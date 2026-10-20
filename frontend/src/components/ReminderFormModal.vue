<script setup lang="ts">
import { ref, watch } from 'vue'
import { X, Check, AlertCircle } from 'lucide-vue-next'
import type { Reminder, RecurrenceType } from '../types'

const props = defineProps<{
  show: boolean
  reminder?: Reminder | null
  loading: boolean
}>()

const emit = defineEmits<{
  (e: 'close'): void
  (e: 'save', payload: Partial<Reminder>): void
}>()

const title = ref('')
const message = ref('')
const timeStr = ref('08:00')
const dateStr = ref(new Date().toISOString().slice(0, 10))
const recurrenceType = ref<RecurrenceType>('daily')
const selectedDays = ref<number[]>([1, 2, 3, 4, 5]) // Mon-Fri default
const timezone = ref('Asia/Jakarta')
const targetPhone = ref('')
const isActive = ref(true)
const errorMsg = ref('')

const dayOptions = [
  { label: 'Sen', value: 1 },
  { label: 'Sel', value: 2 },
  { label: 'Rab', value: 3 },
  { label: 'Kam', value: 4 },
  { label: 'Jum', value: 5 },
  { label: 'Sab', value: 6 },
  { label: 'Min', value: 7 },
]

function toggleDay(val: number) {
  if (selectedDays.value.includes(val)) {
    if (selectedDays.value.length > 1) {
      selectedDays.value = selectedDays.value.filter((d) => d !== val)
    }
  } else {
    selectedDays.value = [...selectedDays.value, val].sort((a, b) => a - b)
  }
}

watch(
  () => props.show,
  (open) => {
    if (open) {
      errorMsg.value = ''
      if (props.reminder) {
        title.value = props.reminder.title
        message.value = props.reminder.message
        recurrenceType.value = props.reminder.recurrence_type
        timezone.value = props.reminder.timezone || 'Asia/Jakarta'
        targetPhone.value = props.reminder.target_phone || ''
        isActive.value = props.reminder.is_active

        if (props.reminder.scheduled_at) {
          const d = new Date(props.reminder.scheduled_at)
          dateStr.value = d.toISOString().slice(0, 10)
          timeStr.value = `${String(d.getHours()).padStart(2, '0')}:${String(d.getMinutes()).padStart(2, '0')}`
        }

        if (props.reminder.recurrence_days) {
          selectedDays.value = props.reminder.recurrence_days
            .split(',')
            .map((s) => parseInt(s.trim(), 10))
            .filter((n) => !isNaN(n))
        }
      } else {
        // Reset defaults
        title.value = ''
        message.value = ''
        timeStr.value = '08:00'
        dateStr.value = new Date().toISOString().slice(0, 10)
        recurrenceType.value = 'daily'
        selectedDays.value = [1, 2, 3, 4, 5]
        timezone.value = 'Asia/Jakarta'
        targetPhone.value = ''
        isActive.value = true
      }
    }
  }
)

function handleSubmit() {
  errorMsg.value = ''
  if (!title.value.trim()) {
    errorMsg.value = 'Judul reminder wajib diisi'
    return
  }
  if (!timeStr.value) {
    errorMsg.value = 'Jam pengingat wajib diisi'
    return
  }

  // Build scheduled_at
  const [h, m] = timeStr.value.split(':').map((v) => parseInt(v, 10))
  let scheduledDate = new Date()
  if (recurrenceType.value === 'once' && dateStr.value) {
    scheduledDate = new Date(dateStr.value)
  }
  scheduledDate.setHours(h, m, 0, 0)

  // If time in the past for today and recurrence is daily/weekly, push to tomorrow/next
  const now = new Date()
  if (recurrenceType.value !== 'once' && scheduledDate.getTime() <= now.getTime()) {
    scheduledDate.setDate(scheduledDate.getDate() + 1)
  }

  const payload: Partial<Reminder> = {
    title: title.value.trim(),
    message: message.value.trim() || title.value.trim(),
    scheduled_at: scheduledDate.toISOString(),
    recurrence_type: recurrenceType.value,
    recurrence_days: recurrenceType.value === 'specific_days' ? selectedDays.value.join(',') : '',
    timezone: timezone.value,
    target_phone: targetPhone.value.trim(),
    is_active: isActive.value,
  }

  emit('save', payload)
}
</script>

