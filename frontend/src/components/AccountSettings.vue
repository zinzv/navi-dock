<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { Icon } from '@iconify/vue'
import {
  changePassword,
  createUser,
  deleteUser,
  fetchUsers,
  type UserPublic,
} from '../api/auth'
import { useAuthStore } from '../stores/auth'
import { useSettingsStore } from '../stores/settings'
import PasswordInput from './PasswordInput.vue'

const emit = defineEmits<{
  saved: []
}>()

const { t } = useI18n()
const router = useRouter()
const auth = useAuthStore()
const settings = useSettingsStore()

const loading = ref(true)
const busy = ref(false)
const error = ref('')
const statusMsg = ref('')
const users = ref<UserPublic[]>([])

const resetOpen = ref(false)
const resetError = ref('')
const oldPassword = ref('')
const newPassword = ref('')
const newConfirm = ref('')

const createOpen = ref(false)
const createError = ref('')
const createUsername = ref('')
const createPassword = ref('')
const createConfirm = ref('')
const createRole = ref<'admin' | 'user'>('user')
const menuOpenId = ref<string | null>(null)

const LIST_PAGE_SIZE = 5
const listPage = ref(1)
const listPageCount = computed(() => Math.max(1, Math.ceil(users.value.length / LIST_PAGE_SIZE)))
const pagedUsers = computed(() => {
  const start = (listPage.value - 1) * LIST_PAGE_SIZE
  return users.value.slice(start, start + LIST_PAGE_SIZE)
})

const createPasswordHint = computed(() => {
  if (!createPassword.value) return ''
  if (createPassword.value.length < 6) return t('settings.accountPasswordHint')
  return ''
})

const createConfirmHint = computed(() => {
  if (!createConfirm.value) return ''
  if (createPassword.value !== createConfirm.value) return t('settings.accountPasswordMismatch')
  return ''
})

const canCreate = computed(
  () =>
    createUsername.value.trim().length >= 2 &&
    createPassword.value.length >= 6 &&
    createPassword.value === createConfirm.value &&
    !busy.value,
)

function clampListPage() {
  if (listPage.value > listPageCount.value) {
    listPage.value = listPageCount.value
  }
  if (listPage.value < 1) listPage.value = 1
}

const isLoggedIn = computed(() => auth.isLoggedIn)
const isAdmin = computed(() => auth.isAdmin)
const anyModalOpen = computed(() => resetOpen.value || createOpen.value)

async function refresh() {
  loading.value = true
  error.value = ''
  try {
    await auth.load()
    if (auth.user && auth.isAdmin) {
      const data = await fetchUsers()
      users.value = data.users
      clampListPage()
    } else {
      users.value = []
      listPage.value = 1
    }
  } catch (e) {
    error.value = e instanceof Error ? e.message : t('settings.accountLoadFailed')
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  void refresh()
})

function flash(msg: string) {
  statusMsg.value = msg
  emit('saved')
  window.setTimeout(() => {
    if (statusMsg.value === msg) statusMsg.value = ''
  }, 1600)
}

async function onLogout() {
  auth.logout()
  users.value = []
  statusMsg.value = ''
  error.value = ''
  await router.replace({ name: 'login' })
}

function openReset() {
  oldPassword.value = ''
  newPassword.value = ''
  newConfirm.value = ''
  resetError.value = ''
  createOpen.value = false
  resetOpen.value = true
}

function closeReset() {
  if (busy.value) return
  resetOpen.value = false
  resetError.value = ''
}

async function onResetPassword() {
  if (busy.value) return
  if (!oldPassword.value || !newPassword.value) return
  if (newPassword.value !== newConfirm.value) {
    resetError.value = t('settings.accountPasswordMismatch')
    return
  }
  busy.value = true
  resetError.value = ''
  error.value = ''
  try {
    await changePassword(oldPassword.value, newPassword.value)
    resetOpen.value = false
    oldPassword.value = ''
    newPassword.value = ''
    newConfirm.value = ''
    flash(t('settings.accountPasswordChanged'))
  } catch (e) {
    resetError.value = e instanceof Error ? e.message : t('settings.accountSaveFailed')
  } finally {
    busy.value = false
  }
}

