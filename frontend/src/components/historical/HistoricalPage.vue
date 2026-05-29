<script setup>
import { ref, computed, watch, onMounted } from 'vue'
import { api } from '../../utils/api.js'

// ─── Downloader State ──────────────────────────────
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

const symbolSearch = ref('')
const symbolOpen = ref(false)
const symbolInput = ref(null)
const highlightIdx = ref(0)

let searchTimer
watch(symbolSearch, (val) => {
  clearTimeout(searchTimer)
  if (!val) { loadUnderlyings(); return }
  searchTimer = setTimeout(() => loadUnderlyings(val), 300)
})

const filteredSymbols = computed(() => underlyings.value)

function toggleSymbol() {
  symbolOpen.value = !symbolOpen.value
  if (symbolOpen.value) { symbolSearch.value = ''; highlightIdx.value = 0; setTimeout(() => symbolInput.value?.focus(), 100) }
}
function selectSymbol(s) { selectedSymbol.value = s; symbolSearch.value = ''; symbolOpen.value = false }
function onSymbolKeydown(e) {
  if (e.key === 'ArrowDown') { e.preventDefault(); if (highlightIdx.value < filteredSymbols.value.length - 1) highlightIdx.value++ }
  else if (e.key === 'ArrowUp') { e.preventDefault(); if (highlightIdx.value > 0) highlightIdx.value-- }
  else if (e.key === 'Enter' || e.key === 'Tab') { e.preventDefault(); if (filteredSymbols.value.length) selectSymbol(filteredSymbols.value[highlightIdx.value]) }
  else if (e.key === 'Escape') { symbolOpen.value = false }
}

function updateDateRange() {
  const now = new Date()
  toDate.value = now.toISOString().split('T')[0]
  const maxDays = { '1m': 30, '3m': 60, '5m': 100, '10m': 100, '15m': 200, '30m': 200, '1h': 400, '1d': 2000 }
  const days = maxDays[selectedInterval.value] || 7
  fromDate.value = new Date(now.getTime() - days * 24 * 60 * 60 * 1000).toISOString().split('T')[0]
}
watch(selectedInterval, updateDateRange)

async function loadExchanges() {
  try {
    exchanges.value = await api('/api/historical/exchanges')
    if (exchanges.value.length > 0 && !selectedExchange.value) selectedExchange.value = exchanges.value[0]
  } catch {}
}
async function loadUnderlyings(q) {
  const ex = typeof selectedExchange.value === 'object' ? selectedExchange.value.code : selectedExchange.value
  if (!ex) return
  try {
    const url = q ? `/api/historical/underlyings?exchange=${ex}&q=${encodeURIComponent(q)}` : `/api/historical/underlyings?exchange=${ex}`
    underlyings.value = await api(url)
  } catch { underlyings.value = [] }
}
async function download() {
  if (!selectedSymbol.value || !selectedExchange.value || !selectedInterval.value || !fromDate.value || !toDate.value) return
  loading.value = true; error.value = ''; candles.value = []
  try {
    const ex = typeof selectedExchange.value === 'object' ? selectedExchange.value.code : selectedExchange.value
    const res = await api('/api/historical/download', {
      method: 'POST',
      body: JSON.stringify({ symbol: selectedSymbol.value, exchange: ex, interval: selectedInterval.value, from: fromDate.value + ' 09:15', to: toDate.value + ' 15:30' }),
    })
    candles.value = res.candles || []
    stats.value = { count: res.count || 0, price_change: res.price_change || 0 }
  } catch (e) { error.value = e.message }
  finally { loading.value = false }
}
watch(selectedExchange, () => { loadUnderlyings() })
onMounted(() => { loadExchanges(); updateDateRange() })

function exchDisplay(ex) { return typeof ex === 'object' ? ex.code : ex }
function exchLabel(ex) { return typeof ex === 'object' ? ex.name : ex }
function formatPrice(v) { return Number(v).toFixed(2) }
function formatVol(v) { return v >= 100000 ? (v/100000).toFixed(1)+'L' : v >= 1000 ? (v/1000).toFixed(1)+'K' : String(v) }
function changeClass(v) { return v > 0 ? 'up' : v < 0 ? 'down' : '' }

