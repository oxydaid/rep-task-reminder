package database

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

// DB wraps the sql.DB instance
type DB struct {
	*sql.DB
	path string
}

// Open initializes SQLite database with WAL and foreign keys
func Open(dbPath string) (*DB, error) {
	dir := filepath.Dir(dbPath)
	if dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, fmt.Errorf("failed to create db directory: %w", err)
		}
	}

	dsn := fmt.Sprintf("file:%s?_pragma=busy_timeout(5000)&_pragma=journal_mode(WAL)&_pragma=foreign_keys(1)&_pragma=synchronous(NORMAL)", dbPath)
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open sqlite database: %w", err)
	}

	db.SetMaxOpenConns(1) // SQLite works safest with single writer
	db.SetMaxIdleConns(1)

	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping sqlite database: %w", err)
	}

	instance := &DB{DB: db, path: dbPath}
	if err := instance.migrate(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	return instance, nil
}

func (db *DB) Path() string {
	return db.path
}

// migrate creates tables if not exist
func (db *DB) migrate() error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS app_settings (
			key TEXT PRIMARY KEY,
			value TEXT NOT NULL
		);`,
		`CREATE TABLE IF NOT EXISTS reminders (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			title TEXT NOT NULL,
			message TEXT NOT NULL,
			scheduled_at TEXT NOT NULL,
			recurrence_type TEXT NOT NULL DEFAULT 'once',
			recurrence_days TEXT NOT NULL DEFAULT '',
			timezone TEXT NOT NULL DEFAULT 'Asia/Jakarta',
			target_phone TEXT NOT NULL DEFAULT '',
			is_active INTEGER NOT NULL DEFAULT 1,
			created_at TEXT NOT NULL,
			updated_at TEXT NOT NULL
		);`,
		`CREATE TABLE IF NOT EXISTS delivery_logs (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			reminder_id INTEGER NOT NULL REFERENCES reminders(id) ON DELETE CASCADE,
			scheduled_for TEXT NOT NULL,
			status TEXT NOT NULL,
			attempts INTEGER NOT NULL DEFAULT 1,
			sent_at TEXT,
			error_message TEXT,
			created_at TEXT NOT NULL,
			CONSTRAINT uq_reminder_scheduled UNIQUE (reminder_id, scheduled_for)
		);`,
		`CREATE INDEX IF NOT EXISTS idx_reminders_active_schedule ON reminders(is_active, scheduled_at);`,
		`CREATE INDEX IF NOT EXISTS idx_delivery_logs_reminder_id ON delivery_logs(reminder_id);`,
		`CREATE INDEX IF NOT EXISTS idx_delivery_logs_created_at ON delivery_logs(created_at DESC);`,
	}

	for _, query := range queries {
		if _, err := db.Exec(query); err != nil {
			return fmt.Errorf("migration error for '%s': %w", query, err)
		}
	}

	return nil
}

// GetSetting retrieves a setting value by key
func (db *DB) GetSetting(key string) (string, error) {
	var val string
	err := db.QueryRow("SELECT value FROM app_settings WHERE key = ?", key).Scan(&val)
	if err == sql.ErrNoRows {
		return "", nil
	}
	return val, err
}

// SetSetting saves or updates a setting
func (db *DB) SetSetting(key, value string) error {
	_, err := db.Exec(`
		INSERT INTO app_settings (key, value) VALUES (?, ?)
		ON CONFLICT(key) DO UPDATE SET value = excluded.value
	`, key, value)
	return err
}

// Backup creates a safe online backup using VACUUM INTO
func (db *DB) Backup(destPath string) error {
	destDir := filepath.Dir(destPath)
	if destDir != "" && destDir != "." {
		if err := os.MkdirAll(destDir, 0755); err != nil {
			return err
		}
	}
	// Remove existing backup if exists so VACUUM INTO won't fail
	_ = os.Remove(destPath)

	query := fmt.Sprintf("VACUUM INTO '%s'", destPath)
	_, err := db.Exec(query)
	return err
}