async function openCreate() {
  createUsername.value = ''
  createPassword.value = ''
  createConfirm.value = ''
  createRole.value = 'user'
  createError.value = ''
  listPage.value = 1
  resetOpen.value = false
  createOpen.value = true
  if (auth.isAdmin) {
    try {
      const data = await fetchUsers()
      users.value = data.users
      clampListPage()
    } catch {
      /* keep existing list */
    }
  }
}

function closeCreate() {
  if (busy.value) return
  createOpen.value = false
  createError.value = ''
  menuOpenId.value = null
}

function toggleMenu(id: string) {
  menuOpenId.value = menuOpenId.value === id ? null : id
}

async function onCreate() {
  if (busy.value || !isAdmin.value || !canCreate.value) return
  const username = createUsername.value.trim()
  const password = createPassword.value
  if (!username || !password) return
  if (password !== createConfirm.value) {
    createError.value = t('settings.accountPasswordMismatch')
    return
  }
  busy.value = true
  createError.value = ''
  error.value = ''
  try {
    await createUser(username, password, createRole.value)
    createOpen.value = false
    createUsername.value = ''
    createPassword.value = ''
    createConfirm.value = ''
    createRole.value = 'user'
    flash(t('settings.accountCreated'))
    await refresh()
  } catch (e) {
    createError.value = e instanceof Error ? e.message : t('settings.accountSaveFailed')
  } finally {
    busy.value = false
  }
}

async function onDelete(user: UserPublic) {
  menuOpenId.value = null
  if (busy.value || !isAdmin.value) return
  if (auth.user?.id === user.id) {
    error.value = t('settings.accountCannotDeleteSelf')
    return
  }
  if (!window.confirm(t('settings.accountDeleteConfirm', { name: user.username }))) return
  busy.value = true
  error.value = ''
  try {
    await deleteUser(user.id)
    if (auth.user?.id === user.id) {
      onLogout()
    }
    flash(t('settings.accountDeleted'))
    await refresh()
  } catch (e) {
    error.value = e instanceof Error ? e.message : t('settings.accountSaveFailed')
  } finally {
    busy.value = false
  }
}

function roleLabel(role: string) {
  return role === 'admin' ? t('settings.accountRoleAdmin') : t('settings.accountRoleUser')
}
</script>

