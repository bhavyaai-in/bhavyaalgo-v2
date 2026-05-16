<script setup>
defineProps({ title: String, data: null })

function cap(str) {
  if (!str) return ''
  return str.replace(/\b\w/g, c => c.toUpperCase())
}

function fmt(v) {
  if (v == null || v === '') return '-'
  return Number(v).toFixed(2)
}

function flatRows(obj) {
  if (!obj) return []
  function flat(o) {
    const rows = []
    for (const [k, v] of Object.entries(o || {})) {
      if (k.startsWith('_')) continue
      if (v && typeof v === 'object' && !Array.isArray(v)) { rows.push(...flat(v)) }
      else if (v !== '') { rows.push({ key: cap(k.replace(/_/g, ' ')), value: Array.isArray(v) ? v.join(', ') : String(v ?? '') }) }
    }
    return rows
  }
  if (obj.data && typeof obj.data === 'object' && Object.keys(obj.data).length) return flat(obj.data)
  return flat(obj)
}
</script>

<template>
  <!-- Orders table (no actions — just view) -->
  <div v-if="title === 'Orders' && Array.isArray(data) && data.length" class="table-wrap">
    <table class="data-table">
      <thead><tr><th>Trading Symbol</th><th>Exchange</th><th>Type</th><th>Qty</th><th>Filled</th><th>Price</th><th>Order Type</th><th>Product</th><th>Status</th></tr></thead>
      <tbody>
        <tr v-for="o in data" :key="o.orderid">
          <td><strong>{{ o.tradingsymbol }}</strong></td>
          <td>{{ o.exchange }}</td><td>{{ o.transactiontype }}</td>
          <td>{{ o.quantity }}</td><td>{{ o.filledshares }}</td><td>{{ o.price }}</td>
          <td>{{ o.ordertype }}</td><td>{{ o.producttype }}</td>
          <td><span class="status-badge" :class="o.orderstatus">{{ o.orderstatus }}</span></td>
        </tr>
      </tbody>
    </table>
  </div>

  <!-- Holdings: summary + table -->
  <div v-else-if="title === 'Holdings' && data?.data" class="compact">
    <div class="summary-row">
      <div v-for="f in [{k:'totalholdingvalue',l:'Total Holding'},{k:'totalinvvalue',l:'Total Investment'},{k:'totalprofitandloss',l:'Total P&L'},{k:'totalpnlpercentage',l:'P&L %'}]" :key="f.k" class="sc" :class="{negative: Number(data.data.totalholding?.[f.k])<0}">
        <div class="sl">{{ f.l }}</div>
        <div class="sv">{{ data.data.totalholding?.[f.k] ?? '-' }}</div>
      </div>
    </div>
    <div v-if="data.data.holdings?.length" class="table-wrap">
      <table class="data-table">
        <thead><tr><th>Trading Symbol</th><th>Exchange</th><th>Qty</th><th>Buy Price</th><th>LTP</th><th>Product</th><th>P&L</th><th>P&L %</th></tr></thead>
        <tbody>
          <tr v-for="h in data.data.holdings" :key="h.symboltoken">
            <td><strong>{{ h.tradingsymbol }}</strong></td>
            <td>{{ h.exchange }}</td><td>{{ h.quantity }}</td>
            <td>{{ h.averageprice }}</td><td>{{ h.ltp }}</td><td>{{ h.product }}</td>
            <td :class="{negative: ((h.ltp||0)-(h.averageprice||0))*(h.quantity||0)<0}">{{ (((h.ltp||0)-(h.averageprice||0))*(h.quantity||0)).toFixed(2) }}</td>
            <td :class="{negative: Number(h.pnlpercentage)<0}">{{ h.pnlpercentage }}%</td>
          </tr>
        </tbody>
      </table>
    </div>
  </div>

  <!-- Positions table -->
  <div v-else-if="title === 'Positions' && Array.isArray(data) && data.length" class="table-wrap">
    <table class="data-table">
      <thead><tr><th>Trading Symbol</th><th>Exchange</th><th>Qty</th><th>Product</th><th>P&L</th></tr></thead>
      <tbody>
        <tr v-for="p in data" :key="p.symboltoken || p.tradingsymbol">
          <td><strong>{{ p.tradingsymbol }}</strong></td>
          <td>{{ p.exchange }}</td><td>{{ p.quantity || p.netqty }}</td>
          <td>{{ p.producttype || p.product }}</td>
          <td :class="{negative: Number(p.profitandloss||p.pnl||0)<0}">{{ fmt(p.profitandloss || p.pnl || 0) }}</td>
        </tr>
      </tbody>
    </table>
  </div>

  <!-- Margin key-value -->
  <div v-else-if="title === 'Margin'" class="table-wrap">
    <table class="data-table kv">
      <tr v-for="(v,k) in data" :key="k"><td class="pkey">{{ cap(k.replace(/_/g,' ')) }}</td><td>{{ v }}</td></tr>
    </table>
  </div>

  <!-- Profile / fallback flat rows -->
  <div v-else class="table-wrap">
    <table class="data-table kv">
      <tr v-for="row in flatRows(data)" :key="row.key"><td class="pkey">{{ row.key }}</td><td>{{ row.value }}</td></tr>
    </table>
  </div>
</template>

<style scoped>
.compact :deep(.summary-row),
.summary-row { display:flex; gap:.4rem; margin-bottom:.75rem; flex-wrap:wrap; }
.sc { flex:1; min-width:80px; background:hsl(var(--card)); border:1px solid hsl(var(--border)); border-radius:var(--radius); padding:.4rem; text-align:center; }
.sl { font-size:var(--font-xs); color:hsl(var(--muted-foreground)); }
.sv { font-size:var(--font-sm); font-weight:700; color:hsl(var(--foreground)); }
.sc.negative .sv { color:hsl(var(--destructive)); }
.table-wrap { overflow-x:auto; }
.data-table { width:100%; border-collapse:collapse; font-size:var(--font-sm); }
.data-table th, .data-table td { padding:.35rem .4rem; border-bottom:1px solid hsl(var(--border)); text-align:left; white-space:nowrap; }
.data-table th { font-weight:600; color:hsl(var(--foreground)); background:hsl(var(--card)); }
.data-table td { color:hsl(var(--muted-foreground)); }
.data-table td.negative { color:hsl(var(--destructive)); }
.data-table.kv td { padding:.5rem .75rem; }
.data-table .pkey { font-weight:600; color:hsl(var(--foreground)); width:40%; }
.status-badge { display:inline-block; padding:.1rem .35rem; border-radius:999px; font-size:var(--font-xs); font-weight:600; }
.status-badge.open { background:hsl(48 100% 50% / .15); color:#b8860b; }
.status-badge.complete { background:hsl(144 80% 55% / .15); color:#16A34A; }
.status-badge.cancelled { background:hsl(0 84% 60% / .15); color:hsl(var(--destructive)); }
</style>
