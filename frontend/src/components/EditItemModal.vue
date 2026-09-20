<script setup lang="ts">
import { computed, reactive, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { Icon } from '@iconify/vue'
import {
  createItem,
  deleteItem,
  updateItem,
  uploadAsset,
  type NavGroup,
  type NavItem,
} from '../api/settings'
import { useSettingsStore } from '../stores/settings'
import AssetGallery from './AssetGallery.vue'

const props = defineProps<{
  open: boolean
  item: NavItem | null
  createGroupId?: string
  groups: NavGroup[]
}>()

const emit = defineEmits<{
  close: []
  saved: [item: NavItem]
  deleted: [id: string]
}>()

const { t } = useI18n()
const settings = useSettingsStore()
const busy = ref(false)
const uploading = ref(false)
const error = ref('')
const showMore = ref(false)
const galleryOpen = ref(false)
const fileInput = ref<HTMLInputElement | null>(null)

const isCreate = computed(() => props.open && !props.item && !!props.createGroupId)

const form = reactive({
  name: '',
  description: '',
  iconType: 'iconify' as 'iconify' | 'text' | 'image',
  icon: '',
  externalUrl: '',
  internalUrl: '',
  openType: '_blank' as '_blank' | '_self',
  groupId: '',
})

function resetForm(groupId = '') {
  error.value = ''
  showMore.value = !!groupId
  galleryOpen.value = false
  form.name = ''
  form.description = ''
  form.iconType = 'iconify'
  form.icon = ''
  form.externalUrl = ''
  form.internalUrl = ''
  form.openType = '_blank'
  form.groupId = groupId
}

watch(
  () => [props.open, props.item, props.createGroupId] as const,
  ([open, item, createGroupId]) => {
    if (!open) return
    if (item) {
      error.value = ''
      showMore.value = false
      galleryOpen.value = false
      form.name = item.name || ''
      form.description = item.description || ''
      const type = item.icon_type
      form.iconType = type === 'text' || type === 'image' ? type : 'iconify'
      form.icon = item.icon || ''
      form.externalUrl = item.external_url || item.url || ''
      form.internalUrl = item.internal_url || ''
      form.openType = item.open_type === '_self' ? '_self' : '_blank'
      form.groupId = item.group_id || ''
      return
    }
    if (createGroupId) {
      resetForm(createGroupId)
    }
  },
)

const nameCount = computed(() => form.name.length)
const descCount = computed(() => form.description.length)
const textContentCount = computed(() => form.icon.length)

const tileText = computed(() => {
  const content = form.icon.trim()
  if (content) return content
  const name = form.name.trim()
  return name || '?'
})

function close() {
  if (busy.value || uploading.value) return
  galleryOpen.value = false
  emit('close')
}

function setIconType(next: 'text' | 'image' | 'iconify') {
  if (form.iconType === next) return
  const prev = form.iconType
  form.iconType = next
  // Avoid carrying iconify id / image url into text content field
  if (next === 'text' && (prev === 'iconify' || prev === 'image')) {
    form.icon = form.name.trim().slice(0, 10)
  } else if (next !== 'text' && prev === 'text') {
    form.icon = ''
  }
}

function triggerLocalUpload() {
  fileInput.value?.click()
}

async function onLocalFile(e: Event) {
  const input = e.target as HTMLInputElement
  const file = input.files?.[0]
  input.value = ''
  if (!file) return
  uploading.value = true
  error.value = ''
  try {
    const res = await uploadAsset('icons', file)
    form.iconType = 'image'
    form.icon = res.url
  } catch (err) {
    error.value = err instanceof Error ? err.message : t('edit.uploadFailed')
  } finally {
    uploading.value = false
  }
}

function onGallerySelect(url: string) {
  form.iconType = 'image'
  form.icon = url
}

async function onSave() {
  if (busy.value) return
  if (!isCreate.value && !props.item) return
  const name = form.name.trim()
  const external = form.externalUrl.trim()
  const internal = form.internalUrl.trim()
  if (!name) {
    error.value = t('edit.nameRequired')
    return
  }
  if (!external && !internal) {
    error.value = t('edit.urlRequired')
    return
  }
  if (form.iconType === 'image' && !form.icon.trim()) {
    error.value = t('edit.imageRequired')
    return
  }
  if (form.iconType === 'text' && !form.icon.trim()) {
    error.value = t('edit.textContentRequired')
    return
  }
  const groupId = form.groupId || props.item?.group_id || props.createGroupId || ''
  if (!groupId) {
    error.value = t('edit.groupRequired')
    return
  }
  busy.value = true
  error.value = ''
  const payload = {
    group_id: groupId,
    name,
    description: form.description.trim(),
    icon_type: form.iconType,
    icon: form.icon.trim(),
    external_url: external,
    internal_url: internal,
    open_type: form.openType,
  }
  try {
    const saved = isCreate.value
      ? await createItem(payload)
      : await updateItem(props.item!.id, payload)
    emit('saved', saved)
    emit('close')
  } catch (e) {
    error.value = e instanceof Error ? e.message : t('edit.saveFailed')
  } finally {
    busy.value = false
  }
}

async function onDelete() {
  if (!props.item || busy.value || isCreate.value) return
  if (!window.confirm(t('home.deleteConfirm', { name: form.name || props.item.name }))) return
  busy.value = true
  error.value = ''
  try {
    await deleteItem(props.item.id)
    emit('deleted', props.item.id)
    emit('close')
  } catch (e) {
    error.value = e instanceof Error ? e.message : t('edit.deleteFailed')
  } finally {
    busy.value = false
  }
}
</script>

<template>
  <Teleport to="body">
    <div
      v-if="open && (item || createGroupId)"
      class="modal-root"
      :data-theme="settings.resolvedTheme"
      @keydown.esc.prevent="close"
    >
      <div class="modal-backdrop" @click="close" />
      <div
        class="modal"
        role="dialog"
        aria-modal="true"
        :aria-label="isCreate ? t('edit.createTitle') : t('edit.title')"
      >
        <header class="modal-head">
          <h2>{{ isCreate ? t('edit.createTitle') : t('edit.title') }}</h2>
          <button type="button" class="close-btn" :aria-label="t('edit.close')" @click="close">
            <Icon icon="mdi:close" width="18" />
          </button>
        </header>

        <div class="modal-body">
          <div class="preview">
            <div class="preview-tile" :class="{ 'is-text': form.iconType === 'text' }">
              <img
                v-if="form.iconType === 'image' && form.icon"
                class="preview-img"
                :src="form.icon"
                alt=""
              />
              <div v-else-if="form.iconType === 'text'" class="preview-text-block">
                <div class="preview-text-main">{{ tileText }}</div>
                <div v-if="form.description.trim()" class="preview-text-desc">
                  {{ form.description }}
                </div>
              </div>
              <Icon v-else :icon="form.icon || 'mdi:application-outline'" width="28" />
            </div>
            <div class="preview-meta">
              <div class="preview-title">{{ form.name || t('edit.namePlaceholder') }}</div>
            </div>
          </div>

          <div class="field">
            <label class="field-label">{{ t('edit.iconStyle') }}</label>
            <div class="seg">
              <button
                type="button"
                :class="{ active: form.iconType === 'text' }"
                @click="setIconType('text')"
              >
                {{ t('edit.iconText') }}
              </button>
              <button
                type="button"
                :class="{ active: form.iconType === 'image' }"
                @click="setIconType('image')"
              >
                {{ t('edit.iconImage') }}
              </button>
              <button
                type="button"
                :class="{ active: form.iconType === 'iconify' }"
                @click="setIconType('iconify')"
              >
                {{ t('edit.iconOnline') }}
              </button>
            </div>
          </div>

          <div v-if="form.iconType === 'text'" class="field">
            <label class="field-label">{{ t('edit.textContent') }}</label>
            <div class="input-with-count w-short">
              <input
                v-model="form.icon"
                class="input"
                type="text"
                maxlength="10"
                :placeholder="t('edit.textContentPlaceholder')"
              />
              <span class="inline-counter">{{ textContentCount }} / 10</span>
            </div>
          </div>

          <div v-else-if="form.iconType === 'image'" class="field">
            <label class="field-label">{{ t('edit.imageAddress') }}</label>
            <div class="image-row">
              <input
                v-model="form.icon"
                class="input image-input"
                type="text"
                :placeholder="t('edit.imageInputPlaceholder')"
              />
              <button type="button" class="action-btn" @click="galleryOpen = true">
                <Icon icon="mdi:view-grid-outline" width="16" />
                {{ t('edit.gallery') }}
              </button>
              <button type="button" class="action-btn" :disabled="uploading" @click="triggerLocalUpload">
                <Icon icon="mdi:upload" width="16" />
                {{ uploading ? t('common.loading') : t('edit.localUpload') }}
              </button>
              <input
                ref="fileInput"
                class="hidden-file"
                type="file"
                accept="image/png,image/jpeg,image/webp,image/gif,image/svg+xml"
                @change="onLocalFile"
              />
            </div>
          </div>

          <div v-else-if="form.iconType === 'iconify'" class="field">
            <label class="field-label">{{ t('edit.iconValue') }}</label>
            <div class="iconify-row">
              <input
                v-model="form.icon"
                class="input w-short"
                type="text"
                placeholder="mdi:nas"
              />
              <a
                class="field-link iconify-lib"
                href="https://icon-sets.iconify.design/"
                target="_blank"
                rel="noopener noreferrer"
              >
                {{ t('edit.iconLibrary') }}
              </a>
            </div>
          </div>

          <div class="field-row" :class="{ single: form.iconType !== 'text' }">
            <div class="field">
              <label class="field-label">
                {{ t('edit.name') }} <span class="req">*</span>
              </label>
              <div class="input-with-count w-short">
                <input
                  v-model="form.name"
                  class="input"
                  type="text"
                  maxlength="20"
                  :placeholder="t('edit.namePlaceholder')"
                />
                <span class="inline-counter">{{ nameCount }} / 20</span>
              </div>
            </div>
            <div v-if="form.iconType === 'text'" class="field">
              <label class="field-label">{{ t('edit.description') }}</label>
              <div class="input-with-count w-short">
                <input
                  v-model="form.description"
                  class="input"
                  type="text"
                  maxlength="100"
                  :placeholder="t('edit.descriptionPlaceholder')"
                />
                <span class="inline-counter">{{ descCount }} / 100</span>
              </div>
            </div>
          </div>

          <div class="field">
            <label class="field-label">{{ t('edit.defaultUrl') }}</label>
            <input
              v-model="form.externalUrl"
              class="input w-url"
              type="url"
              :placeholder="t('edit.urlPlaceholder')"
            />
            <span class="url-hint">{{ t('edit.externalUrlHint') }}</span>
          </div>

          <div class="field">
            <label class="field-label">{{ t('edit.internalUrl') }}</label>
            <input
              v-model="form.internalUrl"
              class="input w-url"
              type="url"
              :placeholder="t('edit.urlPlaceholder')"
            />
            <span class="url-hint">{{ t('edit.internalUrlHint') }}</span>
            <span class="url-hint">{{ t('edit.singleUrlHint') }}</span>
          </div>

          <div class="field">
            <label class="field-label">{{ t('edit.openType') }}</label>
            <select v-model="form.openType" class="input select w-short">
              <option value="_blank">{{ t('edit.openBlank') }}</option>
              <option value="_self">{{ t('edit.openSelf') }}</option>
            </select>
          </div>

          <label class="more-toggle">
            <input v-model="showMore" type="checkbox" />
            <span>{{ t('edit.moreOptions') }}</span>
          </label>

          <div v-if="showMore" class="field">
            <label class="field-label">{{ t('edit.group') }}</label>
            <select v-model="form.groupId" class="input select w-short">
              <option v-for="g in groups" :key="g.id" :value="g.id">{{ g.name }}</option>
            </select>
          </div>

          <p v-if="error" class="error">{{ error }}</p>
        </div>

        <footer class="modal-foot">
          <button
            v-if="!isCreate"
            type="button"
            class="btn danger"
            :disabled="busy || uploading"
            @click="onDelete"
          >
            {{ t('home.delete') }}
          </button>
          <button type="button" class="btn primary" :disabled="busy || uploading" @click="onSave">
            {{ t('settings.save') }}
          </button>
        </footer>
      </div>

      <AssetGallery
        :open="galleryOpen"
        initial-kind="icons"
        @close="galleryOpen = false"
        @select="onGallerySelect"
      />
    </div>
  </Teleport>
</template>

<style scoped>
.modal-root {
  --w-short: 220px;
  --w-url: 320px;

  position: fixed;
  inset: 0;
  z-index: 1200;
  display: grid;
  place-items: center;
  padding: 24px 16px;
}

.modal-root[data-theme='light'] {
  --modal-bg: #ffffff;
  --modal-text: #1f2937;
  --modal-mute: #6b7280;
  --modal-border: #e5e7eb;
  --modal-surface: #f8fafc;
  --modal-input-bg: #ffffff;
  --modal-preview: #f1f5f9;
  --modal-preview-tile: #ffffff;
  --modal-preview-tile-border: #e2e8f0;
  --modal-preview-tile-text: #1e293b;
  --modal-preview-tile-mute: #64748b;
  --modal-preview-tile-shadow: 0 8px 20px rgba(15, 23, 42, 0.08);
  --modal-hover: #f3f4f6;
  --modal-accent: #6366f1;
  --modal-backdrop: rgba(15, 23, 42, 0.48);
  --modal-shadow: 0 24px 64px rgba(15, 23, 42, 0.22);
  --modal-scrollbar: rgba(100, 116, 139, 0.45);
  --modal-scrollbar-hover: rgba(71, 85, 105, 0.7);
}

.modal-root[data-theme='dark'] {
  --modal-bg: #171b24;
  --modal-text: #f3f4f6;
  --modal-mute: #9ca3af;
  --modal-border: rgba(255, 255, 255, 0.1);
  --modal-surface: #111827;
  --modal-input-bg: #0f1218;
  --modal-preview: rgba(99, 102, 241, 0.16);
  --modal-preview-tile: #0b1220;
  --modal-preview-tile-border: rgba(255, 255, 255, 0.1);
  --modal-preview-tile-text: #ffffff;
  --modal-preview-tile-mute: rgba(255, 255, 255, 0.72);
  --modal-preview-tile-shadow: none;
  --modal-hover: rgba(255, 255, 255, 0.06);
  --modal-accent: #818cf8;
  --modal-backdrop: rgba(2, 6, 23, 0.62);
  --modal-shadow: 0 24px 64px rgba(0, 0, 0, 0.5);
  --modal-scrollbar: rgba(148, 163, 184, 0.35);
  --modal-scrollbar-hover: rgba(203, 213, 225, 0.55);
}

.modal-backdrop {
  position: absolute;
  inset: 0;
  background: var(--modal-backdrop);
}

.modal {
  position: relative;
  width: min(560px, 100%);
  max-height: min(92vh, 820px);
  display: flex;
  flex-direction: column;
  overflow: hidden;
  background: var(--modal-bg);
  color: var(--modal-text);
  border-radius: 14px;
  border: 1px solid var(--modal-border);
  box-shadow: var(--modal-shadow);
}

.modal-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 16px 18px 12px;
  border-bottom: 1px solid var(--modal-border);
  flex-shrink: 0;
}

