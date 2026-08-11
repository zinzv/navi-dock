<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { Icon } from '@iconify/vue'
import {
  createGroup,
  deleteGroup,
  fetchGroups,
  sortGroups,
  updateGroup,
  type NavGroup,
} from '../api/settings'

const emit = defineEmits<{
  saved: []
}>()

const { t } = useI18n()
const groups = ref<NavGroup[]>([])
const loading = ref(false)
const busy = ref(false)
const error = ref('')
const editingId = ref<string | null>(null)
const draftName = ref('')
const adding = ref(false)
const newName = ref('')

async function load() {
  loading.value = true
  error.value = ''
  try {
    const data = await fetchGroups()
    groups.value = data.groups
  } catch (e) {
    error.value = e instanceof Error ? e.message : t('settings.groupsLoadFailed')
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  void load()
})

function startEdit(group: NavGroup) {
  adding.value = false
  editingId.value = group.id
  draftName.value = group.name
}

function cancelEdit() {
  editingId.value = null
  draftName.value = ''
}

async function saveEdit(group: NavGroup) {
  const name = draftName.value.trim()
  if (!name || busy.value) return
  busy.value = true
  error.value = ''
  try {
    const updated = await updateGroup(group.id, { name, icon: group.icon || '' })
    const idx = groups.value.findIndex((g) => g.id === group.id)
    if (idx >= 0) {
      groups.value[idx] = { ...groups.value[idx], ...updated }
    }
    cancelEdit()
    emit('saved')
  } catch (e) {
    error.value = e instanceof Error ? e.message : t('settings.groupsSaveFailed')
  } finally {
    busy.value = false
  }
}

function startAdd() {
  editingId.value = null
  adding.value = true
  newName.value = ''
}

function cancelAdd() {
  adding.value = false
  newName.value = ''
}

async function confirmAdd() {
  const name = newName.value.trim()
  if (!name || busy.value) return
  busy.value = true
  error.value = ''
  try {
    const created = await createGroup({ name })
    groups.value = [...groups.value, { ...created, item_count: 0 }]
    cancelAdd()
    emit('saved')
  } catch (e) {
    error.value = e instanceof Error ? e.message : t('settings.groupsSaveFailed')
  } finally {
    busy.value = false
  }
}

async function removeGroup(group: NavGroup) {
  const count = group.item_count ?? group.items?.length ?? 0
  const msg =
    count > 0
      ? t('settings.groupsDeleteConfirmWithItems', { name: group.name, count })
      : t('settings.groupsDeleteConfirm', { name: group.name })
  if (!window.confirm(msg)) return
  busy.value = true
  error.value = ''
  try {
    await deleteGroup(group.id)
    groups.value = groups.value.filter((g) => g.id !== group.id)
    if (editingId.value === group.id) cancelEdit()
    emit('saved')
  } catch (e) {
    error.value = e instanceof Error ? e.message : t('settings.groupsSaveFailed')
  } finally {
    busy.value = false
  }
}

async function move(index: number, delta: number) {
  const next = index + delta
  if (next < 0 || next >= groups.value.length || busy.value) return
  const reordered = [...groups.value]
  const [row] = reordered.splice(index, 1)
  reordered.splice(next, 0, row)
  groups.value = reordered
  busy.value = true
  error.value = ''
  try {
    const data = await sortGroups(reordered.map((g) => g.id))
    groups.value = data.groups
    emit('saved')
  } catch (e) {
    error.value = e instanceof Error ? e.message : t('settings.groupsSaveFailed')
    await load()
  } finally {
    busy.value = false
  }
}
</script>

