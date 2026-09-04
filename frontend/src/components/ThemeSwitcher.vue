<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { Icon } from '@iconify/vue'
import type { ThemeMode } from '../api/settings'
import { useSettingsStore } from '../stores/settings'

const { t } = useI18n()
const settings = useSettingsStore()

const order: ThemeMode[] = ['system', 'light', 'dark']

const meta = computed(() => {
  if (settings.theme === 'light') {
    return { icon: 'mdi:white-balance-sunny', label: t('theme.light') }
  }
  if (settings.theme === 'dark') {
    return { icon: 'mdi:moon-waning-crescent', label: t('theme.dark') }
  }
  return { icon: 'mdi:monitor', label: t('theme.system') }
})

async function cycleTheme() {
  const idx = order.indexOf(settings.theme)
  const next = order[(idx + 1) % order.length]
  await settings.setTheme(next)
}
</script>

<template>
  <button
    class="icon-btn"
    :class="{ active: settings.theme !== 'system' }"
    type="button"
    :title="`${t('theme.title')}: ${meta.label}`"
    :aria-label="`${t('theme.title')}: ${meta.label}`"
    @click.stop="cycleTheme"
  >
    <Icon :icon="meta.icon" width="18" />
  </button>
</template>
