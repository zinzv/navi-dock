import { defineStore } from 'pinia'
import { computed, ref, watch } from 'vue'
import {
  clearBackground,
  clearSiteIcon,
  fetchSettings,
  updateSettings,
  uploadAsset,
  uploadBackground,
  type AppSettings,
  type ThemeMode,
} from '../api/settings'
import i18n, { type AppLocale } from '../i18n'

const BRAND_DEFAULTS = new Set(['导航', 'NaviDock', 'NaviDot', 'naviDock', 'NavVerse'])
const DEFAULT_FAVICON = '/favicon.png'

function brandName(locale: AppLocale) {
  return locale === 'en' ? 'NaviDock' : '导航'
}

function resolveTheme(theme: ThemeMode): 'light' | 'dark' {
  if (theme === 'system') {
    return window.matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'light'
  }
  return theme
}

function applyTheme(theme: ThemeMode) {
  const resolved = resolveTheme(theme)
  document.documentElement.dataset.theme = resolved
  document.documentElement.style.colorScheme = resolved
}

function applyDocumentLang(locale: AppLocale) {
  document.documentElement.lang = locale === 'en' ? 'en' : 'zh-CN'
}

function applyFavicon(href: string) {
  let link = document.querySelector("link[rel*='icon']") as HTMLLinkElement | null
  if (!link) {
    link = document.createElement('link')
    link.rel = 'icon'
    document.head.appendChild(link)
  }
  link.href = href || DEFAULT_FAVICON
}

function clampOpacity(v: number) {
  if (Number.isNaN(v)) return 0.35
  return Math.min(1, Math.max(0, v))
}

export const useSettingsStore = defineStore('settings', () => {
  const loaded = ref(false)
  const saving = ref(false)
  const siteTitle = ref(brandName('zh'))
  const siteIcon = ref('')
  const language = ref<AppLocale>('zh')
  const theme = ref<ThemeMode>('dark')
  const networkMode = ref<'auto' | 'internal' | 'external'>('internal')
  const backgroundImage = ref('')
  const backgroundOpacity = ref(0.35)

  const resolvedTheme = computed(() => resolveTheme(theme.value))
  const displayName = computed(() => brandName(language.value))

  function applyFromServer(data: AppSettings) {
    language.value = (data.language === 'en' ? 'en' : 'zh') as AppLocale
    const incomingTitle = data.site_title?.trim()
    if (!incomingTitle || BRAND_DEFAULTS.has(incomingTitle)) {
      siteTitle.value = brandName(language.value)
    } else {
      siteTitle.value = incomingTitle
    }
    siteIcon.value = data.site_icon || ''
    theme.value = data.theme || 'dark'
    networkMode.value = data.network_mode || 'internal'
    backgroundImage.value = data.background_image || ''
    backgroundOpacity.value = clampOpacity(data.background_opacity ?? 0.35)
    applyLocal()
  }

  function applyLocal() {
    i18n.global.locale.value = language.value
    applyDocumentLang(language.value)
    document.title = siteTitle.value || brandName(language.value)
    applyFavicon(siteIcon.value)
    applyTheme(theme.value)
  }

  async function load() {
    try {
      const data = await fetchSettings()
      applyFromServer(data)
    } catch {
      applyLocal()
    } finally {
      loaded.value = true
    }
  }

  async function save(partial?: Partial<AppSettings>) {
    saving.value = true
    try {
      const payload: AppSettings = {
        site_title: partial?.site_title ?? siteTitle.value,
        site_icon: partial?.site_icon ?? siteIcon.value,
        language: partial?.language ?? language.value,
        theme: partial?.theme ?? theme.value,
        network_mode: partial?.network_mode ?? networkMode.value,
        background_image: partial?.background_image ?? backgroundImage.value,
        background_opacity: partial?.background_opacity ?? backgroundOpacity.value,
        clear_background: partial?.clear_background,
        clear_site_icon: partial?.clear_site_icon,
      }
      const data = await updateSettings(payload)
      applyFromServer(data)
      return data
    } finally {
      saving.value = false
    }
  }

  async function setTheme(next: ThemeMode) {
    theme.value = next
    applyTheme(next)
    await save({ theme: next })
  }

  async function setLanguage(next: AppLocale) {
    const prevTitle = siteTitle.value.trim()
    language.value = next
    i18n.global.locale.value = next
    applyDocumentLang(next)
    if (!prevTitle || BRAND_DEFAULTS.has(prevTitle)) {
      siteTitle.value = brandName(next)
      document.title = siteTitle.value
      await save({ language: next, site_title: siteTitle.value })
      return
    }
    await save({ language: next })
  }

  async function setNetworkMode(next: 'auto' | 'internal' | 'external') {
    networkMode.value = next
    await save({ network_mode: next })
  }

  async function setBackgroundOpacity(next: number) {
    backgroundOpacity.value = clampOpacity(next)
    await save({ background_opacity: backgroundOpacity.value })
  }

  async function uploadHomeBackground(file: File) {
    saving.value = true
    try {
      const data = await uploadBackground(file)
      applyFromServer(data)
      return data
    } finally {
      saving.value = false
    }
  }

  async function removeHomeBackground() {
    saving.value = true
    try {
      const data = await clearBackground()
      applyFromServer(data)
      return data
    } finally {
      saving.value = false
    }
  }

  async function uploadBrandIcon(file: File) {
    saving.value = true
    try {
      // Site icon shares the icons gallery (no separate /assets/site).
      const uploaded = await uploadAsset('icons', file)
      return await save({ site_icon: uploaded.url })
    } finally {
      saving.value = false
    }
  }

  async function removeBrandIcon() {
    saving.value = true
    try {
      const data = await clearSiteIcon()
      applyFromServer(data)
      return data
    } finally {
      saving.value = false
    }
  }

  watch(siteTitle, (v) => {
    document.title = v || brandName(language.value)
  })

  watch(siteIcon, (v) => {
    applyFavicon(v)
  })

  if (typeof window !== 'undefined') {
    window.matchMedia('(prefers-color-scheme: dark)').addEventListener('change', () => {
      if (theme.value === 'system') applyTheme('system')
    })
  }

  return {
    loaded,
    saving,
    siteTitle,
    siteIcon,
    language,
    theme,
    networkMode,
    backgroundImage,
    backgroundOpacity,
    resolvedTheme,
    displayName,
    load,
    save,
    setTheme,
    setLanguage,
    setNetworkMode,
    setBackgroundOpacity,
    uploadHomeBackground,
    removeHomeBackground,
    uploadBrandIcon,
    removeBrandIcon,
    applyLocal,
  }
})
