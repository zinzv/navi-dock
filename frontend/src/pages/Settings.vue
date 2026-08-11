<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { useSettingsStore } from '../stores/settings'
import AccountSettings from '../components/AccountSettings.vue'
import GroupManager from '../components/GroupManager.vue'
import AssetLibrary from '../components/AssetLibrary.vue'
import AssetGallery from '../components/AssetGallery.vue'
import type { AppLocale } from '../i18n'
import { exportNavigation, importNavigation } from '../api/settings'

const VERSION = '0.1.0'

const { t } = useI18n()
const settings = useSettingsStore()
const siteTitle = ref(settings.siteTitle)
const opacityPercent = ref(Math.round(settings.backgroundOpacity * 100))
const fileInput = ref<HTMLInputElement | null>(null)
const iconInput = ref<HTMLInputElement | null>(null)
const importInput = ref<HTMLInputElement | null>(null)
const iconGalleryOpen = ref(false)
const bgGalleryOpen = ref(false)
const activeSection = ref('account')

const sideMenus = [
  { id: 'account', labelKey: 'settings.account' },
  { id: 'general', labelKey: 'settings.general' },
  { id: 'groups', labelKey: 'settings.groups' },
  { id: 'library', labelKey: 'settings.library' },
  { id: 'background', labelKey: 'settings.background' },
  { id: 'backup', labelKey: 'settings.backup' },
  { id: 'about', labelKey: 'settings.about' },
] as const

function scrollToSection(id: string) {
  activeSection.value = id
  const el = document.getElementById(`settings-${id}`)
  el?.scrollIntoView({ behavior: 'smooth', block: 'start' })
}
const savedFlash = ref(false)
const ioBusy = ref(false)
const ioStatus = ref('')
const ioError = ref('')
let titleTimer: number | undefined
let opacityTimer: number | undefined
let skipFlash = true

watch(
  () => settings.siteTitle,
  (v) => {
    siteTitle.value = v
  },
)

watch(
  () => settings.backgroundOpacity,
  (v) => {
    opacityPercent.value = Math.round(v * 100)
  },
)

watch(
  () => [settings.theme, settings.networkMode, settings.backgroundImage, settings.siteIcon],
  () => {
    if (skipFlash) {
      skipFlash = false
      return
    }
    flashSaved()
  },
)

const previewStyle = computed(() => {
  if (!settings.backgroundImage) return undefined
  return {
    backgroundImage: `url(${settings.backgroundImage})`,
    opacity: settings.backgroundOpacity,
  }
})

function flashSaved() {
  savedFlash.value = true
  window.setTimeout(() => {
    savedFlash.value = false
  }, 1200)
}

async function setLanguage(lang: AppLocale) {
  if (settings.language === lang) return
  await settings.setLanguage(lang)
  flashSaved()
}

function queueSaveTitle() {
  window.clearTimeout(titleTimer)
  titleTimer = window.setTimeout(async () => {
    const next = siteTitle.value.trim() || t('app.name')
    siteTitle.value = next
    if (next === settings.siteTitle) return
    await settings.save({ site_title: next })
    flashSaved()
  }, 400)
}

function onOpacityInput() {
  window.clearTimeout(opacityTimer)
  opacityTimer = window.setTimeout(async () => {
    await settings.setBackgroundOpacity(opacityPercent.value / 100)
    flashSaved()
  }, 200)
}

async function onPickFile(e: Event) {
  const input = e.target as HTMLInputElement
  const file = input.files?.[0]
  if (!file) return
  await settings.uploadHomeBackground(file)
  input.value = ''
  flashSaved()
}

async function onRemoveBackground() {
  await settings.removeHomeBackground()
  flashSaved()
}

async function onPickIcon(e: Event) {
  const input = e.target as HTMLInputElement
  const file = input.files?.[0]
  if (!file) return
  await settings.uploadBrandIcon(file)
  input.value = ''
  flashSaved()
}

async function onRemoveIcon() {
  await settings.removeBrandIcon()
  flashSaved()
}

async function onGallerySelect(url: string) {
  await settings.save({ site_icon: url })
  flashSaved()
}

async function onBgGallerySelect(url: string) {
  await settings.save({ background_image: url })
  flashSaved()
}

async function onExport() {
  if (ioBusy.value) return
  ioBusy.value = true
  ioError.value = ''
  ioStatus.value = ''
  try {
    await exportNavigation()
    ioStatus.value = t('settings.exportDone')
    flashSaved()
  } catch (e) {
    ioError.value = e instanceof Error ? e.message : t('settings.exportFailed')
  } finally {
    ioBusy.value = false
  }
}

