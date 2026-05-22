<script setup>
import { ref, onMounted, onUnmounted, watch } from 'vue'
import { api } from '../utils/api.js'
import { useWebSocket } from '../composables/useWebSocket.js'

defineProps({ show: Boolean })
const emit = defineEmits(['close', 'place-order'])

const activeWL = ref(null)
const watchlists = ref([])
const items = ref([])
const searchQ = ref('')
const searchResults = ref([])
const searching = ref(false)
const dragItem = ref(null)
const dragOver = ref(null)
const showPrompt = ref(false)
const promptName = ref('')
const ltpMap = ref({})
const ws = useWebSocket()

// WebSocket: subscribe to current watchlist items, update LTP
let subscribedSymbols = []

function subscribeItems() {
  const syms = items.value.map(i => (i.exchange || '') + '|' + (i.token || '')).filter(Boolean)
  if (syms.length === 0) return
  // Unsubscribe old, subscribe new
  if (subscribedSymbols.length) ws.unsubscribe(subscribedSymbols)
  ws.subscribe(syms)
  subscribedSymbols = syms
}

ws.onTick((tick) => {
  if (tick.token && tick.ltp != null) {
    ltpMap.value = { ...ltpMap.value, [tick.token]: tick }
  }
})

// Subscribe when items change
watch(items, () => subscribeItems(), { deep: true })

onUnmounted(() => {
  if (subscribedSymbols.length) ws.unsubscribe(subscribedSymbols)
})

// Fetch watchlists
async function loadWatchlists() {
  try {
    watchlists.value = await api('/api/watchlists')
  } catch {
    watchlists.value = []
  }
  if (watchlists.value.length && !activeWL.value) {
    activeWL.value = watchlists.value[0]
  }
}

// Fetch items for active watchlist
async function loadItems() {
  if (!activeWL.value) { items.value = []; return }
  try {
    items.value = await api(`/api/watchlists/${activeWL.value.id}/items`)
  } catch {
    items.value = []
  }
}

watch(activeWL, () => loadItems())
onMounted(loadWatchlists)

// Search contracts
watch(searchQ, async (q) => {
  if (!q || q.length < 2) { searchResults.value = []; return }
  searching.value = true
  try {
    const results = await api(`/api/search-contracts?q=${encodeURIComponent(q)}`)
    const existing = new Set(items.value.map(i => i.token + '|' + i.exchange))
    searchResults.value = results.filter(c => !existing.has(c.token + '|' + c.exchange))
  } catch { searchResults.value = [] }
  finally { searching.value = false }
})

async function addToWatchlist(contract) {
  if (!activeWL.value) { return }
  try {
    await api(`/api/watchlists/${activeWL.value.id}/items`, {
      method: 'POST',
      body: JSON.stringify({
      symbol: contract.symbol,
      brsymbol: contract.brsymbol,
      name: contract.name,
      exchange: contract.exchange,
      token: contract.token,
      expiry: contract.expiry,
      strike: contract.strike,
      lotsize: contract.lotsize,
      instrumenttype: contract.instrumenttype,
      tick_size: contract.tick_size,
    }),
    })
    searchQ.value = ''
    searchResults.value = []
    await loadItems()
  } catch (e) {
    if (e.status === 409) {
      alert('Symbol already exists in this watchlist')
    } else {
      console.error('add failed:', e)
    }
  }
}

async function removeItem(item) {
  await api(`/api/watchlist-items/${item.id}`, { method: 'DELETE' })
  await loadItems()
}

async function onDrop(dropIdx) {
  const src = dragItem.value
  if (src === null || src === dropIdx) { dragItem.value = null; return }
  const list = [...items.value]
  if (src < 0 || src >= list.length || dropIdx < 0 || dropIdx >= list.length) { dragItem.value = null; return }
  // Move item from src to dropIdx in the array
  const [moved] = list.splice(src, 1)
  list.splice(dropIdx, 0, moved)
  // Assign sequential sort_order based on new position
  for (let i = 0; i < list.length; i++) {
    if (list[i].sort_order !== i) {
      await api(`/api/watchlist-items/${list[i].id}/reorder`, { method: 'PUT', body: JSON.stringify({ sort_order: i }) })
    }
  }
  dragItem.value = null
  await loadItems()
}

