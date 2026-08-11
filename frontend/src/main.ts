import { createApp } from 'vue'
import { createPinia } from 'pinia'
import App from './App.vue'
import router from './router'
import i18n from './i18n'
import { useSettingsStore } from './stores/settings'
import { useAuthStore } from './stores/auth'
import './styles/main.css'

const app = createApp(App)
const pinia = createPinia()

app.use(pinia)
app.use(router)
app.use(i18n)

const settings = useSettingsStore(pinia)
const auth = useAuthStore(pinia)

auth
  .load()
  .then(async () => {
    if (auth.isLoggedIn) {
      await settings.load()
    }
  })
  .finally(() => {
    app.mount('#app')
  })
