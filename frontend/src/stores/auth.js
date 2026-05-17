import { defineStore } from 'pinia'
import { api } from '../utils/api.js'

export const useAuthStore = defineStore('auth', {
  state: () => ({
    token: localStorage.getItem('token') || null,
    email: null,
    loading: true,
  }),
  actions: {
    async login(email, password) {
      const data = await api('/api/login', {
        method: 'POST',
        body: JSON.stringify({ email, password }),
      })
      this.token = data.token
      localStorage.setItem('token', data.token)
      await this.fetchUser()
    },
    async fetchUser() {
      if (!this.token) {
        this.loading = false
        return
      }
      try {
        const data = await api('/api/me')
        this.email = data.email
      } catch {
        this.logout()
      } finally {
        this.loading = false
      }
    },
    logout() {
      if (this.token) {
        fetch('/api/logout', { method: 'POST', headers: { Authorization: this.token } }).catch(() => {})
      }
      this.token = null
      this.email = null
      localStorage.removeItem('token')
    },
  },
})
