export type RecurrenceType = 'once' | 'daily' | 'specific_days' | 'weekly'

export interface Reminder {
  id: number
  title: string
  message: string
  scheduled_at: string
  recurrence_type: RecurrenceType
  recurrence_days: string
  timezone: string
  target_phone: string
  is_active: boolean
  created_at: string
  updated_at: string
}

export type DeliveryStatus = 'pending' | 'sent' | 'failed'

export interface DeliveryLog {
  id: number
  reminder_id: number
  reminder_title?: string
  scheduled_for: string
  status: DeliveryStatus
  attempts: number
  sent_at?: string
  error_message?: string
  created_at: string
}

export interface Stats {
  total_reminders: number
  active_reminders: number
  inactive_reminders: number
  today_reminders: number
  total_sent: number
  total_failed: number
}

export type WhatsAppConnectionStatus = 'connected' | 'connecting' | 'disconnected'

export interface WhatsAppStatus {
  status: WhatsAppConnectionStatus
  connected_phone: string
  qr_code_data_url?: string
  target_phone: string
}

export interface AuthStatus {
  setup_required: boolean
  locked: boolean
  lock_remaining_seconds: number
}

export interface AppSettings {
  target_phone: string
  timezone: string
}
