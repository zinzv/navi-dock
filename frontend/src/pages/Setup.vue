<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { useAuthStore } from '../stores/auth'
import { useSettingsStore } from '../stores/settings'
import PasswordInput from '../components/PasswordInput.vue'

const { t } = useI18n()
const router = useRouter()
const auth = useAuthStore()
const settings = useSettingsStore()

const username = ref('')
const password = ref('')
const confirm = ref('')
const busy = ref(false)
const error = ref('')

async function onSubmit() {
  if (busy.value) return
  const name = username.value.trim()
  if (!name || !password.value) return
  if (password.value !== confirm.value) {
    error.value = t('setup.passwordMismatch')
    return
  }
  busy.value = true
  error.value = ''
  try {
    await auth.completeSetup(name, password.value)
    await settings.load()
    await router.replace({ name: 'home' })
  } catch (e) {
    error.value = e instanceof Error ? e.message : t('setup.failed')
  } finally {
    busy.value = false
  }
}
</script>

<template>
  <div class="setup-page">
    <div class="setup-card">
      <div class="setup-brand">
        <img
          v-if="settings.siteIcon"
          class="setup-icon"
          :src="settings.siteIcon"
          alt=""
        />
        <span v-else class="setup-mark" aria-hidden="true">🏠</span>
        <h1>{{ settings.siteTitle || t('app.name') }}</h1>
      </div>
      <p class="setup-desc">{{ t('setup.desc') }}</p>

      <form class="setup-form" @submit.prevent="onSubmit">
        <label class="field">
          <span>{{ t('setup.username') }}</span>
          <input
            v-model="username"
            type="text"
            autocomplete="username"
            maxlength="32"
            :placeholder="t('setup.usernamePlaceholder')"
            :disabled="busy"
          />
        </label>
        <label class="field">
          <span>{{ t('setup.password') }}</span>
          <PasswordInput
            v-model="password"
            autocomplete="new-password"
            :placeholder="t('setup.passwordPlaceholder')"
            :disabled="busy"
          />
        </label>
        <label class="field">
          <span>{{ t('setup.confirmPassword') }}</span>
          <PasswordInput
            v-model="confirm"
            autocomplete="new-password"
            :disabled="busy"
          />
        </label>

        <p v-if="error" class="setup-error">{{ error }}</p>

        <button
          type="submit"
          class="setup-btn"
          :disabled="busy || !username.trim() || !password"
        >
          {{ busy ? t('setup.submitting') : t('setup.submit') }}
        </button>
      </form>
    </div>
  </div>
</template>

<style scoped>
.setup-page {
  min-height: 100vh;
  display: grid;
  place-items: center;
  padding: 24px;
  background:
    radial-gradient(ellipse at 20% 10%, color-mix(in srgb, var(--accent) 18%, transparent), transparent 50%),
    radial-gradient(ellipse at 80% 90%, color-mix(in srgb, var(--bg3) 70%, transparent), transparent 45%),
    linear-gradient(160deg, var(--bg0), var(--bg1) 45%, var(--bg2));
  color: var(--text, #f3f4f6);
}

.setup-card {
  width: min(100%, 400px);
  padding: 28px 26px 24px;
  border-radius: 16px;
  border: 1px solid color-mix(in srgb, #fff 14%, transparent);
  background: color-mix(in srgb, #0f172a 55%, transparent);
  backdrop-filter: blur(16px);
  box-shadow: 0 18px 48px rgba(0, 0, 0, 0.28);
}

.setup-brand {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-bottom: 8px;
}

.setup-brand h1 {
  margin: 0;
  font-size: 22px;
  font-weight: 650;
  letter-spacing: 0.01em;
}

.setup-icon {
  width: 36px;
  height: 36px;
  border-radius: 10px;
  object-fit: cover;
}

.setup-mark {
  width: 36px;
  height: 36px;
  display: grid;
  place-items: center;
  font-size: 22px;
}

.setup-desc {
  margin: 0 0 20px;
  font-size: 13px;
  line-height: 1.55;
  color: color-mix(in srgb, #fff 68%, transparent);
}

.setup-form {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.field {
  display: flex;
  flex-direction: column;
  gap: 6px;
  font-size: 13px;
  color: color-mix(in srgb, #fff 72%, transparent);
}

.field input,
.field :deep(input) {
  height: 40px;
  padding: 0 12px;
  border-radius: 10px;
  border: 1px solid color-mix(in srgb, #fff 16%, transparent);
  background: color-mix(in srgb, #020617 45%, transparent);
  color: #f8fafc;
  font-size: 14px;
  outline: none;
}

.field :deep(input) {
  padding-right: 40px;
}

.field input:focus,
.field :deep(input:focus) {
  border-color: color-mix(in srgb, var(--accent) 70%, #fff);
}

.setup-error {
  margin: 0;
  font-size: 13px;
  color: #f87171;
}

.setup-btn {
  margin-top: 6px;
  height: 40px;
  border: 0;
  border-radius: 10px;
  background: var(--accent);
  color: #0b1220;
  font-size: 14px;
  font-weight: 600;
  cursor: pointer;
}

.setup-btn:disabled {
  opacity: 0.55;
  cursor: not-allowed;
}

:global(html[data-theme='light']) .setup-card {
  background: color-mix(in srgb, #fff 82%, transparent);
  border-color: color-mix(in srgb, #0f172a 10%, transparent);
  box-shadow: 0 16px 40px rgba(15, 23, 42, 0.12);
  color: #0f172a;
}

:global(html[data-theme='light']) .setup-desc,
:global(html[data-theme='light']) .field {
  color: #64748b;
}

:global(html[data-theme='light']) .field input,
:global(html[data-theme='light']) .field :deep(input) {
  background: #fff;
  border-color: #e2e8f0;
  color: #0f172a;
}

:global(html[data-theme='light']) .setup-btn {
  color: #fff;
}
</style>
