<script setup lang="ts">
import { ref } from 'vue'
import {
  History,
  CheckCircle2,
  Clock,
  AlertCircle,
  RefreshCw,
  ChevronLeft,
  ChevronRight
} from 'lucide-vue-next'
import type { DeliveryLog } from '../types'

const props = defineProps<{
  logs: DeliveryLog[]
  total: number
  loading: boolean
}>()

const emit = defineEmits<{
  (e: 'refresh'): void
  (e: 'page-change', offset: number): void
}>()

const pageSize = 20
const currentOffset = ref(0)

function formatDateTime(isoStr?: string) {
  if (!isoStr) return '-'
  try {
    const d = new Date(isoStr)
    return d.toLocaleString('id-ID', {
      day: '2-digit',
      month: 'short',
      year: 'numeric',
      hour: '2-digit',
      minute: '2-digit',
    })
  } catch {
    return isoStr
  }
}

function prevPage() {
  if (currentOffset.value >= pageSize) {
    currentOffset.value -= pageSize
    emit('page-change', currentOffset.value)
  }
}

function nextPage() {
  if (currentOffset.value + pageSize < props.total) {
    currentOffset.value += pageSize
    emit('page-change', currentOffset.value)
  }
}
</script>

<template>
  <div class="space-y-4 pb-20 md:pb-8">
    <!-- Header Controls -->
    <div class="flex items-center justify-between">
      <div>
        <h3 class="text-base font-bold text-slate-900">Riwayat Pengiriman Pesan</h3>
        <p class="text-xs text-slate-500">Catatan pengiriman pesan WhatsApp oleh scheduler otomatis</p>
      </div>

      <button
        @click="emit('refresh')"
        :disabled="loading"
        class="h-9 px-3.5 rounded-xl border border-slate-200 bg-white hover:bg-slate-50 text-slate-700 text-xs font-semibold flex items-center gap-1.5 transition shadow-sm"
      >
        <RefreshCw class="w-3.5 h-3.5" :class="{ 'animate-spin': loading }" />
        <span>Segarkan</span>
      </button>
    </div>

    <!-- Skeleton Loading State -->
    <div v-if="loading" class="bg-white rounded-2xl border border-slate-200/80 shadow-sm overflow-hidden p-5 space-y-4 animate-pulse">
      <div v-for="i in 5" :key="i" class="flex items-center justify-between py-2 border-b border-slate-100 last:border-0">
        <div class="space-y-1.5 flex-1">
          <div class="h-4 bg-slate-200 rounded w-1/4" />
          <div class="h-3 bg-slate-100 rounded w-1/3" />
        </div>
        <div class="h-6 w-20 bg-slate-100 rounded-full" />
      </div>
    </div>

    <!-- Empty State -->
    <div
      v-else-if="logs.length === 0"
      class="text-center py-16 bg-white rounded-2xl border border-slate-200/80 p-8 shadow-sm space-y-3"
    >
      <div class="w-14 h-14 mx-auto rounded-2xl bg-slate-100 text-slate-500 flex items-center justify-center">
        <History class="w-7 h-7" />
      </div>
      <h3 class="text-base font-bold text-slate-800">Belum Ada Riwayat Pengiriman</h3>
      <p class="text-xs text-slate-500 max-w-sm mx-auto">
        Ketika jadwal reminder tiba, status pengiriman dan catatan error akan tersimpan di sini.
      </p>
    </div>

    <!-- Logs List (Cards for Mobile, Table for Desktop) -->
    <div v-else class="bg-white rounded-2xl border border-slate-200/80 shadow-sm overflow-hidden">
      <div class="overflow-x-auto">
        <table class="w-full text-left text-xs">
          <thead class="bg-slate-50 text-slate-500 font-semibold border-b border-slate-100 uppercase tracking-wider">
            <tr>
              <th class="py-3 px-4">Reminder</th>
              <th class="py-3 px-4">Jadwal Target</th>
              <th class="py-3 px-4">Status</th>
              <th class="py-3 px-4">Waktu Dikirim</th>
              <th class="py-3 px-4">Percobaan</th>
              <th class="py-3 px-4">Catatan / Error</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-slate-100">
            <tr v-for="log in logs" :key="log.id" class="hover:bg-slate-50/70 transition">
              <td class="py-3 px-4 font-semibold text-slate-900">
                {{ log.reminder_title || `Reminder #${log.reminder_id}` }}
              </td>
              <td class="py-3 px-4 text-slate-600 font-mono">
                {{ formatDateTime(log.scheduled_for) }}
              </td>
              <td class="py-3 px-4">
                <span
                  class="inline-flex items-center gap-1 px-2.5 py-0.5 rounded-full font-medium"
                  :class="{
                    'bg-emerald-50 text-emerald-700': log.status === 'sent',
                    'bg-amber-50 text-amber-700': log.status === 'pending',
                    'bg-rose-50 text-rose-700': log.status === 'failed'
                  }"
                >
                  <CheckCircle2 v-if="log.status === 'sent'" class="w-3 h-3" />
                  <Clock v-else-if="log.status === 'pending'" class="w-3 h-3" />
                  <AlertCircle v-else class="w-3 h-3" />
                  <span>{{ log.status === 'sent' ? 'Terkirim' : log.status === 'pending' ? 'Menunggu' : 'Gagal' }}</span>
                </span>
              </td>
              <td class="py-3 px-4 text-slate-600 font-mono">
                {{ formatDateTime(log.sent_at) }}
              </td>
              <td class="py-3 px-4 text-slate-600 font-mono text-center">
                {{ log.attempts }}x
              </td>
              <td class="py-3 px-4 text-slate-500 max-w-xs truncate" :title="log.error_message">
                <span v-if="log.error_message" class="text-rose-600">{{ log.error_message }}</span>
                <span v-else class="text-slate-400">-</span>
              </td>
            </tr>
          </tbody>
        </table>
      </div>

      <!-- Pagination Footer -->
      <div class="px-4 py-3 border-t border-slate-100 bg-slate-50/50 flex items-center justify-between text-xs text-slate-500">
        <div>
          Menampilkan <span class="font-bold text-slate-700">{{ logs.length }}</span> dari <span class="font-bold text-slate-700">{{ total }}</span> log
        </div>
        <div class="flex items-center gap-2">
          <button
            @click="prevPage"
            :disabled="currentOffset <= 0"
            class="h-7 px-2.5 rounded-lg border border-slate-200 bg-white hover:bg-slate-100 disabled:opacity-40 transition flex items-center gap-1"
          >
            <ChevronLeft class="w-3.5 h-3.5" />
            <span>Sebelumnya</span>
          </button>
          <button
            @click="nextPage"
            :disabled="currentOffset + pageSize >= total"
            class="h-7 px-2.5 rounded-lg border border-slate-200 bg-white hover:bg-slate-100 disabled:opacity-40 transition flex items-center gap-1"
          >
            <span>Berikutnya</span>
            <ChevronRight class="w-3.5 h-3.5" />
          </button>
        </div>
      </div>
    </div>
  </div>
</template>
