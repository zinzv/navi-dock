import { computed, ref } from 'vue'
import { defineStore } from 'pinia'
import {
  clearAuthToken,
  createUser,
  fetchAuthStatus,
  getAuthToken,
  login as apiLogin,
  setAuthToken,
  type UserPublic,
} from '../api/auth'

export const useAuthStore = defineStore('auth', () => {
  const ready = ref(false)
  const setupRequired = ref(false)
  const user = ref<UserPublic | null>(null)

  const isLoggedIn = computed(() => !!user.value)
  const isAdmin = computed(() => user.value?.role === 'admin')

  async function load() {
    try {
      const status = await fetchAuthStatus()
      setupRequired.value = status.setup_required
      user.value = status.user || null
      if (!status.user && getAuthToken()) {
        clearAuthToken()
      }
    } catch {
      setupRequired.value = false
      user.value = null
    } finally {
      ready.value = true
    }
  }

  async function login(username: string, password: string) {
    const result = await apiLogin(username, password)
    setAuthToken(result.token)
    user.value = result.user
    setupRequired.value = false
    return result.user
  }

  async function completeSetup(username: string, password: string) {
    await createUser(username, password, 'admin')
    return login(username, password)
  }

  function logout() {
    clearAuthToken()
    user.value = null
  }

  return {
    ready,
    setupRequired,
    user,
    isLoggedIn,
    isAdmin,
    load,
    login,
    completeSetup,
    logout,
  }
})
