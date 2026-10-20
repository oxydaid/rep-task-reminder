package whatsapp

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"log"
	"regexp"
	"strings"
	"sync"
	"time"

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
	Status         ConnectionStatus `json:"status"`
	ConnectedPhone string           `json:"connected_phone"`
	QRCodeDataURL  string           `json:"qr_code_data_url,omitempty"`
	TargetPhone    string           `json:"target_phone"`
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

// Start loads first device or creates new, and initiates connection if paired session exists
func (c *Client) Start() error {
	deviceStore, err := c.container.GetFirstDevice(context.Background())
	if err != nil {
		return fmt.Errorf("failed to get device store: %w", err)
	}

	clientLog := waLog.Stdout("WhatsApp", "INFO", true)
	c.waClient = whatsmeow.NewClient(deviceStore, clientLog)
	c.waClient.AddEventHandler(c.handleEvent)

	if c.waClient.Store.ID != nil {
		// Session exists, connect in background
		c.mu.Lock()
		c.status = StatusConnecting
		c.mu.Unlock()
		return c.waClient.Connect()
	}

	return nil
}

// Connect starts WhatsApp connection and initiates QR pairing if not yet paired
func (c *Client) Connect() error {
	c.mu.Lock()

	if c.waClient == nil {
		c.mu.Unlock()
		return errors.New("client belum diinisialisasi")
	}

	// 1. If already paired and logged in
	if c.waClient.Store.ID != nil && c.waClient.IsLoggedIn() {
		c.status = StatusConnected
		c.qrDataURL = ""
		c.qrRaw = ""
		c.mu.Unlock()
		return nil
	}

	// 2. If already paired device, but disconnected -> reconnect
	if c.waClient.Store.ID != nil {
		c.status = StatusConnecting
		c.qrDataURL = ""
		c.qrRaw = ""
		c.mu.Unlock()
		if !c.waClient.IsConnected() {
			return c.waClient.Connect()
		}
		return nil
	}

	// 3. Unpaired device: must get QR channel and connect
	// If websocket is already open without QR channel, disconnect first
	if c.waClient.IsConnected() {
		c.waClient.Disconnect()
	}

	c.status = StatusConnecting
	c.qrDataURL = ""
	c.qrRaw = ""

	qrChan, err := c.waClient.GetQRChannel(context.Background())
	if err != nil {
		c.status = StatusDisconnected
		c.mu.Unlock()
		return fmt.Errorf("failed to get qr channel: %w", err)
	}

	err = c.waClient.Connect()
	if err != nil {
		c.status = StatusDisconnected
		c.mu.Unlock()
		return fmt.Errorf("failed to connect: %w", err)
	}

	qrReady := make(chan struct{}, 1)
	go c.listenForQR(qrChan, qrReady)
	c.mu.Unlock()

	// Wait briefly (up to 1.5s) so the immediate response contains QR code
	select {
	case <-qrReady:
	case <-time.After(1500 * time.Millisecond):
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

	log.Println("[WhatsApp] Device session deleted (unpaired). Fresh device ready.")
	return nil
}

func (c *Client) listenForQR(qrChan <-chan whatsmeow.QRChannelItem, qrReady chan<- struct{}) {
	first := true
	for item := range qrChan {
		switch item.Event {
		case "code":
			log.Printf("[WhatsApp] New QR code received, generating image...\n")
			pngBytes, err := qrcode.Encode(item.Code, qrcode.Medium, 280)
			if err == nil {
				dataURL := "data:image/png;base64," + base64.StdEncoding.EncodeToString(pngBytes)
				c.mu.Lock()
				c.qrRaw = item.Code
				c.qrDataURL = dataURL
				c.status = StatusConnecting
				c.mu.Unlock()
			}
			if first {
				first = false
				select {
				case qrReady <- struct{}{}:
				default:
				}
			}
		case "success":
			log.Printf("[WhatsApp] QR pairing successful!\n")
			c.mu.Lock()
			c.status = StatusConnected
			c.qrDataURL = ""
			c.qrRaw = ""
			c.mu.Unlock()
			return
		case "timeout":
			log.Printf("[WhatsApp] QR pairing timed out.\n")
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
		if c.waClient.Store.ID != nil && c.waClient.IsLoggedIn() {
			c.status = StatusConnected
			c.qrDataURL = ""
			c.qrRaw = ""
			log.Printf("[WhatsApp] Authenticated & connected as +%s\n", c.waClient.Store.ID.User)
		}
		c.mu.Unlock()

	case *events.PairSuccess:
		c.mu.Lock()
		c.status = StatusConnected
		c.qrDataURL = ""
		c.qrRaw = ""
		log.Printf("[WhatsApp] Pair success! Connected as +%s\n", v.ID.User)
		c.mu.Unlock()

	case *events.LoggedOut:
		c.mu.Lock()
		c.status = StatusDisconnected
		c.qrDataURL = ""
		c.qrRaw = ""
		log.Println("[WhatsApp] Device logged out")
		c.mu.Unlock()

	case *events.Disconnected:
		c.mu.Lock()
		if c.status == StatusConnected || c.status != StatusConnecting {
			c.status = StatusDisconnected
		}
		c.mu.Unlock()

	case *events.Message:
		c.handleIncomingMessage(v)
	}
}

func (c *Client) handleIncomingMessage(evt *events.Message) {
	if evt == nil || evt.Message == nil {
		return
	}

	// 1. Extract text content
	text := extractMessageText(evt.Message)
	text = strings.TrimSpace(text)
	if text == "" {
		return
	}

	// Only process if it starts with the command prefix '!'
	if !strings.HasPrefix(text, "!") {
		return
	}

	// 2. Authorize sender
	if !c.isSenderAuthorized(evt) {
		log.Printf("[WhatsApp] Ignored command from unauthorized sender: %s (Chat: %s)\n", evt.Info.Sender.String(), evt.Info.Chat.String())
		return
	}

	log.Printf("[WhatsApp] Received command from %s: %s\n", evt.Info.Sender.String(), text)

	// 3. Handle command
	if c.commandHandler == nil {
		return
	}

	reply := c.commandHandler.Handle(text)
	if reply == "" {
		return
	}

	// 4. Send reply back to the chat
	targetJID := evt.Info.Chat.ToNonAD()
	if targetJID.IsEmpty() {
		targetJID = evt.Info.Sender.ToNonAD()
	}

	resp, err := c.waClient.SendMessage(context.Background(), targetJID, &waE2E.Message{
		Conversation: proto.String(reply),
	})
	if err != nil {
		log.Printf("[WhatsApp] Failed to send command reply to %s: %v\n", targetJID.String(), err)
	} else {
		log.Printf("[WhatsApp] Command reply sent to %s (ID: %s)\n", targetJID.String(), resp.ID)
	}
}

func extractMessageText(msg *waE2E.Message) string {
	if msg == nil {
		return ""
	}
	if msg.GetEphemeralMessage() != nil && msg.GetEphemeralMessage().GetMessage() != nil {
		msg = msg.GetEphemeralMessage().GetMessage()
	}
	if msg.GetViewOnceMessage() != nil && msg.GetViewOnceMessage().GetMessage() != nil {
		msg = msg.GetViewOnceMessage().GetMessage()
	}
	if msg.GetViewOnceMessageV2() != nil && msg.GetViewOnceMessageV2().GetMessage() != nil {
		msg = msg.GetViewOnceMessageV2().GetMessage()
	}
	if msg.GetDocumentWithCaptionMessage() != nil && msg.GetDocumentWithCaptionMessage().GetMessage() != nil {
		msg = msg.GetDocumentWithCaptionMessage().GetMessage()
	}

	if text := msg.GetConversation(); text != "" {
		return text
	}
	if ext := msg.GetExtendedTextMessage(); ext != nil && ext.GetText() != "" {
		return ext.GetText()
	}
	return ""
}

func (c *Client) isSenderAuthorized(evt *events.Message) bool {
	if evt == nil {
		return false
	}

	// Candidates for sender's phone number
	candidates := []string{
		evt.Info.Sender.User,
		evt.Info.SenderAlt.User,
		evt.Info.Chat.User,
		evt.Info.RecipientAlt.User,
	}

	c.mu.RLock()
	var connPhone string
	if c.waClient != nil && c.waClient.Store != nil && c.waClient.Store.ID != nil {
		connPhone = NormalizePhone(c.waClient.Store.ID.User)
	}
	c.mu.RUnlock()

	targetPhone, err := c.db.GetSetting("target_phone")
	normTargetPhone := NormalizePhone(targetPhone)

	for _, raw := range candidates {
		phone := NormalizePhone(raw)
		if phone == "" {
			continue
		}
		// 1. Is sender the bot/connected device itself?
		if connPhone != "" && phone == connPhone {
			return true
		}
		// 2. Is sender the configured target_phone?
		if normTargetPhone != "" && phone == normTargetPhone {
			return true
		}
	}

	// 3. If target_phone is not yet configured, and this is a direct private chat (not a group)
	if (normTargetPhone == "" || err != nil) && !evt.Info.IsGroup {
		for _, raw := range candidates {
			phone := NormalizePhone(raw)
			if phone != "" && len(phone) >= 9 && len(phone) <= 15 {
				_ = c.db.SetSetting("target_phone", phone)
				log.Printf("[WhatsApp] Auto-configured target_phone to +%s from incoming command\n", phone)
				return true
			}
		}
		return true
	}

	return false
}

// GetStatus returns the current connection state and details
func (c *Client) GetStatus() *ClientStatus {
	c.mu.RLock()
	defer c.mu.RUnlock()

	var connPhone string
	isLoggedIn := false

	if c.waClient != nil && c.waClient.Store != nil && c.waClient.Store.ID != nil {
		connPhone = c.waClient.Store.ID.User
		isLoggedIn = c.waClient.IsLoggedIn()
	}

	var status ConnectionStatus
	if isLoggedIn {
		status = StatusConnected
	} else if c.waClient != nil && c.waClient.Store != nil && c.waClient.Store.ID != nil {
		if c.waClient.IsConnected() || c.status == StatusConnecting {
			status = StatusConnecting
		} else {
			status = StatusDisconnected
		}
	} else {
		if c.status == StatusConnecting {
			status = StatusConnecting
		} else {
			status = StatusDisconnected
		}
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

	if c.waClient == nil || !c.waClient.IsConnected() || !c.waClient.IsLoggedIn() || c.waClient.Store.ID == nil {
		return errors.New("whatsapp belum terhubung atau belum login")
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
