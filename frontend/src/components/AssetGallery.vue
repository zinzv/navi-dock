<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { Icon } from '@iconify/vue'
import {
  deleteAsset,
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
const deletingUrl = ref('')
const error = ref('')
const selectedUrl = ref('')
const fileInput = ref<HTMLInputElement | null>(null)

const hasItems = computed(() => items.value.length > 0)
const canUse = computed(() => Boolean(selectedUrl.value))

const emptyTitle = computed(() =>
  kind.value === 'wallpapers' ? t('edit.galleryEmptyWallpaperTitle') : t('edit.galleryEmptyIconTitle'),
)
const emptyHint = computed(() =>
  kind.value === 'wallpapers' ? t('edit.galleryEmptyWallpaperHint') : t('edit.galleryEmptyIconHint'),
)
const uploadLabel = computed(() =>
  kind.value === 'wallpapers' ? t('edit.galleryUploadWallpaper') : t('edit.galleryUploadIcon'),
)
const countLabel = computed(() =>
  kind.value === 'wallpapers'
    ? t('edit.galleryWallpaperCount', { count: items.value.length })
    : t('edit.galleryIconCount', { count: items.value.length }),
)

async function load() {
  loading.value = true
  error.value = ''
  selectedUrl.value = ''
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
    selectedUrl.value = ''
    void load()
  },
)

watch(kind, () => {
  if (props.open) void load()
})

onMounted(() => {
  if (props.open) void load()
})

function onSelect(url: string) {
  selectedUrl.value = url
}

function onUse() {
  if (!selectedUrl.value) return
  emit('select', selectedUrl.value)
  emit('close')
}

function onCancel() {
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
    selectedUrl.value = res.url
  } catch (err) {
    error.value = err instanceof Error ? err.message : t('edit.uploadFailed')
  } finally {
    uploading.value = false
  }
}

async function onDelete(item: AssetItem) {
  if (deletingUrl.value) return
  if (!window.confirm(t('edit.galleryDeleteConfirm', { name: item.name }))) return
  deletingUrl.value = item.url
  error.value = ''
  const clearsCurrent =
    settings.siteIcon === item.url || settings.backgroundImage === item.url
  try {
    await deleteAsset(kind.value, item.name)
    items.value = items.value.filter((candidate) => candidate.url !== item.url)
    if (selectedUrl.value === item.url) selectedUrl.value = ''
    if (clearsCurrent) await settings.load()
  } catch (err) {
    error.value = err instanceof Error ? err.message : t('edit.galleryDeleteFailed')
  } finally {
    deletingUrl.value = ''
  }
}
</script>

<template>
  <Teleport to="body">
    <div
      v-if="open"
      class="gallery-root"
      :data-theme="settings.resolvedTheme"
      @keydown.esc.prevent="onCancel"
    >
      <div class="gallery-backdrop" @click="onCancel" />
      <div class="gallery-panel" role="dialog" aria-modal="true" :aria-label="t('edit.gallery')">
        <header class="gallery-head">
          <div class="gallery-head-copy">
            <h3>{{ t('edit.gallery') }}</h3>
            <p>{{ t('edit.galleryHint') }}</p>
          </div>
          <button type="button" class="icon-close" :aria-label="t('edit.close')" @click="onCancel">
            <Icon icon="mdi:close" width="18" />
          </button>
        </header>

        <div class="gallery-toolbar">
          <div class="gallery-tabs" role="tablist" :aria-label="t('edit.gallery')">
            <button
              type="button"
              role="tab"
              :aria-selected="kind === 'icons'"
              :class="{ active: kind === 'icons' }"
              @click="kind = 'icons'"
            >
              {{ t('edit.galleryIcons') }}
            </button>
            <button
              type="button"
              role="tab"
              :aria-selected="kind === 'wallpapers'"
              :class="{ active: kind === 'wallpapers' }"
              @click="kind = 'wallpapers'"
            >
              {{ t('edit.galleryWallpapers') }}
            </button>
          </div>
          <button
            v-if="hasItems"
            type="button"
            class="upload-btn"
            :disabled="uploading"
            @click="triggerUpload"
          >
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

        <div class="gallery-body">
          <p v-if="error" class="gallery-error">{{ error }}</p>
          <p v-else-if="loading" class="gallery-hint">{{ t('common.loading') }}</p>

          <div v-else-if="!hasItems" class="gallery-empty">
            <div class="empty-icon">
              <Icon
                :icon="kind === 'wallpapers' ? 'mdi:image-outline' : 'mdi:image-multiple-outline'"
                width="22"
              />
            </div>
            <strong>{{ emptyTitle }}</strong>
            <span>{{ emptyHint }}</span>
            <button type="button" class="empty-upload-btn" :disabled="uploading" @click="triggerUpload">
              <Icon icon="mdi:upload" width="16" />
              {{ uploading ? t('common.loading') : uploadLabel }}
            </button>
          </div>

          <div
            v-else
            class="gallery-grid"
            :class="{ wallpapers: kind === 'wallpapers', icons: kind === 'icons' }"
          >
            <div
              v-for="item in items"
              :key="item.url"
              class="gallery-card"
            >
              <button
                type="button"
                class="gallery-item"
                :class="{ selected: selectedUrl === item.url }"
                :title="item.name"
                :disabled="deletingUrl === item.url"
                @click="onSelect(item.url)"
                @dblclick="selectedUrl = item.url; onUse()"
              >
                <img :src="item.url" :alt="item.name" />
                <span v-if="selectedUrl === item.url" class="check">
                  <Icon icon="mdi:check" width="12" />
                </span>
              </button>
              <button
                type="button"
                class="delete-asset"
                :class="{ deleting: deletingUrl === item.url }"
                :disabled="Boolean(deletingUrl)"
                :title="t('edit.galleryDelete')"
                :aria-label="t('edit.galleryDelete')"
                @click.stop="onDelete(item)"
              >
                <Icon
                  :icon="deletingUrl === item.url ? 'mdi:loading' : 'mdi:trash-can-outline'"
                  width="14"
                  :class="{ spin: deletingUrl === item.url }"
                />
              </button>
            </div>
          </div>
        </div>

        <footer v-if="hasItems" class="gallery-foot">
          <span class="count">{{ countLabel }}</span>
          <div class="foot-actions">
            <button type="button" class="ghost-btn" @click="onCancel">
              {{ t('settings.groupsCancel') }}
            </button>
            <button type="button" class="primary-btn" :disabled="!canUse" @click="onUse">
              {{ t('edit.galleryUse') }}
            </button>
          </div>
        </footer>
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
  --g-text: #182033;
  --g-mute: #737b8c;
  --g-border: #e5e7eb;
  --g-divider: #eef0f3;
  --g-surface: #f6f7f9;
  --g-hover: #f6f7f9;
  --g-accent: #5f66e8;
  --g-accent-soft: rgba(95, 102, 232, 0.08);
  --g-backdrop: rgba(15, 23, 42, 0.45);
}

