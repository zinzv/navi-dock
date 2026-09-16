<script setup lang="ts">
import { computed, inject, onMounted, ref, type Ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { Icon } from '@iconify/vue'
import {
  deleteItem,
  fetchNavigation,
  sortItems,
  type NavGroup,
  type NavItem,
  type NetworkMode,
} from '../api/settings'
import { useSettingsStore } from '../stores/settings'
import { useItemReorder } from '../composables/useItemReorder'
import AppContextMenu from '../components/AppContextMenu.vue'
import EditItemModal from '../components/EditItemModal.vue'

const { t } = useI18n()
const settings = useSettingsStore()
const searchQuery = inject<Ref<string>>('homeSearchQuery', ref(''))
const groups = ref<NavGroup[]>([])
const toast = ref('')
let toastTimer: number | undefined

const menuOpen = ref(false)
const menuX = ref(0)
const menuY = ref(0)
const menuItem = ref<NavItem | null>(null)

const editOpen = ref(false)
const editItem = ref<NavItem | null>(null)
const createGroupId = ref('')

const filteredGroups = computed(() => {
  const q = searchQuery.value.trim().toLowerCase()
  if (!q) return groups.value

  return groups.value
    .map((group) => {
      if (matchText(group.name, q)) {
        return { ...group, items: group.items || [] }
      }
      const items = (group.items || []).filter((item) => matchItem(item, q))
      return { ...group, items }
    })
    .filter((group) => (group.items || []).length > 0)
})

function matchText(value: string | undefined, q: string) {
  return (value || '').toLowerCase().includes(q)
}

function matchItem(item: NavItem, q: string) {
  return (
    matchText(item.name, q) ||
    matchText(item.description, q) ||
    matchText(item.external_url, q) ||
    matchText(item.internal_url, q) ||
    matchText(item.url, q) ||
    matchText(item.icon, q)
  )
}

const {
  dragging,
  dragItemId,
  hoverGroupId,
  ghost,
  suppressClick,
  onItemPointerDown,
} = useItemReorder({
  groups,
  enabled: () => !searchQuery.value.trim() && !editOpen.value,
  onDragStart: () => closeMenu(),
  onCommit: async ({ destGroupId, destIds }) => {
    await sortItems(destGroupId, destIds)
  },
  onCommitError: async () => {
    showToast(t('home.sortFailed'))
    await reload()
  },
})

const ghostStyle = computed(() => {
  if (!ghost.value) return undefined
  return {
    left: `${ghost.value.x}px`,
    top: `${ghost.value.y}px`,
    width: `${ghost.value.width}px`,
  }
})

onMounted(async () => {
  await reload()
})

async function reload() {
  try {
    const data = await fetchNavigation()
    groups.value = data.groups
  } catch {
    groups.value = []
  }
}

function showToast(msg: string) {
  toast.value = msg
  window.clearTimeout(toastTimer)
  toastTimer = window.setTimeout(() => {
    toast.value = ''
  }, 1600)
}

function defaultUrl(item: NavItem) {
  return (item.external_url || item.url || '').trim()
}

function internalUrl(item: NavItem) {
  return (item.internal_url || '').trim()
}

function resolveOpenUrl(item: NavItem, mode: NetworkMode = settings.networkMode) {
  const external = defaultUrl(item)
  const internal = internalUrl(item)
  if (mode === 'internal') return internal || external
  if (mode === 'external') return external || internal
  return external || internal
}

function openItem(item: NavItem) {
  const url = resolveOpenUrl(item)
  if (!url) {
    showToast(t('home.noUrl'))
    return
  }
  const target = item.open_type === '_self' ? '_self' : '_blank'
  if (target === '_blank') {
    window.open(url, '_blank', 'noopener,noreferrer')
  } else {
    window.location.href = url
  }
}

function onPointerDownItem(e: PointerEvent, item: NavItem, groupId: string) {
  closeMenu()
  onItemPointerDown(e, item, groupId)
}

function onItemClick(item: NavItem) {
  if (dragging.value || suppressClick.value) return
  closeMenu()
  openItem(item)
}

function onItemContextMenu(e: MouseEvent, item: NavItem) {
  if (dragging.value || suppressClick.value) {
    e.preventDefault()
    return
  }
  e.preventDefault()
  e.stopPropagation()
  menuItem.value = item
  menuX.value = e.clientX
  menuY.value = e.clientY
  menuOpen.value = true
}

function closeMenu() {
  menuOpen.value = false
  menuItem.value = null
}

function onEdit(item: NavItem) {
  closeMenu()
  createGroupId.value = ''
  editItem.value = item
  editOpen.value = true
}

function onAddItem(group: NavGroup) {
  closeMenu()
  editItem.value = null
  createGroupId.value = group.id
  editOpen.value = true
}

function closeEditModal() {
  editOpen.value = false
  editItem.value = null
  createGroupId.value = ''
}

async function onRemove(item: NavItem) {
  closeMenu()
  const ok = window.confirm(t('home.deleteConfirm', { name: item.name }))
  if (!ok) return
  try {
    await deleteItem(item.id)
    removeLocal(item.id)
    showToast(t('home.deleted', { name: item.name }))
  } catch {
    showToast(t('edit.deleteFailed'))
  }
}

function removeLocal(id: string) {
  for (const group of groups.value) {
    const items = group.items || []
    const idx = items.findIndex((x) => x.id === id)
    if (idx >= 0) {
      items.splice(idx, 1)
      return
    }
  }
}

function onSaved(item: NavItem) {
  void reload().then(() => {
    showToast(t('settings.saved'))
  })
  editItem.value = item
  createGroupId.value = ''
}

function onDeleted(id: string) {
  removeLocal(id)
  showToast(t('home.deleted', { name: editItem.value?.name || '' }))
}

function itemTileText(item: NavItem) {
  const content = item.icon?.trim()
  if (content) return content
  const n = item.name?.trim()
  return n || '?'
}

const opticallyLargeIconHints = [
  'vercel',
  'adguard',
  'github',
  'circle',
  'disc',
  'globe',
  'earth',
  'play',
  'triangle',
  'hexagon',
  '音元',
]

const opticallySmallIconHints = ['plex', 'nas']

const opticallyNarrowIconHints = [
  'file',
  'document',
  'description',
  'receipt',
  'bookmark',
  'paperclip',
]

function iconOpticalSize(item: NavItem) {
  const identity = `${item.name || ''} ${item.icon || ''}`.toLowerCase()
  if (opticallyLargeIconHints.some((hint) => identity.includes(hint))) return 27
  if (opticallySmallIconHints.some((hint) => identity.includes(hint))) return 31
  if (opticallyNarrowIconHints.some((hint) => identity.includes(hint))) return 30
  return 29
}

function itemLabelClass(name: string) {
  const visualLength = Array.from(name || '').reduce(
    (length, char) => length + (/[\u2e80-\u9fff\uff00-\uffef]/.test(char) ? 2 : 1),
    0,
  )
  return {
    'is-compact': visualLength > 10,
    'is-wrapped': visualLength > 16,
  }
}

</script>

<template>
  <div class="home-board" :class="{ 'is-reordering': dragging }">
    <p v-if="searchQuery.trim() && !filteredGroups.length" class="search-empty">
      {{ t('search.noLocalResults', { q: searchQuery.trim() }) }}
    </p>

    <section
      v-for="group in filteredGroups"
      :key="group.id"
      class="section"
      :data-group-id="group.id"
      :class="{
        'is-drop-target': dragging && hoverGroupId === group.id,
      }"
    >
      <div class="section-head">
        <h2 class="section-title">{{ group.name }}</h2>
        <div class="section-actions">
          <button
            type="button"
            class="section-action"
            :title="t('home.addItem')"
            :aria-label="t('home.addItem')"
            @click="onAddItem(group)"
          >
            <Icon icon="lucide:plus" width="18" height="18" stroke-width="2" />
          </button>
        </div>
      </div>
      <div class="grid" :class="{ 'is-empty': !(group.items && group.items.length) }">
        <button
          v-for="item in group.items || []"
          :key="item.id"
          class="app-item"
          :class="{ 'is-placeholder': dragging && dragItemId === item.id }"
          type="button"
          :data-item-id="item.id"
          :aria-grabbed="dragging && dragItemId === item.id"
          draggable="false"
          @click="onItemClick(item)"
          @contextmenu="onItemContextMenu($event, item)"
          @pointerdown="onPointerDownItem($event, item, group.id)"
          @dragstart.prevent
        >
          <div class="tile" :class="{ 'is-text': item.icon_type === 'text' }">
            <img
              v-if="item.icon_type === 'image' && item.icon"
              class="tile-img"
              :src="item.icon"
              alt=""
              draggable="false"
            />
            <div v-else-if="item.icon_type === 'text'" class="tile-text-block">
              <div class="tile-text-main">{{ itemTileText(item) }}</div>
              <div v-if="item.description" class="tile-text-desc">{{ item.description }}</div>
            </div>
            <Icon
              v-else
              class="tile-icon"
              :icon="item.icon || 'mdi:application-outline'"
              :width="iconOpticalSize(item)"
              :height="iconOpticalSize(item)"
            />
          </div>
          <div class="app-label" :class="itemLabelClass(item.name)">{{ item.name }}</div>
        </button>
        <div
          v-if="!(group.items && group.items.length)"
          class="group-drop-empty"
        >
          {{ dragging ? t('home.dropHere') : t('home.emptyGroup') }}
        </div>
      </div>
    </section>

    <div
      v-if="ghost"
      class="app-item app-item-ghost"
      :class="{ 'is-returning': ghost.returning }"
      :style="ghostStyle"
    >
      <div class="tile" :class="{ 'is-text': ghost.item.icon_type === 'text' }">
        <img
          v-if="ghost.item.icon_type === 'image' && ghost.item.icon"
          class="tile-img"
          :src="ghost.item.icon"
          alt=""
          draggable="false"
        />
        <div v-else-if="ghost.item.icon_type === 'text'" class="tile-text-block">
          <div class="tile-text-main">{{ itemTileText(ghost.item) }}</div>
          <div v-if="ghost.item.description" class="tile-text-desc">{{ ghost.item.description }}</div>
        </div>
        <Icon
          v-else
          class="tile-icon"
          :icon="ghost.item.icon || 'mdi:application-outline'"
          :width="iconOpticalSize(ghost.item)"
          :height="iconOpticalSize(ghost.item)"
        />
      </div>
      <div class="app-label" :class="itemLabelClass(ghost.item.name)">{{ ghost.item.name }}</div>
    </div>

    <AppContextMenu
      :open="menuOpen"
      :item="menuItem"
      :x="menuX"
      :y="menuY"
      @close="closeMenu"
      @edit="onEdit"
      @remove="onRemove"
      @toast="showToast"
    />

    <EditItemModal
      :open="editOpen"
      :item="editItem"
      :create-group-id="createGroupId"
      :groups="groups"
      @close="closeEditModal"
      @saved="onSaved"
      @deleted="onDeleted"
    />

    <div v-if="toast" class="toast show">{{ toast }}</div>
  </div>
</template>
