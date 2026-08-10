<script setup lang="ts">
import { onMounted, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { Icon } from '@iconify/vue'
import {
  listAssets,
  uploadAsset,
  type AssetItem,
  type AssetKind,
} from '../api/settings'
import { useSettingsStore } from '../stores/settings'

const props = defineProps<{
  open: boolean
  initialKind?: AssetKind
}>()

const emit = defineEmits<{
  close: []
  select: [url: string]
}>()

const { t } = useI18n()
const settings = useSettingsStore()
const kind = ref<AssetKind>(props.initialKind || 'icons')
const items = ref<AssetItem[]>([])
const loading = ref(false)
const uploading = ref(false)
const error = ref('')
const fileInput = ref<HTMLInputElement | null>(null)

async function load() {
  loading.value = true
  error.value = ''
  try {
    const res = await listAssets(kind.value)
    items.value = res.items || []
  } catch (e) {
    error.value = e instanceof Error ? e.message : t('edit.galleryLoadFailed')
    items.value = []
  } finally {
    loading.value = false
  }
}

watch(
  () => [props.open, props.initialKind] as const,
  ([open, initial]) => {
    if (!open) return
    kind.value = initial || 'icons'
    void load()
  },
)

watch(kind, () => {
  if (props.open) void load()
})

onMounted(() => {
  if (props.open) void load()
})

function onPick(url: string) {
  emit('select', url)
  emit('close')
}

function triggerUpload() {
  fileInput.value?.click()
}

async function onFileChange(e: Event) {
  const input = e.target as HTMLInputElement
  const file = input.files?.[0]
  input.value = ''
  if (!file) return
  uploading.value = true
  error.value = ''
  try {
    const res = await uploadAsset(kind.value, file)
    await load()
    onPick(res.url)
  } catch (err) {
    error.value = err instanceof Error ? err.message : t('edit.uploadFailed')
  } finally {
    uploading.value = false
  }
}
</script>

<template>
  <Teleport to="body">
    <div
      v-if="open"
      class="gallery-root"
      :data-theme="settings.resolvedTheme"
      @keydown.esc.prevent="emit('close')"
    >
      <div class="gallery-backdrop" @click="emit('close')" />
      <div class="gallery-panel" role="dialog" aria-modal="true" :aria-label="t('edit.gallery')">
        <header class="gallery-head">
          <h3>{{ t('edit.gallery') }}</h3>
          <button type="button" class="icon-close" :aria-label="t('edit.close')" @click="emit('close')">
            <Icon icon="mdi:close" width="18" />
          </button>
        </header>

        <div class="gallery-tabs">
          <button type="button" :class="{ active: kind === 'icons' }" @click="kind = 'icons'">
            {{ t('edit.galleryIcons') }}
          </button>
          <button type="button" :class="{ active: kind === 'wallpapers' }" @click="kind = 'wallpapers'">
            {{ t('edit.galleryWallpapers') }}
          </button>
        </div>

        <div class="gallery-toolbar">
          <button type="button" class="upload-btn" :disabled="uploading" @click="triggerUpload">
            <Icon icon="mdi:upload" width="16" />
            {{ uploading ? t('common.loading') : t('edit.localUpload') }}
          </button>
          <input
            ref="fileInput"
            class="hidden-file"
            type="file"
            accept="image/png,image/jpeg,image/webp,image/gif,image/svg+xml,image/x-icon,.ico"
            @change="onFileChange"
          />
        </div>

        <p v-if="error" class="gallery-error">{{ error }}</p>
        <p v-else-if="loading" class="gallery-hint">{{ t('common.loading') }}</p>
        <p v-else-if="!items.length" class="gallery-hint">{{ t('edit.galleryEmpty') }}</p>

        <div v-else class="gallery-grid">
          <button
            v-for="item in items"
            :key="item.url"
            type="button"
            class="gallery-item"
            :title="item.name"
            @click="onPick(item.url)"
          >
            <img :src="item.url" :alt="item.name" />
          </button>
        </div>
      </div>
    </div>
  </Teleport>
</template>

<style scoped>
.gallery-root {
  position: fixed;
  inset: 0;
  z-index: 1300;
  display: grid;
  place-items: center;
  padding: 20px 16px;
}

.gallery-root[data-theme='light'] {
  --g-bg: #ffffff;
  --g-text: #1f2937;
  --g-mute: #6b7280;
  --g-border: #e5e7eb;
  --g-surface: #f8fafc;
  --g-hover: #f3f4f6;
  --g-accent: #6366f1;
  --g-backdrop: rgba(15, 23, 42, 0.45);
}

.gallery-root[data-theme='dark'] {
  --g-bg: #171b24;
  --g-text: #f3f4f6;
  --g-mute: #9ca3af;
  --g-border: rgba(255, 255, 255, 0.1);
  --g-surface: #111827;
  --g-hover: rgba(255, 255, 255, 0.06);
  --g-accent: #818cf8;
  --g-backdrop: rgba(2, 6, 23, 0.62);
}

.gallery-backdrop {
  position: absolute;
  inset: 0;
  background: var(--g-backdrop);
}

.gallery-panel {
  position: relative;
  width: min(520px, 100%);
  max-height: min(80vh, 640px);
  display: flex;
  flex-direction: column;
  background: var(--g-bg);
  color: var(--g-text);
  border: 1px solid var(--g-border);
  border-radius: 14px;
  overflow: hidden;
  box-shadow: 0 24px 64px rgba(0, 0, 0, 0.28);
}

.gallery-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 14px 16px;
  border-bottom: 1px solid var(--g-border);
  flex-shrink: 0;
}

