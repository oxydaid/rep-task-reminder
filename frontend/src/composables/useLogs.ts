import { ref } from 'vue'
import { api } from '../api/client'
import type { DeliveryLog } from '../types'

export function useLogs() {
  const logs = ref<DeliveryLog[]>([])
  const total = ref(0)
  const loading = ref(true)
  const error = ref('')

  const fetchLogs = async (limit: number = 20, offset: number = 0) => {
    loading.value = true
    error.value = ''
    try {
      const res = await api.get<{ data: DeliveryLog[]; total: number }>(`/api/logs?limit=${limit}&offset=${offset}`)
      logs.value = res.data || []
      total.value = res.total || 0
    } catch (e: any) {
      error.value = e.message || 'Gagal memuat riwayat pengiriman'
    } finally {
      loading.value = false
    }
  }

  return {
    logs,
    total,
    loading,
    error,
    fetchLogs,
  }
}
