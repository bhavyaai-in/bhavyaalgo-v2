<script setup>
import { onMounted } from 'vue'
import { useAuthStore } from './stores/auth.js'
import AppLoader from './components/AppLoader.vue'
import ConfirmModal from './modals/ConfirmModal.vue'

const auth = useAuthStore()

onMounted(() => {
  auth.fetchUser()
})
</script>

<template>
  <AppLoader />
  <ConfirmModal />
  <router-view />
</template>
