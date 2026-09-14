<script setup lang="ts">
import { computed, inject, onMounted, onUnmounted, ref, type Ref } from 'vue'
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
import AppContextMenu from '../components/AppContextMenu.vue'
import EditItemModal from '../components/EditItemModal.vue'
import sortHandIcon from '../assets/sort-hand.png'

const sortIconStyle = {
  '--sort-icon': `url("${sortHandIcon}")`,
}

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

const sortingGroupId = ref('')
const dragItemId = ref('')
const sortBusy = ref(false)
let ignoreOutsideUntil = 0
let suppressClick = false

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

onMounted(async () => {
  window.addEventListener('pointerdown', onOutsidePointerDown, true)
  window.addEventListener('click', onOutsideClickCapture, true)
  await reload()
})

onUnmounted(() => {
  window.removeEventListener('pointerdown', onOutsidePointerDown, true)
  window.removeEventListener('click', onOutsideClickCapture, true)
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

function onItemClick(item: NavItem) {
  if (sortingGroupId.value) return
  closeMenu()
  openItem(item)
}

function onItemContextMenu(e: MouseEvent, item: NavItem) {
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

function isSortKeepTarget(target: EventTarget | null) {
  if (!(target instanceof Element)) return false
  if (target.closest('.section-action--sort')) return true
  if (target.closest('.section.sorting .grid')) return true
  if (target.closest('.app-ctx-menu')) return true
  if (target.closest('.modal-root')) return true
  return false
}

function onOutsidePointerDown(e: PointerEvent) {
  if (!sortingGroupId.value) return
  if (Date.now() < ignoreOutsideUntil) return
  if (isSortKeepTarget(e.target)) return
  suppressClick = true
  void finishSort()
}

function onOutsideClickCapture(e: MouseEvent) {
  if (!suppressClick) return
  suppressClick = false
  e.preventDefault()
  e.stopPropagation()
}

async function finishSort() {
  const groupId = sortingGroupId.value
  if (!groupId) return
  const group = groups.value.find((g) => g.id === groupId)
  sortingGroupId.value = ''
  dragItemId.value = ''
  if (group?.items && !sortBusy.value) {
    try {
      await sortItems(
        groupId,
        group.items.map((x) => x.id),
      )
    } catch {
      showToast(t('home.sortFailed'))
      await reload()
      return
    }
  }
  showToast(t('home.sortDone'))
}

async function toggleSort(group: NavGroup) {
  closeMenu()
  if (sortingGroupId.value === group.id) {
    await finishSort()
    return
  }
  if (sortingGroupId.value) {
    await finishSort()
  }
  sortingGroupId.value = group.id
  dragItemId.value = ''
  showToast(t('home.sortHint'))
}

function onItemDragStart(e: DragEvent, item: NavItem, groupId: string) {
  if (sortingGroupId.value !== groupId) {
    e.preventDefault()
    return
  }
  dragItemId.value = item.id
  if (e.dataTransfer) {
    e.dataTransfer.effectAllowed = 'move'
    e.dataTransfer.setData('text/plain', item.id)
  }
}

function onItemDragOver(e: DragEvent, groupId: string) {
  if (sortingGroupId.value !== groupId || !dragItemId.value) return
  e.preventDefault()
  if (e.dataTransfer) e.dataTransfer.dropEffect = 'move'
}

async function onItemDrop(e: DragEvent, target: NavItem, groupId: string) {
  e.preventDefault()
  if (sortingGroupId.value !== groupId || sortBusy.value) return
  const sourceId = dragItemId.value || e.dataTransfer?.getData('text/plain') || ''
  dragItemId.value = ''
  if (!sourceId || sourceId === target.id) return

  const group = groups.value.find((g) => g.id === groupId)
  if (!group?.items) return
  const items = [...group.items]
  const from = items.findIndex((x) => x.id === sourceId)
  const to = items.findIndex((x) => x.id === target.id)
  if (from < 0 || to < 0 || from === to) return

  const [moved] = items.splice(from, 1)
  items.splice(to, 0, moved)
  group.items = items

  sortBusy.value = true
  try {
    await sortItems(
      groupId,
      items.map((x) => x.id),
    )
  } catch {
    showToast(t('home.sortFailed'))
    await reload()
  } finally {
    sortBusy.value = false
  }
}

function onItemDragEnd() {
  dragItemId.value = ''
  ignoreOutsideUntil = Date.now() + 400
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
  <div>
    <p v-if="searchQuery.trim() && !filteredGroups.length" class="search-empty">
      {{ t('search.noLocalResults', { q: searchQuery.trim() }) }}
    </p>

    <section
      v-for="group in filteredGroups"
      :key="group.id"
      class="section"
      :class="{ sorting: sortingGroupId === group.id }"
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
          <button
            type="button"
            class="section-action section-action--sort"
            :class="{ active: sortingGroupId === group.id }"
            :title="sortingGroupId === group.id ? t('home.sortDone') : t('home.sortItems')"
            :aria-label="sortingGroupId === group.id ? t('home.sortDone') : t('home.sortItems')"
            @click="toggleSort(group)"
          >
            <span
              class="sort-hand-icon"
              aria-hidden="true"
              :style="sortIconStyle"
            />
          </button>
        </div>
      </div>
      <div class="grid">
        <button
          v-for="item in group.items || []"
          :key="item.id"
          class="app-item"
          :class="{ dragging: dragItemId === item.id }"
          type="button"
          :draggable="sortingGroupId === group.id"
          @click="onItemClick(item)"
          @contextmenu="onItemContextMenu($event, item)"
          @dragstart="onItemDragStart($event, item, group.id)"
          @dragover="onItemDragOver($event, group.id)"
          @drop="onItemDrop($event, item, group.id)"
          @dragend="onItemDragEnd"
        >
          <div class="tile" :class="{ 'is-text': item.icon_type === 'text' }">
            <img
              v-if="item.icon_type === 'image' && item.icon"
              class="tile-img"
              :src="item.icon"
              alt=""
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
      </div>
    </section>

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
