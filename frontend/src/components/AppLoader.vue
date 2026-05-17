<script setup>
import { useUiStore } from '../stores/ui.js'

const ui = useUiStore()
</script>

<template>
  <div v-if="ui.isLoading" class="loader-bar" />
</template>

<style scoped>
.loader-bar {
  position: fixed;
  top: 0;
  left: 0;
  width: 100%;
  height: 3px;
  background: linear-gradient(90deg, hsl(var(--primary)), #16A34A, hsl(var(--primary)));
  background-size: 200% 100%;
  animation: shimmer 1s ease infinite;
  z-index: 9999;
}
@keyframes shimmer {
  0% { background-position: 200% 0; }
  100% { background-position: -200% 0; }
}
</style>
