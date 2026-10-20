<script setup lang="ts">
import { LayoutDashboard, Clock, History, Settings } from 'lucide-vue-next'

defineProps<{
  activeTab: 'dashboard' | 'reminders' | 'history' | 'settings'
}>()

const emit = defineEmits<{
  (e: 'update:activeTab', tab: 'dashboard' | 'reminders' | 'history' | 'settings'): void
}>()
</script>

<template>
  <nav class="md:hidden fixed bottom-0 left-0 right-0 z-30 bg-white/95 backdrop-blur-md border-t border-slate-200 px-4 py-1.5 flex items-center justify-around shadow-lg">
    <button
      @click="emit('update:activeTab', 'dashboard')"
      class="flex flex-col items-center gap-1 py-1 px-3 rounded-lg transition"
      :class="activeTab === 'dashboard' ? 'text-blue-600 font-semibold' : 'text-slate-500 hover:text-slate-800'"
    >
      <LayoutDashboard class="w-5 h-5" />
      <span class="text-[10px]">Home</span>
    </button>
    <button
      @click="emit('update:activeTab', 'reminders')"
      class="flex flex-col items-center gap-1 py-1 px-3 rounded-lg transition"
      :class="activeTab === 'reminders' ? 'text-blue-600 font-semibold' : 'text-slate-500 hover:text-slate-800'"
    >
      <Clock class="w-5 h-5" />
      <span class="text-[10px]">Reminder</span>
    </button>
    <button
      @click="emit('update:activeTab', 'history')"
      class="flex flex-col items-center gap-1 py-1 px-3 rounded-lg transition"
      :class="activeTab === 'history' ? 'text-blue-600 font-semibold' : 'text-slate-500 hover:text-slate-800'"
    >
      <History class="w-5 h-5" />
      <span class="text-[10px]">Riwayat</span>
    </button>
    <button
      @click="emit('update:activeTab', 'settings')"
      class="flex flex-col items-center gap-1 py-1 px-3 rounded-lg transition"
      :class="activeTab === 'settings' ? 'text-blue-600 font-semibold' : 'text-slate-500 hover:text-slate-800'"
    >
      <Settings class="w-5 h-5" />
      <span class="text-[10px]">Pengaturan</span>
    </button>
  </nav>
</template>