<template>
  <section class="card">
    <div class="card-header">
      <h2 class="card-title">{{ t('settings.groups') }}</h2>
      <button type="button" class="primary-btn" :disabled="busy || adding" @click="startAdd">
        {{ t('settings.groupsAdd') }}
      </button>
    </div>

    <p v-if="loading" class="status">{{ t('common.loading') }}</p>
    <p v-else-if="error" class="status error">{{ error }}</p>

    <div v-else class="rows">
      <div v-if="adding" class="row row-stack">
        <input
          v-model="newName"
          class="row-input"
          type="text"
          maxlength="64"
          :placeholder="t('settings.groupsNamePlaceholder')"
          @keydown.enter.prevent="confirmAdd"
          @keydown.esc.prevent="cancelAdd"
        />
        <div class="row-actions">
          <button type="button" class="primary-btn" :disabled="busy || !newName.trim()" @click="confirmAdd">
            {{ t('settings.groupsSave') }}
          </button>
          <button type="button" class="link-btn" :disabled="busy" @click="cancelAdd">
            {{ t('settings.groupsCancel') }}
          </button>
        </div>
      </div>

      <div v-for="(group, index) in groups" :key="group.id" class="row group-row">
        <div class="group-main">
          <template v-if="editingId === group.id">
            <input
              v-model="draftName"
              class="row-input"
              type="text"
              maxlength="64"
              @keydown.enter.prevent="saveEdit(group)"
              @keydown.esc.prevent="cancelEdit"
            />
          </template>
          <template v-else>
            <span class="label">{{ group.name }}</span>
            <span class="meta">{{ t('settings.groupsItemCount', { count: group.item_count ?? 0 }) }}</span>
          </template>
        </div>

        <div class="row-actions">
          <template v-if="editingId === group.id">
            <button type="button" class="primary-btn" :disabled="busy || !draftName.trim()" @click="saveEdit(group)">
              {{ t('settings.groupsSave') }}
            </button>
            <button type="button" class="link-btn" :disabled="busy" @click="cancelEdit">
              {{ t('settings.groupsCancel') }}
            </button>
          </template>
          <template v-else>
            <button
              type="button"
              class="icon-action"
              :title="t('settings.groupsMoveUp')"
              :disabled="busy || index === 0"
              @click="move(index, -1)"
            >
              <Icon icon="mdi:arrow-up" width="16" />
            </button>
            <button
              type="button"
              class="icon-action"
              :title="t('settings.groupsMoveDown')"
              :disabled="busy || index === groups.length - 1"
              @click="move(index, 1)"
            >
              <Icon icon="mdi:arrow-down" width="16" />
            </button>
            <button
              type="button"
              class="icon-action"
              :title="t('home.edit')"
              :disabled="busy"
              @click="startEdit(group)"
            >
              <Icon icon="mdi:pencil-outline" width="16" />
            </button>
            <button
              type="button"
              class="icon-action danger"
              :title="t('home.delete')"
              :disabled="busy"
              @click="removeGroup(group)"
            >
              <Icon icon="mdi:trash-can-outline" width="16" />
            </button>
          </template>
        </div>
      </div>

      <div v-if="!groups.length && !adding" class="row">
        <span class="value">{{ t('settings.groupsEmpty') }}</span>
      </div>
    </div>
  </section>
</template>

<style scoped>
.card {
  background: var(--sv-surface);
  border: 1px solid var(--sv-border);
  border-radius: 10px;
  padding: 16px 18px 4px;
  margin-bottom: 16px;
}

.card-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  margin-bottom: 8px;
}

.card-header .card-hint {
  margin: 0;
  flex: 1;
  min-width: 0;
}

.card-header .primary-btn {
  flex-shrink: 0;
}

.card-title {
  margin: 0;
  font-size: 15px;
  font-weight: 700;
  color: var(--sv-text);
}

.card-hint {
  margin: 0 0 8px;
  font-size: 12px;
  line-height: 1.5;
  color: var(--sv-mute);
}

.status {
  margin: 8px 0 12px;
  font-size: 13px;
  color: var(--sv-mute);
}

.status.error {
  color: #ef4444;
}

.rows {
  display: flex;
  flex-direction: column;
}

.row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  min-height: 44px;
  padding: 10px 0;
  font-size: 14px;
  border-top: 1px solid var(--sv-border);
}

.rows .row:first-child {
  border-top: none;
}

.row-stack {
  flex-direction: column;
  align-items: stretch;
  gap: 10px;
}

.group-row {
  align-items: center;
}

.group-main {
  flex: 1;
  min-width: 0;
  display: flex;
  align-items: center;
  gap: 10px;
}

.label {
  color: var(--sv-text);
  font-weight: 500;
}

.meta {
  font-size: 12px;
  color: var(--sv-mute);
  white-space: nowrap;
}

.value {
  color: var(--sv-mute);
  font-size: 13px;
}

.row-input {
  height: 36px;
  width: 100%;
  max-width: 280px;
  border-radius: 8px;
  border: 1px solid var(--sv-border);
  background: var(--sv-bg);
  color: var(--sv-text);
  padding: 0 12px;
  outline: none;
  font-size: 14px;
}

.row-input:focus {
  border-color: var(--sv-accent);
}

.row-actions {
  display: flex;
  align-items: center;
  gap: 6px;
  flex-shrink: 0;
}

.primary-btn {
  background: var(--sv-accent);
  color: #fff;
  border: none;
  border-radius: 8px;
  padding: 6px 14px;
  font-size: 13px;
  cursor: pointer;
}

.primary-btn:disabled,
.link-btn:disabled,
.icon-action:disabled {
  opacity: 0.45;
  cursor: default;
}

.link-btn {
  background: none;
  border: 1px solid var(--sv-border);
  color: var(--sv-text);
  border-radius: 8px;
  padding: 4px 10px;
  font-size: 12px;
  cursor: pointer;
}

.icon-action {
  width: 30px;
  height: 30px;
  display: grid;
  place-items: center;
  border: 1px solid var(--sv-border);
  border-radius: 8px;
  background: transparent;
  color: var(--sv-text);
  cursor: pointer;
  padding: 0;
}

.icon-action:hover:not(:disabled) {
  border-color: var(--sv-accent);
}

.icon-action.danger {
  color: #dc2626;
}

.icon-action.danger:hover:not(:disabled) {
  background: #fef2f2;
  border-color: #fecaca;
}

@media (max-width: 640px) {
  .group-row {
    flex-wrap: wrap;
  }

  .row-actions {
    width: 100%;
    justify-content: flex-end;
  }
}
</style>