.gallery-root[data-theme='dark'] {
  --g-bg: #202124;
  --g-text: rgba(255, 255, 255, 0.9);
  --g-mute: rgba(255, 255, 255, 0.55);
  --g-border: rgba(255, 255, 255, 0.08);
  --g-divider: rgba(255, 255, 255, 0.06);
  --g-surface: #26272b;
  --g-hover: rgba(255, 255, 255, 0.06);
  --g-accent: #6970ef;
  --g-accent-soft: rgba(105, 112, 239, 0.14);
  --g-backdrop: rgba(2, 6, 23, 0.62);
}

.gallery-backdrop {
  position: absolute;
  inset: 0;
  background: var(--g-backdrop);
}

.gallery-panel {
  position: relative;
  width: min(680px, 100%);
  min-height: 440px;
  max-height: 70vh;
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
  align-items: flex-start;
  justify-content: space-between;
  gap: 16px;
  padding: 20px 22px 16px;
  border-bottom: 1px solid var(--g-divider);
  flex-shrink: 0;
}

.gallery-head-copy h3 {
  margin: 0;
  font-size: 18px;
  font-weight: 600;
  line-height: 24px;
}

.gallery-head-copy p {
  margin: 4px 0 0;
  font-size: 13px;
  color: var(--g-mute);
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
  flex-shrink: 0;
}

.icon-close:hover {
  background: var(--g-hover);
  color: var(--g-text);
}

.gallery-toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  padding: 16px 22px 0;
  flex-shrink: 0;
}

.gallery-tabs {
  display: inline-flex;
  align-items: center;
  gap: 24px;
  height: 36px;
}

.gallery-tabs button {
  position: relative;
  border: 0;
  background: transparent;
  color: var(--g-mute);
  padding: 0 0 10px;
  font-size: 14px;
  font-weight: 500;
  cursor: pointer;
}

.gallery-tabs button.active {
  color: var(--g-accent);
  font-weight: 600;
}

.gallery-tabs button.active::after {
  content: '';
  position: absolute;
  left: 0;
  right: 0;
  bottom: 0;
  height: 2px;
  border-radius: 999px;
  background: var(--g-accent);
}

.upload-btn,
.empty-upload-btn {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  height: 34px;
  padding: 0 12px;
  border-radius: 8px;
  border: 1px solid var(--g-border);
  background: transparent;
  color: var(--g-text);
  font-size: 13px;
  font-weight: 500;
  cursor: pointer;
}

.upload-btn:hover:not(:disabled),
.empty-upload-btn:hover:not(:disabled) {
  background: var(--g-accent-soft);
  border-color: color-mix(in srgb, var(--g-accent) 30%, var(--g-border));
  color: var(--g-accent);
}

.upload-btn:disabled,
.empty-upload-btn:disabled {
  opacity: 0.55;
  cursor: default;
}

.hidden-file {
  display: none;
}

.gallery-body {
  flex: 1;
  min-height: 0;
  display: flex;
  flex-direction: column;
  padding: 16px 22px 20px;
}

