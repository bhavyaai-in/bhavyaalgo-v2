<script setup>
import { ref, computed, onMounted, watch } from 'vue'
import { useRoute } from 'vue-router'
import { api } from '../../utils/api.js'
import OrdersTable from './OrdersTable.vue'

const props = defineProps({ data: null, broker: null })

const route = useRoute()
const brokers = ref([])
const selectedId = ref(route.query.broker || '')
const orders = ref([])
const loading = ref(false)
const error = ref('')
const displayOrders = computed(() => props.data || orders.value)
const editing = ref(null)
const editForm = ref({ price: '', quantity: '' })

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
  orders.value = []
  try {
    orders.value = await api(`/api/broker-orders/${selectedId.value}`)
  } catch (e) {
    error.value = e.message
  } finally {
    loading.value = false
  }
}

if (!props.data) {
  onMounted(async () => {
    await fetchBrokers()
    if (props.broker) {
      selectedId.value = String(props.broker.id)
    }
  })
  
  watch(selectedId, () => {
    if (selectedId.value) fetchData()
  })
}

async function cancelOrder(o) {
  if (!confirm('Cancel order ' + o.tradingsymbol + '?')) return
  const brokerId = props.broker?.id || selectedId.value
  if (!brokerId) { alert('No broker selected'); return }
  try {
    await api('/api/broker-cancel-order', {
      method: 'POST',
      body: JSON.stringify({ broker_id: Number(brokerId), order_id: o.orderid }),
    })
    if (!props.data) await fetchData()
  } catch (e) {
    alert('Cancel failed: ' + e.message)
  }
}

function openEdit(o) {
  editing.value = o
  editForm.value = { price: o.price || '', quantity: o.quantity || '' }
}

function closeEdit() {
  editing.value = null
}

async function saveEdit() {
  const o = editing.value
  if (!o) return
  const data = { ...o }
  data.price = Number(editForm.value.price) || o.price
  data.quantity = Number(editForm.value.quantity) || o.quantity
  if (data.variety === 'AMO') data.variety = 'NORMAL'
  delete data.orderid
  const brokerId = props.broker?.id || selectedId.value
  if (!brokerId) { alert('No broker selected'); return }
  try {
    await api('/api/broker-modify-order', {
      method: 'POST',
      body: JSON.stringify({ broker_id: Number(brokerId), order_id: o.orderid, data }),
    })
    closeEdit()
    if (!props.data) await fetchData()
  } catch (e) {
    alert('Modify failed: ' + e.message)
  }
}

function canAct(o) {
  const s = (o.orderstatus || '').toLowerCase()
  return s === 'open' || s === 'pending' || s === 'trigger pending'
}

function fmt(v) {
  if (v == null || v === '') return '-'
  return String(v)
}
</script>

<template>
  <div class="page">
    <header>
      <h2>Orders</h2>
      <select v-model="selectedId" class="broker-select">
        <option v-for="b in brokers" :key="b.id" :value="b.id">{{ b.friendly_name || b.broker_name }}</option>
      </select>
    </header>
    
    <div v-if="loading && props.broker" class="state-msg">Loading...</div>
    <div v-else-if="error && props.broker" class="state-msg error">{{ error }}</div>
    <div v-else-if="loading && !props.broker" class="state-msg">Loading...</div>
    <div v-else-if="error && !props.broker" class="state-msg error">{{ error }}</div>
    
    <OrdersTable :items="displayOrders" :canAct="canAct" @cancel="cancelOrder" @edit="openEdit" />
  </div>
  
  <div v-if="editing" class="modal-overlay" @click.self="closeEdit">
      <div class="edit-modal">
        <h3>Edit Order — {{ editing.tradingsymbol }}</h3>
        <label>Price <input v-model="editForm.price" type="number" step="0.05" /></label>
        <label>Quantity <input v-model="editForm.quantity" type="number" /></label>
        <div class="form-actions">
          <button @click="saveEdit">Save</button>
          <button class="cancel" @click="closeEdit">Cancel</button>
        </div>
      </div>
  </div>
</template>

<style scoped>
.status-badge {
  display: inline-block;
  padding: .15rem .5rem;
  border-radius: 999px;
  font-size: var(--font-xs);
  font-weight: 600;
}
.status-badge.open { background: hsl(48 100% 50% / .15); color: #b8860b; }
.status-badge.complete { background: hsl(144 80% 55% / .15); color: #16A34A; }
.status-badge.cancelled { background: hsl(0 84% 60% / .15); color: hsl(var(--destructive)); }
.actions { display: flex; gap: .3rem; }
.action-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 26px;
  height: 26px;
  border: 1px solid hsl(var(--border));
  border-radius: var(--radius);
  background: transparent;
  cursor: pointer;
  color: hsl(var(--muted-foreground));
}
.action-btn.cancel { border-color: hsl(var(--destructive)); color: hsl(var(--destructive)); }
.action-btn:hover { opacity: .7; }

.edit-modal {
  background: hsl(var(--card));
  padding: 2rem;
  border-radius: var(--radius);
  width: 90%;
  max-width: 360px;
  display: flex;
  flex-direction: column;
  gap: .8rem;
  box-shadow: 0 4px 24px rgba(0,0,0,.12);
}
.edit-modal h3 { margin: 0; font-size: var(--font-base); }
.edit-modal label {
  display: flex;
  flex-direction: column;
  gap: .3rem;
  font-size: var(--font-sm);
  color: hsl(var(--foreground));
}
.edit-modal input {
  padding: .5rem .7rem;
  border: 1px solid hsl(var(--input));
  border-radius: var(--radius);
  font-size: var(--font-sm);
}
.edit-modal input:focus {
  outline: none;
  border-color: hsl(var(--ring));
  box-shadow: 0 0 0 2px hsl(var(--ring) / .2);
}
.form-actions { display: flex; gap: .5rem; margin-top: .5rem; }
.form-actions button { flex: 1; padding: .6rem; border: none; border-radius: var(--radius); cursor: pointer; font-weight: 500; color: hsl(var(--primary-foreground)); background: hsl(var(--primary)); }
.form-actions .cancel { background: hsl(var(--muted-foreground)); }
</style>
