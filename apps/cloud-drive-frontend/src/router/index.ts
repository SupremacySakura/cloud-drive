import { createRouter, createWebHistory } from 'vue-router'
import { routes } from './route.ts'
import { checkLogin } from '../services/apis/auth'
import { useUserStore } from '../stores/user.ts'

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: routes,
})

router.beforeEach(async (to) => {
  const userStore = useUserStore()

  if (to.meta.requiresAuth) {
    if (!userStore.token) {
      return { path: '/require-login' }
    }
    
    try {
      const res = await checkLogin()
      if (res.code !== 0) {
        userStore.logout()
        return { path: '/require-login' }
      }
    } catch {
      userStore.logout()
      return { path: '/require-login' }
    }
  }

  return true
})

export default router
