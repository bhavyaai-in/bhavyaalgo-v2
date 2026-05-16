<script setup>
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { useAuth } from '../stores/auth.js'

const router = useRouter()
const { login } = useAuth()

const email = ref('admin@example.com')
const password = ref('password123')
const error = ref('')
const submitting = ref(false)

async function handleSubmit() {
  error.value = ''
  submitting.value = true
  try {
    await login(email.value, password.value)
    router.push('/')
  } catch (e) {
    error.value = e.message
  } finally {
    submitting.value = false
  }
}
</script>

<template>
  <main class="login-page">
    <form @submit.prevent="handleSubmit">
      <img src="/logo.png" alt="BhavyaAI" class="logo" />
      <h1>Sign in</h1>
      <p v-if="error" class="error">{{ error }}</p>
      <label>
        Email
        <input v-model="email" type="email" required autocomplete="email" />
      </label>
      <label>
        Password
        <input v-model="password" type="password" required autocomplete="current-password" />
      </label>
      <button type="submit" :disabled="submitting">
        {{ submitting ? 'Signing in...' : 'Sign in' }}
      </button>
    </form>
  </main>
</template>

<style scoped>
.login-page {
  display: flex;
  justify-content: center;
  align-items: center;
  min-height: 100vh;
  background: hsl(var(--background));
}
.logo {
  display: block;
  width: 120px;
  margin: 0 auto 1rem;
}
form {
  background: hsl(var(--card));
  padding: 2.5rem;
  border-radius: var(--radius);
  box-shadow: 0 2px 12px rgba(0,0,0,.08);
  width: 100%;
  max-width: 360px;
  display: flex;
  flex-direction: column;
  gap: 1rem;
}
h1 {
  margin: 0;
  font-size: var(--font-2xl);
  text-align: center;
  color: var(--heading-color);
}
.error {
  margin: 0;
  color: hsl(var(--destructive));
  font-size: var(--font-sm);
  text-align: center;
}
label {
  display: flex;
  flex-direction: column;
  gap: .3rem;
  font-size: var(--font-sm);
  color: hsl(var(--foreground));
}
input {
  padding: .6rem .8rem;
  border: 1px solid hsl(var(--input));
  border-radius: var(--radius);
  font-size: var(--font-base);
}
input:focus {
  outline: none;
  border-color: hsl(var(--ring));
  box-shadow: 0 0 0 2px hsl(var(--ring) / .2);
}
button {
  padding: .7rem;
  background: hsl(var(--primary));
  color: hsl(var(--primary-foreground));
  border: none;
  border-radius: var(--radius);
  font-size: var(--font-base);
  cursor: pointer;
  font-weight: 500;
}
button:disabled {
  opacity: .6;
}
button:not(:disabled):hover {
  opacity: .9;
}
</style>