<template>
  <section class="card">
    <div class="card-header">
      <h2 class="card-title">{{ t('settings.account') }}</h2>
      <button
        v-if="isAdmin && isLoggedIn && !loading"
        type="button"
        class="primary-btn"
        :disabled="busy || anyModalOpen"
        @click="openCreate"
      >
        {{ t('settings.accountCreate') }}
      </button>
    </div>

    <p v-if="loading" class="status">{{ t('settings.accountLoading') }}</p>
    <p v-else-if="error" class="status error">{{ error }}</p>
    <p v-else-if="statusMsg" class="status ok">{{ statusMsg }}</p>

    <template v-if="!loading && isLoggedIn">
      <div class="rows">
        <div class="row current-row">
          <div class="current-line">
            <span class="value">{{ auth.user?.username }}</span>
            <span class="role-tag">{{ roleLabel(auth.user?.role || '') }}</span>
          </div>
          <div class="toolbar-actions">
            <button type="button" class="secondary-btn" :disabled="busy || anyModalOpen" @click="openReset">
              {{ t('settings.accountResetPassword') }}
            </button>
            <button type="button" class="logout-btn" :disabled="busy" @click="onLogout">
              {{ t('settings.accountLogout') }}
            </button>
          </div>
        </div>
      </div>
    </template>

    <Teleport to="body">
      <div
        v-if="resetOpen"
        class="account-modal-root"
        role="dialog"
        aria-modal="true"
        :data-theme="settings.resolvedTheme"
        :aria-label="t('settings.accountResetPassword')"
      >
        <div class="account-backdrop" @click="closeReset" />
        <div class="account-modal">
          <div class="account-head">
            <h2>{{ t('settings.accountResetPassword') }}</h2>
            <button type="button" class="account-close" :disabled="busy" @click="closeReset">
              <Icon icon="mdi:close" width="18" />
            </button>
          </div>
          <div class="account-body">
            <label class="field">
              <span class="field-label">{{ t('settings.accountOldPassword') }}</span>
              <PasswordInput
                v-model="oldPassword"
                input-class="field-input"
                autocomplete="current-password"
                :disabled="busy"
              />
            </label>
            <label class="field">
              <span class="field-label">{{ t('settings.accountNewPassword') }}</span>
              <PasswordInput
                v-model="newPassword"
                input-class="field-input"
                autocomplete="new-password"
                :placeholder="t('settings.accountPasswordPlaceholder')"
                :disabled="busy"
              />
            </label>
            <label class="field">
              <span class="field-label">{{ t('settings.accountConfirmPassword') }}</span>
              <PasswordInput
                v-model="newConfirm"
                input-class="field-input"
                autocomplete="new-password"
                :disabled="busy"
                @keyup.enter="onResetPassword"
              />
            </label>
            <p v-if="resetError" class="modal-error">{{ resetError }}</p>
          </div>
          <div class="account-foot">
            <button type="button" class="ghost-btn" :disabled="busy" @click="closeReset">
              {{ t('settings.groupsCancel') }}
            </button>
            <button
              type="button"
              class="primary-btn"
              :disabled="busy || !oldPassword || !newPassword"
              @click="onResetPassword"
            >
              {{ t('settings.accountResetPassword') }}
            </button>
          </div>
        </div>
      </div>
    </Teleport>

    <Teleport to="body">
      <div
        v-if="createOpen"
        class="account-modal-root"
        role="dialog"
        aria-modal="true"
        :data-theme="settings.resolvedTheme"
        :aria-label="t('settings.accountCreate')"
        @click="menuOpenId = null"
      >
        <div class="account-backdrop" @click="closeCreate" />
        <div class="account-modal account-modal-wide" @click.stop>
          <div class="account-head">
            <div class="account-head-copy">
              <h2>{{ t('settings.accountCreate') }}</h2>
              <p>{{ t('settings.accountCreateHint') }}</p>
            </div>
            <button type="button" class="account-close" :disabled="busy" @click="closeCreate">
              <Icon icon="mdi:close" width="18" />
            </button>
          </div>
          <div class="account-body create-body">
            <div class="list-block">
              <p class="list-title">{{ t('settings.accountList') }}</p>
              <div class="user-list">
                <div
                  v-for="user in pagedUsers"
                  :key="user.id"
                  class="user-item"
                  :class="{ current: auth.user?.id === user.id }"
                >
                  <div class="user-meta">
                    <span class="value">{{ user.username }}</span>
                    <span class="role-tag">{{ roleLabel(user.role) }}</span>
                  </div>
                  <div class="user-menu" @click.stop>
                    <button
                      type="button"
                      class="icon-action"
                      :title="t('settings.accountMore')"
                      :disabled="busy"
                      @click="toggleMenu(user.id)"
                    >
                      <Icon icon="mdi:dots-horizontal" width="18" />
                    </button>
                    <div v-if="menuOpenId === user.id" class="user-menu-popover">
                      <button
                        type="button"
                        class="danger"
                        :disabled="auth.user?.id === user.id"
                        :title="
                          auth.user?.id === user.id
                            ? t('settings.accountCannotDeleteSelf')
                            : t('home.delete')
                        "
                        @click="onDelete(user)"
                      >
                        <Icon icon="mdi:trash-can-outline" width="15" />
                        {{ t('home.delete') }}
                      </button>
                    </div>
                  </div>
                </div>
                <div v-if="!users.length" class="user-item empty">
                  <span class="value mute">{{ t('settings.accountEmpty') }}</span>
                </div>
              </div>
              <div v-if="users.length > LIST_PAGE_SIZE" class="list-pager">
                <button
                  type="button"
                  class="pager-btn"
                  :disabled="listPage <= 1"
                  :title="t('settings.accountPagePrev')"
                  @click="listPage -= 1"
                >
                  <Icon icon="mdi:chevron-left" width="18" />
                </button>
                <span class="pager-info">{{ listPage }} / {{ listPageCount }}</span>
                <button
                  type="button"
                  class="pager-btn"
                  :disabled="listPage >= listPageCount"
                  :title="t('settings.accountPageNext')"
                  @click="listPage += 1"
                >
                  <Icon icon="mdi:chevron-right" width="18" />
                </button>
              </div>
            </div>

            <div class="form-block">
              <div class="form-head">
                <p class="form-title">{{ t('settings.accountInfo') }}</p>
                <p class="form-hint">{{ t('settings.accountInfoHint') }}</p>
              </div>
              <label class="field">
                <span class="field-label">{{ t('settings.accountUsername') }}</span>
                <input
                  v-model="createUsername"
                  class="field-input"
                  type="text"
                  autocomplete="off"
                  maxlength="32"
                  :placeholder="t('settings.accountUsernamePlaceholder')"
                  :disabled="busy"
                />
              </label>
              <label class="field">
                <span class="field-label">{{ t('settings.accountRole') }}</span>
                <select
                  v-model="createRole"
                  class="role-select"
                  :disabled="busy"
                  :aria-label="t('settings.accountRole')"
                >
                  <option value="user">{{ t('settings.accountRoleUser') }}</option>
                  <option value="admin">{{ t('settings.accountRoleAdmin') }}</option>
                </select>
              </label>
              <label class="field">
                <span class="field-label">{{ t('settings.accountPassword') }}</span>
                <PasswordInput
                  v-model="createPassword"
                  input-class="field-input"
                  autocomplete="new-password"
                  :placeholder="t('settings.accountPasswordInputPlaceholder')"
                  :disabled="busy"
                />
                <span class="field-hint" :class="{ error: createPasswordHint }">
                  {{ createPasswordHint || t('settings.accountPasswordHint') }}
                </span>
              </label>
              <label class="field">
                <span class="field-label">{{ t('settings.accountConfirmPassword') }}</span>
                <PasswordInput
                  v-model="createConfirm"
                  input-class="field-input"
                  autocomplete="new-password"
                  :placeholder="t('settings.accountConfirmPlaceholder')"
                  :disabled="busy"
                  @keyup.enter="onCreate"
                />
                <span v-if="createConfirmHint" class="field-hint error">{{ createConfirmHint }}</span>
                <span
                  v-else-if="createConfirm && createPassword === createConfirm"
                  class="field-hint ok"
                >
                  {{ t('settings.accountPasswordMatch') }}
                </span>
              </label>
              <p v-if="createError" class="modal-error">{{ createError }}</p>
            </div>
          </div>
          <div class="account-foot">
            <button type="button" class="ghost-btn" :disabled="busy" @click="closeCreate">
              {{ t('settings.groupsCancel') }}
            </button>
            <button
              type="button"
              class="primary-btn"
              :disabled="!canCreate"
              @click="onCreate"
            >
              {{ t('settings.accountCreate') }}
            </button>
          </div>
        </div>
      </div>
    </Teleport>
  </section>
