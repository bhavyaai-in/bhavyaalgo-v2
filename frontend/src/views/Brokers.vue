<script setup>
import { ref, onMounted } from 'vue'
import { api, confirm } from '../utils/api.js'
import { useNotificationStore } from '../stores/notification.js'
import BrokerFormModal from '../modals/brokers/BrokerFormModal.vue'
import BrokerModal from '../modals/brokers/BrokerModal.vue'
import PlaceOrderModal from '../modals/brokers/PlaceOrderModal.vue'
import ProfileModal from '../modals/brokers/ProfileModal.vue'
import OrdersPageDisplay from '../components/brokers/OrdersPageDisplay.vue'
import HoldingsPageDisplay from '../components/brokers/HoldingsPageDisplay.vue'
import PositionsPageDisplay from '../components/brokers/PositionsPageDisplay.vue'
import MarginPageDisplay from '../components/brokers/MarginPageDisplay.vue'


const notif = useNotificationStore()

const brokers = ref([])

function cap(str) {
  if (!str) return ''
  return str.replace(/\b\w/g, c => c.toUpperCase())
}
const brokerList = ref([])
const showForm = ref(false)
const editing = ref(null)

const form = ref({
  friendly_name: '',
  broker_userid: '',
  broker_password: '',
  broker_pin: '',
  broker_qr_key: '',
  broker_api: '',
  broker_api_secret: '',
  broker_name: '',
  is_active: false,
  is_autologin: false,
})

async function fetchBrokers() {
  brokers.value = await api('/api/brokers')
}

async function fetchBrokerList() {
  const list = await api('/api/broker-list')
  brokerList.value = list.filter(e => e.is_active)
}

function openAdd() {
  editing.value = null
  form.value = {
    friendly_name: '',
    broker_userid: '',
    broker_password: '',
    broker_pin: '',
    broker_qr_key: '',
    broker_api: '',
    broker_api_secret: '',
    broker_name: '',
    is_active: false,
    is_autologin: false,
  }
  showForm.value = true
}

function openEdit(b) {
  editing.value = b.id
  form.value = { ...b }
  showForm.value = true
}

async function save() {
  try {
    if (editing.value) {
      await api(`/api/brokers/${editing.value}`, {
        method: 'PUT',
        body: JSON.stringify(form.value),
      })
    } else {
      await api('/api/brokers', {
        method: 'POST',
        body: JSON.stringify(form.value),
      })
    }
    showForm.value = false
    await fetchBrokers()
  } catch (e) {
    alert('Error: ' + e.message)
  }
}

const connecting = ref(null)
const activeModal = ref(null) // 'orders' | 'holdings' | 'positions' | 'margin' | null
const modalBroker = ref(null)

function openModal(type, b) {
  activeModal.value = type
  modalBroker.value = b
}
function closeModal() {
  activeModal.value = null
  modalBroker.value = null
}

  async function connectBroker(b) {
  connecting.value = b.id
  try {
    const res = await api('/api/connect-broker', {
      method: 'POST',
      body: JSON.stringify({ broker_id: b.id }),
      skipLoader: true,
    })
    b.token_status = 'connected'
    b.broker_token = res.auth_token
    b.feed_token = res.feed_token || ''
    b.message = res.profile_name
    notif.add({
      title: 'Connected',
      message: (b.friendly_name || cap(b.broker_name)) + ' (' + cap(b.broker_name) + ') (' + (res.profile_name || b.broker_userid) + ')',
      type: 'success',
    })
  } catch (e) {
    await fetchBrokers()
  } finally {
    connecting.value = null
  }
}

async function remove(id) {
  if (!await confirm('Delete Broker', 'Delete this broker?')) return
  try {
    await api(`/api/brokers/${id}`, { method: 'DELETE' })
    await fetchBrokers()
  } catch (e) {
    alert('Error: ' + e.message)
  }
}

onMounted(async () => {
  await fetchBrokerList()
  await fetchBrokers()
})
</script>

