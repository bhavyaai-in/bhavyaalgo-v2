<script setup>
import { ref, computed, watch, onMounted } from 'vue'
import { api } from '../../utils/api.js'

const exchanges = ref([])
const selectedExchange = ref('')
const underlyings = ref([])
const selectedSymbol = ref('')
const intervals = ['1m', '3m', '5m', '10m', '15m', '30m', '1h', '1d']
const selectedInterval = ref('15m')
const fromDate = ref('')
const toDate = ref('')
const loading = ref(false)
const candles = ref([])
const stats = ref({ count: 0, price_change: 0 })
const error = ref('')

// Set default dates (last 7 days)
function setDefaultDates() {
  const now = new Date()
  const to = now.toISOString().split('T')[0]
  const from = new Date(now.getTime() - 7 * 24 * 60 * 60 * 1000).toISOString().split('T')[0]
  fromDate.value = from
  toDate.value = to
}
setDefaultDates()

async function loadExchanges() {
  try {
    exchanges.value = await api('/api/historical/exchanges')
    if (exchanges.value.length > 0) selectedExchange.value = exchanges.value[0]
  } catch {}
}

async function loadUnderlyings() {
  if (!selectedExchange.value) return
  try {
    underlyings.value = await api(`/api/historical/underlyings?exchange=${selectedExchange.value}`)
    if (underlyings.value.length > 0) selectedSymbol.value = underlyings.value[0]
  } catch {
    underlyings.value = []
  }
}

async function download() {
  if (!selectedSymbol.value || !selectedExchange.value || !selectedInterval.value || !fromDate.value || !toDate.value) return
  loading.value = true
  error.value = ''
  candles.value = []
  try {
    const res = await api('/api/historical/download', {
      method: 'POST',
      body: JSON.stringify({
        symbol: selectedSymbol.value,
        exchange: selectedExchange.value,
        interval: selectedInterval.value,
        from: fromDate.value + ' 09:15',
        to: toDate.value + ' 15:30',
      }),
    })
    candles.value = res.candles || []
    stats.value = { count: res.count || 0, price_change: res.price_change || 0 }
  } catch (e) {
    error.value = e.message
  } finally {
    loading.value = false
  }
}

watch(selectedExchange, () => { loadUnderlyings() })
onMounted(() => { loadExchanges() })

function formatPrice(v) { return Number(v).toFixed(2) }
function formatVol(v) { return v >= 100000 ? (v/100000).toFixed(1)+'L' : v >= 1000 ? (v/1000).toFixed(1)+'K' : String(v) }
function changeClass(v) { return v > 0 ? 'up' : v < 0 ? 'down' : '' }
</script>

<template>
  <div class="page">
    <header><h2>Historical Data Downloader</h2></header>

    <div class="controls">
      <select v-model="selectedExchange">
        <option value="">Select exchange</option>
        <option v-for="e in exchanges" :key="e" :value="e">{{ e }}</option>
      </select>
      <select v-model="selectedSymbol">
        <option value="">Select symbol</option>
        <option v-for="s in underlyings" :key="s" :value="s">{{ s }}</option>
      </select>
      <select v-model="selectedInterval">
        <option v-for="i in intervals" :key="i" :value="i">{{ i }}</option>
      </select>
      <label>From: <input v-model="fromDate" type="date" class="date-input" /></label>
      <label>To: <input v-model="toDate" type="date" class="date-input" /></label>
      <button class="chip primary" @click="download" :disabled="loading">{{ loading ? 'Downloading...' : 'Download' }}</button>
    </div>

    <div v-if="error" class="error">{{ error }}</div>

    <div v-if="candles.length" class="stats">
      <div class="card">
        <span class="card-label">Records</span>
        <span class="card-value">{{ stats.count }}</span>
      </div>
      <div class="card">
        <span class="card-label">Price Change</span>
        <span class="card-value" :class="changeClass(stats.price_change)">{{ stats.price_change > 0 ? '+' : '' }}{{ formatPrice(stats.price_change) }}</span>
      </div>
      <div class="card">
        <span class="card-label">{{ selectedSymbol }}</span>
        <span class="card-value">{{ selectedInterval }}</span>
        <span class="card-sub">{{ selectedExchange }}</span>
      </div>
    </div>

    <div v-if="candles.length" class="table-wrap">
      <table>
        <thead>
          <tr>
            <th>Timestamp</th>
            <th class="rt">Open</th>
            <th class="rt">High</th>
            <th class="rt">Low</th>
            <th class="rt">Close</th>
            <th class="rt">Volume</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="c in candles" :key="c.timestamp">
            <td class="ts">{{ c.timestamp }}</td>
            <td class="rt">{{ formatPrice(c.open) }}</td>
            <td class="rt">{{ formatPrice(c.high) }}</td>
            <td class="rt">{{ formatPrice(c.low) }}</td>
            <td class="rt" :class="changeClass(c.close - c.open)">{{ formatPrice(c.close) }}</td>
            <td class="rt vol">{{ formatVol(c.volume) }}</td>
          </tr>
        </tbody>
      </table>
    </div>
  </div>
</template>

<style scoped>
.page { }
header h2 { margin:0 0 1rem; }
.controls { display:flex; flex-wrap:wrap; gap:.5rem; margin-bottom:1rem; align-items:center; }
.controls select, .controls button, .date-input {
  padding:.35rem .55rem; border:1px solid hsl(var(--input)); border-radius:var(--radius);
  font-size:var(--font-sm); background:hsl(var(--card)); outline:none;
}
.controls label { display:flex; align-items:center; gap:4px; font-size:var(--font-sm); color:hsl(var(--muted-foreground)); }
.date-input { width:140px; }
.chip { padding:.25rem .6rem; border-radius:var(--radius); font-size:var(--font-xs); cursor:pointer; background:transparent; color:hsl(var(--muted-foreground)); border:1px solid hsl(var(--border)); }
.chip.primary { background:hsl(var(--primary)); color:#fff; border-color:hsl(var(--primary)); font-weight:600; }
.chip.primary:disabled { opacity:.5; cursor:not-allowed; }
.stats { display:flex; gap:.75rem; margin-bottom:1rem; flex-wrap:wrap; }
.card { flex:1; min-width:120px; background:hsl(var(--card)); border:1px solid hsl(var(--border)); border-radius:var(--radius); padding:.6rem .8rem; display:flex; flex-direction:column; gap:2px; }
.card-label { font-size:var(--font-xs); color:hsl(var(--muted-foreground)); }
.card-value { font-size:var(--font-lg); font-weight:700; }
.card-value.up { color:#16A34A; }
.card-value.down { color:hsl(0 84% 60%); }
.card-sub { font-size:var(--font-xs); color:hsl(var(--muted-foreground)); }
.table-wrap { overflow-x:auto; border:1px solid hsl(var(--border)); border-radius:var(--radius); max-height:60vh; overflow-y:auto; }
table { width:100%; border-collapse:collapse; font-size:var(--font-sm); }
th, td { padding:.35rem .5rem; white-space:nowrap; border-bottom:1px solid hsl(var(--border)/.5); }
th { font-weight:600; font-size:var(--font-xs); color:hsl(var(--muted-foreground)); position:sticky; top:0; background:hsl(var(--card)); }
.rt { text-align:right; font-family:monospace; }
.ts { font-family:monospace; font-size:var(--font-xs); color:hsl(var(--muted-foreground)); }
.vol { color:hsl(var(--muted-foreground)); }
tr:hover td { background:hsl(var(--muted)/.3); }
.error { color:hsl(var(--destructive)); margin-bottom:.5rem; }
</style>
