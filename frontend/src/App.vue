<script setup>
import { onMounted, onUnmounted } from 'vue'
import { useAuthStore } from './stores/auth.js'
import { useNotificationStore } from './stores/notification.js'
import AppLoader from './components/AppLoader.vue'
import ConfirmModal from './modals/ConfirmModal.vue'
import ToastNotification from './components/ToastNotification.vue'
import { useWebSocket } from './composables/useWebSocket.js'

const auth = useAuthStore()
const notif = useNotificationStore()
const ws = useWebSocket()

function handleNotification(data) {
  notif.add(data)
}

onMounted(() => {
  auth.fetchUser()
  ws.onNotification(handleNotification)
})

onUnmounted(() => {
  ws.offNotification(handleNotification)
})
</script>

<template>
  <AppLoader />
  <ConfirmModal />
  <ToastNotification />
  <router-view />
</template>
