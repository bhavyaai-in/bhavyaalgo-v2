<script setup>
import { ref, computed, watch } from 'vue'
import { api } from '../../utils/api.js'
import { useWebSocket } from '../../composables/useWebSocket.js'

const props = defineProps({ show: Boolean, broker: Object, strategy: Object, presetContract: { type: Object, default: null } })
const emit = defineEmits(['close'])

const ws = useWebSocket()

const selectedContract = ref(null)
const searchQ = ref('')
const savedSearchQ = ref('')
const editing = ref(true)
const searchResults = ref([])
const searching = ref(false)
const showResults = ref(false)
const transactionType = ref('BUY')
const quantity = ref(1)
const price = ref(0)
const priceLinked = ref(false)
const triggerPrice = ref(0)
const orderType = ref('MARKET')
const productType = ref('DELIVERY')
const variety = ref('NORMAL')
const squareoff = ref(0)
const stoploss = ref(0)
const submitting = ref(false)
const error = ref('')
const success = ref(false)
const ltpMap = ref({})

ws.onTick((tick) => {
  if (tick.token && tick.ltp != null) {
    ltpMap.value = { ...ltpMap.value, [tick.token]: tick }
  }
})

const tick = computed(() => selectedContract.value ? ltpMap.value[selectedContract.value.token] : null)
const ltp = computed(() => {
  if (!tick.value) return null
  const value = tick.value.ltp != null ? tick.value.ltp : (tick.value.close != null ? tick.value.close : 0)
  return Number(value).toFixed(2)
})
const changeVal = computed(() => {
  if (!tick.value || tick.value.change == null) return null
  const ch = Number(tick.value.change)
  return `${ch >= 0 ? '+' : ''}${ch.toFixed(2)}`
})
const pctVal = computed(() => {
  if (!tick.value || tick.value.change == null || !tick.value.close) return null
  const pct = (Number(tick.value.change) / Number(tick.value.close)) * 100
  return `${pct >= 0 ? '+' : ''}${pct.toFixed(2)}%`
})

watch(() => props.show, async (val) => {
  if (!val) return
  selectedContract.value = null
  searchQ.value = ''
  savedSearchQ.value = ''
  editing.value = true
  searchResults.value = []
  transactionType.value = 'BUY'
  quantity.value = 1
  price.value = 0
  priceLinked.value = false
  triggerPrice.value = 0
  orderType.value = 'MARKET'
  productType.value = 'DELIVERY'
  variety.value = 'NORMAL'
  squareoff.value = 0
  stoploss.value = 0
  success.value = false
  error.value = ''
  submitting.value = false
  ltpMap.value = {}
  if (props.presetContract) {
    selectContract(props.presetContract)
  }
})

watch(searchQ, async (q) => {
  if (!q || q.length < 2) { searchResults.value = []; showResults.value = false; return }
  searching.value = true
  showResults.value = true
  try {
    searchResults.value = await api(`/api/search-contracts?q=${encodeURIComponent(q)}`)
  } catch { searchResults.value = [] }
  finally { searching.value = false }
})

// Auto-link price to LTP when LIMIT is selected and price not manually changed
watch(orderType, (val) => {
  if (val === 'LIMIT' && ltp.value) {
    priceLinked.value = true
    price.value = Number(ltp.value)
  } else {
    priceLinked.value = false
  }
})

watch(ltp, (val) => {
  if (orderType.value === 'LIMIT' && priceLinked.value && val) {
    price.value = Number(val)
  }
})

watch(selectedContract, () => {
  if (orderType.value === 'LIMIT' && ltp.value) {
    priceLinked.value = true
    price.value = Number(ltp.value)
  }
})

function onPriceInput() {
  priceLinked.value = false
}

function selectContract(c) {
  savedSearchQ.value = searchQ.value
  selectedContract.value = c
  searchQ.value = ''
  editing.value = false
  showResults.value = false
  searchResults.value = []
  if (c.token && c.exchange) {
    ws.subscribe([c.exchange + '|' + c.token])
  }
}

function editContract() {
  editing.value = true
  searchQ.value = savedSearchQ.value
}

