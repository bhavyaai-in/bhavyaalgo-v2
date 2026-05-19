<script setup>
import { ref, computed, watch } from 'vue'
import { api } from '../utils/api.js'
import { useWebSocket } from '../composables/useWebSocket.js'

const props = defineProps({
  symbol: String,
  token: String,
  exchange: String,
  show: Boolean,
})
const emit = defineEmits(['close'])

const ws = useWebSocket()

const transactionType = ref('BUY')
const orderTab = ref('regular')
const productType = ref('DELIVERY')
const quantity = ref(1)
const price = ref(0)
const triggerPrice = ref(0)
const orderType = ref('MARKET')
const submitting = ref(false)
const error = ref('')
const success = ref(false)
const connectedBroker = ref(null)
const ltpMap = ref({})
const availableMargin = ref(0)

ws.onTick((tick) => {
  if (tick.token && tick.ltp != null) {
    ltpMap.value = { ...ltpMap.value, [tick.token]: tick }
  }
})

const tick = computed(() => props.token ? ltpMap.value[props.token] : null)
const ltp = computed(() => tick.value ? Number(tick.value.ltp || tick.value.close || 0).toFixed(2) : '-')

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

const tradingSymbol = computed(() => {
  if (!props.symbol) return ''
  return props.exchange === 'NSE' ? `${props.symbol}-EQ` : props.symbol
})

watch(() => props.show, async (val) => {
  if (!val) return
  transactionType.value = 'BUY'
  orderTab.value = 'regular'
  productType.value = 'DELIVERY'
  quantity.value = 1
  price.value = 0
  triggerPrice.value = 0
  orderType.value = 'MARKET'
  submitting.value = false
  error.value = ''
  success.value = false

  if (props.token && props.exchange) {
    ws.subscribe([`${props.exchange}|${props.token}`])
  }

  try {
    const brokers = await api('/api/brokers')
    connectedBroker.value = brokers.find(b => b.token_status === 'connected') || null
  } catch {}

  if (connectedBroker.value) {
    try {
      const margin = await api(`/api/broker-margin/${connectedBroker.value.id}`)
      availableMargin.value = margin?.availablecash || 0
    } catch {}
  }
})

async function placeOrder() {
  if (!connectedBroker.value) {
    error.value = 'No connected broker found'
    return
  }
  submitting.value = true
  error.value = ''

  const variety = orderTab.value === 'stoploss' ? 'STOPLOSS' : 'NORMAL'
  const priceVal = orderType.value === 'MARKET' ? '0' : String(price.value)
  const triggerVal = orderTab.value === 'stoploss' ? String(triggerPrice.value) : '0'

  const data = {
    variety,
    tradingsymbol: tradingSymbol.value,
    symboltoken: props.token,
    transactiontype: transactionType.value,
    exchange: props.exchange,
    ordertype: orderType.value,
    producttype: productType.value,
    duration: 'DAY',
    price: priceVal,
    triggerprice: triggerVal,
    squareoff: '0',
    stoploss: '0',
    quantity: String(quantity.value),
  }

  try {
    await api('/api/broker-place-order', {
      method: 'POST',
      body: JSON.stringify({ broker_id: connectedBroker.value.id, data }),
    })
    success.value = true
    setTimeout(() => { success.value = false; emit('close') }, 2000)
  } catch (e) {
    console.log('Order error:', e)
    error.value = e.message || 'Order could not be placed'
  } finally {
    submitting.value = false
  }
}
</script>

