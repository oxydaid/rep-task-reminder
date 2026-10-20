package reminder

import (
	"database/sql"
	"errors"
	"fmt"
	"sync"
	"time"

	"whatsapp-reminder/backend/internal/database"
)

type Service struct {
	db *database.DB
	mu sync.Mutex
}

func NewService(db *database.DB) *Service {
	return &Service{db: db}
}

// CreateReminder adds a new reminder
func (s *Service) CreateReminder(r *Reminder) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now().UTC()
	r.CreatedAt = now
	r.UpdatedAt = now

	if r.Timezone == "" {
		r.Timezone = "Asia/Jakarta"
	}
	if r.RecurrenceType == "" {
		r.RecurrenceType = RecurrenceOnce
	}

	query := `
		INSERT INTO reminders (title, message, scheduled_at, recurrence_type, recurrence_days, timezone, target_phone, is_active, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	res, err := s.db.Exec(query,
		r.Title,
		r.Message,
		r.ScheduledAt.UTC().Format(time.RFC3339),
		string(r.RecurrenceType),
		r.RecurrenceDays,
		r.Timezone,
		r.TargetPhone,
		r.IsActive,
		r.CreatedAt.Format(time.RFC3339),
		r.UpdatedAt.Format(time.RFC3339),
	)
	if err != nil {
		return fmt.Errorf("failed to create reminder: %w", err)
	}

	id, err := res.LastInsertId()
	if err != nil {
		return err
	}
	r.ID = id
	return nil
}

// GetReminder finds a reminder by ID
func (s *Service) GetReminder(id int64) (*Reminder, error) {
	query := `
		SELECT id, title, message, scheduled_at, recurrence_type, recurrence_days, timezone, target_phone, is_active, created_at, updated_at
		FROM reminders WHERE id = ?
	`
	r := &Reminder{}
	var schedStr, createdStr, updatedStr string

	err := s.db.QueryRow(query, id).Scan(
		&r.ID, &r.Title, &r.Message, &schedStr, &r.RecurrenceType, &r.RecurrenceDays,
		&r.Timezone, &r.TargetPhone, &r.IsActive, &createdStr, &updatedStr,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("reminder tidak ditemukan")
		}
		return nil, err
	}

	r.ScheduledAt, _ = time.Parse(time.RFC3339, schedStr)
	r.CreatedAt, _ = time.Parse(time.RFC3339, createdStr)
	r.UpdatedAt, _ = time.Parse(time.RFC3339, updatedStr)
	return r, nil
}

// ListReminders returns reminders filtered by status ("all", "active", "inactive")
func (s *Service) ListReminders(filter string) ([]*Reminder, error) {
	query := `
		SELECT id, title, message, scheduled_at, recurrence_type, recurrence_days, timezone, target_phone, is_active, created_at, updated_at
		FROM reminders
	`
	var args []any
	switch filter {
	case "active":
		query += " WHERE is_active = 1"
	case "inactive":
		query += " WHERE is_active = 0"
	}
	query += " ORDER BY is_active DESC, scheduled_at ASC"

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []*Reminder
	for rows.Next() {
		r := &Reminder{}
		var schedStr, createdStr, updatedStr string
		err := rows.Scan(
			&r.ID, &r.Title, &r.Message, &schedStr, &r.RecurrenceType, &r.RecurrenceDays,
			&r.Timezone, &r.TargetPhone, &r.IsActive, &createdStr, &updatedStr,
		)
		if err != nil {
			return nil, err
		}
		r.ScheduledAt, _ = time.Parse(time.RFC3339, schedStr)
		r.CreatedAt, _ = time.Parse(time.RFC3339, createdStr)
		r.UpdatedAt, _ = time.Parse(time.RFC3339, updatedStr)
		list = append(list, r)
	}

	return list, nil
}

// UpdateReminder updates an existing reminder
func (s *Service) UpdateReminder(r *Reminder) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	r.UpdatedAt = time.Now().UTC()
	query := `
		UPDATE reminders
		SET title = ?, message = ?, scheduled_at = ?, recurrence_type = ?, recurrence_days = ?, timezone = ?, target_phone = ?, is_active = ?, updated_at = ?
		WHERE id = ?
	`
	res, err := s.db.Exec(query,
		r.Title,
		r.Message,
		r.ScheduledAt.UTC().Format(time.RFC3339),
		string(r.RecurrenceType),
		r.RecurrenceDays,
		r.Timezone,
		r.TargetPhone,
		r.IsActive,
		r.UpdatedAt.Format(time.RFC3339),
		r.ID,
	)
	if err != nil {
		return err
	}
	rows, _ := res.RowsAffected()
	if rows == 0 {
		return errors.New("reminder tidak ditemukan")
	}
	return nil
}

// DeleteReminder removes a reminder and cascades delivery_logs
func (s *Service) DeleteReminder(id int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	res, err := s.db.Exec("DELETE FROM reminders WHERE id = ?", id)
	if err != nil {
		return err
	}
	rows, _ := res.RowsAffected()
	if rows == 0 {
		return errors.New("reminder tidak ditemukan")
	}
	return nil
}

// ToggleReminder sets active state
func (s *Service) ToggleReminder(id int64, active bool) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now().UTC().Format(time.RFC3339)
	res, err := s.db.Exec("UPDATE reminders SET is_active = ?, updated_at = ? WHERE id = ?", active, now, id)
	if err != nil {
		return err
	}
	rows, _ := res.RowsAffected()
	if rows == 0 {
		return errors.New("reminder tidak ditemukan")
	}
	return nil
}

// GetNextUpcomingReminder gets the soonest active reminder
func (s *Service) GetNextUpcomingReminder() (*Reminder, error) {
	query := `
		SELECT id, title, message, scheduled_at, recurrence_type, recurrence_days, timezone, target_phone, is_active, created_at, updated_at
		FROM reminders
		WHERE is_active = 1
		ORDER BY scheduled_at ASC
		LIMIT 1
	`
	r := &Reminder{}
	var schedStr, createdStr, updatedStr string
	err := s.db.QueryRow(query).Scan(
		&r.ID, &r.Title, &r.Message, &schedStr, &r.RecurrenceType, &r.RecurrenceDays,
		&r.Timezone, &r.TargetPhone, &r.IsActive, &createdStr, &updatedStr,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	r.ScheduledAt, _ = time.Parse(time.RFC3339, schedStr)
	r.CreatedAt, _ = time.Parse(time.RFC3339, createdStr)
	r.UpdatedAt, _ = time.Parse(time.RFC3339, updatedStr)
	return r, nil
}

// GetStats returns summary statistics for dashboard
func (s *Service) GetStats() (*Stats, error) {
	stats := &Stats{}

	_ = s.db.QueryRow("SELECT COUNT(*) FROM reminders").Scan(&stats.TotalReminders)
	_ = s.db.QueryRow("SELECT COUNT(*) FROM reminders WHERE is_active = 1").Scan(&stats.ActiveReminders)
	stats.InactiveReminders = stats.TotalReminders - stats.ActiveReminders

	// Today's scheduled reminders in UTC or local
	startOfDay := time.Now().Truncate(24 * time.Hour).UTC().Format(time.RFC3339)
	endOfDay := time.Now().Truncate(24 * time.Hour).Add(24 * time.Hour).UTC().Format(time.RFC3339)
	_ = s.db.QueryRow("SELECT COUNT(*) FROM reminders WHERE is_active = 1 AND scheduled_at >= ? AND scheduled_at < ?", startOfDay, endOfDay).Scan(&stats.TodayReminders)

	_ = s.db.QueryRow("SELECT COUNT(*) FROM delivery_logs WHERE status = 'sent'").Scan(&stats.TotalSent)
	_ = s.db.QueryRow("SELECT COUNT(*) FROM delivery_logs WHERE status = 'failed'").Scan(&stats.TotalFailed)

	return stats, nil
}

// GetDueReminders returns active reminders that are due at or before 'now'
func (s *Service) GetDueReminders(now time.Time) ([]*Reminder, error) {
	query := `
		SELECT id, title, message, scheduled_at, recurrence_type, recurrence_days, timezone, target_phone, is_active, created_at, updated_at
		FROM reminders
		WHERE is_active = 1 AND scheduled_at <= ?
		ORDER BY scheduled_at ASC
	`
	rows, err := s.db.Query(query, now.UTC().Format(time.RFC3339))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []*Reminder
	for rows.Next() {
		r := &Reminder{}
		var schedStr, createdStr, updatedStr string
		err := rows.Scan(
			&r.ID, &r.Title, &r.Message, &schedStr, &r.RecurrenceType, &r.RecurrenceDays,
			&r.Timezone, &r.TargetPhone, &r.IsActive, &createdStr, &updatedStr,
		)
		if err != nil {
			return nil, err
		}
		r.ScheduledAt, _ = time.Parse(time.RFC3339, schedStr)
		r.CreatedAt, _ = time.Parse(time.RFC3339, createdStr)
		r.UpdatedAt, _ = time.Parse(time.RFC3339, updatedStr)
		list = append(list, r)
	}

	return list, nil
}

var ErrAlreadyDispatched = errors.New("reminder already dispatched for this schedule")

// RecordDeliveryAttempt tries to insert a delivery log record.
// If UNIQUE constraint fails, it returns ErrAlreadyDispatched (duplicate prevention).
func (s *Service) RecordDeliveryAttempt(reminderID int64, scheduledFor time.Time, status DeliveryStatus, errMsg string) (int64, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now().UTC()
	schedStr := scheduledFor.UTC().Format(time.RFC3339)

	var sentAtStr *string
	if status == StatusSent {
		tStr := now.Format(time.RFC3339)
		sentAtStr = &tStr
	}

	query := `
		INSERT INTO delivery_logs (reminder_id, scheduled_for, status, attempts, sent_at, error_message, created_at)
		VALUES (?, ?, ?, 1, ?, ?, ?)
	`
	res, err := s.db.Exec(query, reminderID, schedStr, string(status), sentAtStr, errMsg, now.Format(time.RFC3339))
	if err != nil {
		return 0, ErrAlreadyDispatched
	}

	return res.LastInsertId()
}

// MarkLogStatus updates delivery status for an existing log
func (s *Service) MarkLogStatus(logID int64, status DeliveryStatus, errMsg string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	var sentAtStr *string
	if status == StatusSent {
		tStr := time.Now().UTC().Format(time.RFC3339)
		sentAtStr = &tStr
	}

	query := `
		UPDATE delivery_logs
		SET status = ?, sent_at = ?, error_message = ?
		WHERE id = ?
	`
	_, err := s.db.Exec(query, string(status), sentAtStr, errMsg, logID)
	return err
}

// ListDeliveryLogs returns delivery history
func (s *Service) ListDeliveryLogs(limit, offset int) ([]*DeliveryLog, int64, error) {
	if limit <= 0 {
		limit = 20
	}
	var total int64
	_ = s.db.QueryRow("SELECT COUNT(*) FROM delivery_logs").Scan(&total)

	query := `
		SELECT l.id, l.reminder_id, COALESCE(r.title, '[Dihapus]') as title, l.scheduled_for, l.status, l.attempts, l.sent_at, l.error_message, l.created_at
		FROM delivery_logs l
		LEFT JOIN reminders r ON l.reminder_id = r.id
		ORDER BY l.created_at DESC
		LIMIT ? OFFSET ?
	`
	rows, err := s.db.Query(query, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var logs []*DeliveryLog
	for rows.Next() {
		l := &DeliveryLog{}
		var schedStr, createdStr string
		var sentStr sql.NullString
		var errStr sql.NullString
		err := rows.Scan(
			&l.ID, &l.ReminderID, &l.ReminderTitle, &schedStr, &l.Status, &l.Attempts,
			&sentStr, &errStr, &createdStr,
		)
		if err != nil {
			return nil, 0, err
		}
		l.ScheduledFor, _ = time.Parse(time.RFC3339, schedStr)
		l.CreatedAt, _ = time.Parse(time.RFC3339, createdStr)
		if sentStr.Valid {
			t, _ := time.Parse(time.RFC3339, sentStr.String)
			l.SentAt = &t
		}
		if errStr.Valid {
			l.ErrorMessage = errStr.String
		}
		logs = append(logs, l)
	}

	return logs, total, nil
}
