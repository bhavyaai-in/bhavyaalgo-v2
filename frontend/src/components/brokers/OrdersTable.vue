<script setup>
const props = defineProps({ items: { type: Array, default: () => [] }, canAct: { type: Function, default: () => () => false } })
const emit = defineEmits(['cancel','edit'])

function handleCancel(o) { emit('cancel', o) }
function handleEdit(o) { emit('edit', o) }

function orderTime(o) {
  const raw = o.ordertime || o.requesttime || o.timestamp || o.created_at || o.createdAt || o.orderDate || o.orderDateTime || o.dateTime
  if (!raw) return '-'
  const asNumber = Number(raw)
  if (!Number.isNaN(asNumber) && String(raw).match(/^\d+$/)) {
    const millis = String(raw).length === 10 ? asNumber * 1000 : asNumber
    const parsed = new Date(millis)
    if (!Number.isNaN(parsed.getTime())) {
      return parsed.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit', second: '2-digit' })
    }
  }
  const parsed = new Date(raw)
  if (!Number.isNaN(parsed.getTime())) {
    return parsed.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit', second: '2-digit' })
  }
  return String(raw)
}

function statusClass(o) {
  const s = (o.orderstatus || '').toLowerCase()
  if (s === 'rejected' || s === 'cancel') return 'rejected'
  if (s === 'complete' || s === 'filled') return 'complete'
  if (s === 'open' || s === 'pending' || s === 'trigger pending') return 'open'
  return ''
}
</script>

<template>
  <div v-if="items && items.length" class="table-wrap">
    <table class="data-table">
      <thead>
        <tr>
          <th>Trading Symbol</th><th>Time</th><th>Type</th><th>Qty</th><th>Filled</th><th>Price</th><th>Order Type</th><th>Product</th><th>Status</th><th>Actions</th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="o in items" :key="o.orderid">
          <td><strong>{{ o.tradingsymbol }}</strong></td>
          <td>{{ o.orderTime }}</td><td>{{ o.transactiontype }}</td>
          <td>{{ o.quantity }}</td><td>{{ o.filledshares }}</td><td>{{ o.price }}</td>
          <td>{{ o.ordertype }}</td><td>{{ o.producttype }}</td>
          <td>
            <span class="status-badge" :class="statusClass(o)">{{ o.orderstatus }}</span>
            <span v-if="o.text" class="status-tooltip">{{ o.text }}</span>
          </td>
          <td>
            <div class="actions" v-if="canAct(o)">
              <button class="action-btn" title="Edit" @click="handleEdit(o)"><svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M11 4H4a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2v-7"/><path d="M18.5 2.5a2.121 2.121 0 0 1 3 3L12 15l-4 1 1-4 9.5-9.5z"/></svg></button>
              <button class="action-btn cancel" title="Cancel" @click="handleCancel(o)"><svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/></svg></button>
            </div>
          </td>
        </tr>
      </tbody>
    </table>
  </div>
  <div v-else class="state-msg">No orders.</div>
</template>

<style scoped>
.table-wrap { overflow-x: auto; }
.data-table { width: 100%; border-collapse: collapse; font-size: var(--font-sm); }
.data-table th, .data-table td { padding: .5rem .6rem; border-bottom: 1px solid hsl(var(--border)); text-align: left; white-space: nowrap; }
.data-table th { font-weight: 600; color: hsl(var(--foreground)); position: sticky; top: 0; background: hsl(var(--card)); }
.data-table td { color: hsl(var(--muted-foreground)); }
.status-badge { display: inline-block; padding: .15rem .5rem; border-radius: 999px; font-size: var(--font-xs); font-weight: 600; cursor: default; }
.status-badge.open { background: hsl(48 100% 50% / .15); color: #b8860b; }
.status-badge.complete { background: hsl(144 80% 55% / .15); color: #16A34A; }
.status-badge.rejected { background: hsl(0 84% 60% / .15); color: hsl(var(--destructive)); }
.status-tooltip {
  display: none;
  position: absolute;
  background: hsl(var(--foreground));
  color: hsl(var(--card));
  font-size: var(--font-xs);
  padding: .4rem .6rem;
  border-radius: var(--radius);
  white-space: nowrap;
  max-width: 320px;
  z-index: 10;
  bottom: calc(100% + 4px);
  left: 50%;
  transform: translateX(-50%);
  box-shadow: 0 2px 8px rgba(0,0,0,.15);
}
td { position: relative; }
td:hover .status-tooltip { display: block; }
.actions { display: flex; gap: .3rem; }
.action-btn { display: inline-flex; align-items: center; justify-content: center; width: 26px; height: 26px; border: 1px solid hsl(var(--border)); border-radius: var(--radius); background: transparent; cursor: pointer; color: hsl(var(--muted-foreground)); }
.action-btn.cancel { border-color: hsl(var(--destructive)); color: hsl(var(--destructive)); }
.action-btn:hover { opacity: .7; }
.state-msg { text-align: center; padding: 2rem; color: hsl(var(--muted-foreground)); }
</style>
