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
const menuOpenId = ref<string | null>(null)

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
  menuOpenId.value = null
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
  menuOpenId.value = null
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

function toggleMenu(groupId: string) {
  menuOpenId.value = menuOpenId.value === groupId ? null : groupId
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
            <div class="group-menu">
              <button
                type="button"
                class="icon-action"
                :title="t('settings.groupsMore')"
                :aria-expanded="menuOpenId === group.id"
                :disabled="busy"
                @click="toggleMenu(group.id)"
              >
                <Icon icon="mdi:dots-horizontal" width="18" />
              </button>
              <div v-if="menuOpenId === group.id" class="group-menu-popover">
                <button type="button" @click="startEdit(group)">
                  <Icon icon="mdi:pencil-outline" width="15" />
                  {{ t('home.edit') }}
                </button>
                <button type="button" class="danger" @click="removeGroup(group)">
                  <Icon icon="mdi:trash-can-outline" width="15" />
                  {{ t('home.delete') }}
                </button>
              </div>
            </div>
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
  border-radius: 12px;
  padding: 22px 26px;
  margin-bottom: 20px;
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
  font-size: 18px;
  font-weight: 600;
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
  min-height: 58px;
  padding: 8px 0;
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
  color: #87909c;
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
  height: 36px;
  padding: 0 14px;
  font-size: 13px;
  cursor: pointer;
}

.primary-btn:disabled,
.link-btn:disabled {
  opacity: 0.45;
  cursor: not-allowed;
}

.primary-btn:hover:not(:disabled) {
  background: var(--sv-accent-hover);
}

.icon-action:disabled {
  opacity: 1;
  color: var(--sv-disabled);
  cursor: not-allowed;
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
  width: 34px;
  height: 34px;
  display: grid;
  place-items: center;
  border: 1px solid transparent;
  border-radius: 8px;
  background: transparent;
  color: var(--sv-mute);
  cursor: pointer;
  padding: 0;
}

.icon-action:hover:not(:disabled) {
  color: var(--sv-text);
  background: var(--sv-surface-hover);
}

.icon-action.danger {
  color: var(--sv-mute);
}

.icon-action.danger:hover:not(:disabled) {
  color: var(--sv-danger);
  background: var(--sv-danger-soft);
}

.group-menu {
  position: relative;
}

.group-menu-popover {
  position: absolute;
  top: calc(100% + 6px);
  right: 0;
  z-index: 20;
  min-width: 112px;
  padding: 5px;
  border: 1px solid var(--sv-border);
  border-radius: 9px;
  background: var(--sv-bg);
  box-shadow: 0 12px 30px rgba(15, 23, 42, 0.14);
}

.group-menu-popover button {
  width: 100%;
  min-height: 32px;
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 0 9px;
  border: 0;
  border-radius: 6px;
  background: transparent;
  color: var(--sv-text);
  font-size: 12px;
  cursor: pointer;
}

.group-menu-popover button:hover {
  background: var(--sv-surface-hover);
}

.group-menu-popover button.danger {
  color: var(--sv-mute);
}

.group-menu-popover button.danger:hover {
  color: var(--sv-danger);
  background: var(--sv-danger-soft);
}

:global(html[data-theme='dark']) .meta {
  color: rgba(255, 255, 255, 0.52);
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
