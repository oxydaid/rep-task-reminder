<script setup lang="ts">
import { ref, computed } from 'vue'
import { ShieldCheck, Lock, AlertTriangle, ArrowRight, Check } from 'lucide-vue-next'

const props = defineProps<{
  setupRequired: boolean
  isLocked: boolean
  lockRemaining: number
  loading: boolean
  errorMessage?: string
}>()

const emit = defineEmits<{
  (e: 'login', pin: string): void
  (e: 'setup', pin: string): void
}>()

const pin = ref('')
const confirmPin = ref('')
const localError = ref('')

const isPinValid = computed(() => pin.value.length >= 4 && pin.value.length <= 12)
const isConfirmMatch = computed(() => !props.setupRequired || pin.value === confirmPin.value)

const displayError = computed(() => localError.value || props.errorMessage)

function handleKey(digit: string) {
  if (props.isLocked || props.loading) return
  localError.value = ''
  if (!props.setupRequired) {
    if (pin.value.length < 12) {
      pin.value += digit
    }
  } else {
    if (pin.value.length < 6) {
      pin.value += digit
    } else if (confirmPin.value.length < 6) {
      confirmPin.value += digit
    }
  }
}

function handleBackspace() {
  if (props.isLocked || props.loading) return
  localError.value = ''
  if (!props.setupRequired) {
    pin.value = pin.value.slice(0, -1)
  } else {
    if (confirmPin.value.length > 0) {
      confirmPin.value = confirmPin.value.slice(0, -1)
    } else {
      pin.value = pin.value.slice(0, -1)
    }
  }
}

function submit() {
  localError.value = ''
  if (!isPinValid.value) {
    localError.value = 'PIN minimal 4 digit'
    return
  }

  if (props.setupRequired) {
    if (pin.value !== confirmPin.value) {
      localError.value = 'Konfirmasi PIN tidak cocok'
      return
    }
    emit('setup', pin.value)
  } else {
    emit('login', pin.value)
  }
}
</script>

