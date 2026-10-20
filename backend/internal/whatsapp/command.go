package whatsapp

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"whatsapp-reminder/backend/internal/reminder"
)

type CommandHandler struct {
	reminderService *reminder.Service
	defaultTimezone string
}

func NewCommandHandler(svc *reminder.Service, tz string) *CommandHandler {
	if tz == "" {
		tz = "Asia/Jakarta"
	}
	return &CommandHandler{
		reminderService: svc,
		defaultTimezone: tz,
	}
}

func (h *CommandHandler) GetLocation() *time.Location {
	loc, err := time.LoadLocation(h.defaultTimezone)
	if err != nil {
		return time.FixedZone("WIB", 7*3600)
	}
	return loc
}

// Handle processes incoming text message with '!' prefix and returns response text
func (h *CommandHandler) Handle(input string) string {
	text := strings.TrimSpace(input)
	if !strings.HasPrefix(text, "!") {
		// Ignore any chat that does not start with '!' prefix
		return ""
	}

	content := strings.TrimSpace(strings.TrimPrefix(text, "!"))
	if content == "" {
		return "Ketik *!help* untuk melihat panduan perintah WhatsApp Reminder."
	}

	fields := strings.Fields(content)
	if len(fields) == 0 {
		return "Ketik *!help* untuk melihat panduan perintah WhatsApp Reminder."
	}

	cmd := strings.ToLower(fields[0])

	switch cmd {
	case "help", "bantuan":
		return h.helpMessage()

	case "ping":
		return "🏓 *Pong!* Bot WhatsApp Reminder aktif & siap menerima perintah.\nKetik *!help* untuk melihat panduan lengkap."

	case "daftar", "list":
		return h.handleList()

	case "detail":
		return h.handleDetail(content)

	case "buat", "tambah":
		return h.handleCreate(content)

	case "edit":
		return h.handleEdit(content)

	case "aktifkan", "on":
		return h.handleToggle(content, true)

	case "nonaktifkan", "off":
		return h.handleToggle(content, false)

	case "hapus", "delete":
		return h.handleDelete(content)

	default:
		return "Perintah tidak dikenali. Ketik *!help* untuk melihat daftar perintah WhatsApp Reminder."
	}
}

func (h *CommandHandler) helpMessage() string {
	return `*Daftar Perintah WhatsApp Reminder* (Prefix: *!*)

1. *Lihat Reminder:*
• *!daftar* (atau *!list*) - Tampilkan semua reminder
• *!detail <id>* - Lihat detail reminder (contoh: *!detail 1*)
• *!ping* - Cek koneksi bot

2. *Buat Reminder:*
Format: *!buat <Judul> | <Waktu> | <Pengulangan>*
(Bisa juga pakai *!tambah*)
Contoh:
• *!buat Absen kerja | 08:00 | setiap hari*
• *!buat Minum obat | 12:00 | setiap hari*
• *!buat Olahraga | 06:30 | senin, rabu, jumat*
• *!buat Deadline Project | 2026-10-25 15:00 | sekali*

3. *Ubah & Kelola:*
• *!edit <id> jam <HH:mm>* - Ubah jam (contoh: *!edit 1 jam 13:00*)
• *!aktifkan <id>* - Aktifkan reminder (contoh: *!aktifkan 1*)
• *!nonaktifkan <id>* - Matikan reminder (contoh: *!nonaktifkan 1*)
• *!hapus <id>* - Hapus reminder (contoh: *!hapus 1*)`
}

