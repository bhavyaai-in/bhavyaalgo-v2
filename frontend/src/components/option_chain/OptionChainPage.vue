<script setup>
import { ref, computed, watch, onMounted } from 'vue'
import { api } from '../../utils/api.js'
import { useWebSocket } from '../../composables/useWebSocket.js'

const ws = useWebSocket()

const exchanges = ['NFO', 'BFO']
const selectedExchange = ref('NFO')
const underlyings = ref([])
const selectedUnderlying = ref('')
const underlyingSearch = ref('')
const underlyingOpen = ref(false)
const underlyingInput = ref(null)
const highlightIdx = ref(0)
const expiries = ref([])
const selectedExpiry = ref('')
const strikeCount = ref(10)
const chain = ref([])
const loading = ref(false)
const error = ref('')
const underlyingLTP = ref(0)
const underlyingClose = ref(0)
const atmStrike = ref(0)
const pcr = ref(0)
const ltpMap = ref({})

// Live tick subscription
ws.onTick((tick) => {
  if (tick.token && tick.ltp != null) {
    ltpMap.value = { ...ltpMap.value, [tick.token]: tick }
    if (tick.token_999) ltpMap.value[tick.token_999] = ltpMap.value[tick.token]
  }
})

const strikeCounts = [5, 10, 15, 20, 25]

watch(underlyingSearch, () => { highlightIdx.value = 0 })

const filteredUnderlyings = computed(() => {
  if (!underlyingSearch.value) return underlyings.value
  const q = underlyingSearch.value.toUpperCase()
  const matches = underlyings.value.filter(u => u.toUpperCase().includes(q))
  // Sort: exact prefix match first, then alphabetical
  matches.sort((a, b) => {
    const pa = a.toUpperCase().startsWith(q)
    const pb = b.toUpperCase().startsWith(q)
    if (pa !== pb) return pa ? -1 : 1
    return a < b ? -1 : 1
  })
  return matches
})

function toggleUnderlying() {
  underlyingOpen.value = !underlyingOpen.value
  if (underlyingOpen.value) {
    underlyingSearch.value = ''
    highlightIdx.value = 0
    setTimeout(() => underlyingInput.value?.focus(), 100)
  }
}

function onKeydown(e) {
  if (e.key === 'ArrowDown') {
    e.preventDefault()
    if (highlightIdx.value < filteredUnderlyings.value.length - 1) highlightIdx.value++
  } else if (e.key === 'ArrowUp') {
    e.preventDefault()
    if (highlightIdx.value > 0) highlightIdx.value--
  } else if (e.key === 'Enter' || e.key === 'Tab') {
    if (filteredUnderlyings.value.length > 0) {
      selectUnderlying(filteredUnderlyings.value[highlightIdx.value])
    }
  }
}
function selectUnderlying(u) {
  selectedUnderlying.value = u
  underlyingSearch.value = ''
  underlyingOpen.value = false
  expiries.value = []
  selectedExpiry.value = ''
  loadExpiries()
}

function onUnderlyingChange() {
  expiries.value = []
  selectedExpiry.value = ''
  loadExpiries()
}

async function loadUnderlyings() {
  try {
    underlyings.value = await api(`/api/option-chain/underlyings?exchange=${selectedExchange.value}`)
    if (underlyings.value.length > 0) {
      selectedUnderlying.value = underlyings.value[0]
    }
  } catch {}
}

async function loadExpiries() {
  if (!selectedUnderlying.value) return
  try {
    expiries.value = await api(`/api/option-chain/expiries?exchange=${selectedExchange.value}&underlying=${encodeURIComponent(selectedUnderlying.value)}`)
    if (expiries.value.length > 0) {
      selectedExpiry.value = expiries.value[0]
    }
  } catch {}
}