.modal-head h2 {
  margin: 0;
  font-size: 17px;
  font-weight: 600;
  color: var(--modal-text);
}

.close-btn {
  width: 32px;
  height: 32px;
  display: grid;
  place-items: center;
  border: 0;
  border-radius: 8px;
  background: transparent;
  color: var(--modal-mute);
  cursor: pointer;
}

.close-btn:hover {
  background: var(--modal-hover);
  color: var(--modal-text);
}

.modal-body {
  padding: 14px 18px;
  display: flex;
  flex-direction: column;
  gap: 10px;
  overflow: auto;
  min-height: 0;
  scrollbar-width: thin;
  scrollbar-color: var(--modal-scrollbar) transparent;
}

.modal-body::-webkit-scrollbar {
  width: 6px;
  height: 6px;
}

.modal-body::-webkit-scrollbar-track {
  background: transparent;
}

.modal-body::-webkit-scrollbar-thumb {
  background: var(--modal-scrollbar);
  border-radius: 999px;
}

.modal-body::-webkit-scrollbar-thumb:hover {
  background: var(--modal-scrollbar-hover);
}

.preview {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 8px;
  padding: 16px 12px;
  border-radius: 12px;
  background: var(--modal-preview);
  border: 1px solid var(--modal-border);
}

.preview-tile {
  width: 64px;
  height: 64px;
  border-radius: 14px;
  background: var(--modal-preview-tile);
  color: var(--modal-preview-tile-text);
  display: grid;
  place-items: center;
  overflow: hidden;
  border: 1px solid var(--modal-preview-tile-border);
  box-shadow: var(--modal-preview-tile-shadow);
  padding: 0;
}

