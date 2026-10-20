package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"whatsapp-reminder/backend/internal/auth"
	"whatsapp-reminder/backend/internal/database"
	"whatsapp-reminder/backend/internal/reminder"
	"whatsapp-reminder/backend/internal/whatsapp"
)

func setupTestApp(t *testing.T) (*fiber.App, *auth.AuthService, func()) {
	t.Helper()
	tempDir, err := os.MkdirTemp("", "api_test_*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	dbPath := filepath.Join(tempDir, "test.db")
	db, err := database.Open(dbPath)
	if err != nil {
		t.Fatalf("failed to open test db: %v", err)
	}

	authSvc, err := auth.NewAuthService(db)
	if err != nil {
		t.Fatalf("failed to create auth svc: %v", err)
	}
	remSvc := reminder.NewService(db)
	cmdHandler := whatsapp.NewCommandHandler(remSvc, "Asia/Jakarta")
	waDBPath := filepath.Join(tempDir, "wa.db")
	waClient, err := whatsapp.NewClient(db, cmdHandler, waDBPath)
	if err != nil {
		t.Fatalf("failed to create wa client: %v", err)
	}

	app := fiber.New()
	handler := NewHandler(db, authSvc, remSvc, waClient, filepath.Join(tempDir, "backups"))
	SetupRouter(app, handler, "")

	cleanup := func() {
		db.Close()
		os.RemoveAll(tempDir)
	}
	return app, authSvc, cleanup
}

func TestHealthEndpoint(t *testing.T) {
	app, _, cleanup := setupTestApp(t)
	defer cleanup()

	req := httptest.NewRequest("GET", "/health", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("health request failed: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", resp.StatusCode)
	}
}

func TestAuthAndRemindersFlow(t *testing.T) {
	app, _, cleanup := setupTestApp(t)
	defer cleanup()

	// 1. Initial status -> setup_required should be true
	req := httptest.NewRequest("GET", "/api/auth/status", nil)
	resp, _ := app.Test(req)
	var status struct {
		SetupRequired bool `json:"setup_required"`
	}
	_ = json.NewDecoder(resp.Body).Decode(&status)
	if !status.SetupRequired {
		t.Fatalf("expected setup_required to be true initially")
	}

	// 2. Setup initial PIN
	setupBody := []byte(`{"pin":"123456"}`)
	req = httptest.NewRequest("POST", "/api/auth/setup", bytes.NewBuffer(setupBody))
	req.Header.Set("Content-Type", "application/json")
	resp, _ = app.Test(req)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("setup failed with status %d", resp.StatusCode)
	}

	var authRes struct {
		Token string `json:"token"`
	}
	_ = json.NewDecoder(resp.Body).Decode(&authRes)
	token := authRes.Token
	if token == "" {
		t.Fatalf("expected non-empty token")
	}

	// 3. Unauthorized access without token
	req = httptest.NewRequest("GET", "/api/reminders", nil)
	resp, _ = app.Test(req)
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected 401 Unauthorized, got %d", resp.StatusCode)
	}

	// 4. Authorized access: Create reminder
	remBody := []byte(`{
		"title": "Minum Vitamin",
		"message": "Minum vitamin C setelah makan",
		"scheduled_at": "` + time.Now().Add(1*time.Hour).Format(time.RFC3339) + `",
		"recurrence_type": "daily",
		"timezone": "Asia/Jakarta"
	}`)
	req = httptest.NewRequest("POST", "/api/reminders", bytes.NewBuffer(remBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	resp, _ = app.Test(req)
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("create reminder failed with status %d", resp.StatusCode)
	}

	var created reminder.Reminder
	_ = json.NewDecoder(resp.Body).Decode(&created)
	if created.ID == 0 || created.Title != "Minum Vitamin" {
		t.Fatalf("unexpected reminder data: %+v", created)
	}

	// 5. List reminders
	req = httptest.NewRequest("GET", "/api/reminders", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	resp, _ = app.Test(req)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("list reminders failed with status %d", resp.StatusCode)
	}
	var list []reminder.Reminder
	_ = json.NewDecoder(resp.Body).Decode(&list)
	if len(list) != 1 {
		t.Fatalf("expected 1 reminder, got %d", len(list))
	}

	// 6. Delete reminder
	req = httptest.NewRequest("DELETE", "/api/reminders/1", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	resp, _ = app.Test(req)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("delete reminder failed with status %d", resp.StatusCode)
	}
}
