import { defineStore } from 'pinia'
import { api } from '../utils/api.js'

export const useBrokerDataStore = defineStore('brokerData', {
  state: () => ({
    orders: {},
    positions: {},
    holdings: {},
    margin: {},
    loading: {},
    errors: {},
  }),
  actions: {
    async fetchOrders(id) {
      if (this.orders[id]) return
      this.loading = { ...this.loading, [id]: true }
      try {
        this.orders = { ...this.orders, [id]: await api(`/api/broker-orders/${id}`) }
      } catch (e) {
        this.errors = { ...this.errors, [id]: e.message }
      } finally {
        const { [id]: _, ...rest } = this.loading
        this.loading = rest
      }
    },
    async fetchPositions(id) {
      if (this.positions[id]) return
      this.loading = { ...this.loading, [id]: true }
      try {
        this.positions = { ...this.positions, [id]: await api(`/api/broker-positions/${id}`) }
      } catch (e) {
        this.errors = { ...this.errors, [id]: e.message }
      } finally {
        const { [id]: _, ...rest } = this.loading
        this.loading = rest
      }
    },
    async fetchHoldings(id) {
      if (this.holdings[id]) return
      this.loading = { ...this.loading, [id]: true }
      try {
        this.holdings = { ...this.holdings, [id]: await api(`/api/broker-holdings/${id}`) }
      } catch (e) {
        this.errors = { ...this.errors, [id]: e.message }
      } finally {
        const { [id]: _, ...rest } = this.loading
        this.loading = rest
      }
    },
    async fetchMargin(id) {
      if (this.margin[id]) return
      this.loading = { ...this.loading, [id]: true }
      try {
        this.margin = { ...this.margin, [id]: await api(`/api/broker-margin/${id}`) }
      } catch (e) {
        this.errors = { ...this.errors, [id]: e.message }
      } finally {
        const { [id]: _, ...rest } = this.loading
        this.loading = rest
      }
    },
    async refreshOrders(id) {
      delete this.orders[id]
      await this.fetchOrders(id)
    },
    async refreshPositions(id) {
      delete this.positions[id]
      await this.fetchPositions(id)
    },
    async refreshHoldings(id) {
      delete this.holdings[id]
      await this.fetchHoldings(id)
    },
    async refreshMargin(id) {
      delete this.margin[id]
      await this.fetchMargin(id)
    },
    getOrders(id) { return this.orders[id] || [] },
    getPositions(id) { return this.positions[id] || [] },
    getHoldings(id) { return this.holdings[id] || null },
    getMargin(id) { return this.margin[id] || null },
    isLoading(id) { return !!this.loading[id] },
    getError(id) { return this.errors[id] || '' },
  },
})
