<script setup>
import { ref, computed, watch, onMounted } from 'vue'
import { api } from '../../utils/api.js'

const groups = ref([])
const activeGroup = ref(null)
const items = ref([])
const logs = ref([])
const showSettings = ref(false)
const settingsForm = ref({ cron: '0 15 * * 1-5', broker_priority: '', is_active: 1 })
const showAddItem = ref(false)
const itemForm = ref({ symbol: '', exchange: 'NSE', interval: '1d' })
const loading = ref(false)
const searchQ = ref('')
const searchResults = ref([])

const intervals = ['1m','3m','5m','10m','15m','30m','1h','1d']

async function loadGroups() {
  groups.value = await api('/api/scheduler/groups')
}

async function createGroup() {
  const name = prompt('Group name:')
  if (!name) return
  await api('/api/scheduler/groups', { method: 'POST', body: JSON.stringify({ name }) })
  await loadGroups()
}

async function deleteGroup(g) {
  if (!confirm(`Delete "${g.name}"?`)) return
  await api(`/api/scheduler/groups/${g.id}`, { method: 'DELETE' })
  if (activeGroup.value?.id === g.id) { activeGroup.value = null; items.value = []; logs.value = [] }
  await loadGroups()
}

async function selectGroup(g) {
  activeGroup.value = g
  await Promise.all([loadItems(g.id), loadLogs(g.id)])
  await loadSettings(g.id)
}

async function loadItems(gid) {
  items.value = await api(`/api/scheduler/groups/${gid}/items`)
}

async function loadLogs(gid) {
  logs.value = await api(`/api/scheduler/groups/${gid}/logs`)
}

async function loadSettings(gid) {
  try {
    const s = await api(`/api/scheduler/groups/${gid}/settings`)
    settingsForm.value = { cron: s.cron, broker_priority: s.broker_priority || '', is_active: s.is_active }
  } catch {}
}

async function saveSettings() {
  await api(`/api/scheduler/groups/${activeGroup.value.id}/settings`, {
    method: 'PUT',
    body: JSON.stringify(settingsForm.value),
  })
  showSettings.value = false
  await loadGroups()
}

async function addItem() {
  if (!itemForm.value.symbol) return
  await api(`/api/scheduler/groups/${activeGroup.value.id}/items`, {
    method: 'POST', body: JSON.stringify(itemForm.value),
  })
  itemForm.value = { symbol: '', exchange: 'NSE', interval: '1d' }
  showAddItem.value = false
  searchQ.value = ''
  searchResults.value = []
  await loadItems(activeGroup.value.id)
  await loadGroups()
}

async function deleteItem(it) {
  await api(`/api/scheduler/items/${it.id}`, { method: 'DELETE' })
  await loadItems(activeGroup.value.id)
  await loadGroups()
}

async function runNow(gid) {
  await api(`/api/scheduler/groups/${gid}/run`, { method: 'POST' })
  setTimeout(() => loadLogs(gid), 2000)
}

let searchTimer
watch(searchQ, (val) => {
  clearTimeout(searchTimer)
  if (!val || val.length < 2) { searchResults.value = []; return }
  searchTimer = setTimeout(async () => {
    try { searchResults.value = await api(`/api/search-contracts?q=${val}`) } catch { searchResults.value = [] }
  }, 300)
})

function pickContract(c) {
  itemForm.value.symbol = c.symbol
  itemForm.value.exchange = c.exchange
  searchQ.value = ''
  searchResults.value = []
}

onMounted(loadGroups)
</script>

