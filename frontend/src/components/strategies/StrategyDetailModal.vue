<script setup>
import { ref, computed, watch } from 'vue'
import { api, confirm } from '../../utils/api.js'
import PlaceOrderModal from '../../modals/brokers/PlaceOrderModal.vue'

const props = defineProps({ show: Boolean, data: Object, initialTab: { type: String, default: 'overview' }, typeName: Function, brokerName: Function, brokers: { type: Array, default: () => [] } })
const emit = defineEmits(['close'])

const activeTab = ref('overview')
const showAddJoiner = ref(false)
const useQty = ref(false)
const joinerForm = ref({ broker_id: 0, quantity: 1, multiplier: 1, re_entry: 1, is_active: true })
const submitting = ref(false)
const showPlaceOrder = ref(false)
const editingJoiner = ref(null)

const availableBrokers = computed(() => {
  const joined = new Set((props.data?.joiners || []).map(j => j.broker_id))
  if (editingJoiner.value) joined.delete(editingJoiner.value.broker_id)
  return (props.brokers || []).filter(b => !joined.has(b.id))
})

watch(() => props.show, (val) => {
  if (!val) { showAddJoiner.value = false; activeTab.value = 'overview' }
  if (val) { activeTab.value = props.initialTab; useQty.value = false; joinerForm.value = { broker_id: 0, quantity: 1, multiplier: 1, re_entry: 1, is_active: true } }
})

function cap(str) { if (!str) return ''; return str.replace(/\b\w/g, c => c.toUpperCase()) }
function fmt(v) { return v == null || v === '' ? '-' : Number(v).toFixed(2) }

function editJoiner(j) {
  editingJoiner.value = j
  useQty.value = j.quantity > 0
  joinerForm.value = {
    broker_id: j.broker_id,
    quantity: j.quantity || 1,
    multiplier: j.multiplier || 1,
    re_entry: j.re_entry || 1,
    is_active: j.is_active ? true : false,
  }
  showAddJoiner.value = true
}

async function saveJoiner() {
  if (!joinerForm.value.broker_id) return
  submitting.value = true
  try {
    const body = {
      broker_id: Number(joinerForm.value.broker_id),
      quantity: useQty.value ? Number(joinerForm.value.quantity) : 0,
      re_entry: Number(joinerForm.value.re_entry),
      re_entry_triggered: 0,
      multiplier: useQty.value ? 0 : Number(joinerForm.value.multiplier),
      is_active: joinerForm.value.is_active ? 1 : 0,
    }
    if (editingJoiner.value) {
      await api('/api/strategies/' + props.data.strategy.id + '/joiners/' + editingJoiner.value.id, {
        method: 'PUT', body: JSON.stringify(body),
      })
    } else {
      await api('/api/strategies/' + props.data.strategy.id + '/joiners', {
        method: 'POST', body: JSON.stringify(body),
      })
    }
    const joiners = await api('/api/strategies/' + props.data.strategy.id + '/joiners')
    props.data.joiners = joiners
    showAddJoiner.value = false
    editingJoiner.value = null
    useQty.value = false
    joinerForm.value = { broker_id: 0, quantity: 1, multiplier: 1, re_entry: 1, is_active: true }
  } catch (e) { alert(e.message) }
  finally { submitting.value = false }
}

async function removeJoiner(j) {
  if (!await confirm('Remove', 'Remove this broker from strategy?')) return
  await api('/api/strategies/' + props.data.strategy.id + '/joiners/' + j.id, { method: 'DELETE' })
  props.data.joiners = props.data.joiners.filter(x => x.id !== j.id)
}

async function toggleJoinerActive(j) {
  const newVal = j.is_active ? 0 : 1
  await api('/api/strategies/' + props.data.strategy.id + '/joiners/' + j.id, {
    method: 'PUT',
    body: JSON.stringify({
      broker_id: j.broker_id,
      quantity: j.quantity || 0,
      re_entry: j.re_entry || 1,
      re_entry_triggered: j.re_entry_triggered || 0,
      multiplier: j.multiplier || 0,
      is_active: newVal,
    }),
  })
  j.is_active = newVal
}

