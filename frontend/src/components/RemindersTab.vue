<script setup lang="ts">
import { ref, computed } from 'vue'
import {
  Plus,
  Clock,
  Edit3,
  Trash2,
  Smartphone,
  Search
} from 'lucide-vue-next'
import type { Reminder } from '../types'

const props = defineProps<{
  reminders: Reminder[]
  loading: boolean
}>()

const emit = defineEmits<{
  (e: 'create'): void
  (e: 'edit', item: Reminder): void
  (e: 'delete', id: number): void
  (e: 'toggle', id: number, active: boolean): void
}>()

const filter = ref<'all' | 'active' | 'inactive'>('all')
const searchQuery = ref('')
const confirmDeleteId = ref<number | null>(null)

const filteredReminders = computed(() => {
  return props.reminders.filter((r) => {
    if (filter.value === 'active' && !r.is_active) return false
    if (filter.value === 'inactive' && r.is_active) return false
    if (searchQuery.value.trim()) {
      const q = searchQuery.value.toLowerCase()
      const titleMatch = r.title.toLowerCase().includes(q)
      const msgMatch = r.message.toLowerCase().includes(q)
      return titleMatch || msgMatch
    }
    return true
  })
})

function formatTime(isoStr: string, _tz?: string) {
  if (!isoStr) return ''
  try {
    const d = new Date(isoStr)
    return d.toLocaleString('id-ID', {
      weekday: 'short',
      day: 'numeric',
      month: 'short',
      hour: '2-digit',
      minute: '2-digit',
    })
  } catch {
    return isoStr
  }
}

function getRecurrenceLabel(type: string, days?: string) {
  switch (type) {
    case 'daily': return 'Setiap Hari'
    case 'weekly': return 'Mingguan'
    case 'specific_days': return `Hari: ${days || '-'}`
    case 'once': return 'Sekali Saja'
    default: return type
  }
}

function confirmDelete(id: number) {
  confirmDeleteId.value = id
}

function cancelDelete() {
  confirmDeleteId.value = null
}

function executeDelete(id: number) {
  emit('delete', id)
  confirmDeleteId.value = null
}
</script>

