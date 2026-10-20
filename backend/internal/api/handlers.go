package api

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"whatsapp-reminder/backend/internal/auth"
	"whatsapp-reminder/backend/internal/database"
	"whatsapp-reminder/backend/internal/reminder"
	"whatsapp-reminder/backend/internal/whatsapp"
)

type Handler struct {
	db          *database.DB
	authSvc     *auth.AuthService
	remSvc      *reminder.Service
	waClient    *whatsapp.Client
	backupDir   string
}

func NewHandler(db *database.DB, authSvc *auth.AuthService, remSvc *reminder.Service, waClient *whatsapp.Client, backupDir string) *Handler {
	return &Handler{
		db:        db,
		authSvc:   authSvc,
		remSvc:    remSvc,
		waClient:  waClient,
		backupDir: backupDir,
	}
}

// Health check
func (h *Handler) Health(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"status": "ok",
		"time":   time.Now().UTC().Format(time.RFC3339),
	})
}

// Auth handlers
func (h *Handler) GetAuthStatus(c *fiber.Ctx) error {
	status, err := h.authSvc.GetStatus()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(status)
}

func (h *Handler) SetupPIN(c *fiber.Ctx) error {
	var body struct {
		PIN string `json:"pin"`
	}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Format request tidak valid"})
	}

	body.PIN = strings.TrimSpace(body.PIN)
	token, err := h.authSvc.SetupInitialPIN(body.PIN)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	c.Cookie(&fiber.Cookie{
		Name:     "auth_token",
		Value:    token,
		Expires:  time.Now().Add(7 * 24 * time.Hour),
		HTTPOnly: true,
		SameSite: "Lax",
	})

	return c.JSON(fiber.Map{
		"message": "PIN berhasil diatur",
		"token":   token,
	})
}

func (h *Handler) Login(c *fiber.Ctx) error {
	var body struct {
		PIN string `json:"pin"`
	}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Format request tidak valid"})
	}

	body.PIN = strings.TrimSpace(body.PIN)
	token, err := h.authSvc.Login(body.PIN)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": err.Error()})
	}

	c.Cookie(&fiber.Cookie{
		Name:     "auth_token",
		Value:    token,
		Expires:  time.Now().Add(7 * 24 * time.Hour),
		HTTPOnly: true,
		SameSite: "Lax",
	})

	return c.JSON(fiber.Map{
		"message": "Login berhasil",
		"token":   token,
	})
}

func (h *Handler) Logout(c *fiber.Ctx) error {
	c.ClearCookie("auth_token")
	return c.JSON(fiber.Map{"message": "Logout berhasil"})
}

func (h *Handler) ChangePIN(c *fiber.Ctx) error {
	var body struct {
		OldPIN string `json:"old_pin"`
		NewPIN string `json:"new_pin"`
	}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Format request tidak valid"})
	}

	body.OldPIN = strings.TrimSpace(body.OldPIN)
	body.NewPIN = strings.TrimSpace(body.NewPIN)

	if err := h.authSvc.ChangePIN(body.OldPIN, body.NewPIN); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"message": "PIN berhasil diperbarui"})
}

// Dashboard statistics
func (h *Handler) GetStats(c *fiber.Ctx) error {
	stats, err := h.remSvc.GetStats()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(stats)
}

// Reminders CRUD
func (h *Handler) ListReminders(c *fiber.Ctx) error {
	status := c.Query("status", "all")
	list, err := h.remSvc.ListReminders(status)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	if list == nil {
		list = []*reminder.Reminder{}
	}
	return c.JSON(list)
}

func (h *Handler) GetUpcomingReminder(c *fiber.Ctx) error {
	item, err := h.remSvc.GetNextUpcomingReminder()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(item)
}

func (h *Handler) GetReminder(c *fiber.Ctx) error {
	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "ID tidak valid"})
	}

	item, err := h.remSvc.GetReminder(id)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(item)
}

type ReminderPayload struct {
	Title          string `json:"title"`
	Message        string `json:"message"`
	ScheduledAt    string `json:"scheduled_at"`
	RecurrenceType string `json:"recurrence_type"`
	RecurrenceDays string `json:"recurrence_days"`
	Timezone       string `json:"timezone"`
	TargetPhone    string `json:"target_phone"`
	IsActive       *bool  `json:"is_active"`
}

