<script setup lang="ts">
import { computed } from 'vue'
import { useCountdown } from '../composables/useCountdown'
import {
  Bell,
  Clock,
  Calendar,
  CheckCircle2,
  AlertCircle,
  Plus,
  WifiOff,
  QrCode,
  ArrowRight
} from 'lucide-vue-next'
import type { Reminder, Stats, WhatsAppConnectionStatus } from '../types'

const props = defineProps<{
  upcoming: Reminder | null
  stats: Stats
  waStatus: WhatsAppConnectionStatus
  connectedPhone?: string
  loading: boolean
  isWaLoading?: boolean
  isUpcomingLoading?: boolean
  isStatsLoading?: boolean
}>()

const emit = defineEmits<{
  (e: 'create'): void
  (e: 'open-qr'): void
  (e: 'view-reminders'): void
  (e: 'toggle-reminder', id: number, active: boolean): void
}>()

const { formattedCountdown } = useCountdown(() => props.upcoming?.scheduled_at)

const formattedUpcomingTime = computed(() => {
  if (!props.upcoming || !props.upcoming.scheduled_at) return ''
  const d = new Date(props.upcoming.scheduled_at)
  return d.toLocaleString('id-ID', {
    weekday: 'short',
    day: 'numeric',
    month: 'short',
    hour: '2-digit',
    minute: '2-digit',
  })
})

const recurrenceLabel = computed(() => {
  if (!props.upcoming) return ''
  switch (props.upcoming.recurrence_type) {
    case 'daily': return 'Setiap Hari'
    case 'weekly': return 'Mingguan'
    case 'specific_days': return `Hari: ${props.upcoming.recurrence_days}`
    case 'once': return 'Sekali'
    default: return ''
  }
})
</script>

