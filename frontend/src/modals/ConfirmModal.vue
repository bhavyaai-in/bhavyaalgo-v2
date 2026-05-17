<script setup>
import { useUiStore } from '../stores/ui.js'

const ui = useUiStore()
</script>

<template>
  <div v-if="ui.confirm" class="modal-overlay">
    <div class="confirm-box">
      <h3>{{ ui.confirm.title }}</h3>
      <p>{{ ui.confirm.message }}</p>
      <div class="actions">
        <button class="yes" @click="ui.resolveConfirm(true)">Yes</button>
        <button class="no" @click="ui.resolveConfirm(false)">No</button>
      </div>
    </div>
  </div>
</template>

<style scoped>
.modal-overlay {
  position: fixed; inset: 0;
  background: rgba(0,0,0,.4);
  display: flex; justify-content: center; align-items: center;
  z-index: 9998;
}
.confirm-box {
  background: hsl(var(--card));
  padding: 2rem;
  border-radius: var(--radius);
  width: 90%;
  max-width: 360px;
  box-shadow: 0 4px 24px rgba(0,0,0,.12);
}
.confirm-box h3 { margin: 0 0 .5rem; }
.confirm-box p { margin: 0 0 1.5rem; color: hsl(var(--muted-foreground)); }
.actions { display: flex; gap: .5rem; }
.actions button {
  flex: 1; padding: .6rem; border: none; border-radius: var(--radius);
  cursor: pointer; font-weight: 500;
}
.yes { background: hsl(var(--destructive)); color: #fff; }
.no { background: hsl(var(--muted)); color: hsl(var(--foreground)); }
</style>
