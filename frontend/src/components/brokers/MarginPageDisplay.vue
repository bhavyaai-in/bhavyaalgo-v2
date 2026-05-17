<script setup>
import { useBrokerData } from '../../composables/useBrokerData.js'

const props = defineProps({ data: null, broker: null })
const { brokers, selectedId, data, loading, error } = useBrokerData(props, 'margin')

function cap(str) { if (!str) return ''; return str.replace(/\b\w/g, c => c.toUpperCase()) }
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
    <div v-else-if="data" class="table-wrap">
      <table class="data-table kv">
        <tr v-for="(v,k) in data" :key="k"><td class="pkey">{{ cap(k.replace(/_/g,' ')) }}</td><td>{{ v }}</td></tr>
      </table>
    </div>
    <div v-else class="state-msg">No margin data.</div>
  </div>
</template>

<style scoped>
</style>