</template>

<style scoped>
.card {
  background: var(--sv-surface);
  border: 1px solid var(--sv-border);
  border-radius: 12px;
  padding: 20px 24px 16px;
  margin-bottom: 20px;
}

.card-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  margin-bottom: 8px;
}

.card-header .primary-btn {
  flex-shrink: 0;
}

.card-title {
  margin: 0;
  font-size: 18px;
  font-weight: 600;
  color: var(--sv-text);
}

.section-label {
  margin: 14px 0 0;
  padding-top: 10px;
  font-size: 13px;
  font-weight: 600;
  color: var(--sv-text);
  border-top: 1px solid var(--sv-border);
}

.status {
  margin: 8px 0 12px;
  font-size: 13px;
  color: var(--sv-mute);
}

.status.error {
  color: #ef4444;
}

.status.ok {
  color: var(--sv-accent);
}

.rows {
  display: flex;
  flex-direction: column;
}

.row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  min-height: 58px;
  padding: 8px 0;
  font-size: 14px;
  border-top: 1px solid var(--sv-border);
}

.rows > .row:first-child {
  border-top: 0;
}

.row.actions {
  justify-content: flex-end;
  gap: 8px;
}

.row.current-row {
  gap: 18px;
  min-height: 52px;
  padding: 8px 0 0;
  max-width: 820px;
}

.current-line {
  display: flex;
  align-items: center;
  gap: 8px;
  min-width: 0;
  flex: 1;
  flex-wrap: nowrap;
}

.toolbar-actions {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-shrink: 0;
}

.current-line .label {
  min-width: auto;
}

