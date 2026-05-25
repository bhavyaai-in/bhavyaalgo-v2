<script setup>
import { ref, onMounted } from 'vue'
import { api, confirm } from '../../utils/api.js'

const activeTab = ref('types')
const types = ref([])
const settings = ref([])
const loading = ref(true)

// Strategy Types
const typeForm = ref({ name: '', rules_explanation: '' })
const editingType = ref(null)
const showTypeForm = ref(false)
const typeSaving = ref(false)

// System Settings
const settingForm = ref({ key: '', value: '' })
const showSettingForm = ref(false)
const settingSaving = ref(false)

async function fetchTypes() {
  try {
    const data = await api('/api/strategy-types')
    types.value = Array.isArray(data) ? data : []
  } catch { types.value = [] }
}

async function fetchSettings() {
  try {
    const data = await api('/api/settings')
    settings.value = Array.isArray(data) ? data : []
  } catch { settings.value = [] }
}

function openTypeForm(t) {
  editingType.value = t
  typeForm.value = t ? { name: t.name, rules_explanation: t.rules_explanation } : { name: '', rules_explanation: '' }
  showTypeForm.value = true
}

async function saveType() {
  typeSaving.value = true
  try {
    if (editingType.value) {
      await api('/api/strategy-types/' + editingType.value.id, { method: 'PUT', body: JSON.stringify(typeForm.value) })
    } else {
      await api('/api/strategy-types', { method: 'POST', body: JSON.stringify(typeForm.value) })
    }
    showTypeForm.value = false
    editingType.value = null
    await fetchTypes()
  } catch (e) { alert(e.message) }
  finally { typeSaving.value = false }
}

async function deleteType(t) {
  if (!await confirm('Delete', 'Delete "' + t.name + '"?')) return
  await api('/api/strategy-types/' + t.id, { method: 'DELETE' })
  await fetchTypes()
}

function openSettingForm(s) {
  settingForm.value = s ? { key: s.key, value: s.value } : { key: '', value: '' }
  showSettingForm.value = true
}

async function saveSetting() {
  settingSaving.value = true
  try {
    await api('/api/settings', { method: 'POST', body: JSON.stringify(settingForm.value) })
    showSettingForm.value = false
    await fetchSettings()
  } catch (e) { alert(e.message) }
  finally { settingSaving.value = false }
}

onMounted(async () => {
  await Promise.allSettled([fetchTypes(), fetchSettings()])
  loading.value = false
})
</script>

<template>
  <div class="page">
    <h2>Settings</h2>

    <!-- Sub-tabs -->
    <div class="sub-tabs">
      <button :class="{ active: activeTab === 'types' }" @click="activeTab = 'types'">Strategy Types</button>
      <button :class="{ active: activeTab === 'system' }" @click="activeTab = 'system'">System Settings</button>
    </div>

    <div v-if="loading" class="state-msg">Loading...</div>

    <!-- Strategy Types -->
    <template v-else-if="activeTab === 'types'">
      <div class="section-header">
        <span class="type-count">{{ types.length }} types</span>
        <button class="add-btn" @click="openTypeForm(null)">+ New Type</button>
      </div>

      <div v-if="!types.length" class="state-msg">No strategy types defined.</div>

      <div v-else class="type-list">
        <div v-for="t in types" :key="t.id" class="type-card">
          <div class="type-main">
            <strong>{{ t.name }}</strong>
            <p>{{ t.rules_explanation }}</p>
          </div>
          <div class="type-actions">
            <button class="chip" @click="openTypeForm(t)">Edit</button>
            <button class="chip danger" @click="deleteType(t)">Delete</button>
          </div>
        </div>
      </div>

      <!-- Type Form Modal -->
      <div v-if="showTypeForm" class="modal-overlay" @click.self="showTypeForm = false; editingType = null">
        <div class="modal-box small">
          <h3>{{ editingType ? 'Edit' : 'New' }} Strategy Type</h3>
          <div class="field-stack">
            <label>Name <input v-model="typeForm.name" /></label>
            <label>Rules Explanation <textarea v-model="typeForm.rules_explanation" rows="3"></textarea></label>
          </div>
          <div class="form-actions">
            <button @click="saveType" :disabled="typeSaving || !typeForm.name">{{ typeSaving ? 'Saving...' : 'Save' }}</button>
            <button class="cancel" @click="showTypeForm = false; editingType = null">Cancel</button>
          </div>
        </div>
      </div>
    </template>

    <!-- System Settings -->
    <template v-else-if="activeTab === 'system'">
      <div class="section-header">
        <button class="add-btn" @click="openSettingForm(null)">+ New Setting</button>
      </div>

      <table v-if="settings.length" class="data-table">
        <thead><tr><th>Key</th><th>Value</th><th></th></tr></thead>
        <tbody>
          <tr v-for="s in settings" :key="s.key">
            <td class="pkey">{{ s.key }}</td>
            <td>{{ s.value }}</td>
            <td><button class="chip" @click="openSettingForm(s)">Edit</button></td>
          </tr>
        </tbody>
      </table>
      <div v-else class="state-msg">No settings defined.</div>

      <!-- Setting Form Modal -->
      <div v-if="showSettingForm" class="modal-overlay" @click.self="showSettingForm = false">
        <div class="modal-box small">
          <h3>{{ settingForm.key && !editingType ? 'Edit' : 'New' }} Setting</h3>
          <div class="field-stack">
            <label>Key <input v-model="settingForm.key" :disabled="!!settingForm.key" /></label>
            <label>Value <input v-model="settingForm.value" /></label>
          </div>
          <div class="form-actions">
            <button @click="saveSetting" :disabled="settingSaving || !settingForm.key">{{ settingSaving ? 'Saving...' : 'Save' }}</button>
            <button class="cancel" @click="showSettingForm = false">Cancel</button>
          </div>
        </div>
      </div>
    </template>
  </div>
