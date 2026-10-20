import { ref, onMounted, onUnmounted } from 'vue'
import { api } from '../api/client'
import type { WhatsAppStatus, WhatsAppConnectionStatus } from '../types'

export function useWhatsApp() {
  const status = ref<WhatsAppConnectionStatus>('disconnected')
  const connectedPhone = ref('')
  const qrCode = ref('')
  const targetPhone = ref('')
  const loading = ref(false)
  const error = ref('')

  let pollInterval: any = null

  const fetchStatus = async () => {
    try {
      const data = await api.get<WhatsAppStatus>('/api/whatsapp/status')
      status.value = data.status
      connectedPhone.value = data.connected_phone
      qrCode.value = data.qr_code_data_url || ''
      targetPhone.value = data.target_phone || ''
      error.value = ''

      // If status is connecting, ensure polling is active
      if (status.value === 'connecting' && !pollInterval) {
        startPolling()
      } else if (status.value === 'connected' && pollInterval) {
        stopPolling()
      }
    } catch (e: any) {
      error.value = e.message || 'Gagal memuat status WhatsApp'
    }
  }

  const connect = async () => {
    loading.value = true
    error.value = ''
    try {
      const data = await api.post<WhatsAppStatus>('/api/whatsapp/connect')
      status.value = data.status
      connectedPhone.value = data.connected_phone
      qrCode.value = data.qr_code_data_url || ''
      startPolling()
    } catch (e: any) {
      error.value = e.message || 'Gagal menghubungkan WhatsApp'
    } finally {
      loading.value = false
    }
  }

  const disconnect = async () => {
    loading.value = true
    error.value = ''
    try {
      await api.post('/api/whatsapp/disconnect')
      status.value = 'disconnected'
      qrCode.value = ''
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
        connectedPhone.value = data.connected_phone
        qrCode.value = data.qr_code_data_url || ''
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

  onMounted(() => {
    onUnmounted(() => {
      stopPolling()
    })
  })

  return {
    status,
    connectedPhone,
    qrCode,
    targetPhone,
    loading,
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
