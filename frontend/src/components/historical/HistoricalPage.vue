<script setup>
import { ref, computed, watch, onMounted } from 'vue'
import { api } from '../../utils/api.js'

const exchanges = ref([])
const selectedExchange = ref('NSE')
const underlyings = ref([])
const selectedSymbol = ref('NIFTY')
const intervals = ['1m', '3m', '5m', '10m', '15m', '30m', '1h', '1d']
const selectedInterval = ref('1d')
const fromDate = ref('')
const toDate = ref('')
const loading = ref(false)
const candles = ref([])
const stats = ref({ count: 0, price_change: 0 })
const error = ref('')

// Searchable symbol dropdown
const symbolSearch = ref('')
const symbolOpen = ref(false)
const symbolInput = ref(null)
const highlightIdx = ref(0)

const filteredSymbols = computed(() => {
  if (!symbolSearch.value) return underlyings.value
  const q = symbolSearch.value.toUpperCase()
  const matches = underlyings.value.filter(u => u.toUpperCase().includes(q))
  matches.sort((a, b) => {
    const pa = a.toUpperCase().startsWith(q)
    const pb = b.toUpperCase().startsWith(q)
    if (pa !== pb) return pa ? -1 : 1
    return a < b ? -1 : 1
  })
  return matches
})

function toggleSymbol() {
  symbolOpen.value = !symbolOpen.value
  if (symbolOpen.value) {
    symbolSearch.value = ''
    highlightIdx.value = 0
    setTimeout(() => symbolInput.value?.focus(), 100)
  }
}

function selectSymbol(s) {
  selectedSymbol.value = s
  symbolSearch.value = ''
  symbolOpen.value = false
}

function onSymbolKeydown(e) {
  if (e.key === 'ArrowDown') { e.preventDefault(); if (highlightIdx.value < filteredSymbols.value.length - 1) highlightIdx.value++ }
  else if (e.key === 'ArrowUp') { e.preventDefault(); if (highlightIdx.value > 0) highlightIdx.value-- }
  else if (e.key === 'Enter' || e.key === 'Tab') { e.preventDefault(); if (filteredSymbols.value.length) selectSymbol(filteredSymbols.value[highlightIdx.value]) }
  else if (e.key === 'Escape') { symbolOpen.value = false }
}

// Calculate date range based on interval
function updateDateRange() {
  const now = new Date()
  const to = now.toISOString().split('T')[0]
  toDate.value = to
  
  // Max days back based on interval (approximate)
  const maxDays = {
    '1m': 30, '3m': 60, '5m': 100, '10m': 100, '15m': 200,
    '30m': 200, '1h': 400, '1d': 2000  // Angel API max days per interval
  }
  const days = maxDays[selectedInterval.value] || 7
  const from = new Date(now.getTime() - days * 24 * 60 * 60 * 1000)
  fromDate.value = from.toISOString().split('T')[0]
}

watch(selectedInterval, updateDateRange)

async function loadExchanges() {
  try {
    exchanges.value = await api('/api/historical/exchanges')
    if (exchanges.value.length > 0 && !selectedExchange.value) selectedExchange.value = exchanges.value[0]
  } catch {}
}

