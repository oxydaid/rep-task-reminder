import { ref, onUnmounted } from 'vue'
import { api } from '../api/client'
import type { WhatsAppStatus, WhatsAppConnectionStatus } from '../types'

const STORAGE_KEY_STATUS = 'wa_status'
const STORAGE_KEY_PHONE = 'wa_connected_phone'

function getCachedStatus(): WhatsAppConnectionStatus {
  if (typeof window === 'undefined') return 'connecting'
  const cached = localStorage.getItem(STORAGE_KEY_STATUS) as WhatsAppConnectionStatus | null
  if (cached === 'connected' || cached === 'connecting' || cached === 'disconnected') {
    return cached
  }
  return 'connecting'
}

function getCachedPhone(): string {
  if (typeof window === 'undefined') return ''
  return localStorage.getItem(STORAGE_KEY_PHONE) || ''
}

export function useWhatsApp() {
  const status = ref<WhatsAppConnectionStatus>(getCachedStatus())
  const connectedPhone = ref<string>(getCachedPhone())
  const qrCode = ref('')
  const targetPhone = ref('')
  const loading = ref(false)
  const isInitialLoading = ref(true)
  const error = ref('')

  let pollInterval: any = null

  const fetchStatus = async () => {
    try {
      const data = await api.get<WhatsAppStatus>('/api/whatsapp/status')
      status.value = data.status
      connectedPhone.value = data.connected_phone || ''
      qrCode.value = data.qr_code_data_url || ''
      targetPhone.value = data.target_phone || ''
      error.value = ''

      if (typeof window !== 'undefined') {
        localStorage.setItem(STORAGE_KEY_STATUS, data.status)
        if (data.connected_phone) {
          localStorage.setItem(STORAGE_KEY_PHONE, data.connected_phone)
        } else if (data.status === 'disconnected') {
          localStorage.removeItem(STORAGE_KEY_PHONE)
        }
      }

      // If status is connecting, ensure polling is active
      if (status.value === 'connecting' && !pollInterval) {
        startPolling()
      } else if (status.value === 'connected' && pollInterval) {
        stopPolling()
      }
    } catch (e: any) {
      error.value = e.message || 'Gagal memuat status WhatsApp'
    } finally {
      isInitialLoading.value = false
    }
  }

  const connect = async () => {
    loading.value = true
    error.value = ''
    try {
      const data = await api.post<WhatsAppStatus>('/api/whatsapp/connect')
      status.value = data.status
      connectedPhone.value = data.connected_phone || ''
      qrCode.value = data.qr_code_data_url || ''

      if (typeof window !== 'undefined') {
        localStorage.setItem(STORAGE_KEY_STATUS, data.status)
        if (data.connected_phone) {
          localStorage.setItem(STORAGE_KEY_PHONE, data.connected_phone)
        }
      }
      startPolling()
    } catch (e: any) {
      error.value = e.message || 'Gagal menghubungkan WhatsApp'
    } finally {
      loading.value = false
      isInitialLoading.value = false
    }
  }

  const disconnect = async () => {
    loading.value = true
    error.value = ''
    try {
      await api.post('/api/whatsapp/disconnect')
      status.value = 'disconnected'
      qrCode.value = ''
      if (typeof window !== 'undefined') {
        localStorage.setItem(STORAGE_KEY_STATUS, 'disconnected')
      }
      stopPolling()
    } catch (e: any) {
      error.value = e.message || 'Gagal memutuskan WhatsApp'
    } finally {
      loading.value = false
    }
  }

  const logoutSession = async () => {
    loading.value = true
    error.value = ''
    try {
      await api.post('/api/whatsapp/logout')
      status.value = 'disconnected'
      connectedPhone.value = ''
      qrCode.value = ''
      if (typeof window !== 'undefined') {
        localStorage.setItem(STORAGE_KEY_STATUS, 'disconnected')
        localStorage.removeItem(STORAGE_KEY_PHONE)
      }
      stopPolling()
    } catch (e: any) {
      error.value = e.message || 'Gagal menghapus sesi WhatsApp'
    } finally {
      loading.value = false
    }
  }

  const sendTestMessage = async (phone: string, message: string) => {
    loading.value = true
    error.value = ''
    try {
      await api.post('/api/whatsapp/test', { phone, message })
      return true
    } catch (e: any) {
      error.value = e.message || 'Gagal mengirim pesan uji coba'
      throw e
    } finally {
      loading.value = false
    }
  }

  const startPolling = () => {
    if (pollInterval) return
    pollInterval = setInterval(async () => {
      try {
        const data = await api.get<WhatsAppStatus>('/api/whatsapp/status')
        status.value = data.status
        connectedPhone.value = data.connected_phone || ''
        qrCode.value = data.qr_code_data_url || ''

        if (typeof window !== 'undefined') {
          localStorage.setItem(STORAGE_KEY_STATUS, data.status)
          if (data.connected_phone) {
            localStorage.setItem(STORAGE_KEY_PHONE, data.connected_phone)
          } else if (data.status === 'disconnected') {
            localStorage.removeItem(STORAGE_KEY_PHONE)
          }
        }

        if (data.status === 'connected') {
          stopPolling()
        }
      } catch {
        // Polling error silently handled
      }
    }, 2500)
  }

  const stopPolling = () => {
    if (pollInterval) {
      clearInterval(pollInterval)
      pollInterval = null
    }
  }

  onUnmounted(() => {
    stopPolling()
  })

  return {
    status,
    connectedPhone,
    qrCode,
    targetPhone,
    loading,
    isInitialLoading,
    error,
    fetchStatus,
    connect,
    disconnect,
    logoutSession,
    sendTestMessage,
    startPolling,
    stopPolling,
  }
}