</template>

<style scoped>
.page { padding:0; }
h2 { margin:0 0 1rem; }

.sub-tabs { display:flex; gap:0; border-bottom:1px solid hsl(var(--border)); margin-bottom:1rem; }
.sub-tabs button {
  padding:.5rem 1rem; border:none; background:transparent;
  font-size:var(--font-sm); color:hsl(var(--muted-foreground)); cursor:pointer;
  border-bottom:2px solid transparent; font-weight:500;
}
.sub-tabs button.active { color:hsl(var(--primary)); border-bottom-color:hsl(var(--primary)); }
.sub-tabs button:hover { color:hsl(var(--foreground)); }

.section-header { display:flex; justify-content:space-between; align-items:center; margin-bottom:.75rem; }
.type-count { font-size:var(--font-sm); color:hsl(var(--muted-foreground)); }
.add-btn {
  padding:.4rem .8rem; border:none; border-radius:var(--radius);
  background:hsl(var(--primary)); color:#fff; font-weight:500; cursor:pointer; font-size:var(--font-xs);
}

.type-list { display:flex; flex-direction:column; gap:.5rem; }
.type-card {
  display:flex; justify-content:space-between; align-items:center;
  background:hsl(var(--card)); border:1px solid hsl(var(--border));
  border-radius:var(--radius); padding:.6rem .75rem;
}
.type-main strong { font-size:var(--font-sm); color:hsl(var(--foreground)); }
.type-main p { margin:.15rem 0 0; font-size:var(--font-xs); color:hsl(var(--muted-foreground)); }
.type-actions { display:flex; gap:.4rem; flex-shrink:0; }

.chip.danger:hover { border-color:hsl(var(--destructive)); color:hsl(var(--destructive)); }







.state-msg { text-align:center; padding:2rem; color:hsl(var(--muted-foreground)); font-size:var(--font-sm); }

/* Modal */
.modal-overlay {
  position:fixed; inset:0; background:rgba(0,0,0,.4);
  display:flex; justify-content:center; align-items:center; z-index:100;
}
.modal-box.small { background:hsl(var(--card)); border-radius:var(--radius); padding:1.25rem; width:90%; max-width:420px; }
.modal-box h3 { margin:0 0 .75rem; font-size:var(--font-base); }
.field-stack { display:flex; flex-direction:column; gap:.6rem; }
.field-stack label { display:flex; flex-direction:column; gap:.2rem; font-size:var(--font-sm); color:hsl(var(--foreground)); }
.field-stack input, .field-stack textarea { padding:.45rem .6rem; border:1px solid hsl(var(--input)); border-radius:var(--radius); font-size:var(--font-sm); outline:none; background:hsl(var(--card)); }
.field-stack input:focus, .field-stack textarea:focus { border-color:hsl(var(--ring)); box-shadow:0 0 0 2px hsl(var(--ring)/.2); }
.field-stack input:disabled { background:hsl(var(--muted)); }
.form-actions { display:flex; gap:.5rem; margin-top:.75rem; }
.form-actions button { flex:1; padding:.5rem; border:none; border-radius:var(--radius); cursor:pointer; font-weight:500; color:#fff; background:hsl(var(--primary)); font-size:var(--font-sm); }
.form-actions button:disabled { opacity:.5; }
.form-actions .cancel { background:hsl(var(--muted-foreground)); }
</style>