<template>
  <div class="space-y-4 pb-20 md:pb-8">
    <!-- Top Action & Search Bar -->
    <div class="flex flex-col sm:flex-row items-stretch sm:items-center justify-between gap-3">
      <!-- Search Input -->
      <div class="relative flex-1">
        <Search class="w-4 h-4 text-slate-400 absolute left-3.5 top-1/2 -translate-y-1/2" />
        <input
          v-model="searchQuery"
          type="text"
          placeholder="Cari reminder..."
          class="w-full h-10 pl-10 pr-3.5 text-xs sm:text-sm rounded-xl border border-slate-200 bg-white focus:border-blue-600 focus:ring-1 focus:ring-blue-600 outline-none transition"
        />
      </div>

      <!-- Filter Buttons & Create CTA -->
      <div class="flex items-center gap-2">
        <div class="flex items-center p-1 bg-slate-100 rounded-xl">
          <button
            @click="filter = 'all'"
            :class="filter === 'all' ? 'bg-white text-slate-900 shadow-sm font-semibold' : 'text-slate-600 hover:text-slate-900'"
            class="px-3 py-1.5 text-xs rounded-lg transition"
          >
            Semua
          </button>
          <button
            @click="filter = 'active'"
            :class="filter === 'active' ? 'bg-white text-blue-600 shadow-sm font-semibold' : 'text-slate-600 hover:text-slate-900'"
            class="px-3 py-1.5 text-xs rounded-lg transition"
          >
            Aktif
          </button>
          <button
            @click="filter = 'inactive'"
            :class="filter === 'inactive' ? 'bg-white text-slate-900 shadow-sm font-semibold' : 'text-slate-600 hover:text-slate-900'"
            class="px-3 py-1.5 text-xs rounded-lg transition"
          >
            Nonaktif
          </button>
        </div>

        <button
          @click="emit('create')"
          class="h-10 px-4 rounded-xl bg-blue-600 hover:bg-blue-700 text-white font-semibold text-xs flex items-center gap-1.5 transition shadow-sm shadow-blue-500/20 flex-shrink-0"
        >
          <Plus class="w-4 h-4 stroke-[2.5]" />
          <span>Tambah</span>
        </button>
      </div>
    </div>

    <!-- Empty State -->
    <div
      v-if="!loading && filteredReminders.length === 0"
      class="text-center py-16 bg-white rounded-2xl border border-slate-200/80 p-8 shadow-sm space-y-3"
    >
      <div class="w-14 h-14 mx-auto rounded-2xl bg-blue-50 text-blue-600 flex items-center justify-center">
        <Clock class="w-7 h-7" />
      </div>
      <h3 class="text-base font-bold text-slate-800">Belum Ada Reminder</h3>
      <p class="text-xs text-slate-500 max-w-sm mx-auto">
        {{ searchQuery ? 'Tidak ada reminder yang cocok dengan pencarian Anda.' : 'Buat reminder alarm pertama Anda agar sistem mengirim pesan WhatsApp otomatis.' }}
      </p>
      <button
        v-if="!searchQuery"
        @click="emit('create')"
        class="mt-2 inline-flex items-center gap-1.5 px-4 py-2 bg-blue-600 text-white text-xs font-semibold rounded-xl hover:bg-blue-700 transition"
      >
        <Plus class="w-4 h-4" />
        <span>Buat Reminder Sekarang</span>
      </button>
    </div>

    <!-- Reminder Cards List -->
    <div v-else class="grid grid-cols-1 md:grid-cols-2 gap-3.5">
      <div
        v-for="r in filteredReminders"
        :key="r.id"
        class="bg-white rounded-2xl border p-4 sm:p-5 shadow-sm transition hover:shadow-md relative overflow-hidden"
        :class="r.is_active ? 'border-slate-200/80' : 'border-slate-200/40 opacity-75 bg-slate-50/50'"
      >
        <!-- Card Header -->
        <div class="flex items-start justify-between gap-3 mb-2">
          <div class="flex-1">
            <div class="flex items-center gap-2">
              <span class="text-xs font-mono font-bold text-slate-400">#{{ r.id }}</span>
              <h4 class="font-bold text-slate-900 text-base leading-snug">
                {{ r.title }}
              </h4>
            </div>
            <p v-if="r.message && r.message !== r.title" class="text-xs text-slate-600 mt-1 line-clamp-2">
              {{ r.message }}
            </p>
          </div>

          <!-- Active Toggle Switch -->
          <button
            type="button"
            @click="emit('toggle', r.id, !r.is_active)"
            class="w-11 h-6 flex items-center rounded-full p-1 transition duration-200 flex-shrink-0 cursor-pointer"
            :class="r.is_active ? 'bg-blue-600' : 'bg-slate-300'"
            title="Klik untuk mengaktifkan/menonaktifkan"
          >
            <div
              class="bg-white w-4 h-4 rounded-full shadow-md transform transition duration-200"
              :class="{ 'translate-x-5': r.is_active }"
            />
          </button>
        </div>

        <!-- Schedule Badges & Meta -->
        <div class="flex flex-wrap items-center gap-2 py-2 text-xs">
          <span class="inline-flex items-center gap-1 px-2.5 py-1 rounded-lg bg-blue-50 text-blue-700 font-medium">
            <Clock class="w-3.5 h-3.5" />
            <span>{{ formatTime(r.scheduled_at, r.timezone) }} ({{ r.timezone }})</span>
          </span>
          <span class="inline-flex items-center gap-1 px-2.5 py-1 rounded-lg bg-slate-100 text-slate-700 font-medium">
            <span>🔄 {{ getRecurrenceLabel(r.recurrence_type, r.recurrence_days) }}</span>
          </span>
          <span v-if="r.target_phone" class="inline-flex items-center gap-1 px-2 py-1 rounded-lg bg-emerald-50 text-emerald-700 font-medium">
            <Smartphone class="w-3.5 h-3.5" />
            <span>+{{ r.target_phone }}</span>
          </span>
        </div>

        <!-- Card Footer Actions -->
        <div class="flex items-center justify-between pt-3 mt-1 border-t border-slate-100 text-xs">
          <span class="text-[11px] text-slate-400">
            {{ r.is_active ? '🟢 Aktif' : '⚪ Nonaktif' }}
          </span>

          <div class="flex items-center gap-1.5">
            <button
              @click="emit('edit', r)"
              class="h-8 px-2.5 rounded-lg text-slate-600 hover:text-blue-600 hover:bg-blue-50 font-medium flex items-center gap-1 transition"
            >
              <Edit3 class="w-3.5 h-3.5" />
              <span>Edit</span>
            </button>
            <button
              @click="confirmDelete(r.id)"
              class="h-8 px-2.5 rounded-lg text-slate-400 hover:text-rose-600 hover:bg-rose-50 font-medium flex items-center gap-1 transition"
            >
              <Trash2 class="w-3.5 h-3.5" />
              <span>Hapus</span>
            </button>
          </div>
        </div>

        <!-- Delete Confirmation Overlay -->
        <div
          v-if="confirmDeleteId === r.id"
          class="absolute inset-0 bg-white/95 backdrop-blur-sm p-4 flex flex-col items-center justify-center text-center space-y-3 z-10 animate-fade-in"
        >
          <div class="text-xs font-semibold text-slate-900">Hapus reminder "{{ r.title }}"?</div>
          <p class="text-[11px] text-slate-500">Tindakan ini tidak dapat dibatalkan.</p>
          <div class="flex items-center gap-2">
            <button
              @click="cancelDelete"
              class="h-8 px-3 rounded-lg border border-slate-300 text-slate-700 text-xs font-medium hover:bg-slate-50"
            >
              Batal
            </button>
            <button
              @click="executeDelete(r.id)"
              class="h-8 px-3 rounded-lg bg-rose-600 hover:bg-rose-700 text-white text-xs font-semibold shadow-sm"
            >
              Ya, Hapus
            </button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>