<template>
  <div class="page">
    <header><h2>Data Scheduler</h2></header>

    <div class="sched-layout">
      <!-- Groups Sidebar -->
      <aside class="groups-panel">
        <div class="groups-header"><span>Groups</span><button class="chip primary" @click="createGroup" style="font-size:10px;padding:2px 8px">+</button></div>
        <div v-for="g in groups" :key="g.id" class="group-item" :class="{ active: activeGroup?.id === g.id }" @click="selectGroup(g)">
          <div class="group-info">
            <strong>{{ g.name }}</strong>
            <span class="group-meta">{{ g.item_count }} items · {{ g.last_run || 'never' }}</span>
          </div>
          <div class="group-actions">
            <button class="chip sm" @click.stop="selectGroup(g); showSettings = true" title="Settings">&#9881;</button>
            <button class="chip sm" @click.stop="runNow(g.id)" title="Run Now">&#9654;</button>
            <button class="chip sm danger" @click.stop="deleteGroup(g)" title="Delete">&#10005;</button>
          </div>
        </div>
        <div v-if="!groups.length" class="empty-msg">No groups yet. Click + to create one.</div>
      </aside>

      <!-- Detail Panel -->
      <main v-if="activeGroup" class="detail-panel">
        <div class="detail-header">
          <h3>{{ activeGroup.name }}</h3>
          <div class="hdr-actions">
            <button class="chip" @click="showSettings = true">Settings</button>
            <button class="chip primary" @click="showAddItem = true">+ Add Item</button>
            <button class="chip" @click="runNow(activeGroup.id)">Run Now</button>
          </div>
        </div>

        <!-- Settings Modal -->
        <div v-if="showSettings" class="modal-overlay" @click.self="showSettings = false">
          <div class="modal-box sm">
            <h4>Settings: {{ activeGroup.name }}</h4>
            <div class="form-row">
              <label>Cron Expression</label>
              <input v-model="settingsForm.cron" placeholder="0 15 * * 1-5" class="inp" />
              <span class="hint">min hour day month weekday</span>
            </div>
            <div class="form-row">
              <label>Broker Priority</label>
              <select v-model="settingsForm.broker_priority" class="inp">
                <option value="">Any connected broker</option>
                <option value="angel">Angel Only</option>
                <option value="aliceblue">Alice Blue Only</option>
              </select>
            </div>
            <div class="form-row">
              <label>Active</label>
              <label class="switch-label">
                <span class="switch"><input v-model="settingsForm.is_active" type="checkbox" true-value="1" false-value="0" /><span class="slider"></span></span>
                {{ settingsForm.is_active == 1 ? 'Yes' : 'No' }}
              </label>
            </div>
            <div class="modal-actions">
              <button class="chip primary" @click="saveSettings">Save</button>
              <button class="chip" @click="showSettings = false">Cancel</button>
            </div>
          </div>
        </div>

        <!-- Add Item -->
        <div v-if="showAddItem" class="add-item-panel">
          <div class="search-wrap">
            <input v-model="searchQ" placeholder="Search symbol..." class="inp" />
            <div v-if="searchResults.length" class="search-dropdown">
              <div v-for="c in searchResults" :key="c.id" class="search-item" @mousedown="pickContract(c)">
                {{ c.symbol }} <span class="ex-badge">{{ c.exchange }}</span>
              </div>
            </div>
          </div>
          <div class="form-row">
            <label>Symbol</label>
            <input v-model="itemForm.symbol" class="inp" readonly />
          </div>
          <div class="form-row">
            <label>Exchange</label>
            <select v-model="itemForm.exchange" class="inp"><option value="NSE">NSE</option><option value="NFO">NFO</option><option value="BSE">BSE</option></select>
          </div>
          <div class="form-row">
            <label>Interval</label>
            <select v-model="itemForm.interval" class="inp"><option v-for="i in intervals" :key="i" :value="i">{{ i }}</option></select>
          </div>
          <div class="modal-actions">
            <button class="chip primary" @click="addItem">Add</button>
            <button class="chip" @click="showAddItem = false; searchQ = ''; searchResults = []">Cancel</button>
          </div>
        </div>

        <!-- Items Table -->
        <div class="section">
          <div class="section-title">Items ({{ items.length }})</div>
          <table v-if="items.length" class="data-table">
            <thead><tr><th>Symbol</th><th>Exchange</th><th>Interval</th><th>Token</th><th></th></tr></thead>
            <tbody>
              <tr v-for="it in items" :key="it.id">
                <td><strong>{{ it.symbol }}</strong></td><td>{{ it.exchange }}</td><td>{{ it.interval }}</td><td class="muted">{{ it.token || '-' }}</td>
                <td><button class="chip sm danger" @click="deleteItem(it)">&#10005;</button></td>
              </tr>
            </tbody>
          </table>
          <div v-else class="empty-msg">No items. Click + Add Item to add symbols.</div>
        </div>

        <!-- Logs -->
        <div class="section">
          <div class="section-title">Recent Runs</div>
          <table v-if="logs.length" class="data-table">
            <thead><tr><th>Time</th><th>Status</th><th>Items</th><th>Message</th></tr></thead>
            <tbody>
              <tr v-for="l in logs" :key="l.id">
                <td class="muted">{{ l.time }}</td>
                <td><span class="status-badge" :class="l.status">{{ l.status }}</span></td>
                <td>{{ l.success }}/{{ l.total }}</td>
                <td class="muted">{{ l.message }}</td>
              </tr>
            </tbody>
          </table>
          <div v-else class="empty-msg">No runs yet.</div>
        </div>
      </main>

      <!-- Empty state -->
      <main v-else class="detail-panel">
        <div class="empty-state">Select a group or create one to start scheduling data downloads.</div>
      </main>
    </div>
  </div>
</template>

