<script setup>
import { onMounted } from 'vue'
import { useAuth } from './stores/auth.js'

const { fetchUser } = useAuth()

onMounted(() => {
  fetchUser()
})
</script>

<template>
  <router-view />
</template>