</script>

<template>
  <div v-if="show && data" class="modal-overlay" @click.self="emit('close')">
    <div class="modal-box">
      <h3>{{ data.strategy.name }}</h3>

      <!-- Sub-tabs -->
      <div class="sub-tabs">
        <button :class="{ active: activeTab === 'overview' }" @click="activeTab = 'overview'">Overview</button>
        <button :class="{ active: activeTab === 'joiners' }" @click="activeTab = 'joiners'">Joiners <span class="tab-badge">{{ (data.joiners || []).length }}</span></button>
        <button :class="{ active: activeTab === 'positions' }" @click="activeTab = 'positions'">Positions <span class="tab-badge">{{ (data.positions || []).length }}</span></button>
        <button :class="{ active: activeTab === 'orders' }" @click="activeTab = 'orders'">Orders <span class="tab-badge">{{ (data.orders || []).length }}</span></button>
      </div>

      <!-- Overview Tab -->
      <div v-if="activeTab === 'overview'" class="tab-panel">
        <div class="info-grid">
          <div class="info-item"><span>Type</span><strong>{{ typeName(data.strategy.strategy_type) }}</strong></div>
          <div class="info-item"><span>Side</span><strong>{{ cap(data.strategy.side) }}</strong></div>
          <div class="info-item"><span>Exchange</span><strong>{{ data.strategy.exchange || '-' }}</strong></div>
          <div class="info-item"><span>Expiry</span><strong>{{ data.strategy.expiry_date }}</strong></div>
          <div class="info-item"><span>Status</span><strong :style="{ color: data.strategy.is_active ? '#16A34A' : 'hsl(var(--muted-foreground))' }">{{ data.strategy.is_active ? 'Active' : 'Inactive' }}</strong></div>
          <div class="info-item"><span>Color</span><strong><span class="color-dot" :style="{ background: data.strategy.color }"></span> {{ data.strategy.color }}</strong></div>
        </div>
      </div>

      <!-- Joiners Tab -->
      <div v-if="activeTab === 'joiners'" class="tab-panel">
          <div class="section-bar">
            <span class="bar-title">{{ (data.joiners || []).length }} broker(s) joined</span>
            <button v-if="!showAddJoiner" class="small-btn" @click="showAddJoiner = true; editingJoiner = null; joinerForm = { broker_id: 0, quantity: 1, multiplier: 1, re_entry: 1, is_active: true }; useQty = false">+ Add Joiner</button>
          </div>

        <!-- Add Joiner Form -->
        <div v-if="showAddJoiner" class="joiner-form">
          <div class="jf-dual">
            <div class="jf-row">
              <label class="jf-label">Broker</label>
              <select v-model.number="joinerForm.broker_id">
                <option :value="0">-- Select Broker --</option>
                <option v-for="b in availableBrokers" :key="b.id" :value="b.id">{{ b.friendly_name || cap(b.broker_name) }}</option>
              </select>
            </div>
            <div class="jf-row">
              <label class="jf-label">Re-entry</label>
              <input v-model.number="joinerForm.re_entry" type="number" step="1" min="1" max="99" placeholder="Number of re-entries" />
            </div>
          </div>

          <div class="jf-dual">
            <div class="jf-row">
              <label class="jf-label">Mode</label>
              <div class="mode-toggle">
                <button class="mode-btn" :class="{ active: !useQty }" @click="useQty = false">Multiplier</button>
                <button class="mode-btn" :class="{ active: useQty }" @click="useQty = true">Quantity</button>
              </div>
            </div>
            <div class="jf-row">
              <label class="jf-label">{{ useQty ? 'Quantity' : 'Multiplier' }}</label>
              <input v-if="!useQty" v-model.number="joinerForm.multiplier" type="number" step="0.25" min="0.25" placeholder="e.g. 1.5" />
              <input v-else v-model.number="joinerForm.quantity" type="number" step="1" min="1" placeholder="e.g. 100" />
            </div>
          </div>

          <div class="jf-row">
            <label class="jf-label">Status</label>
            <label class="switch-label">
              <span class="switch">
                <input v-model="joinerForm.is_active" type="checkbox" />
                <span class="slider"></span>
              </span>
              {{ joinerForm.is_active ? 'Active' : 'Inactive' }}
            </label>
          </div>

          <div class="jf-actions">
            <button class="small-btn primary" :disabled="submitting || !joinerForm.broker_id" @click="saveJoiner">{{ submitting ? 'Saving...' : (editingJoiner ? 'Update' : 'Save') }}</button>
            <button class="small-btn" @click="showAddJoiner = false; editingJoiner = null">Cancel</button>
          </div>
        </div>

        <!-- Joiners List -->
        <div v-if="(data.joiners || []).length" class="joiner-list">
          <div v-for="j in data.joiners" :key="j.id" class="joiner-row">
            <div class="joiner-left">
              <label class="switch" @click.prevent="toggleJoinerActive(j)">
                <input type="checkbox" :checked="j.is_active" />
                <span class="slider"></span>
              </label>
              <div class="joiner-info">
                <strong>{{ brokerName(j.broker_id) }}</strong>
                <span class="joiner-meta">{{ j.quantity ? 'Qty: ' + j.quantity : 'Mult: ' + j.multiplier + 'x' }} · Re-entry: {{ j.re_entry }}</span>
              </div>
            </div>
            <div class="joiner-row-actions">
              <button class="chip" @click="editJoiner(j)">Edit</button>
              <button class="chip danger" @click="removeJoiner(j)">Remove</button>
            </div>
          </div>
        </div>
        <div v-else-if="!showAddJoiner" class="empty-msg">No brokers joined to this strategy.</div>
      </div>

      <!-- Positions Tab -->
      <div v-if="activeTab === 'positions'" class="tab-panel">
        <table v-if="(data.positions || []).length" class="data-table">
          <thead><tr><th>Symbol</th><th>Side</th><th>Qty</th><th>Buy Price</th><th>Sell Price</th><th>LTP</th></tr></thead>
          <tbody>
            <tr v-for="p in data.positions" :key="p.id">
              <td>{{ p.tradingsymbol }}</td>
              <td>{{ p.side }}</td>
              <td>{{ p.quantity }}</td>
              <td>{{ fmt(p.buy_price) }}</td>
              <td>{{ fmt(p.sell_price) }}</td>
              <td>{{ fmt(p.last_price) }}</td>
            </tr>
          </tbody>
        </table>
        <div v-else class="empty-msg">No positions.</div>
      </div>

      <!-- Orders Tab -->
      <div v-if="activeTab === 'orders'" class="tab-panel">
        <table v-if="(data.orders || []).length" class="data-table">
          <thead><tr><th>Order ID</th><th>Symbol</th><th>Type</th><th>Qty</th><th>Price</th><th>Status</th></tr></thead>
          <tbody>
            <tr v-for="o in data.orders" :key="o.id">
              <td>{{ o.order_id }}</td>
              <td>{{ o.tradingsymbol }}</td>
              <td>{{ o.transaction_type }}</td>
              <td>{{ o.quantity }}</td>
              <td>{{ fmt(o.price) }}</td>
              <td>{{ o.status }}</td>
            </tr>
          </tbody>
        </table>
        <div v-else class="empty-msg">No orders.</div>
      </div>

      <!-- Place Order button -->
      <div class="modal-footer">
        <button class="order-btn" @click="showPlaceOrder = true">Place Order</button>
        <button class="close-btn" @click="emit('close')">Close</button>
      </div>
    </div>

    <PlaceOrderModal
      :show="showPlaceOrder"
      :strategy="data.strategy"
      @close="showPlaceOrder = false"
    />
  </div>
