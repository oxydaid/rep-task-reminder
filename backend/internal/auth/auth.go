package auth

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"whatsapp-reminder/backend/internal/database"
)

const (
	SettingPinHash       = "pin_hash"
	SettingJWTSecret     = "jwt_secret"
	SettingFailedLogins  = "failed_login_attempts"
	SettingLockoutUntil  = "login_locked_until"
	MaxFailedAttempts    = 5
	LockoutDuration      = 5 * time.Minute
	TokenExpiryDuration  = 7 * 24 * time.Hour
)

type AuthService struct {
	db        *database.DB
	jwtSecret []byte
	mu        sync.Mutex
}

type AuthStatus struct {
	SetupRequired        bool  `json:"setup_required"`
	Locked               bool  `json:"locked"`
	LockRemainingSeconds int64 `json:"lock_remaining_seconds"`
}

type Claims struct {
	Subject string `json:"sub"`
	jwt.RegisteredClaims
}

func NewAuthService(db *database.DB) (*AuthService, error) {
	s := &AuthService{db: db}

	// Ensure JWT secret exists in DB
	secretHex, err := db.GetSetting(SettingJWTSecret)
	if err != nil {
		return nil, fmt.Errorf("failed to get jwt secret: %w", err)
	}

	if secretHex == "" {
		bytes := make([]byte, 32)
		if _, err := rand.Read(bytes); err != nil {
			return nil, fmt.Errorf("failed to generate random secret: %w", err)
		}
		secretHex = hex.EncodeToString(bytes)
		if err := db.SetSetting(SettingJWTSecret, secretHex); err != nil {
			return nil, fmt.Errorf("failed to save jwt secret: %w", err)
		}
	}

	secret, err := hex.DecodeString(secretHex)
	if err != nil {
		return nil, fmt.Errorf("failed to decode jwt secret: %w", err)
	}
	s.jwtSecret = secret

	return s, nil
}

// GetStatus checks if PIN is configured and whether login is currently locked
func (s *AuthService) GetStatus() (*AuthStatus, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	hash, err := s.db.GetSetting(SettingPinHash)
	if err != nil {
		return nil, err
	}

	status := &AuthStatus{
		SetupRequired: hash == "",
	}

	lockUntilStr, err := s.db.GetSetting(SettingLockoutUntil)
	if err == nil && lockUntilStr != "" {
		if lockUntil, parseErr := strconv.ParseInt(lockUntilStr, 10, 64); parseErr == nil {
			now := time.Now().Unix()
			if lockUntil > now {
				status.Locked = true
				status.LockRemainingSeconds = lockUntil - now
			}
		}
	}

	return status, nil
}

// SetupInitialPIN sets the PIN if none is currently configured
func (s *AuthService) SetupInitialPIN(pin string) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if len(pin) < 4 || len(pin) > 12 {
		return "", errors.New("PIN harus terdiri dari 4 hingga 12 digit")
	}

	existingHash, err := s.db.GetSetting(SettingPinHash)
	if err != nil {
		return "", err
	}
	if existingHash != "" {
		return "", errors.New("PIN sudah diatur sebelumnya. Silakan login")
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(pin), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("failed to hash pin: %w", err)
	}

	if err := s.db.SetSetting(SettingPinHash, string(hashed)); err != nil {
		return "", err
	}

	// Reset any lockouts
	_ = s.db.SetSetting(SettingFailedLogins, "0")
	_ = s.db.SetSetting(SettingLockoutUntil, "0")

	return s.generateToken()
}

// Login verifies the entered PIN with rate limiting
func (s *AuthService) Login(pin string) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Check lockout
	now := time.Now().Unix()
	lockUntilStr, err := s.db.GetSetting(SettingLockoutUntil)
	if err == nil && lockUntilStr != "" {
		if lockUntil, parseErr := strconv.ParseInt(lockUntilStr, 10, 64); parseErr == nil && lockUntil > now {
			remaining := lockUntil - now
			return "", fmt.Errorf("terlalu banyak percobaan gagal. Coba lagi dalam %d detik", remaining)
		}
	}

	hash, err := s.db.GetSetting(SettingPinHash)
	if err != nil {
		return "", err
	}
	if hash == "" {
		return "", errors.New("PIN belum diatur. Silakan lakukan setup terlebih dahulu")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(pin)); err != nil {
		// Increment failed counter
		failedCount := 1
		cntStr, _ := s.db.GetSetting(SettingFailedLogins)
		if cnt, parseErr := strconv.Atoi(cntStr); parseErr == nil {
			failedCount = cnt + 1
		}
		_ = s.db.SetSetting(SettingFailedLogins, strconv.Itoa(failedCount))

		if failedCount >= MaxFailedAttempts {
			lockUntil := time.Now().Add(LockoutDuration).Unix()
			_ = s.db.SetSetting(SettingLockoutUntil, strconv.FormatInt(lockUntil, 10))
			_ = s.db.SetSetting(SettingFailedLogins, "0")
			return "", fmt.Errorf("PIN salah. Batas percobaan terlampaui. Akun terkunci selama %d menit", int(LockoutDuration.Minutes()))
		}

		remaining := MaxFailedAttempts - failedCount
		return "", fmt.Errorf("PIN salah. Sisa percobaan: %d kali", remaining)
	}

	// Success: reset failures
	_ = s.db.SetSetting(SettingFailedLogins, "0")
	_ = s.db.SetSetting(SettingLockoutUntil, "0")

	return s.generateToken()
}

// ChangePIN verifies the old PIN before updating to the new one
func (s *AuthService) ChangePIN(oldPin, newPin string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if len(newPin) < 4 || len(newPin) > 12 {
		return errors.New("PIN baru harus terdiri dari 4 hingga 12 digit")
	}

	hash, err := s.db.GetSetting(SettingPinHash)
	if err != nil {
		return err
	}
	if hash == "" {
		return errors.New("PIN belum diatur")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(oldPin)); err != nil {
		return errors.New("PIN lama tidak sesuai")
	}

	newHashed, err := bcrypt.GenerateFromPassword([]byte(newPin), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash new pin: %w", err)
	}

	return s.db.SetSetting(SettingPinHash, string(newHashed))
}

// ValidateToken verifies JWT token string and returns Claims
func (s *AuthService) ValidateToken(tokenStr string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return s.jwtSecret, nil
	})
	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("token tidak valid")
}

func (s *AuthService) generateToken() (string, error) {
	claims := Claims{
		Subject: "dashboard_admin",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(TokenExpiryDuration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "whatsapp-reminder",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.jwtSecret)
}