<template>
  <div class="brokers-page">
    <header>
      <h2>Brokers</h2>
      <button class="add-btn" @click="openAdd">+ Add Broker</button>
    </header>

    <BrokerFormModal
      :show="showForm"
      :editing="editing"
      :form="form"
      :broker-list="brokerList"
      @close="showForm = false"
      @save="save"
    />

    <div v-if="brokers.length === 0" class="empty">
      <p>No brokers yet.</p>
    </div>
    <div v-else class="broker-grid">
      <div v-for="b in brokers" :key="b.id" class="broker-card">
        <div class="card-header">
          <div class="card-title-row">
            <strong>{{ cap(b.friendly_name || b.broker_name) }}</strong>
            <a v-if="b.broker_name === 'aliceblue'" :href="'https://ant.aliceblueonline.com/?appcode=' + b.broker_api" target="_blank" class="info-icon" title="Verify app on Alice Blue">&nbsp;i</a>
          </div>
          <span class="badge" :class="{ connected: b.token_status === 'connected', error: b.token_status === 'error', expired: b.token_status === 'expired' }">
            <template v-if="b.token_status === 'connected'">{{ b.message }}</template>
            <template v-else-if="b.token_status === 'error'">Error</template>
            <template v-else-if="b.token_status === 'expired'">Expired</template>
            <template v-else>Not Connected</template>
          </span>
        </div>
        <div class="card-body">
          <p>User: {{ b.broker_userid }} ({{ cap(b.broker_name) }})</p>
        </div>
        <div class="chip-grid">
          <button class="chip connect-chip"
            :class="{ connected: b.token_status === 'connected', loading: connecting === b.id }"
            :disabled="connecting === b.id"
            :title="b.token_status === 'connected' ? 'Connected (' + b.message + ')' : 'Connect'"
            @click="connectBroker(b)">
            <template v-if="connecting === b.id">
              <svg class="spin" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><circle cx="12" cy="12" r="10" opacity=".3"/><path d="M12 2a10 10 0 0 1 10 10" stroke-linecap="round"/></svg>
              <span>Connecting...</span>
            </template>
            <template v-else>
              <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M10 13a5 5 0 0 0 7.54.54l3-3a5 5 0 0 0-7.07-7.07l-1.72 1.71"/><path d="M14 11a5 5 0 0 0-7.54-.54l-3 3a5 5 0 0 0 7.07 7.07l1.71-1.71"/></svg>
              <span>{{ b.token_status === 'connected' ? 'Connected' : 'Connect' }}</span>
            </template>
          </button>
          <button class="chip" :disabled="b.token_status !== 'connected'" @click="openModal('profile', b)">
            <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M20 21v-2a4 4 0 0 0-4-4H8a4 4 0 0 0-4 4v2"/><circle cx="12" cy="7" r="4"/></svg>
            <span>Profile</span>
          </button>
          <button class="chip" :disabled="b.token_status !== 'connected'" @click="openModal('orders', b)">
            <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"/><polyline points="14 2 14 8 20 8"/><line x1="16" y1="13" x2="8" y2="13"/><line x1="16" y1="17" x2="8" y2="17"/><polyline points="10 9 9 9 8 9"/></svg>
            <span>Orders</span>
          </button>
          <button class="chip" :disabled="b.token_status !== 'connected'" @click="openModal('positions', b)">
            <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><line x1="18" y1="20" x2="18" y2="10"/><line x1="12" y1="20" x2="12" y2="4"/><line x1="6" y1="20" x2="6" y2="14"/></svg>
            <span>Positions</span>
          </button>
          <button class="chip" :disabled="b.token_status !== 'connected'" @click="openModal('holdings', b)">
            <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><rect x="3" y="11" width="18" height="11" rx="2" ry="2"/><path d="M7 11V7a5 5 0 0 1 10 0v4"/></svg>
            <span>Holdings</span>
          </button>
          <button class="chip" :disabled="b.token_status !== 'connected'" @click="openModal('margin', b)">
            <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="10"/><path d="M16 8h-6a2 2 0 1 0 0 4h4a2 2 0 1 1 0 4H8"/><path d="M12 18V6"/></svg>
            <span>Margin</span>
          </button>
          <button class="chip" :disabled="b.token_status !== 'connected'" @click="openModal('place-order', b)">
            <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M12 5l-11 7h22l-11-7z"/><polyline points="12 12 12 22"/><line x1="5" y1="14" x2="5" y2="22"/><line x1="19" y1="14" x2="19" y2="22"/></svg>
            <span>Place Order</span>
          </button>
          <button class="chip" @click="openEdit(b)">
            <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M11 4H4a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2v-7"/><path d="M18.5 2.5a2.121 2.121 0 0 1 3 3L12 15l-4 1 1-4 9.5-9.5z"/></svg>
            <span>Edit</span>
          </button>
          <button class="chip danger" :disabled="b.token_status !== 'connected'" @click="remove(b.id)">
            <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><polyline points="3 6 5 6 21 6"/><path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"/><line x1="10" y1="11" x2="10" y2="17"/><line x1="14" y1="11" x2="14" y2="17"/></svg>
            <span>Delete</span>
          </button>
        </div>
      </div>
    </div>

    <!-- Generic Modals -->
    <BrokerModal :show="activeModal === 'orders'" title="Orders" :broker="modalBroker" @close="closeModal">
      <OrdersPageDisplay :broker="modalBroker" />
    </BrokerModal>
    <BrokerModal :show="activeModal === 'holdings'" title="Holdings" :broker="modalBroker" @close="closeModal">
      <HoldingsPageDisplay :broker="modalBroker" />
    </BrokerModal>
    <BrokerModal :show="activeModal === 'positions'" title="Positions" :broker="modalBroker" @close="closeModal">
      <PositionsPageDisplay :broker="modalBroker" />
    </BrokerModal>
    <BrokerModal :show="activeModal === 'margin'" title="Margin" :broker="modalBroker" @close="closeModal">
      <MarginPageDisplay :broker="modalBroker" />
    </BrokerModal>
    <ProfileModal :show="activeModal === 'profile'" :broker="modalBroker" @close="closeModal" />
    <PlaceOrderModal :show="activeModal === 'place-order'" :broker="modalBroker" @close="closeModal" />
  </div>
