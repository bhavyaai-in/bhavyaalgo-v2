<script setup>
import { ref, watch } from 'vue'

const props = defineProps({ show: Boolean, broker: Object })
const emit = defineEmits(['close'])

const data = ref(null)
const loading = ref(false)
const error = ref('')

function token() { return localStorage.getItem('token') }
async function api(path, opts = {}) {
  const res = await fetch(path, { headers: { 'Content-Type': 'application/json', Authorization: token() }, ...opts })
  if (!res.ok) throw new Error(await res.text())
  return res.json()
}
watch(() => props.show, async (val) => {
  if (!val || !props.broker) return
  loading.value = true; error.value = ''; data.value = null
  try { data.value = await api(`/api/broker-profile/${props.broker.id}`) }
  catch (e) { error.value = e.message }
  finally { loading.value = false }
})

function cap(str) { if (!str) return ''; return str.replace(/\b\w/g, c => c.toUpperCase()) }

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
.modal-overlay { position:fixed; inset:0; background:rgba(0,0,0,.4); display:flex; justify-content:center; align-items:center; z-index:100; }
.modal-box { background:hsl(var(--card)); border-radius:var(--radius); padding:1.5rem; width:90%; max-width:480px; max-height:80vh; overflow-y:auto; box-shadow:0 4px 24px rgba(0,0,0,.12); }
.modal-box h3 { margin:0 0 1rem; font-size:var(--font-base); }
.state-msg { text-align:center; padding:2rem; color:hsl(var(--muted-foreground)); }
.state-msg.error { color:hsl(var(--destructive)); }
.table-wrap { overflow-x:auto; }
.data-table { width:100%; border-collapse:collapse; font-size:var(--font-sm); }
.data-table td { padding:.5rem .75rem; border-bottom:1px solid hsl(var(--border)); }
.data-table .pkey { font-weight:600; color:hsl(var(--foreground)); width:40%; }
.data-table td:last-child { color:hsl(var(--muted-foreground)); }
.close-btn { display:block; margin:1rem auto 0; padding:.5rem 1.5rem; border:1px solid hsl(var(--border)); background:hsl(var(--card)); border-radius:var(--radius); cursor:pointer; font-weight:500; }
</style>
