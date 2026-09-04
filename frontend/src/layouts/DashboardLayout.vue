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
const searchInput = ref<HTMLInputElement | null>(null)

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

function onGlobalKeydown(e: KeyboardEvent) {
  if (e.key !== '/' || e.ctrlKey || e.metaKey || e.altKey) return
  const target = e.target as HTMLElement | null
  if (
    target instanceof HTMLInputElement ||
    target instanceof HTMLTextAreaElement ||
    target instanceof HTMLSelectElement ||
    target?.isContentEditable
  ) {
    return
  }
  e.preventDefault()
  searchInput.value?.focus()
}

onMounted(() => {
  document.addEventListener('click', closeMenus)
  document.addEventListener('keydown', onGlobalKeydown)
})
onUnmounted(() => {
  document.removeEventListener('click', closeMenus)
  document.removeEventListener('keydown', onGlobalKeydown)
})
</script>

<template>
  <div class="app-shell" @click="closeMenus">
    <div v-if="bgStyle" class="app-shell-bg" :style="bgStyle" />
    <SceneDecor />

    <header class="header">
      <div class="header-inner">
        <RouterLink class="brand" to="/">
          <img
            v-if="settings.siteIcon"
            class="brand-icon"
            :src="settings.siteIcon"
            alt=""
          />
        <Icon v-else class="brand-mark" icon="mdi:home-outline" width="18" aria-hidden="true" />
          <span class="brand-name">{{ settings.siteTitle || t('app.name') }}</span>
        </RouterLink>

        <div class="search">
          <span class="search-provider" title="Google">
            <Icon icon="mdi:google" width="18" aria-hidden="true" />
          </span>
          <input
            id="webSearch"
            ref="searchInput"
            v-model="searchQuery"
            type="search"
            autocomplete="off"
            :placeholder="t('search.placeholder')"
            @keydown="onSearchKeydown"
            @click.stop
          />
          <kbd class="search-shortcut" aria-hidden="true">/</kbd>
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
      </div>
    </header>

    <main class="main">
      <RouterView />
    </main>
  </div>
</template>
