<script setup>
import { ref, computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useAuthStore } from '../stores/auth.js'
import Brokers from './Brokers.vue'
import HoldingsPage from '../components/brokers/HoldingsPageDisplay.vue'
import OrdersPage from '../components/brokers/OrdersPageDisplay.vue'
import PositionsPage from '../components/brokers/PositionsPageDisplay.vue'
import MarginPage from '../components/brokers/MarginPageDisplay.vue'

const router = useRouter()
const route = useRoute()
const auth = useAuthStore()

const authLoading = computed(() => auth.loading)
const email = computed(() => auth.email)
const activeTab = computed(() => route.path === '/' ? 'dashboard' : route.path.slice(1))
const menuOpen = ref(false)

function go(tab) {
  menuOpen.value = false
  router.push(tab === 'dashboard' ? '/' : '/' + tab)
}

function handleLogout() {
  auth.logout()
  router.push('/login')
}
</script>

<template>
  <main class="dashboard">
    <header>
      <img src="/logo.png" alt="BhavyaAI" class="logo" />
      <div class="header-right">
        <button class="logout-btn desktop-logout" @click="handleLogout">Logout</button>
        <button class="hamburger" @click="menuOpen = !menuOpen">
          <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><line x1="3" y1="6" x2="21" y2="6"/><line x1="3" y1="12" x2="21" y2="12"/><line x1="3" y1="18" x2="21" y2="18"/></svg>
        </button>
      </div>
    </header>

    <nav class="tabs desktop-nav">
      <button :class="{ active: activeTab === 'dashboard' }" @click="go('dashboard')">Dashboard</button>
      <button :class="{ active: activeTab === 'brokers' }" @click="go('brokers')">Brokers</button>
      <button :class="{ active: activeTab === 'orders' }" @click="go('orders')">Orders</button>
      <button :class="{ active: activeTab === 'positions' }" @click="go('positions')">Positions</button>
      <button :class="{ active: activeTab === 'holdings' }" @click="go('holdings')">Holdings</button>
      <button :class="{ active: activeTab === 'margin' }" @click="go('margin')">Margin</button>
    </nav>

    <nav v-if="menuOpen" class="mobile-menu">
      <button :class="{ active: activeTab === 'dashboard' }" @click="go('dashboard')">Dashboard</button>
      <button :class="{ active: activeTab === 'brokers' }" @click="go('brokers')">Brokers</button>
      <button :class="{ active: activeTab === 'orders' }" @click="go('orders')">Orders</button>
      <button :class="{ active: activeTab === 'positions' }" @click="go('positions')">Positions</button>
      <button :class="{ active: activeTab === 'holdings' }" @click="go('holdings')">Holdings</button>
      <button :class="{ active: activeTab === 'margin' }" @click="go('margin')">Margin</button>
      <button class="mobile-logout" @click="handleLogout">Logout</button>
    </nav>

    <section v-if="activeTab === 'dashboard'" class="tab-content">
      <div class="welcome-card">
        <p v-if="email">Welcome, <strong>{{ email }}</strong></p>
        <p v-else-if="authLoading">Loading...</p>
      </div>
    </section>

    <section v-if="activeTab === 'brokers'" class="tab-content"><Brokers /></section>
    <section v-if="activeTab === 'orders'" class="tab-content"><OrdersPage /></section>
    <section v-if="activeTab === 'positions'" class="tab-content"><PositionsPage /></section>
    <section v-if="activeTab === 'holdings'" class="tab-content"><HoldingsPage /></section>
    <section v-if="activeTab === 'margin'" class="tab-content"><MarginPage /></section>
  </main>
</template>

<style scoped>
.dashboard { max-width:960px; margin:0 auto; padding:1.5rem 2rem; }
header { display:flex; justify-content:space-between; align-items:center; margin-bottom:1.5rem; }
.header-right { display:flex; align-items:center; gap:.5rem; }
.logo { height:36px; width:auto; }
.logout-btn {
  padding:.5rem 1rem; background:transparent; color:hsl(var(--destructive));
  border:1px solid hsl(var(--destructive)); border-radius:var(--radius); cursor:pointer; font-weight:500;
}
.logout-btn:hover { background:hsl(var(--destructive)); color:hsl(var(--destructive-foreground)); }
.tabs { display:flex; gap:0; border-bottom:2px solid hsl(var(--border)); margin-bottom:1.5rem; }
.tabs button {
  padding:.6rem 1rem; border:none; background:transparent; cursor:pointer;
  font-size:var(--font-sm); color:hsl(var(--muted-foreground)); border-bottom:2px solid transparent;
  margin-bottom:-2px; font-weight:500; white-space:nowrap;
}
.tabs button.active { color:hsl(var(--primary)); border-bottom-color:hsl(var(--primary)); }
.tab-content { min-height:200px; }
.welcome-card { background:hsl(var(--card)); border:1px solid hsl(var(--border)); border-radius:var(--radius); padding:1.5rem; }

.hamburger { display:none; background:none; border:none; cursor:pointer; color:hsl(var(--foreground)); padding:0; }
@media (max-width:768px) {
  .desktop-nav { display:none; }
  .desktop-logout { display:none; }
  .hamburger { display:flex; align-items:center; }
  .dashboard { padding:1rem; }
}
.mobile-menu {
  display:flex; flex-direction:column; gap:0;
  border:1px solid hsl(var(--border)); border-radius:var(--radius);
  margin-bottom:1.5rem; overflow:hidden;
}
.mobile-menu button {
  padding:.75rem 1rem; border:none; background:transparent; cursor:pointer;
  font-size:var(--font-sm); color:hsl(var(--muted-foreground)); font-weight:500;
  text-align:left; border-bottom:1px solid hsl(var(--border));
}
.mobile-menu button:last-child { border-bottom:none; }
.mobile-menu button.active { color:hsl(var(--primary)); background:hsl(var(--primary)/.08); }
.mobile-logout { color:hsl(var(--destructive)) !important; border-top:1px solid hsl(var(--border)); }
</style>