// ─── Scheduler / Groups State ──────────────────────
const groups = ref([])
const activeGroup = ref(null)
const groupItems = ref([])
const logs = ref([])
const showCreateGroup = ref(false)
const showDeleteGroup = ref(null)
const showSettings = ref(false)
const showAddItem = ref(false)
const groupName = ref('')
const settingsForm = ref({ cron: '0 15 * * 1-5', broker_priority: '', is_active: 1 })
const itemForm = ref({ symbol: '', exchange: 'NSE', interval: '1d' })
const aiQuery = ref('')
const aiLoading = ref(false)
const aiSuggestions = ref([])
const aiError = ref('')
const groupSearchQ = ref('')
const groupSearchResults = ref([])

async function loadGroups() { groups.value = await api('/api/scheduler/groups') }
async function loadGroupItems(gid) { groupItems.value = await api(`/api/scheduler/groups/${gid}/items`) }
async function loadLogs(gid) { logs.value = await api(`/api/scheduler/groups/${gid}/logs`) }
async function loadSettings(gid) {
  try {
    const s = await api(`/api/scheduler/groups/${gid}/settings`)
    settingsForm.value = { cron: s.cron, broker_priority: s.broker_priority || '', is_active: s.is_active }
  } catch {}
}

async function selectGroup(g) {
  activeGroup.value = g
  await Promise.all([loadGroupItems(g.id), loadLogs(g.id)])
  await loadSettings(g.id)
}

async function submitCreateGroup() {
  if (!groupName.value) return
  await api('/api/scheduler/groups', { method: 'POST', body: JSON.stringify({ name: groupName.value }) })
  showCreateGroup.value = false; groupName.value = ''; await loadGroups()
}
async function confirmDelete() {
  if (!showDeleteGroup.value) return
  const g = showDeleteGroup.value
  await api(`/api/scheduler/groups/${g.id}`, { method: 'DELETE' })
  if (activeGroup.value?.id === g.id) { activeGroup.value = null; groupItems.value = []; logs.value = [] }
  showDeleteGroup.value = false; await loadGroups()
}
async function saveSettings() {
  await api(`/api/scheduler/groups/${activeGroup.value.id}/settings`, {
    method: 'PUT', body: JSON.stringify(settingsForm.value),
  })
  showSettings.value = false; await loadGroups()
}
async function addGroupItem() {
  if (!itemForm.value.symbol) return
  await api(`/api/scheduler/groups/${activeGroup.value.id}/items`, {
    method: 'POST', body: JSON.stringify(itemForm.value),
  })
  itemForm.value = { symbol: '', exchange: 'NSE', interval: '1d' }
  showAddItem.value = false; groupSearchQ.value = ''; groupSearchResults.value = []
  await loadGroupItems(activeGroup.value.id); await loadGroups()
}
async function askAI() {
  if (!aiQuery.value || aiQuery.value.length < 3) return
  aiLoading.value = true; aiError.value = ''; aiSuggestions.value = []
  try {
    const ex = typeof selectedExchange.value === 'object' ? selectedExchange.value.code : selectedExchange.value
    const res = await api('/api/historical/ai-suggest', {
      method: 'POST',
      body: JSON.stringify({ query: aiQuery.value, exchange: ex || 'NSE' }),
    })
    aiSuggestions.value = res.symbols || []
    if (!aiSuggestions.value.length) aiError.value = 'No symbols found. Try a different query.'
  } catch (e) { aiError.value = e.message }
  finally { aiLoading.value = false }
}

async function acceptAll() {
  const existing = await api('/api/scheduler/groups/' + activeGroup.value.id + '/items')
  const seen = new Set(existing.map(i => i.symbol + '|' + i.exchange))
  const adds = aiSuggestions.value.filter(s => !seen.has(s.symbol + '|' + (s.exchange || 'NSE')))
  await Promise.all(adds.map(s =>
    api('/api/scheduler/groups/' + activeGroup.value.id + '/items', {
      method: 'POST', body: JSON.stringify({ symbol: s.symbol, exchange: s.exchange, interval: '1d' }),
    }).catch(() => {})
  ))
  aiSuggestions.value = []
  aiQuery.value = ''
  await Promise.all([loadGroupItems(activeGroup.value.id), loadGroups()])
}

async function replaceAll() {
  const existing = await api('/api/scheduler/groups/' + activeGroup.value.id + '/items')
  await Promise.all(existing.map(it =>
    api('/api/scheduler/items/' + it.id, { method: 'DELETE' }).catch(() => {})
  ))
  await Promise.all(aiSuggestions.value.map(s =>
    api('/api/scheduler/groups/' + activeGroup.value.id + '/items', {
      method: 'POST', body: JSON.stringify({ symbol: s.symbol, exchange: s.exchange, interval: '1d' }),
    }).catch(() => {})
  ))
  aiSuggestions.value = []
  aiQuery.value = ''
  await Promise.all([loadGroupItems(activeGroup.value.id), loadGroups()])
}

