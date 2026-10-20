package scheduler

import (
	"fmt"
	"log"
	"sync"
	"time"

	"whatsapp-reminder/backend/internal/reminder"
)

type WhatsAppSender interface {
	SendMessage(targetPhone, text string) error
}

type Scheduler struct {
	reminderService *reminder.Service
	sender          WhatsAppSender
	interval        time.Duration
	stopChan        chan struct{}
	wg              sync.WaitGroup
	running         bool
	mu              sync.Mutex
}

func NewScheduler(remSvc *reminder.Service, sender WhatsAppSender, interval time.Duration) *Scheduler {
	if interval <= 0 {
		interval = 15 * time.Second
	}
	return &Scheduler{
		reminderService: remSvc,
		sender:          sender,
		interval:        interval,
		stopChan:        make(chan struct{}),
	}
}

// Start initiates the background scheduler loop
func (s *Scheduler) Start() {
	s.mu.Lock()
	if s.running {
		s.mu.Unlock()
		return
	}
	s.running = true
	s.stopChan = make(chan struct{})
	s.wg.Add(1)
	s.mu.Unlock()

	go s.loop()
}

// Stop gracefully terminates the scheduler loop
func (s *Scheduler) Stop() {
	s.mu.Lock()
	if !s.running {
		s.mu.Unlock()
		return
	}
	s.running = false
	close(s.stopChan)
	s.mu.Unlock()

	s.wg.Wait()
}

func (s *Scheduler) loop() {
	defer s.wg.Done()

	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()

	// Initial check on startup
	s.ProcessDueReminders(time.Now().UTC())

	for {
		select {
		case <-s.stopChan:
			return
		case now := <-ticker.C:
			s.ProcessDueReminders(now.UTC())
		}
	}
}

// ProcessDueReminders evaluates and dispatches due reminders
func (s *Scheduler) ProcessDueReminders(now time.Time) {
	dueReminders, err := s.reminderService.GetDueReminders(now)
	if err != nil {
		log.Printf("[Scheduler] Error querying due reminders: %v\n", err)
		return
	}

	for _, item := range dueReminders {
		s.dispatchReminder(item, now)
	}
}

func (s *Scheduler) dispatchReminder(item *reminder.Reminder, now time.Time) {
	scheduledFor := item.ScheduledAt

	// 1. Record attempt in delivery_logs; UNIQUE(reminder_id, scheduled_for) guarantees no duplicate send
	logID, err := s.reminderService.RecordDeliveryAttempt(item.ID, scheduledFor, reminder.StatusPending, "")
	if err != nil {
		// Unique conflict or error -> already dispatched or attempted
		return
	}

	// 2. Format message
	loc, err := time.LoadLocation(item.Timezone)
	if err != nil {
		loc = time.FixedZone("WIB", 7*3600)
	}
	timeDisplay := scheduledFor.In(loc).Format("15:04")
	dateDisplay := scheduledFor.In(loc).Format("02 Jan 2006")

	msgBody := item.Message
	if msgBody == "" {
		msgBody = item.Title
	}

	text := fmt.Sprintf("⏰ *PENGINGAT: %s*\n\n%s\n\n📅 Waktu: %s, %s (%s)",
		item.Title,
		msgBody,
		dateDisplay,
		timeDisplay,
		item.Timezone,
	)

	// 3. Send message via WhatsApp
	var sendErr error
	if s.sender != nil {
		sendErr = s.sender.SendMessage(item.TargetPhone, text)
	} else {
		sendErr = fmt.Errorf("whatsapp sender not initialized")
	}

	// 4. Update delivery log status
	if sendErr != nil {
		log.Printf("[Scheduler] Failed sending reminder #%d (%s): %v\n", item.ID, item.Title, sendErr)
		if logID > 0 {
			_ = s.reminderService.MarkLogStatus(logID, reminder.StatusFailed, sendErr.Error())
		}
	} else {
		log.Printf("[Scheduler] Successfully sent reminder #%d (%s)\n", item.ID, item.Title)
		if logID > 0 {
			_ = s.reminderService.MarkLogStatus(logID, reminder.StatusSent, "")
		}
	}

	// 5. Advance schedule or deactivate if once
	nextRun := reminder.CalculateNextRun(item.ScheduledAt, item.RecurrenceType, item.RecurrenceDays, loc)
	if nextRun.IsZero() || item.RecurrenceType == reminder.RecurrenceOnce {
		item.IsActive = false
	} else {
		// If server was offline for an extended period, advance until nextRun is in the future
		for nextRun.Before(now) {
			candidate := reminder.CalculateNextRun(nextRun, item.RecurrenceType, item.RecurrenceDays, loc)
			if candidate.IsZero() || candidate.Equal(nextRun) {
				break
			}
			nextRun = candidate
		}
		item.ScheduledAt = nextRun
	}

	_ = s.reminderService.UpdateReminder(item)
}
