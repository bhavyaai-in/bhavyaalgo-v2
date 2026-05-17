<script setup>
import { ref, computed } from 'vue'

const activeTab = ref('watchlist')
const activeSubList = ref('mywatchlist')
const searchQuery = ref('')
const hoveredSymbol = ref(null)
const showActionMenu = ref(null)

const watchlist = ref([
  { symbol: 'ADANIENT', exchange: 'NSE', ltp: 2845.65, change: 21.70, pct: 1.22, folder: true },
  { symbol: 'SBIN', exchange: 'NSE', ltp: 968.20, change: 10.55, pct: 1.10, folder: false },
  { symbol: 'RELIANCE', exchange: 'NSE', ltp: 2145.30, change: -15.20, pct: -0.70, folder: true },
  { symbol: 'TCS', exchange: 'NSE', ltp: 3892.00, change: -22.45, pct: -0.57, folder: false },
  { symbol: 'HDFCBANK', exchange: 'NSE', ltp: 1678.90, change: 8.35, pct: 0.50, folder: true },
  { symbol: 'INFY', exchange: 'NSE', ltp: 1567.45, change: -5.60, pct: -0.36, folder: false },
  { symbol: 'ICICIBANK', exchange: 'NSE', ltp: 1234.55, change: 15.80, pct: 1.30, folder: false },
  { symbol: 'WIPRO', exchange: 'NSE', ltp: 478.20, change: -3.10, pct: -0.64, folder: false },
  { symbol: 'ITC', exchange: 'NSE', ltp: 432.15, change: 2.45, pct: 0.57, folder: false },
  { symbol: 'BAJFINANCE', exchange: 'NSE', ltp: 6789.00, change: -45.50, pct: -0.66, folder: true },
])

const filteredList = computed(() => {
  if (!searchQuery.value) return watchlist.value
  const q = searchQuery.value.toUpperCase()
  return watchlist.value.filter(item => item.symbol.includes(q))
})

function isPositive(item) {
  return item.change >= 0
}

function toggleAction(symbol) {
  showActionMenu.value = showActionMenu.value === symbol ? null : symbol
}
</script>

<template>
  <aside class="watchlist-sidebar">
    <!-- Header -->
    <div class="sidebar-header">
      <div class="header-tabs">
        <button
          class="header-tab"
          :class="{ active: activeTab === 'watchlist' }"
          @click="activeTab = 'watchlist'"
        >Watchlist</button>
        <button
          class="header-tab"
          :class="{ active: activeTab === 'options' }"
          @click="activeTab = 'options'"
        >Option Chain</button>
      </div>
      <div class="header-actions">
        <button class="icon-btn" title="Settings">⚙️</button>
        <button class="icon-btn" title="Close">✕</button>
      </div>
    </div>

    <!-- Sub Watchlist Row -->
    <div class="sub-watchlist">
      <div class="sub-tabs">
        <button
          class="sub-tab"
          :class="{ active: activeSubList === 'mywatchlist' }"
          @click="activeSubList = 'mywatchlist'"
        >mywatchlist</button>
        <button
          class="sub-tab"
          :class="{ active: activeSubList === 'fut' }"
          @click="activeSubList = 'fut'"
        >FUT</button>
      </div>
      <div class="sub-actions">
        <button class="icon-btn sm" title="Add symbol">➕</button>
        <button class="icon-btn sm" title="Layout">&#9674;</button>
      </div>
    </div>

    <!-- Search -->
    <div class="search-bar">
      <div class="search-wrap">
        <span class="search-icon">🔍</span>
        <input v-model="searchQuery" type="text" placeholder="Search" class="search-input" />
        <span class="filter-icon">⊶</span>
      </div>
    </div>

    <!-- Watchlist -->
    <div class="watchlist-body">
      <div
        v-for="item in filteredList"
        :key="item.symbol"
        class="watchlist-row"
        @mouseenter="hoveredSymbol = item.symbol"
        @mouseleave="hoveredSymbol = null"
      >
        <div class="row-left">
          <div class="symbol-info">
            <span class="symbol-name">{{ item.symbol }}</span>
            <span class="symbol-exchange">{{ item.exchange }}</span>
          </div>
          <span v-if="item.folder" class="badge-folder">📁 1</span>
        </div>
        <div class="row-right">
          <div class="price-row" :class="isPositive(item) ? 'up' : 'down'">
            <span class="arrow">{{ isPositive(item) ? '▲' : '▼' }}</span>
            <span class="ltp">{{ item.ltp.toFixed(2) }}</span>
          </div>
          <div class="change-row" :class="isPositive(item) ? 'up' : 'down'">
            <span>{{ item.change >= 0 ? '+' : '' }}{{ item.change.toFixed(2) }}</span>
            <span> ({{ item.pct >= 0 ? '+' : '' }}{{ item.pct.toFixed(2) }}%)</span>
          </div>
        </div>

        <!-- Floating Action Menu -->
        <div v-if="showActionMenu === item.symbol && hoveredSymbol === item.symbol" class="action-menu">
          <button class="act-btn buy">B</button>
          <button class="act-btn sell">S</button>
          <span class="act-sep" />
          <button class="act-icon" title="Chart">📊</button>
          <button class="act-icon" title="Link">🔗</button>
          <button class="act-icon" title="Flag">🏳️</button>
          <button class="act-icon" title="Delete">🗑️</button>
          <button class="act-icon" title="More">⋮</button>
        </div>
      </div>
    </div>

    <!-- Footer -->
    <div class="sidebar-footer">
      <span class="footer-label">OPTIONS QUICK LIST</span>
      <button class="footer-expand">❯</button>
    </div>
  </aside>