.preview-tile.is-text {
  padding: 6px 5px;
  align-content: center;
}

.preview-img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.preview-text-block {
  width: 100%;
  height: 100%;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 0;
  text-align: center;
  overflow: hidden;
}

.preview-text-main {
  font-size: 13px;
  font-weight: 700;
  line-height: 1.05;
  word-break: break-word;
  max-width: 100%;
}

.preview-text-desc {
  font-size: 9px;
  font-weight: 400;
  line-height: 1.05;
  margin-top: 5px;
  color: var(--modal-preview-tile-mute);
  opacity: 1;
  word-break: break-word;
  max-width: 100%;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}

.preview-meta {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 2px;
  max-width: 100%;
  text-align: center;
}

.preview-title {
  font-size: 14px;
  font-weight: 600;
  color: var(--modal-text);
  word-break: break-word;
}

.field {
  display: flex;
  flex-direction: column;
  gap: 6px;
  min-width: 0;
}

.w-short {
  width: var(--w-short);
  max-width: 100%;
  flex: 0 0 auto;
}

.w-url {
  width: var(--w-url);
  max-width: 100%;
}

.field-row {
  display: flex;
  flex-wrap: wrap;
  gap: 12px;
  align-items: flex-start;
}

.field-row > .field {
  flex: 0 0 auto;
}

