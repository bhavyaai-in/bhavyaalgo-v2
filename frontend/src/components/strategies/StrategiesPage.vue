<script setup>
import { ref, onMounted } from 'vue'
import { api, confirm } from '../../utils/api.js'
import StrategyFormModal from './StrategyFormModal.vue'
import StrategyDetailModal from './StrategyDetailModal.vue'
import PlaceOrderModal from '../../modals/brokers/PlaceOrderModal.vue'

const strategies = ref([])
const types = ref([])
const brokers = ref([])
const loading = ref(true)
const error = ref('')
const showForm = ref(false)
const editing = ref(null)
const detailData = ref(null)
const showDetail = ref(false)
const detailTab = ref('overview')
const showPlaceOrder = ref(false)
const placeOrderStrategy = ref(null)

async function fetchStrategies() {
  try {
    const data = await api('/api/strategies')
    strategies.value = Array.isArray(data) ? data : []
  } catch { strategies.value = [] }
}

async function fetchTypes() {
  try { types.value = await api('/api/strategy-types') }
  catch { types.value = [] }
}

async function fetchBrokers() {
  try { brokers.value = await api('/api/brokers') }
  catch { brokers.value = [] }
}

function cap(str) { if (!str) return ''; return str.replace(/\b\w/g, c => c.toUpperCase()) }

function typeName(id) {
  const t = types.value.find(t => t.id === id)
  return t ? t.name : '-'
}

function brokerName(id) {
  const b = brokers.value.find(b => b.id === id)
  return b ? (b.friendly_name || cap(b.broker_name)) : '-'
}

function openCreate() { editing.value = null; showForm.value = true }
function openEdit(s) { editing.value = s; showForm.value = true }

function onFormSaved() { showForm.value = false; editing.value = null; fetchStrategies() }

async function deleteStrategy(s) {
  if (!await confirm('Delete Strategy', 'Delete "' + s.name + '"?')) return
  await api('/api/strategies/' + s.id, { method: 'DELETE' })
  fetchStrategies()
}

async function viewDetail(s) {
  try {
    const [joiners, positions, orders] = await Promise.all([
      api('/api/strategies/' + s.id + '/joiners'),
      api('/api/strategies/' + s.id + '/positions'),
      api('/api/strategies/' + s.id + '/orders'),
    ])
    detailData.value = { strategy: s, joiners, positions, orders }
    detailTab.value = 'overview'
    showDetail.value = true
  } catch {}
}

async function viewJoiners(s) {
  try {
    const [joiners, positions, orders] = await Promise.all([
      api('/api/strategies/' + s.id + '/joiners'),
      api('/api/strategies/' + s.id + '/positions'),
      api('/api/strategies/' + s.id + '/orders'),
    ])
    detailData.value = { strategy: s, joiners, positions, orders }
    detailTab.value = 'joiners'
    showDetail.value = true
  } catch {}
}

onMounted(() => {
  Promise.allSettled([fetchStrategies(), fetchTypes(), fetchBrokers()])
    .finally(() => { loading.value = false })
})
</script>

<template>
  <div class="page">
    <header>
      <h2>Strategies</h2>
      <button class="add-btn" @click="openCreate">+ New Strategy</button>
    </header>

    <div v-if="loading" class="state-msg">Loading...</div>
    <div v-else-if="error" class="state-msg error">{{ error }}</div>
    <div v-else-if="!strategies.length" class="state-msg">No strategies yet.</div>

    <div v-else class="strategy-grid">
      <div v-for="s in strategies" :key="s.id" class="strategy-card" @click="viewDetail(s)">
        <div class="sc-top">
          <span class="sc-type">{{ typeName(s.strategy_type) }}</span>
          <span v-if="s.is_active" class="sc-badge active">Active</span>
          <span v-else class="sc-badge">Inactive</span>
        </div>
        <div class="sc-name">{{ s.name }}</div>
        <div class="sc-meta">
          <span>Side: {{ cap(s.side) }}</span>
          <span>Exp: {{ s.expiry_date }}</span>
        </div>
        <div class="sc-actions" @click.stop>
          <button class="chip" @click.stop="openEdit(s)">Edit</button>
          <button class="chip danger" @click.stop="deleteStrategy(s)">Delete</button>
          <button class="chip" @click.stop="viewJoiners(s)">Joiners</button>
          <button class="chip primary" @click.stop="placeOrderStrategy = s; showPlaceOrder = true">Place Order</button>
        </div>
      </div>
    </div>

    <PlaceOrderModal
      :show="showPlaceOrder"
      :strategy="placeOrderStrategy"
      @close="showPlaceOrder = false; placeOrderStrategy = null"
    />

    <StrategyFormModal
      :show="showForm"
      :strategy="editing"
      :types="types"
      :brokers="brokers"
      @close="showForm = false; editing = null"
      @saved="onFormSaved"
    />
    <StrategyDetailModal
      :show="showDetail"
      :data="detailData"
      :initial-tab="detailTab"
      :type-name="typeName"
      :broker-name="brokerName"
      :brokers="brokers"
      @close="showDetail = false; detailData = null"
    />
  </div>
</template>

<style scoped>
.page { padding: 0; }
header { display:flex; justify-content:space-between; align-items:center; margin-bottom:1rem; }
header h2 { margin:0; }
.add-btn {
  padding:.5rem 1rem; border:none; border-radius:var(--radius);
  background:hsl(var(--primary)); color:#fff; font-weight:500; cursor:pointer; font-size:var(--font-sm);
}
.strategy-grid { display:grid; grid-template-columns:repeat(auto-fill,minmax(260px,1fr)); gap:.75rem; }
.strategy-card {
  background:hsl(var(--card)); border:1px solid hsl(var(--border));
  border-radius:var(--radius); padding:.75rem 1rem; cursor:pointer;
  transition:border-color .15s; display:flex; flex-direction:column; gap:.5rem;
}
.strategy-card:hover { border-color:hsl(var(--primary)); }
.sc-top { display:flex; justify-content:space-between; align-items:center; }
.sc-type { font-size:var(--font-xs); color:hsl(var(--muted-foreground)); }
.sc-badge {
  font-size:.625rem; font-weight:600; padding:1px 6px; border-radius:999px;
  background:hsl(var(--muted)); color:hsl(var(--muted-foreground));
}
.sc-badge.active { background:hsl(144 80% 55% / .15); color:#16A34A; }
.sc-name { font-size:var(--font-base); font-weight:700; color:hsl(var(--foreground)); }
.sc-meta { display:flex; gap:.75rem; font-size:var(--font-xs); color:hsl(var(--muted-foreground)); }
.sc-actions { display:flex; gap:.4rem; margin-top:.25rem; }

.chip.primary:hover { opacity:.9; }
.chip.danger:hover { border-color:hsl(var(--destructive)); color:hsl(var(--destructive)); }
.state-msg { text-align:center; padding:2rem; color:hsl(var(--muted-foreground)); }
.state-msg.error { color:hsl(var(--destructive)); }
</style>