</template>

<style scoped>
.watchlist-sidebar {
  position: fixed;
  right: 0;
  top: 0;
  bottom: 0;
  width: 380px;
  height: 100vh;
  display: flex;
  flex-direction: column;
  background: hsl(var(--card));
  border-left: 1px solid hsl(var(--border));
  z-index: 50;
  font-family: system-ui, -apple-system, sans-serif;
}

/* Header */
.sidebar-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  height: 48px;
  padding: 0 1rem;
  border-bottom: 1px solid hsl(var(--border));
  flex-shrink: 0;
}
.header-tabs { display: flex; align-items: center; gap: 0; height: 100%; }
.header-tab {
  height: 100%;
  padding: 0 .75rem;
  border: none;
  background: transparent;
  cursor: pointer;
  font-size: .875rem;
  font-weight: 600;
  color: hsl(var(--muted-foreground));
  border-bottom: 2px solid transparent;
  transition: color .15s, border-color .15s;
}
.header-tab.active {
  color: hsl(var(--primary));
  border-bottom-color: hsl(var(--primary));
}
.header-tab:hover { color: hsl(var(--foreground)); }
.header-actions { display: flex; align-items: center; gap: .25rem; }
.icon-btn {
  width: 28px; height: 28px;
  display: flex; align-items: center; justify-content: center;
  border: none; background: transparent; cursor: pointer;
  border-radius: 4px; font-size: .85rem;
  color: hsl(var(--muted-foreground));
}
.icon-btn:hover { background: hsl(var(--muted)); color: hsl(var(--foreground)); }
.icon-btn.sm { width: 24px; height: 24px; font-size: .75rem; }

/* Sub Watchlist */
.sub-watchlist {
  display: flex; align-items: center; justify-content: space-between;
  height: 40px; padding: 0 1rem;
  border-bottom: 1px solid hsl(var(--border));
  flex-shrink: 0;
}
.sub-tabs { display: flex; align-items: center; gap: .75rem; }
.sub-tab {
  border: none; background: transparent; cursor: pointer;
  font-size: .75rem; font-weight: 500; padding: 0;
  color: hsl(var(--muted-foreground));
  border-bottom: 2px solid transparent;
  transition: color .15s, border-color .15s;
}
.sub-tab.active {
  color: hsl(var(--primary));
  border-bottom-color: hsl(var(--primary));
}
.sub-tab:hover { color: hsl(var(--foreground)); }
.sub-actions { display: flex; gap: .125rem; }

