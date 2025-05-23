import { createApp } from 'vue'
import ElementPlus from 'element-plus'
import 'element-plus/dist/index.css'
import * as ElementPlusIconsVue from '@element-plus/icons-vue'

import App from './App.vue'
import router from './router'
import pinia from './store'

import './assets/styles/main.scss'
import { autoRefreshToken } from '@/utils/token'

const app = createApp(App)

// Register all Element Plus icons
for (const [key, component] of Object.entries(ElementPlusIconsVue)) {
  app.component(key, component)
}

app.use(ElementPlus)
app.use(pinia)
app.use(router)

app.mount('#app')

// Start the periodic token refresh task
// Check every 5 minutes if the token needs to be refreshed
const REFRESH_TOKEN_INTERVAL = 5 * 60 * 1000 // 5 minutes
setInterval(async () => {
  try {
    await autoRefreshToken()
  } catch (error) {
    console.error('Background token refresh failed:', error)
  }
}, REFRESH_TOKEN_INTERVAL)