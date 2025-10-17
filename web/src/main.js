import { createApp } from 'vue'
import { createPinia } from 'pinia'
import App from './App.vue'

// Добавим отладочную информацию
console.log('🚀 Starting Vue application...')

const app = createApp(App)
const pinia = createPinia()

app.use(pinia)

// Попробуем отловить ошибки
app.config.errorHandler = (err, instance, info) => {
  console.error('Vue Error:', err)
  console.error('Instance:', instance)
  console.error('Info:', info)
}

app.mount('#app')

console.log('✅ Vue app mounted successfully!')