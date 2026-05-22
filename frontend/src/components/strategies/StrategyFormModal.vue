<script setup>
import { ref, watch } from 'vue'
import { api } from '../../utils/api.js'

const props = defineProps({ show: Boolean, strategy: Object, types: Array, brokers: Array })
const emit = defineEmits(['close', 'saved'])

const form = ref({
  name: '',
  strategy_type: '',
  exchange: '',
  side: 'SELL',
  atm_otm: 0,
  color: '#6366f1',
  is_active: 1,
  is_locked: 0,
  message: '',
  expiry_date: '',
  image_url: '',
})
const submitting = ref(false)
const error = ref('')
const success = ref(false)

watch(() => props.show, (val) => {
  if (!val) return
  error.value = ''
  success.value = false
  if (props.strategy) {
    form.value = {
      name: props.strategy.name || '',
      strategy_type: String(props.strategy.strategy_type || ''),
      exchange: props.strategy.exchange || '',
      side: props.strategy.side || 'SELL',
      atm_otm: props.strategy.atm_otm || 0,
      color: props.strategy.color || '#6366f1',
      is_active: props.strategy.is_active || 0,
      is_locked: props.strategy.is_locked || 0,
      message: props.strategy.message || '',
      expiry_date: props.strategy.expiry_date || '',
      image_url: props.strategy.image_url || '',
    }
  } else {
    form.value = {
      name: '', strategy_type: '', exchange: '',
      side: 'SELL', atm_otm: 0, color: '#6366f1',
      is_active: 1, is_locked: 0, message: '', expiry_date: '', image_url: '',
    }
  }
})

async function save() {
  submitting.value = true
  error.value = ''
  const body = {
    name: form.value.name,
    strategy_secret_key: '',
    strategy_type: Number(form.value.strategy_type),
    position_status: 0,
    instrument_token: 0,
    exchange: form.value.exchange,
    side: form.value.side,
    atm_otm: Number(form.value.atm_otm),
    image_url: form.value.image_url,
    color: form.value.color,
    is_active: Number(form.value.is_active),
    is_locked: Number(form.value.is_locked),
    message: form.value.message,
    expiry_date: form.value.expiry_date,
  }
  try {
    if (props.strategy) {
      await api('/api/strategies/' + props.strategy.id, { method: 'PUT', body: JSON.stringify(body) })
    } else {
      await api('/api/strategies', { method: 'POST', body: JSON.stringify(body) })
    }
    success.value = true
    setTimeout(() => emit('saved'), 800)
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
      <h3>{{ strategy ? 'Edit Strategy' : 'New Strategy' }}</h3>

      <template v-if="success">
        <div class="success-msg">Strategy saved!</div>
      </template>

      <template v-else>
        <div class="field-grid">
          <label>Name <input v-model="form.name" required /></label>
          <label>Strategy Type
            <select v-model="form.strategy_type">
              <option value="">-- Select --</option>
              <option v-for="t in types" :key="t.id" :value="t.id">{{ t.name }}</option>
            </select>
          </label>
          <label>Exchange(s) <input v-model="form.exchange" placeholder="NSE,BSE,NFO" /></label>
          <label>Side
            <select v-model="form.side">
              <option>BUY</option><option>SELL</option><option>BOTH</option>
            </select>
          </label>
          <label>ATM/OTM <input v-model.number="form.atm_otm" type="number" step="0.5" /></label>
          <label>Color <input v-model="form.color" type="color" style="height:2.2rem;padding:2px" /></label>
          <label>Expiry Date <input v-model="form.expiry_date" type="date" /></label>
          <label>Active
            <select v-model.number="form.is_active">
              <option :value="1">Yes</option><option :value="0">No</option>
            </select>
          </label>
          <label>Locked
            <select v-model.number="form.is_locked">
              <option :value="0">No</option><option :value="1">Yes</option>
            </select>
          </label>
          <label>Image URL <input v-model="form.image_url" /></label>
          <label class="full-width">Message <input v-model="form.message" /></label>
        </div>
        <div v-if="error" class="error-msg">{{ error }}</div>
        <div class="form-actions">
          <button @click="save" :disabled="submitting || !form.name || !form.strategy_type">{{ submitting ? 'Saving...' : (strategy ? 'Update' : 'Create') }}</button>
          <button class="cancel" @click="emit('close')">Cancel</button>
        </div>
      </template>
    </div>
  </div>
</template>

<style scoped>
.modal-box { max-width:520px; }
.field-grid { display:grid; grid-template-columns:1fr 1fr; gap:.6rem; }
.field-grid label { display:flex; flex-direction:column; gap:.2rem; font-size:var(--font-sm); color:hsl(var(--foreground)); }
.full-width { grid-column:1/-1; }
.field-grid input, .field-grid select { padding:.45rem .6rem; border:1px solid hsl(var(--input)); border-radius:var(--radius); font-size:var(--font-sm); outline:none; background:hsl(var(--card)); }
.field-grid input:focus, .field-grid select:focus { border-color:hsl(var(--ring)); box-shadow:0 0 0 2px hsl(var(--ring)/.2); }
.error-msg { padding:.5rem; color:hsl(var(--destructive)); font-size:var(--font-sm); text-align:center; }
.success-msg { text-align:center; padding:2rem; font-size:var(--font-lg); font-weight:700; color:#16A34A; }
.form-actions { display:flex; gap:.5rem; margin-top:.75rem; }
.form-actions button { flex:1; padding:.6rem; border:none; border-radius:var(--radius); cursor:pointer; font-weight:500; color:#fff; background:hsl(var(--primary)); }
.form-actions button:disabled { opacity:.5; }
.form-actions .cancel { background:hsl(var(--muted-foreground)); }
</style>
