import { defineStore } from 'pinia'

export const useNotificationStore = defineStore('notification', {
  state: () => ({
    notifications: [],
  }),
  actions: {
    add(notif) {
      const id = Date.now() + Math.random()
      this.notifications.push({ id, ...notif })
      setTimeout(() => this.remove(id), 5000)
    },
    remove(id) {
      this.notifications = this.notifications.filter(n => n.id !== id)
    },
  },
})