async function loadChain() {
  if (!selectedUnderlying.value || !selectedExpiry.value) return
  loading.value = true
  error.value = ''
  try {
    // Convert expiry DD-MMM-YY to DDMMMYY for the API
    const expiryAPI = selectedExpiry.value.replace(/-/g, '')
    const res = await api(`/api/option-chain?underlying=${encodeURIComponent(selectedUnderlying.value)}&exchange=${encodeURIComponent(selectedExchange.value)}&expiry=${encodeURIComponent(expiryAPI)}&strike_count=${strikeCount.value}`)
    chain.value = res.chain || []
    underlyingLTP.value = res.underlying_ltp || 0
    underlyingClose.value = res.underlying_close || 0
    atmStrike.value = res.atm_strike || 0
    pcr.value = res.pcr || 0

    // Subscribe all option tokens for live updates
    const tokens = []
    chain.value.forEach(item => {
      if (item.ce?.symbol) tokens.push({ exchange: selectedExchange.value, symbol: item.ce.symbol, token: item.ce.token })
      if (item.pe?.symbol) tokens.push({ exchange: selectedExchange.value, symbol: item.pe.symbol, token: item.pe.token })
    })
    if (tokens.length > 0) {
      ws.subscribe(tokens.map(t => t.exchange + '|' + t.token))
    }
  } catch (e) {
    error.value = e.message
  } finally {
    loading.value = false
  }
}

watch(selectedExchange, () => { loadUnderlyings() })
watch(selectedUnderlying, () => { loadExpiries() })
watch([selectedExpiry, strikeCount], () => { loadChain() })

onMounted(() => {
  loadUnderlyings()
})

function formatPrice(v) { return v != null && v > 0 ? Number(v).toFixed(2) : '-' }
function formatLakhs(v) {
  if (!v || v <= 0) return '-'
  if (v >= 100000) return (v / 100000).toFixed(1) + 'L'
  if (v >= 1000) return (v / 1000).toFixed(1) + 'K'
  return String(Math.round(v))
}

const liveChain = computed(() => {
  return chain.value.map(item => {
    const ce = item.ce ? { ...item.ce } : null
    const pe = item.pe ? { ...item.pe } : null
    // Merge WebSocket live data
    if (ce) {
      const tick = ltpMap.value[ce.token]
      if (tick?.ltp != null) { ce.ltp = tick.ltp; ce.bid = tick.bid ?? ce.bid; ce.ask = tick.ask ?? ce.ask }
      if (tick?.oi != null) ce.oi = tick.oi
      if (tick?.volume != null) ce.volume = tick.volume
    }
    if (pe) {
      const tick = ltpMap.value[pe.token]
      if (tick?.ltp != null) { pe.ltp = tick.ltp; pe.bid = tick.bid ?? pe.bid; pe.ask = tick.ask ?? pe.ask }
      if (tick?.oi != null) pe.oi = tick.oi
      if (tick?.volume != null) pe.volume = tick.volume
    }
    return { ...item, ce, pe }
  })
})

function strikeLabel(strike) {
  if (strike === atmStrike.value) return 'ATM'
  const idx = chain.value.findIndex(c => c.strike === strike)
  return chain.value[idx]?.ce?.label || chain.value[idx]?.pe?.label || ''
}

function changeClass(item, side, strike) {
  const ltp = item[side]?.ltp ?? 0
  const close = item[side]?.close ?? 0
  if (!ltp || !close) return ''
  return ltp >= close ? 'up' : 'down'
}
</script>

