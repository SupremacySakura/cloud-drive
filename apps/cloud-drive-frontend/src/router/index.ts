import { createRouter, createWebHistory } from 'vue-router'
// 导入路由
import { routes } from './route.ts'
import { checkLogin } from '../services/apis/auth'

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: routes,
})

router.beforeEach(async (to, _from, next) => {
  const requiresAuth = to.matched.some(record => record.meta.requiresAuth)

  if (!requiresAuth) {
    next()
    return
  }

  try {
    const res = await checkLogin()
    console.log(res)
    if (res.code === 0) {
      next()
    } else {
      next({
        path: '/require-login',
        query: { redirect: to.fullPath }
      })
    }
  } catch (error) {
    next({
      path: '/require-login',
      query: { redirect: to.fullPath }
    })
  }
})

export default router
