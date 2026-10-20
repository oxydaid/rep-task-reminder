package whatsapp

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"sync"

	"github.com/skip2/go-qrcode"
	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/proto/waE2E"
	"go.mau.fi/whatsmeow/store/sqlstore"
	"go.mau.fi/whatsmeow/types"
	"go.mau.fi/whatsmeow/types/events"
	waLog "go.mau.fi/whatsmeow/util/log"
	"google.golang.org/protobuf/proto"
	_ "modernc.org/sqlite"

	"whatsapp-reminder/backend/internal/database"
)

type ConnectionStatus string

const (
	StatusConnected    ConnectionStatus = "connected"
	StatusConnecting   ConnectionStatus = "connecting"
	StatusDisconnected ConnectionStatus = "disconnected"
)

type ClientStatus struct {
	Status        ConnectionStatus `json:"status"`
	ConnectedPhone string          `json:"connected_phone"`
	QRCodeDataURL  string          `json:"qr_code_data_url,omitempty"`
	TargetPhone    string          `json:"target_phone"`
}

type Client struct {
	db             *database.DB
	container      *sqlstore.Container
	waClient       *whatsmeow.Client
	commandHandler *CommandHandler
	status         ConnectionStatus
	qrDataURL      string
	qrRaw          string
	mu             sync.RWMutex
	stopQRChan     chan struct{}
}

func NewClient(db *database.DB, cmdHandler *CommandHandler, sessionDBPath string) (*Client, error) {
	dbLog := waLog.Stdout("WhatsAppDB", "ERROR", true)

	container, err := sqlstore.New(context.Background(), "sqlite", fmt.Sprintf("file:%s?_pragma=foreign_keys(1)&_pragma=journal_mode(WAL)&_pragma=busy_timeout(5000)", sessionDBPath), dbLog)
	if err != nil {
		return nil, fmt.Errorf("failed to open whatsapp sqlstore: %w", err)
	}

	c := &Client{
		db:             db,
		container:      container,
		commandHandler: cmdHandler,
		status:         StatusDisconnected,
	}

	return c, nil
}

// Start loads first device or creates new, and initiates connection if session exists
func (c *Client) Start() error {
	deviceStore, err := c.container.GetFirstDevice(context.Background())
	if err != nil {
		return fmt.Errorf("failed to get device store: %w", err)
	}

	clientLog := waLog.Stdout("WhatsApp", "INFO", true)
	c.waClient = whatsmeow.NewClient(deviceStore, clientLog)
	c.waClient.AddEventHandler(c.handleEvent)

	if c.waClient.Store.ID != nil {
		// Session exists, connect
		c.mu.Lock()
		c.status = StatusConnecting
		c.mu.Unlock()
		return c.waClient.Connect()
	}

	return nil
}

// Connect starts WhatsApp connection and initiates QR pairing if needed
func (c *Client) Connect() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.waClient == nil {
		return errors.New("client belum diinisialisasi")
	}

	if c.waClient.IsConnected() {
		c.status = StatusConnected
		return nil
	}

	c.status = StatusConnecting
	c.qrDataURL = ""
	c.qrRaw = ""

	if c.waClient.Store.ID == nil {
		// No session yet -> listen for QR channel
		qrChan, err := c.waClient.GetQRChannel(context.Background())
		if err != nil {
			c.status = StatusDisconnected
			return fmt.Errorf("failed to get qr channel: %w", err)
		}

		err = c.waClient.Connect()
		if err != nil {
			c.status = StatusDisconnected
			return fmt.Errorf("failed to connect: %w", err)
		}

		go c.listenForQR(qrChan)
	} else {
		// Existing session
		err := c.waClient.Connect()
		if err != nil {
			c.status = StatusDisconnected
			return err
		}
	}

	return nil
}

// Disconnect closes the socket connection
func (c *Client) Disconnect() {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.waClient != nil && c.waClient.IsConnected() {
		c.waClient.Disconnect()
	}
	c.status = StatusDisconnected
	c.qrDataURL = ""
	c.qrRaw = ""
}

// Logout disconnects and unpairs device, deleting session from DB
func (c *Client) Logout() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.waClient != nil {
		if c.waClient.IsConnected() {
			_ = c.waClient.Logout(context.Background())
		}
		c.waClient.Disconnect()
		if c.waClient.Store != nil {
			_ = c.waClient.Store.Delete(context.Background())
		}
	}

	c.status = StatusDisconnected
	c.qrDataURL = ""
	c.qrRaw = ""

	// Recreate fresh device store
	deviceStore := c.container.NewDevice()
	clientLog := waLog.Stdout("WhatsApp", "INFO", true)
	c.waClient = whatsmeow.NewClient(deviceStore, clientLog)
	c.waClient.AddEventHandler(c.handleEvent)

	return nil
}

