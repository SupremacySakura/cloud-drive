import { createRouter, createWebHistory } from 'vue-router'
// 导入路由
import { routes } from './route.ts'
import { checkLogin } from '../services/apis/auth'
import { useUserStore } from '../stores/user.ts'

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: routes,
})

router.beforeEach(async (_to, _from, next) => {
  const userStore = useUserStore()

  if (userStore.token) {
    try {
      const res = await checkLogin()
      if (res.code !== 0) {
        userStore.logout()
      }
    } catch {
      userStore.logout()
    }
  }

  next()
})

export default router