.gallery-hint,
.gallery-error {
  margin: 0;
  font-size: 13px;
  color: var(--g-mute);
}

.gallery-error {
  color: #ef4444;
}

.gallery-empty {
  flex: 1;
  min-height: 260px;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 6px;
  text-align: center;
}

.empty-icon {
  width: 48px;
  height: 48px;
  border-radius: 12px;
  display: grid;
  place-items: center;
  background: var(--g-accent-soft);
  color: color-mix(in srgb, var(--g-accent) 70%, var(--g-text));
  margin-bottom: 6px;
}

.gallery-empty strong {
  font-size: 14px;
  font-weight: 500;
  color: var(--g-text);
}

.gallery-empty span {
  font-size: 13px;
  color: var(--g-mute);
}

.empty-upload-btn {
  margin-top: 10px;
}

.gallery-grid {
  display: grid;
  gap: 12px;
  overflow: auto;
  min-height: 0;
  flex: 1;
  align-content: start;
  scrollbar-width: thin;
}

.gallery-grid.icons {
  grid-template-columns: repeat(auto-fill, minmax(68px, 1fr));
}

.gallery-grid.wallpapers {
  grid-template-columns: repeat(4, minmax(0, 1fr));
}

.gallery-card {
  position: relative;
  min-width: 0;
}

.gallery-grid.icons .gallery-card {
  aspect-ratio: 1;
}

.gallery-grid.wallpapers .gallery-card {
  aspect-ratio: 16 / 10;
}

.gallery-item {
  width: 100%;
  height: 100%;
  border: 1px solid transparent;
  border-radius: 8px;
  padding: 0;
  background: var(--g-surface);
  overflow: hidden;
  cursor: pointer;
  transition:
    transform 140ms ease,
    box-shadow 140ms ease,
    border-color 140ms ease;
}

.gallery-item:hover {
  transform: translateY(-1px);
  box-shadow: 0 3px 10px rgba(0, 0, 0, 0.08);
}

.gallery-item.selected {
  border: 2px solid var(--g-accent);
}

.gallery-item:disabled {
  opacity: 0.5;
  cursor: wait;
}

.gallery-item img {
  width: 100%;
  height: 100%;
  object-fit: cover;
  display: block;
}

.gallery-grid.icons .gallery-item img {
  object-fit: contain;
  padding: 14px;
  box-sizing: border-box;
}

.check {
  position: absolute;
  bottom: 7px;
  right: 8px;
  width: 20px;
  height: 20px;
  border-radius: 999px;
  background: var(--g-accent);
  color: #fff;
  display: grid;
  place-items: center;
}

.delete-asset {
  position: absolute;
  z-index: 2;
  top: 6px;
  right: 6px;
  width: 26px;
  height: 26px;
  display: grid;
  place-items: center;
  padding: 0;
  border: 1px solid color-mix(in srgb, #fff 12%, transparent);
  border-radius: 7px;
  background: rgba(12, 18, 28, 0.78);
  color: rgba(255, 255, 255, 0.82);
  box-shadow: 0 3px 10px rgba(0, 0, 0, 0.16);
  opacity: 0;
  transform: translateY(-2px);
  cursor: pointer;
  transition: opacity 140ms ease, transform 140ms ease, background 140ms ease;
}

.gallery-card:hover .delete-asset,
.delete-asset:focus-visible,
.delete-asset.deleting {
  opacity: 1;
  transform: translateY(0);
}

.delete-asset:hover:not(:disabled) {
  background: rgba(220, 55, 65, 0.9);
  color: #fff;
}

.delete-asset:disabled:not(.deleting) {
  pointer-events: none;
}

.spin {
  animation: gallery-spin 0.8s linear infinite;
}

@keyframes gallery-spin {
  to {
    transform: rotate(360deg);
  }
}

.gallery-foot {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  height: 64px;
  padding: 12px 22px;
  border-top: 1px solid var(--g-divider);
  flex-shrink: 0;
}

.count {
  font-size: 13px;
  color: var(--g-mute);
}

.foot-actions {
  display: inline-flex;
  align-items: center;
  gap: 8px;
}

.ghost-btn,
.primary-btn {
  height: 34px;
  padding: 0 14px;
  border-radius: 8px;
  font-size: 13px;
  font-weight: 500;
  cursor: pointer;
}

.ghost-btn {
  border: 1px solid var(--g-border);
  background: transparent;
  color: var(--g-text);
}

.ghost-btn:hover {
  background: var(--g-hover);
}

.primary-btn {
  border: 0;
  background: var(--g-accent);
  color: #fff;
}

.primary-btn:disabled {
  opacity: 0.45;
  cursor: not-allowed;
}

.primary-btn:hover:not(:disabled) {
  filter: brightness(1.05);
}

@media (max-width: 720px) {
  .gallery-grid.wallpapers {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }

  .gallery-toolbar {
    flex-wrap: wrap;
  }
}

@media (hover: none) {
  .delete-asset {
    opacity: 1;
    transform: none;
  }
}
</style>
