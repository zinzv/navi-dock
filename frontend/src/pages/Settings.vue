<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { Icon } from '@iconify/vue'
import { useSettingsStore } from '../stores/settings'
import AccountSettings from '../components/AccountSettings.vue'
import GroupManager from '../components/GroupManager.vue'
import AssetLibrary from '../components/AssetLibrary.vue'
import AssetGallery from '../components/AssetGallery.vue'
import type { AppLocale } from '../i18n'
import { exportNavigation, importNavigation } from '../api/settings'
import { fetchAppVersion } from '../api/health'

const version = ref('…')

onMounted(async () => {
  version.value = await fetchAppVersion()
})

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
                class="secondary-btn"
                :disabled="settings.saving"
                @click="iconGalleryOpen = true"
              >
                <Icon icon="mdi:image-multiple-outline" width="15" />
                {{ t('settings.siteIconGallery') }}
              </button>
              <button
                type="button"
                class="primary-btn"
                :disabled="settings.saving"
                @click="iconInput?.click()"
              >
                <Icon icon="mdi:upload" width="15" />
                {{ t('settings.siteIconUpload') }}
              </button>
              <button
                type="button"
                class="danger-ghost-btn"
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

      <div v-if="previewStyle" class="bg-preview-wrap">
        <div class="bg-preview-base" />
        <div class="bg-preview-image" :style="previewStyle" />
        <div class="mini-home">
          <div class="mini-home-head">
            <span class="mini-home-brand"><i />{{ settings.siteTitle || t('app.name') }}</span>
            <span class="mini-home-search" />
          </div>
          <div class="mini-home-cards"><i /><i /><i /></div>
        </div>
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
                class="secondary-btn"
                :disabled="settings.saving"
                @click="bgGalleryOpen = true"
              >
                <Icon icon="mdi:image-multiple-outline" width="15" />
                {{ t('settings.backgroundGallery') }}
              </button>
              <button
                type="button"
                class="primary-btn"
                :disabled="settings.saving"
                @click="fileInput?.click()"
              >
                <Icon icon="mdi:upload" width="15" />
                {{ t('settings.backgroundUpload') }}
              </button>
              <button
                type="button"
                class="danger-ghost-btn"
                :disabled="settings.saving || !settings.backgroundImage"
                @click="onRemoveBackground"
              >
                {{ t('settings.backgroundRemove') }}
              </button>
            </div>
          </div>
        </div>

        <div class="row">
          <span class="label">{{ t('settings.backgroundOpacity') }}</span>
          <div class="range-control">
            <input
              v-model.number="opacityPercent"
              class="opacity-range"
              type="range"
              min="0"
              max="100"
              step="1"
              :style="{ '--range-progress': `${opacityPercent}%` }"
              @input="onOpacityInput"
            />
            <span class="value mono">{{ opacityPercent }}%</span>
          </div>
        </div>
      </div>
    </section>

        <section id="settings-backup" class="card">
          <h2 class="card-title">{{ t('settings.backup') }}</h2>
      <div class="rows">
        <div class="row migration-row">
          <div class="setting-copy">
            <span class="label">{{ t('settings.export') }}</span>
            <span class="hint">{{ t('settings.exportHint') }}</span>
          </div>
          <div>
            <button type="button" class="secondary-btn" :disabled="ioBusy" @click="onExport">
              {{ t('settings.exportAction') }}
            </button>
          </div>
        </div>
        <div class="row migration-row">
          <div class="setting-copy">
            <span class="label">{{ t('settings.import') }}</span>
            <span class="hint">{{ t('settings.importHint') }}</span>
          </div>
          <div>
            <input
              ref="importInput"
              class="hidden-file"
              type="file"
              accept="application/json,.json"
              @change="onImportFile"
            />
            <button type="button" class="secondary-btn restore-btn" :disabled="ioBusy" @click="triggerImport">
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
              <div class="setting-copy">
                <span class="about-product">{{ t('app.name') }}</span>
                <span class="hint">{{ t('app.tagline') }}</span>
              </div>
            </div>
            <div class="row">
              <span class="label">{{ t('settings.version') }}</span>
              <span class="value mono about-version">v{{ version }}</span>
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
  padding: 24px 32px 72px;
  max-width: 1180px;
  margin: 0 auto;
}

