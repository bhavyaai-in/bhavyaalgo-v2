<script setup>
import HoldingsTable from './HoldingsTable.vue'
import { useBrokerData } from '../../composables/useBrokerData.js'

const props = defineProps({ data: null, broker: null })
const { brokers, selectedId, data, loading, error } = useBrokerData(props, 'holdings')
</script>

<template>
  <div class="page">
    <header>
      <h2>Holdings</h2>
      <select v-model="selectedId" class="broker-select">
        <option v-for="b in brokers" :key="b.id" :value="b.id">{{ b.friendly_name || b.broker_name }}</option>
      </select>
    </header>
    <div v-if="loading" class="state-msg">Loading...</div>
    <div v-else-if="error" class="state-msg error">{{ error }}</div>
    <HoldingsTable v-else-if="data" :data="data" />
    <div v-else class="state-msg">No holdings.</div>
  </div>
</template>

<style scoped>
</style>
