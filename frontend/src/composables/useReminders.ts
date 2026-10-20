import { ref, watch, toRef, toValue, type MaybeRefOrGetter } from 'vue'
import { api } from '../api/client'
import type { Reminder, Stats } from '../types'

export interface UseRemindersOptions {
  statusFilter?: MaybeRefOrGetter<string>
  immediate?: boolean
}

export function useReminders(options: UseRemindersOptions = {}) {
  const reminders = ref<Reminder[]>([])
  const upcoming = ref<Reminder | null>(null)
  const stats = ref<Stats>({
    total_reminders: 0,
    active_reminders: 0,
    inactive_reminders: 0,
    today_reminders: 0,
    total_sent: 0,
    total_failed: 0,
  })
  const loading = ref(false)
  const error = ref('')

  const fetchReminders = async (status?: MaybeRefOrGetter<string>) => {
    loading.value = true
    error.value = ''
    const resolvedStatus = status !== undefined ? toValue(status) : (options.statusFilter ? toValue(options.statusFilter) : 'all')
    try {
      const data = await api.get<Reminder[]>(`/api/reminders?status=${resolvedStatus}`)
      reminders.value = data || []
    } catch (e: any) {
      error.value = e.message || 'Gagal memuat reminder'
    } finally {
      loading.value = false
    }
  }

  const fetchUpcoming = async () => {
    try {
      const data = await api.get<Reminder | null>('/api/reminders/upcoming')
      upcoming.value = data
    } catch {
      upcoming.value = null
    }
  }

  const fetchStats = async () => {
    try {
      const data = await api.get<Stats>('/api/stats')
      if (data) stats.value = data
    } catch {
      // Ignore
    }
  }

  const createReminder = async (payload: Partial<Reminder>) => {
    loading.value = true
    error.value = ''
    try {
      const created = await api.post<Reminder>('/api/reminders', payload)
      await Promise.all([fetchReminders(), fetchUpcoming(), fetchStats()])
      return created
    } catch (e: any) {
      error.value = e.message || 'Gagal membuat reminder'
      throw e
    } finally {
      loading.value = false
    }
  }

  const updateReminder = async (id: number, payload: Partial<Reminder>) => {
    loading.value = true
    error.value = ''
    try {
      const updated = await api.put<Reminder>(`/api/reminders/${id}`, payload)
      await Promise.all([fetchReminders(), fetchUpcoming(), fetchStats()])
      return updated
    } catch (e: any) {
      error.value = e.message || 'Gagal memperbarui reminder'
      throw e
    } finally {
      loading.value = false
    }
  }

  const deleteReminder = async (id: number) => {
    loading.value = true
    error.value = ''
    try {
      await api.delete(`/api/reminders/${id}`)
      await Promise.all([fetchReminders(), fetchUpcoming(), fetchStats()])
      return true
    } catch (e: any) {
      error.value = e.message || 'Gagal menghapus reminder'
      throw e
    } finally {
      loading.value = false
    }
  }

  const toggleReminder = async (id: number, isActive: boolean) => {
    try {
      await api.patch(`/api/reminders/${id}/toggle`, { is_active: isActive })
      const item = reminders.value.find((r) => r.id === id)
      if (item) item.is_active = isActive
      await Promise.all([fetchUpcoming(), fetchStats()])
    } catch (e: any) {
      error.value = e.message || 'Gagal mengubah status reminder'
      await fetchReminders()
    }
  }

  // Reactive listener for statusFilter if provided
  if (options.statusFilter !== undefined) {
    watch(
      toRef(options.statusFilter),
      (newStatus) => {
        fetchReminders(toValue(newStatus))
      },
      { immediate: options.immediate ?? false }
    )
  }

  return {
    reminders,
    upcoming,
    stats,
    loading,
    error,
    fetchReminders,
    fetchUpcoming,
    fetchStats,
    createReminder,
    updateReminder,
    deleteReminder,
    toggleReminder,
  }
}
