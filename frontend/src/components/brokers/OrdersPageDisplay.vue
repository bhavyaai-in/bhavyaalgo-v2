<script setup>
import { ref, computed, onMounted, onUnmounted, watch } from 'vue'
import { api, confirm } from '../../utils/api.js'
import { useBrokerDataStore } from '../../stores/brokerData.js'
import { useBrokerData } from '../../composables/useBrokerData.js'
import OrdersTable from './OrdersTable.vue'

const props = defineProps({ data: null, broker: null })
const { brokers, selectedId, data, loading, error } = useBrokerData(props, 'orders')

const editing = ref(null)
const editForm = ref({ price: '', quantity: '' })

// Dropdown State Logic
const isDropdownOpen = ref(false)
const dropdownRef = ref(null)

// 1. Bulletproof Watch: Jaise hi brokers list load hogi, yeh auto-select karega 1st broker
watch(() => brokers.value, (newBrokers) => {
  if (newBrokers && newBrokers.length > 0) {
    const exists = newBrokers.some(b => b.id === selectedId.value)
    if (!selectedId.value || !exists) {
      selectedId.value = newBrokers[0].id
    }
  }
}, { immediate: true, deep: true })

// 2. Updated Computed Property: UI text aur list state dono ko dynamic sync rakkhega
const selectedBrokerName = computed(() => {
  if (!brokers.value || brokers.value.length === 0) return 'No Broker Available'
  
  const current = brokers.value.find(b => b.id === selectedId.value)
  if (current) {
    return current.friendly_name || current.broker_name
  }
  
  return brokers.value[0].friendly_name || brokers.value[0].broker_name
})

function toggleDropdown() {
  isDropdownOpen.value = !isDropdownOpen.value
}

function selectBroker(id) {
  selectedId.value = id
  isDropdownOpen.value = false
}

// Bahar click karne par dropdown band karne ke liye
function handleClickOutside(event) {
  if (dropdownRef.value && !dropdownRef.value.contains(event.target)) {
    isDropdownOpen.value = false
  }
}

onMounted(() => { window.addEventListener('click', handleClickOutside) })
onUnmounted(() => { window.removeEventListener('click', handleClickOutside) })

function canAct(o) {
  const s = (o.orderstatus || '').toLowerCase()
  return s === 'open' || s === 'pending' || s === 'trigger pending'
}

function openEdit(o) {
  editing.value = o
  editForm.value = { price: o.price || '', quantity: o.quantity || '' }
}
function closeEdit() { editing.value = null }

async function saveEdit() {
  const o = editing.value
  if (!o) return
  const data2 = { ...o, price: Number(editForm.value.price) || o.price, quantity: Number(editForm.value.quantity) || o.quantity }
  if (data2.variety === 'AMO') data2.variety = 'NORMAL'
  delete data2.orderid
  const store = useBrokerDataStore()
  try {
    await api('/api/broker-modify-order', { method: 'POST', body: JSON.stringify({ broker_id: Number(selectedId.value), order_id: o.orderid, data: data2 }) })
    if (!props.data) await store.refreshOrders(selectedId.value)
    closeEdit()
  } catch (e) { alert('Modify failed: ' + e.message) }
}

async function cancelOrder(o) {
  if (!await confirm('Cancel Order', 'Cancel order ' + o.tradingsymbol + '?')) return
  const store = useBrokerDataStore()
  try {
    await api('/api/broker-cancel-order', { method: 'POST', body: JSON.stringify({ broker_id: Number(selectedId.value), order_id: o.orderid }) })
    if (!props.data) await store.refreshOrders(selectedId.value)
  } catch (e) { alert('Cancel failed: ' + e.message) }
}
</script>

