package whatsapp

import (
	"os"
	"path/filepath"
	"testing"

	"whatsapp-reminder/backend/internal/database"
	"whatsapp-reminder/backend/internal/reminder"
)

func TestWhatsAppClient_InitialStatus(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "wa_client_test_*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	dbPath := filepath.Join(tempDir, "test.db")
	db, err := database.Open(dbPath)
	if err != nil {
		t.Fatalf("failed to open test db: %v", err)
	}
	defer db.Close()

	remSvc := reminder.NewService(db)
	cmdHandler := NewCommandHandler(remSvc, "Asia/Jakarta")

	waDBPath := filepath.Join(tempDir, "wa.db")
	client, err := NewClient(db, cmdHandler, waDBPath)
	if err != nil {
		t.Fatalf("failed to create whatsapp client: %v", err)
	}

	err = client.Start()
	if err != nil {
		t.Fatalf("failed to start client: %v", err)
	}
	defer client.Disconnect()

	// Initial status must be disconnected and phone must be empty
	status := client.GetStatus()
	if status.Status != StatusDisconnected {
		t.Errorf("expected StatusDisconnected, got %s", status.Status)
	}
	if status.ConnectedPhone != "" {
		t.Errorf("expected empty ConnectedPhone, got %s", status.ConnectedPhone)
	}
	if status.QRCodeDataURL != "" {
		t.Errorf("expected empty QRCodeDataURL initially, got %s", status.QRCodeDataURL)
	}
}
