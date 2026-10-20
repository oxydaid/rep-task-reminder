package scheduler

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"whatsapp-reminder/backend/internal/database"
	"whatsapp-reminder/backend/internal/reminder"
)

type mockSender struct {
	mu       sync.Mutex
	messages []string
	shouldFail bool
}

func (m *mockSender) SendMessage(targetPhone, text string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.shouldFail {
		return fmt.Errorf("simulated network error")
	}
	m.messages = append(m.messages, text)
	return nil
}

func (m *mockSender) getMessages() []string {
	m.mu.Lock()
	defer m.mu.Unlock()
	dst := make([]string, len(m.messages))
	copy(dst, m.messages)
	return dst
}

func setupSchedulerTest(t *testing.T) (*reminder.Service, *mockSender, *Scheduler, func()) {
	t.Helper()
	tempDir, err := os.MkdirTemp("", "sched_test_*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	dbPath := filepath.Join(tempDir, "test.db")
	db, err := database.Open(dbPath)
	if err != nil {
		t.Fatalf("failed to open test db: %v", err)
	}

	remSvc := reminder.NewService(db)
	sender := &mockSender{}
	sched := NewScheduler(remSvc, sender, 100*time.Millisecond)

	cleanup := func() {
		db.Close()
		os.RemoveAll(tempDir)
	}
	return remSvc, sender, sched, cleanup
}

func TestScheduler_ProcessDueDaily(t *testing.T) {
	remSvc, sender, sched, cleanup := setupSchedulerTest(t)
	defer cleanup()

	loc, _ := time.LoadLocation("Asia/Jakarta")
	schedTime := time.Now().Add(-5 * time.Minute).In(loc)

	rem := &reminder.Reminder{
		Title:          "Absen Masuk",
		Message:        "Jangan lupa absen kehadiran kantor",
		ScheduledAt:    schedTime,
		RecurrenceType: reminder.RecurrenceDaily,
		Timezone:       "Asia/Jakarta",
		IsActive:       true,
	}
	if err := remSvc.CreateReminder(rem); err != nil {
		t.Fatalf("CreateReminder failed: %v", err)
	}

	// Run scheduler pass
	sched.ProcessDueReminders(time.Now().UTC())

	// Check mock sender
	msgs := sender.getMessages()
	if len(msgs) != 1 {
		t.Fatalf("expected 1 message sent, got %d", len(msgs))
	}

	// Check reminder schedule was updated to tomorrow
	updated, err := remSvc.GetReminder(rem.ID)
	if err != nil {
		t.Fatalf("GetReminder failed: %v", err)
	}
	if !updated.IsActive {
		t.Errorf("daily reminder should remain active")
	}
	if !updated.ScheduledAt.After(schedTime) {
		t.Errorf("expected ScheduledAt to advance into future, got %v", updated.ScheduledAt)
	}

	// Check delivery log
	logs, total, err := remSvc.ListDeliveryLogs(10, 0)
	if err != nil || total != 1 {
		t.Fatalf("expected 1 delivery log, got total=%d, err=%v", total, err)
	}
	if logs[0].Status != reminder.StatusSent {
		t.Errorf("expected log status 'sent', got %v", logs[0].Status)
	}
}

func TestScheduler_ProcessDueOnce(t *testing.T) {
	remSvc, sender, sched, cleanup := setupSchedulerTest(t)
	defer cleanup()

	pastTime := time.Now().Add(-10 * time.Minute)
	rem := &reminder.Reminder{
		Title:          "Meeting Klien",
		Message:        "Zoom link 123",
		ScheduledAt:    pastTime,
		RecurrenceType: reminder.RecurrenceOnce,
		Timezone:       "Asia/Jakarta",
		IsActive:       true,
	}
	_ = remSvc.CreateReminder(rem)

	sched.ProcessDueReminders(time.Now().UTC())

	msgs := sender.getMessages()
	if len(msgs) != 1 {
		t.Fatalf("expected 1 message sent, got %d", len(msgs))
	}

	updated, _ := remSvc.GetReminder(rem.ID)
	if updated.IsActive {
		t.Errorf("once reminder should become inactive after dispatch")
	}
}

func TestScheduler_DuplicatePrevention(t *testing.T) {
	remSvc, sender, sched, cleanup := setupSchedulerTest(t)
	defer cleanup()

	pastTime := time.Now().Add(-10 * time.Minute)
	rem := &reminder.Reminder{
		Title:          "Single Send Test",
		Message:        "Test duplicate prevention",
		ScheduledAt:    pastTime,
		RecurrenceType: reminder.RecurrenceOnce,
		Timezone:       "Asia/Jakarta",
		IsActive:       true,
	}
	_ = remSvc.CreateReminder(rem)

	// Call dispatch directly twice with same scheduled time
	sched.dispatchReminder(rem, time.Now().UTC())
	sched.dispatchReminder(rem, time.Now().UTC())

	// Should only send 1 message
	msgs := sender.getMessages()
	if len(msgs) != 1 {
		t.Fatalf("expected exactly 1 message despite multiple dispatches, got %d", len(msgs))
	}
}
