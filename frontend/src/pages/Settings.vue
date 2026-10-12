<script setup lang="ts">
import { computed, nextTick, onMounted, onUnmounted, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { useSettingsStore } from '../stores/settings'
import AccountSettings from '../components/AccountSettings.vue'
import GroupManager from '../components/GroupManager.vue'
import AssetGallery from '../components/AssetGallery.vue'
import type { AppLocale } from '../i18n'
import { exportNavigation, importNavigation, validateNavigationImport } from '../api/settings'
import { fetchAppVersion } from '../api/health'

const version = ref('…')
let scrollFrame = 0
let flashTimer: number | undefined

function syncActiveSection() {
  const sections = sideMenus.map(item => document.getElementById(`settings-${item.id}`))
    .filter((el): el is HTMLElement => Boolean(el))
  const offset = window.innerWidth <= 640 ? 152 : 100
  const current = sections.filter(el => el.getBoundingClientRect().top <= offset).at(-1)
  const atBottom = window.scrollY > 0 && window.scrollY + window.innerHeight >= document.documentElement.scrollHeight - 2
  activeSection.value = (atBottom ? sections.at(-1) : current || sections[0])?.id.replace('settings-', '') || 'account'
}

function onScroll() {
  cancelAnimationFrame(scrollFrame)
  scrollFrame = requestAnimationFrame(syncActiveSection)
}

onMounted(async () => {
  window.addEventListener('scroll', onScroll, { passive: true })
  window.addEventListener('resize', onScroll)
  await nextTick()
  syncActiveSection()
  version.value = await fetchAppVersion()
})

onUnmounted(() => {
  window.removeEventListener('scroll', onScroll)
  window.removeEventListener('resize', onScroll)
  cancelAnimationFrame(scrollFrame)
  window.clearTimeout(opacityTimer)
  window.clearTimeout(flashTimer)
})

const { t } = useI18n()
const settings = useSettingsStore()
const siteTitle = ref(settings.siteTitle)
const opacityPercent = ref(Math.round(settings.backgroundOpacity * 100))
const importInput = ref<HTMLInputElement | null>(null)
const iconGalleryOpen = ref(false)
const bgGalleryOpen = ref(false)
const activeSection = ref('account')

const sideMenus = [
  { id: 'account', labelKey: 'settings.account' },
  { id: 'general', labelKey: 'settings.general' },
  { id: 'groups', labelKey: 'settings.groups' },
  { id: 'background', labelKey: 'settings.background' },
  { id: 'backup', labelKey: 'settings.backup' },
  { id: 'about', labelKey: 'settings.about' },
] as const

function scrollToSection(id: string) {
  activeSection.value = id
  const el = document.getElementById(`settings-${id}`)
  el?.scrollIntoView({ behavior: window.matchMedia('(prefers-reduced-motion: reduce)').matches ? 'auto' : 'smooth', block: 'start' })
}
const savedFlash = ref(false)
const ioBusy = ref(false)
const groupsRevision = ref(0)
const ioStatus = ref('')
const ioError = ref('')
let opacityTimer: number | undefined

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

const previewStyle = computed(() => {
  if (!settings.backgroundImage) return undefined
  return {
    backgroundImage: `url(${settings.backgroundImage})`,
    opacity: settings.backgroundOpacity,
  }
})

function flashSaved() {
  savedFlash.value = true
  window.clearTimeout(flashTimer)
  flashTimer = window.setTimeout(() => {
    savedFlash.value = false
  }, 1200)
}

async function setLanguage(lang: AppLocale) {
  if (settings.language === lang) return
  await settings.setLanguage(lang)
  flashSaved()
}

const titleError = ref('')
async function saveTitle() {
  const next = siteTitle.value.trim() || t('app.name')
  siteTitle.value = next
  if (next === settings.siteTitle) return
  titleError.value = ''
  try {
    await settings.save({ site_title: next })
    flashSaved()
  } catch (error) {
    titleError.value = error instanceof Error ? error.message : t('settings.saveFailed')
  }
}

function onOpacityInput() {
  window.clearTimeout(opacityTimer)
  opacityTimer = window.setTimeout(async () => {
    await settings.setBackgroundOpacity(opacityPercent.value / 100)
    flashSaved()
  }, 200)
}

async function onRemoveBackground() {
  await settings.removeHomeBackground()
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
  ioBusy.value = true
  ioError.value = ''
  ioStatus.value = ''
  try {
    const payload: unknown = JSON.parse(await file.text())
    const validation = validateNavigationImport(payload)
    if (validation) throw new Error(t(`settings.${validation}`))
    if (!window.confirm(t('settings.importConfirm'))) return
    const result = await importNavigation(file)
    groupsRevision.value++
    ioStatus.value = t('settings.importDone', {
      groups: result.group_count,
      items: result.item_count,
    })
    flashSaved()
  } catch (err) {
    ioError.value = err instanceof SyntaxError ? t('settings.importInvalid') : err instanceof Error ? err.message : t('settings.importFailed')
  } finally {
    ioBusy.value = false
  }
}
</script>

<template>
  <div class="settings">
    <div class="settings-top">
      <h1>{{ t('settings.title') }}</h1>
      <span class="saved-tip" role="status" aria-live="polite"><template v-if="savedFlash">{{ t('settings.saved') }}</template></span>
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
            :aria-current="activeSection === item.id ? 'location' : undefined"
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
          <p v-if="titleError" class="io-error" role="alert">{{ titleError }}</p>
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
                :aria-label="t('settings.siteTitle')"
                @blur="saveTitle"
                @keydown.enter="saveTitle"
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
                <div class="action-btns">
                  <button
                    type="button"
                    class="secondary-btn"
                    :disabled="settings.saving"
                    @click="iconGalleryOpen = true"
                  >
                    {{ t('settings.changeImage') }}
                  </button>
                  <button
                    v-if="settings.siteIcon"
                    type="button"
                    class="danger-ghost-btn"
                    :disabled="settings.saving"
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
          <GroupManager :key="groupsRevision" @saved="flashSaved" />
        </div>

        <section id="settings-background" class="card">
          <h2 class="card-title">{{ t('settings.background') }}</h2>

          <div class="rows">
            <div class="row">
              <span class="label">{{ t('settings.backgroundImage') }}</span>
              <div class="bg-actions">
                <div v-if="previewStyle" class="bg-preview-wrap" aria-hidden="true">
                  <div class="bg-preview-image" :style="previewStyle" />
                </div>
                <div class="action-btns">
                  <button
                    type="button"
                    class="secondary-btn"
                    :disabled="settings.saving"
                    @click="bgGalleryOpen = true"
                  >
                    {{ t('settings.changeImage') }}
                  </button>
                  <button
                    v-if="settings.backgroundImage"
                    type="button"
                    class="danger-ghost-btn"
                    :disabled="settings.saving"
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
                 :aria-label="t('settings.backgroundOpacity')"
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
              <span class="value mono about-version" :aria-label="t('settings.version')">v{{ version }}</span>
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
  width: min(1440px, calc(100% - 64px));
  padding: 28px 0 64px;
  margin: 0 auto;
}