func (h *CommandHandler) handleList() string {
	items, err := h.reminderService.ListReminders("all")
	if err != nil {
		return "Gagal mengambil daftar reminder: " + err.Error()
	}
	if len(items) == 0 {
		return "Belum ada reminder yang terdaftar. Ketik *!help* untuk cara membuat reminder baru."
	}

	loc := h.GetLocation()
	var sb strings.Builder
	sb.WriteString("*Daftar Reminder Anda:*\n\n")

	for _, item := range items {
		status := "🟢 Aktif"
		if !item.IsActive {
			status = "⚪ Nonaktif"
		}

		timeStr := item.ScheduledAt.In(loc).Format("02 Jan 15:04")
		recurrenceDesc := string(item.RecurrenceType)
		switch item.RecurrenceType {
		case reminder.RecurrenceDaily:
			recurrenceDesc = "Setiap hari"
			timeStr = item.ScheduledAt.In(loc).Format("15:04")
		case reminder.RecurrenceSpecificDays:
			recurrenceDesc = "Hari: " + item.RecurrenceDays
			timeStr = item.ScheduledAt.In(loc).Format("15:04")
		case reminder.RecurrenceWeekly:
			recurrenceDesc = "Mingguan"
		case reminder.RecurrenceOnce:
			recurrenceDesc = "Sekali"
		}

		sb.WriteString(fmt.Sprintf("*[%d]* %s\n", item.ID, item.Title))
		sb.WriteString(fmt.Sprintf("⏰ %s (%s) | %s\n\n", timeStr, recurrenceDesc, status))
	}

	sb.WriteString("Ketik *!detail <id>* untuk melihat detail.")
	return sb.String()
}

func (h *CommandHandler) handleDetail(content string) string {
	parts := strings.Fields(content)
	if len(parts) < 2 {
		return "Format salah. Gunakan: *!detail <id>* (contoh: *!detail 1*)"
	}
	id, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		return "ID reminder harus berupa angka. Contoh: *!detail 1*"
	}

	item, err := h.reminderService.GetReminder(id)
	if err != nil {
		return "Reminder dengan ID tersebut tidak ditemukan."
	}

	loc := h.GetLocation()
	status := "Aktif 🟢"
	if !item.IsActive {
		status = "Nonaktif ⚪"
	}

	msg := item.Message
	if msg == "" {
		msg = "(Sama dengan judul)"
	}

	return fmt.Sprintf(`*Detail Reminder #%d*
• *Judul:* %s
• *Pesan:* %s
• *Jadwal Berikutnya:* %s %s
• *Pengulangan:* %s %s
• *Status:* %s`,
		item.ID,
		item.Title,
		msg,
		item.ScheduledAt.In(loc).Format("02 Jan 2006 15:04"),
		item.Timezone,
		item.RecurrenceType,
		item.RecurrenceDays,
		status,
	)
}

func (h *CommandHandler) handleCreate(content string) string {
	idx := strings.Index(content, " ")
	if idx == -1 {
		return "Format salah. Gunakan: *!buat <Judul> | <Waktu> | <Pengulangan>*\nContoh: *!buat Minum obat | 12:00 | setiap hari*"
	}
	rawArgs := strings.TrimSpace(content[idx+1:])
	parts := strings.Split(rawArgs, "|")
	if len(parts) < 2 {
		return "Format salah. Gunakan: *!buat <Judul> | <Waktu> | <Pengulangan>*\nContoh: *!buat Minum obat | 12:00 | setiap hari*"
	}

	title := strings.TrimSpace(parts[0])
	timeStr := strings.TrimSpace(parts[1])
	repeatStr := "sekali"
	if len(parts) >= 3 {
		repeatStr = strings.ToLower(strings.TrimSpace(parts[2]))
	}

	if title == "" {
		return "Judul reminder tidak boleh kosong."
	}

	loc := h.GetLocation()
	now := time.Now().In(loc)

	var recurrenceType reminder.RecurrenceType
	var recurrenceDays string

	switch {
	case repeatStr == "sekali" || repeatStr == "once":
		recurrenceType = reminder.RecurrenceOnce
	case repeatStr == "setiap hari" || repeatStr == "harian" || repeatStr == "daily" || repeatStr == "tiap hari":
		recurrenceType = reminder.RecurrenceDaily
	case repeatStr == "mingguan" || repeatStr == "weekly" || repeatStr == "setiap minggu":
		recurrenceType = reminder.RecurrenceWeekly
	default:
		parsedDays := reminder.ParseRecurrenceDays(repeatStr)
		if len(parsedDays) > 0 {
			recurrenceType = reminder.RecurrenceSpecificDays
			recurrenceDays = reminder.FormatRecurrenceDays(parsedDays)
		} else {
			recurrenceType = reminder.RecurrenceOnce
		}
	}

	var scheduledAt time.Time
	if t, err := time.Parse("15:04", timeStr); err == nil {
		scheduledAt = reminder.CalculateInitialRun(t.Hour(), t.Minute(), recurrenceType, recurrenceDays, loc, now)
	} else if fullT, err := time.ParseInLocation("2006-01-02 15:04", timeStr, loc); err == nil {
		scheduledAt = fullT
	} else {
		return "Format waktu tidak valid. Gunakan jam (contoh: 08:00) atau tanggal lengkap (contoh: 2026-10-25 15:00)."
	}

	newReminder := &reminder.Reminder{
		Title:          title,
		Message:        title,
		ScheduledAt:    scheduledAt,
		RecurrenceType: recurrenceType,
		RecurrenceDays: recurrenceDays,
		Timezone:       h.defaultTimezone,
		IsActive:       true,
	}

	if err := h.reminderService.CreateReminder(newReminder); err != nil {
		return "Gagal menyimpan reminder: " + err.Error()
	}

	return fmt.Sprintf("✅ *Reminder Berhasil Dibuat!*\n\n• *ID:* %d\n• *Judul:* %s\n• *Jadwal Pertama:* %s (%s)\n• *Pengulangan:* %s",
		newReminder.ID,
		newReminder.Title,
		newReminder.ScheduledAt.In(loc).Format("02 Jan 2006 15:04"),
		newReminder.Timezone,
		newReminder.RecurrenceType,
	)
}