func (h *Handler) CreateReminder(c *fiber.Ctx) error {
	var body ReminderPayload
	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Format data tidak valid"})
	}

	body.Title = strings.TrimSpace(body.Title)
	if body.Title == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Judul reminder wajib diisi"})
	}

	if body.Message == "" {
		body.Message = body.Title
	}

	if body.Timezone == "" {
		body.Timezone = "Asia/Jakarta"
	}
	loc, err := time.LoadLocation(body.Timezone)
	if err != nil {
		loc = time.FixedZone("WIB", 7*3600)
	}

	var scheduledTime time.Time
	if body.ScheduledAt != "" {
		if t, err := time.Parse(time.RFC3339, body.ScheduledAt); err == nil {
			scheduledTime = t
		} else if t, err := time.ParseInLocation("2006-01-02 15:04", body.ScheduledAt, loc); err == nil {
			scheduledTime = t
		} else if t, err := time.ParseInLocation("2006-01-02T15:04", body.ScheduledAt, loc); err == nil {
			scheduledTime = t
		}
	}

	if scheduledTime.IsZero() {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Waktu jadwal tidak valid"})
	}

	recurrenceType := reminder.RecurrenceType(body.RecurrenceType)
	if recurrenceType == "" {
		recurrenceType = reminder.RecurrenceOnce
	}

	isActive := true
	if body.IsActive != nil {
		isActive = *body.IsActive
	}

	item := &reminder.Reminder{
		Title:          body.Title,
		Message:        body.Message,
		ScheduledAt:    scheduledTime,
		RecurrenceType: recurrenceType,
		RecurrenceDays: body.RecurrenceDays,
		Timezone:       body.Timezone,
		TargetPhone:    whatsapp.NormalizePhone(body.TargetPhone),
		IsActive:       isActive,
	}

	if err := h.remSvc.CreateReminder(item); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(fiber.StatusCreated).JSON(item)
}

func (h *Handler) UpdateReminder(c *fiber.Ctx) error {
	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "ID tidak valid"})
	}

	item, err := h.remSvc.GetReminder(id)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Reminder tidak ditemukan"})
	}

	var body ReminderPayload
	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Format data tidak valid"})
	}

	if strings.TrimSpace(body.Title) != "" {
		item.Title = strings.TrimSpace(body.Title)
	}
	if body.Message != "" {
		item.Message = body.Message
	}
	if body.Timezone != "" {
		item.Timezone = body.Timezone
	}
	loc, _ := time.LoadLocation(item.Timezone)
	if loc == nil {
		loc = time.FixedZone("WIB", 7*3600)
	}

	if body.ScheduledAt != "" {
		if t, err := time.Parse(time.RFC3339, body.ScheduledAt); err == nil {
			item.ScheduledAt = t
		} else if t, err := time.ParseInLocation("2006-01-02 15:04", body.ScheduledAt, loc); err == nil {
			item.ScheduledAt = t
		} else if t, err := time.ParseInLocation("2006-01-02T15:04", body.ScheduledAt, loc); err == nil {
			item.ScheduledAt = t
		}
	}

	if body.RecurrenceType != "" {
		item.RecurrenceType = reminder.RecurrenceType(body.RecurrenceType)
	}
	item.RecurrenceDays = body.RecurrenceDays

	if body.TargetPhone != "" {
		item.TargetPhone = whatsapp.NormalizePhone(body.TargetPhone)
	}
	if body.IsActive != nil {
		item.IsActive = *body.IsActive
	}

	if err := h.remSvc.UpdateReminder(item); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(item)
}

func (h *Handler) DeleteReminder(c *fiber.Ctx) error {
	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "ID tidak valid"})
	}

	if err := h.remSvc.DeleteReminder(id); err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"message": "Reminder berhasil dihapus"})
}

