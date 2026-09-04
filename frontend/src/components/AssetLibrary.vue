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

const emit = defineEmits<{
  saved: []
}>()

const { t } = useI18n()
const kind = ref<AssetKind>('icons')
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
    error.value = e instanceof Error ? e.message : t('settings.libraryLoadFailed')
    items.value = []
  } finally {
    loading.value = false
  }
}

watch(kind, () => {
  void load()
})

onMounted(() => {
  void load()
})

function triggerUpload() {
  fileInput.value?.click()
}

async function onFileChange(e: Event) {
  const input = e.target as HTMLInputElement
  const file = input.files?.[0]
  input.value = ''
  if (!file || uploading.value) return
  uploading.value = true
  error.value = ''
  try {
    await uploadAsset(kind.value, file)
    await load()
    emit('saved')
  } catch (err) {
    error.value = err instanceof Error ? err.message : t('settings.libraryUploadFailed')
  } finally {
    uploading.value = false
  }
}
</script>

<template>
  <section class="card">
    <div class="card-header">
      <h2 class="card-title">{{ t('settings.library') }}</h2>
      <button type="button" class="primary-btn" :disabled="uploading" @click="triggerUpload">
        {{ uploading ? t('common.loading') : t('settings.libraryUpload') }}
      </button>
      <input
        ref="fileInput"
        class="hidden-file"
        type="file"
        accept="image/png,image/jpeg,image/webp,image/gif,image/svg+xml,image/x-icon,.ico"
        @change="onFileChange"
      />
    </div>

    <div class="tabs" role="tablist" :aria-label="t('settings.library')">
      <button
        type="button"
        role="tab"
        :aria-selected="kind === 'icons'"
        :class="{ active: kind === 'icons' }"
        @click="kind = 'icons'"
      >
        {{ t('settings.libraryIcons') }}
      </button>
      <button
        type="button"
        role="tab"
        :aria-selected="kind === 'wallpapers'"
        :class="{ active: kind === 'wallpapers' }"
        @click="kind = 'wallpapers'"
      >
        {{ t('settings.libraryWallpapers') }}
      </button>
    </div>

    <p v-if="loading" class="status">{{ t('common.loading') }}</p>
    <p v-else-if="error" class="status error">{{ error }}</p>
    <div v-else-if="!items.length" class="library-empty">
      <Icon icon="mdi:image-multiple-outline" width="26" />
      <strong>{{ t('settings.libraryEmptyTitle') }}</strong>
      <span>{{ t('settings.libraryEmptyHint') }}</span>
      <button type="button" class="empty-upload-btn" :disabled="uploading" @click="triggerUpload">
        {{ t('settings.libraryUpload') }}
      </button>
    </div>
    <div v-else class="grid" :class="{ wallpapers: kind === 'wallpapers' }">
      <div v-for="item in items" :key="item.url" class="tile" :title="item.name">
        <img :src="item.url" :alt="item.name" />
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
  margin: 0 0 12px;
  font-size: 12px;
  line-height: 1.5;
  color: var(--sv-mute);
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
  transition: opacity 0.18s ease-out;
}

.primary-btn:hover:not(:disabled) {
  background: var(--sv-accent-hover);
}

.primary-btn:disabled {
  opacity: 0.55;
  cursor: default;
}

.hidden-file {
  display: none;
}

.tabs {
  display: inline-flex;
  border: 1px solid var(--sv-border);
  border-radius: 8px;
  overflow: hidden;
  margin-bottom: 12px;
}

.tabs button {
  background: transparent;
  border: none;
  color: var(--sv-mute);
  padding: 6px 14px;
  font-size: 13px;
  cursor: pointer;
  transition: background-color 0.18s ease-out, color 0.18s ease-out;
}

.tabs button + button {
  border-left: 1px solid var(--sv-border);
}

.tabs button.active {
  background: color-mix(in srgb, var(--sv-accent) 12%, transparent);
  color: var(--sv-accent);
}

.status {
  margin: 0;
  font-size: 13px;
  color: var(--sv-mute);
}

.status.error {
  color: #ef4444;
}

.library-empty {
  min-height: 132px;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 6px;
  color: var(--sv-mute);
  text-align: center;
}

.library-empty strong {
  color: var(--sv-text);
  font-size: 14px;
  font-weight: 600;
}

.library-empty span {
  font-size: 12px;
}

.empty-upload-btn {
  margin-top: 6px;
  height: 30px;
  padding: 0 12px;
  border: 1px solid var(--sv-border);
  border-radius: 8px;
  background: transparent;
  color: var(--sv-text);
  font-size: 12px;
  cursor: pointer;
}

.empty-upload-btn:hover:not(:disabled) {
  color: var(--sv-accent);
  background: color-mix(in srgb, var(--sv-accent) 8%, transparent);
}

.grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(72px, 1fr));
  gap: 10px;
}

.grid.wallpapers {
  grid-template-columns: repeat(auto-fill, minmax(120px, 1fr));
}

.tile {
  aspect-ratio: 1;
  border: 1px solid var(--sv-border);
  border-radius: 10px;
  overflow: hidden;
  background: var(--sv-bg);
}

.grid.wallpapers .tile {
  aspect-ratio: 16 / 10;
}

.tile img {
  width: 100%;
  height: 100%;
  object-fit: cover;
  display: block;
}
</style>