.settings-top {
  display: flex;
  align-items: center;
  gap: 16px;
  margin-bottom: 24px;
}

.settings h1 {
  font-size: 28px;
  line-height: 36px;
  font-weight: 600;
  margin: 0;
}

.saved-tip {
  font-size: 12px;
  color: var(--sv-accent-text);
  min-height: 18px;
}

.settings-body {
  display: grid;
  grid-template-columns: 200px minmax(0, 960px);
  gap: 32px;
  align-items: start;
}

.settings-side {
  position: sticky;
  top: 88px;
}

.side-nav {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.side-nav-item {
  border: 0;
  background: transparent;
  color: var(--sv-mute);
  text-align: left;
  min-height: 40px;
  padding: 8px 14px;
  border-radius: 8px;
  font-size: 14px;
  cursor: pointer;
}

.side-nav-item:hover {
  background: var(--sv-surface-hover);
  color: var(--sv-text);
}

.side-nav-item.active {
  color: var(--sv-accent-text);
  background: var(--sv-accent-weak);
}

.settings-main {
  min-width: 0;
}

.section-anchor, .card {
  scroll-margin-top: 88px;
}

.card, .settings-main :deep(.card) {
  background: var(--sv-surface);
  border: 1px solid var(--sv-border);
  border-radius: 12px;
  padding: 20px 24px 12px;
  margin-bottom: 20px;
  box-shadow: none;
}

.card-title, .settings-main :deep(.card-title) {
  margin: 0 0 8px;
  font-size: 18px;
  line-height: 26px;
  font-weight: 600;
  color: var(--sv-text);
}

.settings-main :deep(.card-header .card-title) {
  margin: 0;
}

.rows {
  display: flex;
  flex-direction: column;
}

.row {
  display: grid;
  grid-template-columns: minmax(0, 1fr) minmax(240px, 300px);
  gap: 20px;
  align-items: center;
  font-size: 14px;
}

.row, .settings-main :deep(.row) {
  box-sizing: border-box;
  min-height: 60px;
  padding: 10px 0;
  border-top: 1px solid var(--sv-border);
}

.rows > .row:first-child, .settings-main :deep(.rows > .row:first-child) {
  border-top: 0;
}

.settings-main :deep(.group-row) {
  min-height: 56px;
  padding: 8px 0;
}

.row > :last-child {
  justify-self: end;
}

.label {
  font-size: 14px;
  font-weight: 500;
  color: var(--sv-text);
}

.hint {
  font-size: 12px;
  line-height: 1.5;
  color: var(--sv-mute);
}

.setting-copy {
  display: flex;
  flex-direction: column;
  gap: 4px;
  min-width: 0;
}

.value {
  font-size: 13px;
  color: var(--sv-mute);
}

.mono {
  font-family: ui-monospace, SFMono-Regular, Menlo, Consolas, monospace;
}

.segmented {
  display: inline-flex;
  width: 240px;
  height: 36px;
  border: 1px solid var(--sv-border);
  border-radius: 8px;
  overflow: hidden;
}

.segmented button {
  flex: 1;
  border: 0;
  background: transparent;
  color: var(--sv-mute);
  font-size: 13px;
  cursor: pointer;
}

.segmented button + button {
  border-left: 1px solid var(--sv-border);
}

.segmented button.active {
  background: var(--sv-accent-weak);
  color: var(--sv-accent-text);
}

.secondary-btn, .danger-ghost-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  height: 36px;
  min-width: 64px;
  padding: 0 12px;
  border: 1px solid var(--sv-border);
  border-radius: 8px;
  background: var(--sv-surface);
  color: var(--sv-text);
  font-size: 13px;
  cursor: pointer;
}