async function deleteGroupItem(it) {
  await api(`/api/scheduler/items/${it.id}`, { method: 'DELETE' })
  await loadGroupItems(activeGroup.value.id); await loadGroups()
}
async function runNow() {
  await api(`/api/scheduler/groups/${activeGroup.value.id}/run`, { method: 'POST' })
  setTimeout(() => loadLogs(activeGroup.value.id), 2000)
}

let gsTimer
watch(groupSearchQ, (val) => {
  clearTimeout(gsTimer)
  if (!val || val.length < 2) { groupSearchResults.value = []; return }
  gsTimer = setTimeout(async () => {
    try { groupSearchResults.value = await api(`/api/search-contracts?q=${val}`) } catch { groupSearchResults.value = [] }
  }, 300)
})
function pickContract(c) { itemForm.value.symbol = c.symbol; itemForm.value.exchange = c.exchange; groupSearchQ.value = ''; groupSearchResults.value = [] }

onMounted(loadGroups)
</script>

<template>
  <div class="page">
    <header><h2>Historical Data Downloader</h2></header>

    <!-- Download Controls -->
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
      <select v-model="selectedInterval"><option v-for="i in intervals" :key="i" :value="i">{{ i }}</option></select>
      <label>From: <input v-model="fromDate" type="date" class="date-input" /></label>
      <label>To: <input v-model="toDate" type="date" class="date-input" /></label>
      <button class="chip primary" @click="download" :disabled="loading">{{ loading ? 'Downloading...' : 'Download' }}</button>
    </div>

    <div v-if="error" class="error">{{ error }}</div>

    <!-- Downloaded Data Display -->
    <div v-if="candles.length" class="stats">
      <div class="card"><span class="card-label">Records</span><span class="card-value">{{ stats.count }}</span></div>
      <div class="card"><span class="card-label">Price Change</span><span class="card-value" :class="changeClass(stats.price_change)">{{ stats.price_change > 0 ? '+' : '' }}{{ formatPrice(stats.price_change) }}</span></div>
      <div class="card"><span class="card-label">{{ selectedSymbol }}</span><span class="card-value">{{ selectedInterval }}</span><span class="card-sub">{{ selectedExchange }}</span></div>
    </div>
    <div v-if="candles.length" class="table-wrap">
      <table>
        <thead><tr><th>Timestamp</th><th class="rt">Open</th><th class="rt">High</th><th class="rt">Low</th><th class="rt">Close</th><th class="rt">Volume</th></tr></thead>
        <tbody><tr v-for="c in candles" :key="c.timestamp">
          <td class="ts">{{ c.timestamp }}</td><td class="rt">{{ formatPrice(c.open) }}</td><td class="rt">{{ formatPrice(c.high) }}</td>
          <td class="rt">{{ formatPrice(c.low) }}</td><td class="rt" :class="changeClass(c.close - c.open)">{{ formatPrice(c.close) }}</td><td class="rt vol">{{ formatVol(c.volume) }}</td>
        </tr></tbody>
      </table>
    </div>

    <!-- ─── Scheduler / Groups Section ──────────────── -->
    <hr class="section-divider" />

    <div class="sched-section">
      <div class="sched-header">
        <h3>Groups & Schedule</h3>
        <button class="chip primary" @click="showCreateGroup = true" style="font-size:10px;padding:2px 10px">+ New Group</button>
      </div>

      <div class="sched-layout">
        <!-- Groups List -->
        <aside class="groups-panel">
          <div v-for="g in groups" :key="g.id" class="group-item" :class="{ active: activeGroup?.id === g.id }" @click="selectGroup(g)">
            <div class="group-info">
              <strong>{{ g.name }}</strong>
              <span class="group-meta">{{ g.item_count }} items · {{ g.last_run || 'never' }}</span>
            </div>
            <div class="group-actions">
              <button class="chip sm" @click.stop="selectGroup(g); showSettings = true" title="Settings">&#9881;</button>
              <button class="chip sm" @click.stop="activeGroup = g; runNow()" title="Run Now">&#9654;</button>
              <button class="chip sm danger" @click.stop="showDeleteGroup = g" title="Delete">&#10005;</button>
            </div>
          </div>
          <div v-if="!groups.length" class="empty-msg">No groups yet.</div>
        </aside>

        <!-- Detail -->
        <main v-if="activeGroup" class="sched-detail">
          <div class="detail-header">
            <strong>{{ activeGroup.name }}</strong>
            <div class="hdr-actions">
              <button class="chip" @click="showAddItem = true">+ Item</button>
              <button class="chip" @click="showSettings = true">Settings</button>
              <button class="chip primary" @click="runNow()">Run Now</button>
            </div>
          </div>

          <!-- AI Assistant -->
          <div class="ai-section">
            <div class="ai-header"><span class="ai-icon" style="font-size:1.1rem">&#129302;</span> AI Symbol Finder</div>
            <div class="ai-input-row">
              <input v-model="aiQuery" placeholder="e.g. Nifty 50 stocks, companies >500Cr turnover..." class="inp" @keyup.enter="askAI" :disabled="aiLoading" />
              <button class="chip primary" @click="askAI" :disabled="aiLoading || aiQuery.length < 3" style="flex-shrink:0">{{ aiLoading ? '...' : 'Ask AI' }}</button>
            </div>
            <div v-if="aiError" class="ai-error">{{ aiError }}</div>
            <div v-if="aiSuggestions.length" class="ai-suggestions">
              <div class="ai-sugg-header">
                <span>{{ aiSuggestions.length }} symbols found · {{ groupItems.length }} in group</span>
                <div class="ai-sugg-actions">
                  <button class="chip primary sm" @click="acceptAll">+ Add New</button>
                  <button class="chip sm" @click="replaceAll">Replace All</button>
                </div>
              </div>
              <div v-for="s in aiSuggestions" :key="s.symbol" class="ai-sugg-item">
                <strong>{{ s.symbol }}</strong>
                <span class="muted">{{ s.exchange }}</span>
                <span class="muted" style="flex:1;font-size:.65rem">{{ s.name }}</span>
                <span class="hint">{{ s.reason }}</span>
              </div>
            </div>
          </div>

          <!-- Add Item -->
          <div v-if="showAddItem" class="add-item-panel">
            <div class="search-wrap">
              <input v-model="groupSearchQ" placeholder="Search symbol..." class="inp" />
              <div v-if="groupSearchResults.length" class="search-dropdown">
                <div v-for="c in groupSearchResults" :key="c.id" class="search-item" @mousedown="pickContract(c)">{{ c.symbol }} <span class="ex-badge">{{ c.exchange }}</span></div>
              </div>
            </div>
            <div class="af-row"><label>Symbol</label><input v-model="itemForm.symbol" class="inp" readonly /></div>
            <div class="af-row"><label>Exchange</label><select v-model="itemForm.exchange" class="inp"><option value="NSE">NSE</option><option value="NFO">NFO</option></select></div>
            <div class="af-row"><label>Interval</label><select v-model="itemForm.interval" class="inp"><option v-for="i in intervals" :key="i" :value="i">{{ i }}</option></select></div>
            <div class="modal-actions"><button class="chip primary" @click="addGroupItem">Add</button><button class="chip" @click="showAddItem = false; groupSearchQ = ''; groupSearchResults = []">Cancel</button></div>
          </div>

          <!-- Items -->
          <div class="section">
            <div class="section-title">Items ({{ groupItems.length }})</div>
            <table v-if="groupItems.length" class="data-table">
              <thead><tr><th>Symbol</th><th>Ex</th><th>Interval</th><th>Token</th><th></th></tr></thead>
              <tbody><tr v-for="it in groupItems" :key="it.id">
                <td><strong>{{ it.symbol }}</strong></td><td>{{ it.exchange }}</td><td>{{ it.interval }}</td><td class="muted">{{ it.token || '-' }}</td>
                <td><button class="chip sm danger" @click="deleteGroupItem(it)">&#10005;</button></td>
              </tr></tbody>
            </table>
            <div v-else class="empty-msg">No items.</div>
          </div>

          <!-- Logs -->
          <div class="section">
            <div class="section-title">Recent Runs</div>
            <table v-if="logs.length" class="data-table">
              <thead><tr><th>Time</th><th>Status</th><th>Items</th><th>Message</th></tr></thead>
              <tbody><tr v-for="l in logs" :key="l.id">
                <td class="muted">{{ l.time }}</td>
                <td><span class="status-badge" :class="l.status">{{ l.status }}</span></td>
                <td>{{ l.success }}/{{ l.total }}</td>
                <td class="muted">{{ l.message }}</td>
              </tr></tbody>
            </table>
            <div v-else class="empty-msg">No runs yet.</div>
          </div>
        </main>
        <main v-else class="sched-detail"><div class="empty-state">Select a group or create one.</div></main>
      </div>
    </div>

    <!-- Modals -->
    <div v-if="showCreateGroup" class="modal-overlay" @click.self="showCreateGroup = false">
      <div class="modal-box sm"><h4>Create Group</h4>
        <div class="form-row"><label>Group Name</label><input v-model="groupName" placeholder="Enter group name..." class="inp" @keyup.enter="submitCreateGroup" autofocus /></div>
        <div class="modal-actions"><button class="chip primary" @click="submitCreateGroup">Create</button><button class="chip" @click="showCreateGroup = false">Cancel</button></div>
      </div>
    </div>
    <div v-if="showDeleteGroup" class="modal-overlay" @click.self="showDeleteGroup = false">
      <div class="modal-box sm"><h4>Delete Group</h4>
        <p style="font-size:var(--font-sm);color:hsl(var(--muted-foreground));margin:.5rem 0">Delete "<strong>{{ showDeleteGroup?.name }}</strong>"?</p>
        <div class="modal-actions"><button class="chip danger" @click="confirmDelete">Delete</button><button class="chip" @click="showDeleteGroup = false">Cancel</button></div>
      </div>
    </div>
    <div v-if="showSettings" class="modal-overlay" @click.self="showSettings = false">
      <div class="modal-box sm"><h4>Settings: {{ activeGroup?.name }}</h4>
        <div class="form-row"><label>Cron Expression</label><input v-model="settingsForm.cron" placeholder="0 15 * * 1-5" class="inp" /><span class="hint">min hour day month weekday</span></div>
        <div class="form-row"><label>Broker Priority</label><select v-model="settingsForm.broker_priority" class="inp"><option value="">Any</option><option value="angel">Angel</option><option value="aliceblue">Alice Blue</option></select></div>
        <div class="form-row"><label>Active</label><label class="switch-label"><span class="switch"><input v-model="settingsForm.is_active" type="checkbox" true-value="1" false-value="0" /><span class="slider"></span></span>{{ settingsForm.is_active == 1 ? 'Yes' : 'No' }}</label></div>
        <div class="modal-actions"><button class="chip primary" @click="saveSettings">Save</button><button class="chip" @click="showSettings = false">Cancel</button></div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.page { }
