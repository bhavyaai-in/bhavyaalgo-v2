<script setup>
import { ref, computed, watch } from 'vue'

const props = defineProps({
  show: Boolean,
  editing: [Number, null],
  form: Object,
  brokerList: Array,
})

const emit = defineEmits(['close', 'save'])

const columns = ref({})
const fieldsVisible = ref(false)

const fieldMeta = {
  friendly_name: { label: 'Friendly Name' },
  broker_userid: { label: 'User ID' },
  broker_password: { label: 'Password' },
  broker_pin: { label: 'PIN' },
  broker_qr_key: { label: 'QR Key' },
  broker_api: { label: 'API Key' },
  broker_api_secret: { label: 'API Secret' },
}

const visibleFields = computed(() => {
  const cols = columns.value[props.form.broker_name] || []
  return cols.map(name => ({
    key: name,
    label: (fieldMeta[name] || {}).label || name,
  }))
})

watch(() => props.show, async (open) => {
  if (!open) return
  const res = await fetch('/api/broker-columns', {
    headers: { Authorization: localStorage.getItem('token') },
  })
  columns.value = await res.json()
  fieldsVisible.value = !!props.editing
})

watch(() => props.form.broker_name, (name) => {
  if (props.show && !props.editing && name) {
    fieldsVisible.value = true
  }
})
</script>

<template>
  <div v-if="show" class="modal-overlay" @click.self="emit('close')">
    <form class="broker-form" @submit.prevent="emit('save')">
      <h3>{{ editing ? 'Edit Broker' : 'Add Broker' }}</h3>

      <label>Broker
        <select v-model="form.broker_name" required>
          <option value="" disabled>select</option>
          <option v-for="e in brokerList" :key="e.id" :value="e.broker_name">
            {{ e.broker_name }}
          </option>
        </select>
      </label>

      <template v-if="fieldsVisible">
        <div class="field-grid">
          <label v-for="f in visibleFields" :key="f.key">
            {{ f.label }}
            <input v-model="form[f.key]" />
          </label>
        </div>
        <div class="checkboxes">
          <label class="checkbox">
            <input type="checkbox" v-model="form.is_active" /> Active
          </label>
          <label class="checkbox">
            <input type="checkbox" v-model="form.is_autologin" /> Auto Login
          </label>
        </div>
        <div class="form-actions">
          <button type="submit">{{ editing ? 'Update' : 'Create' }}</button>
          <button type="button" class="cancel" @click="emit('close')">Cancel</button>
        </div>
      </template>
    </form>
  </div>
</template>

<style scoped>
.modal-overlay {
  position: fixed;
  inset: 0;
  background: rgba(0,0,0,.4);
  display: flex;
  justify-content: center;
  align-items: center;
  z-index: 100;
}
.broker-form {
  background: hsl(var(--card));
  padding: 2rem;
  border-radius: var(--radius);
  width: 90%;
  max-width: 480px;
  max-height: 90vh;
  overflow-y: auto;
  display: flex;
  flex-direction: column;
  gap: .8rem;
  box-shadow: 0 4px 24px rgba(0,0,0,.12);
}
.broker-form h3 { margin: 0 0 .5rem; }
.broker-form label {
  display: flex;
  flex-direction: column;
  gap: .3rem;
  font-size: var(--font-sm);
  color: hsl(var(--foreground));
}
.broker-form input,
.broker-form select {
  padding: .5rem .7rem;
  border: 1px solid hsl(var(--input));
  border-radius: var(--radius);
  font-size: var(--font-sm);
}
.broker-form input:focus,
.broker-form select:focus {
  outline: none;
  border-color: hsl(var(--ring));
  box-shadow: 0 0 0 2px hsl(var(--ring) / .2);
}
.field-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: .6rem;
}
@media (max-width: 640px) {
  .field-grid {
    grid-template-columns: 1fr;
  }
}
.broker-form .checkbox {
  flex-direction: row;
  align-items: center;
  gap: .5rem;
}
.checkboxes {
  display: flex;
  gap: 1rem;
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
.form-actions .cancel {
  background: hsl(var(--muted-foreground));
}
</style>
