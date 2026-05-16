<script setup>
import { ref, computed, watch } from 'vue'

const props = defineProps({ show: Boolean, broker: Object })
const emit = defineEmits(['close'])

const raw = ref(null)
const loading = ref(false)
const error = ref('')

function cap(str) { if (!str) return ''; return str.replace(/\b\w/g, c => c.toUpperCase()) }
function token() { return localStorage.getItem('token') }
async function api(path, opts = {}) {
  const res = await fetch(path, { headers: { 'Content-Type': 'application/json', Authorization: token() }, ...opts })
  if (!res.ok) throw new Error(await res.text())
  return res.json()
}
watch(() => props.show, async (val) => {
  if (!val || !props.broker) return
  loading.value = true; error.value = ''; raw.value = null
  try { raw.value = await api(`/api/broker-holdings/${props.broker.id}`) }
  catch (e) { error.value = e.message }
  finally { loading.value = false }
})

const summaryCards = computed(() => {
  const t = raw.value?.data?.totalholding || {}
  return [
    { k: 'totalholdingvalue', l: 'Total Holding', v: t.totalholdingvalue },
    { k: 'totalinvvalue', l: 'Total Investment', v: t.totalinvvalue },
    { k: 'totalprofitandloss', l: 'Total P&L', v: t.totalprofitandloss },
    { k: 'totalpnlpercentage', l: 'P&L %', v: t.totalpnlpercentage },
  ]
})

const holdings = computed(() => raw.value?.data?.holdings || [])
</script>

<template>
  <div v-if="show" class="modal-overlay" @click.self="emit('close')">
    <div class="modal-box">
      <h3>Holdings — {{ cap(broker?.friendly_name || broker?.broker_name || '') }}</h3>
      <div v-if="loading" class="state-msg">Loading...</div>
      <div v-else-if="error" class="state-msg error">{{ error }}</div>
      <template v-else-if="raw">
        <div class="summary-row">
          <div v-for="c in summaryCards" :key="c.k" class="sc" :class="{negative: Number(c.v)<0}">
            <div class="sl">{{ c.l }}</div>
            <div class="sv">{{ c.v ?? '-' }}</div>
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
                <td :class="{negative: ((h.ltp||0)-(h.averageprice||0))*(h.quantity||0)<0}">{{ (((h.ltp||0)-(h.averageprice||0))*(h.quantity||0)).toFixed(2) }}</td>
                <td :class="{negative: Number(h.pnlpercentage)<0}">{{ h.pnlpercentage }}%</td>
              </tr>
            </tbody>
          </table>
        </div>
        <div v-else class="state-msg">No holdings.</div>
      </template>
      <button class="close-btn" @click="emit('close')">Close</button>
    </div>
  </div>
</template>

<style scoped>
.modal-overlay { position:fixed; inset:0; background:rgba(0,0,0,.4); display:flex; justify-content:center; align-items:center; z-index:100; }
.modal-box { background:hsl(var(--card)); border-radius:var(--radius); padding:1.5rem; width:90%; max-width:700px; max-height:80vh; overflow-y:auto; box-shadow:0 4px 24px rgba(0,0,0,.12); }
.modal-box h3 { margin:0 0 1rem; font-size:var(--font-base); }
.state-msg { text-align:center; padding:2rem; color:hsl(var(--muted-foreground)); }
.state-msg.error { color:hsl(var(--destructive)); }
.summary-row { display:flex; gap:.4rem; margin-bottom:.75rem; flex-wrap:wrap; }
.sc { flex:1; min-width:80px; background:hsl(var(--card)); border:1px solid hsl(var(--border)); border-radius:var(--radius); padding:.4rem; text-align:center; }
.sl { font-size:var(--font-xs); color:hsl(var(--muted-foreground)); }
.sv { font-size:var(--font-sm); font-weight:700; color:hsl(var(--foreground)); }
.sc.negative .sv { color:hsl(var(--destructive)); }
.table-wrap { overflow-x:auto; }
.data-table { width:100%; border-collapse:collapse; font-size:var(--font-sm); }
.data-table th, .data-table td { padding:.35rem .4rem; border-bottom:1px solid hsl(var(--border)); text-align:left; white-space:nowrap; }
.data-table th { font-weight:600; color:hsl(var(--foreground)); }
.data-table td { color:hsl(var(--muted-foreground)); }
.data-table td.negative { color:hsl(var(--destructive)); }
.close-btn { display:block; margin:1rem auto 0; padding:.5rem 1.5rem; border:1px solid hsl(var(--border)); background:hsl(var(--card)); border-radius:var(--radius); cursor:pointer; font-weight:500; }
</style>
