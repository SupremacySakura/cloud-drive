import { describe, it, expect, vi, afterEach } from 'vitest'
import request from '../request'
import { deleteAIProvider, saveAIProvider, getAIConfiguration } from './aiSettings'
import type { ProviderPayload } from './aiSettings'
vi.mock('../request', () => ({
  default: { get: vi.fn(), post: vi.fn(), put: vi.fn(), delete: vi.fn() },
}))
afterEach(() => vi.clearAllMocks())
describe('aiSettings API', () => {
  it('uses the provider ID on update and sends deletion versions in the body', async () => {
    const data = { providers: [], settings: {} }
    vi.mocked(request.post).mockResolvedValue({ data: { code: 0, data } })
    vi.mocked(request.put).mockResolvedValue({ data: { code: 0, data } })
    vi.mocked(request.delete).mockResolvedValue({ data: { code: 0, data } })
    const payload = { provider_id: 'gateway', expected_version: 5 } as ProviderPayload
    const signal = new AbortController().signal
    await saveAIProvider(null, payload, signal)
    await saveAIProvider(7, payload, signal)
    await deleteAIProvider(7, 6, signal)
    expect(request.post).toHaveBeenCalledWith('/api/settings/ai/providers', payload, { signal })
    expect(request.put).toHaveBeenCalledWith('/api/settings/ai/providers/7', payload, { signal })
    expect(request.delete).toHaveBeenCalledWith('/api/settings/ai/providers/7', {
      data: { expected_version: 6 },
      signal,
    })
  })
  it('rejects business failures without treating them as saved configuration', async () => {
    vi.mocked(request.get).mockResolvedValue({ data: { code: 1004, msg: '配置不可用' } })
    await expect(getAIConfiguration()).rejects.toThrow('配置不可用')
  })
})
