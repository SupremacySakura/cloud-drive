import { describe, it, expect, beforeEach, vi } from 'vitest'
import { setActivePinia, createPinia } from 'pinia'
import axios from 'axios'
import { useUserStore } from '../../stores/user'

// Mock axios
vi.mock('axios', async () => {
  const actual = await vi.importActual('axios')
  return {
    default: {
      ...actual,
      create: vi.fn(() => ({
        interceptors: {
          request: { use: vi.fn() },
          response: { use: vi.fn() },
        },
        defaults: { timeout: 30000 },
        post: vi.fn(),
      })),
      post: vi.fn(),
    },
  }
})

// Mock window.location
Object.defineProperty(window, 'location', {
  writable: true,
  value: { href: '' },
})

describe('request', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    window.location.href = ''
    vi.clearAllMocks()
  })

  it('should create axios instance with correct timeout', async () => {
    const axiosCreateSpy = vi.spyOn(axios, 'create')
    await import('../request')
    expect(axiosCreateSpy).toHaveBeenCalledWith(
      expect.objectContaining({
        timeout: 30000,
      }),
    )
  })

  it('should inject token into request headers when token exists', async () => {
    const store = useUserStore()
    store.setToken('test-bearer-token')

    // 重新加载模块以获取拦截器
    vi.resetModules()
    const requestModule = await import('../request')
    const request = requestModule.default

    // 模拟请求拦截器调用
    const mockConfig = { headers: {} }
    const requestInterceptor = (request.interceptors.request.use as any).mock.calls[0][0]
    const result = requestInterceptor(mockConfig)

    expect(result.headers.Authorization).toBe('Bearer test-bearer-token')
  })

  it('should not inject token when token is empty', async () => {
    const store = useUserStore()
    store.logout()

    vi.resetModules()
    const requestModule = await import('../request')
    const request = requestModule.default

    const mockConfig = { headers: {} }
    const requestInterceptor = (request.interceptors.request.use as any).mock.calls[0][0]
    const result = requestInterceptor(mockConfig)

    expect(result.headers.Authorization).toBeUndefined()
  })

  it('should handle 401 response by attempting token refresh', async () => {
    const store = useUserStore()
    store.setToken('old-token')
    // @ts-expect-error -- Pinia store type inference issue in test
    store
      .setRefreshToken('refresh-token-123')(
        // Mock successful token refresh
        axios.post as any,
      )
      .mockResolvedValueOnce({
        data: {
          data: {
            token: 'new-token-456',
            refreshToken: 'new-refresh-token',
          },
        },
      })

    vi.resetModules()
    const requestModule = await import('../request')
    const request = requestModule.default

    const responseInterceptor = (request.interceptors.response.use as any).mock.calls[0][1]

    const mockError = {
      response: { status: 401 },
      config: { headers: {} },
    }

    // 由于 responseInterceptor 是异步的，我们需要等待它
    await expect(responseInterceptor(mockError)).rejects.toBeDefined()

    // 验证 token 刷新请求被调用
    expect(axios.post).toHaveBeenCalledWith('/auth/refresh', { refreshToken: 'refresh-token-123' })
  })

  it('should logout and redirect to login on 401 when no refresh token', async () => {
    const store = useUserStore()
    store.setToken('some-token')
    store.logout() // 清除所有 token

    vi.resetModules()
    const requestModule = await import('../request')
    const request = requestModule.default

    const responseInterceptor = (request.interceptors.response.use as any).mock.calls[0][1]

    const mockError = {
      response: { status: 401 },
      config: { headers: {} },
    }

    await expect(responseInterceptor(mockError)).rejects.toBe(mockError)
    expect(window.location.href).toBe('/login')
  })

  it('should handle 403 response with toast', async () => {
    const consoleSpy = vi.spyOn(console, 'warn').mockImplementation(() => {})

    vi.resetModules()
    const requestModule = await import('../request')
    const request = requestModule.default

    const responseInterceptor = (request.interceptors.response.use as any).mock.calls[0][1]

    const mockError = {
      response: { status: 403 },
      config: {},
    }

    await expect(responseInterceptor(mockError)).rejects.toBe(mockError)
    expect(consoleSpy).toHaveBeenCalledWith('Toast: ', '无权限')

    consoleSpy.mockRestore()
  })

  it('should handle 500 response with toast', async () => {
    const consoleSpy = vi.spyOn(console, 'warn').mockImplementation(() => {})

    vi.resetModules()
    const requestModule = await import('../request')
    const request = requestModule.default

    const responseInterceptor = (request.interceptors.response.use as any).mock.calls[0][1]

    const mockError = {
      response: { status: 500 },
      config: {},
    }

    await expect(responseInterceptor(mockError)).rejects.toBe(mockError)
    expect(consoleSpy).toHaveBeenCalledWith('Toast: ', '服务器错误')

    consoleSpy.mockRestore()
  })

  it('should extract and store tokens from login response', async () => {
    const store = useUserStore()

    vi.resetModules()
    const requestModule = await import('../request')
    const request = requestModule.default

    const successInterceptor = (request.interceptors.response.use as any).mock.calls[0][0]

    const mockResponse = {
      config: { url: '/auth/login' },
      data: {
        data: {
          token: 'login-token',
          refreshToken: 'login-refresh-token',
        },
      },
    }

    successInterceptor(mockResponse)

    expect(store.token).toBe('login-token')
    expect(store.refreshToken).toBe('login-refresh-token')
  })
})
