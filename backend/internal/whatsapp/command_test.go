package whatsapp

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"whatsapp-reminder/backend/internal/database"
	"whatsapp-reminder/backend/internal/reminder"
)

func setupTestServices(t *testing.T) (*reminder.Service, *CommandHandler, func()) {
	t.Helper()
	tempDir, err := os.MkdirTemp("", "cmd_test_*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	dbPath := filepath.Join(tempDir, "test.db")
	db, err := database.Open(dbPath)
	if err != nil {
		t.Fatalf("failed to open test db: %v", err)
	}

	remSvc := reminder.NewService(db)
	cmdHandler := NewCommandHandler(remSvc, "Asia/Jakarta")

	cleanup := func() {
		db.Close()
		os.RemoveAll(tempDir)
	}
	return remSvc, cmdHandler, cleanup
}

func TestCommandHandler_PrefixRequirement(t *testing.T) {
	_, handler, cleanup := setupTestServices(t)
	defer cleanup()

	// Regular chats without '!' must be completely ignored (return "")
	ignoredResp := handler.Handle("halo apa kabar")
	if ignoredResp != "" {
		t.Errorf("expected empty string for non-prefixed message, got: %s", ignoredResp)
	}

	ignoredHelp := handler.Handle("help")
	if ignoredHelp != "" {
		t.Errorf("expected empty string for 'help' without '!', got: %s", ignoredHelp)
	}

	// Commands starting with '!' must be processed
	helpResp := handler.Handle("!help")
	if !strings.Contains(helpResp, "Daftar Perintah") {
		t.Errorf("expected help message for !help, got: %s", helpResp)
	}

	bantuanResp := handler.Handle("!bantuan")
	if !strings.Contains(bantuanResp, "Daftar Perintah") {
		t.Errorf("expected help message for !bantuan, got: %s", bantuanResp)
	}

	pingResp := handler.Handle("!ping")
	if !strings.Contains(pingResp, "Pong") {
		t.Errorf("expected pong message for !ping, got: %s", pingResp)
	}

	unknownResp := handler.Handle("!blablabla")
	if !strings.Contains(unknownResp, "Perintah tidak dikenali") {
		t.Errorf("expected unknown command message for !blablabla, got: %s", unknownResp)
	}
}

func TestCommandHandler_CreateAndList(t *testing.T) {
	_, handler, cleanup := setupTestServices(t)
	defer cleanup()

	// Create reminder using !buat
	createResp := handler.Handle("!buat Minum obat | 12:00 | setiap hari")
	if !strings.Contains(createResp, "Berhasil Dibuat") || !strings.Contains(createResp, "Minum obat") {
		t.Fatalf("create failed, got: %s", createResp)
	}

	// Create reminder using alias !tambah
	createResp2 := handler.Handle("!tambah Olahraga pagi | 06:30 | senin, rabu, jumat")
	if !strings.Contains(createResp2, "Berhasil Dibuat") || !strings.Contains(createResp2, "Olahraga pagi") {
		t.Fatalf("create with !tambah failed, got: %s", createResp2)
	}

	// List using !daftar
	listResp := handler.Handle("!daftar")
	if !strings.Contains(listResp, "Minum obat") || !strings.Contains(listResp, "Olahraga pagi") {
		t.Fatalf("list failed, got: %s", listResp)
	}

	// List using alias !list
	listResp2 := handler.Handle("!list")
	if !strings.Contains(listResp2, "Minum obat") {
		t.Fatalf("list with !list failed, got: %s", listResp2)
	}

	// Detail using !detail 1
	detailResp := handler.Handle("!detail 1")
	if !strings.Contains(detailResp, "Detail Reminder #1") {
		t.Fatalf("detail failed, got: %s", detailResp)
	}
}

func TestCommandHandler_EditToggleDelete(t *testing.T) {
	remSvc, handler, cleanup := setupTestServices(t)
	defer cleanup()

	// Seed reminder
	rem := &reminder.Reminder{
		Title:          "Rapat Mingguan",
		Message:        "Ikuti rapat via zoom",
		ScheduledAt:    time.Now().Add(2 * time.Hour),
		RecurrenceType: reminder.RecurrenceDaily,
		Timezone:       "Asia/Jakarta",
		IsActive:       true,
	}
	_ = remSvc.CreateReminder(rem)

	// Edit time using !edit
	editResp := handler.Handle("!edit 1 jam 14:00")
	if !strings.Contains(editResp, "berhasil diubah ke *14:00*") {
		t.Fatalf("edit failed: %s", editResp)
	}

	// Deactivate using !nonaktifkan
	deactResp := handler.Handle("!nonaktifkan 1")
	if !strings.Contains(deactResp, "dinonaktifkan") {
		t.Fatalf("deactivate failed: %s", deactResp)
	}

	// Activate using !aktifkan
	actResp := handler.Handle("!aktifkan 1")
	if !strings.Contains(actResp, "diaktifkan") {
		t.Fatalf("activate failed: %s", actResp)
	}

	// Delete using !hapus
	delResp := handler.Handle("!hapus 1")
	if !strings.Contains(delResp, "telah berhasil dihapus") {
		t.Fatalf("delete failed: %s", delResp)
	}

	// Detail should now fail
	detailResp := handler.Handle("!detail 1")
	if !strings.Contains(detailResp, "tidak ditemukan") {
		t.Fatalf("expected not found after delete, got: %s", detailResp)
	}
}
