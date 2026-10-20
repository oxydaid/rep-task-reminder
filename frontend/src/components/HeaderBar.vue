<script setup lang="ts">
import { Bell, Wifi, WifiOff, RefreshCw, LogOut, LayoutDashboard, Clock, History, Settings } from 'lucide-vue-next'
import type { WhatsAppConnectionStatus } from '../types'

defineProps<{
  activeTab: 'dashboard' | 'reminders' | 'history' | 'settings'
  waStatus: WhatsAppConnectionStatus
  connectedPhone?: string
}>()

const emit = defineEmits<{
  (e: 'update:activeTab', tab: 'dashboard' | 'reminders' | 'history' | 'settings'): void
  (e: 'logout'): void
  (e: 'open-qr'): void
}>()
</script>

<template>
  <header class="sticky top-0 z-30 bg-white/90 backdrop-blur-md border-b border-slate-200">
    <div class="max-w-5xl mx-auto px-4 h-16 flex items-center justify-between gap-2">
      <!-- Brand Logo & Title -->
      <div class="flex items-center gap-2.5">
        <div class="w-9 h-9 rounded-xl bg-blue-600 flex items-center justify-center text-white shadow-md shadow-blue-500/20">
          <Bell class="w-5 h-5" />
        </div>
        <div>
          <span class="font-bold text-slate-900 tracking-tight text-base sm:text-lg block leading-none">
            WhatsApp Reminder
          </span>
          <span class="text-[11px] text-slate-400 font-medium hidden sm:inline">
            Personal Alarm & Scheduler
          </span>
        </div>
      </div>

      <!-- Center Nav (Desktop) -->
      <nav class="hidden md:flex items-center gap-1 bg-slate-100 p-1 rounded-xl">
        <button
          @click="emit('update:activeTab', 'dashboard')"
          :class="activeTab === 'dashboard' ? 'bg-white text-blue-600 shadow-sm font-semibold' : 'text-slate-600 hover:text-slate-900'"
          class="px-3.5 py-1.5 text-xs rounded-lg transition flex items-center gap-1.5"
        >
          <LayoutDashboard class="w-4 h-4" />
          <span>Dashboard</span>
        </button>
        <button
          @click="emit('update:activeTab', 'reminders')"
          :class="activeTab === 'reminders' ? 'bg-white text-blue-600 shadow-sm font-semibold' : 'text-slate-600 hover:text-slate-900'"
          class="px-3.5 py-1.5 text-xs rounded-lg transition flex items-center gap-1.5"
        >
          <Clock class="w-4 h-4" />
          <span>Reminder</span>
        </button>
        <button
          @click="emit('update:activeTab', 'history')"
          :class="activeTab === 'history' ? 'bg-white text-blue-600 shadow-sm font-semibold' : 'text-slate-600 hover:text-slate-900'"
          class="px-3.5 py-1.5 text-xs rounded-lg transition flex items-center gap-1.5"
        >
          <History class="w-4 h-4" />
          <span>Riwayat</span>
        </button>
        <button
          @click="emit('update:activeTab', 'settings')"
          :class="activeTab === 'settings' ? 'bg-white text-blue-600 shadow-sm font-semibold' : 'text-slate-600 hover:text-slate-900'"
          class="px-3.5 py-1.5 text-xs rounded-lg transition flex items-center gap-1.5"
        >
          <Settings class="w-4 h-4" />
          <span>Pengaturan</span>
        </button>
      </nav>

      <!-- Right: WhatsApp Status Pill & Logout -->
      <div class="flex items-center gap-2">
        <button
          @click="emit('open-qr')"
          class="h-8 px-2.5 rounded-full text-xs font-medium flex items-center gap-1.5 transition border cursor-pointer"
          :class="{
            'bg-emerald-50 text-emerald-700 border-emerald-200 hover:bg-emerald-100': waStatus === 'connected',
            'bg-amber-50 text-amber-700 border-amber-200 animate-pulse': waStatus === 'connecting',
            'bg-rose-50 text-rose-700 border-rose-200 hover:bg-rose-100': waStatus === 'disconnected'
          }"
        >
          <span class="w-2 h-2 rounded-full" :class="{
            'bg-emerald-500': waStatus === 'connected',
            'bg-amber-500': waStatus === 'connecting',
            'bg-rose-500': waStatus === 'disconnected'
          }" />
          <Wifi v-if="waStatus === 'connected'" class="w-3.5 h-3.5" />
          <RefreshCw v-else-if="waStatus === 'connecting'" class="w-3.5 h-3.5 animate-spin" />
          <WifiOff v-else class="w-3.5 h-3.5" />
          
          <span class="hidden sm:inline">
            {{ waStatus === 'connected' ? 'WA Terhubung' : waStatus === 'connecting' ? 'Menghubungkan...' : 'WA Terputus' }}
          </span>
        </button>

        <button
          @click="emit('logout')"
          title="Keluar"
          class="w-8 h-8 rounded-lg text-slate-500 hover:text-slate-900 hover:bg-slate-100 flex items-center justify-center transition"
        >
          <LogOut class="w-4 h-4" />
        </button>
      </div>
    </div>
  </header>
</template>
