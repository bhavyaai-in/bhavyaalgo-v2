<script setup>
import { ref, watch } from 'vue'
import HoldingsPageDisplay from './HoldingsPageDisplay.vue'

const props = defineProps({ show: Boolean, broker: Object })
const emit = defineEmits(['close'])

const raw = ref(null)
const loading = ref(false)
const error = ref('')

function cap(str) { if (!str) return ''; return str.replace(/\b\w/g, c => c.toUpperCase()) }
function token() { return localStorage.getItem('token') }
async function api(path, opts = {}) {
  const res = await fetch(path, { headers: { 'Content-Type': 'application/json', Authorization: token() }, ...opts })
  if (!res.ok) throw new Error(await res.text())
  return res.json()
}
watch(() => props.show, async (val) => {
  if (!val || !props.broker) return
  loading.value = true; error.value = ''; raw.value = null
  try { raw.value = await api(`/api/broker-holdings/${props.broker.id}`) }
  catch (e) { error.value = e.message }
  finally { loading.value = false }
})
</script>

<template>
  <div v-if="show" class="modal-overlay" @click.self="emit('close')">
    <div class="modal-box">
      <h3>Holdings — {{ cap(broker?.friendly_name || broker?.broker_name || '') }}</h3>
      <div v-if="loading" class="state-msg">Loading...</div>
      <div v-else-if="error" class="state-msg error">{{ error }}</div>
      <HoldingsPageDisplay v-else-if="raw" :data="raw" />
      <div v-else class="state-msg">No holdings.</div>
      <button class="close-btn" @click="emit('close')">Close</button>
    </div>
  </div>
</template>

<style scoped>
.modal-overlay { position:fixed; inset:0; background:rgba(0,0,0,.4); display:flex; justify-content:center; align-items:center; z-index:100; }
.modal-box { background:hsl(var(--card)); border-radius:var(--radius); padding:1.5rem; width:90%; max-width:700px; max-height:80vh; overflow-y:auto; box-shadow:0 4px 24px rgba(0,0,0,.12); }
.modal-box h3 { margin:0 0 1rem; font-size:var(--font-base); }
.state-msg { text-align:center; padding:2rem; color:hsl(var(--muted-foreground)); }
.state-msg.error { color:hsl(var(--destructive)); }
.close-btn { display:block; margin:1rem auto 0; padding:.5rem 1.5rem; border:1px solid hsl(var(--border)); background:hsl(var(--card)); border-radius:var(--radius); cursor:pointer; font-weight:500; }
</style>