<template>
  <div v-if="show" class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-slate-900/50 backdrop-blur-sm animate-fade-in overflow-y-auto">
    <div class="w-full max-w-lg bg-white rounded-2xl shadow-2xl border border-slate-200 overflow-hidden my-8">
      <!-- Header -->
      <div class="px-6 py-4 border-b border-slate-100 flex items-center justify-between">
        <h2 class="font-bold text-slate-900 text-lg">
          {{ reminder ? 'Ubah Reminder' : 'Buat Reminder Baru' }}
        </h2>
        <button
          @click="emit('close')"
          class="w-7 h-7 rounded-lg text-slate-400 hover:text-slate-600 hover:bg-slate-100 flex items-center justify-center transition"
        >
          <X class="w-4 h-4" />
        </button>
      </div>

      <!-- Form -->
      <div class="p-6 space-y-4 max-h-[75vh] overflow-y-auto">
        <div v-if="errorMsg" class="p-3 rounded-xl bg-red-50 border border-red-200 text-red-600 text-xs flex items-center gap-2">
          <AlertCircle class="w-4 h-4 flex-shrink-0" />
          <span>{{ errorMsg }}</span>
        </div>

        <!-- Title -->
        <div>
          <label class="block text-xs font-semibold text-slate-700 uppercase tracking-wider mb-1">
            Judul Reminder *
          </label>
          <input
            v-model="title"
            type="text"
            placeholder="Contoh: Absen Kerja, Minum Obat, Meeting"
            class="w-full h-10 px-3.5 text-sm rounded-xl border border-slate-300 focus:border-blue-600 focus:ring-1 focus:ring-blue-600 outline-none"
          />
        </div>

        <!-- Message -->
        <div>
          <label class="block text-xs font-semibold text-slate-700 uppercase tracking-wider mb-1">
            Pesan Pengingat (Opsional)
          </label>
          <textarea
            v-model="message"
            rows="2"
            placeholder="Pesan tambahan yang akan dikirim via WhatsApp"
            class="w-full p-3 text-sm rounded-xl border border-slate-300 focus:border-blue-600 focus:ring-1 focus:ring-blue-600 outline-none resize-none"
          />
        </div>

        <!-- Time & Timezone Row -->
        <div class="grid grid-cols-1 sm:grid-cols-2 gap-3">
          <div>
            <label class="block text-xs font-semibold text-slate-700 uppercase tracking-wider mb-1">
              Jam Pengingat *
            </label>
            <div class="relative">
              <input
                v-model="timeStr"
                type="time"
                class="w-full h-10 px-3.5 text-sm rounded-xl border border-slate-300 focus:border-blue-600 focus:ring-1 focus:ring-blue-600 outline-none"
              />
            </div>
          </div>
          <div>
            <label class="block text-xs font-semibold text-slate-700 uppercase tracking-wider mb-1">
              Zona Waktu
            </label>
            <select
              v-model="timezone"
              class="w-full h-10 px-3 text-sm rounded-xl border border-slate-300 focus:border-blue-600 focus:ring-1 focus:ring-blue-600 outline-none bg-white"
            >
              <option value="Asia/Jakarta">WIB (Asia/Jakarta)</option>
              <option value="Asia/Makassar">WITA (Asia/Makassar)</option>
              <option value="Asia/Jayapura">WIT (Asia/Jayapura)</option>
            </select>
          </div>
        </div>

        <!-- Recurrence Type -->
        <div>
          <label class="block text-xs font-semibold text-slate-700 uppercase tracking-wider mb-1.5">
            Pengulangan
          </label>
          <div class="grid grid-cols-2 sm:grid-cols-4 gap-2">
            <button
              type="button"
              @click="recurrenceType = 'daily'"
              class="h-9 text-xs rounded-xl border font-medium transition"
              :class="recurrenceType === 'daily' ? 'bg-blue-50 border-blue-600 text-blue-600 font-semibold' : 'border-slate-200 text-slate-600 hover:bg-slate-50'"
            >
              Setiap Hari
            </button>
            <button
              type="button"
              @click="recurrenceType = 'specific_days'"
              class="h-9 text-xs rounded-xl border font-medium transition"
              :class="recurrenceType === 'specific_days' ? 'bg-blue-50 border-blue-600 text-blue-600 font-semibold' : 'border-slate-200 text-slate-600 hover:bg-slate-50'"
            >
              Hari Tertentu
            </button>
            <button
              type="button"
              @click="recurrenceType = 'weekly'"
              class="h-9 text-xs rounded-xl border font-medium transition"
              :class="recurrenceType === 'weekly' ? 'bg-blue-50 border-blue-600 text-blue-600 font-semibold' : 'border-slate-200 text-slate-600 hover:bg-slate-50'"
            >
              Mingguan
            </button>
            <button
              type="button"
              @click="recurrenceType = 'once'"
              class="h-9 text-xs rounded-xl border font-medium transition"
              :class="recurrenceType === 'once' ? 'bg-blue-50 border-blue-600 text-blue-600 font-semibold' : 'border-slate-200 text-slate-600 hover:bg-slate-50'"
            >
              Sekali Saja
            </button>
          </div>
        </div>

        <!-- Specific Days Selector -->
        <div v-if="recurrenceType === 'specific_days'" class="p-3 rounded-xl bg-slate-50 border border-slate-200 space-y-2">
          <div class="text-xs text-slate-600 font-medium">Pilih hari aktif:</div>
          <div class="flex flex-wrap gap-1.5">
            <button
              v-for="d in dayOptions"
              :key="d.value"
              type="button"
              @click="toggleDay(d.value)"
              class="w-10 h-8 rounded-lg text-xs font-semibold transition"
              :class="selectedDays.includes(d.value) ? 'bg-blue-600 text-white shadow-sm' : 'bg-white border border-slate-200 text-slate-600 hover:bg-slate-100'"
            >
              {{ d.label }}
            </button>
          </div>
        </div>

        <!-- Date Picker for Once -->
        <div v-if="recurrenceType === 'once'">
          <label class="block text-xs font-semibold text-slate-700 uppercase tracking-wider mb-1">
            Tanggal Pelaksanaan
          </label>
          <input
            v-model="dateStr"
            type="date"
            class="w-full h-10 px-3.5 text-sm rounded-xl border border-slate-300 focus:border-blue-600 focus:ring-1 focus:ring-blue-600 outline-none"
          />
        </div>

        <!-- Optional Target Phone -->
        <div>
          <label class="block text-xs font-semibold text-slate-700 uppercase tracking-wider mb-1">
            Nomor Penerima Khusus (Opsional)
          </label>
          <input
            v-model="targetPhone"
            type="text"
            placeholder="Kosongkan untuk dikirim ke nomor WhatsApp Anda sendiri"
            class="w-full h-10 px-3.5 text-sm rounded-xl border border-slate-300 focus:border-blue-600 focus:ring-1 focus:ring-blue-600 outline-none"
          />
          <p class="text-[11px] text-slate-400 mt-1">Default adalah nomor WhatsApp akun yang terhubung.</p>
        </div>

        <!-- Is Active Switch -->
        <div class="flex items-center justify-between pt-2 border-t border-slate-100">
          <div>
            <div class="text-sm font-semibold text-slate-800">Status Aktif</div>
            <div class="text-xs text-slate-500">Scheduler akan menjalankan reminder ini jika aktif</div>
          </div>
          <button
            type="button"
            @click="isActive = !isActive"
            class="w-12 h-6 flex items-center rounded-full p-1 transition duration-200 cursor-pointer"
            :class="isActive ? 'bg-blue-600' : 'bg-slate-300'"
          >
            <div
              class="bg-white w-4 h-4 rounded-full shadow-md transform transition duration-200"
              :class="{ 'translate-x-6': isActive }"
            />
          </button>
        </div>
      </div>

      <!-- Footer Actions -->
      <div class="px-6 py-4 bg-slate-50 border-t border-slate-100 flex items-center justify-end gap-2.5">
        <button
          type="button"
          @click="emit('close')"
          class="h-10 px-4 rounded-xl border border-slate-300 hover:bg-slate-100 text-slate-700 text-xs font-medium transition"
        >
          Batal
        </button>
        <button
          type="button"
          @click="handleSubmit"
          :disabled="loading"
          class="h-10 px-5 rounded-xl bg-blue-600 hover:bg-blue-700 text-white text-xs font-semibold flex items-center gap-1.5 transition disabled:opacity-50 shadow-md shadow-blue-500/20"
        >
          <Check class="w-4 h-4" />
          <span>{{ loading ? 'Menyimpan...' : 'Simpan Reminder' }}</span>
        </button>
      </div>
    </div>
  </div>
</template>
