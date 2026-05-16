<script setup>
defineProps({ show: Boolean, title: String, broker: Object })
const emit = defineEmits(['close'])
function cap(str) { if (!str) return ''; return str.replace(/\b\w/g, c => c.toUpperCase()) }
</script>

<template>
  <div v-if="show" class="modal-overlay" @click.self="emit('close')">
    <div class="modal-box">
      <h3>{{ title }} — {{ cap(broker?.friendly_name || broker?.broker_name || '') }}</h3>
      <slot />
      <button class="close-btn" @click="emit('close')">Close</button>
    </div>
  </div>
</template>
