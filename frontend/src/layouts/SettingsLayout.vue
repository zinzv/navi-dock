<script setup lang="ts">
import { RouterLink, RouterView } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { Icon } from '@iconify/vue'
import ThemeSwitcher from '../components/ThemeSwitcher.vue'
import { useSettingsStore } from '../stores/settings'

const { t } = useI18n()
const settings = useSettingsStore()
</script>

<template>
  <div class="settings-shell">
    <header class="settings-nav">
      <div class="settings-nav-inner">
        <RouterLink class="settings-brand" to="/">
          <img
            v-if="settings.siteIcon"
            class="settings-brand-icon"
            :src="settings.siteIcon"
            alt=""
          />
          <Icon v-else class="settings-brand-mark" icon="mdi:home-outline" width="18" aria-hidden="true" />
          <span>{{ settings.siteTitle || t('app.name') }}</span>
        </RouterLink>

        <div class="settings-nav-title" aria-hidden="true">{{ t('settings.title') }}</div>

        <div class="settings-nav-actions settings-control-capsule">
          <ThemeSwitcher />
          <RouterLink
            class="icon-btn"
            to="/"
            :title="t('settings.navHome')"
            :aria-label="t('settings.navHome')"
          >
            <Icon icon="mdi:view-grid-outline" width="18" />
          </RouterLink>
        </div>
      </div>
    </header>

    <RouterView />
  </div>
</template>
