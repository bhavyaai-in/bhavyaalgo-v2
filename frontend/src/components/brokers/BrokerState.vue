<script setup>
defineProps({
  type: { type: String, default: 'empty' },
  message: { type: String, default: '' },
})
</script>

<template>
  <div class="state-msg" :class="type">
    <svg v-if="type === 'disconnected'" class="state-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round">
      <path d="M12 22v-5"/>
      <path d="M9 10V2"/>
      <path d="M15 10V2"/>
      <path d="M18 10v4a6 6 0 0 1-12 0v-4"/>
      <line x1="2" x2="22" y1="2" y2="22"/>
    </svg>
    <svg v-else-if="type === 'empty'" class="state-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round">
      <path d="M13 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V9z"/>
      <polyline points="13 2 13 9 20 9"/>
    </svg>
    <svg v-else-if="type === 'loading'" class="state-icon spin" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round">
      <path d="M21 12a9 9 0 1 1-6.219-8.56"/>
    </svg>
    <svg v-else-if="type === 'error'" class="state-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round">
      <circle cx="12" cy="12" r="10"/>
      <line x1="15" y1="9" x2="9" y2="15"/>
      <line x1="9" y1="9" x2="15" y2="15"/>
    </svg>
    <span class="state-text">{{ message }}</span>
  </div>
</template>

<style scoped>
.state-msg {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 0.75rem;
  padding: 2.5rem 1rem;
  text-align: center;
}
.state-msg.empty,
.state-msg.disconnected {
  color: hsl(var(--muted-foreground));
}
.state-msg.error {
  color: hsl(var(--destructive));
}
.state-msg.loading {
  color: hsl(var(--muted-foreground));
}
.state-icon {
  width: 36px;
  height: 36px;
  opacity: 0.7;
}
.state-text {
  font-size: var(--font-sm);
  line-height: 1.5;
  max-width: 280px;
}
.spin {
  animation: spin 1s linear infinite;
}
@keyframes spin {
  to { transform: rotate(360deg); }
}
</style>