.field-row.single {
  display: flex;
}

.field-label {
  font-size: 13px;
  color: var(--modal-text);
  font-weight: 500;
}

.url-hint {
  color: var(--modal-mute);
  font-size: 11px;
  line-height: 1.4;
}

.field-label-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
}

.field-link {
  font-size: 12px;
  color: var(--modal-accent);
  text-decoration: none;
  white-space: nowrap;
}

.field-link:hover {
  text-decoration: underline;
}

.iconify-row {
  display: flex;
  align-items: center;
  gap: 10px;
}

.iconify-lib {
  flex-shrink: 0;
  font-size: 13px;
}

.req {
  color: #ef4444;
}

.input {
  height: 38px;
  border-radius: 8px;
  border: 1px solid var(--modal-border);
  background: var(--modal-input-bg);
  color: var(--modal-text);
  padding: 0 12px;
  font-size: 14px;
  outline: none;
  box-sizing: border-box;
}

.input-with-count {
  position: relative;
  display: block;
  box-sizing: border-box;
}

.input-with-count .input {
  width: 100%;
  padding-right: 58px;
}

.inline-counter {
  position: absolute;
  right: 10px;
  top: 50%;
  transform: translateY(-50%);
  font-size: 11px;
  color: var(--modal-mute);
  pointer-events: none;
}