<template>
  <div class="page">
    <header>
      <h2>Option Chain</h2>
    </header>

    <!-- Controls -->
    <div class="controls">
      <select v-model="selectedExchange">
        <option v-for="ex in exchanges" :key="ex" :value="ex">{{ ex }}</option>
      </select>
      <div class="search-wrap">
        <input v-if="underlyingOpen" ref="underlyingInput" v-model="underlyingSearch" placeholder="Search underlying..." class="search-input" @blur="setTimeout(() => underlyingOpen = false, 200)" @keydown="onKeydown" />
        <button v-else class="search-btn" @click="toggleUnderlying">{{ selectedUnderlying || 'Select underlying...' }} <span class="arrow">▾</span></button>
        <div v-if="underlyingOpen && filteredUnderlyings.length" class="search-dropdown">
          <div v-for="(u, i) in filteredUnderlyings" :key="u" class="search-item" :class="{ highlighted: i === highlightIdx }" @mousedown="selectUnderlying(u)" @mouseenter="highlightIdx = i">{{ u }}</div>
        </div>
        <div v-if="underlyingOpen && !filteredUnderlyings.length" class="search-dropdown">
          <div class="search-empty">No matches</div>
        </div>
      </div>
      <select v-model="selectedExpiry" @change="loadChain">
        <option v-if="!expiries.length" value="">Loading...</option>
        <option v-for="e in expiries" :key="e" :value="e">{{ e }}</option>
      </select>
      <select v-model.number="strikeCount">
        <option v-for="n in strikeCounts" :key="n" :value="n">{{ n }} strikes</option>
      </select>
      <button class="chip primary" @click="loadChain" :disabled="loading">{{ loading ? 'Loading...' : 'Refresh' }}</button>
    </div>

    <!-- Summary cards -->
    <div v-if="chain.length" class="summary">
      <div class="card">
        <span class="card-label">{{ selectedUnderlying }}</span>
        <span class="card-value primary">{{ formatPrice(underlyingLTP) }}</span>
        <span class="card-sub">Close: {{ formatPrice(underlyingClose) }}</span>
      </div>
      <div class="card">
        <span class="card-label">ATM Strike</span>
        <span class="card-value">{{ atmStrike }}</span>
        <span class="card-sub">{{ selectedExpiry }}</span>
      </div>
      <div class="card">
        <span class="card-label">PCR (OI)</span>
        <span class="card-value" :class="pcr > 1 ? 'green' : 'yellow'">{{ pcr.toFixed(2) }}</span>
        <span class="card-sub">Put/Call Ratio</span>
      </div>
    </div>

    <!-- Chain table -->
    <div v-if="chain.length" class="table-wrap">
      <table>
        <thead>
          <tr class="section-hdr">
            <th colspan="4" class="calls-hdr">CALLS</th>
            <th class="strike-hdr">Strike</th>
            <th colspan="5" class="puts-hdr">PUTS</th>
          </tr>
          <tr class="col-hdr">
            <th class="rt">OI</th>
            <th class="rt">LTP</th>
            <th class="rt">Bid</th>
            <th class="rt">Ask</th>
            <th class="strike-col"></th>
            <th class="lt">Ask</th>
            <th class="lt">Bid</th>
            <th class="lt">LTP</th>
            <th class="lt">OI</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="item in liveChain" :key="item.strike" class="chain-row">
            <!-- CE side -->
            <td class="rt oi-cell">{{ item.ce ? formatLakhs(item.ce.oi) : '-' }}</td>
            <td class="rt ltp-cell" :class="changeClass(item, 'ce', item.strike)">{{ formatPrice(item.ce?.ltp) }}</td>
            <td class="rt">{{ formatPrice(item.ce?.bid) }}</td>
            <td class="rt">{{ formatPrice(item.ce?.ask) }}</td>

            <!-- Strike -->
            <td class="strike-cell" :class="{ 'atm': strikeLabel(item.strike) === 'ATM' }">
              <div class="strike-inner">
                <span class="strike-label" v-if="item.ce?.label">{{ item.ce.label }}</span>
                <span class="strike-val">{{ item.strike }}</span>
                <span class="strike-label" v-if="item.pe?.label">{{ item.pe.label }}</span>
              </div>
            </td>

            <!-- PE side -->
            <td class="lt">{{ formatPrice(item.pe?.ask) }}</td>
            <td class="lt">{{ formatPrice(item.pe?.bid) }}</td>
            <td class="lt ltp-cell" :class="changeClass(item, 'pe', item.strike)">{{ formatPrice(item.pe?.ltp) }}</td>
            <td class="lt oi-cell">{{ item.pe ? formatLakhs(item.pe.oi) : '-' }}</td>
          </tr>
        </tbody>
      </table>
    </div>

    <div v-if="error" class="error">{{ error }}</div>
    <div v-if="loading && !chain.length" class="loading">Loading...</div>
  </div>
</template>

<style scoped>
.page { }
header { margin-bottom:1rem; }
header h2 { margin:0; }

.controls { display:flex; flex-wrap:wrap; gap:.5rem; margin-bottom:1rem; align-items:center; }
.controls select, .controls button {
  padding:.4rem .6rem; border:1px solid hsl(var(--input)); border-radius:var(--radius);
  font-size:var(--font-sm); background:hsl(var(--card)); outline:none;
}
.controls select:focus { border-color:hsl(var(--ring)); }
.search-wrap { position:relative; min-width:160px; }
.search-btn {
  display:flex; align-items:center; gap:4px; cursor:pointer; color:hsl(var(--foreground));
  padding:.4rem .6rem; border:1px solid hsl(var(--input)); border-radius:var(--radius);
  font-size:var(--font-sm); background:hsl(var(--card)); outline:none; width:100%;
}
.search-btn .arrow { font-size:10px; color:hsl(var(--muted-foreground)); margin-left:auto; }
.search-input {
  padding:.4rem .6rem; border:1px solid hsl(var(--input)); border-radius:var(--radius);
  font-size:var(--font-sm); background:hsl(var(--card)); outline:none; width:100%;
  box-sizing:border-box;
}
.search-input:focus { border-color:hsl(var(--ring)); }
.search-dropdown {
  position:absolute; top:100%; left:0; right:0; z-index:20;
  background:hsl(var(--card)); border:1px solid hsl(var(--border)); border-radius:var(--radius);
  box-shadow:0 4px 12px rgba(0,0,0,.1); margin-top:2px;
}

