<script setup>
import { ref, watch } from 'vue'
import { api } from '../../utils/api.js'

const props = defineProps({ show: Boolean, broker: Object })
const emit = defineEmits(['close'])

const data = ref(null)
const loading = ref(false)
const error = ref('')

function cap(str) { if (!str) return ''; return str.replace(/\b\w/g, c => c.toUpperCase()) }
watch(() => props.show, async (val) => {
  if (!val || !props.broker) return
  loading.value = true; error.value = ''; data.value = null
  try { data.value = await api(`/api/broker-profile/${props.broker.id}`) }
  catch (e) { error.value = e.message }
  finally { loading.value = false }
})

function flatRows(obj) {
  if (!obj) return []
  function flat(o) {
    const rows = []
    for (const [k, v] of Object.entries(o || {})) {
      if (k.startsWith('_')) continue
      if (v && typeof v === 'object' && !Array.isArray(v)) { rows.push(...flat(v)) }
      else if (v !== '') { rows.push({ key: cap(k.replace(/_/g, ' ')), value: Array.isArray(v) ? v.join(', ') : String(v ?? '') }) }
    }
    return rows
  }
  if (obj.data && typeof obj.data === 'object' && Object.keys(obj.data).length) return flat(obj.data)
  return flat(obj)
}
</script>

<template>
  <div v-if="show" class="modal-overlay" @click.self="emit('close')">
    <div class="modal-box">
      <h3>Profile — {{ cap(broker?.friendly_name || broker?.broker_name || '') }}</h3>
      <div v-if="loading" class="state-msg">Loading...</div>
      <div v-else-if="error" class="state-msg error">{{ error }}</div>
      <div v-else class="table-wrap">
        <table class="data-table kv">
          <tr v-for="row in flatRows(data)" :key="row.key"><td class="pkey">{{ row.key }}</td><td>{{ row.value }}</td></tr>
        </table>
      </div>
      <button class="close-btn" @click="emit('close')">Close</button>
    </div>
  </div>
</template>

<style scoped>
.modal-box { max-width:480px; }
</style>
