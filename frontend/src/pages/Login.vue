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
const busy = ref(false)
const error = ref('')

async function onSubmit() {
  if (busy.value) return
  const name = username.value.trim()
  if (!name || !password.value) return
  busy.value = true
  error.value = ''
  try {
    await auth.login(name, password.value)
    await settings.load()
    await router.replace({ name: 'home' })
  } catch (e) {
    error.value = e instanceof Error ? e.message : t('login.failed')
  } finally {
    busy.value = false
  }
}
</script>

<template>
  <div class="login-page">
    <div class="login-card">
      <div class="login-brand">
        <img
          v-if="settings.siteIcon"
          class="login-icon"
          :src="settings.siteIcon"
          alt=""
        />
        <span v-else class="login-mark" aria-hidden="true">🏠</span>
        <h1>{{ settings.siteTitle || t('app.name') }}</h1>
      </div>
      <p class="login-desc">{{ t('login.desc') }}</p>

      <form class="login-form" @submit.prevent="onSubmit">
        <label class="field">
          <span>{{ t('login.username') }}</span>
          <input
            v-model="username"
            type="text"
            autocomplete="username"
            maxlength="32"
            :disabled="busy"
          />
        </label>
        <label class="field">
          <span>{{ t('login.password') }}</span>
          <PasswordInput
            v-model="password"
            autocomplete="current-password"
            :disabled="busy"
          />
        </label>

        <p v-if="error" class="login-error">{{ error }}</p>

        <button
          type="submit"
          class="login-btn"
          :disabled="busy || !username.trim() || !password"
        >
          {{ busy ? t('login.submitting') : t('login.submit') }}
        </button>
      </form>
    </div>
  </div>
</template>

<style scoped>
.login-page {
  min-height: 100vh;
  display: grid;
  place-items: center;
  padding: 24px;
  background:
    radial-gradient(ellipse at 20% 10%, color-mix(in srgb, var(--accent) 18%, transparent), transparent 50%),
    radial-gradient(ellipse at 80% 90%, color-mix(in srgb, var(--bg3) 70%, transparent), transparent 45%),
    linear-gradient(160deg, var(--bg0), var(--bg1) 45%, var(--bg2));
}

.login-card {
  width: min(100%, 400px);
  padding: 28px 26px 24px;
  border-radius: 16px;
  border: 1px solid color-mix(in srgb, #fff 14%, transparent);
  background: color-mix(in srgb, #0f172a 55%, transparent);
  backdrop-filter: blur(16px);
  box-shadow: 0 18px 48px rgba(0, 0, 0, 0.28);
  color: #f8fafc;
}

.login-brand {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-bottom: 8px;
}

.login-brand h1 {
  margin: 0;
  font-size: 22px;
  font-weight: 650;
}

.login-icon {
  width: 36px;
  height: 36px;
  border-radius: 10px;
  object-fit: cover;
}

.login-mark {
  width: 36px;
  height: 36px;
  display: grid;
  place-items: center;
  font-size: 22px;
}

.login-desc {
  margin: 0 0 20px;
  font-size: 13px;
  line-height: 1.55;
  color: color-mix(in srgb, #fff 68%, transparent);
}

.login-form {
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
  border: 1.5px solid color-mix(in srgb, #fff 20%, transparent);
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

.login-error {
  margin: 0;
  font-size: 13px;
  color: #f87171;
}

.login-btn {
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

.login-btn:disabled {
  opacity: 0.55;
  cursor: not-allowed;
}

:global(html[data-theme='light']) .login-card {
  background: color-mix(in srgb, #fff 82%, transparent);
  border-color: color-mix(in srgb, #0f172a 10%, transparent);
  box-shadow: 0 16px 40px rgba(15, 23, 42, 0.12);
  color: #0f172a;
}

:global(html[data-theme='light']) .login-desc,
:global(html[data-theme='light']) .field {
  color: #64748b;
}

:global(html[data-theme='light']) .field input,
:global(html[data-theme='light']) .field :deep(input) {
  background: #fff;
  border-color: #cbd5e1;
  color: #0f172a;
}

:global(html[data-theme='light']) .login-btn {
  color: #fff;
}
</style>
