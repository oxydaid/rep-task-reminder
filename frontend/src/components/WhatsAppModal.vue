<script setup lang="ts">
import { ref, computed } from 'vue'
import { QrCode, RefreshCw, Smartphone, LogOut, Send, CheckCircle2, X } from 'lucide-vue-next'
import type { WhatsAppConnectionStatus } from '../types'

const props = defineProps<{
  show: boolean
  status: WhatsAppConnectionStatus
  connectedPhone?: string
  qrCode?: string
  loading: boolean
  targetPhone?: string
}>()

const emit = defineEmits<{
  (e: 'close'): void
  (e: 'connect'): void
  (e: 'disconnect'): void
  (e: 'logout'): void
  (e: 'send-test', phone: string, message: string): void
}>()

const isConnected = computed(() => props.status === 'connected' && !!props.connectedPhone)

const testPhone = ref(props.targetPhone || '')
const testMessage = ref('Halo! Ini adalah pesan uji coba WhatsApp Reminder.')
const testSuccess = ref(false)
const testError = ref('')
const sendingTest = ref(false)

async function handleSendTest() {
  testSuccess.value = false
  testError.value = ''
  sendingTest.value = true
  try {
    emit('send-test', testPhone.value, testMessage.value)
    testSuccess.value = true
    setTimeout(() => {
      testSuccess.value = false
    }, 4000)
  } catch (e: any) {
    testError.value = e.message || 'Gagal mengirim pesan'
  } finally {
    sendingTest.value = false
  }
}
</script>