</template>

<style scoped>
.modal-box { max-width:720px; }
h3 { margin-bottom:.75rem; }

/* Sub-tabs */
.sub-tabs { display:flex; gap:0; border-bottom:1px solid hsl(var(--border)); margin-bottom:.75rem; }
.sub-tabs button {
  padding:.45rem .9rem; border:none; background:transparent;
  font-size:var(--font-sm); color:hsl(var(--muted-foreground)); cursor:pointer;
  border-bottom:2px solid transparent; font-weight:500; display:flex; align-items:center; gap:4px;
  transition:color .15s;
}
.sub-tabs button.active { color:hsl(var(--primary)); border-bottom-color:hsl(var(--primary)); }
.sub-tabs button:hover { color:hsl(var(--foreground)); }
.tab-badge {
  font-size:.625rem; font-weight:600; padding:0 5px; border-radius:999px;
  background:hsl(var(--muted)); color:hsl(var(--muted-foreground)); line-height:1.4;
}
.sub-tabs button.active .tab-badge { background:hsl(var(--primary)/.15); color:hsl(var(--primary)); }

.tab-panel { min-height:100px; }

/* Overview */
.info-grid { display:grid; grid-template-columns:1fr 1fr; gap:.5rem; }
.info-item { display:flex; flex-direction:column; gap:2px; padding:.35rem .5rem; border-radius:var(--radius); background:hsl(var(--muted)/.3); }
.info-item span { font-size:var(--font-xs); color:hsl(var(--muted-foreground)); }
.info-item strong { font-size:var(--font-sm); color:hsl(var(--foreground)); }
.color-dot { display:inline-block; width:12px; height:12px; border-radius:999px; vertical-align:middle; margin-right:4px; }