func (h *CommandHandler) handleEdit(content string) string {
	parts := strings.Fields(content)
	if len(parts) < 4 || strings.ToLower(parts[2]) != "jam" {
		return "Format salah. Gunakan: *!edit <id> jam <HH:mm>*\nContoh: *!edit 1 jam 13:00*"
	}

	id, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		return "ID reminder harus berupa angka."
	}

	item, err := h.reminderService.GetReminder(id)
	if err != nil {
		return "Reminder tidak ditemukan."
	}

	timeStr := parts[3]
	t, err := time.Parse("15:04", timeStr)
	if err != nil {
		return "Format jam tidak valid. Gunakan format 24 jam (contoh: 13:00)."
	}

	loc := h.GetLocation()
	now := time.Now().In(loc)

	item.ScheduledAt = reminder.CalculateInitialRun(t.Hour(), t.Minute(), item.RecurrenceType, item.RecurrenceDays, loc, now)

	if err := h.reminderService.UpdateReminder(item); err != nil {
		return "Gagal mengupdate reminder: " + err.Error()
	}

	return fmt.Sprintf("✅ Jam reminder *[%d] %s* berhasil diubah ke *%s* (Jadwal berikutnya: %s)",
		item.ID, item.Title, timeStr, item.ScheduledAt.In(loc).Format("02 Jan 15:04"))
}

func (h *CommandHandler) handleToggle(content string, active bool) string {
	parts := strings.Fields(content)
	if len(parts) < 2 {
		cmdName := "aktifkan"
		if !active {
			cmdName = "nonaktifkan"
		}
		return "Format salah. Gunakan: *!" + cmdName + " <id>*"
	}

	id, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		return "ID reminder harus berupa angka."
	}

	item, err := h.reminderService.GetReminder(id)
	if err != nil {
		return "Reminder tidak ditemukan."
	}

	if err := h.reminderService.ToggleReminder(id, active); err != nil {
		return "Gagal mengubah status reminder: " + err.Error()
	}

	action := "diaktifkan 🟢"
	if !active {
		action = "dinonaktifkan ⚪"
	}

	return fmt.Sprintf("✅ Reminder *[%d] %s* berhasil %s.", item.ID, item.Title, action)
}

func (h *CommandHandler) handleDelete(content string) string {
	parts := strings.Fields(content)
	if len(parts) < 2 {
		return "Format salah. Gunakan: *!hapus <id>*"
	}

	id, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		return "ID reminder harus berupa angka."
	}

	item, err := h.reminderService.GetReminder(id)
	if err != nil {
		return "Reminder tidak ditemukan."
	}

	if err := h.reminderService.DeleteReminder(id); err != nil {
		return "Gagal menghapus reminder: " + err.Error()
	}

	return fmt.Sprintf("✅ Reminder *[%d] %s* telah berhasil dihapus.", id, item.Title)
}