.secondary-btn:hover:not(:disabled) {
  background: var(--sv-surface-hover);
}

.danger-ghost-btn {
  min-width: auto;
  border-color: transparent;
  color: var(--sv-mute);
}

.danger-ghost-btn:hover:not(:disabled) {
  color: var(--sv-danger);
  background: var(--sv-danger-soft);
}

button:disabled {
  opacity: .45;
  cursor: not-allowed;
}

.row-input {
  height: 36px;
  width: 240px;
  max-width: 100%;
  box-sizing: border-box;
  padding: 0 12px;
  border: 1px solid var(--sv-border);
  border-radius: 8px;
  background: var(--sv-input-bg);
  color: var(--sv-text);
  font-size: 14px;
}

.row-input:focus {
  outline: 2px solid var(--sv-accent);
  outline-offset: 2px;
}

.bg-actions, .action-btns {
  display: flex;
  align-items: center;
  gap: 8px;
}

.site-icon-preview {
  width: 32px;
  height: 32px;
  border: 1px solid var(--sv-border);
  border-radius: 6px;
  object-fit: contain;
  background: var(--sv-bg);
}

.bg-preview-wrap {
  width: 64px;
  height: 36px;
  position: relative;
  overflow: hidden;
  border-radius: 6px;
  border: 1px solid var(--sv-border);
  background: var(--sv-bg);
}

