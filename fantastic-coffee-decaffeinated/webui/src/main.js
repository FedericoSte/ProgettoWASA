import { createApp } from 'vue'
import App from './App.vue'
import router from './router'

// Importa il file CSS globale
import './assets/main.css'

const app = createApp(App)

app.use(router)

app.mount('#app')