function triggerImport() {
  importInput.value?.click()
}

async function onImportFile(e: Event) {
  const input = e.target as HTMLInputElement
  const file = input.files?.[0]
  input.value = ''
  if (!file || ioBusy.value) return
  if (!window.confirm(t('settings.importConfirm'))) return
  ioBusy.value = true
  ioError.value = ''
  ioStatus.value = ''
  try {
    const result = await importNavigation(file)
    ioStatus.value = t('settings.importDone', {
      groups: result.group_count,
      items: result.item_count,
    })
    flashSaved()
  } catch (err) {
    ioError.value = err instanceof Error ? err.message : t('settings.importFailed')
  } finally {
    ioBusy.value = false
  }
}
</script>

<template>
  <div class="settings">
    <div class="settings-top">
      <h1>{{ t('settings.title') }}</h1>
      <span v-if="savedFlash" class="saved-tip">{{ t('settings.saved') }}</span>
    </div>

    <div class="settings-body">
      <aside class="settings-side" :aria-label="t('settings.title')">
        <nav class="side-nav">
          <button
            v-for="item in sideMenus"
            :key="item.id"
            type="button"
            class="side-nav-item"
            :class="{ active: activeSection === item.id }"
            @click="scrollToSection(item.id)"
          >
            {{ t(item.labelKey) }}
          </button>
        </nav>
      </aside>

      <div class="settings-main">
        <div id="settings-account" class="section-anchor">
          <AccountSettings @saved="flashSaved" />
        </div>

        <section id="settings-general" class="card">
          <h2 class="card-title">{{ t('settings.general') }}</h2>
          <div class="rows">
        <div class="row">
          <span class="label">{{ t('settings.language') }}</span>
          <div class="segmented" role="group" :aria-label="t('settings.language')">
            <button
              type="button"
              :class="{ active: settings.language === 'zh' }"
              @click="setLanguage('zh')"
            >
              {{ t('settings.langZh') }}
            </button>
            <button
              type="button"
              :class="{ active: settings.language === 'en' }"
              @click="setLanguage('en')"
            >
              English
            </button>
          </div>
        </div>
        <div class="row">
          <span class="label">{{ t('settings.siteTitle') }}</span>
          <input
            v-model="siteTitle"
            class="row-input site-title-input"
            type="text"
            maxlength="64"
            :title="t('settings.siteTitleHint')"
            @input="queueSaveTitle"
            @blur="queueSaveTitle"
          />
        </div>
        <div class="row">
          <span class="label">{{ t('settings.siteIcon') }}</span>
          <div class="bg-actions">
            <img
              v-if="settings.siteIcon"
              class="site-icon-preview"
              :src="settings.siteIcon"
              alt=""
            />
            <input
              ref="iconInput"
              class="hidden-file"
              type="file"
              accept="image/png,image/jpeg,image/webp,image/gif,image/svg+xml,image/x-icon,.ico"
              @change="onPickIcon"
            />
            <div class="action-btns">
              <button
                type="button"
                class="primary-btn"
                :disabled="settings.saving"
                @click="iconGalleryOpen = true"
              >
                {{ t('settings.siteIconGallery') }}
              </button>
              <button
                type="button"
                class="primary-btn"
                :disabled="settings.saving"
                @click="iconInput?.click()"
              >
                {{ t('settings.siteIconUpload') }}
              </button>
              <button
                type="button"
                class="primary-btn"
                :disabled="settings.saving || !settings.siteIcon"
                @click="onRemoveIcon"
              >
                {{ t('settings.siteIconRemove') }}
              </button>
            </div>
          </div>
        </div>
      </div>
    </section>

        <div id="settings-groups" class="section-anchor">
          <GroupManager @saved="flashSaved" />
        </div>

        <div id="settings-library" class="section-anchor">
          <AssetLibrary @saved="flashSaved" />
        </div>

        <section id="settings-background" class="card">
          <h2 class="card-title">{{ t('settings.background') }}</h2>

      <div class="bg-preview-wrap">
        <div class="bg-preview-base" />
        <div v-if="previewStyle" class="bg-preview-image" :style="previewStyle" />
        <div v-else class="bg-preview-empty">{{ t('settings.backgroundEmpty') }}</div>
      </div>

      <div class="rows">
        <div class="row">
          <span class="label">{{ t('settings.backgroundImage') }}</span>
          <div class="bg-actions">
            <input
              ref="fileInput"
              class="hidden-file"
              type="file"
              accept="image/png,image/jpeg,image/webp,image/gif"
              @change="onPickFile"
            />
            <div class="action-btns">
              <button
                type="button"
                class="primary-btn"
                :disabled="settings.saving"
                @click="bgGalleryOpen = true"
              >
                {{ t('settings.backgroundGallery') }}
              </button>
              <button
                type="button"
                class="primary-btn"
                :disabled="settings.saving"
                @click="fileInput?.click()"
              >
                {{ t('settings.backgroundUpload') }}
              </button>
              <button
                type="button"
                class="primary-btn"
                :disabled="settings.saving || !settings.backgroundImage"
                @click="onRemoveBackground"
              >
                {{ t('settings.backgroundRemove') }}
              </button>
            </div>
          </div>
        </div>

        <div class="row row-stack">
          <div class="row-inline">
            <span class="label">{{ t('settings.backgroundOpacity') }}</span>
            <span class="value mono">{{ opacityPercent }}%</span>
          </div>
          <input
            v-model.number="opacityPercent"
            class="opacity-range"
            type="range"
            min="0"
            max="100"
            step="1"
            @input="onOpacityInput"
          />
        </div>
      </div>
    </section>

        <section id="settings-backup" class="card">
          <h2 class="card-title">{{ t('settings.backup') }}</h2>
      <div class="rows">
        <div class="row">
          <div class="bg-actions">
            <button type="button" class="primary-btn" :disabled="ioBusy" @click="onExport">
              {{ t('settings.exportAction') }}
            </button>
            <input
              ref="importInput"
              class="hidden-file"
              type="file"
              accept="application/json,.json"
              @change="onImportFile"
            />
            <button type="button" class="primary-btn" :disabled="ioBusy" @click="triggerImport">
              {{ t('settings.importAction') }}
            </button>
          </div>
        </div>
        <p v-if="ioStatus" class="io-status">{{ ioStatus }}</p>
        <p v-if="ioError" class="io-error">{{ ioError }}</p>
      </div>
    </section>

        <section id="settings-about" class="card">
          <h2 class="card-title">{{ t('settings.about') }}</h2>
          <div class="rows">
            <div class="row">
              <span class="label">{{ t('app.name') }}</span>
              <span class="value">{{ t('app.tagline') }}</span>
            </div>
            <div class="row">
              <span class="label">{{ t('settings.version') }}</span>
              <span class="value mono">{{ VERSION }}</span>
            </div>
          </div>
        </section>
      </div>
    </div>

    <AssetGallery
      :open="iconGalleryOpen"
      initial-kind="icons"
      @close="iconGalleryOpen = false"
      @select="onGallerySelect"
    />
    <AssetGallery
      :open="bgGalleryOpen"
      initial-kind="wallpapers"
      @close="bgGalleryOpen = false"
      @select="onBgGallerySelect"
    />
  </div>
