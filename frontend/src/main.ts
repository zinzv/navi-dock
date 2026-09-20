import { createApp } from 'vue'
import { createPinia } from 'pinia'
import App from './App.vue'
import router from './router'
import i18n from './i18n'
import { useSettingsStore } from './stores/settings'
import { useAuthStore } from './stores/auth'
import { useNetworkStore } from './stores/network'
import './styles/main.css'

const app = createApp(App)
const pinia = createPinia()

app.use(pinia)
app.use(router)
app.use(i18n)

const settings = useSettingsStore(pinia)
const auth = useAuthStore(pinia)
const network = useNetworkStore(pinia)

auth
  .load()
  .then(async () => {
    if (auth.isLoggedIn) {
      await settings.load()
      await network.start()
    }
  })
  .finally(() => {
    app.mount('#app')
  })
