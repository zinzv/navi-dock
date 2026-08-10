<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { Icon } from '@iconify/vue'
import type { NavItem } from '../api/settings'

const props = defineProps<{
  item: NavItem | null
  x: number
  y: number
  open: boolean
}>()

const emit = defineEmits<{
  close: []
  edit: [item: NavItem]
  remove: [item: NavItem]
  toast: [message: string]
}>()

const { t } = useI18n()
const menuRef = ref<HTMLElement | null>(null)
const pos = ref({ left: 0, top: 0 })

const defaultUrl = computed(() => {
  if (!props.item) return ''
  return (props.item.external_url || props.item.url || '').trim()
})

const internalUrl = computed(() => {
  if (!props.item) return ''
  return (props.item.internal_url || '').trim()
})

function clampPosition() {
  const el = menuRef.value
  if (!el) return
  const pad = 8
  const rect = el.getBoundingClientRect()
  let left = props.x
  let top = props.y
  if (left + rect.width > window.innerWidth - pad) {
    left = Math.max(pad, window.innerWidth - rect.width - pad)
  }
  if (top + rect.height > window.innerHeight - pad) {
    top = Math.max(pad, window.innerHeight - rect.height - pad)
  }
  pos.value = { left, top }
}

watch(
  () => [props.open, props.x, props.y, props.item?.id] as const,
  async ([open]) => {
    if (!open) return
    pos.value = { left: props.x, top: props.y }
    await nextTick()
    clampPosition()
  },
)

function onKeydown(e: KeyboardEvent) {
  if (e.key === 'Escape') emit('close')
}

function onPointerDown(e: PointerEvent) {
  if (!props.open) return
  const el = menuRef.value
  if (el && e.target instanceof Node && el.contains(e.target)) return
  emit('close')
}

function onViewportChange() {
  if (props.open) emit('close')
}

onMounted(() => {
  window.addEventListener('keydown', onKeydown)
  window.addEventListener('pointerdown', onPointerDown, true)
  window.addEventListener('resize', onViewportChange)
  window.addEventListener('scroll', onViewportChange, true)
})

onBeforeUnmount(() => {
  window.removeEventListener('keydown', onKeydown)
  window.removeEventListener('pointerdown', onPointerDown, true)
  window.removeEventListener('resize', onViewportChange)
  window.removeEventListener('scroll', onViewportChange, true)
})

function openUrl(url: string, target: '_blank' | '_self' = '_blank') {
  if (!url) {
    emit('toast', t('home.noUrl'))
    return
  }
  if (target === '_blank') {
    window.open(url, '_blank', 'noopener,noreferrer')
  } else {
    window.location.href = url
  }
  emit('close')
}

async function copyUrl(url: string) {
  if (!url) {
    emit('toast', t('home.noUrl'))
    return
  }
  try {
    await navigator.clipboard.writeText(url)
    emit('toast', t('home.copied'))
  } catch {
    emit('toast', t('home.copyFailed'))
  }
  emit('close')
}

function onEdit() {
  if (!props.item) return
  emit('edit', props.item)
  emit('close')
}

function onRemove() {
  if (!props.item) return
  emit('remove', props.item)
  emit('close')
}
</script>

<template>
  <Teleport to="body">
    <div
      v-if="open && item"
      ref="menuRef"
      class="app-ctx-menu"
      :style="{ left: `${pos.left}px`, top: `${pos.top}px` }"
      role="menu"
      @contextmenu.prevent
    >
      <div v-if="defaultUrl" class="ctx-block">
        <div class="ctx-label">{{ t('home.openDefault') }}</div>
        <div class="ctx-actions">
          <button
            type="button"
            class="ctx-icon-btn"
            :title="t('home.openNewTab')"
            @click="openUrl(defaultUrl, '_blank')"
          >
            <Icon icon="mdi:open-in-new" width="16" />
          </button>
          <button
            type="button"
            class="ctx-icon-btn"
            :title="t('home.copyUrl')"
            @click="copyUrl(defaultUrl)"
          >
            <Icon icon="mdi:content-copy" width="16" />
          </button>
          <button
            type="button"
            class="ctx-icon-btn"
            :title="t('home.openSameTab')"
            @click="openUrl(defaultUrl, '_self')"
          >
            <Icon icon="mdi:link-variant" width="16" />
          </button>
        </div>
      </div>

      <div v-if="internalUrl" class="ctx-block">
        <div class="ctx-label">{{ t('home.openInternal') }}</div>
        <div class="ctx-actions">
          <button
            type="button"
            class="ctx-icon-btn"
            :title="t('home.openNewTab')"
            @click="openUrl(internalUrl, '_blank')"
          >
            <Icon icon="mdi:open-in-new" width="16" />
          </button>
          <button
            type="button"
            class="ctx-icon-btn"
            :title="t('home.copyUrl')"
            @click="copyUrl(internalUrl)"
          >
            <Icon icon="mdi:content-copy" width="16" />
          </button>
          <button
            type="button"
            class="ctx-icon-btn"
            :title="t('home.openSameTab')"
            @click="openUrl(internalUrl, '_self')"
          >
            <Icon icon="mdi:link-variant" width="16" />
          </button>
        </div>
      </div>

      <div class="ctx-divider" />

      <button type="button" class="ctx-row" role="menuitem" @click="onEdit">
        <Icon icon="mdi:pencil-outline" width="16" />
        <span>{{ t('home.edit') }}</span>
      </button>
      <button type="button" class="ctx-row danger" role="menuitem" @click="onRemove">
        <Icon icon="mdi:trash-can-outline" width="16" />
        <span>{{ t('home.delete') }}</span>
      </button>
    </div>
  </Teleport>
</template>

<style scoped>
.app-ctx-menu {
  position: fixed;
  z-index: 1000;
  min-width: 168px;
  padding: 8px;
  border-radius: 12px;
  background: #fff;
  color: #1f2937;
  border: 1px solid #e5e7eb;
  box-shadow: 0 12px 32px rgba(15, 23, 42, 0.16);
  font-size: 13px;
}

.ctx-block {
  padding: 4px 4px 8px;
}

.ctx-label {
  font-size: 12px;
  color: #6b7280;
  margin-bottom: 6px;
  padding: 0 2px;
}

.ctx-actions {
  display: flex;
  gap: 6px;
}

.ctx-icon-btn {
  width: 32px;
  height: 32px;
  display: grid;
  place-items: center;
  border: 1px solid #e5e7eb;
  border-radius: 8px;
  background: #fff;
  color: #374151;
  cursor: pointer;
  padding: 0;
  transition: background 0.15s ease, border-color 0.15s ease;
}

.ctx-icon-btn:hover {
  background: #f3f4f6;
  border-color: #d1d5db;
}

.ctx-divider {
  height: 1px;
  background: #e5e7eb;
  margin: 4px 0 6px;
}

.ctx-row {
  width: 100%;
  display: flex;
  align-items: center;
  gap: 8px;
  border: 0;
  background: transparent;
  color: #1f2937;
  padding: 8px 6px;
  border-radius: 8px;
  cursor: pointer;
  text-align: left;
}

.ctx-row:hover {
  background: #f3f4f6;
}

.ctx-row.danger {
  color: #dc2626;
}

.ctx-row.danger:hover {
  background: #fef2f2;
}
</style>
