<script setup lang="ts">
import { RouterLink, RouterView } from 'vue-router'
import ThemeSwitcher from '../components/ThemeSwitcher.vue'
import { useSettingsStore } from '../stores/settings'

const settings = useSettingsStore()
</script>

<template>
  <div class="settings-shell">
    <header class="settings-nav">
      <RouterLink class="settings-brand" to="/">
        <img
          v-if="settings.siteIcon"
          class="settings-brand-icon"
          :src="settings.siteIcon"
          alt=""
        />
        <span v-else class="settings-brand-mark">🏠</span>
        <span>{{ settings.siteTitle || t('app.name') }}</span>
      </RouterLink>

      <div class="settings-nav-actions">
        <ThemeSwitcher />
      </div>
    </header>

    <RouterView />
  </div>
</template>