<template>
  <div v-if="show" class="orderpad-overlay" @click.self="emit('close')">
    <div class="orderpad">
      <!-- Header -->
      <div class="orderpad-header" :class="transactionType === 'BUY' ? 'bg-up' : 'bg-down'">
        <div>
          <h5 class="header-symbol">{{ symbol }}</h5>
          <div class="header-meta">
            <span class="exch-badge">{{ exchange }}</span>
            <span class="header-ltp">{{ ltp }}</span>
            <span v-if="changeVal" class="header-change">{{ changeVal }} ({{ pctVal }})</span>
          </div>
        </div>
        <div class="header-right">
          <div class="bs-group">
            <button class="bs-btn" :class="{ active: transactionType === 'BUY' }" @click="transactionType = 'BUY'">B</button>
            <button class="bs-btn" :class="{ active: transactionType === 'SELL' }" @click="transactionType = 'SELL'">S</button>
          </div>
          <button class="close-icon" @click="emit('close')">✕</button>
        </div>
      </div>

      <!-- Success -->
      <div v-if="success" class="success-section">
        <svg width="48" height="48" viewBox="0 0 24 24" fill="none" stroke="var(--success-600)" stroke-width="2"><path d="M22 11.08V12a10 10 0 1 1-5.93-9.14"/><polyline points="22 4 12 14.01 9 11.01"/></svg>
        <div class="success-text">Order Placed!</div>
      </div>

      <!-- Form -->
      <div v-else class="orderpad-body">
        <!-- Order type tabs -->
        <div class="tabs">
          <button class="tab" :class="{ active: orderTab === 'regular' }" @click="orderTab = 'regular'">Regular</button>
          <button class="tab" :class="{ active: orderTab === 'stoploss' }" @click="orderTab = 'stoploss'">Stop Loss</button>
          <button class="tab" :class="{ active: orderTab === 'gtt' }" @click="orderTab = 'gtt'">GTT</button>
          <button class="tab" :class="{ active: orderTab === 'sip' }" @click="orderTab = 'sip'">SIP</button>
        </div>

        <!-- Product + Quantity + Price in one row -->
        <div class="fields-row">
          <div class="field-block">
            <label class="field-label">Product Type</label>
            <div class="segmented">
              <button class="seg-btn" :class="{ active: productType === 'INTRADAY' }" @click="productType = 'INTRADAY'">INT</button>
              <button class="seg-btn" :class="{ active: productType === 'DELIVERY' }" @click="productType = 'DELIVERY'">DEL</button>
            </div>
          </div>
          <div class="field-block">
            <label class="field-label">Quantity</label>
            <div class="qty-wrap">
              <input type="number" v-model.number="quantity" min="1" class="field-input" />
              <div class="qty-arrows">
                <button @click="quantity = Math.max(1, quantity + 1)">▲</button>
                <button @click="quantity = Math.max(1, quantity - 1)">▼</button>
              </div>
            </div>
          </div>
          <div class="field-block price-block">
            <label class="field-label">Price</label>
            <div class="price-inline">
              <input type="number" v-model.number="price" :disabled="orderType === 'MARKET'" step="0.05" min="0" class="field-input" />
              <div class="segmented">
                <button class="seg-btn" :class="{ active: orderType === 'LIMIT' }" @click="orderType = 'LIMIT'">Limit</button>
                <button class="seg-btn" :class="{ active: orderType === 'MARKET' }" @click="orderType = 'MARKET'">Market</button>
              </div>
            </div>
          </div>
        </div>

        <!-- Trigger Price (Stop Loss) -->
        <div v-if="orderTab === 'stoploss'" class="trigger-row">
          <label class="field-label">Trigger Price</label>
          <input type="number" v-model.number="triggerPrice" step="0.05" min="0" class="field-input" />
        </div>

        <div v-if="error" class="error-msg">{{ error }}</div>

        <!-- Bottom -->
        <div class="bottom-row">
          <div>
            <div class="avail-label">Available</div>
            <div class="avail-amount">₹ {{ Number(availableMargin).toFixed(2) }}</div>
          </div>
          <button class="order-btn" :class="transactionType === 'BUY' ? 'buy' : 'sell'" :disabled="submitting" @click="placeOrder">
            {{ submitting ? 'Placing...' : transactionType }}
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.orderpad-overlay {
  position: fixed; top: 0; right: 0; bottom: 0;
  width: 540px; z-index: 100;
  background: hsl(var(--card));
  border-left: 1px solid hsl(var(--border));
  display: flex; flex-direction: column;
  box-shadow: -4px 0 24px rgba(0,0,0,.08);
}

.orderpad { display: flex; flex-direction: column; height: 100%; }

/* Header */
.orderpad-header {
  padding: 12px 16px; display: flex; justify-content: space-between; align-items: flex-start;
}
.bg-up { background: var(--success-600); }
.bg-down { background: var(--danger-600); }

