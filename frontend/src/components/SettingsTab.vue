<script setup lang="ts">
import { ref, onMounted } from 'vue'
import {
  Smartphone,
  Lock,
  Globe,
  Database,
  Download,
  CheckCircle2,
  Save,
  QrCode,
  LogOut
} from 'lucide-vue-next'
import { api } from '../api/client'
import type { WhatsAppConnectionStatus, AppSettings } from '../types'

const props = defineProps<{
  waStatus: WhatsAppConnectionStatus
  connectedPhone?: string
}>()

const emit = defineEmits<{
  (e: 'open-qr'): void
  (e: 'logout-wa'): void
  (e: 'change-pin', oldPin: string, newPin: string): void
}>()

// Settings state
const defaultPhone = ref('')
const defaultTz = ref('Asia/Jakarta')
const savingSettings = ref(false)
const settingsSuccess = ref(false)
const settingsError = ref('')

// Change PIN state
const oldPin = ref('')
const newPin = ref('')
const confirmPin = ref('')
const pinError = ref('')
const pinSuccess = ref(false)
const changingPin = ref(false)

// Backup state
const backupSuccess = ref('')
const backupError = ref('')
const backingUp = ref(false)

async function loadSettings() {
  try {
    const res = await api.get<AppSettings>('/api/settings')
    defaultPhone.value = res.target_phone || ''
    defaultTz.value = res.timezone || 'Asia/Jakarta'
  } catch {
    // Ignore
  }
}

async function saveSettings() {
  savingSettings.value = true
  settingsSuccess.value = false
  settingsError.value = ''
  try {
    await api.post('/api/settings', {
      target_phone: defaultPhone.value,
      timezone: defaultTz.value,
    })
    settingsSuccess.value = true
    setTimeout(() => {
      settingsSuccess.value = false
    }, 3000)
  } catch (e: any) {
    settingsError.value = e.message || 'Gagal menyimpan pengaturan'
  } finally {
    savingSettings.value = false
  }
}

async function handleChangePIN() {
  pinError.value = ''
  pinSuccess.value = false
  if (newPin.value.length < 4 || newPin.value.length > 12) {
    pinError.value = 'PIN baru harus 4-12 digit'
    return
  }
  if (newPin.value !== confirmPin.value) {
    pinError.value = 'Konfirmasi PIN tidak sesuai'
    return
  }

  changingPin.value = true
  try {
    await api.post('/api/auth/change-pin', {
      old_pin: oldPin.value,
      new_pin: newPin.value,
    })
    pinSuccess.value = true
    oldPin.value = ''
    newPin.value = ''
    confirmPin.value = ''
    setTimeout(() => {
      pinSuccess.value = false
    }, 4000)
  } catch (e: any) {
    pinError.value = e.message || 'Gagal mengubah PIN'
  } finally {
    changingPin.value = false
  }
}

async function triggerBackup() {
  backingUp.value = true
  backupSuccess.value = ''
  backupError.value = ''
  try {
    const res = await api.post<{ message: string; filename: string }>('/api/backup')
    backupSuccess.value = `Backup tersimpan: ${res.filename}`
  } catch (e: any) {
    backupError.value = e.message || 'Gagal membuat backup database'
  } finally {
    backingUp.value = false
  }
}

onMounted(() => {
  loadSettings()
})
</script>