header h2 { margin:0 0 1rem; }
.controls { display:flex; flex-wrap:wrap; gap:.5rem; margin-bottom:1rem; align-items:center; }
.controls select, .date-input { padding:.35rem .55rem; border:1px solid hsl(var(--input)); border-radius:var(--radius); font-size:var(--font-sm); background:hsl(var(--card)); outline:none; }
.controls label { display:flex; align-items:center; gap:4px; font-size:var(--font-sm); color:hsl(var(--muted-foreground)); }
.date-input { width:140px; }
.chip { padding:.35rem .65rem; border:1px solid hsl(var(--border)); border-radius:var(--radius); font-size:var(--font-xs); cursor:pointer; background:transparent; color:hsl(var(--muted-foreground)); display:inline-flex; align-items:center; gap:4px; }
.chip:hover { border-color:hsl(var(--primary)); color:hsl(var(--primary)); }
.chip.primary { background:hsl(var(--primary)); color:#fff; border-color:hsl(var(--primary)); font-weight:600; }
.chip.primary:hover { opacity:.9; }
.chip.primary:disabled { opacity:.5; cursor:not-allowed; }
.chip.danger:hover { border-color:hsl(var(--destructive)); color:hsl(var(--destructive)); }
.chip.sm { font-size:.625rem; padding:2px 6px; }
.search-wrap { position:relative; min-width:160px; }
.search-btn { display:flex; align-items:center; gap:4px; cursor:pointer; color:hsl(var(--foreground)); padding:.35rem .55rem; border:1px solid hsl(var(--input)); border-radius:var(--radius); font-size:var(--font-sm); background:hsl(var(--card)); outline:none; width:100%; }
.search-btn .arrow { font-size:10px; color:hsl(var(--muted-foreground)); margin-left:auto; }
.search-input { padding:.35rem .55rem; border:1px solid hsl(var(--input)); border-radius:var(--radius); font-size:var(--font-sm); background:hsl(var(--card)); outline:none; width:100%; box-sizing:border-box; }
.search-input:focus { border-color:hsl(var(--ring)); }
.search-dropdown { position:absolute; top:100%; left:0; right:0; z-index:20; background:hsl(var(--card)); border:1px solid hsl(var(--border)); border-radius:var(--radius); box-shadow:0 4px 12px rgba(0,0,0,.1); margin-top:2px; max-height:180px; overflow-y:auto; }
.search-item { padding:.35rem .55rem; cursor:pointer; font-size:var(--font-sm); }
.search-item:hover { background:hsl(var(--muted)); }
.search-item.highlighted { background:hsl(var(--primary)/.15); }
.stats { display:flex; gap:.75rem; margin-bottom:1rem; flex-wrap:wrap; }
.card { flex:1; min-width:120px; background:hsl(var(--card)); border:1px solid hsl(var(--border)); border-radius:var(--radius); padding:.6rem .8rem; display:flex; flex-direction:column; gap:2px; }
.card-label { font-size:var(--font-xs); color:hsl(var(--muted-foreground)); }
.card-value { font-size:var(--font-lg); font-weight:700; }
.card-value.up { color:#16A34A; }
.card-value.down { color:hsl(0 84% 60%); }
.card-sub { font-size:var(--font-xs); color:hsl(var(--muted-foreground)); }
.table-wrap { overflow-x:auto; border:1px solid hsl(var(--border)); border-radius:var(--radius); max-height:50vh; overflow-y:auto; }
table { width:100%; border-collapse:collapse; font-size:var(--font-sm); }
th, td { padding:.3rem .45rem; white-space:nowrap; border-bottom:1px solid hsl(var(--border)/.5); }
th { font-weight:600; font-size:var(--font-xs); color:hsl(var(--foreground)); position:sticky; top:0; background:hsl(var(--card)); }
.rt { text-align:right; font-family:monospace; }
.ts { font-family:monospace; font-size:var(--font-xs); color:hsl(var(--muted-foreground)); }
.vol { color:hsl(var(--muted-foreground)); }
tr:hover td { background:hsl(var(--muted)/.3); }
.error { color:hsl(var(--destructive)); margin-bottom:.5rem; }

/* AI section */
.ai-section { background:hsl(var(--primary)/.04); border:1px solid hsl(var(--primary)/.2); border-radius:var(--radius); padding:.7rem; margin-bottom:.6rem; }
.ai-header { font-size:var(--font-sm); font-weight:600; margin-bottom:.5rem; display:flex; align-items:center; gap:6px; }
.ai-input-row { display:flex; gap:.4rem; }
.ai-input-row .inp { flex:1; }
.ai-error { font-size:var(--font-xs); color:hsl(var(--destructive)); margin-top:.3rem; }
.ai-suggestions { margin-top:.5rem; border:1px solid hsl(var(--border)); border-radius:var(--radius); overflow:hidden; }
.ai-sugg-header { display:flex; justify-content:space-between; align-items:center; padding:.3rem .5rem; background:hsl(var(--muted)/.3); font-size:var(--font-xs); font-weight:600; }
.ai-sugg-actions { display:flex; gap:4px; }
.ai-sugg-item { display:flex; align-items:center; gap:6px; padding:.25rem .5rem; border-bottom:1px solid hsl(var(--border)/.5); font-size:var(--font-xs); }
.ai-sugg-item:last-child { border-bottom:none; }
.ai-sugg-item strong { font-size:var(--font-sm); min-width:90px; }
.hint { font-size:.6rem; color:hsl(var(--muted-foreground)); white-space:nowrap; overflow:hidden; text-overflow:ellipsis; max-width:120px; }

/* Scheduler section */
.section-divider { margin:1.5rem 0; border:none; border-top:2px solid hsl(var(--border)); }
.sched-section { }
.sched-header { display:flex; justify-content:space-between; align-items:center; margin-bottom:.75rem; }
.sched-header h3 { margin:0; font-size:var(--font-base); }
.sched-layout { display:flex; gap:.75rem; }
.groups-panel { width:260px; flex-shrink:0; overflow-y:auto; max-height:400px; }
.group-item { display:flex; justify-content:space-between; align-items:center; padding:.4rem .5rem; border:1px solid hsl(var(--border)); border-radius:var(--radius); margin-bottom:.3rem; cursor:pointer; transition:border-color .12s; }
.group-item:hover, .group-item.active { border-color:hsl(var(--primary)); }
.group-info { display:flex; flex-direction:column; gap:1px; min-width:0; }
.group-info strong { font-size:var(--font-sm); color:hsl(var(--foreground)); }
.group-meta { font-size:.6rem; color:hsl(var(--muted-foreground)); }
.group-actions { display:flex; gap:2px; flex-shrink:0; }
.sched-detail { flex:1; overflow-y:auto; min-width:0; max-height:400px; }
.detail-header { display:flex; justify-content:space-between; align-items:center; margin-bottom:.5rem; flex-wrap:wrap; gap:.4rem; }
.detail-header strong { font-size:var(--font-sm); }
.hdr-actions { display:flex; gap:.3rem; }
.add-item-panel { background:hsl(var(--muted)/.2); border:1px solid hsl(var(--border)); border-radius:var(--radius); padding:.6rem; margin-bottom:.6rem; }
.af-row { margin-bottom:.4rem; }
.af-row label { display:block; font-size:var(--font-xs); font-weight:600; margin-bottom:2px; }
.inp { width:100%; padding:.35rem .5rem; border:1px solid hsl(var(--input)); border-radius:var(--radius); font-size:var(--font-sm); background:hsl(var(--card)); outline:none; box-sizing:border-box; }
.inp:focus { border-color:hsl(var(--ring)); }
.hint { font-size:.6rem; color:hsl(var(--muted-foreground)); }
.section { margin-bottom:.75rem; }
.section-title { font-size:var(--font-sm); font-weight:600; margin-bottom:.3rem; color:hsl(var(--foreground)); }
.data-table { width:100%; border-collapse:collapse; font-size:var(--font-xs); }
.data-table th, .data-table td { padding:.25rem .4rem; border-bottom:1px solid hsl(var(--border)); text-align:left; white-space:nowrap; }
.data-table th { font-weight:600; color:hsl(var(--foreground)); }
.muted { color:hsl(var(--muted-foreground)); }
.empty-msg { text-align:center; padding:1rem; font-size:var(--font-sm); color:hsl(var(--muted-foreground)); }
.empty-state { text-align:center; padding:2rem; color:hsl(var(--muted-foreground)); font-size:var(--font-sm); }
.modal-box.sm { max-width:380px; background:hsl(var(--card)); border:1px solid hsl(var(--border)); border-radius:var(--radius); padding:1rem; margin:3rem auto; }
.modal-overlay { position:fixed; inset:0; z-index:50; background:rgba(0,0,0,.3); display:flex; align-items:flex-start; justify-content:center; padding-top:3rem; }
.form-row { margin-bottom:.5rem; }
.form-row label { display:block; font-size:var(--font-xs); font-weight:600; color:hsl(var(--foreground)); margin-bottom:2px; }
.modal-actions { display:flex; gap:.5rem; margin-top:.6rem; }
.switch-label { display:flex; align-items:center; gap:6px; font-size:var(--font-sm); cursor:pointer; }
.switch { position:relative; display:inline-block; width:34px; height:20px; }
.switch input { opacity:0; width:0; height:0; }
.slider { position:absolute; inset:0; background:hsl(var(--muted-foreground)); border-radius:999px; transition:background .2s; cursor:pointer; }
.slider::before { content:''; position:absolute; left:3px; top:3px; width:14px; height:14px; background:#fff; border-radius:999px; transition:transform .2s; }
.switch input:checked + .slider { background:hsl(var(--primary)); }
.switch input:checked + .slider::before { transform:translateX(14px); }
.ex-badge { font-size:.6rem; color:hsl(var(--muted-foreground)); background:hsl(var(--muted)); padding:0 4px; border-radius:4px; }
.status-badge { font-size:.6rem; font-weight:600; padding:1px 5px; border-radius:999px; }
.status-badge.success, .status-badge.placed { background:hsl(142 70% 45%/.15); color:#16A34A; }
.status-badge.failed, .status-badge.error { background:hsl(0 84% 60%/.15); color:hsl(0 84% 60%); }
.status-badge.partial { background:hsl(38 92% 50%/.15); color:#D97706; }
.status-badge.running { background:hsl(210 100% 50%/.15); color:hsl(210 100% 50%); }
</style>