.settings-top {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 30px;
}

.settings-body {
  display: grid;
  grid-template-columns: 168px minmax(0, 1fr);
  align-items: flex-start;
  gap: 28px;
}

.settings-side {
  position: sticky;
  top: 88px;
  width: 168px;
}

.side-nav {
  display: flex;
  flex-direction: column;
  gap: 4px;
  padding: 0;
  border: 0;
  background: transparent;
}

.side-nav-item {
  border: 0;
  background: transparent;
  color: rgba(20, 30, 45, 0.62);
  text-align: left;
  min-height: 40px;
  padding: 8px 14px;
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
  color: #5d63e8;
  background: rgba(99, 102, 241, 0.09);
}

.settings-main {
  flex: 1;
  min-width: 0;
}

.section-anchor {
  scroll-margin-top: 96px;
}

#settings-account,
#settings-general,
#settings-background,
#settings-library,
#settings-backup,
#settings-about {
  scroll-margin-top: 96px;
}

.settings h1 {
  font-size: 28px;
  font-weight: 600;
  margin: 0;
}

.saved-tip {
  color: var(--sv-accent);
  font-size: 13px;
}

.card {
  background: var(--sv-surface);
  border: 1px solid var(--sv-border);
  border-radius: 12px;
  padding: 22px 26px;
  margin-bottom: 20px;
  box-shadow: none;
}