function clearContract() {
  if (selectedContract.value) {
    ws.unsubscribe([selectedContract.value.exchange + '|' + selectedContract.value.token])
  }
  selectedContract.value = null
  savedSearchQ.value = ''
  editing.value = true
  searchQ.value = ''
}

async function getConnectedBroker() {
  if (props.broker) return props.broker
  try {
    const brokers = await api('/api/brokers')
    return brokers.find(b => b.token_status === 'connected') || null
  } catch { return null }
}

async function placeOrder() {
  submitting.value = true
  error.value = ''
  const tradingsymbol = selectedContract.value?.exchange === 'NSE'
    ? selectedContract.value.symbol + '-EQ'
    : selectedContract.value.symbol
  const data = {
    variety: variety.value,
    tradingsymbol,
    symboltoken: selectedContract.value.token,
    transactiontype: transactionType.value,
    exchange: selectedContract.value.exchange,
    ordertype: orderType.value,
    producttype: productType.value === 'MIS' ? 'INTRADAY' : productType.value,
    duration: 'DAY',
    price: String(price.value),
    triggerprice: String(triggerPrice.value),
    squareoff: String(squareoff.value),
    stoploss: String(stoploss.value),
    quantity: String(quantity.value),
  }
  try {
    if (props.strategy) {
      const res = await api('/api/strategies/' + props.strategy.id + '/place-order', {
        method: 'POST',
        body: JSON.stringify({ data }),
      })
      const failed = (res.results || []).filter(r => !r.success)
      if (failed.length) {
        error.value = failed.map(r => (r.broker_name || '') + ': ' + (r.error || 'failed')).join('; ')
      } else {
        success.value = true
        setTimeout(() => emit('close'), 3000)
      }
    } else {
      const broker = await getConnectedBroker()
      if (!broker) {
        error.value = 'No connected broker found'
        submitting.value = false
        return
      }
      await api('/api/broker-place-order', {
        method: 'POST',
        body: JSON.stringify({ broker_id: broker.id, data }),
      })
      success.value = true
      setTimeout(() => emit('close'), 3000)
    }
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
      <h3>Place Order</h3>

      <template v-if="success">
        <div class="success-anim">
          <svg width="48" height="48" viewBox="0 0 24 24" fill="none" stroke="#16A34A" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M22 11.08V12a10 10 0 1 1-5.93-9.14"/><polyline points="22 4 12 14.01 9 11.01"/></svg>
          <div class="success-text">Order Placed!</div>
        </div>
      </template>

      <template v-else>
        <!-- Search box (hidden when contract selected and not editing) -->
        <div v-show="!selectedContract || editing" class="search-wrap">
          <label class="field-label">Trading Symbol</label>
          <div class="search-input-wrap">
            <input
              v-model="searchQ"
              type="text"
              class="search-input"
              placeholder="Search symbol..."
              @focus="showResults = searchResults.length > 0"
            />
            <button v-if="selectedContract" class="clear-btn" @click="clearContract">&times;</button>
          </div>
          <div v-if="showResults && searchResults.length" class="search-dropdown">
            <div
              v-for="c in searchResults"
              :key="c.token + c.exchange"
              class="search-item"
              @click="selectContract(c)"
            >
              <strong>{{ c.symbol }}</strong>
              <span class="search-item-meta">{{ c.exchange }} {{ c.name || c.symbol }}</span>
            </div>
          </div>
          <div v-else-if="showResults && !searching && searchQ.length >= 2 && !searchResults.length" class="search-dropdown">
            <div class="search-empty">No results</div>
          </div>
        </div>

        <!-- Selected contract info bar (hidden when editing) -->
        <div v-if="selectedContract && !editing" class="contract-bar" :class="transactionType === 'BUY' ? 'buy-bg' : 'sell-bg'" @dblclick="editContract">
          <div class="contract-left">
            <div class="contract-symbol">{{ selectedContract.symbol }}</div>
            <div class="contract-exch">{{ selectedContract.exchange }}</div>
          </div>
          <div class="contract-right">
            <div class="contract-ltp">{{ ltp || '-' }}</div>
            <div v-if="changeVal" class="contract-change">{{ changeVal }} ({{ pctVal }})</div>
          </div>
        </div>

        <!-- Buy/Sell toggle -->
        <div class="bs-row">
          <div class="bs-group">
            <button class="bs-btn" :class="{ active: transactionType === 'BUY' }" @click="transactionType = 'BUY'">Buy</button>
            <button class="bs-btn" :class="{ active: transactionType === 'SELL' }" @click="transactionType = 'SELL'">Sell</button>
          </div>
        </div>

        <div class="field-grid">
          <label>Qty <input v-model.number="quantity" type="number" min="1" /></label>
          <label>Order Type
            <select v-model="orderType">
              <option>MARKET</option><option>LIMIT</option><option>SL</option><option>SL-M</option>
            </select>
          </label>
          <label>Price
            <div class="price-wrap">
              <input v-model.number="price" type="number" step="0.05" min="0" :disabled="orderType === 'MARKET'" @input="onPriceInput" />
              <span v-if="orderType === 'LIMIT' && priceLinked" class="price-link-badge" title="Auto-linked to LTP">LTP</span>
            </div>
          </label>
          <label v-if="orderType === 'SL' || orderType === 'SL-M'">Trigger Price <input v-model.number="triggerPrice" type="number" step="0.05" min="0" /></label>
          <label>Product
            <select v-model="productType">
              <option>DELIVERY</option><option>CARRYFORWARD</option><option>MARGIN</option><option>MIS</option>
            </select>
          </label>
          <label>Square Off <input v-model.number="squareoff" type="number" step="0.05" min="0" /></label>
          <label>Stop Loss <input v-model.number="stoploss" type="number" step="0.05" min="0" /></label>
        </div>
        <div v-if="error" class="error-msg">{{ error }}</div>
        <div class="form-actions">
          <button class="cancel" @click="emit('close')">Cancel</button>
          <button class="order-btn" :class="transactionType === 'BUY' ? 'buy' : 'sell'" @click="placeOrder" :disabled="submitting || !selectedContract">{{ submitting ? 'Placing...' : (transactionType === 'BUY' ? 'Buy' : 'Sell') }}</button>
        </div>
      </template>
    </div>
  </div>
</template>

<style scoped>
.modal-box { max-width:480px; }

/* Search */
.search-wrap { position: relative; margin-bottom: .5rem; }
.field-label { display:block; font-size:var(--font-sm); color:hsl(var(--foreground)); margin-bottom:.25rem; font-weight:500; }
.search-input-wrap { position: relative; }
.search-input {
  width:100%; padding:.45rem .6rem; padding-right:2rem;
  border:1px solid hsl(var(--input)); border-radius:var(--radius);
  font-size:var(--font-sm); outline:none; background:hsl(var(--card));
}
.search-input:focus { border-color:hsl(var(--ring)); box-shadow:0 0 0 2px hsl(var(--ring)/.2); }
.clear-btn {
  position:absolute; right:4px; top:50%; transform:translateY(-50%);
  border:none; background:none; cursor:pointer;
  font-size:1.2rem; line-height:1; color:hsl(var(--muted-foreground)); padding:2px;
}
.search-dropdown {
  position:absolute; top:100%; left:0; right:0; z-index:10;
  background:hsl(var(--card)); border:1px solid hsl(var(--border));
  border-radius:var(--radius); max-height:200px; overflow-y:auto;
  box-shadow:0 4px 12px rgba(0,0,0,.1);
}
.search-item {
  padding:.4rem .6rem; cursor:pointer; display:flex; flex-direction:column; gap:1px;
  border-bottom:1px solid hsl(var(--border)/.5);
}
.search-item:last-child { border-bottom:none; }
.search-item:hover { background:hsl(var(--muted)); }
.search-item strong { font-size:var(--font-sm); color:hsl(var(--foreground)); }
.search-item-meta { font-size:var(--font-xs); color:hsl(var(--muted-foreground)); }
.search-empty { padding:.5rem; text-align:center; font-size:var(--font-sm); color:hsl(var(--muted-foreground)); }

/* Contract info bar */
.contract-bar {
  display:flex; justify-content:space-between; align-items:center;
  padding:.6rem .75rem; border-radius:var(--radius); margin-bottom:.75rem;
  cursor:default;
}
.contract-bar.buy-bg { background:hsl(144 80% 55% / .12); }
.contract-bar.sell-bg { background:hsl(0 84% 60% / .12); }
.contract-symbol { font-size:var(--font-base); font-weight:700; color:hsl(var(--foreground)); }
.contract-exch { font-size:var(--font-xs); color:hsl(var(--muted-foreground)); }
.contract-right { text-align:right; }
.contract-ltp { font-size:var(--font-lg); font-weight:700; color:hsl(var(--foreground)); }
.contract-change { font-size:var(--font-xs); font-weight:500; }
.contract-bar.buy-bg .contract-change { color:#16A34A; }
.contract-bar.sell-bg .contract-change { color:hsl(var(--destructive)); }

/* Fields */
.field-grid { display:grid; grid-template-columns:1fr 1fr; gap:.6rem; margin-top:.5rem; }
.field-grid label { display:flex; flex-direction:column; gap:.2rem; font-size:var(--font-sm); color:hsl(var(--foreground)); }
.field-grid input, .field-grid select { padding:.45rem .6rem; border:1px solid hsl(var(--input)); border-radius:var(--radius); font-size:var(--font-sm); outline:none; background:hsl(var(--card)); }
.field-grid input:focus, .field-grid select:focus { border-color:hsl(var(--ring)); box-shadow:0 0 0 2px hsl(var(--ring)/.2); }
.field-grid input:disabled { background:hsl(var(--muted)); cursor:not-allowed; }
.price-wrap { position:relative; display:flex; align-items:center; }
.price-wrap input { flex:1; }
.price-link-badge {
  position:absolute; right:6px; font-size:.625rem; font-weight:600;
  color:var(--success-600); background:hsl(144 80% 55% / .15);
  padding:1px 5px; border-radius:3px; pointer-events:none;
}

.error-msg { padding:.5rem; color:hsl(var(--destructive)); font-size:var(--font-sm); text-align:center; }
.success-anim { display:flex; flex-direction:column; align-items:center; gap:.5rem; padding:2rem; }
.success-anim svg { animation: popIn .5s cubic-bezier(.68,-.55,.27,1.55); }
.success-text { font-size:var(--font-lg); font-weight:700; color:#16A34A; animation: fadeIn .5s ease; }
@keyframes popIn { 0% { transform:scale(0); opacity:0; } 100% { transform:scale(1); opacity:1; } }
@keyframes fadeIn { 0% { opacity:0; transform:translateY(10px); } 100% { opacity:1; transform:translateY(0); } }
/* Buy/Sell toggle */
.bs-row { display:flex; align-items:center; gap:.75rem; margin-bottom:.75rem; padding:.5rem 0; }
.bs-group { display:flex; border-radius:8px; overflow:hidden; border:2px solid hsl(var(--border)); }
.bs-btn {
  display:flex; align-items:center; gap:4px; padding:7px 18px; border:none;
  font-size:.8125rem; font-weight:600; cursor:pointer; background:transparent;
  color:hsl(var(--muted-foreground)); transition:all .15s;
}
.bs-btn:hover:not(.active) { background:hsl(var(--muted)); }
.bs-btn.active { color:#fff; }
.bs-btn.active:first-child { background:var(--success-600); border-color:var(--success-600); }
.bs-btn.active:last-child { background:var(--danger-600); border-color:var(--danger-600); }

.form-actions { display:flex; gap:.5rem; margin-top:.75rem; }
.order-btn { flex:1; padding:.6rem; border:none; border-radius:var(--radius); cursor:pointer; font-weight:600; color:#fff; font-size:.875rem; transition:opacity .15s; }
.order-btn.buy { background:var(--success-600); }
.order-btn.sell { background:var(--danger-600); }
.order-btn:disabled { opacity:.5; cursor:not-allowed; }
.order-btn:hover:not(:disabled) { opacity:.9; }
.form-actions .cancel { flex:1; padding:.6rem; border:none; border-radius:var(--radius); cursor:pointer; font-weight:500; background:hsl(var(--muted-foreground)); color:#fff; }
</style>