function newWatchlist() {
  promptName.value = ''
  showPrompt.value = true
}

async function confirmNew() {
  const name = promptName.value.trim()
  if (!name) return
  showPrompt.value = false
  await api('/api/watchlists', { method: 'POST', body: JSON.stringify({ name, sort_order: watchlists.value.length }) })
  await loadWatchlists()
}

function tickData(item) {
  if (!item?.token) return null
  return ltpMap.value[item.token] || ltpMap.value['999' + item.token] || null
}

function ltpDisplay(item) {
  const t = tickData(item)
  if (!t) return '-'
  const value = t.ltp != null ? t.ltp : (t.close != null ? t.close : 0)
  return Number(value).toFixed(2)
}

function changeDisplay(item) {
  const t = tickData(item)
  if (!t || t.change == null) return null
  const ch = Number(t.change)
  const sign = ch >= 0 ? '+' : ''
  return `${sign}${ch.toFixed(2)}`
}

function pctDisplay(item) {
  const t = tickData(item)
  if (!t || t.change == null || !t.close) return null
  const pct = (Number(t.change) / Number(t.close)) * 100
  const sign = pct >= 0 ? '+' : ''
  return `${sign}${pct.toFixed(2)}%`
}

function priceClass(item) {
  const t = tickData(item)
  if (!t || t.change == null) return ''
  return Number(t.change) >= 0 ? 'up' : 'down'
}
</script>

<template>
  <div v-if="show" class="mobile-backdrop" @click="emit('close')" />
  <aside class="watchlist-sidebar" :class="{ 'mobile-show': show }">
    <div class="sidebar-header">
      <div class="header-tabs">
        <span class="header-title">Watchlist</span>
      </div>
      <div class="header-actions">
        <button class="icon-btn" title="Close" @click="emit('close')">✕</button>
      </div>
    </div>

    <div class="sub-watchlist">
      <div class="sub-tabs">
        <button v-for="wl in watchlists" :key="wl.id" class="sub-tab" :class="{ active: activeWL?.id === wl.id }" @click="activeWL = wl">{{ wl.name }}</button>
      </div>
      <div class="sub-actions">
        <button class="icon-btn sm" title="New watchlist" @click="newWatchlist">➕</button>
      </div>
    </div>

    <!-- Search box right below sub-watchlist -->
    <div class="search-wrap">
      <span class="search-icon">🔍</span>
      <input v-model="searchQ" type="text" placeholder="Search & add symbols..." class="search-input" />
    </div>
    <div v-if="searchQ" class="search-results">
      <div v-if="searching" class="empty-msg">Searching...</div>
      <div v-for="c in searchResults" :key="c.id" class="watchlist-row" @click="addToWatchlist(c)">
        <div class="row-left">
          <div class="symbol-info">
            <span class="symbol-name">{{ c.symbol }}</span>
            <div class="symbol-badge">{{ c.exchange }}</div>
          </div>
        </div>
        <div class="row-right">
          <button class="add-btn" title="Add">+</button>
        </div>
      </div>
      <div v-if="searchQ && !searchResults.length && !searching" class="empty-msg">No results.</div>
    </div>

    <!-- Watchlist items (scrollable) -->
    <div class="watchlist-body">
      <div v-for="(item, idx) in items" :key="item.id" class="watchlist-row" draggable="true"
        @dragstart="dragItem = idx" @dragover.prevent @dragenter="dragOver = idx"
        :class="{ 'drag-over': dragOver === idx }" @drop="onDrop(idx)">
        <div class="row-left">
          <div class="symbol-info">
            <strong class="symbol-name clickable" @click.stop="emit('place-order', item)">{{ item.symbol }}</strong>
            <div class="symbol-badge">{{ item.exchange }}</div>
          </div>
        </div>
        <div class="row-right" :class="priceClass(item)">
          <div class="price-col">
            <span class="ltp-val">{{ ltpDisplay(item) }}</span>
            <span v-if="changeDisplay(item)" class="change-val">
              {{ changeDisplay(item) }} ({{ pctDisplay(item) }})
            </span>
          </div>
          <button class="icon-btn sm remove-btn" title="Remove" @click="removeItem(item)">✕</button>
        </div>
      </div>
      <div v-if="!items.length && watchlists.length" class="empty-msg">No symbols in this list.</div>
      <div v-if="!watchlists.length" class="empty-msg">Create a watchlist to start.</div>
    </div>

    <div class="sidebar-footer">
      <span class="footer-label">{{ watchlists.length }} watchlist{{ watchlists.length !== 1 ? 's' : '' }}</span>
    </div>

    <!-- Prompt modal -->
    <div v-if="showPrompt" class="modal-overlay" @click.self="showPrompt = false">
      <div class="prompt-box">
        <h3>New Watchlist</h3>
        <input v-model="promptName" type="text" placeholder="Watchlist name..." class="prompt-input" @keyup.enter="confirmNew" />
        <div class="prompt-actions">
          <button class="yes" @click="confirmNew">Create</button>
          <button class="no" @click="showPrompt = false">Cancel</button>
        </div>
      </div>
    </div>
  </aside>