<style scoped>
.page { height:100%; display:flex; flex-direction:column; }
header h2 { margin:0 0 .75rem; }
.sched-layout { display:flex; gap:1rem; flex:1; min-height:0; }
.groups-panel { width:280px; flex-shrink:0; overflow-y:auto; }
.groups-header { display:flex; justify-content:space-between; align-items:center; margin-bottom:.5rem; font-size:var(--font-sm); font-weight:600; color:hsl(var(--foreground)); }
.group-item { display:flex; justify-content:space-between; align-items:center; padding:.5rem .6rem; border:1px solid hsl(var(--border)); border-radius:var(--radius); margin-bottom:.35rem; cursor:pointer; transition:border-color .12s; }
.group-item:hover, .group-item.active { border-color:hsl(var(--primary)); }
.group-info { display:flex; flex-direction:column; gap:2px; min-width:0; }
.group-info strong { font-size:var(--font-sm); color:hsl(var(--foreground)); }
.group-meta { font-size:.625rem; color:hsl(var(--muted-foreground)); }
.group-actions { display:flex; gap:3px; flex-shrink:0; }
.chip.sm { font-size:.625rem; padding:2px 6px; border:1px solid hsl(var(--border)); border-radius:var(--radius); cursor:pointer; background:transparent; color:hsl(var(--muted-foreground)); }
.chip.sm:hover { border-color:hsl(var(--primary)); color:hsl(var(--primary)); }
.chip.primary { background:hsl(var(--primary)); color:#fff; border-color:hsl(var(--primary)); }
.chip.primary:hover { opacity:.9; }
.chip.danger:hover { border-color:hsl(var(--destructive)); color:hsl(var(--destructive)); }
.detail-panel { flex:1; overflow-y:auto; min-width:0; }
.detail-header { display:flex; justify-content:space-between; align-items:center; margin-bottom:.75rem; flex-wrap:wrap; gap:.5rem; }
.detail-header h3 { margin:0; }
.hdr-actions { display:flex; gap:.35rem; }
.modal-box.sm { max-width:400px; background:hsl(var(--card)); border:1px solid hsl(var(--border)); border-radius:var(--radius); padding:1rem; margin:3rem auto; }
.modal-overlay { position:fixed; inset:0; z-index:50; background:rgba(0,0,0,.3); display:flex; align-items:flex-start; justify-content:center; padding-top:3rem; }
.form-row { margin-bottom:.6rem; }
.form-row label { display:block; font-size:var(--font-xs); font-weight:600; color:hsl(var(--foreground)); margin-bottom:3px; }
.inp { width:100%; padding:.4rem .55rem; border:1px solid hsl(var(--input)); border-radius:var(--radius); font-size:var(--font-sm); background:hsl(var(--card)); outline:none; box-sizing:border-box; }
.inp:focus { border-color:hsl(var(--ring)); }
.hint { font-size:.625rem; color:hsl(var(--muted-foreground)); }
.modal-actions { display:flex; gap:.5rem; margin-top:.75rem; }
.switch-label { display:flex; align-items:center; gap:6px; font-size:var(--font-sm); cursor:pointer; }
.switch { position:relative; display:inline-block; width:34px; height:20px; }
.switch input { opacity:0; width:0; height:0; }
.slider { position:absolute; inset:0; background:hsl(var(--muted-foreground)); border-radius:999px; transition:background .2s; cursor:pointer; }
.slider::before { content:''; position:absolute; left:3px; top:3px; width:14px; height:14px; background:#fff; border-radius:999px; transition:transform .2s; }
.switch input:checked + .slider { background:hsl(var(--primary)); }
.switch input:checked + .slider::before { transform:translateX(14px); }
.search-wrap { position:relative; margin-bottom:.5rem; }
.search-dropdown { position:absolute; top:100%; left:0; right:0; z-index:20; max-height:180px; overflow-y:auto; background:hsl(var(--card)); border:1px solid hsl(var(--border)); border-radius:var(--radius); box-shadow:0 4px 12px rgba(0,0,0,.1); }
.search-item { padding:.35rem .55rem; cursor:pointer; font-size:var(--font-sm); display:flex; align-items:center; gap:6px; }
.search-item:hover { background:hsl(var(--muted)); }
.ex-badge { font-size:.625rem; color:hsl(var(--muted-foreground)); background:hsl(var(--muted)); padding:0 4px; border-radius:4px; }
.add-item-panel { background:hsl(var(--muted)/.2); border:1px solid hsl(var(--border)); border-radius:var(--radius); padding:.75rem; margin-bottom:.75rem; }
.section { margin-bottom:1rem; }
.section-title { font-size:var(--font-sm); font-weight:600; margin-bottom:.4rem; color:hsl(var(--foreground)); }
.data-table { width:100%; border-collapse:collapse; font-size:var(--font-sm); }
.data-table th, .data-table td { padding:.3rem .45rem; border-bottom:1px solid hsl(var(--border)); text-align:left; white-space:nowrap; }
.data-table th { font-weight:600; color:hsl(var(--foreground)); }
.muted { color:hsl(var(--muted-foreground)); }
.empty-msg { text-align:center; padding:1.5rem; font-size:var(--font-sm); color:hsl(var(--muted-foreground)); }
.empty-state { text-align:center; padding:3rem; color:hsl(var(--muted-foreground)); font-size:var(--font-sm); }
</style>