</template>

<style scoped>
.brokers-page { padding: 1rem 0; }
header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 1.25rem;
}
header h2 { margin: 0; }
.add-btn {
  background: hsl(var(--primary));
  color: hsl(var(--primary-foreground));
  border: none;
  padding: .5rem 1rem;
  border-radius: var(--radius);
  cursor: pointer;
  font-weight: 500;
}
.add-btn:hover { opacity: .9; }

.empty { text-align: center; color: hsl(var(--muted-foreground)); padding: 2rem; }

.broker-grid {
  columns: 2;
  column-gap: .75rem;
}
@media (max-width: 640px) {
  .broker-grid {
    columns: 1;
  }
}
.broker-card {
  break-inside: avoid;
  margin-bottom: .75rem;
  border: 1px solid hsl(var(--border));
  border-radius: var(--radius);
  padding: 1rem;
  background: hsl(var(--card));
}
@media (max-width: 480px) {
  .broker-grid {
    gap: .5rem;
  }
  .broker-card {
    padding: .6rem;
  }
  .broker-card .card-header strong {
    font-size: var(--font-sm);
  }
  .broker-card .card-body p {
    font-size: var(--font-xs);
  }
  .broker-card .card-actions button {
    font-size: var(--font-xs);
    padding: .3rem;
  }
}
.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: .5rem;
}
.card-title-row { display:flex; align-items:center; gap:2px; }
.info-icon {
  display:inline-flex; align-items:center; justify-content:center;
  width:18px; height:18px; border-radius:999px;
  background:hsl(var(--muted)); color:hsl(var(--muted-foreground));
  font-size:11px; font-weight:700; font-style:italic; cursor:pointer; text-decoration:none;
  line-height:1;
}
.info-icon:hover { background:hsl(var(--primary)); color:#fff; }
.badge {
  font-size: var(--font-xs);
  padding: .2rem .5rem;
  border-radius: 999px;
  background: hsl(var(--muted));
  color: hsl(var(--muted-foreground));
}
.badge.connected {
  background: hsl(144 80% 55% / .15);
  color: #16A34A;
}
.badge.error {
  background: hsl(0 84% 60% / .15);
  color: hsl(var(--destructive));
}
.badge.expired {
  background: hsl(39 100% 50% / .15);
  color: #D97706;
}
.card-body p { margin: .2rem 0; font-size: var(--font-sm); color: hsl(var(--foreground)); }
.token-status { font-size: var(--font-xs) !important; margin-top: .3rem !important; }
.token-status.connected { color: #16A34A; }
.token-status.error { color: hsl(var(--destructive)); }
.chip-grid { display: grid; grid-template-columns: repeat(3, 1fr); gap: .4rem; margin-top: .75rem; }
.chip.danger:hover { border-color: hsl(var(--destructive)); color: hsl(var(--destructive)); }
.chip.connect-chip { border-color: #16A34A; color: #16A34A; }
.chip.connect-chip.connected { background: #16A34A; color: #fff; border-color: #16A34A; }
.chip.connect-chip.loading { opacity: .7; }
.chip:disabled { opacity: .4; cursor: default; }
.chip:disabled:hover { border-color: hsl(var(--border)); color: hsl(var(--muted-foreground)); }
.chip.danger:disabled:hover { border-color: hsl(var(--border)); color: hsl(var(--muted-foreground)); }
@keyframes spin { to { transform: rotate(360deg); } }
.spin { animation: spin .8s linear infinite; }

.table-wrap { overflow-x: auto; }
.modal-data-table { width: 100%; border-collapse: collapse; font-size: var(--font-sm); }
.modal-data-table th, .modal-data-table td { padding: .4rem .5rem; border-bottom: 1px solid hsl(var(--border)); text-align: left; white-space: nowrap; }
.modal-data-table th { font-weight: 600; color: hsl(var(--foreground)); position: sticky; top: 0; background: hsl(var(--card)); }
.modal-data-table td { color: hsl(var(--muted-foreground)); }
.modal-data-table td.negative { color: hsl(var(--destructive)); }
.modal-data-table td.positive { color: #16A34A; }
.status-badge { display: inline-block; padding: .1rem .4rem; border-radius: 999px; font-size: var(--font-xs); font-weight: 600; }
.status-badge.open { background: hsl(48 100% 50% / .15); color: #b8860b; }
.status-badge.complete { background: hsl(144 80% 55% / .15); color: #16A34A; }
.status-badge.cancelled { background: hsl(0 84% 60% / .15); color: hsl(var(--destructive)); }

.summary-row { display: flex; gap: .5rem; margin-bottom: 1rem; flex-wrap: wrap; }
.summary-card { flex: 1; min-width: 100px; background: hsl(var(--card)); border: 1px solid hsl(var(--border)); border-radius: var(--radius); padding: .6rem; text-align: center; }
.summary-label { font-size: var(--font-xs); color: hsl(var(--muted-foreground)); }
.summary-value { font-size: var(--font-base); font-weight: 700; color: hsl(var(--foreground)); }
.summary-card.negative .summary-value { color: hsl(var(--destructive)); }




</style>