.bg-preview-image {
  position: absolute;
  inset: 0;
  background-size: cover;
  background-position: center;
}

.hidden-file {
  display: none;
}

.range-control {
  display: grid;
  grid-template-columns: minmax(0, 1fr) 44px;
  align-items: center;
  gap: 12px;
  width: 300px;
  max-width: 100%;
}

.range-control .value {
  text-align: right;
  font-variant-numeric: tabular-nums;
}

.opacity-range {
  width: 100%;
  margin: 0;
  appearance: none;
  background: transparent;
  height: 24px;
  cursor: pointer;
}

.opacity-range::-webkit-slider-runnable-track {
  height: 4px;
  border-radius: 999px;
  background: linear-gradient(to right, var(--sv-accent) 0 var(--range-progress), var(--sv-border) var(--range-progress) 100%);
}

.opacity-range::-webkit-slider-thumb {
  appearance: none;
  width: 16px;
  height: 16px;
  margin-top: -6px;
  border: 2px solid var(--sv-surface);
  border-radius: 50%;
  background: var(--sv-accent);
}

.opacity-range::-moz-range-track {
  height: 4px;
  border-radius: 999px;
  background: var(--sv-border);
}

.opacity-range::-moz-range-progress {
  height: 4px;
  background: var(--sv-accent);
}

.opacity-range::-moz-range-thumb {
  width: 12px;
  height: 12px;
  border: 2px solid var(--sv-surface);
  border-radius: 50%;
  background: var(--sv-accent);
}

.about-product {
  font-size: 15px;
  font-weight: 600;
}

.about-version {
  padding: 4px 10px;
  border-radius: 6px;
  background: var(--sv-bg);
  white-space: nowrap;
}

.io-status, .io-error {
  margin: 4px 0 12px;
  font-size: 12px;
  color: var(--sv-mute);
}

.io-error {
  color: var(--sv-danger);
}

.settings-main :deep(.primary-btn), .settings-main :deep(.secondary-btn), .settings-main :deep(.logout-btn), .settings-main :deep(.link-btn) {
  height: 36px;
  border-radius: 8px;
}

@media (max-width: 1000px) {
  .settings {
    width: calc(100% - 40px);
  }

  .settings-body {
    grid-template-columns: 144px minmax(0, 1fr);
    gap: 24px;
  }

  .row {
    grid-template-columns: minmax(0, 1fr) minmax(180px, 240px);
  }

}

@media (max-width: 640px) {
  .settings {
    width: calc(100% - 32px);
    padding-top: 20px;
  }

  .settings-top {
    margin-bottom: 16px;
  }

  .settings-body {
    grid-template-columns: 1fr;
    gap: 16px;
  }

  .settings-side {
    top: 64px;
    z-index: 30;
    background: var(--sv-bg);
    padding: 8px 0;
  }

  .side-nav {
    flex-direction: row;
    overflow-x: auto;
  }

  .side-nav-item {
    flex-shrink: 0;
    text-align: center;
    padding: 8px 12px;
  }

  .section-anchor, .card {
    scroll-margin-top: 136px;
  }

  .card, .settings-main :deep(.card) {
    padding: 16px 16px 8px;
    margin-bottom: 16px;
  }

  .row {
    grid-template-columns: 1fr;
    gap: 8px;
  }

  .row > :last-child {
    justify-self: start;
  }

  .row-input, .segmented, .range-control {
    width: 100%;
  }

  .bg-actions {
    flex-wrap: wrap;
  }

  .migration-row, #settings-about .row {
    grid-template-columns: minmax(0, 1fr) auto;
    gap: 12px;
  }

}
</style>