<template>
  <div class="space-y-6 pb-20 md:pb-8">
    <!-- WhatsApp Alert Banner if Disconnected (hidden while verifying to prevent FOUC) -->
    <Transition
      enter-active-class="transition-all duration-300 ease-out"
      enter-from-class="opacity-0 -translate-y-2 max-h-0 overflow-hidden"
      enter-to-class="opacity-100 translate-y-0 max-h-40"
      leave-active-class="transition-all duration-200 ease-in"
      leave-from-class="opacity-100 translate-y-0 max-h-40"
      leave-to-class="opacity-0 -translate-y-2 max-h-0 overflow-hidden"
    >
      <div
        v-if="waStatus === 'disconnected' && !isWaLoading"
        class="p-4 rounded-2xl bg-amber-500/10 border border-amber-300/60 flex flex-col sm:flex-row items-start sm:items-center justify-between gap-3 shadow-sm"
      >
        <div class="flex items-center gap-3">
          <div class="w-10 h-10 rounded-xl bg-amber-500/20 text-amber-700 flex items-center justify-center flex-shrink-0">
            <WifiOff class="w-5 h-5" />
          </div>
          <div>
            <div class="font-bold text-slate-900 text-sm">WhatsApp Belum Terhubung</div>
            <div class="text-xs text-slate-600">Scan QR Code agar reminder dapat dikirimkan tepat waktu.</div>
          </div>
        </div>
        <button
          @click="emit('open-qr')"
          class="px-4 py-2 bg-amber-600 hover:bg-amber-700 text-white text-xs font-semibold rounded-xl flex items-center gap-1.5 transition shadow-sm flex-shrink-0"
        >
          <QrCode class="w-4 h-4" />
          <span>Hubungkan WhatsApp</span>
        </button>
      </div>
    </Transition>

    <!-- Hero Card: Next Upcoming Reminder -->
    <div class="relative overflow-hidden rounded-3xl bg-gradient-to-br from-blue-600 to-indigo-700 text-white p-6 sm:p-8 shadow-xl shadow-blue-500/15">
      <!-- Background Graphic elements -->
      <div class="absolute -right-8 -bottom-8 w-44 h-44 rounded-full bg-white/5 blur-xl pointer-events-none" />
      <div class="absolute right-12 top-6 w-20 h-20 rounded-full bg-white/10 blur-md pointer-events-none" />

      <div class="relative z-10 flex flex-col md:flex-row md:items-center justify-between gap-6">
        <!-- Skeleton State when Upcoming is loading -->
        <div v-if="isUpcomingLoading" class="space-y-4 w-full max-w-lg animate-pulse py-1">
          <div class="inline-flex items-center gap-1.5 px-3 py-1 rounded-full bg-white/15 text-blue-50 text-xs font-medium backdrop-blur-sm">
            <Clock class="w-3.5 h-3.5 animate-spin" />
            <span>Memeriksa Jadwal...</span>
          </div>
          <div class="space-y-2">
            <div class="h-8 bg-white/25 rounded-xl w-3/4" />
            <div class="h-4 bg-white/15 rounded-lg w-1/2" />
          </div>
          <div class="flex items-center gap-2 pt-1">
            <div class="h-6 w-32 bg-white/20 rounded-lg" />
            <div class="h-6 w-24 bg-white/15 rounded-lg" />
          </div>
        </div>

        <!-- Loaded State: Has upcoming reminder -->
        <div v-else-if="upcoming" class="space-y-3">
          <div class="inline-flex items-center gap-1.5 px-3 py-1 rounded-full bg-white/15 text-blue-50 text-xs font-medium backdrop-blur-sm">
            <Clock class="w-3.5 h-3.5" />
            <span>Reminder Berikutnya</span>
          </div>

          <div>
            <h2 class="text-2xl sm:text-3xl font-bold tracking-tight text-white mb-1">
              {{ upcoming.title }}
            </h2>
            <p v-if="upcoming.message && upcoming.message !== upcoming.title" class="text-blue-100 text-sm max-w-lg line-clamp-2">
              {{ upcoming.message }}
            </p>
            <div class="flex flex-wrap items-center gap-3 mt-3 text-xs text-blue-100">
              <span class="inline-flex items-center gap-1 font-semibold text-white bg-blue-500/40 px-2.5 py-1 rounded-lg">
                <Calendar class="w-3.5 h-3.5" />
                {{ formattedUpcomingTime }} {{ upcoming.timezone }}
              </span>
              <span class="px-2.5 py-1 rounded-lg bg-white/10">
                🔄 {{ recurrenceLabel }}
              </span>
              <span v-if="formattedCountdown" class="px-2.5 py-1 rounded-lg bg-white/20 font-medium">
                ⏳ {{ formattedCountdown }}
              </span>
            </div>
          </div>
        </div>

        <!-- Loaded State: No upcoming reminder -->
        <div v-else class="py-2">
          <div class="inline-flex items-center gap-1.5 px-3 py-1 rounded-full bg-white/15 text-blue-50 text-xs font-medium backdrop-blur-sm mb-2">
            <Clock class="w-3.5 h-3.5" />
            <span>Reminder Berikutnya</span>
          </div>
          <h2 class="text-xl sm:text-2xl font-bold text-white mb-1">
            Tidak Ada Reminder Mendatang
          </h2>
          <p class="text-blue-100 text-xs sm:text-sm">
            Semua jadwal telah selesai atau belum ada reminder yang aktif.
          </p>
        </div>

        <!-- Hero Actions -->
        <div class="flex items-center gap-3 flex-shrink-0">
          <button
            @click="emit('create')"
            class="px-5 py-3 rounded-2xl bg-white hover:bg-blue-50 text-blue-700 font-bold text-xs sm:text-sm flex items-center gap-2 transition shadow-lg shadow-black/10 active:scale-95"
          >
            <Plus class="w-4 h-4 stroke-[3]" />
            <span>Buat Reminder</span>
          </button>
        </div>
      </div>
    </div>

    <!-- Quick Stats Grid -->
    <div class="grid grid-cols-2 lg:grid-cols-4 gap-3 sm:gap-4">
      <!-- Skeleton Stats Grid -->
      <template v-if="isStatsLoading">
        <div
          v-for="i in 4"
          :key="i"
          class="bg-white p-4 sm:p-5 rounded-2xl border border-slate-200/80 shadow-sm flex items-center gap-3.5 animate-pulse"
        >
          <div class="w-11 h-11 rounded-xl bg-slate-100 flex-shrink-0" />
          <div class="space-y-2 flex-1">
            <div class="h-6 bg-slate-200 rounded-md w-12" />
            <div class="h-3 bg-slate-100 rounded w-24" />
          </div>
        </div>
      </template>

      <!-- Loaded Stats Grid -->
      <template v-else>
        <!-- Active Reminders -->
        <div class="bg-white p-4 sm:p-5 rounded-2xl border border-slate-200/80 shadow-sm flex items-center gap-3.5">
          <div class="w-11 h-11 rounded-xl bg-blue-50 text-blue-600 flex items-center justify-center flex-shrink-0">
            <Bell class="w-5 h-5" />
          </div>
          <div>
            <div class="text-2xl font-bold text-slate-900 leading-none">{{ stats.active_reminders }}</div>
            <div class="text-xs text-slate-500 mt-1 font-medium">Reminder Aktif</div>
          </div>
        </div>

        <!-- Today's Schedule -->
        <div class="bg-white p-4 sm:p-5 rounded-2xl border border-slate-200/80 shadow-sm flex items-center gap-3.5">
          <div class="w-11 h-11 rounded-xl bg-amber-50 text-amber-600 flex items-center justify-center flex-shrink-0">
            <Calendar class="w-5 h-5" />
          </div>
          <div>
            <div class="text-2xl font-bold text-slate-900 leading-none">{{ stats.today_reminders }}</div>
            <div class="text-xs text-slate-500 mt-1 font-medium">Jadwal Hari Ini</div>
          </div>
        </div>

        <!-- Sent -->
        <div class="bg-white p-4 sm:p-5 rounded-2xl border border-slate-200/80 shadow-sm flex items-center gap-3.5">
          <div class="w-11 h-11 rounded-xl bg-emerald-50 text-emerald-600 flex items-center justify-center flex-shrink-0">
            <CheckCircle2 class="w-5 h-5" />
          </div>
          <div>
            <div class="text-2xl font-bold text-slate-900 leading-none">{{ stats.total_sent }}</div>
            <div class="text-xs text-slate-500 mt-1 font-medium">Berhasil Terkirim</div>
          </div>
        </div>

        <!-- Failed -->
        <div class="bg-white p-4 sm:p-5 rounded-2xl border border-slate-200/80 shadow-sm flex items-center gap-3.5">
          <div class="w-11 h-11 rounded-xl bg-rose-50 text-rose-600 flex items-center justify-center flex-shrink-0">
            <AlertCircle class="w-5 h-5" />
          </div>
          <div>
            <div class="text-2xl font-bold text-slate-900 leading-none">{{ stats.total_failed }}</div>
            <div class="text-xs text-slate-500 mt-1 font-medium">Pengiriman Gagal</div>
          </div>
        </div>
      </template>
    </div>

    <!-- Overview Section: WhatsApp Bot Guide & Action Card -->
    <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
      <!-- Bot Chat Instructions -->
      <div class="bg-white p-5 rounded-2xl border border-slate-200/80 shadow-sm space-y-3">
        <div class="flex items-center justify-between">
          <h3 class="font-bold text-slate-900 text-sm flex items-center gap-2">
            <span>💬 Kelola via WhatsApp Chat</span>
          </h3>
          <span class="text-[11px] font-semibold px-2 py-0.5 rounded bg-emerald-50 text-emerald-700">Bot Aktif</span>
        </div>
        <p class="text-xs text-slate-600 leading-relaxed">
          Kirim pesan langsung ke akun WhatsApp Anda untuk mengelola alarm reminder secara instan:
        </p>
        <div class="p-3 bg-slate-50 rounded-xl border border-slate-200/60 font-mono text-xs text-slate-800 space-y-1.5">
          <div><b class="text-blue-600">!help</b> - Panduan lengkap (atau <b class="text-blue-600">!ping</b>)</div>
          <div><b class="text-blue-600">!buat Absen | 08:00 | setiap hari</b></div>
          <div><b class="text-blue-600">!buat Minum obat | 12:00 | harian</b></div>
          <div><b class="text-blue-600">!daftar</b> - Lihat daftar reminder</div>
          <div><b class="text-blue-600">!edit 1 jam 13:00</b> - Ubah jam</div>
        </div>
      </div>

      <!-- Quick Link to All Reminders -->
      <div class="bg-white p-5 rounded-2xl border border-slate-200/80 shadow-sm flex flex-col justify-between space-y-4">
        <div>
          <h3 class="font-bold text-slate-900 text-sm mb-1">Daftar Semua Reminder</h3>
          <p class="text-xs text-slate-600 leading-relaxed">
            Kelola pengulangan harian, hari tertentu dalam seminggu, nomor tujuan khusus, dan status aktif.
          </p>
        </div>
        <button
          @click="emit('view-reminders')"
          class="w-full h-11 rounded-xl bg-slate-100 hover:bg-slate-200 text-slate-800 font-semibold text-xs flex items-center justify-center gap-2 transition"
        >
          <span>Buka Tab Reminder {{ isStatsLoading ? '' : `(${stats.total_reminders})` }}</span>
          <ArrowRight class="w-4 h-4" />
        </button>
      </div>
    </div>
  </div>
</template>
