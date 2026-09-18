<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import type { NetworkMode } from '../api/settings'
import { useSettingsStore } from '../stores/settings'

const { t } = useI18n()
const settings = useSettingsStore()

const order: NetworkMode[] = ['auto', 'internal', 'external']

const meta = computed(() => {
  if (settings.networkMode === 'internal') {
    return { kind: 'internal' as const, label: t('network.internal') }
  }
  if (settings.networkMode === 'external') {
    return { kind: 'external' as const, label: t('network.external') }
  }
  return { kind: 'auto' as const, label: t('network.auto') }
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
    <!-- 自动：雷达扫描，表示按环境自动探测 -->
    <svg
      v-if="meta.kind === 'auto'"
      class="net-glyph"
      viewBox="1 1 22 22"
      fill="none"
      aria-hidden="true"
    >
      <circle cx="12" cy="12" r="1.65" fill="currentColor" />
      <path
        d="M12 12 L18.35 5.65"
        stroke="currentColor"
        stroke-width="1.85"
        stroke-linecap="round"
      />
      <path
        d="M16.97 10.48 A 5.2 5.2 0 1 1 13.52 7.03"
        stroke="currentColor"
        stroke-width="1.85"
        stroke-linecap="round"
      />
      <path
        d="M19.89 9.59 A 8.25 8.25 0 1 1 14.41 4.11"
        stroke="currentColor"
        stroke-width="1.85"
        stroke-linecap="round"
      />
    </svg>

    <!-- 内网：三台设备互联，表示局域网 -->
    <svg
      v-else-if="meta.kind === 'internal'"
      class="net-glyph"
      viewBox="1 1 22 22"
      fill="none"
      aria-hidden="true"
    >
      <rect x="9" y="2.4" width="6" height="6" rx="1.15" stroke="currentColor" stroke-width="1.85" />
      <rect x="2.2" y="15.6" width="6" height="6" rx="1.15" stroke="currentColor" stroke-width="1.85" />
      <rect x="15.8" y="15.6" width="6" height="6" rx="1.15" stroke="currentColor" stroke-width="1.85" />
      <path
        d="M12 8.4v2.05c0 .7-.55 1.25-1.25 1.25H6.7c-.6 0-1.1.5-1.1 1.1v2.8"
        stroke="currentColor"
        stroke-width="1.85"
        stroke-linecap="round"
        stroke-linejoin="round"
      />
      <path
        d="M12 8.4v2.05c0 .7.55 1.25 1.25 1.25h4.05c.6 0 1.1.5 1.1 1.1v2.8"
        stroke="currentColor"
        stroke-width="1.85"
        stroke-linecap="round"
        stroke-linejoin="round"
      />
    </svg>

    <!-- 外网：地球，表示公网 / WAN -->
    <svg
      v-else
      class="net-glyph"
      viewBox="1 1 22 22"
      fill="none"
      aria-hidden="true"
    >
      <circle cx="12" cy="12" r="8.15" stroke="currentColor" stroke-width="1.85" />
      <path
        d="M12 3.85c2.15 2.35 3.35 5.15 3.35 8.15s-1.2 5.8-3.35 8.15C9.85 17.8 8.65 15 8.65 12s1.2-5.8 3.35-8.15Z"
        stroke="currentColor"
        stroke-width="1.85"
        stroke-linejoin="round"
      />
      <path d="M4.2 12h15.6" stroke="currentColor" stroke-width="1.85" stroke-linecap="round" />
    </svg>
  </button>
</template>

<style scoped>
.net-glyph {
  width: 18px;
  height: 18px;
  display: block;
  overflow: visible;
}
</style>