.search-list { max-height:200px; overflow-y:auto; }
.search-item {
  padding:.35rem .5rem; cursor:pointer; font-size:var(--font-sm); color:hsl(var(--foreground)); border-radius:4px;
}
.search-item:hover { background:hsl(var(--muted)); }
.search-item.highlighted { background:hsl(var(--primary)/.15); }
.search-empty { padding:.5rem; text-align:center; font-size:var(--font-sm); color:hsl(var(--muted-foreground)); }
.controls .chip { padding:.35rem .65rem; font-size:var(--font-sm); }
.chip {
  padding:.25rem .6rem; border:1px solid hsl(var(--border)); border-radius:var(--radius);
  font-size:var(--font-xs); cursor:pointer; background:transparent; color:hsl(var(--muted-foreground));
}
.chip.primary { background:hsl(var(--primary)); color:#fff; border-color:hsl(var(--primary)); font-weight:600; }
.chip.primary:disabled { opacity:.5; cursor:not-allowed; }

.summary { display:flex; gap:.75rem; margin-bottom:1rem; flex-wrap:wrap; }
.card {
  flex:1; min-width:140px; background:hsl(var(--card)); border:1px solid hsl(var(--border));
  border-radius:var(--radius); padding:.65rem .85rem; display:flex; flex-direction:column; gap:2px;
}
.card-label { font-size:var(--font-xs); color:hsl(var(--muted-foreground)); }
.card-value { font-size:var(--font-lg); font-weight:700; color:hsl(var(--foreground)); }
.card-value.primary { color:hsl(var(--primary)); }
.card-value.green { color:#16A34A; }
.card-value.yellow { color:#EAB308; }
.card-sub { font-size:var(--font-xs); color:hsl(var(--muted-foreground)); }

.table-wrap { overflow-x:auto; border:1px solid hsl(var(--border)); border-radius:var(--radius); }
table { width:100%; border-collapse:collapse; font-size:var(--font-sm); }
th, td { padding:.3rem .5rem; white-space:nowrap; border-bottom:1px solid hsl(var(--border)/.5); }
th { font-weight:600; font-size:var(--font-xs); color:hsl(var(--muted-foreground)); }
.rt { text-align:right; }
.lt { text-align:left; }
.section-hdr th { padding:.4rem .5rem; font-size:var(--font-sm); font-weight:700; }
.calls-hdr { text-align:right; color:#16A34A; }
.puts-hdr { text-align:left; color:hsl(0 84% 60%); }
.strike-hdr { text-align:center; }
.col-hdr th { border-bottom:2px solid hsl(var(--border)); }
.strike-col { text-align:center; }
.strike-cell { text-align:center; font-weight:700; background:hsl(var(--muted)/.3); padding:.35rem .5rem; }
.strike-cell.atm { background:hsl(var(--primary)/.12); }
.strike-inner { display:flex; flex-direction:column; align-items:center; gap:1px; }
.strike-label { font-size:.625rem; font-weight:600; color:hsl(var(--muted-foreground)); }
.strike-val { font-size:var(--font-sm); color:hsl(var(--foreground)); }
.ltp-cell { font-weight:600; }
.ltp-cell.up { color:#16A34A; }
.ltp-cell.down { color:hsl(0 84% 60%); }
.oi-cell { color:hsl(var(--muted-foreground)); font-size:var(--font-xs); }
.chain-row:hover td { background:hsl(var(--muted)/.4); }
.chain-row:last-child td { border-bottom:none; }
.error { color:hsl(var(--destructive)); padding:1rem; text-align:center; }
.loading { text-align:center; padding:2rem; color:hsl(var(--muted-foreground)); }
</style>