<template>
  <div v-if="show" class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-slate-900/50 backdrop-blur-sm animate-fade-in">
    <div class="w-full max-w-md bg-white rounded-2xl shadow-2xl border border-slate-200 overflow-hidden">
      <!-- Modal Header -->
      <div class="px-5 py-4 border-b border-slate-100 flex items-center justify-between">
        <div class="flex items-center gap-2">
          <div class="w-8 h-8 rounded-lg bg-emerald-50 text-emerald-600 flex items-center justify-center">
            <Smartphone class="w-5 h-5" />
          </div>
          <h2 class="font-bold text-slate-900 text-base">Koneksi WhatsApp</h2>
        </div>
        <button
          @click="emit('close')"
          class="w-7 h-7 rounded-lg text-slate-400 hover:text-slate-600 hover:bg-slate-100 flex items-center justify-center transition"
        >
          <X class="w-4 h-4" />
        </button>
      </div>

      <!-- Modal Body -->
      <div class="p-6">
        <!-- Case 1: Connected with verified phone -->
        <div v-if="isConnected" class="text-center space-y-5">
          <div class="w-14 h-14 mx-auto rounded-full bg-emerald-100 text-emerald-600 flex items-center justify-center">
            <CheckCircle2 class="w-8 h-8" />
          </div>
          <div>
            <h3 class="font-bold text-slate-900 text-lg">WhatsApp Terhubung!</h3>
            <p class="text-sm text-slate-500 mt-1">
              Akun aktif: <span class="font-semibold text-slate-800 font-mono">+{{ connectedPhone }}</span>
            </p>
          </div>

          <!-- Test Message Section -->
          <div class="p-4 rounded-xl bg-slate-50 border border-slate-200 text-left space-y-3">
            <div class="text-xs font-semibold text-slate-700 uppercase tracking-wider">Uji Coba Pengiriman Pesan</div>
            <input
              v-model="testPhone"
              type="text"
              placeholder="Nomor tujuan (contoh: 08123456789)"
              class="w-full h-9 px-3 text-sm rounded-lg border border-slate-300 focus:border-blue-500 focus:ring-1 focus:ring-blue-500 outline-none"
            />
            <input
              v-model="testMessage"
              type="text"
              placeholder="Teks pesan"
              class="w-full h-9 px-3 text-sm rounded-lg border border-slate-300 focus:border-blue-500 focus:ring-1 focus:ring-blue-500 outline-none"
            />
            <button
              @click="handleSendTest"
              :disabled="sendingTest"
              class="w-full h-9 bg-blue-600 hover:bg-blue-700 text-white text-xs font-semibold rounded-lg flex items-center justify-center gap-1.5 transition disabled:opacity-50"
            >
              <Send class="w-3.5 h-3.5" />
              <span>{{ sendingTest ? 'Mengirim...' : 'Kirim Pesan Uji Coba' }}</span>
            </button>
            <div v-if="testSuccess" class="text-xs text-emerald-600 font-medium">✅ Pesan uji coba berhasil dikirim!</div>
            <div v-if="testError" class="text-xs text-rose-600 font-medium">{{ testError }}</div>
          </div>

          <!-- Disconnect / Unpair -->
          <div class="flex gap-2.5 pt-2">
            <button
              @click="emit('disconnect')"
              :disabled="loading"
              class="flex-1 h-10 border border-slate-300 hover:bg-slate-50 text-slate-700 font-medium text-xs rounded-xl transition"
            >
              Putuskan Sementara
            </button>
            <button
              @click="emit('logout')"
              :disabled="loading"
              class="flex-1 h-10 bg-rose-50 hover:bg-rose-100 text-rose-600 font-medium text-xs rounded-xl flex items-center justify-center gap-1.5 transition"
            >
              <LogOut class="w-3.5 h-3.5" />
              <span>Hapus Sesi (Unpair)</span>
            </button>
          </div>
        </div>

        <!-- Case 2: Connecting / Scan QR -->
        <div v-else-if="status === 'connecting' || qrCode" class="text-center space-y-4">
          <div v-if="qrCode" class="space-y-3">
            <div class="inline-block p-3 rounded-2xl bg-white border border-slate-200 shadow-md">
              <img :src="qrCode" alt="Scan QR WhatsApp" class="w-56 h-56 mx-auto object-contain" />
            </div>
            <p class="text-xs text-slate-500">
              QR Code berganti otomatis. Silakan scan segera.
            </p>
          </div>
          <div v-else class="space-y-3">
            <div class="inline-block p-3 rounded-2xl bg-white border border-slate-200 shadow-md">
              <div class="w-56 h-56 rounded-xl bg-slate-100 flex flex-col items-center justify-center p-4 text-center animate-pulse">
                <RefreshCw class="w-8 h-8 text-blue-600 animate-spin mb-3" />
                <span class="text-xs font-semibold text-slate-700">Menyiapkan QR Code...</span>
                <span class="text-[11px] text-slate-400 mt-1">Menghubungkan ke WhatsApp</span>
              </div>
            </div>
            <p class="text-xs text-slate-400">
              Mohon tunggu beberapa saat...
            </p>
          </div>

          <div class="bg-blue-50 border border-blue-100 p-3 rounded-xl text-left text-xs text-blue-900 space-y-1">
            <div class="font-semibold text-blue-950">Cara Scan:</div>
            <ol class="list-decimal list-inside space-y-0.5 text-blue-800">
              <li>Buka WhatsApp di ponsel Anda</li>
              <li>Buka Menu (titik tiga) atau Pengaturan &gt; <b>Perangkat Tertaut</b></li>
              <li>Pilih <b>Tautkan Perangkat</b></li>
              <li>Arahkan kamera ke QR Code di atas</li>
            </ol>
          </div>

          <div class="flex gap-2 pt-2">
            <button
              @click="emit('connect')"
              :disabled="loading"
              class="flex-1 h-9 border border-slate-200 hover:bg-slate-50 text-slate-700 text-xs font-semibold rounded-xl flex items-center justify-center gap-1.5 transition"
            >
              <RefreshCw :class="['w-3.5 h-3.5', loading ? 'animate-spin' : '']" />
              <span>Muat Ulang QR</span>
            </button>
            <button
              @click="emit('close')"
              class="h-9 px-4 border border-slate-200 hover:bg-slate-50 text-slate-600 text-xs font-medium rounded-xl transition"
            >
              Tutup
            </button>
          </div>
        </div>

        <!-- Case 3: Disconnected -->
        <div v-else class="text-center space-y-4 py-4">
          <div class="w-14 h-14 mx-auto rounded-full bg-slate-100 text-slate-500 flex items-center justify-center">
            <QrCode class="w-7 h-7" />
          </div>
          <div>
            <h3 class="font-bold text-slate-900 text-lg">WhatsApp Belum Terhubung</h3>
            <p class="text-sm text-slate-500 mt-1 max-w-xs mx-auto">
              Hubungkan akun WhatsApp agar sistem dapat mengirimkan pesan alarm pengingat otomatis.
            </p>
          </div>

          <button
            @click="emit('connect')"
            :disabled="loading"
            class="w-full h-11 bg-blue-600 hover:bg-blue-700 text-white font-medium text-sm rounded-xl flex items-center justify-center gap-2 transition shadow-md shadow-blue-500/20"
          >
            <RefreshCw v-if="loading" class="w-4 h-4 animate-spin" />
            <template v-else>
              <QrCode class="w-4 h-4" />
              <span>Tampilkan QR Code</span>
            </template>
          </button>
        </div>
      </div>
    </div>
  </div>
</template>
