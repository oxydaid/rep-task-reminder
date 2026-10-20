import { ref, onMounted, onUnmounted } from 'vue'
import { api, getAuthToken, setAuthToken } from '../api/client'
import type { AuthStatus } from '../types'

const STORAGE_KEY_SETUP_REQUIRED = 'auth_setup_required_cache'

export function useAuth() {
  const isAuthenticated = ref(!!getAuthToken())
  const setupRequired = ref(
    typeof window !== 'undefined' ? localStorage.getItem(STORAGE_KEY_SETUP_REQUIRED) === 'true' : false
  )
  const isLocked = ref(false)
  const lockRemaining = ref(0)
  const loading = ref(false)
  const error = ref('')

  let lockTimer: any = null

  const startLockTimer = (seconds: number) => {
    lockRemaining.value = seconds
    isLocked.value = true
    if (lockTimer) clearInterval(lockTimer)
    lockTimer = setInterval(() => {
      if (lockRemaining.value > 1) {
        lockRemaining.value -= 1
      } else {
        lockRemaining.value = 0
        isLocked.value = false
        clearInterval(lockTimer)
      }
    }, 1000)
  }

  const checkStatus = async () => {
    loading.value = true
    error.value = ''
    try {
      const data = await api.get<AuthStatus>('/api/auth/status')
      setupRequired.value = data.setup_required
      if (typeof window !== 'undefined') {
        localStorage.setItem(STORAGE_KEY_SETUP_REQUIRED, String(data.setup_required))
      }
      if (data.locked && data.lock_remaining_seconds > 0) {
        startLockTimer(data.lock_remaining_seconds)
      } else {
        isLocked.value = false
        lockRemaining.value = 0
      }
    } catch (e: any) {
      error.value = e.message || 'Gagal memeriksa status keamanan'
    } finally {
      loading.value = false
    }
  }

  const login = async (pin: string) => {
    loading.value = true
    error.value = ''
    try {
      const res = await api.post<{ token: string; message: string }>('/api/auth/login', { pin })
      if (res.token) {
        setAuthToken(res.token)
        isAuthenticated.value = true
        return true
      }
      return false
    } catch (e: any) {
      error.value = e.message || 'Login gagal'
      await checkStatus()
      return false
    } finally {
      loading.value = false
    }
  }

  const setupPIN = async (pin: string) => {
    loading.value = true
    error.value = ''
    try {
      const res = await api.post<{ token: string; message: string }>('/api/auth/setup', { pin })
      if (res.token) {
        setAuthToken(res.token)
        isAuthenticated.value = true
        setupRequired.value = false
        if (typeof window !== 'undefined') {
          localStorage.setItem(STORAGE_KEY_SETUP_REQUIRED, 'false')
        }
        return true
      }
      return false
    } catch (e: any) {
      error.value = e.message || 'Setup PIN gagal'
      return false
    } finally {
      loading.value = false
    }
  }

  const changePIN = async (oldPin: string, newPin: string) => {
    loading.value = true
    error.value = ''
    try {
      await api.post('/api/auth/change-pin', { old_pin: oldPin, new_pin: newPin })
      return true
    } catch (e: any) {
      error.value = e.message || 'Gagal mengubah PIN'
      throw e
    } finally {
      loading.value = false
    }
  }

  const logout = async () => {
    try {
      await api.post('/api/auth/logout')
    } catch {
      // Ignore
    } finally {
      setAuthToken('')
      isAuthenticated.value = false
    }
  }

  onMounted(() => {
    const onAuthExpired = () => {
      isAuthenticated.value = false
    }
    window.addEventListener('auth:expired', onAuthExpired)
    onUnmounted(() => {
      window.removeEventListener('auth:expired', onAuthExpired)
      if (lockTimer) clearInterval(lockTimer)
    })
  })

  return {
    isAuthenticated,
    setupRequired,
    isLocked,
    lockRemaining,
    loading,
    error,
    checkStatus,
    login,
    setupPIN,
    changePIN,
    logout,
  }
}
