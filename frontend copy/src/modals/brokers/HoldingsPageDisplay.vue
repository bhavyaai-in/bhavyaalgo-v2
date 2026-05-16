<script setup>
import { ref, computed, onMounted, watch } from 'vue'
import { useRoute } from 'vue-router'

const props = defineProps({ data: null })

const route = useRoute()
const brokers = ref([])
const selectedId = ref(route.query.broker || '')
const raw = ref(null)
const loading = ref(false)
const error = ref('')
const displayData = computed(() => props.data || raw.value)

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
  raw.value = null
  try {
    raw.value = await api(`/api/broker-holdings/${selectedId.value}`)
  } catch (e) {
    error.value = e.message
  } finally {
    loading.value = false
  }
}

watch(selectedId, () => { if (brokers.value.length) fetchData() })
onMounted(fetchBrokers)

const summaryCards = computed(() => {
  const src = displayData.value?.data || displayData.value || {}
  const t = src.totalholding || {}
  return [
    { key: 'totalholdingvalue', label: 'Total Holding', val: t.totalholdingvalue },
    { key: 'totalinvvalue', label: 'Total Investment', val: t.totalinvvalue },
    { key: 'totalprofitandloss', label: 'Total P&L', val: t.totalprofitandloss },
    { key: 'totalpnlpercentage', label: 'P&L %', val: t.totalpnlpercentage },
  ]
})

const holdings = computed(() => {
  const src = displayData.value?.data || displayData.value || {}
  return src.holdings || []
})
</script>

<template>
  <!-- Embedded mode -->
  <template v-if="props.data">
    <div class="summary-row">
      <div v-for="c in summaryCards" :key="c.key" class="summary-card" :class="{negative: Number(c.val)<0}">
        <div class="summary-label">{{ c.label }}</div>
        <div class="summary-value">{{ c.val ?? '-' }}</div>
      </div>
    </div>
    <div v-if="holdings.length" class="table-wrap">
      <table class="data-table">
        <thead><tr><th>Trading Symbol</th><th>Exchange</th><th>Qty</th><th>Buy Price</th><th>LTP</th><th>Product</th><th>P&L</th><th>P&L %</th></tr></thead>
        <tbody>
          <tr v-for="h in holdings" :key="h.symboltoken">
            <td><strong>{{ h.tradingsymbol }}</strong></td>
            <td>{{ h.exchange }}</td><td>{{ h.quantity }}</td>
            <td>{{ h.averageprice }}</td><td>{{ h.ltp }}</td><td>{{ h.product }}</td>
            <td :class="{negative: ((h.ltp||0)-(h.averageprice||0))*(h.quantity||0) < 0}">{{ (((h.ltp||0)-(h.averageprice||0))*(h.quantity||0)).toFixed(2) }}</td>
            <td :class="{negative: Number(h.pnlpercentage) < 0}">{{ h.pnlpercentage }}%</td>
          </tr>
        </tbody>
      </table>
    </div>
    <div v-else class="state-msg">No holdings.</div>
  </template>

  <!-- Full page mode -->
  <template v-else>
    <div class="page">
      <header>
        <h2>Holdings</h2>
        <select v-model="selectedId" class="broker-select">
          <option v-for="b in brokers" :key="b.id" :value="b.id">{{ b.friendly_name || b.broker_name }}</option>
        </select>
      </header>
      <div v-if="loading" class="state-msg">Loading...</div>
      <div v-else-if="error" class="state-msg error">{{ error }}</div>
      <template v-else-if="displayData">
        <div class="summary-row">
          <div v-for="c in summaryCards" :key="c.key" class="summary-card" :class="{negative: Number(c.val)<0}">
            <div class="summary-label">{{ c.label }}</div>
            <div class="summary-value">{{ c.val ?? '-' }}</div>
          </div>
        </div>
        <div v-if="holdings.length" class="table-wrap">
          <table class="data-table">
            <thead><tr><th>Trading Symbol</th><th>Exchange</th><th>Qty</th><th>Buy Price</th><th>LTP</th><th>Product</th><th>P&L</th><th>P&L %</th></tr></thead>
            <tbody>
              <tr v-for="h in holdings" :key="h.symboltoken">
                <td><strong>{{ h.tradingsymbol }}</strong></td>
                <td>{{ h.exchange }}</td><td>{{ h.quantity }}</td>
                <td>{{ h.averageprice }}</td><td>{{ h.ltp }}</td><td>{{ h.product }}</td>
                <td :class="{negative: ((h.ltp||0)-(h.averageprice||0))*(h.quantity||0) < 0}">{{ (((h.ltp||0)-(h.averageprice||0))*(h.quantity||0)).toFixed(2) }}</td>
                <td :class="{negative: Number(h.pnlpercentage) < 0}">{{ h.pnlpercentage }}%</td>
              </tr>
            </tbody>
          </table>
        </div>
        <div v-else class="state-msg">No holdings.</div>
      </template>
    </div>
  </template>
</template>

<style scoped>
.page { padding: 1rem 0; }
header { display:flex; justify-content:space-between; align-items:center; margin-bottom:1.25rem; }
header h2 { margin:0; }
.broker-select { padding:.5rem 1rem; border:1px solid hsl(var(--primary)); border-radius:var(--radius); font-size:var(--font-sm); background:hsl(var(--primary)); color:hsl(var(--primary-foreground)); cursor:pointer; font-weight:500; }
.state-msg { text-align:center; padding:2rem; color:hsl(var(--muted-foreground)); }
.state-msg.error { color:hsl(var(--destructive)); }
.summary-row { display:flex; gap:.5rem; margin-bottom:1rem; flex-wrap:wrap; }
.summary-card { flex:1; min-width:100px; background:hsl(var(--card)); border:1px solid hsl(var(--border)); border-radius:var(--radius); padding:.6rem; text-align:center; }
.summary-label { font-size:var(--font-xs); color:hsl(var(--muted-foreground)); }
.summary-value { font-size:var(--font-base); font-weight:700; color:hsl(var(--foreground)); }
.summary-card.negative .summary-value { color:hsl(var(--destructive)); }
.table-wrap { overflow-x:auto; }
.data-table { width:100%; border-collapse:collapse; font-size:var(--font-sm); }
.data-table th, .data-table td { padding:.4rem .5rem; border-bottom:1px solid hsl(var(--border)); text-align:left; white-space:nowrap; }
.data-table th { font-weight:600; color:hsl(var(--foreground)); position:sticky; top:0; background:hsl(var(--card)); }
.data-table td { color:hsl(var(--muted-foreground)); }
.data-table td.negative { color:hsl(var(--destructive)); }
</style>