func (c *Client) listenForQR(qrChan <-chan whatsmeow.QRChannelItem) {
	for item := range qrChan {
		switch item.Event {
		case "code":
			pngBytes, err := qrcode.Encode(item.Code, qrcode.Medium, 280)
			if err == nil {
				dataURL := "data:image/png;base64," + base64.StdEncoding.EncodeToString(pngBytes)
				c.mu.Lock()
				c.qrRaw = item.Code
				c.qrDataURL = dataURL
				c.status = StatusConnecting
				c.mu.Unlock()
			}
		case "success":
			c.mu.Lock()
			c.status = StatusConnected
			c.qrDataURL = ""
			c.qrRaw = ""
			c.mu.Unlock()
			return
		case "timeout":
			c.mu.Lock()
			c.status = StatusDisconnected
			c.qrDataURL = ""
			c.qrRaw = ""
			c.mu.Unlock()
			return
		}
	}
}

func (c *Client) handleEvent(evt interface{}) {
	switch v := evt.(type) {
	case *events.Connected:
		c.mu.Lock()
		c.status = StatusConnected
		c.qrDataURL = ""
		c.qrRaw = ""
		c.mu.Unlock()

	case *events.LoggedOut:
		c.mu.Lock()
		c.status = StatusDisconnected
		c.qrDataURL = ""
		c.qrRaw = ""
		c.mu.Unlock()

	case *events.Disconnected:
		c.mu.Lock()
		if c.status != StatusConnecting {
			c.status = StatusDisconnected
		}
		c.mu.Unlock()

	case *events.Message:
		c.handleIncomingMessage(v)
	}
}

func (c *Client) handleIncomingMessage(evt *events.Message) {
	if evt == nil || evt.Message == nil || evt.Info.IsFromMe {
		return
	}

	// Extract text content
	text := evt.Message.GetConversation()
	if text == "" && evt.Message.ExtendedTextMessage != nil {
		text = evt.Message.ExtendedTextMessage.GetText()
	}
	text = strings.TrimSpace(text)
	if text == "" {
		return
	}

	// Authorize sender
	senderPhone := evt.Info.Sender.User
	if !c.isSenderAuthorized(senderPhone) {
		return
	}

	// Handle command
	if c.commandHandler == nil {
		return
	}

	reply := c.commandHandler.Handle(text)
	if reply != "" {
		_, _ = c.waClient.SendMessage(context.Background(), evt.Info.Chat, &waE2E.Message{
			Conversation: proto.String(reply),
		})
	}
}

func (c *Client) isSenderAuthorized(phone string) bool {
	phone = NormalizePhone(phone)
	if phone == "" {
		return false
	}

	// 1. Check if sender is the connected device itself
	c.mu.RLock()
	if c.waClient != nil && c.waClient.Store != nil && c.waClient.Store.ID != nil {
		connPhone := NormalizePhone(c.waClient.Store.ID.User)
		if connPhone != "" && connPhone == phone {
			c.mu.RUnlock()
			return true
		}
	}
	c.mu.RUnlock()

	// 2. Check if sender matches target_phone in database
	targetPhone, err := c.db.GetSetting("target_phone")
	if err == nil && targetPhone != "" {
		if NormalizePhone(targetPhone) == phone {
			return true
		}
	}

	return false
}

// GetStatus returns the current connection state and details
func (c *Client) GetStatus() *ClientStatus {
	c.mu.RLock()
	defer c.mu.RUnlock()

	var connPhone string
	if c.waClient != nil && c.waClient.Store != nil && c.waClient.Store.ID != nil {
		connPhone = c.waClient.Store.ID.User
	}

	status := c.status
	if c.waClient != nil && c.waClient.IsConnected() {
		status = StatusConnected
	}

	targetPhone, _ := c.db.GetSetting("target_phone")

	return &ClientStatus{
		Status:         status,
		ConnectedPhone: connPhone,
		QRCodeDataURL:  c.qrDataURL,
		TargetPhone:    targetPhone,
	}
}

// SendMessage sends a text message to the specified phone number or fallback
func (c *Client) SendMessage(targetPhone string, messageText string) error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.waClient == nil || !c.waClient.IsConnected() {
		return errors.New("whatsapp belum terhubung")
	}

	phone := NormalizePhone(targetPhone)
	if phone == "" {
		// Fallback to configured target_phone
		cfgPhone, _ := c.db.GetSetting("target_phone")
		phone = NormalizePhone(cfgPhone)
	}
	if phone == "" {
		// Fallback to connected device own phone
		if c.waClient.Store != nil && c.waClient.Store.ID != nil {
			phone = NormalizePhone(c.waClient.Store.ID.User)
		}
	}

	if phone == "" {
		return errors.New("tidak ada nomor tujuan pengiriman WhatsApp yang valid")
	}

	recipientJID := types.NewJID(phone, types.DefaultUserServer)

	msg := &waE2E.Message{
		Conversation: proto.String(messageText),
	}

	_, err := c.waClient.SendMessage(context.Background(), recipientJID, msg)
	if err != nil {
		return fmt.Errorf("gagal mengirim pesan whatsapp: %w", err)
	}

	return nil
}

// NormalizePhone formats Indonesian and international phone numbers into standard digits
func NormalizePhone(input string) string {
	clean := regexp.MustCompile(`[^0-9]`).ReplaceAllString(input, "")
	if clean == "" {
		return ""
	}

	// 08xx -> 628xx
	if strings.HasPrefix(clean, "0") {
		clean = "62" + clean[1:]
	}

	return clean
}
