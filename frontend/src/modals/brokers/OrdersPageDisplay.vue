<script setup>
import { ref, onMounted, watch } from 'vue'
import { useRoute } from 'vue-router'

const route = useRoute()
const brokers = ref([])
const selectedId = ref(route.query.broker || '')
const orders = ref([])
const loading = ref(false)
const error = ref('')
const editing = ref(null)
const editForm = ref({ price: '', quantity: '' })

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
  orders.value = []
  try {
    orders.value = await api(`/api/broker-orders/${selectedId.value}`)
  } catch (e) {
    error.value = e.message
  } finally {
    loading.value = false
  }
}

watch(selectedId, () => {
  if (brokers.value.length) fetchData()
})
onMounted(fetchBrokers)

async function cancelOrder(o) {
  if (!confirm('Cancel order ' + o.tradingsymbol + '?')) return
  try {
    await api('/api/broker-cancel-order', {
      method: 'POST',
      body: JSON.stringify({ broker_id: Number(selectedId.value), order_id: o.orderid }),
    })
    await fetchData()
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
  try {
    await api('/api/broker-modify-order', {
      method: 'POST',
      body: JSON.stringify({ broker_id: Number(selectedId.value), order_id: o.orderid, data }),
    })
    closeEdit()
    await fetchData()
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
  <div class="orders-page">
    <header>
      <h2>Orders</h2>
      <select v-model="selectedId" class="broker-select">
        <option v-for="b in brokers" :key="b.id" :value="b.id">
          {{ b.friendly_name || b.broker_name }}
        </option>
      </select>
    </header>

    <div v-if="loading" class="state-msg">Loading...</div>
    <div v-else-if="error" class="state-msg error">{{ error }}</div>
    <div v-else-if="orders.length === 0" class="state-msg">No orders.</div>
    <div v-else class="table-wrap">
      <table class="data-table">
        <thead><tr><th>Trading Symbol</th><th>Exchange</th><th>Type</th><th>Qty</th><th>Filled</th><th>Price</th><th>Order Type</th><th>Product</th><th>Status</th><th>Actions</th></tr></thead>
        <tbody>
          <tr v-for="o in orders" :key="o.orderid">
            <td><strong>{{ o.tradingsymbol }}</strong></td>
            <td>{{ o.exchange }}</td>
            <td>{{ o.transactiontype }}</td>
            <td>{{ o.quantity }}</td>
            <td>{{ o.filledshares }}</td>
            <td>{{ o.price }}</td>
            <td>{{ o.ordertype }}</td>
            <td>{{ o.producttype }}</td>
            <td><span class="status-badge" :class="o.orderstatus">{{ o.orderstatus }}</span></td>
            <td>
              <div class="actions" v-if="canAct(o)">
                <button class="action-btn" title="Edit" @click="openEdit(o)">
                  <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M11 4H4a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2v-7"/><path d="M18.5 2.5a2.121 2.121 0 0 1 3 3L12 15l-4 1 1-4 9.5-9.5z"/></svg>
                </button>
                <button class="action-btn cancel" title="Cancel" @click="cancelOrder(o)">
                  <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/></svg>
                </button>
              </div>
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <!-- Edit Modal -->
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
  </div>
</template>

<style scoped>
.orders-page { padding: 1rem 0; }
header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 1.25rem;
}
header h2 { margin: 0; }
.broker-select {
  padding: .5rem 1rem;
  border: 1px solid hsl(var(--primary));
  border-radius: var(--radius);
  font-size: var(--font-sm);
  background: hsl(var(--primary));
  color: hsl(var(--primary-foreground));
  cursor: pointer;
  font-weight: 500;
}
.state-msg { text-align: center; padding: 2rem; color: hsl(var(--muted-foreground)); }
.state-msg.error { color: hsl(var(--destructive)); }
.table-wrap { overflow-x: auto; }
.data-table { width: 100%; border-collapse: collapse; font-size: var(--font-sm); }
.data-table th, .data-table td {
  padding: .5rem .6rem;
  border-bottom: 1px solid hsl(var(--border));
  text-align: left;
  white-space: nowrap;
}
.data-table th {
  font-weight: 600;
  color: hsl(var(--foreground));
  position: sticky;
  top: 0;
  background: hsl(var(--card));
}
.data-table td { color: hsl(var(--muted-foreground)); }
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

.modal-overlay {
  position: fixed;
  inset: 0;
  background: rgba(0,0,0,.4);
  display: flex;
  justify-content: center;
  align-items: center;
  z-index: 100;
}
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
.form-actions {
  display: flex;
  gap: .5rem;
  margin-top: .5rem;
}
.form-actions button {
  flex: 1;
  padding: .6rem;
  border: none;
  border-radius: var(--radius);
  cursor: pointer;
  font-weight: 500;
  color: hsl(var(--primary-foreground));
  background: hsl(var(--primary));
}
.form-actions .cancel { background: hsl(var(--muted-foreground)); }
</style>
