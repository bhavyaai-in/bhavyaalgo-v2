<script setup>
import { ref, computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useAuthStore } from '../stores/auth.js'
import Brokers from './Brokers.vue'
import HoldingsPage from '../components/brokers/HoldingsPageDisplay.vue'
import OrdersPage from '../components/brokers/OrdersPageDisplay.vue'
import PositionsPage from '../components/brokers/PositionsPageDisplay.vue'
import MarginPage from '../components/brokers/MarginPageDisplay.vue'
import TradingWatchlist from '../components/TradingWatchlist.vue'
import PlaceOrderModal from '../modals/brokers/PlaceOrderModal.vue'
import StrategiesPage from '../components/strategies/StrategiesPage.vue'
import SettingsPage from '../components/settings/SettingsPage.vue'
import AppsPage from '../components/apps/AppsPage.vue'
import OptionChainPage from '../components/option_chain/OptionChainPage.vue'

const router = useRouter()
const route = useRoute()
const auth = useAuthStore()

const authLoading = computed(() => auth.loading)
const email = computed(() => auth.email)
const activeTab = computed(() => route.path === '/' ? 'dashboard' : route.path.slice(1))
const menuOpen = ref(false)
const watchlistOpen = ref(false)
const orderContract = ref(null)
const orderModalOpen = ref(false)

function onPlaceOrder(item) {
  orderContract.value = { symbol: item.symbol, token: item.token, exchange: item.exchange, name: item.name || item.symbol }
  orderModalOpen.value = true
}

function closeOrderModal() {
  orderModalOpen.value = false
  orderContract.value = null
}

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
  <div class="app-layout">
    <!-- Top bar: full width -->
    <header class="top-bar">
      <div class="top-left">
        <img src="/logo.png" alt="BhavyaAI" class="logo" />
        <div class="sidebar-spacer desktop-nav" />
        <nav class="top-nav desktop-nav">
          <button :class="{ active: activeTab === 'dashboard' }" @click="go('dashboard')">Dashboard</button>
          <button :class="{ active: activeTab === 'brokers' }" @click="go('brokers')">Brokers</button>
          <button :class="{ active: activeTab === 'orders' }" @click="go('orders')">Orders</button>
          <button :class="{ active: activeTab === 'positions' }" @click="go('positions')">Positions</button>
          <button :class="{ active: activeTab === 'holdings' }" @click="go('holdings')">Holdings</button>
          <button :class="{ active: activeTab === 'margin' }" @click="go('margin')">Margin</button>
          <button :class="{ active: activeTab === 'strategies' }" @click="go('strategies')">Strategies</button>
          <button :class="{ active: activeTab === 'apps' }" @click="go('apps')">Apps</button>
          <button :class="{ active: activeTab === 'settings' }" @click="go('settings')">Settings</button>
        </nav>
      </div>
      <div class="top-right">
        <button class="logout-btn desktop-logout" @click="handleLogout">Logout</button>
        <button class="hamburger" @click="menuOpen = !menuOpen">
          <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><line x1="3" y1="6" x2="21" y2="6"/><line x1="3" y1="12" x2="21" y2="12"/><line x1="3" y1="18" x2="21" y2="18"/></svg>
        </button>
      </div>
    </header>

    <!-- Mobile menu dropdown -->
    <nav v-if="menuOpen" class="mobile-menu">
      <button :class="{ active: activeTab === 'dashboard' }" @click="go('dashboard')">Dashboard</button>
      <button :class="{ active: activeTab === 'brokers' }" @click="go('brokers')">Brokers</button>
      <button :class="{ active: activeTab === 'orders' }" @click="go('orders')">Orders</button>
      <button :class="{ active: activeTab === 'positions' }" @click="go('positions')">Positions</button>
      <button :class="{ active: activeTab === 'holdings' }" @click="go('holdings')">Holdings</button>
      <button :class="{ active: activeTab === 'margin' }" @click="go('margin')">Margin</button>
      <button :class="{ active: activeTab === 'strategies' }" @click="go('strategies')">Strategies</button>
      <button :class="{ active: activeTab === 'apps' }" @click="go('apps')">Apps</button>
      <button :class="{ active: activeTab === 'settings' }" @click="go('settings')">Settings</button>
      <button @click="watchlistOpen = true; menuOpen = false">📋 Watchlist</button>
      <button class="mobile-logout" @click="handleLogout">Logout</button>
    </nav>

    <!-- Body: sidebar + content -->
    <div class="body">
      <TradingWatchlist :show="watchlistOpen" @close="watchlistOpen = false" @place-order="onPlaceOrder" />
      <main class="content">
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
        <section v-if="activeTab === 'strategies'" class="tab-content"><StrategiesPage /></section>
        <section v-if="activeTab === 'optionchain'" class="tab-content wide"><OptionChainPage /></section>
        <section v-if="activeTab === 'apps'" class="tab-content"><AppsPage /></section>
        <section v-if="activeTab === 'settings'" class="tab-content"><SettingsPage /></section>
      </main>
      <PlaceOrderModal
        :show="orderModalOpen"
        :preset-contract="orderContract"
        @close="closeOrderModal"
      />
    </div>
  </div>
</template>

<style scoped>
.app-layout { display:flex; flex-direction:column; height:100vh; overflow:hidden; }

/* Top bar — fixed at top */
.top-bar {
  display:flex; align-items:center; justify-content:space-between;
  height:52px; padding:0 1.5rem;
  border-bottom:1px solid hsl(var(--border));
  background:hsl(var(--card));
  flex-shrink:0; position:sticky; top:0; z-index:30;
}
.top-left { display:flex; align-items:center; }
.sidebar-spacer { width:380px; flex-shrink:0; }
.logo { height:32px; width:auto; margin-right:1.5rem; }
.top-nav { display:flex; align-items:center; gap:0; height:52px; }
.top-nav button {
  height:100%; padding:0 1rem; border:none; background:transparent; cursor:pointer;
  font-size:var(--font-sm); color:hsl(var(--muted-foreground)); font-weight:500;
  border-bottom:2px solid transparent;
  transition:color .15s, border-color .15s;
}
.top-nav button.active { color:hsl(var(--primary)); border-bottom-color:hsl(var(--primary)); }
.top-nav button:hover { color:hsl(var(--foreground)); }
.top-right { display:flex; align-items:center; gap:.5rem; }
.logout-btn {
  padding:.5rem 1rem; background:transparent; color:hsl(var(--destructive));
  border:1px solid hsl(var(--destructive)); border-radius:var(--radius); cursor:pointer; font-weight:500;
}
.logout-btn:hover { background:hsl(var(--destructive)); color:hsl(var(--destructive-foreground)); }

/* Body — flex row, fills remaining height */
.body { display:flex; flex:1; min-height:0; overflow:hidden; }

.content {
  flex:1; min-width:0;
  max-width:960px;
  margin:0 auto;
  padding:1.5rem 2rem;
  overflow-y:auto;
  height:100%;
}

.tab-content { min-height:200px; }
.tab-content.wide { max-width:none; }
.welcome-card { background:hsl(var(--card)); border:1px solid hsl(var(--border)); border-radius:var(--radius); padding:1.5rem; }

/* Mobile */
.hamburger { display:none; background:none; border:none; cursor:pointer; color:hsl(var(--foreground)); padding:0; }
@media (max-width:768px) {
  .desktop-nav { display:none; }
  .desktop-logout { display:none; }
  .hamburger { display:flex; align-items:center; }
  .content { padding:1rem; }
}
.mobile-menu {
  display:flex; flex-direction:column; gap:0;
  border-bottom:1px solid hsl(var(--border));
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
