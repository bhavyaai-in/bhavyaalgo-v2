<script setup>
import { ref, onMounted, watch } from 'vue'
import { useRoute } from 'vue-router'

const route = useRoute()
const brokers = ref([])
const selectedId = ref(route.query.broker || '')
const margin = ref(null)
const loading = ref(false)
const error = ref('')

function token() { return localStorage.getItem('token') }
async function api(path, opts = {}) {
  const res = await fetch(path, { headers: { 'Content-Type': 'application/json', Authorization: token() }, ...opts })
  if (!res.ok) throw new Error(await res.text())
  return res.json()
}
async function fetchBrokers() {
  brokers.value = await api('/api/brokers')
  if (brokers.value.length > 0 && !selectedId.value) selectedId.value = String(brokers.value[0].id)
}
async function fetchData() {
  if (!selectedId.value) return
  loading.value = true; error.value = ''; margin.value = null
  try { margin.value = await api(`/api/broker-margin/${selectedId.value}`) }
  catch (e) { error.value = e.message }
  finally { loading.value = false }
}
watch(selectedId, () => { if (brokers.value.length) fetchData() })
onMounted(fetchBrokers)

function cap(str) {
  if (!str) return ''
  return str.replace(/\b\w/g, c => c.toUpperCase())
}
</script>

<template>
  <div class="page">
    <header>
      <h2>Margin</h2>
      <select v-model="selectedId" class="broker-select">
        <option v-for="b in brokers" :key="b.id" :value="b.id">{{ b.friendly_name || b.broker_name }}</option>
      </select>
    </header>
    <div v-if="loading" class="state-msg">Loading...</div>
    <div v-else-if="error" class="state-msg error">{{ error }}</div>
    <div v-else-if="!margin" class="state-msg">No margin data.</div>
    <div v-else class="table-wrap">
      <table class="data-table kv">
        <tr v-for="(v,k) in margin" :key="k"><td class="pkey">{{ cap(k.replace(/_/g,' ')) }}</td><td>{{ v }}</td></tr>
      </table>
    </div>
  </div>
</template>

<style scoped>
.page { padding: 1rem 0; }
header { display:flex; justify-content:space-between; align-items:center; margin-bottom:1.25rem; }
header h2 { margin:0; }
.broker-select { padding:.5rem 1rem; border:1px solid hsl(var(--primary)); border-radius:var(--radius); font-size:var(--font-sm); background:hsl(var(--primary)); color:hsl(var(--primary-foreground)); cursor:pointer; font-weight:500; }
.state-msg { text-align:center; padding:2rem; color:hsl(var(--muted-foreground)); }
.state-msg.error { color:hsl(var(--destructive)); }
.table-wrap { overflow-x:auto; }
.data-table { width:100%; border-collapse:collapse; font-size:var(--font-sm); }
.data-table td { padding:.5rem .75rem; border-bottom:1px solid hsl(var(--border)); }
.data-table .pkey { font-weight:600; color:hsl(var(--foreground)); width:40%; }
.data-table td:last-child { color:hsl(var(--muted-foreground)); }
</style>