async function loadUnderlyings() {
  const ex = typeof selectedExchange.value === 'object' ? selectedExchange.value.code : selectedExchange.value
  if (!ex) return
  try {
    underlyings.value = await api(`/api/historical/underlyings?exchange=${ex}`)
    // Set NIFTY as default if available
    if (underlyings.value.includes('NIFTY')) selectedSymbol.value = 'NIFTY'
    else if (underlyings.value.length > 0) selectedSymbol.value = underlyings.value[0]
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
        exchange: typeof selectedExchange.value === 'object' ? selectedExchange.value.code : selectedExchange.value,
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
function exchDisplay(ex) {
  if (typeof ex === 'object') return ex.code
  return ex
}
function exchLabel(ex) {
  if (typeof ex === 'object') return ex.name
  return ex
}

onMounted(async () => { 
  await loadExchanges()
  loadUnderlyings()
  updateDateRange() 
})

function formatPrice(v) { return Number(v).toFixed(2) }
function formatVol(v) { return v >= 100000 ? (v/100000).toFixed(1)+'L' : v >= 1000 ? (v/1000).toFixed(1)+'K' : String(v) }
function changeClass(v) { return v > 0 ? 'up' : v < 0 ? 'down' : '' }
</script>

<template>
  <div class="page">
    <header><h2>Historical Data Downloader</h2></header>

    <div class="controls">
      <select v-model="selectedExchange">
        <option v-for="e in exchanges" :key="exchDisplay(e)" :value="exchDisplay(e)">{{ exchLabel(e) }}</option>
      </select>

      <div class="search-wrap">
        <button v-if="!symbolOpen" class="search-btn" @click="toggleSymbol">{{ selectedSymbol || 'Select symbol...' }} <span class="arrow">▾</span></button>
        <input v-else ref="symbolInput" v-model="symbolSearch" placeholder="Search symbol..." class="search-input" @blur="setTimeout(() => symbolOpen = false, 200)" @keydown="onSymbolKeydown" />
        <div v-if="symbolOpen && filteredSymbols.length" class="search-dropdown">
          <div v-for="(s, i) in filteredSymbols" :key="s" class="search-item" :class="{ highlighted: i === highlightIdx }" @mousedown="selectSymbol(s)" @mouseenter="highlightIdx = i">{{ s }}</div>
        </div>
      </div>

      <select v-model="selectedInterval">
        <option v-for="i in intervals" :key="i" :value="i">{{ i }}</option>
      </select>
      <label>From: <input v-model="fromDate" type="date" class="date-input" /></label>
      <label>To: <input v-model="toDate" type="date" class="date-input" /></label>
      <button class="chip primary" @click="download" :disabled="loading">{{ loading ? 'Downloading...' : 'Download' }}</button>
    </div>

    <div v-if="error" class="error">{{ error }}</div>
    <div v-if="candles.length === 0 && !loading && fromDate && selectedSymbol && selectedExchange && !error" class="warning">
      No data returned. Try a different date range or symbol.
    </div>

    <div v-if="candles.length" class="stats">
      <div class="card"><span class="card-label">Records</span><span class="card-value">{{ stats.count }}</span></div>
      <div class="card"><span class="card-label">Price Change</span><span class="card-value" :class="changeClass(stats.price_change)">{{ stats.price_change > 0 ? '+' : '' }}{{ formatPrice(stats.price_change) }}</span></div>
      <div class="card"><span class="card-label">{{ selectedSymbol }}</span><span class="card-value">{{ selectedInterval }}</span><span class="card-sub">{{ selectedExchange }}</span></div>
    </div>

    <div v-if="candles.length" class="table-wrap">
      <table>
        <thead><tr><th>Timestamp</th><th class="rt">Open</th><th class="rt">High</th><th class="rt">Low</th><th class="rt">Close</th><th class="rt">Volume</th></tr></thead>
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
.page {}
header h2 { margin:0 0 1rem; }
.controls { display:flex; flex-wrap:wrap; gap:.5rem; margin-bottom:1rem; align-items:center; }
.controls select, .date-input { padding:.35rem .55rem; border:1px solid hsl(var(--input)); border-radius:var(--radius); font-size:var(--font-sm); background:hsl(var(--card)); outline:none; }
.controls label { display:flex; align-items:center; gap:4px; font-size:var(--font-sm); color:hsl(var(--muted-foreground)); }
.date-input { width:140px; }
.chip { padding:.25rem .6rem; border-radius:var(--radius); font-size:var(--font-xs); cursor:pointer; background:transparent; color:hsl(var(--muted-foreground)); border:1px solid hsl(var(--border)); }
.chip.primary { background:hsl(var(--primary)); color:#fff; border-color:hsl(var(--primary)); font-weight:600; }
.chip.primary:disabled { opacity:.5; cursor:not-allowed; }
.search-wrap { position:relative; min-width:160px; }
.search-btn { display:flex; align-items:center; gap:4px; cursor:pointer; color:hsl(var(--foreground)); padding:.35rem .55rem; border:1px solid hsl(var(--input)); border-radius:var(--radius); font-size:var(--font-sm); background:hsl(var(--card)); outline:none; width:100%; }
.search-btn .arrow { font-size:10px; color:hsl(var(--muted-foreground)); margin-left:auto; }
.search-input { padding:.35rem .55rem; border:1px solid hsl(var(--input)); border-radius:var(--radius); font-size:var(--font-sm); background:hsl(var(--card)); outline:none; width:100%; box-sizing:border-box; }
.search-input:focus { border-color:hsl(var(--ring)); }
.search-dropdown { position:absolute; top:100%; left:0; right:0; z-index:20; background:hsl(var(--card)); border:1px solid hsl(var(--border)); border-radius:var(--radius); box-shadow:0 4px 12px rgba(0,0,0,.1); margin-top:2px; max-height:240px; overflow-y:auto; }
.search-item { padding:.35rem .55rem; cursor:pointer; font-size:var(--font-sm); color:hsl(var(--foreground)); }
.search-item:hover { background:hsl(var(--muted)); }
.search-item.highlighted { background:hsl(var(--primary)/.15); }
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
.warning { color:hsl(var(--warning, 38 92% 50%)); margin-bottom:.5rem; font-size:var(--font-sm); padding:.5rem; background:hsl(var(--card)); border:1px solid hsl(var(--warning, 38 92% 50%)/.3); border-radius:var(--radius); }
</style>
