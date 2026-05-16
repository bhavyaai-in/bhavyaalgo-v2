<script setup>
import { ref, watch } from 'vue'

const props = defineProps({ show: Boolean, broker: Object })
const emit = defineEmits(['close'])

const positions = ref([])
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
  loading.value = true; error.value = ''; positions.value = []
  try { positions.value = await api(`/api/broker-positions/${props.broker.id}`) }
  catch (e) { error.value = e.message }
  finally { loading.value = false }
})

function fmt(v) { return v == null || v === '' ? '-' : Number(v).toFixed(2) }
</script>

<template>
  <div v-if="show" class="modal-overlay" @click.self="emit('close')">
    <div class="modal-box">
      <h3>Positions — {{ cap(broker?.friendly_name || broker?.broker_name || '') }}</h3>
      <div v-if="loading" class="state-msg">Loading...</div>
      <div v-else-if="error" class="state-msg error">{{ error }}</div>
      <div v-else-if="!positions.length" class="state-msg">No positions.</div>
      <div v-else class="table-wrap">
        <table class="data-table">
          <thead><tr><th>Trading Symbol</th><th>Exchange</th><th>Qty</th><th>Product</th><th>P&L</th></tr></thead>
          <tbody>
            <tr v-for="p in positions" :key="p.symboltoken || p.tradingsymbol">
              <td><strong>{{ p.tradingsymbol }}</strong></td>
              <td>{{ p.exchange }}</td><td>{{ p.quantity || p.netqty }}</td>
              <td>{{ p.producttype || p.product }}</td>
              <td :class="{negative: Number(p.profitandloss||p.pnl||0)<0}">{{ fmt(p.profitandloss || p.pnl || 0) }}</td>
            </tr>
          </tbody>
        </table>
      </div>
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
.table-wrap { overflow-x:auto; }
.data-table { width:100%; border-collapse:collapse; font-size:var(--font-sm); }
.data-table th, .data-table td { padding:.35rem .4rem; border-bottom:1px solid hsl(var(--border)); text-align:left; white-space:nowrap; }
.data-table th { font-weight:600; color:hsl(var(--foreground)); }
.data-table td { color:hsl(var(--muted-foreground)); }
.data-table td.negative { color:hsl(var(--destructive)); }
.close-btn { display:block; margin:1rem auto 0; padding:.5rem 1.5rem; border:1px solid hsl(var(--border)); background:hsl(var(--card)); border-radius:var(--radius); cursor:pointer; font-weight:500; }
</style>