</template>

<style scoped>
.settings {
  padding: 28px 32px 48px;
  max-width: 860px;
  margin: 0 auto;
}

.settings-top {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 24px;
}

.settings-body {
  display: flex;
  align-items: flex-start;
  gap: 20px;
}

.settings-side {
  position: sticky;
  top: 72px;
  flex: 0 0 128px;
  width: 128px;
}

.side-nav {
  display: flex;
  flex-direction: column;
  gap: 4px;
  padding: 6px;
  border: 1px solid var(--sv-border);
  border-radius: 10px;
  background: var(--sv-surface);
}

.side-nav-item {
  border: 0;
  background: transparent;
  color: var(--sv-mute);
  text-align: left;
  padding: 8px 12px;
  border-radius: 8px;
  font-size: 13px;
  cursor: pointer;
  transition: background-color 0.15s ease, color 0.15s ease;
}

.side-nav-item:hover {
  color: var(--sv-text);
  background: color-mix(in srgb, var(--sv-accent) 10%, transparent);
}

.side-nav-item.active {
  color: #fff;
  background: var(--sv-accent);
}

.settings-main {
  flex: 1;
  min-width: 0;
}

.section-anchor {
  scroll-margin-top: 72px;
}

#settings-account,
#settings-general,
#settings-background,
#settings-library,
#settings-backup,
#settings-about {
  scroll-margin-top: 72px;
}

.settings h1 {
  font-size: 22px;
  font-weight: 700;
  margin: 0;
}

.saved-tip {
  color: var(--sv-accent);
  font-size: 13px;
}

.card {
  background: var(--sv-surface);
  border: 1px solid var(--sv-border);
  border-radius: 10px;
  padding: 16px 18px 4px;
  margin-bottom: 16px;
}

