<script setup lang="ts">
import { ref, onMounted, watch } from 'vue'
import { useAuth } from './composables/useAuth'
import { useWhatsApp } from './composables/useWhatsApp'
import { useReminders } from './composables/useReminders'
import { useLogs } from './composables/useLogs'
import { useDocumentTitle } from './composables/useDocumentTitle'
import type { Reminder } from './types'

// Components
import PinAuthScreen from './components/PinAuthScreen.vue'
import HeaderBar from './components/HeaderBar.vue'
import BottomNav from './components/BottomNav.vue'
import DashboardTab from './components/DashboardTab.vue'
import RemindersTab from './components/RemindersTab.vue'
import HistoryTab from './components/HistoryTab.vue'
import SettingsTab from './components/SettingsTab.vue'
import WhatsAppModal from './components/WhatsAppModal.vue'
import ReminderFormModal from './components/ReminderFormModal.vue'

// Auth Composable
const {
  isAuthenticated,
  setupRequired,
  isLocked,
  lockRemaining,
  loading: authLoading,
  error: authError,
  checkStatus,
  login,
  setupPIN,
  logout
} = useAuth()

// WhatsApp Composable
const {
  status: waStatus,
  connectedPhone,
  qrCode,
  targetPhone,
  loading: waLoading,
  fetchStatus: fetchWAStatus,
  connect: connectWA,
  disconnect: disconnectWA,
  logoutSession: logoutWA,
  sendTestMessage
} = useWhatsApp()

// Reminders Composable
const {
  reminders,
  upcoming,
  stats,
  loading: remLoading,
  fetchReminders,
  fetchUpcoming,
  fetchStats,
  createReminder,
  updateReminder,
  deleteReminder,
  toggleReminder
} = useReminders()

// Dynamic Document Title using Adaptable Composable
useDocumentTitle(() => {
  if (!isAuthenticated.value) return 'Masuk'
  if (stats.value.active_reminders > 0) {
    return `(${stats.value.active_reminders} Aktif)`
  }
  return 'Dashboard'
})

// Logs Composable
const {
  logs,
  total: totalLogs,
  loading: logsLoading,
  fetchLogs
} = useLogs()

// Navigation & Modals State
const activeTab = ref<'dashboard' | 'reminders' | 'history' | 'settings'>('dashboard')
const showQAModal = ref(false)
const showFormModal = ref(false)
const editingReminder = ref<Reminder | null>(null)

// Init data upon authentication
async function initAppData() {
  await Promise.allSettled([
    fetchWAStatus(),
    fetchReminders(),
    fetchUpcoming(),
    fetchStats(),
    fetchLogs(),
  ])
}

onMounted(async () => {
  await checkStatus()
  if (isAuthenticated.value) {
    await initAppData()
  }
})

watch(isAuthenticated, (authenticated) => {
  if (authenticated) {
    initAppData()
  }
})

// Handlers
function openCreateModal() {
  editingReminder.value = null
  showFormModal.value = true
}

function openEditModal(item: Reminder) {
  editingReminder.value = item
  showFormModal.value = true
}

async function handleSaveReminder(payload: Partial<Reminder>) {
  if (editingReminder.value) {
    await updateReminder(editingReminder.value.id, payload)
  } else {
    await createReminder(payload)
  }
  showFormModal.value = false
}

async function handleToggleReminder(id: number, active: boolean) {
  await toggleReminder(id, active)
}

async function handleDeleteReminder(id: number) {
  await deleteReminder(id)
}

function handleViewReminders() {
  activeTab.value = 'reminders'
}

function handleOpenQR() {
  showQAModal.value = true
  if (waStatus.value === 'disconnected') {
    connectWA()
  }
}
</script>

<template>
  <div class="min-h-screen bg-slate-50 text-slate-900 flex flex-col font-sans">
    <!-- Unauthenticated / PIN Screen -->
    <PinAuthScreen
      v-if="!isAuthenticated"
      :setup-required="setupRequired"
      :is-locked="isLocked"
      :lock-remaining="lockRemaining"
      :loading="authLoading"
      :error-message="authError"
      @login="login"
      @setup="setupPIN"
    />

    <!-- Main Application UI -->
    <template v-else>
      <HeaderBar
        :active-tab="activeTab"
        :wa-status="waStatus"
        :connected-phone="connectedPhone"
        @update:active-tab="activeTab = $event"
        @logout="logout"
        @open-qr="handleOpenQR"
      />

      <!-- Content Area -->
      <main class="flex-1 max-w-5xl w-full mx-auto px-4 pt-5">
        <DashboardTab
          v-if="activeTab === 'dashboard'"
          :upcoming="upcoming"
          :stats="stats"
          :wa-status="waStatus"
          :connected-phone="connectedPhone"
          :loading="remLoading"
          @create="openCreateModal"
          @open-qr="handleOpenQR"
          @view-reminders="handleViewReminders"
          @toggle-reminder="handleToggleReminder"
        />

        <RemindersTab
          v-else-if="activeTab === 'reminders'"
          :reminders="reminders"
          :loading="remLoading"
          @create="openCreateModal"
          @edit="openEditModal"
          @delete="handleDeleteReminder"
          @toggle="handleToggleReminder"
        />

        <HistoryTab
          v-else-if="activeTab === 'history'"
          :logs="logs"
          :total="totalLogs"
          :loading="logsLoading"
          @refresh="fetchLogs"
          @page-change="fetchLogs(20, $event)"
        />

        <SettingsTab
          v-else-if="activeTab === 'settings'"
          :wa-status="waStatus"
          :connected-phone="connectedPhone"
          @open-qr="handleOpenQR"
          @logout-wa="logoutWA"
        />
      </main>

      <!-- Bottom Nav for Mobile -->
      <BottomNav
        :active-tab="activeTab"
        @update:active-tab="activeTab = $event"
      />

      <!-- WhatsApp QR / Connection Modal -->
      <WhatsAppModal
        :show="showQAModal"
        :status="waStatus"
        :connected-phone="connectedPhone"
        :qr-code="qrCode"
        :loading="waLoading"
        :target-phone="targetPhone"
        @close="showQAModal = false"
        @connect="connectWA"
        @disconnect="disconnectWA"
        @logout="logoutWA"
        @send-test="sendTestMessage"
      />

      <!-- Reminder Create / Edit Modal -->
      <ReminderFormModal
        :show="showFormModal"
        :reminder="editingReminder"
        :loading="remLoading"
        @close="showFormModal = false"
        @save="handleSaveReminder"
      />
    </template>
  </div>
</template>
