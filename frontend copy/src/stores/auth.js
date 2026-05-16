import { reactive } from 'vue'

const state = reactive({
  token: localStorage.getItem('token') || null,
  email: null,
  loading: true,
})

export function useAuth() {
  async function login(email, password) {
    const res = await fetch('/api/login', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ email, password }),
    })
    if (!res.ok) {
      const err = await res.json().catch(() => ({ error: 'login failed' }))
      throw new Error(err.error)
    }
    const data = await res.json()
    state.token = data.token
    localStorage.setItem('token', data.token)
    const me = await fetchUser()
    return me
  }

  async function fetchUser() {
    if (!state.token) {
      state.loading = false
      return
    }
    const res = await fetch('/api/me', {
      headers: { Authorization: state.token },
    })
    if (!res.ok) {
      logout()
      state.loading = false
      return
    }
    const data = await res.json()
    state.email = data.email
    state.loading = false
  }

  function logout() {
    if (state.token) {
      fetch('/api/logout', {
        method: 'POST',
        headers: { Authorization: state.token },
      }).catch(() => {})
    }
    state.token = null
    state.email = null
    localStorage.removeItem('token')
  }

  return { state, login, fetchUser, logout }
}
