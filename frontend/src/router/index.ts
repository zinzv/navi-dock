import { createRouter, createWebHistory } from 'vue-router'
import DashboardLayout from '../layouts/DashboardLayout.vue'
import SettingsLayout from '../layouts/SettingsLayout.vue'
import Home from '../pages/Home.vue'
import Settings from '../pages/Settings.vue'

const router = createRouter({
  history: createWebHistory(),
  routes: [
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

export default router
