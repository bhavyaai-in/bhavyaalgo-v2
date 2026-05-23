<script setup>
import { ref, onMounted, watch } from 'vue'
import { useRoute } from 'vue-router'
import BrokerState from '../components/brokers/BrokerState.vue'

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

function stateType() {
  if (loading.value) return 'loading'
  if (error.value) {
    if (error.value === 'could not connect broker' || error.value === 'broker token not generated') {
      return 'disconnected'
    }
    return 'error'
  }
  if (!data.value || (Array.isArray(data.value) && data.value.length === 0)) return 'empty'
  return null
}

function stateMessage() {
  if (loading.value) return 'Loading...'
  if (error.value) return error.value
  if (!data.value || (Array.isArray(data.value) && data.value.length === 0)) return 'No data.'
  return ''
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

    <BrokerState v-if="stateType()" :type="stateType()" :message="stateMessage()" />
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