.input::placeholder {
  color: var(--modal-mute);
}

.input:focus {
  border-color: var(--modal-accent);
  box-shadow: 0 0 0 3px color-mix(in srgb, var(--modal-accent) 22%, transparent);
}

.select {
  cursor: pointer;
}

.counter {
  font-size: 11px;
  color: var(--modal-mute);
  text-align: right;
}

.seg {
  display: inline-flex;
  border: 1px solid var(--modal-border);
  border-radius: 8px;
  overflow: hidden;
  width: fit-content;
  background: var(--modal-input-bg);
}

.seg button {
  border: 0;
  background: transparent;
  color: var(--modal-mute);
  padding: 7px 14px;
  font-size: 13px;
  cursor: pointer;
}

.seg button + button {
  border-left: 1px solid var(--modal-border);
}

.seg button.active {
  background: var(--modal-accent);
  color: #fff;
}

.image-row {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-wrap: wrap;
}

.image-input {
  width: var(--w-url);
  max-width: 100%;
  flex: 0 1 auto;
}

.action-btn {
  display: inline-flex;
  align-items: center;
  gap: 5px;
  height: 38px;
  padding: 0 12px;
  border-radius: 8px;
  border: 1px solid var(--modal-accent);
  background: transparent;
  color: var(--modal-accent);
  font-size: 13px;
  white-space: nowrap;
  cursor: pointer;
  flex-shrink: 0;
}

.action-btn:hover {
  background: color-mix(in srgb, var(--modal-accent) 12%, transparent);
}

.action-btn:disabled {
  opacity: 0.55;
  cursor: default;
}

.hidden-file {
  display: none;
}

.more-toggle {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 13px;
  color: var(--modal-mute);
  cursor: pointer;
  user-select: none;
}

.error {
  margin: 0;
  font-size: 13px;
  color: #ef4444;
}

.modal-foot {
  display: flex;
  justify-content: flex-end;
  gap: 10px;
  padding: 12px 18px 16px;
  border-top: 1px solid var(--modal-border);
  background: var(--modal-bg);
  flex-shrink: 0;
}

.btn {
  border: 0;
  border-radius: 8px;
  padding: 8px 16px;
  font-size: 13px;
  font-weight: 600;
  cursor: pointer;
}

.btn:disabled {
  opacity: 0.55;
  cursor: default;
}

.btn.danger {
  background: #ef4444;
  color: #fff;
}

.btn.primary {
  background: var(--modal-accent);
  color: #fff;
}

@media (max-width: 640px) {
  .field-row {
    flex-direction: column;
  }

  .w-short,
  .w-url {
    width: 100%;
  }

  .seg {
    width: 100%;
  }

  .seg button {
    flex: 1;
  }

  .action-btn {
    flex: 1;
    justify-content: center;
  }
}
</style>
