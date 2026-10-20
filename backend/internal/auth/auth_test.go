package auth

import (
	"os"
	"path/filepath"
	"testing"

	"whatsapp-reminder/backend/internal/database"
)

func setupTestDB(t *testing.T) (*database.DB, func()) {
	t.Helper()
	tempDir, err := os.MkdirTemp("", "auth_test_*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	dbPath := filepath.Join(tempDir, "test.db")
	db, err := database.Open(dbPath)
	if err != nil {
		t.Fatalf("failed to open test db: %v", err)
	}

	cleanup := func() {
		db.Close()
		os.RemoveAll(tempDir)
	}
	return db, cleanup
}

func TestAuthService_SetupAndLogin(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	svc, err := NewAuthService(db)
	if err != nil {
		t.Fatalf("NewAuthService failed: %v", err)
	}

	status, err := svc.GetStatus()
	if err != nil {
		t.Fatalf("GetStatus failed: %v", err)
	}
	if !status.SetupRequired {
		t.Errorf("expected SetupRequired=true")
	}

	// Setup PIN
	token, err := svc.SetupInitialPIN("123456")
	if err != nil {
		t.Fatalf("SetupInitialPIN failed: %v", err)
	}
	if token == "" {
		t.Errorf("expected non-empty token")
	}

	// Validate token
	claims, err := svc.ValidateToken(token)
	if err != nil || claims.Subject != "dashboard_admin" {
		t.Fatalf("ValidateToken failed: %v", err)
	}

	// Status should now be SetupRequired=false
	status, err = svc.GetStatus()
	if err != nil || status.SetupRequired {
		t.Errorf("expected SetupRequired=false, got %v", status)
	}

	// Login with correct PIN
	loginToken, err := svc.Login("123456")
	if err != nil || loginToken == "" {
		t.Fatalf("Login with correct PIN failed: %v", err)
	}

	// Login with wrong PIN
	_, err = svc.Login("000000")
	if err == nil {
		t.Fatalf("expected error on wrong PIN")
	}
}

func TestAuthService_Lockout(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	svc, _ := NewAuthService(db)
	_, _ = svc.SetupInitialPIN("123456")

	// Trigger 5 failed attempts
	for i := 0; i < MaxFailedAttempts-1; i++ {
		_, err := svc.Login("wrong")
		if err == nil {
			t.Fatalf("expected error on wrong attempt %d", i)
		}
	}

	// 5th attempt triggers lockout
	_, err := svc.Login("wrong")
	if err == nil {
		t.Fatalf("expected lockout error on 5th attempt")
	}

	status, _ := svc.GetStatus()
	if !status.Locked {
		t.Errorf("expected Locked=true after max failures")
	}

	// Even correct PIN should be rejected during lockout
	_, err = svc.Login("123456")
	if err == nil {
		t.Fatalf("expected login rejection while locked")
	}
}

func TestAuthService_ChangePIN(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	svc, _ := NewAuthService(db)
	_, _ = svc.SetupInitialPIN("123456")

	// Change with wrong old PIN
	err := svc.ChangePIN("999999", "654321")
	if err == nil {
		t.Fatalf("expected error when old PIN is wrong")
	}

	// Change with correct old PIN
	err = svc.ChangePIN("123456", "654321")
	if err != nil {
		t.Fatalf("ChangePIN failed: %v", err)
	}

	// Verify old PIN no longer works
	_, err = svc.Login("123456")
	if err == nil {
		t.Fatalf("old PIN should no longer work")
	}

	// Verify new PIN works
	_, err = svc.Login("654321")
	if err != nil {
		t.Fatalf("new PIN should work: %v", err)
	}
}