/* Search */
.search-bar {
  padding: .75rem;
  flex-shrink: 0;
}
.search-wrap {
  position: relative;
  display: flex;
  align-items: center;
}
.search-icon {
  position: absolute;
  left: .625rem;
  font-size: .8rem;
  color: hsl(var(--muted-foreground));
  pointer-events: none;
}
.search-input {
  width: 100%;
  height: 36px;
  padding: 0 2rem 0 2rem;
  border: 1px solid hsl(var(--input));
  border-radius: .5rem;
  background: hsl(var(--muted));
  font-size: .8125rem;
  color: hsl(var(--foreground));
  outline: none;
}
.search-input::placeholder { color: hsl(var(--muted-foreground)); }
.search-input:focus { border-color: hsl(var(--ring)); box-shadow: 0 0 0 2px hsl(var(--ring)/.15); }
.filter-icon {
  position: absolute;
  right: .625rem;
  font-size: .8rem;
  color: hsl(var(--muted-foreground));
  cursor: pointer;
}
.filter-icon:hover { color: hsl(var(--foreground)); }

/* Watchlist Body */
.watchlist-body {
  flex: 1;
  overflow-y: auto;
}

.watchlist-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: .75rem 1rem;
  border-bottom: 1px solid hsl(var(--border)/.5);
  position: relative;
  cursor: pointer;
  transition: background .1s;
}
.watchlist-row:hover { background: hsl(var(--muted)/.5); }

.row-left {
  display: flex;
  align-items: center;
  gap: .5rem;
  min-width: 0;
}
.symbol-info {
  display: flex;
  flex-direction: column;
  gap: .1rem;
}
.symbol-name {
  font-size: .8125rem;
  font-weight: 700;
  color: hsl(var(--foreground));
  line-height: 1.2;
}
.symbol-exchange {
  font-size: .6875rem;
  color: hsl(var(--muted-foreground));
}
.badge-folder {
  font-size: .625rem;
  padding: .1rem .3rem;
  background: hsl(var(--muted));
  border-radius: 4px;
  white-space: nowrap;
}

.row-right {
  display: flex;
  flex-direction: column;
  align-items: flex-end;
  gap: .1rem;
}
.price-row {
  display: flex;
  align-items: center;
  gap: .35rem;
  font-size: .8125rem;
  font-weight: 700;
}
.price-row .arrow { font-size: .65rem; }
.price-row.up { color: #16A34A; }
.price-row.down { color: hsl(var(--destructive)); }
.change-row {
  font-size: .6875rem;
  font-weight: 500;
  display: flex;
}
.change-row.up { color: #16A34A; }
.change-row.down { color: hsl(var(--destructive)); }

/* Action Menu */
.action-menu {
  position: absolute;
  left: 50%;
  transform: translateX(-50%);
  top: -10px;
  z-index: 50;
  display: flex;
  align-items: center;
  gap: .2rem;
  padding: .35rem .5rem;
  background: hsl(var(--card));
  border: 1px solid hsl(var(--border));
  border-radius: .5rem;
  box-shadow: 0 4px 16px rgba(0,0,0,.12);
}
.act-btn {
  height: 28px; width: 28px;
  display: flex; align-items: center; justify-content: center;
  border: none; border-radius: 6px;
  font-size: .75rem; font-weight: 700;
  cursor: pointer;
}
.act-btn.buy { background: #16A34A; color: #fff; }
.act-btn.sell { background: hsl(var(--destructive)); color: #fff; }
.act-btn:hover { opacity: .85; }
.act-sep { width: 1px; height: 20px; background: hsl(var(--border)); }
.act-icon {
  width: 24px; height: 24px;
  display: flex; align-items: center; justify-content: center;
  border: none; background: transparent; cursor: pointer;
  border-radius: 4px; font-size: .8rem;
  color: hsl(var(--muted-foreground));
}
.act-icon:hover { background: hsl(var(--muted)); color: hsl(var(--foreground)); }

/* Footer */
.sidebar-footer {
  display: flex;
  align-items: center;
  justify-content: space-between;
  height: 44px;
  padding: 0 1rem;
  border-top: 1px solid hsl(var(--border));
  background: hsl(var(--card));
  flex-shrink: 0;
}
.footer-label {
  font-size: .6875rem;
  font-weight: 700;
  letter-spacing: .05em;
  color: hsl(var(--primary));
  text-transform: uppercase;
}
.footer-expand {
  width: 28px; height: 28px;
  display: flex; align-items: center; justify-content: center;
  border: 1px solid hsl(var(--border)); background: transparent; cursor: pointer;
  border-radius: 999px; font-size: .75rem;
  color: hsl(var(--muted-foreground));
}
.footer-expand:hover { background: hsl(var(--muted)); }
</style>