</template>

<style scoped>
.watchlist-sidebar {
  width:380px; flex-shrink:0; display:flex; flex-direction:column;
  background:hsl(var(--card)); border-right:1px solid hsl(var(--border));
  font-family:system-ui,-apple-system,sans-serif;
}

.sidebar-header { display:flex; align-items:center; justify-content:space-between; height:48px; padding:0 1rem; border-bottom:1px solid hsl(var(--border)); flex-shrink:0; }
.header-tabs { display:flex; align-items:center; gap:0; height:100%; }
.header-title { font-size:.875rem; font-weight:700; color:hsl(var(--foreground)); }

.header-actions { display:flex; align-items:center; gap:.25rem; }
.icon-btn { width:28px; height:28px; display:flex; align-items:center; justify-content:center; border:none; background:transparent; cursor:pointer; border-radius:4px; font-size:.85rem; color:hsl(var(--muted-foreground)); }
.icon-btn:hover { background:hsl(var(--muted)); color:hsl(var(--foreground)); }
.icon-btn.sm { width:24px; height:24px; font-size:.75rem; }

.sub-watchlist { display:flex; align-items:center; justify-content:space-between; height:40px; padding:0 1rem; border-bottom:1px solid hsl(var(--border)); flex-shrink:0; }
.sub-tabs { display:flex; align-items:center; gap:.5rem; overflow-x:auto; flex:1; }
.sub-tab { border:none; background:transparent; cursor:pointer; font-size:.75rem; font-weight:500; padding:.2rem .4rem; white-space:nowrap; color:hsl(var(--muted-foreground)); border-bottom:2px solid transparent; }
.sub-tab.active { color:hsl(var(--primary)); border-bottom-color:hsl(var(--primary)); }
.sub-tab:hover { color:hsl(var(--foreground)); }
.sub-actions { display:flex; gap:.125rem; }

.search-wrap { position:relative; display:flex; align-items:center; margin:.5rem .75rem; }
.search-icon { position:absolute; left:.625rem; font-size:.8rem; color:hsl(var(--muted-foreground)); pointer-events:none; }
.search-input { width:100%; height:36px; padding:0 .75rem 0 2rem; border:1px solid hsl(var(--input)); border-radius:.5rem; background:hsl(var(--muted)); font-size:.8125rem; color:hsl(var(--foreground)); outline:none; }
.search-input:focus { border-color:hsl(var(--ring)); box-shadow:0 0 0 2px hsl(var(--ring)/.15); }
.search-results { border-top:1px solid hsl(var(--border)); }
.search-icon { position:absolute; left:.625rem; font-size:.8rem; color:hsl(var(--muted-foreground)); pointer-events:none; }
.search-input { width:100%; height:36px; padding:0 .75rem 0 2rem; border:1px solid hsl(var(--input)); border-radius:.5rem; background:hsl(var(--muted)); font-size:.8125rem; color:hsl(var(--foreground)); outline:none; }
.search-input:focus { border-color:hsl(var(--ring)); box-shadow:0 0 0 2px hsl(var(--ring)/.15); }