<template>
  <div class="min-h-screen flex items-center justify-center p-4 bg-slate-50">
    <div class="w-full max-w-sm bg-white rounded-2xl shadow-xl border border-slate-200/80 p-6 md:p-8">
      <!-- Header -->
      <div class="text-center mb-6">
        <div class="w-14 h-14 mx-auto mb-3 rounded-2xl bg-blue-600 flex items-center justify-center text-white shadow-lg shadow-blue-500/25">
          <ShieldCheck v-if="setupRequired" class="w-8 h-8" />
          <Lock v-else class="w-7 h-7" />
        </div>
        <h1 class="text-xl font-bold text-slate-900 tracking-tight">
          {{ setupRequired ? 'Atur PIN Keamanan' : 'Masuk ke Dashboard' }}
        </h1>
        <p class="text-sm text-slate-500 mt-1">
          {{ setupRequired ? 'Buat PIN 4-12 digit untuk melindungi aplikasi' : 'Masukkan PIN untuk mengakses WhatsApp Reminder' }}
        </p>
      </div>

      <!-- Lockout Alert -->
      <div v-if="isLocked" class="mb-5 p-3 rounded-xl bg-red-50 border border-red-200 text-red-700 text-sm flex items-start gap-2.5">
        <AlertTriangle class="w-5 h-5 flex-shrink-0 text-red-500 mt-0.5" />
        <div>
          <div class="font-semibold">Akses Terkunci</div>
          <div>Terlalu banyak percobaan gagal. Silakan tunggu <span class="font-bold underline">{{ lockRemaining }}</span> detik.</div>
        </div>
      </div>

      <!-- Error Message -->
      <div v-else-if="displayError" class="mb-5 p-3 rounded-xl bg-red-50 border border-red-200 text-red-600 text-sm flex items-center gap-2">
        <AlertTriangle class="w-4 h-4 flex-shrink-0" />
        <span>{{ displayError }}</span>
      </div>

      <!-- Form Mode: Setup -->
      <div v-if="setupRequired" class="space-y-4 mb-6">
        <div>
          <label class="block text-xs font-semibold text-slate-700 uppercase tracking-wider mb-1.5">PIN Baru</label>
          <input
            v-model="pin"
            type="password"
            inputmode="numeric"
            placeholder="Ketik 4-12 digit PIN"
            class="w-full h-11 px-3.5 text-center tracking-widest text-lg font-mono rounded-xl border border-slate-300 focus:border-blue-600 focus:ring-2 focus:ring-blue-100 outline-none transition"
            maxlength="12"
          />
        </div>
        <div>
          <label class="block text-xs font-semibold text-slate-700 uppercase tracking-wider mb-1.5">Konfirmasi PIN</label>
          <input
            v-model="confirmPin"
            type="password"
            inputmode="numeric"
            placeholder="Ulangi PIN baru"
            class="w-full h-11 px-3.5 text-center tracking-widest text-lg font-mono rounded-xl border border-slate-300 focus:border-blue-600 focus:ring-2 focus:ring-blue-100 outline-none transition"
            maxlength="12"
          />
        </div>
        <button
          @click="submit"
          :disabled="loading || !isPinValid || !isConfirmMatch"
          class="w-full h-11 mt-2 rounded-xl bg-blue-600 hover:bg-blue-700 text-white font-medium flex items-center justify-center gap-2 transition disabled:opacity-50 disabled:cursor-not-allowed shadow-md shadow-blue-500/20"
        >
          <span v-if="loading">Menyimpan...</span>
          <template v-else>
            <span>Simpan & Masuk</span>
            <ArrowRight class="w-4 h-4" />
          </template>
        </button>
      </div>

      <!-- Form Mode: Enter PIN -->
      <div v-else class="space-y-5">
        <!-- Dots Display -->
        <div class="flex justify-center gap-3 py-2">
          <div
            v-for="i in 6"
            :key="i"
            class="w-3.5 h-3.5 rounded-full transition-all duration-200"
            :class="pin.length >= i ? 'bg-blue-600 scale-110 shadow-sm shadow-blue-500/50' : 'bg-slate-200'"
          />
        </div>

        <!-- Numeric Keypad -->
        <div class="grid grid-cols-3 gap-2.5 max-w-[260px] mx-auto">
          <button
            v-for="digit in ['1', '2', '3', '4', '5', '6', '7', '8', '9']"
            :key="digit"
            @click="handleKey(digit)"
            :disabled="isLocked || loading"
            type="button"
            class="h-14 rounded-xl text-xl font-medium text-slate-800 bg-slate-100 hover:bg-slate-200 active:bg-blue-50 active:text-blue-600 transition flex items-center justify-center disabled:opacity-40"
          >
            {{ digit }}
          </button>
          <button
            @click="handleBackspace"
            :disabled="isLocked || loading || pin.length === 0"
            type="button"
            class="h-14 rounded-xl text-sm font-medium text-slate-600 bg-slate-100 hover:bg-slate-200 active:bg-slate-300 transition flex items-center justify-center disabled:opacity-40"
          >
            Hapus
          </button>
          <button
            @click="handleKey('0')"
            :disabled="isLocked || loading"
            type="button"
            class="h-14 rounded-xl text-xl font-medium text-slate-800 bg-slate-100 hover:bg-slate-200 active:bg-blue-50 active:text-blue-600 transition flex items-center justify-center disabled:opacity-40"
          >
            0
          </button>
          <button
            @click="submit"
            :disabled="isLocked || loading || pin.length < 4"
            type="button"
            class="h-14 rounded-xl font-semibold text-white bg-blue-600 hover:bg-blue-700 active:bg-blue-800 transition flex items-center justify-center disabled:opacity-40 shadow-sm shadow-blue-500/20"
          >
            <Check class="w-6 h-6" />
          </button>
        </div>
      </div>
    </div>
  </div>
</template>