func (h *Handler) ToggleReminder(c *fiber.Ctx) error {
	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "ID tidak valid"})
	}

	var body struct {
		IsActive bool `json:"is_active"`
	}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Format data tidak valid"})
	}

	if err := h.remSvc.ToggleReminder(id, body.IsActive); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{
		"message":   "Status berhasil diperbarui",
		"is_active": body.IsActive,
	})
}

// Delivery Logs
func (h *Handler) ListLogs(c *fiber.Ctx) error {
	limit, _ := strconv.Atoi(c.Query("limit", "20"))
	offset, _ := strconv.Atoi(c.Query("offset", "0"))

	logs, total, err := h.remSvc.ListDeliveryLogs(limit, offset)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	if logs == nil {
		logs = []*reminder.DeliveryLog{}
	}

	return c.JSON(fiber.Map{
		"data":   logs,
		"total":  total,
		"limit":  limit,
		"offset": offset,
	})
}

// WhatsApp Connection
func (h *Handler) GetWhatsAppStatus(c *fiber.Ctx) error {
	if h.waClient == nil {
		return c.JSON(fiber.Map{"status": "disconnected"})
	}
	return c.JSON(h.waClient.GetStatus())
}

func (h *Handler) ConnectWhatsApp(c *fiber.Ctx) error {
	if h.waClient == nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "WhatsApp client belum siap"})
	}
	if err := h.waClient.Connect(); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(h.waClient.GetStatus())
}

func (h *Handler) DisconnectWhatsApp(c *fiber.Ctx) error {
	if h.waClient != nil {
		h.waClient.Disconnect()
	}
	return c.JSON(fiber.Map{"message": "WhatsApp terputus"})
}

func (h *Handler) LogoutWhatsApp(c *fiber.Ctx) error {
	if h.waClient == nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "WhatsApp client belum siap"})
	}
	if err := h.waClient.Logout(); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"message": "Sesi WhatsApp berhasil dihapus (unpair)"})
}

func (h *Handler) TestSendWhatsApp(c *fiber.Ctx) error {
	var body struct {
		Phone   string `json:"phone"`
		Message string `json:"message"`
	}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Format data tidak valid"})
	}

	if body.Message == "" {
		body.Message = "Halo! Ini adalah pesan uji coba dari sistem WhatsApp Reminder."
	}

	if err := h.waClient.SendMessage(body.Phone, body.Message); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"message": "Pesan uji coba berhasil dikirim"})
}

// Settings
func (h *Handler) GetSettings(c *fiber.Ctx) error {
	targetPhone, _ := h.db.GetSetting("target_phone")
	tz, _ := h.db.GetSetting("timezone")
	if tz == "" {
		tz = "Asia/Jakarta"
	}

	return c.JSON(fiber.Map{
		"target_phone": targetPhone,
		"timezone":     tz,
	})
}

func (h *Handler) UpdateSettings(c *fiber.Ctx) error {
	var body struct {
		TargetPhone string `json:"target_phone"`
		Timezone    string `json:"timezone"`
	}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Format data tidak valid"})
	}

	if body.TargetPhone != "" {
		cleanPhone := whatsapp.NormalizePhone(body.TargetPhone)
		_ = h.db.SetSetting("target_phone", cleanPhone)
	} else if body.TargetPhone == "" {
		_ = h.db.SetSetting("target_phone", "")
	}

	if body.Timezone != "" {
		_ = h.db.SetSetting("timezone", body.Timezone)
	}

	return c.JSON(fiber.Map{"message": "Pengaturan berhasil disimpan"})
}

// Backup database
func (h *Handler) TriggerBackup(c *fiber.Ctx) error {
	if h.backupDir == "" {
		h.backupDir = "data/backups"
	}
	_ = os.MkdirAll(h.backupDir, 0755)

	backupFileName := fmt.Sprintf("backup_%s.db", time.Now().Format("20060102_150405"))
	destPath := filepath.Join(h.backupDir, backupFileName)

	if err := h.db.Backup(destPath); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": fmt.Sprintf("Gagal backup: %v", err)})
	}

	return c.JSON(fiber.Map{
		"message":  "Backup berhasil dibuat",
		"filename": backupFileName,
		"path":     destPath,
	})
}
