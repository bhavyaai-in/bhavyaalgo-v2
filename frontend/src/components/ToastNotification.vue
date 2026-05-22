<script setup>
import { useNotificationStore } from '../stores/notification.js'
const notif = useNotificationStore()
</script>

<template>
  <div class="toast-container">
    <TransitionGroup name="toast">
      <div v-for="n in notif.notifications" :key="n.id" class="toast" :class="n.type || 'info'">
        <div class="toast-body">
          <strong>{{ n.title }}</strong>
          <p v-if="n.message">{{ n.message }}</p>
        </div>
        <button class="toast-close" @click="notif.remove(n.id)">&times;</button>
      </div>
    </TransitionGroup>
  </div>
</template>

<style scoped>
.toast-container {
  position: fixed;
  top: 1rem;
  right: 1rem;
  z-index: 9999;
  display: flex;
  flex-direction: column;
  gap: .5rem;
  max-width: 360px;
  width: 100%;
  pointer-events: none;
}
.toast {
  display: flex;
  align-items: flex-start;
  gap: .5rem;
  padding: .75rem 1rem;
  border-radius: var(--radius);
  background: hsl(var(--card));
  border: 1px solid hsl(var(--border));
  box-shadow: 0 4px 12px rgba(0,0,0,.1);
  pointer-events: auto;
}
.toast.success { border-left: 3px solid #16A34A; }
.toast.error { border-left: 3px solid #DC2626; }
.toast.info { border-left: 3px solid #0284C7; }
.toast.warning { border-left: 3px solid #D97706; }
.toast-body { flex: 1; min-width: 0; }
.toast-body strong { display: block; font-size: var(--font-sm); color: hsl(var(--foreground)); }
.toast-body p { margin: .15rem 0 0; font-size: var(--font-xs); color: hsl(var(--muted-foreground)); }
.toast-close {
  background: none; border: none; cursor: pointer;
  font-size: 1.25rem; line-height: 1; padding: 0;
  color: hsl(var(--muted-foreground)); flex-shrink: 0;
}
.toast-close:hover { color: hsl(var(--foreground)); }

.toast-enter-active { transition: all .3s ease; }
.toast-leave-active { transition: all .3s ease; }
.toast-enter-from { opacity: 0; transform: translateX(100%); }
.toast-leave-to { opacity: 0; transform: translateX(100%); }
</style>
