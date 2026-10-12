<script setup lang="ts">
import { RouterLink, RouterView } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { useSettingsStore } from '../stores/settings'
import ThemeSwitcher from '../components/ThemeSwitcher.vue'
const { t } = useI18n()
const settings = useSettingsStore()
</script>

<template>
  <div class="settings-shell" :data-theme="settings.resolvedTheme">
    <header class="settings-header">
      <nav class="settings-header-inner" :aria-label="t('settings.title')">
        <RouterLink to="/">{{ t('app.name') }}</RouterLink>
        <span class="breadcrumb-divider" aria-hidden="true">/</span>
        <span aria-current="page">{{ t('settings.title') }}</span>
        <ThemeSwitcher />
      </nav>
    </header>
    <RouterView />
  </div>
</template>

<style scoped>
.settings-shell {
  --sv-accent: #6366f1;
  --sv-accent-hover: #5558df;
  --sv-accent-text: #4f46e5;
  --sv-accent-weak: #eef2ff;
  --sv-bg: #f7f8fa;
  --sv-surface: #fff;
  --sv-border: #eaecf0;
  --sv-text: #1f2937;
  --sv-regular: #1f2937;
  --sv-mute: #667085;
  --sv-input-bg: #fff;
  background: var(--sv-bg);
}

.settings-shell[data-theme='dark'] {
  --sv-accent: #6366f1;
  --sv-accent-text: #a5a8ff;
  --sv-accent-weak: #2c2c49;
  --sv-bg: #15171c;
  --sv-surface: #1e2128;
  --sv-border: #363b46;
  --sv-text: #edf0f5;
  --sv-regular: #edf0f5;
  --sv-mute: #a4adbd;
  --sv-input-bg: #191c22;
  background: var(--sv-bg);
  color-scheme: dark;
}

.settings-header {
  position: sticky;
  top: 0;
  z-index: 40;
  height: 64px;
  border-bottom: 1px solid var(--sv-border);
  background: var(--sv-surface);
}

.settings-header-inner {
  width: min(1440px, calc(100% - 64px));
  height: 100%;
  margin: auto;
  display: flex;
  align-items: center;
  gap: 16px;
  font-size: 14px;
}

.settings-header a {
  color: var(--sv-text);
  text-decoration: none;
  font-weight: 600;
}

.settings-header a:hover {
  color: var(--sv-accent-text);
}

.breadcrumb-divider {
  color: var(--sv-mute);
}

.settings-header :deep(.icon-btn) {
  margin-left: auto;
  width: 36px;
  height: 36px;
  display: grid;
  place-items: center;
  border: 1px solid var(--sv-border);
  border-radius: 8px;
  color: var(--sv-text);
  background: var(--sv-surface);
  cursor: pointer;
}

.settings-header :deep(.icon-btn:hover) {
  background: var(--sv-accent-weak);
}

@media (max-width: 1000px) {
  .settings-header-inner {
    width: calc(100% - 40px);
  }

}

@media (max-width: 640px) {
  .settings-header-inner {
    width: calc(100% - 32px);
  }

}
</style>
