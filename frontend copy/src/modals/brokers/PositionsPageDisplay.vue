<script setup>
import { ref, computed, onMounted, watch } from 'vue'
import { useRoute } from 'vue-router'

const props = defineProps({ data: null })
const route = useRoute()
const brokers = ref([])
const selectedId = ref(route.query.broker || '')
const positions = ref([])
const loading = ref(false)
const error = ref('')
const display = computed(() => props.data || positions.value)

function token() { return localStorage.getItem('token') }
async function api(path, opts = {}) {
  const res = await fetch(path, { headers: { 'Content-Type': 'application/json', Authorization: token() }, ...opts })
  if (!res.ok) throw new Error(await res.text())
  return res.json()
}
async function fetchBrokers() {
  brokers.value = await api('/api/brokers')
  if (brokers.value.length > 0 && !selectedId.value) selectedId.value = String(brokers.value[0].id)
}
async function fetchData() {
  if (!selectedId.value) return
  loading.value = true; error.value = ''; positions.value = []
  try { positions.value = await api(`/api/broker-positions/${selectedId.value}`) }
  catch (e) { error.value = e.message }
  finally { loading.value = false }
}
watch(selectedId, () => { if (brokers.value.length) fetchData() })
onMounted(fetchBrokers)
function fmt(v) { return v == null || v === '' ? '-' : Number(v).toFixed(2) }
</script>

<template>
  <template v-if="props.data">
    <div v-if="display.length" class="table-wrap">
      <table class="data-table">
        <thead><tr><th>Trading Symbol</th><th>Exchange</th><th>Qty</th><th>Product</th><th>P&L</th></tr></thead>
        <tbody>
          <tr v-for="p in display" :key="p.symboltoken || p.tradingsymbol">
            <td><strong>{{ p.tradingsymbol }}</strong></td>
            <td>{{ p.exchange }}</td><td>{{ p.quantity || p.netqty }}</td>
            <td>{{ p.producttype || p.product }}</td>
            <td :class="{negative: Number(p.profitandloss||p.pnl||0)<0}">{{ fmt(p.profitandloss || p.pnl || 0) }}</td>
          </tr>
        </tbody>
      </table>
    </div>
    <div v-else class="state-msg">No positions.</div>
  </template>

  <template v-else>
    <div class="page">
      <header>
        <h2>Positions</h2>
        <select v-model="selectedId" class="broker-select">
          <option v-for="b in brokers" :key="b.id" :value="b.id">{{ b.friendly_name || b.broker_name }}</option>
        </select>
      </header>
      <div v-if="loading" class="state-msg">Loading...</div>
      <div v-else-if="error" class="state-msg error">{{ error }}</div>
      <div v-else-if="!display.length" class="state-msg">No positions.</div>
      <div v-else class="table-wrap">
        <table class="data-table">
          <thead><tr><th>Trading Symbol</th><th>Exchange</th><th>Qty</th><th>Product</th><th>P&L</th></tr></thead>
          <tbody>
            <tr v-for="p in display" :key="p.symboltoken || p.tradingsymbol">
              <td><strong>{{ p.tradingsymbol }}</strong></td>
              <td>{{ p.exchange }}</td><td>{{ p.quantity || p.netqty }}</td>
              <td>{{ p.producttype || p.product }}</td>
              <td :class="{negative: Number(p.profitandloss||p.pnl||0)<0}">{{ fmt(p.profitandloss || p.pnl || 0) }}</td>
            </tr>
          </tbody>
        </table>
      </div>
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
.table-wrap { overflow-x:auto; }
.data-table { width:100%; border-collapse:collapse; font-size:var(--font-sm); }
.data-table th, .data-table td { padding:.4rem .5rem; border-bottom:1px solid hsl(var(--border)); text-align:left; white-space:nowrap; }
.data-table th { font-weight:600; color:hsl(var(--foreground)); position:sticky; top:0; background:hsl(var(--card)); }
.data-table td { color:hsl(var(--muted-foreground)); }
.data-table td.negative { color:hsl(var(--destructive)); }
</style>