.header-symbol { font-size: 1rem; font-weight: 600; color: #fff; margin: 0 0 4px; }
.header-meta { display: flex; align-items: center; gap: 10px; font-size: .8125rem; color: rgba(255,255,255,.9); }
.exch-badge {
  font-size: .6875rem; font-weight: 500; padding: 1px 6px; border-radius: 3px;
  background: rgba(255,255,255,.2); color: #fff;
}
.header-ltp { font-weight: 600; font-size: .875rem; color: #fff; }
.header-change { font-size: .75rem; color: rgba(255,255,255,.9); }

.header-right { display: flex; align-items: center; gap: 8px; }
.bs-group { display: flex; border-radius: 6px; overflow: hidden; border: 1.5px solid rgba(255,255,255,.35); }
.bs-btn {
  width: 30px; height: 28px; border: none; font-size: .8125rem; font-weight: 700;
  cursor: pointer; background: transparent; color: rgba(255,255,255,.7);
  transition: all .15s;
}
.bs-btn.active:first-child { background: #fff; color: var(--success-600); }
.bs-btn.active:last-child { background: #fff; color: var(--danger-600); }
.bs-btn:hover:not(.active) { background: rgba(255,255,255,.15); color: #fff; }

.close-icon {
  width: 26px; height: 26px; border: none; border-radius: 50%;
  background: rgba(255,255,255,.2); color: #fff; font-size: .8125rem;
  cursor: pointer; display: flex; align-items: center; justify-content: center;
}
.close-icon:hover { background: rgba(255,255,255,.35); }

/* Body */
.orderpad-body { flex: 1; padding: 16px; display: flex; flex-direction: column; gap: 16px; }

.success-section {
  flex: 1; display: flex; flex-direction: column; align-items: center; justify-content: center; gap: 8px;
}
.success-text { font-size: 1.125rem; font-weight: 700; color: var(--success-600); }

/* Tabs */
.tabs { display: flex; border-bottom: 1px solid hsl(var(--border)); }
.tab {
  padding: 8px 16px; border: none; background: transparent;
  font-size: .8125rem; color: hsl(var(--muted-foreground)); cursor: pointer;
  border-bottom: 2px solid transparent; font-weight: 500; transition: color .15s;
}
.tab.active { color: hsl(var(--primary)); border-bottom-color: hsl(var(--primary)); }
.tab:hover { color: hsl(var(--foreground)); }

/* Fields row */
.fields-row { display: flex; gap: 12px; align-items: flex-start; }
.field-block { flex: 1; min-width: 0; }
.price-block { flex: 1.5; }
.field-label { display: block; font-size: .75rem; color: hsl(var(--muted-foreground)); margin-bottom: 6px; font-weight: 500; white-space: nowrap; }

.segmented { display: inline-flex; border-radius: 6px; overflow: hidden; border: 1px solid hsl(var(--border)); }
.seg-btn {
  padding: 6px 14px; border: none; background: transparent;
  font-size: .75rem; font-weight: 500; cursor: pointer;
  color: hsl(var(--muted-foreground)); transition: all .15s;
}
.seg-btn.active { background: hsl(var(--primary)); color: #fff; }
.seg-btn:hover:not(.active) { background: hsl(var(--muted)); }

.field-input {
  width: 100%; padding: 8px 10px; border: 1px solid hsl(var(--input));
  border-radius: 6px; font-size: .875rem; outline: none; background: hsl(var(--card));
}
.field-input:focus { border-color: hsl(var(--ring)); box-shadow: 0 0 0 2px hsl(var(--ring)/.15); }
.field-input:disabled { background: hsl(var(--muted)); cursor: not-allowed; }

.qty-wrap { position: relative; }
.qty-arrows { position: absolute; right: 4px; top: 4px; display: flex; flex-direction: column; }
.qty-arrows button {
  display: flex; align-items: center; justify-content: center;
  width: 20px; height: 14px; border: none; background: transparent;
  font-size: .4375rem; cursor: pointer; color: hsl(var(--muted-foreground)); padding: 0;
}
.qty-arrows button:hover { color: hsl(var(--foreground)); }

.price-inline { display: flex; align-items: center; gap: 6px; }
.price-inline .field-input { flex: 1; min-width: 0; }

.error-msg { padding: 8px; color: hsl(var(--destructive)); font-size: .8125rem; text-align: center; }

.bottom-row { margin-top: auto; display: flex; align-items: center; justify-content: space-between; padding: 16px 0 0; border-top: 1px solid hsl(var(--border)); }
.avail-label { font-size: .6875rem; color: hsl(var(--muted-foreground)); margin-bottom: 2px; }
.avail-amount { font-size: .875rem; font-weight: 600; color: hsl(var(--foreground)); }

.order-btn {
  padding: 10px 36px; border: none; border-radius: 8px;
  font-size: .875rem; font-weight: 600; cursor: pointer; color: #fff;
  transition: opacity .15s;
}
.order-btn.buy { background: var(--success-600); }
.order-btn.sell { background: var(--danger-600); }
.order-btn:disabled { opacity: .5; cursor: not-allowed; }
.order-btn:hover:not(:disabled) { opacity: .9; }
</style>
