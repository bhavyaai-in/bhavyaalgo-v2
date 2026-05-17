import { ref, computed, onMounted, watch } from 'vue'
import { useRoute } from 'vue-router'
import { api } from '../utils/api.js'
import { useBrokerDataStore } from '../stores/brokerData.js'

/**
 * Shared composable for OrdersPageDisplay, HoldingsPageDisplay, etc.
 *
 * Eliminates duplicated watch/onMounted/fetchBrokers logic across all
 * PageDisplay components by providing:
 *  - broker list fetching
 *  - selected broker tracking (from route query, prop, or first broker)
 *  - data fetching via Pinia store (cached, deduplicated)
 *  - loading / error states
 */
export function useBrokerData(props, endpoint) {
  const route = useRoute()
  const store = useBrokerDataStore()

  const brokers = ref([])
  const selectedId = ref(route.query.broker || '')

  async function fetchBrokers() {
    brokers.value = await api('/api/brokers')
    if (brokers.value.length > 0 && !selectedId.value && !props.broker) {
      selectedId.value = String(brokers.value[0].id)
    }
  }

  const data = computed(() => {
    const id = selectedId.value
    if (!id) return null
    switch (endpoint) {
      case 'orders': return store.getOrders(id)
      case 'positions': return store.getPositions(id)
      case 'holdings': return store.getHoldings(id)
      case 'margin': return store.getMargin(id)
      default: return null
    }
  })

  const loading = computed(() => store.isLoading(selectedId.value))
  const error = computed(() => store.getError(selectedId.value))

  async function loadData() {
    const id = selectedId.value
    if (!id) return
    switch (endpoint) {
      case 'orders': await store.fetchOrders(id); break
      case 'positions': await store.fetchPositions(id); break
      case 'holdings': await store.fetchHoldings(id); break
      case 'margin': await store.fetchMargin(id); break
    }
  }

  watch(selectedId, () => {
    if (selectedId.value) loadData()
  })

  onMounted(async () => {
    await fetchBrokers()
    if (props.broker) selectedId.value = String(props.broker.id)
    else if (selectedId.value) loadData()
  })

  return { brokers, selectedId, data, loading, error, loadData }
}