.card-title {
  margin: 0 0 8px;
  font-size: 15px;
  font-weight: 700;
  color: var(--sv-text);
}

.card-hint {
  margin: 0 0 12px;
  font-size: 12px;
  line-height: 1.5;
  color: var(--sv-mute);
}

.rows {
  display: flex;
  flex-direction: column;
}

.row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
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

.row-inline,
.row-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
}

.label {
  flex-shrink: 0;
  color: var(--sv-text);
}

.hint {
  font-size: 12px;
  color: var(--sv-mute);
}

.value {
  text-align: right;
  color: var(--sv-mute);
  font-size: 13px;
  line-height: 1.4;
  word-break: break-all;
}

.mono {
  font-family: ui-monospace, SFMono-Regular, Menlo, Consolas, monospace;
}

.segmented {
  display: inline-flex;
  width: 180px;
  box-sizing: border-box;
  border: 1px solid var(--sv-border);
  border-radius: 8px;
  overflow: hidden;
}

.segmented button {
  flex: 1;
  background: transparent;
  border: none;
  color: var(--sv-mute);
  padding: 6px 10px;
  font-size: 13px;
  cursor: pointer;
  transition: background-color 0.18s ease-out, color 0.18s ease-out;
}

.segmented button + button {
  border-left: 1px solid var(--sv-border);
}

.segmented button.active {
  background: var(--sv-accent);
  color: #fff;
}

.primary-btn {
  background: var(--sv-accent);
  color: #fff;
  border: none;
  border-radius: 8px;
  padding: 6px 14px;
  font-size: 13px;
  cursor: pointer;
  transition: opacity 0.18s ease-out;
}

.primary-btn:hover:not(:disabled) {
  opacity: 0.9;
}

.primary-btn:disabled,
.link-btn:disabled {
  opacity: 0.5;
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
  transition: background-color 0.18s ease-out;
}

.link-btn:hover:not(:disabled) {
  background: var(--sv-border);
}

.row-input {
  height: 36px;
  width: 100%;
  border-radius: 8px;
  border: 1px solid var(--sv-border);
  background: var(--sv-bg);
  color: var(--sv-text);
  padding: 0 12px;
  outline: none;
  font-size: 14px;
}

.site-title-input {
  width: 180px;
  max-width: 100%;
  box-sizing: border-box;
}

.action-btns {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  flex-wrap: nowrap;
}

.action-btns .primary-btn {
  margin: 0;
}

.site-icon-preview {
  width: 28px;
  height: 28px;
  border-radius: 6px;
  object-fit: contain;
  background: var(--sv-bg);
  border: 1px solid var(--sv-border);
}

.row-input:focus {
  border-color: var(--sv-accent);
}

.bg-preview-wrap {
  position: relative;
  height: 120px;
  margin: 4px 0 8px;
  border-radius: 10px;
  overflow: hidden;
  border: 1px solid var(--sv-border);
}

.bg-preview-base {
  position: absolute;
  inset: 0;
  background: linear-gradient(180deg, #16324a 0%, #1f4a3d 100%);
}

.bg-preview-image {
  position: absolute;
  inset: 0;
  background-position: center;
  background-size: cover;
  background-repeat: no-repeat;
}

.bg-preview-empty {
  position: absolute;
  inset: 0;
  display: grid;
  place-items: center;
  color: rgba(255, 255, 255, 0.75);
  font-size: 13px;
}

.bg-actions {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
  align-items: center;
}

.hidden-file {
  display: none;
}

.opacity-range {
  width: 100%;
  accent-color: var(--sv-accent);
}

.io-status {
  margin: 0 0 12px;
  font-size: 13px;
  color: var(--sv-mute);
}

.io-error {
  margin: 0 0 12px;
  font-size: 13px;
  color: #ef4444;
}

@media (max-width: 640px) {
  .settings {
    padding: 20px 16px 40px;
  }

  .settings-body {
    flex-direction: column;
    gap: 12px;
  }

  .settings-side {
    position: static;
    width: 100%;
    flex-basis: auto;
  }

  .side-nav {
    flex-direction: row;
    flex-wrap: wrap;
  }

  .side-nav-item {
    flex: 1 1 auto;
    text-align: center;
  }

  .row:not(.row-stack) {
    flex-wrap: wrap;
    gap: 8px;
  }

  .value {
    text-align: left;
    width: 100%;
  }

  .segmented {
    width: 100%;
  }

  .segmented button {
    flex: 1;
  }
}
</style>
