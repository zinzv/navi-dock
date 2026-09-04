<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { Icon } from '@iconify/vue'
import type { NetworkMode } from '../api/settings'
import { useSettingsStore } from '../stores/settings'

const { t } = useI18n()
const settings = useSettingsStore()

const order: NetworkMode[] = ['auto', 'internal', 'external']

const meta = computed(() => {
  // Outline icons to match ThemeSwitcher / settings grid buttons
  if (settings.networkMode === 'internal') {
    return { icon: 'mdi:lan-connect', label: t('network.internal'), badge: '内' }
  }
  if (settings.networkMode === 'external') {
    return { icon: 'mdi:cloud-outline', label: t('network.external'), badge: '外' }
  }
  return { icon: 'mdi:web', label: t('network.auto'), badge: 'A' }
})

async function cycleNetwork() {
  const idx = order.indexOf(settings.networkMode)
  const next = order[(idx + 1) % order.length]
  await settings.setNetworkMode(next)
}
</script>

<template>
  <button
    class="icon-btn"
    :class="{ active: settings.networkMode !== 'auto' }"
    type="button"
    :title="`${t('network.title')}: ${meta.label}`"
    :aria-label="`${t('network.title')}: ${meta.label}`"
    @click.stop="cycleNetwork"
  >
    <Icon :icon="meta.icon" width="18" />
    <span class="net-badge">{{ meta.badge }}</span>
  </button>
</template>
