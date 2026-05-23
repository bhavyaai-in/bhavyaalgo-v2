<script setup>
import { useRouter } from 'vue-router'

const router = useRouter()
const apps = [
  { name: 'Data Downloader', desc: 'Download master contracts and market data for offline analysis', icon: '⬇️', color: '#3B82F6' },
  { name: 'Option Chain', desc: 'View real-time option chain with CE/PE data, OI, and live prices', icon: '📊', color: '#8B5CF6', route: '/optionchain' },
  { name: 'Scalper', desc: 'Quick scalp trades with hotkeys and preset configurations', icon: '⚡', color: '#F59E0B' },
  { name: 'Backtest Engine', desc: 'Run historical backtests on your strategies', icon: '🧪', color: '#10B981' },
  { name: 'Alerts', desc: 'Set price and indicator alerts across your watchlist', icon: '🔔', color: '#EF4444' },
  { name: 'Market Scanner', desc: 'Scan multiple securities for technical patterns', icon: '🔍', color: '#EC4899' },
  { name: 'Trade Journal', desc: 'Log and analyze your trades with notes and screenshots', icon: '📓', color: '#14B8A6' },
  { name: 'Paper Trading', desc: 'Test strategies with virtual money in real-time market', icon: '📝', color: '#6366F1' },
]
</script>

<template>
  <div class="apps-page">
    <header><h2>Apps</h2></header>
    <p class="subtitle">Launch trading tools and utilities</p>
    <div class="tile-grid">
      <div v-for="app in apps" :key="app.name" class="tile" @click="app.route ? router.push(app.route) : alert(app.name + ' coming soon')">
        <div class="tile-icon" :style="{ background: app.color + '20', color: app.color }">{{ app.icon }}</div>
        <div class="tile-info">
          <div class="tile-name">{{ app.name }}</div>
          <div class="tile-desc">{{ app.desc }}</div>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.apps-page { }
header { margin-bottom:.25rem; }
header h2 { margin:0; }
.subtitle { font-size:var(--font-sm); color:hsl(var(--muted-foreground)); margin-bottom:1.25rem; }
.tile-grid { display:grid; grid-template-columns:repeat(auto-fill,minmax(280px,1fr)); gap:.75rem; }
.tile {
  display:flex; align-items:center; gap:.85rem;
  background:hsl(var(--card)); border:1px solid hsl(var(--border));
  border-radius:var(--radius); padding:.85rem 1rem; cursor:pointer;
  transition:border-color .15s, box-shadow .15s;
}
.tile:hover { border-color:hsl(var(--primary)); box-shadow:0 2px 8px hsl(var(--primary)/.1); }
.tile-icon {
  width:44px; height:44px; border-radius:10px; display:flex; align-items:center;
  justify-content:center; font-size:1.3rem; flex-shrink:0;
}
.tile-info { display:flex; flex-direction:column; gap:2px; min-width:0; }
.tile-name { font-size:var(--font-sm); font-weight:700; color:hsl(var(--foreground)); }
.tile-desc { font-size:var(--font-xs); color:hsl(var(--muted-foreground)); line-height:1.35; }
</style>
