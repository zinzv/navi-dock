import { createRouter, createWebHistory } from 'vue-router'
import DashboardLayout from '../layouts/DashboardLayout.vue'
import SettingsLayout from '../layouts/SettingsLayout.vue'
import Home from '../pages/Home.vue'
import Settings from '../pages/Settings.vue'
import Setup from '../pages/Setup.vue'
import Login from '../pages/Login.vue'
import { useAuthStore } from '../stores/auth'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    {
      path: '/setup',
      name: 'setup',
      component: Setup,
    },
    {
      path: '/login',
      name: 'login',
      component: Login,
    },
    {
      path: '/',
      component: DashboardLayout,
      children: [{ path: '', name: 'home', component: Home }],
    },
    {
      path: '/settings',
      component: SettingsLayout,
      children: [{ path: '', name: 'settings', component: Settings }],
    },
  ],
})

router.beforeEach(async (to) => {
  const auth = useAuthStore()
  if (!auth.ready) {
    await auth.load()
  }

  if (auth.setupRequired) {
    if (to.name !== 'setup') return { name: 'setup' }
    return true
  }

  if (to.name === 'setup') {
    return { name: auth.isLoggedIn ? 'home' : 'login' }
  }

  if (!auth.isLoggedIn) {
    if (to.name !== 'login') return { name: 'login' }
    return true
  }

  if (to.name === 'login') {
    return { name: 'home' }
  }

  return true
})

export default router
