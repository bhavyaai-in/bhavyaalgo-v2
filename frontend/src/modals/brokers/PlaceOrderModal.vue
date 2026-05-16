<script setup>
import { ref, watch } from 'vue'

const props = defineProps({ show: Boolean, broker: Object })
const emit = defineEmits(['close'])

const success = ref(false)
const tradingSymbol = ref('')
const symbolToken = ref('')
const exchange = ref('NSE')
const transactionType = ref('BUY')
const quantity = ref(1)
const price = ref(0)
const triggerPrice = ref(0)
const orderType = ref('MARKET')
const productType = ref('DELIVERY')
const duration = ref('DAY')
const variety = ref('NORMAL')
const squareoff = ref(0)
const stoploss = ref(0)
const submitting = ref(false)
const error = ref('')

function token() { return localStorage.getItem('token') }

async function api(path, opts = {}) {
  const res = await fetch(path, {
    headers: { 'Content-Type': 'application/json', Authorization: token() },
    ...opts,
  })
  if (!res.ok) throw new Error(await res.text())
  return res.json()
}

watch(() => props.show, (val) => {
  if (!val) return
  tradingSymbol.value = ''
  symbolToken.value = ''
  exchange.value = 'NSE'
  transactionType.value = 'BUY'
  quantity.value = 1
  price.value = 0
  triggerPrice.value = 0
  orderType.value = 'MARKET'
  productType.value = 'DELIVERY'
  duration.value = 'DAY'
  variety.value = 'NORMAL'
  squareoff.value = 0
  stoploss.value = 0
  success.value = false
  error.value = ''
  submitting.value = false
})

async function placeOrder() {
  submitting.value = true
  error.value = ''
  const data = {
    variety: variety.value,
    tradingsymbol: tradingSymbol.value,
    symboltoken: symbolToken.value,
    transactiontype: transactionType.value,
    exchange: exchange.value,
    ordertype: orderType.value,
    producttype: productType.value,
    duration: duration.value,
    price: String(price.value),
    triggerprice: String(triggerPrice.value),
    squareoff: String(squareoff.value),
    stoploss: String(stoploss.value),
    quantity: String(quantity.value),
  }
  try {
    await api('/api/broker-place-order', {
      method: 'POST',
      body: JSON.stringify({ broker_id: props.broker.id, data }),
    })
    success.value = true
    setTimeout(() => emit('close'), 3000)
  } catch (e) {
    error.value = e.message
  } finally {
    submitting.value = false
  }
}

function cap(str) { if (!str) return ''; return str.replace(/\b\w/g, c => c.toUpperCase()) }
</script>

<template>
  <div v-if="show" class="modal-overlay" @click.self="emit('close')">
    <div class="modal-box">
      <h3>Place Order — {{ cap(broker?.friendly_name || broker?.broker_name || '') }}</h3>

      <template v-if="success">
        <div class="success-anim">
          <svg width="48" height="48" viewBox="0 0 24 24" fill="none" stroke="#16A34A" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M22 11.08V12a10 10 0 1 1-5.93-9.14"/><polyline points="22 4 12 14.01 9 11.01"/></svg>
          <div class="success-text">Order Placed!</div>
        </div>
      </template>

      <template v-else-if="error && !submitting">
        <div class="error-msg">{{ error }}</div>
        <button class="close-btn" @click="emit('close')">Close</button>
      </template>

      <template v-else>
        <div class="field-grid">
          <label>Trading Symbol <input v-model="tradingSymbol" required /></label>
          <label>Symbol Token <input v-model="symbolToken" required /></label>
          <label>Exchange <select v-model="exchange"><option>NSE</option><option>BSE</option><option>NFO</option><option>MCX</option></select></label>
          <label>Type <select v-model="transactionType"><option>BUY</option><option>SELL</option></select></label>
          <label>Qty <input v-model.number="quantity" type="number" min="1" /></label>
          <label>Price <input v-model.number="price" type="number" step="0.05" min="0" /></label>
          <label>Trigger Price <input v-model.number="triggerPrice" type="number" step="0.05" min="0" /></label>
          <label>Order Type <select v-model="orderType"><option>MARKET</option><option>LIMIT</option><option>SL</option><option>SL-M</option></select></label>
          <label>Product <select v-model="productType"><option>DELIVERY</option><option>CARRYFORWARD</option><option>MARGIN</option><option>MIS</option></select></label>
          <label>Square Off <input v-model.number="squareoff" type="number" step="0.05" min="0" /></label>
          <label>Stop Loss <input v-model.number="stoploss" type="number" step="0.05" min="0" /></label>
        </div>
        <div v-if="error" class="error-msg">{{ error }}</div>
        <div class="form-actions">
          <button @click="placeOrder" :disabled="submitting || !tradingSymbol || !symbolToken">{{ submitting ? 'Placing...' : 'Place Order' }}</button>
          <button class="cancel" @click="emit('close')">Cancel</button>
        </div>
      </template>
    </div>
  </div>
</template>

<style scoped>
.modal-box { max-width:480px; }
.field-grid { display:grid; grid-template-columns:1fr 1fr; gap:.6rem; }
.field-grid label { display:flex; flex-direction:column; gap:.2rem; font-size:var(--font-sm); color:hsl(var(--foreground)); }
.field-grid input, .field-grid select { padding:.45rem .6rem; border:1px solid hsl(var(--input)); border-radius:var(--radius); font-size:var(--font-sm); }
.field-grid input:focus, .field-grid select:focus { outline:none; border-color:hsl(var(--ring)); box-shadow:0 0 0 2px hsl(var(--ring)/.2); }
.error-msg { padding:.5rem; color:hsl(var(--destructive)); font-size:var(--font-sm); text-align:center; }
.success-anim { display:flex; flex-direction:column; align-items:center; gap:.5rem; padding:2rem; }
.success-anim svg { animation: popIn .5s cubic-bezier(.68,-.55,.27,1.55); }
.success-text { font-size:var(--font-lg); font-weight:700; color:#16A34A; animation: fadeIn .5s ease; }
@keyframes popIn { 0% { transform:scale(0); opacity:0; } 100% { transform:scale(1); opacity:1; } }
@keyframes fadeIn { 0% { opacity:0; transform:translateY(10px); } 100% { opacity:1; transform:translateY(0); } }
.form-actions { display:flex; gap:.5rem; margin-top:.75rem; }
.form-actions button { flex:1; padding:.6rem; border:none; border-radius:var(--radius); cursor:pointer; font-weight:500; color:hsl(var(--primary-foreground)); background:hsl(var(--primary)); }
.form-actions button:disabled { opacity:.5; }
.form-actions .cancel { background:hsl(var(--muted-foreground)); }
.close-btn { display:block; margin:1rem auto 0; padding:.5rem 1.5rem; border:1px solid hsl(var(--border)); background:hsl(var(--card)); border-radius:var(--radius); cursor:pointer; font-weight:500; }
.table-wrap { overflow-x:auto; }
.data-table { width:100%; border-collapse:collapse; font-size:var(--font-sm); }
.data-table td { padding:.5rem .75rem; border-bottom:1px solid hsl(var(--border)); }
.data-table .pkey { font-weight:600; color:hsl(var(--foreground)); width:40%; }
.data-table td:last-child { color:hsl(var(--muted-foreground)); }
</style>