.watchlist-body { flex:1; overflow-y:auto; }
.watchlist-row { display:flex; align-items:center; justify-content:space-between; padding:.6rem 1rem; border-bottom:1px solid hsl(var(--border)/.5); cursor:pointer; transition:background .1s; }
.watchlist-row:hover { background:hsl(var(--muted)/.5); }
.row-left { display:flex; align-items:center; gap:.5rem; min-width:0; }
.symbol-info { display:flex; flex-direction:column; gap:.1rem; }
.symbol-name { font-size:.8125rem; font-weight:700; color:hsl(var(--foreground)); line-height:1.2; }
.symbol-name.clickable { cursor:pointer; }
.symbol-name.clickable:hover { color:hsl(var(--primary)); }
.symbol-badge { display:inline-block; font-size:.625rem; font-weight:500; color:hsl(var(--muted-foreground)); background:hsl(var(--muted)); padding:0 .25rem; border-radius:3px; width:fit-content; }
.row-right { display:flex; align-items:center; gap:.5rem; }
.row-right.up .price-col { color:#16A34A; }
.row-right.down .price-col { color:hsl(var(--destructive)); }
.price-col { display:flex; flex-direction:column; align-items:flex-end; gap:0; min-width:80px; }
.ltp-val { font-size:.8125rem; font-weight:700; line-height:1.3; }
.change-val { font-size:.6875rem; font-weight:400; white-space:nowrap; }
.remove-btn { opacity:0; transition:opacity .15s; }
.watchlist-row:hover .remove-btn { opacity:1; }
.watchlist-row.drag-over { border-top:2px solid hsl(var(--primary)); }
.add-btn { width:24px; height:24px; display:flex; align-items:center; justify-content:center; border:1px solid hsl(var(--primary)); background:transparent; color:hsl(var(--primary)); border-radius:4px; cursor:pointer; font-size:.8rem; font-weight:700; }
.add-btn:hover { background:hsl(var(--primary)); color:#fff; }

.empty-msg { padding:2rem 1rem; text-align:center; font-size:.8125rem; color:hsl(var(--muted-foreground)); }

.sidebar-footer { display:flex; align-items:center; justify-content:space-between; height:44px; padding:0 1rem; border-top:1px solid hsl(var(--border)); background:hsl(var(--card)); flex-shrink:0; }
.footer-label { font-size:.6875rem; font-weight:700; letter-spacing:.05em; color:hsl(var(--primary)); text-transform:uppercase; }

.modal-overlay { position:fixed; inset:0; background:rgba(0,0,0,.4); display:flex; justify-content:center; align-items:center; z-index:9998; }
.prompt-box { background:hsl(var(--card)); padding:2rem; border-radius:var(--radius); width:90%; max-width:320px; box-shadow:0 4px 24px rgba(0,0,0,.12); }
.prompt-box h3 { margin:0 0 .75rem; }
.prompt-input { width:100%; padding:.5rem .7rem; border:1px solid hsl(var(--input)); border-radius:var(--radius); font-size:var(--font-sm); box-sizing:border-box; }
.prompt-input:focus { outline:none; border-color:hsl(var(--ring)); box-shadow:0 0 0 2px hsl(var(--ring)/.2); }
.prompt-actions { display:flex; gap:.5rem; margin-top:1rem; }
.prompt-actions button { flex:1; padding:.6rem; border:none; border-radius:var(--radius); cursor:pointer; font-weight:500; }
.yes { background:hsl(var(--primary)); color:#fff; }
.no { background:hsl(var(--muted)); color:hsl(var(--foreground)); }

.mobile-backdrop { display:none; }
@media (max-width:768px) {
  .watchlist-sidebar { display:none; }
  .watchlist-sidebar.mobile-show { display:flex; position:fixed; left:0; top:0; bottom:0; z-index:99; height:100vh; }
  .mobile-backdrop { display:block; position:fixed; inset:0; z-index:98; background:rgba(0,0,0,.3); }
}
@media (min-width:769px) { .watchlist-sidebar { display:flex; position:static; height:auto; } }
</style>
