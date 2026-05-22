<script setup>
const props = defineProps({ items: { type: Array, default: () => [] } })
function fmt(v) { return v == null || v === '' ? '-' : Number(v).toFixed(2) }
</script>

<template>
  <div v-if="items.length" class="table-wrap">
    <table class="data-table">
      <thead><tr><th>Trading Symbol</th><th>Exchange</th><th>Qty</th><th>Product</th><th>P&L</th></tr></thead>
      <tbody>
        <tr v-for="p in items" :key="p.symboltoken || p.tradingsymbol">
          <td><strong>{{ p.tradingsymbol }}</strong></td>
          <td>{{ p.exchange }}</td><td>{{ p.quantity || p.netqty }}</td>
          <td>{{ p.producttype || p.product }}</td>
          <td :class="{negative: Number(p.profitandloss||p.pnl||0)<0, positive: Number(p.profitandloss||p.pnl||0)>0}">{{ fmt(p.profitandloss || p.pnl || 0) }}</td>
        </tr>
      </tbody>
    </table>
  </div>
  <div v-else class="state-msg">No positions.</div>
</template>

<style scoped>
.table-wrap { overflow-x:auto; }
.data-table { width:100%; border-collapse:collapse; font-size:var(--font-sm); }
.data-table th, .data-table td { padding:.4rem .5rem; border-bottom:1px solid hsl(var(--border)); text-align:left; white-space:nowrap; }
.data-table th { font-weight:600; color:hsl(var(--foreground)); position:sticky; top:0; background:hsl(var(--card)); }
.data-table td { color:hsl(var(--muted-foreground)); }
.data-table td.negative { color:hsl(var(--destructive)); }
.data-table td.positive { color:#16A34A; }
.state-msg { text-align:center; padding:2rem; color:hsl(var(--muted-foreground)); }
</style>
