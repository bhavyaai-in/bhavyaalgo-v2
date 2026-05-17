import { defineStore } from 'pinia'

export const useUiStore = defineStore('ui', {
  state: () => ({
    requestCount: 0,
    confirm: null,
  }),
  getters: {
    isLoading: (state) => state.requestCount > 0,
  },
  actions: {
    startRequest() { this.requestCount++ },
    endRequest() { if (this.requestCount > 0) this.requestCount-- },
    showConfirm(title, message) {
      return new Promise((resolve) => {
        this.confirm = { title, message, resolve }
      })
    },
    resolveConfirm(result) {
      if (this.confirm) {
        this.confirm.resolve(result)
        this.confirm = null
      }
    },
  },
})