.current-line .value {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.label {
  color: var(--sv-mute);
  flex-shrink: 0;
  min-width: 88px;
}

.value {
  color: var(--sv-text);
}

.account-modal-root .value {
  color: var(--modal-text);
}

.value.mute {
  color: var(--sv-mute);
}

.account-modal-root .value.mute {
  color: var(--modal-mute);
}

.row-input {
  flex: 1;
  max-width: 280px;
  height: 34px;
  padding: 0 10px;
  border: 1px solid var(--sv-border);
  border-radius: 8px;
  background: color-mix(in srgb, var(--sv-surface) 80%, #000 4%);
  color: var(--sv-text);
  font-size: 13px;
  outline: none;
}

.row-input:focus {
  border-color: color-mix(in srgb, var(--sv-accent) 55%, var(--sv-border));
}

.primary-btn {
  border: 0;
  background: var(--sv-accent);
  color: #fff;
  height: 36px;
  padding: 0 14px;
  border-radius: 8px;
  font-size: 13px;
  cursor: pointer;
}

.primary-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.primary-btn:hover:not(:disabled) {
  background: var(--sv-accent-hover);
}

.secondary-btn,
.logout-btn {
  border: 1px solid var(--sv-border);
  background: transparent;
  color: var(--sv-text);
  height: 36px;
  padding: 0 14px;
  border-radius: 8px;
  font-size: 13px;
  cursor: pointer;
}

.secondary-btn:hover:not(:disabled) {
  background: var(--sv-surface-hover);
}

.logout-btn {
  border-color: transparent;
  color: var(--sv-mute);
}

.logout-btn:hover:not(:disabled) {
  color: var(--sv-danger);
  background: var(--sv-danger-soft);
}

.secondary-btn:disabled,
.logout-btn:disabled {
  opacity: 0.4;
  cursor: not-allowed;
}

.ghost-btn {
  border: 1px solid var(--sv-border);
  background: transparent;
  color: var(--sv-text);
  height: 32px;
  padding: 0 14px;
  border-radius: 8px;
  font-size: 13px;
  cursor: pointer;
}

.ghost-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.user-meta {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-wrap: wrap;
  justify-content: flex-start;
}

.role-tag {
  font-size: 11px;
  padding: 2px 7px;
  border-radius: 999px;
  color: color-mix(in srgb, var(--sv-accent) 78%, var(--sv-text));
  background: color-mix(in srgb, var(--sv-accent) 10%, transparent);
}

.account-modal-root .role-tag {
  color: var(--modal-mute);
  background: color-mix(in srgb, var(--modal-accent) 14%, transparent);
}

.icon-action {
  border: 0;
  background: transparent;
  color: var(--sv-mute);
  width: 30px;
  height: 30px;
  border-radius: 8px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
}

.account-modal-root .icon-action {
  color: var(--modal-mute);
}

.icon-action:hover:not(:disabled) {
  color: var(--sv-text);
  background: color-mix(in srgb, var(--sv-accent) 12%, transparent);
}

.account-modal-root .icon-action:hover:not(:disabled) {
  color: var(--modal-text);
  background: color-mix(in srgb, var(--modal-accent) 12%, transparent);
}

.icon-action.danger:hover:not(:disabled) {
  color: #ef4444;
  background: color-mix(in srgb, #ef4444 12%, transparent);
}

.icon-action:disabled {
  opacity: 0.4;
  cursor: not-allowed;
}

.account-modal-root {
  --modal-accent: #5f66e8;
  --modal-bg: #ffffff;
  --modal-text: #182033;
  --modal-mute: #737b8c;
  --modal-border: rgba(15, 23, 42, 0.08);
  --modal-input-bg: #f8fafc;
  --modal-list-bg: transparent;
  --modal-hover: rgba(15, 23, 42, 0.04);
  --modal-backdrop: rgba(15, 23, 42, 0.48);
  --modal-shadow: 0 24px 64px rgba(15, 23, 42, 0.22);
  --modal-disabled-bg: #e7e8ec;
  --modal-disabled-text: #a2a7b2;

  position: fixed;
  inset: 0;
  z-index: 1200;
  display: grid;
  place-items: center;
  padding: 20px;
  color: var(--modal-text);
}

.account-modal-root[data-theme='dark'] {
  --modal-bg: #202124;
  --modal-text: rgba(255, 255, 255, 0.92);
  --modal-mute: rgba(255, 255, 255, 0.55);
  --modal-border: rgba(255, 255, 255, 0.08);
  --modal-input-bg: rgba(0, 0, 0, 0.22);
  --modal-list-bg: transparent;
  --modal-hover: rgba(255, 255, 255, 0.05);
  --modal-backdrop: rgba(2, 6, 23, 0.62);
  --modal-shadow: 0 24px 64px rgba(0, 0, 0, 0.55);
  --modal-disabled-bg: rgba(255, 255, 255, 0.08);
  --modal-disabled-text: rgba(255, 255, 255, 0.35);
}

.account-backdrop {
  position: absolute;
  inset: 0;
  background: var(--modal-backdrop);
}

.account-modal {
  position: relative;
  width: min(420px, 100%);
  max-height: min(90vh, 720px);
  overflow: auto;
  border-radius: 14px;
  border: 1px solid var(--modal-border);
  background: var(--modal-bg);
  color: var(--modal-text);
  box-shadow: var(--modal-shadow);
}

.account-modal-wide {
  width: min(780px, 100%);
  min-height: 500px;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.account-head {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 12px;
  min-height: 76px;
  padding: 18px 20px 16px;
  border-bottom: 1px solid var(--modal-border);
  flex-shrink: 0;
}

.account-head-copy h2,
.account-head h2 {
  margin: 0;
  font-size: 18px;
  font-weight: 600;
  line-height: 24px;
  color: var(--modal-text);
}

.account-head-copy p {
  margin: 4px 0 0;
  font-size: 13px;
  color: var(--modal-mute);
}

.account-close {
  border: 0;
  background: transparent;
  color: var(--modal-mute);
  width: 32px;
  height: 32px;
  border-radius: 8px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  flex-shrink: 0;
}

.account-close:hover:not(:disabled) {
  color: var(--modal-text);
  background: var(--modal-hover);
}

.account-body {
  display: flex;
  flex-direction: column;
  gap: 12px;
  padding: 16px;
}

.create-body {
  display: grid;
  grid-template-columns: 280px minmax(0, 1fr);
  gap: 0;
  align-items: stretch;
  padding: 0;
  flex: 1;
  min-height: 0;
}

.list-block,
.form-block {
  display: flex;
  flex-direction: column;
  gap: 12px;
  min-width: 0;
  min-height: 0;
}

.list-block {
  padding: 20px;
  border-right: 1px solid var(--modal-border);
}

.form-block {
  padding: 20px 24px;
  border-left: 0;
  gap: 16px;
}

.list-title,
.form-title {
  margin: 0;
  font-size: 13px;
  font-weight: 600;
  color: var(--modal-text);
  line-height: 20px;
}

.form-head {
  display: flex;
  flex-direction: column;
  gap: 4px;
  margin-bottom: 2px;
}

.form-hint {
  margin: 0;
  font-size: 12px;
  color: var(--modal-mute);
}

.user-list {
  width: 100%;
  flex: 1;
  max-height: none;
  overflow: auto;
  border: 0;
  border-radius: 0;
  background: transparent;
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.user-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
  min-height: 50px;
  padding: 0 12px;
  border: 1px solid transparent;
  border-radius: 10px;
  background: var(--modal-hover);
}

.user-item:first-child {
  border-top: 0;
}

.user-item.current {
  background: color-mix(in srgb, var(--modal-accent) 6%, transparent);
  border-color: color-mix(in srgb, var(--modal-accent) 18%, transparent);
}

.user-item.empty {
  justify-content: flex-start;
  background: transparent;
}

.user-menu {
  position: relative;
  flex-shrink: 0;
}

.user-menu-popover {
  position: absolute;
  top: calc(100% + 4px);
  right: 0;
  z-index: 20;
  min-width: 120px;
  padding: 4px;
  border: 1px solid var(--modal-border);
  border-radius: 8px;
  background: var(--modal-bg);
  box-shadow: 0 10px 24px rgba(15, 23, 42, 0.14);
}

.user-menu-popover button {
  width: 100%;
  min-height: 32px;
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 0 10px;
  border: 0;
  border-radius: 6px;
  background: transparent;
  color: var(--modal-text);
  font-size: 12px;
  cursor: pointer;
}

.user-menu-popover button:hover:not(:disabled) {
  background: var(--modal-hover);
}

.user-menu-popover button.danger {
  color: var(--modal-mute);
}

.user-menu-popover button.danger:hover:not(:disabled) {
  color: #ef6a6a;
  background: rgba(239, 106, 106, 0.08);
}

.user-menu-popover button:disabled {
  opacity: 0.4;
  cursor: not-allowed;
}

.list-pager {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  padding-top: 2px;
}

.pager-btn {
  width: 28px;
  height: 28px;
  border: 1px solid var(--modal-border);
  border-radius: 8px;
  background: var(--modal-input-bg);
  color: var(--modal-text);
  display: inline-flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
}

.pager-btn:disabled {
  opacity: 0.4;
  cursor: not-allowed;
}

.pager-info {
  min-width: 48px;
  text-align: center;
  font-size: 12px;
  color: var(--modal-mute);
}

.user-meta {
  display: flex;
  align-items: center;
  gap: 8px;
  justify-content: flex-start;
  min-width: 0;
  flex-wrap: wrap;
}

.user-meta .value {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.field {
  display: flex;
  flex-direction: column;
  align-items: stretch;
  gap: 7px;
  width: 100%;
  max-width: 320px;
}

.field-label {
  font-size: 14px;
  font-weight: 500;
  color: var(--modal-text);
}

.field-hint {
  font-size: 12px;
  color: var(--modal-mute);
  line-height: 1.4;
}

.field-hint.error {
  color: #ef6a6a;
}

.field-hint.ok {
  color: #22a06b;
}

.role-select,
.field-input,
:deep(.field-input) {
  width: 100%;
  max-width: 100%;
  height: 42px;
  padding: 0 14px;
  border: 1px solid var(--modal-border);
  border-radius: 8px;
  background: var(--modal-input-bg);
  color: var(--modal-text);
  font-size: 14px;
  outline: none;
  box-sizing: border-box;
}

.role-select:focus,
.field-input:focus,
:deep(.field-input:focus) {
  border-color: color-mix(in srgb, var(--modal-accent) 50%, transparent);
  box-shadow: 0 0 0 3px color-mix(in srgb, var(--modal-accent) 10%, transparent);
}

.role-select:disabled {
  opacity: 0.55;
  cursor: not-allowed;
}

.field :deep(.password-wrap) {
  width: 100%;
  max-width: 100%;
}

@media (max-width: 720px) {
  .create-body {
    grid-template-columns: 1fr;
  }

  .list-block {
    border-right: 0;
    border-bottom: 1px solid var(--modal-border);
  }

  .form-block {
    padding-top: 16px;
  }
}

:deep(.field-input) {
  padding-right: 40px;
}

.field-input::placeholder,
:deep(.field-input::placeholder) {
  color: var(--modal-mute);
}

:deep(.toggle) {
  color: var(--modal-mute);
}

.modal-error {
  margin: 0;
  font-size: 13px;
  color: #ef6a6a;
}

.account-foot {
  display: flex;
  justify-content: flex-end;
  align-items: center;
  gap: 10px;
  height: 64px;
  padding: 0 20px;
  border-top: 1px solid var(--modal-border);
  flex-shrink: 0;
}

.account-modal-root .primary-btn {
  background: var(--modal-accent);
  height: 38px;
  padding: 0 18px;
}

.account-modal-root .primary-btn:disabled {
  background: var(--modal-disabled-bg);
  color: var(--modal-disabled-text);
  opacity: 1;
}

.account-modal-root .ghost-btn {
  border-color: var(--modal-border);
  color: var(--modal-text);
  height: 38px;
  padding: 0 14px;
}

.account-modal-root .icon-action {
  width: 32px;
  height: 32px;
  color: var(--modal-mute);
}

.account-modal-root .icon-action:hover:not(:disabled) {
  color: var(--modal-text);
  background: var(--modal-hover);
}

.account-modal-root .role-tag {
  height: 22px;
  padding: 0 8px;
  display: inline-flex;
  align-items: center;
  font-size: 12px;
  color: color-mix(in srgb, var(--modal-accent) 80%, var(--modal-text));
  background: color-mix(in srgb, var(--modal-accent) 8%, transparent);
}
</style>