/* Joiners */
.section-bar { display:flex; justify-content:space-between; align-items:center; margin-bottom:.6rem; }
.bar-title { font-size:var(--font-sm); color:hsl(var(--muted-foreground)); }
.joiner-form {
  background:hsl(var(--muted)/.2); border:1px solid hsl(var(--border));
  border-radius:var(--radius); padding:.75rem 1rem; margin-bottom:.75rem;
  display:flex; flex-direction:column; gap:.65rem;
}
.jf-row { display:flex; flex-direction:column; gap:3px; }
.jf-label { font-size:var(--font-xs); font-weight:600; color:hsl(var(--foreground)); }
.jf-dual { display:flex; flex-direction:column; gap:.65rem; }
@media (min-width:500px) {
  .jf-dual { flex-direction:row; gap:.75rem; }
  .jf-dual .jf-row { flex:1; }
}
.jf-row select, .jf-row input {
  padding:.45rem .55rem; border:1px solid hsl(var(--input)); border-radius:var(--radius);
  font-size:var(--font-sm); outline:none; background:hsl(var(--card)); width:100%;
}
.jf-row select:focus, .jf-row input:focus { border-color:hsl(var(--ring)); box-shadow:0 0 0 2px hsl(var(--ring)/.2); }
.mode-toggle { display:flex; border:1px solid hsl(var(--input)); border-radius:var(--radius); overflow:hidden; }
.mode-btn {
  flex:1; padding:.4rem .5rem; border:none; font-size:var(--font-xs); font-weight:600;
  cursor:pointer; background:transparent; color:hsl(var(--muted-foreground));
}
.mode-btn.active { background:hsl(var(--primary)); color:#fff; }
.mode-btn:not(.active):hover { background:hsl(var(--muted)); }
.switch-label { display:flex; align-items:center; gap:6px; font-size:var(--font-sm); color:hsl(var(--muted-foreground)); cursor:pointer; }
.switch { position:relative; display:inline-block; width:34px; height:20px; flex-shrink:0; }
.switch input { opacity:0; width:0; height:0; }
.slider {
  position:absolute; inset:0; background:hsl(var(--muted-foreground)); border-radius:999px;
  transition:background .2s; cursor:pointer;
}
.slider::before {
  content:''; position:absolute; left:3px; top:3px; width:14px; height:14px;
  background:#fff; border-radius:999px; transition:transform .2s;
}
.switch input:checked + .slider { background:hsl(var(--primary)); }
.switch input:checked + .slider::before { transform:translateX(14px); }
.jf-actions { display:flex; gap:.5rem; margin-top:.25rem; }
.jf-actions .small-btn { flex:1; text-align:center; }
.joiner-list { display:flex; flex-direction:column; gap:.4rem; }
.joiner-row {
  display:flex; justify-content:space-between; align-items:center;
  padding:.5rem .65rem; border:1px solid hsl(var(--border)); border-radius:var(--radius);
}
.joiner-left { display:flex; align-items:center; gap:.6rem; }
.joiner-row-actions { display:flex; gap:.35rem; flex-shrink:0; }
.joiner-info { display:flex; flex-direction:column; gap:2px; }
.joiner-info strong { font-size:var(--font-sm); color:hsl(var(--foreground)); }
.joiner-meta { font-size:var(--font-xs); color:hsl(var(--muted-foreground)); }
.joiner-row .switch { position:relative; display:inline-block; width:34px; height:20px; flex-shrink:0; cursor:pointer; }
.joiner-row .switch input { opacity:0; width:0; height:0; }
.joiner-row .switch .slider {
  position:absolute; inset:0; background:hsl(var(--muted-foreground)); border-radius:999px;
  transition:background .2s; cursor:pointer;
}
.joiner-row .switch .slider::before {
  content:''; position:absolute; left:3px; top:3px; width:14px; height:14px;
  background:#fff; border-radius:999px; transition:transform .2s;
}
.joiner-row .switch input:checked + .slider { background:#16A34A; }
.joiner-row .switch input:checked + .slider::before { transform:translateX(14px); }
.chip {
  font-size:var(--font-xs); padding:.25rem .55rem; border:1px solid hsl(var(--border));
  border-radius:var(--radius); cursor:pointer; background:transparent; color:hsl(var(--muted-foreground));
}
.chip:hover { border-color:hsl(var(--primary)); color:hsl(var(--primary)); }
.chip.danger:hover { border-color:hsl(var(--destructive)); color:hsl(var(--destructive)); }

/* Data table */
.data-table { width:100%; border-collapse:collapse; font-size:var(--font-sm); }
.data-table th, .data-table td { padding:.35rem .45rem; border-bottom:1px solid hsl(var(--border)); text-align:left; white-space:nowrap; }
.data-table th { font-weight:600; color:hsl(var(--foreground)); }
.data-table td { color:hsl(var(--muted-foreground)); }

.empty-msg { text-align:center; padding:1.5rem; font-size:var(--font-sm); color:hsl(var(--muted-foreground)); }
.small-btn {
  padding:.35rem .65rem; border:1px solid hsl(var(--border)); border-radius:var(--radius);
  font-size:var(--font-xs); cursor:pointer; background:transparent; color:hsl(var(--muted-foreground));
}
.small-btn:hover { border-color:hsl(var(--primary)); color:hsl(var(--primary)); }
.small-btn.primary { background:hsl(var(--primary)); color:#fff; border-color:hsl(var(--primary)); }
.small-btn.primary:hover { opacity:.9; }
.small-btn:disabled { opacity:.5; cursor:default; }
.close-btn { padding:.5rem 1.5rem; border:1px solid hsl(var(--border)); background:hsl(var(--card)); border-radius:var(--radius); cursor:pointer; font-weight:500; }
.modal-footer { display:flex; gap:.5rem; justify-content:center; margin-top:1rem; }
.order-btn {
  padding:.5rem 1.5rem; border:none; border-radius:var(--radius); cursor:pointer;
  font-weight:600; color:#fff; background:hsl(var(--primary));
}
.order-btn:hover { opacity:.9; }
</style>
