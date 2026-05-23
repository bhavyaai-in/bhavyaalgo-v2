<script setup>
import { ref, computed, onMounted, onUnmounted, watch } from 'vue'
import BrokerState from './BrokerState.vue'
import PositionsTable from './PositionsTable.vue'
import { useBrokerData } from '../../composables/useBrokerData.js'

const props = defineProps({ data: null, broker: null })
const { brokers, selectedId, data, loading, error } = useBrokerData(props, 'positions')

// Dropdown state logic
const isDropdownOpen = ref(false)
const dropdownRef = ref(null)

// 1. Is watch ko update karo (taaki brokers list aate hi 1st id perfectly bind ho jaye)
watch(() => brokers.value, (newBrokers) => {
  if (newBrokers && newBrokers.length > 0) {
    // Agar selectedId pehle se empty hai ya list me exist nahi karti, toh 1st broker ki id daal do
    const exists = newBrokers.some(b => b.id === selectedId.value)
    if (!selectedId.value || !exists) {
      selectedId.value = newBrokers[0].id
    }
  }
}, { immediate: true, deep: true })

// 2. Is computed property ko update karo (yeh UI aur list dono ko sync rakkhega)
const selectedBrokerName = computed(() => {
  if (!brokers.value || brokers.value.length === 0) return 'No Broker Available'
  
  const current = brokers.value.find(b => b.id === selectedId.value)
  if (current) {
    return current.friendly_name || current.broker_name
  }
  
  // Fallback: agar kuch match na ho toh first waala return kare aur safe side id bhi set kar de
  return brokers.value[0].friendly_name || brokers.value[0].broker_name
})

const stateType = computed(() => {
  if (loading.value) return 'loading'
  if (error.value) {
    if (error.value === 'could not connect broker' || error.value === 'broker token not generated') {
      return 'disconnected'
    }
    return 'error'
  }
  if (data.value && data.value.length === 0) return 'empty'
  return null
})

const stateMessage = computed(() => {
  if (loading.value) return 'Loading...'
  if (error.value) return error.value
  if (data.value && data.value.length === 0) return 'No positions.'
  return ''
})

function toggleDropdown() {
  isDropdownOpen.value = !isDropdownOpen.value
}

function selectBroker(id) {
  selectedId.value = id
  isDropdownOpen.value = false
}

// Bahar click karne par list ko close karne ke liye
function handleClickOutside(event) {
  if (dropdownRef.value && !dropdownRef.value.contains(event.target)) {
    isDropdownOpen.value = false
  }
}

onMounted(() => { window.addEventListener('click', handleClickOutside) })
onUnmounted(() => { window.removeEventListener('click', handleClickOutside) })
</script>

<template>
  <div class="page">
    <header>
      <h2>Positions</h2>
      
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

    <BrokerState v-if="stateType" :type="stateType" :message="stateMessage" />
    <PositionsTable v-else-if="data && data.length" :items="data" />
  </div>
</template>

<style scoped>
/* Custom Dropdown Styles */
.custom-select-container {
  position: relative;
  display: inline-block;
  user-select: none;
}

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
</style>