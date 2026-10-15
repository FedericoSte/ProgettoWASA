import { createRouter, createWebHistory } from 'vue-router'
// Importeremo le viste che creeremo nel prossimo step
import LoginView from '../views/LoginView.vue'
import ChatView from '../views/ChatView.vue'

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    { path: '/', redirect: '/login' },
    { path: '/login', name: 'login', component: LoginView },
    { path: '/app', name: 'app', component: ChatView }
  ]
})

// Navigation Guard: Protegge la rotta '/app'
router.beforeEach((to, from, next) => {
  const isAuthenticated = !!localStorage.getItem('wasaToken');
  
  if (to.name !== 'login' && !isAuthenticated) {
    // Se cerca di andare nella chat ma non ha il token, lo rimanda al login
    next({ name: 'login' })
  } else if (to.name === 'login' && isAuthenticated) {
    // Se è già loggato e va sul login, lo manda diretto alla chat
    next({ name: 'app' })
  } else {
    next()
  }
})

export default router