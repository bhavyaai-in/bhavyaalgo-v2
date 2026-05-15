<script setup>
import { ref, onMounted } from 'vue'

const message = ref('Loading...')

onMounted(async () => {
  try {
    const res = await fetch('/api/hello')
    const data = await res.json()
    message.value = data.text
  } catch {
    message.value = 'Could not reach backend'
  }
})
</script>

<template>
  <main>
    <h1>{{ message }}</h1>
  </main>
</template>

<style>
body {
  margin: 0;
  font-family: system-ui, sans-serif;
  display: flex;
  place-items: center;
  min-height: 100vh;
}
main {
  max-width: 800px;
  margin: 0 auto;
  padding: 2rem;
  text-align: center;
}
h1 {
  font-size: 2.5rem;
  color: #333;
}
</style>