.gallery-head h3 {
  margin: 0;
  font-size: 16px;
  font-weight: 600;
}

.icon-close {
  width: 32px;
  height: 32px;
  border: 0;
  border-radius: 8px;
  background: transparent;
  color: var(--g-mute);
  display: grid;
  place-items: center;
  cursor: pointer;
}

.icon-close:hover {
  background: var(--g-hover);
  color: var(--g-text);
}

.gallery-tabs {
  display: flex;
  gap: 0;
  margin: 12px 16px 0;
  border: 1px solid var(--g-border);
  border-radius: 8px;
  overflow: hidden;
  flex-shrink: 0;
}

.gallery-tabs button {
  flex: 1;
  border: 0;
  background: transparent;
  color: var(--g-mute);
  padding: 8px 12px;
  font-size: 13px;
  cursor: pointer;
}

.gallery-tabs button + button {
  border-left: 1px solid var(--g-border);
}

.gallery-tabs button.active {
  background: var(--g-accent);
  color: #fff;
}

.gallery-toolbar {
  padding: 12px 16px;
  flex-shrink: 0;
}

.upload-btn {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  height: 34px;
  padding: 0 12px;
  border-radius: 8px;
  border: 1px solid var(--g-accent);
  background: transparent;
  color: var(--g-accent);
  font-size: 13px;
  cursor: pointer;
}

.upload-btn:disabled {
  opacity: 0.55;
  cursor: default;
}

.hidden-file {
  display: none;
}

.gallery-hint,
.gallery-error {
  margin: 0 16px 8px;
  font-size: 13px;
  color: var(--g-mute);
}

.gallery-error {
  color: #ef4444;
}

.gallery-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(88px, 1fr));
  gap: 10px;
  padding: 0 16px 16px;
  overflow: auto;
  min-height: 0;
  scrollbar-width: thin;
}

.gallery-item {
  aspect-ratio: 1;
  border: 1px solid var(--g-border);
  border-radius: 10px;
  padding: 0;
  background: var(--g-surface);
  overflow: hidden;
  cursor: pointer;
}

.gallery-item:hover {
  border-color: var(--g-accent);
}

.gallery-item img {
  width: 100%;
  height: 100%;
  object-fit: cover;
  display: block;
}
</style>
