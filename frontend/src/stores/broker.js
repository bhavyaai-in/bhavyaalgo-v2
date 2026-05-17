import { defineStore } from 'pinia'
import { api } from '../utils/api.js'

export const useBrokerStore = defineStore('broker', {
  state: () => ({
    brokers: [],
    brokerList: [],
    loading: false,
  }),
  actions: {
    async fetchBrokers() {
      this.brokers = await api('/api/brokers')
    },
    async fetchBrokerList() {
      const list = await api('/api/broker-list')
      this.brokerList = list.filter(e => e.is_active)
    },
    async fetchAll() {
      this.loading = true
      try {
        await Promise.all([this.fetchBrokers(), this.fetchBrokerList()])
      } finally {
        this.loading = false
      }
    },
    async connect(id) {
      const res = await api('/api/connect-broker', {
        method: 'POST',
        body: JSON.stringify({ broker_id: id }),
      })
      return res
    },
    async create(data) {
      return await api('/api/brokers', { method: 'POST', body: JSON.stringify(data) })
    },
    async update(id, data) {
      return await api('/api/brokers/' + id, { method: 'PUT', body: JSON.stringify(data) })
    },
    async remove(id) {
      await api('/api/brokers/' + id, { method: 'DELETE' })
    },
  },
})
