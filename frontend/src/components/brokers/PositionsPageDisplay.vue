<script setup>
import { ref, computed, onMounted, watch } from 'vue'
import { useRoute } from 'vue-router'
import { api } from '../../utils/api.js'
import PositionsTable from './PositionsTable.vue'

const props = defineProps({ data: null, broker: null })

const route = useRoute()
const brokers = ref([])
const selectedId = ref(route.query.broker || '')
const positions = ref([])
const loading = ref(false)
const error = ref('')
const display = computed(() => props.data || positions.value)

async function fetchBrokers() {
  brokers.value = await api('/api/brokers')
  if (brokers.value.length > 0 && !selectedId.value) selectedId.value = String(brokers.value[0].id)
}
async function fetchData() {
  const id = selectedId.value
  if (!id) return
  loading.value = true; error.value = ''; positions.value = []
  try { positions.value = await api(`/api/broker-positions/${id}`) }
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
      <h2>Positions</h2>
      <select v-model="selectedId" class="broker-select">
        <option v-for="b in brokers" :key="b.id" :value="b.id">{{ b.friendly_name || b.broker_name }}</option>
      </select>
    </header>
    <div v-if="loading" class="state-msg">Loading...</div>
    <div v-else-if="error" class="state-msg error">{{ error }}</div>
    <PositionsTable v-else-if="display.length" :items="display" />
    <div v-else class="state-msg">No positions.</div>
  </div>
</template>

<style scoped>
</style>
