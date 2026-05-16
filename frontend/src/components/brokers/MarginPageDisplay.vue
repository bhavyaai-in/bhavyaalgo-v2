<script setup>
import { ref, computed, onMounted, watch } from 'vue'
import { useRoute } from 'vue-router'
import { api } from '../../utils/api.js'

const props = defineProps({ data: null, broker: null })
const route = useRoute()
const brokers = ref([])
const selectedId = ref(route.query.broker || '')
const margin = ref(null)
const loading = ref(false)
const error = ref('')
const display = computed(() => props.data || margin.value)

function cap(str) { if (!str) return ''; return str.replace(/\b\w/g, c => c.toUpperCase()) }
async function fetchBrokers() {
  brokers.value = await api('/api/brokers')
  if (brokers.value.length > 0 && !selectedId.value) selectedId.value = String(brokers.value[0].id)
}
async function fetchData() {
  const id = selectedId.value
  if (!id) return
  loading.value = true; error.value = ''; margin.value = null
  try { margin.value = await api(`/api/broker-margin/${id}`) }
  catch (e) { error.value = e.message }
  finally { loading.value = false }
}
watch(selectedId, () => { if (selectedId.value) fetchData() })
onMounted(async () => {
  await fetchBrokers()
  if (props.broker) selectedId.value = String(props.broker.id)
})
</script>

<template>
  <div class="page">
    <header>
      <h2>Margin</h2>
      <select v-model="selectedId" class="broker-select">
        <option v-for="b in brokers" :key="b.id" :value="b.id">{{ b.friendly_name || b.broker_name }}</option>
      </select>
    </header>
    <div v-if="loading" class="state-msg">Loading...</div>
    <div v-else-if="error" class="state-msg error">{{ error }}</div>
    <div v-else-if="display" class="table-wrap">
      <table class="data-table kv">
        <tr v-for="(v,k) in display" :key="k"><td class="pkey">{{ cap(k.replace(/_/g,' ')) }}</td><td>{{ v }}</td></tr>
      </table>
    </div>
    <div v-else class="state-msg">No margin data.</div>
  </div>
</template>

<style scoped>
</style>
