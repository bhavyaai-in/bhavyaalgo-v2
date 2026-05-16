<script setup>
import { ref, onMounted, watch } from 'vue'
import { useRoute } from 'vue-router'

const route = useRoute()

const props = defineProps({
  title: String,
  endpoint: String,
})

const brokers = ref([])
const selectedId = ref(route.query.broker || '')
const data = ref(null)
const loading = ref(false)
const error = ref('')

function token() {
  return localStorage.getItem('token')
}

async function api(path, opts = {}) {
  const res = await fetch(path, {
    headers: { 'Content-Type': 'application/json', Authorization: token() },
    ...opts,
  })
  if (!res.ok) throw new Error(await res.text())
  return res.json()
}

async function fetchBrokers() {
  brokers.value = await api('/api/brokers')
  if (brokers.value.length > 0 && !selectedId.value) {
    selectedId.value = String(brokers.value[0].id)
  }
}

async function fetchData() {
  if (!selectedId.value) return
  loading.value = true
  error.value = ''
  data.value = null
  try {
    data.value = await api(`/api/broker-${props.endpoint}/${selectedId.value}`)
  } catch (e) {
    error.value = e.message
  } finally {
    loading.value = false
  }
}

watch(selectedId, () => {
  if (brokers.value.length) fetchData()
})
onMounted(fetchBrokers)

function cap(str) {
  if (!str) return ''
  return str.replace(/\b\w/g, c => c.toUpperCase())
}
</script>

<template>
  <div class="data-page">
    <header>
      <h2>{{ title }}</h2>
      <select v-model="selectedId" class="broker-select">
        <option v-for="b in brokers" :key="b.id" :value="b.id">
          {{ b.friendly_name || b.broker_name }}
        </option>
      </select>
    </header>

    <div v-if="loading" class="state-msg">Loading...</div>
    <div v-else-if="error" class="state-msg error">{{ error }}</div>
    <div v-else-if="!data || (Array.isArray(data) && data.length === 0)" class="state-msg">No data.</div>
    <div v-else class="table-wrap">
      <table v-if="Array.isArray(data)" class="data-table">
        <thead>
          <tr>
            <th v-for="col in Object.keys(data[0] || {})" :key="col">{{ cap(col.replace(/_/g, ' ')) }}</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="(row, ri) in data" :key="ri">
            <td v-for="col in Object.keys(data[0] || {})" :key="col">{{ row[col] ?? '-' }}</td>
          </tr>
        </tbody>
      </table>
      <table v-else class="data-table">
        <tr v-for="(v, k) in data" :key="k">
          <td class="pkey">{{ cap(k.replace(/_/g, ' ')) }}</td>
          <td>{{ v }}</td>
        </tr>
      </table>
    </div>
  </div>
</template>

<style scoped>
.data-page { padding: 1rem 0; }
header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 1.25rem;
}
header h2 { margin: 0; }
.broker-select {
  padding: .5rem 1rem;
  border: 1px solid hsl(var(--primary));
  border-radius: var(--radius);
  font-size: var(--font-sm);
  background: hsl(var(--primary));
  color: hsl(var(--primary-foreground));
  cursor: pointer;
  font-weight: 500;
}
.state-msg { text-align: center; padding: 2rem; color: hsl(var(--muted-foreground)); }
.state-msg.error { color: hsl(var(--destructive)); }
.table-wrap { overflow-x: auto; }
.data-table { width: 100%; border-collapse: collapse; font-size: var(--font-sm); }
.data-table th, .data-table td {
  padding: .5rem .6rem;
  border-bottom: 1px solid hsl(var(--border));
  text-align: left;
  white-space: nowrap;
}
.data-table th {
  font-weight: 600;
  color: hsl(var(--foreground));
  position: sticky;
  top: 0;
  background: hsl(var(--card));
}
.data-table td { color: hsl(var(--muted-foreground)); }
</style>