<template>
  <div class="page">
    <header>
      <h2>Orders</h2>
      
      <div class="custom-select-container" ref="dropdownRef">
        <button class="broker-select-trigger" @click="toggleDropdown">
          <span>{{ selectedBrokerName }}</span>
          <span class="arrow-icon">▼</span>
        </button>

        <div v-if="isDropdownOpen" class="broker-options-dropdown">
          <div 
            v-for="b in brokers" 
            :key="b.id" 
            class="broker-option"
            :class="{ 'is-selected': selectedId === b.id }"
            @click="selectBroker(b.id)"
          >
            <span class="check-mark" v-if="selectedId === b.id">✓</span>
            <span class="option-text">{{ b.friendly_name || b.broker_name }}</span>
          </div>
        </div>
      </div>
    </header>

    <div v-if="loading" class="state-msg">Loading...</div>
    <div v-else-if="error" class="state-msg error">{{ error }}</div>
    <OrdersTable v-else-if="data && data.length" :items="data" :canAct="canAct" @cancel="cancelOrder" @edit="openEdit" />
    <div v-else class="state-msg">No orders.</div>
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
/* Dropdown Wrapper Styling */
.custom-select-container {
  position: relative;
  display: inline-block;
  user-select: none;
}

/* Trigger Button Layout */
.broker-select-trigger {
  box-sizing: border-box;
  font-family: var(--body-font);
  padding: .5rem 1rem;
  border: 1px solid hsl(var(--primary));
  border-radius: var(--radius);
  font-size: var(--font-sm);
  background: hsl(var(--primary));
  color: hsl(var(--primary-foreground));
  cursor: pointer;
  font-weight: 500;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 1rem;
  min-width: 140px;
}

/* Dropdown Menu List Box */
.broker-options-dropdown {
  position: absolute;
  top: calc(100% + 4px);
  left: 0;
  width: 100%;
  min-width: 160px;
  background-color: #ffffff;
  border: 1px solid var(--neutral-200);
  border-radius: var(--radius);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
  z-index: 50;
  overflow: hidden;
  display: flex;
  flex-direction: column;
}

/* Single Option Element */
.broker-option {
  padding: 0.6rem 0.8rem;
  font-size: var(--font-sm);
  color: var(--neutral-800);
  cursor: pointer;
  display: flex;
  align-items: center;
  gap: 0.4rem;
  transition: background 0.15s ease;
}

.broker-option:hover {
  background-color: var(--neutral-100);
}

/* Selected Option State Styling */
.broker-option.is-selected {
  background-color: hsl(var(--accent)) !important;
  color: hsl(var(--accent-foreground)) !important;
  font-weight: 500;
}

.check-mark {
  font-size: var(--font-xs);
  font-weight: bold;
}

.arrow-icon {
  font-size: 8px;
  opacity: 0.8;
}

/* --- Table & Modals CSS --- */
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
  display: inline-flex; align-items: center; justify-content: center;
  width: 26px; height: 26px;
  border: 1px solid hsl(var(--border)); border-radius: var(--radius);
  background: transparent; cursor: pointer; color: hsl(var(--muted-foreground));
}
.action-btn.cancel { border-color: hsl(var(--destructive)); color: hsl(var(--destructive)); }
.action-btn:hover { opacity: .7; }
.edit-modal {
  background: hsl(var(--card)); padding: 2rem; border-radius: var(--radius);
  width: 90%; max-width: 360px; display: flex; flex-direction: column; gap: .8rem;
  box-shadow: 0 4px 24px rgba(0,0,0,.12);
}
.edit-modal h3 { margin: 0; font-size: var(--font-base); }
.edit-modal label { display: flex; flex-direction: column; gap: .3rem; font-size: var(--font-sm); color: hsl(var(--foreground)); }
.edit-modal input { padding: .5rem .7rem; border: 1px solid hsl(var(--input)); border-radius: var(--radius); font-size: var(--font-sm); }
.edit-modal input:focus { outline: none; border-color: hsl(var(--ring)); box-shadow: 0 0 0 2px hsl(var(--ring) / .2); }
.form-actions { display: flex; gap: .5rem; margin-top: .5rem; }
.form-actions button { flex: 1; padding: .6rem; border: none; border-radius: var(--radius); cursor: pointer; font-weight: 500; color: hsl(var(--primary-foreground)); background: hsl(var(--primary)); }
.form-actions .cancel { background: hsl(var(--muted-foreground)); }
</style>