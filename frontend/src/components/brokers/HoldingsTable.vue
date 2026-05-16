<script setup>
const props = defineProps({
  data: { type: Object, default: null }
})

function pnl(h) {
  return (((h.ltp||0)-(h.averageprice||0))*(h.quantity||0)).toFixed(2)
}

const summaryCards = [
  { label: 'Total Holding', key: 'totalholdingvalue' },
  { label: 'Total Investment', key: 'totalinvvalue' },
  { label: 'Total P&L', key: 'totalprofitandloss' },
  { label: 'P&L %', key: 'totalpnlpercentage' },
]
</script>

<template>
  <div>
    <div v-if="data?.data?.totalholding" class="summary-row">
      <div v-for="c in summaryCards" :key="c.key" class="summary-card" :class="{negative: Number(data.data.totalholding[c.key])<0}">
        <div class="summary-label">{{ c.label }}</div>
        <div class="summary-value">{{ data.data.totalholding[c.key] ?? '-' }}</div>
      </div>
    </div>
    <div v-if="data?.data?.holdings?.length" class="table-wrap">
      <table class="data-table">
        <thead><tr><th>Trading Symbol</th><th>Exchange</th><th>Qty</th><th>Buy Price</th><th>LTP</th><th>Product</th><th>P&L</th><th>P&L %</th></tr></thead>
        <tbody>
          <tr v-for="h in data.data.holdings" :key="h.symboltoken">
            <td><strong>{{ h.tradingsymbol }}</strong></td>
            <td>{{ h.exchange }}</td><td>{{ h.quantity }}</td>
            <td>{{ h.averageprice }}</td><td>{{ h.ltp }}</td><td>{{ h.product }}</td>
            <td :class="{negative: Number(pnl(h)) < 0}">{{ pnl(h) }}</td>
            <td :class="{negative: Number(h.pnlpercentage) < 0}">{{ h.pnlpercentage }}%</td>
          </tr>
        </tbody>
      </table>
    </div>
    <div v-else class="state-msg">No holdings.</div>
  </div>
</template>

<style scoped>
.summary-row { display:flex; gap:.5rem; margin-bottom:1rem; flex-wrap:wrap; }
.summary-card { flex:1; min-width:100px; background:hsl(var(--card)); border:1px solid hsl(var(--border)); border-radius:var(--radius); padding:.6rem; text-align:center; }
.summary-label { font-size:var(--font-xs); color:hsl(var(--muted-foreground)); }
.summary-value { font-size:var(--font-base); font-weight:700; color:hsl(var(--foreground)); }
.summary-card.negative .summary-value { color:hsl(var(--destructive)); }
.table-wrap { overflow-x:auto; }
.data-table { width:100%; border-collapse:collapse; font-size:var(--font-sm); }
.data-table th, .data-table td { padding:.4rem .5rem; border-bottom:1px solid hsl(var(--border)); text-align:left; white-space:nowrap; }
.data-table th { font-weight:600; color:hsl(var(--foreground)); position:sticky; top:0; background:hsl(var(--card)); }
.data-table td { color:hsl(var(--muted-foreground)); }
.data-table td.negative { color:hsl(var(--destructive)); }
.state-msg { text-align:center; padding:2rem; color:hsl(var(--muted-foreground)); }
</style>