.card-title {
  margin: 0 0 10px;
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

.rows {
  display: flex;
  flex-direction: column;
}

.row {
  display: grid;
  grid-template-columns: 160px minmax(0, 1fr);
  align-items: center;
  gap: 16px;
  min-height: 58px;
  padding: 8px 0;
  font-size: 14px;
  border-top: 1px solid color-mix(in srgb, var(--sv-border) 60%, transparent);
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
  color: var(--sv-regular);
  font-size: 15px;
  font-weight: 500;
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
  background: rgba(99, 102, 241, 0.12);
  color: #5b5fef;
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
  display: inline-flex;
  align-items: center;
  gap: 6px;
}

.primary-btn:hover:not(:disabled) {
  background: var(--sv-accent-hover);
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

.secondary-btn,
.danger-ghost-btn {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  height: 36px;
  padding: 0 12px;
  border: 1px solid var(--sv-border);
  border-radius: 8px;
  background: transparent;
  color: var(--sv-text);
  font-size: 13px;
  cursor: pointer;
}

.secondary-btn:hover:not(:disabled) {
  background: var(--sv-surface-hover);
  border-color: color-mix(in srgb, var(--sv-accent) 30%, var(--sv-border));
}

.danger-ghost-btn {
  border-color: transparent;
  color: var(--sv-mute);
}

.danger-ghost-btn:hover:not(:disabled),
.danger-ghost-btn:hover:not(:disabled) {
  color: var(--sv-danger);
  border-color: transparent;
  background: var(--sv-danger-soft);
}

.secondary-btn:disabled,
.danger-ghost-btn:disabled {
  opacity: 0.4;
  cursor: not-allowed;
}

.row-input {
  height: 40px;
  width: 100%;
  border-radius: 8px;
  border: 1px solid var(--sv-border);
  background: var(--sv-input-bg);
  color: var(--sv-text);
  padding: 0 12px;
  outline: none;
  font-size: 14px;
}

.site-title-input {
  width: 300px;
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
  border-color: color-mix(in srgb, var(--sv-accent) 50%, transparent);
  box-shadow: 0 0 0 3px color-mix(in srgb, var(--sv-accent) 8%, transparent);
}

.row-input::placeholder {
  color: color-mix(in srgb, var(--sv-text) 43%, transparent);
}

.bg-preview-wrap {
  position: relative;
  height: 112px;
  margin: 4px 0 18px;
  border-radius: 10px;
  overflow: hidden;
  border: 1px solid var(--sv-border);
}

.bg-preview-base {
  position: absolute;
  inset: 0;
  background: linear-gradient(180deg, #16324a 0%, #1f4a3d 100%);
  filter: saturate(0.82);
}

.bg-preview-image {
  position: absolute;
  inset: 0;
  background-position: center;
  background-size: cover;
  background-repeat: no-repeat;
}

.mini-home {
  position: absolute;
  inset: 0;
  z-index: 2;
  padding: 14px 18px;
  color: rgba(255, 255, 255, 0.78);
}

.mini-home-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.mini-home-brand {
  display: inline-flex;
  align-items: center;
  gap: 5px;
  font-size: 8px;
}

.mini-home-brand i {
  width: 7px;
  height: 7px;
  border: 1px solid currentColor;
  border-radius: 2px;
}

.mini-home-search {
  width: 38%;
  height: 10px;
  border: 1px solid rgba(255, 255, 255, 0.28);
  border-radius: 999px;
  background: rgba(5, 20, 28, 0.18);
}

.mini-home-cards {
  display: flex;
  gap: 9px;
  margin-top: 22px;
}

.mini-home-cards i {
  width: 24px;
  height: 24px;
  border-radius: 5px;
  background: rgba(5, 20, 28, 0.34);
  border: 1px solid rgba(255, 255, 255, 0.12);
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
  height: 16px;
  margin: 0;
}

.range-control {
  display: grid;
  grid-template-columns: minmax(0, 1fr) 48px;
  align-items: center;
  gap: 12px;
  width: min(480px, 100%);
  justify-self: end;
}

.range-control .value {
  text-align: right;
}

.opacity-range::-webkit-slider-runnable-track {
  height: 4px;
  border-radius: 999px;
  background: linear-gradient(
    to right,
    var(--sv-accent) 0 var(--range-progress),
    rgba(20, 30, 40, 0.12) var(--range-progress) 100%
  );
}

.opacity-range::-webkit-slider-thumb {
  width: 15px;
  height: 15px;
  margin-top: -5.5px;
  border: 0;
  border-radius: 50%;
  background: var(--sv-accent);
  -webkit-appearance: none;
}

.setting-copy {
  display: flex;
  flex-direction: column;
  gap: 3px;
}

.migration-row > :last-child {
  justify-self: end;
}

.about-product {
  font-size: 15px;
  font-weight: 600;
}

.about-version {
  font-size: 14px;
}

.row > :last-child:not(:only-child),
.segmented,
.bg-actions {
  justify-self: end;
}

.row > .setting-copy:only-child {
  grid-column: 1 / -1;
}

.settings-main :deep(.card) {
  background: var(--sv-surface);
  border: 1px solid var(--sv-border);
  border-radius: 12px;
  margin-bottom: 20px;
  box-shadow: none;
}

.settings-main :deep(.card-title) {
  font-size: 18px;
  font-weight: 600;
}

.settings-main :deep(.row) {
  min-height: 58px;
  border-color: color-mix(in srgb, var(--sv-border) 60%, transparent);
}

.settings-main :deep(.rows) {
  max-width: 900px;
}

:global(html[data-theme='dark']) .side-nav-item.active {
  color: #8e92ff;
  background: var(--sv-accent-weak);
}

:global(html[data-theme='dark']) .side-nav-item {
  color: var(--sv-mute);
}

:global(html[data-theme='dark']) .row-input {
  background: var(--sv-input-bg);
  border-color: rgba(255, 255, 255, 0.08);
  color: rgba(255, 255, 255, 0.9);
}

:global(html[data-theme='dark']) .segmented button.active {
  color: #8c91ff;
  background: rgba(95, 102, 232, 0.15);
}

:global(html[data-theme='dark']) .segmented button:not(.active) {
  color: var(--sv-mute);
}

:global(html[data-theme='dark']) .opacity-range::-webkit-slider-runnable-track {
  background: linear-gradient(
    to right,
    var(--sv-accent) 0 var(--range-progress),
    rgba(255, 255, 255, 0.14) var(--range-progress) 100%
  );
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
    grid-template-columns: 1fr;
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
    grid-template-columns: 1fr;
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
