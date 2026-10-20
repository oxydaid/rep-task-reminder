package reminder

import "time"

type RecurrenceType string

const (
	RecurrenceOnce         RecurrenceType = "once"
	RecurrenceDaily        RecurrenceType = "daily"
	RecurrenceSpecificDays RecurrenceType = "specific_days"
	RecurrenceWeekly       RecurrenceType = "weekly"
)

type Reminder struct {
	ID             int64          `json:"id"`
	Title          string         `json:"title"`
	Message        string         `json:"message"`
	ScheduledAt    time.Time      `json:"scheduled_at"`
	RecurrenceType RecurrenceType `json:"recurrence_type"`
	RecurrenceDays string         `json:"recurrence_days"` // e.g. "1,3,5" (1=Mon ... 7=Sun)
	Timezone       string         `json:"timezone"`
	TargetPhone    string         `json:"target_phone"`
	IsActive       bool           `json:"is_active"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
}

type DeliveryStatus string

const (
	StatusPending DeliveryStatus = "pending"
	StatusSent    DeliveryStatus = "sent"
	StatusFailed  DeliveryStatus = "failed"
)

type DeliveryLog struct {
	ID           int64          `json:"id"`
	ReminderID   int64          `json:"reminder_id"`
	ReminderTitle string        `json:"reminder_title,omitempty"`
	ScheduledFor time.Time      `json:"scheduled_for"`
	Status       DeliveryStatus `json:"status"`
	Attempts     int            `json:"attempts"`
	SentAt       *time.Time     `json:"sent_at,omitempty"`
	ErrorMessage string         `json:"error_message,omitempty"`
	CreatedAt    time.Time      `json:"created_at"`
}

type Stats struct {
	TotalReminders    int `json:"total_reminders"`
	ActiveReminders   int `json:"active_reminders"`
	InactiveReminders int `json:"inactive_reminders"`
	TodayReminders    int `json:"today_reminders"`
	TotalSent         int `json:"total_sent"`
	TotalFailed       int `json:"total_failed"`
}
