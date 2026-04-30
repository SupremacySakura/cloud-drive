import axios from 'axios'
import { useUserStore } from '../stores/user'

// 简单的全局 toast 展示，尽量复用现有应用的 toast 机制
function showToast(message: string) {
  try {
    const anyWindow: any = window
    // 常见的全局 toast API 兜底方案
    if (anyWindow.$toast && typeof anyWindow.$toast === 'function') {
      anyWindow.$toast(message)
      return
    }
    if (anyWindow.$toastStore && typeof anyWindow.$toastStore === 'function') {
      anyWindow.$toastStore(message)
      return
    }
  } catch {
    // 忽略自定义 toast 加载失败的情况
  }
  // 兜底：尽量不打断体验，使用 console 日志，必要时可改成 alert/页面导航
  console.warn('Toast: ', message)
}

function redirectToLogin() {
  // 使用硬跳转，确保路由跳转在拦截器中可控
  window.location.href = '/login'
}

// token 刷新逻辑：仅在有 refreshToken 时尝试刷新
async function attemptRefresh(): Promise<boolean> {
  const store = useUserStore()
  const refreshToken = (store as any).refreshToken || ''
  if (!refreshToken) return false
  try {
    const res = await axios.post('/auth/refresh', { refreshToken })
    const data = res?.data?.data || {}
    if (data?.token) {
      store.setToken(data.token)
    }
    if (data?.refreshToken && (store as any).setRefreshToken) {
      ;(store as any).setRefreshToken(data.refreshToken)
    }
    return true
  } catch {
    return false
  }
}

// 请求实例
const request = axios.create({
  timeout: 30000, // 请求超时 30s
})

// 请求拦截：附带 token
request.interceptors.request.use(
  config => {
    const token = useUserStore().token
    if (token) {
      config.headers.Authorization = `Bearer ${token}`
    }
    return config
  },
  error => {
    return Promise.reject(error)
  },
)

request.interceptors.response.use(
  config => {
    // 更新 token/refreshToken 时机：登录、注册、刷新
    const url = config?.config?.url
    if (
      url &&
      (url.includes('/auth/login') ||
        url.includes('/auth/register') ||
        url.includes('/auth/refresh'))
    ) {
      const data = config?.data?.data || {}
      if (data?.token) {
        useUserStore().setToken(data.token)
      }
      if (data?.refreshToken && (useUserStore() as any).setRefreshToken) {
        ;(useUserStore() as any).setRefreshToken(data.refreshToken)
      }
    }
    return config
  },
  async error => {
    const { config, response, message, code } = error
    const status = response?.status

    // 全局错误处理
    if (status === 403) {
      showToast('无权限')
      return Promise.reject(error)
    }
    if (status === 500) {
      showToast('服务器错误')
      return Promise.reject(error)
    }

    // 网络错误重试：指数退避，最多2次
    const isNetwork =
      !response && (code === 'ECONNABORTED' || (message ?? '').toString().includes('Network Error'))
    if (isNetwork) {
      ;(config as any).__retryCount = (config as any).__retryCount || 0
      if ((config as any).__retryCount < 2) {
        ;(config as any).__retryCount += 1
        const backoff = Math.pow(2, (config as any).__retryCount) * 1000
        await new Promise(resolve => setTimeout(resolve, backoff))
        return request(config)
      }
    }

    // 401：尝试刷新令牌
    if (status === 401) {
      const refreshed = await attemptRefresh()
      if (refreshed) {
        ;(config as any)._retryAuth = true
        const t = useUserStore().token
        if (t) {
          config.headers.Authorization = `Bearer ${t}`
        }
        return request(config)
      } else {
        // 刷新失败，登出
        showToast('请登录')
        useUserStore().logout()
        redirectToLogin()
        return Promise.reject(error)
      }
    }

    return Promise.reject(error)
  },
)

export default request
