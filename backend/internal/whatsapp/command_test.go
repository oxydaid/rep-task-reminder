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

func TestCommandHandler_HelpAndUnknown(t *testing.T) {
	_, handler, cleanup := setupTestServices(t)
	defer cleanup()

	helpResp := handler.Handle("help")
	if !strings.Contains(helpResp, "Daftar Perintah") {
		t.Errorf("expected help message, got %s", helpResp)
	}

	unknownResp := handler.Handle("halo apa kabar")
	if !strings.Contains(unknownResp, "Perintah tidak dikenali") {
		t.Errorf("expected unknown message, got %s", unknownResp)
	}
}

func TestCommandHandler_CreateAndList(t *testing.T) {
	_, handler, cleanup := setupTestServices(t)
	defer cleanup()

	// Create reminder
	createResp := handler.Handle("buat Minum obat | 12:00 | setiap hari")
	if !strings.Contains(createResp, "Berhasil Dibuat") || !strings.Contains(createResp, "Minum obat") {
		t.Fatalf("create failed, got: %s", createResp)
	}

	// List
	listResp := handler.Handle("daftar")
	if !strings.Contains(listResp, "Minum obat") || !strings.Contains(listResp, "[1]") {
		t.Fatalf("list failed, got: %s", listResp)
	}

	// Detail
	detailResp := handler.Handle("detail 1")
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

	// Edit time
	editResp := handler.Handle("edit 1 jam 14:00")
	if !strings.Contains(editResp, "berhasil diubah ke *14:00*") {
		t.Fatalf("edit failed: %s", editResp)
	}

	// Deactivate
	deactResp := handler.Handle("nonaktifkan 1")
	if !strings.Contains(deactResp, "dinonaktifkan") {
		t.Fatalf("deactivate failed: %s", deactResp)
	}

	// Activate
	actResp := handler.Handle("aktifkan 1")
	if !strings.Contains(actResp, "diaktifkan") {
		t.Fatalf("activate failed: %s", actResp)
	}

	// Delete
	delResp := handler.Handle("hapus 1")
	if !strings.Contains(delResp, "telah berhasil dihapus") {
		t.Fatalf("delete failed: %s", delResp)
	}

	// Detail should now fail
	detailResp := handler.Handle("detail 1")
	if !strings.Contains(detailResp, "tidak ditemukan") {
		t.Fatalf("expected not found after delete, got: %s", detailResp)
	}
}