<template>
  <div class="space-y-6 max-w-2xl mx-auto pb-20 md:pb-8">
    <!-- WhatsApp Connection Card -->
    <div class="bg-white rounded-2xl border border-slate-200/80 shadow-sm p-5 space-y-4">
      <div class="flex items-center gap-2.5">
        <div class="w-9 h-9 rounded-xl bg-emerald-50 text-emerald-600 flex items-center justify-center">
          <Smartphone class="w-5 h-5" />
        </div>
        <div>
          <h3 class="font-bold text-slate-900 text-sm">Status WhatsApp</h3>
          <p class="text-xs text-slate-500">Koneksi perangkat WhatsApp multi-device</p>
        </div>
      </div>

      <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-3 p-3.5 rounded-xl bg-slate-50 border border-slate-200/70">
        <div>
          <div class="text-xs font-semibold text-slate-700">Akun Terhubung:</div>
          <div class="font-mono text-sm font-bold text-slate-900 mt-0.5">
            {{ connectedPhone ? `+${connectedPhone}` : 'Belum terhubung' }}
          </div>
        </div>
        <div class="flex items-center gap-2">
          <button
            @click="emit('open-qr')"
            class="h-9 px-3.5 rounded-xl bg-blue-600 hover:bg-blue-700 text-white font-semibold text-xs flex items-center gap-1.5 transition shadow-sm"
          >
            <QrCode class="w-3.5 h-3.5" />
            <span>{{ waStatus === 'connected' ? 'Lihat Status / Test' : 'Scan QR Code' }}</span>
          </button>
          <button
            v-if="waStatus === 'connected'"
            @click="emit('logout-wa')"
            class="h-9 px-3.5 rounded-xl bg-rose-50 hover:bg-rose-100 text-rose-600 font-semibold text-xs flex items-center gap-1.5 transition"
          >
            <LogOut class="w-3.5 h-3.5" />
            <span>Unpair</span>
          </button>
        </div>
      </div>
    </div>

    <!-- General Preferences -->
    <div class="bg-white rounded-2xl border border-slate-200/80 shadow-sm p-5 space-y-4">
      <div class="flex items-center gap-2.5">
        <div class="w-9 h-9 rounded-xl bg-blue-50 text-blue-600 flex items-center justify-center">
          <Globe class="w-5 h-5" />
        </div>
        <div>
          <h3 class="font-bold text-slate-900 text-sm">Preferensi Aplikasi</h3>
          <p class="text-xs text-slate-500">Atur zona waktu default dan nomor penerima</p>
        </div>
      </div>

      <div class="space-y-3.5 pt-1">
        <div>
          <label class="block text-xs font-semibold text-slate-700 uppercase tracking-wider mb-1">
            Zona Waktu Default
          </label>
          <select
            v-model="defaultTz"
            class="w-full h-10 px-3.5 text-xs sm:text-sm rounded-xl border border-slate-300 focus:border-blue-600 outline-none bg-white"
          >
            <option value="Asia/Jakarta">WIB — Asia/Jakarta (UTC+7)</option>
            <option value="Asia/Makassar">WITA — Asia/Makassar (UTC+8)</option>
            <option value="Asia/Jayapura">WIT — Asia/Jayapura (UTC+9)</option>
          </select>
        </div>

        <div>
          <label class="block text-xs font-semibold text-slate-700 uppercase tracking-wider mb-1">
            Nomor Penerima Default (Opsional)
          </label>
          <input
            v-model="defaultPhone"
            type="text"
            placeholder="0812xxxx (kosongkan jika sama dengan nomor bot)"
            class="w-full h-10 px-3.5 text-xs sm:text-sm rounded-xl border border-slate-300 focus:border-blue-600 outline-none"
          />
        </div>

        <div v-if="settingsSuccess" class="text-xs text-emerald-600 font-semibold flex items-center gap-1">
          <CheckCircle2 class="w-3.5 h-3.5" />
          <span>Pengaturan berhasil disimpan!</span>
        </div>
        <div v-if="settingsError" class="text-xs text-rose-600 font-semibold">{{ settingsError }}</div>

        <button
          @click="saveSettings"
          :disabled="savingSettings"
          class="h-10 px-5 rounded-xl bg-blue-600 hover:bg-blue-700 text-white font-semibold text-xs flex items-center gap-1.5 transition disabled:opacity-50 shadow-sm"
        >
          <Save class="w-3.5 h-3.5" />
          <span>{{ savingSettings ? 'Menyimpan...' : 'Simpan Preferensi' }}</span>
        </button>
      </div>
    </div>

    <!-- Security: Change PIN -->
    <div class="bg-white rounded-2xl border border-slate-200/80 shadow-sm p-5 space-y-4">
      <div class="flex items-center gap-2.5">
        <div class="w-9 h-9 rounded-xl bg-amber-50 text-amber-600 flex items-center justify-center">
          <Lock class="w-5 h-5" />
        </div>
        <div>
          <h3 class="font-bold text-slate-900 text-sm">Keamanan: Ubah PIN</h3>
          <p class="text-xs text-slate-500">Perbarui PIN perlindungan dashboard web</p>
        </div>
      </div>

      <div class="space-y-3 pt-1">
        <div>
          <label class="block text-xs font-semibold text-slate-700 uppercase tracking-wider mb-1">PIN Lama</label>
          <input
            v-model="oldPin"
            type="password"
            inputmode="numeric"
            placeholder="Masukkan PIN lama"
            class="w-full h-10 px-3.5 text-sm rounded-xl border border-slate-300 focus:border-blue-600 outline-none font-mono"
            maxlength="12"
          />
        </div>

        <div class="grid grid-cols-1 sm:grid-cols-2 gap-3">
          <div>
            <label class="block text-xs font-semibold text-slate-700 uppercase tracking-wider mb-1">PIN Baru</label>
            <input
              v-model="newPin"
              type="password"
              inputmode="numeric"
              placeholder="PIN baru (4-12 digit)"
              class="w-full h-10 px-3.5 text-sm rounded-xl border border-slate-300 focus:border-blue-600 outline-none font-mono"
              maxlength="12"
            />
          </div>
          <div>
            <label class="block text-xs font-semibold text-slate-700 uppercase tracking-wider mb-1">Konfirmasi PIN Baru</label>
            <input
              v-model="confirmPin"
              type="password"
              inputmode="numeric"
              placeholder="Ulangi PIN baru"
              class="w-full h-10 px-3.5 text-sm rounded-xl border border-slate-300 focus:border-blue-600 outline-none font-mono"
              maxlength="12"
            />
          </div>
        </div>

        <div v-if="pinSuccess" class="text-xs text-emerald-600 font-semibold flex items-center gap-1">
          <CheckCircle2 class="w-3.5 h-3.5" />
          <span>PIN berhasil diperbarui!</span>
        </div>
        <div v-if="pinError" class="text-xs text-rose-600 font-semibold">{{ pinError }}</div>

        <button
          @click="handleChangePIN"
          :disabled="changingPin || !oldPin || !newPin"
          class="h-10 px-5 rounded-xl bg-slate-900 hover:bg-slate-800 text-white font-semibold text-xs flex items-center gap-1.5 transition disabled:opacity-50 shadow-sm"
        >
          <Lock class="w-3.5 h-3.5" />
          <span>{{ changingPin ? 'Menyimpan...' : 'Perbarui PIN' }}</span>
        </button>
      </div>
    </div>

    <!-- Database Backup -->
    <div class="bg-white rounded-2xl border border-slate-200/80 shadow-sm p-5 space-y-4">
      <div class="flex items-center gap-2.5">
        <div class="w-9 h-9 rounded-xl bg-purple-50 text-purple-600 flex items-center justify-center">
          <Database class="w-5 h-5" />
        </div>
        <div>
          <h3 class="font-bold text-slate-900 text-sm">Backup Database SQLite</h3>
          <p class="text-xs text-slate-500">Buat cadangan konsisten menggunakan SQLite VACUUM INTO</p>
        </div>
      </div>

      <div class="space-y-3 pt-1">
        <p class="text-xs text-slate-600">
          File backup akan dibuat secara aman di folder <code>data/backups/</code> pada server tanpa mengganggu scheduler yang sedang berjalan.
        </p>

        <div v-if="backupSuccess" class="text-xs text-emerald-600 font-semibold flex items-center gap-1">
          <CheckCircle2 class="w-3.5 h-3.5" />
          <span>{{ backupSuccess }}</span>
        </div>
        <div v-if="backupError" class="text-xs text-rose-600 font-semibold">{{ backupError }}</div>

        <button
          @click="triggerBackup"
          :disabled="backingUp"
          class="h-10 px-5 rounded-xl bg-purple-600 hover:bg-purple-700 text-white font-semibold text-xs flex items-center gap-1.5 transition disabled:opacity-50 shadow-sm"
        >
          <Download class="w-3.5 h-3.5" />
          <span>{{ backingUp ? 'Memproses Backup...' : 'Buat Backup Sekarang' }}</span>
        </button>
      </div>
    </div>
  </div>
</template>
