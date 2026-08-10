<script setup lang="ts">
import { computed, onMounted, onUnmounted, provide, ref } from 'vue'
import { RouterLink, RouterView } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { Icon } from '@iconify/vue'
import ThemeSwitcher from '../components/ThemeSwitcher.vue'
import NetworkSwitcher from '../components/NetworkSwitcher.vue'
import SceneDecor from '../components/SceneDecor.vue'
import { useSettingsStore } from '../stores/settings'

const { t } = useI18n()
const settings = useSettingsStore()
const searchQuery = ref('')

provide('homeSearchQuery', searchQuery)

const bgStyle = computed(() => {
  if (!settings.backgroundImage) return undefined
  return {
    backgroundImage: `url(${settings.backgroundImage})`,
    opacity: settings.backgroundOpacity,
  }
})

function closeMenus() {
  document.dispatchEvent(new CustomEvent('navidock:close-menus'))
}

function openGoogleSearch() {
  const q = searchQuery.value.trim()
  if (!q) {
    window.open('https://www.google.com/', '_blank', 'noopener,noreferrer')
    return
  }
  const url = `https://www.google.com/search?q=${encodeURIComponent(q)}`
  window.open(url, '_blank', 'noopener,noreferrer')
}

function onSearchKeydown(e: KeyboardEvent) {
  if (e.key === 'Enter') {
    e.preventDefault()
    openGoogleSearch()
  }
}

onMounted(() => document.addEventListener('click', closeMenus))
onUnmounted(() => document.removeEventListener('click', closeMenus))
</script>

<template>
  <div class="app-shell" @click="closeMenus">
    <div v-if="bgStyle" class="app-shell-bg" :style="bgStyle" />
    <SceneDecor />

    <header class="header">
      <RouterLink class="brand" to="/">
        <img
          v-if="settings.siteIcon"
          class="brand-icon"
          :src="settings.siteIcon"
          alt=""
        />
        <span v-else class="brand-mark" aria-hidden="true">🏠</span>
        <span class="brand-name">{{ settings.siteTitle || t('app.name') }}</span>
      </RouterLink>

      <div class="search">
        <svg class="g-logo" viewBox="0 0 24 24" aria-hidden="true">
          <path fill="#4285F4" d="M22.56 12.25c0-.78-.07-1.53-.2-2.25H12v4.26h5.92a5.06 5.06 0 0 1-2.2 3.32v2.77h3.57c2.08-1.92 3.28-4.74 3.28-8.1z" />
          <path fill="#34A853" d="M12 23c2.97 0 5.46-.98 7.28-2.66l-3.57-2.77c-.98.66-2.23 1.06-3.71 1.06-2.86 0-5.29-1.93-6.16-4.53H2.18v2.84C3.99 20.53 7.7 23 12 23z" />
          <path fill="#FBBC05" d="M5.84 14.09c-.22-.66-.35-1.36-.35-2.09s.13-1.43.35-2.09V7.07H2.18C1.43 8.55 1 10.22 1 12s.43 3.45 1.18 4.93l2.85-2.22.81-.62z" />
          <path fill="#EA4335" d="M12 5.38c1.62 0 3.06.56 4.21 1.64l3.15-3.15C17.45 2.09 14.97 1 12 1 7.7 1 3.99 3.47 2.18 7.07l3.66 2.84c.87-2.6 3.3-4.53 6.16-4.53z" />
        </svg>
        <input
          id="webSearch"
          v-model="searchQuery"
          type="search"
          autocomplete="off"
          :placeholder="t('search.placeholder')"
          @keydown="onSearchKeydown"
          @click.stop
        />
        <button
          type="button"
          class="search-submit"
          :title="t('search.google')"
          :aria-label="t('search.google')"
          @click.stop="openGoogleSearch"
        >
          <Icon icon="mdi:magnify" width="20" />
        </button>
      </div>

      <div class="header-actions" @click.stop>
        <NetworkSwitcher />
        <ThemeSwitcher />
        <RouterLink
          class="icon-btn"
          to="/settings"
          :title="t('settings.title')"
          :aria-label="t('settings.title')"
        >
          <Icon icon="mdi:view-grid-outline" width="18" />
        </RouterLink>
      </div>
    </header>

    <main class="main">
      <RouterView />
    </main>
  </div>
</template>
