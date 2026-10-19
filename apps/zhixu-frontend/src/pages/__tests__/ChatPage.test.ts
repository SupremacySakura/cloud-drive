import { mount, flushPromises } from '@vue/test-utils'
import { afterEach, describe, expect, it, vi } from 'vitest'
import ChatPage from '../ChatPage.vue'
import { getAIConfiguration, getAIProviderCatalog } from '../../services/apis/aiSettings'
import { streamChat } from '../../services/apis/chat'
vi.mock('../../services/apis/aiSettings', () => ({
  getAIConfiguration: vi.fn(),
  getAIProviderCatalog: vi.fn(),
  saveAIProvider: vi.fn(),
  saveAIDefaults: vi.fn(),
  deleteAIProvider: vi.fn(),
}))
vi.mock('../../services/apis/chat', () => ({ streamChat: vi.fn() }))
vi.mock('@iconify/vue', () => ({ Icon: { template: '<span />' } }))
const global = { stubs: { RouterLink: { template: '<a><slot /></a>' }, teleport: true } }
afterEach(() => vi.clearAllMocks())
describe('ChatPage', () => {
  it('shows configuration guidance without requiring embedding', async () => {
    vi.mocked(getAIConfiguration).mockResolvedValue({
      providers: [],
      settings: {
        chat_model_id: 0,
        embedding_model_id: 0,
        config_version: 0,
        embedding_version: 0,
        needs_reindex: false,
      },
    })
    vi.mocked(getAIProviderCatalog).mockResolvedValue([])
    const wrapper = mount(ChatPage, { global })
    await flushPromises()
    expect(wrapper.text()).toContain('普通对话不需要 Embedding 配置')
    expect(wrapper.get('textarea').attributes('disabled')).toBeDefined()
    expect(wrapper.find('select').exists()).toBe(false)
    wrapper.unmount()
  })
  it('submits the configured model and displays actual streamed text', async () => {
    vi.mocked(getAIProviderCatalog).mockResolvedValue([])
    vi.mocked(getAIConfiguration).mockResolvedValue({
      providers: [
        {
          id: 1,
          provider_id: 'gateway',
          catalog_id: '',
          name: '模型网关',
          base_url: 'https://example.com',
          protocol: 'chat',
          auth_type: 'none',
          credential_configured: false,
          models: [{ id: 2, provider_id: 1, model_id: 'chat', kind: 'chat', dimension: 0 }],
        },
      ],
      settings: {
        chat_model_id: 2,
        embedding_model_id: 0,
        config_version: 1,
        embedding_version: 0,
        needs_reindex: false,
      },
    })
    vi.mocked(streamChat).mockImplementation(async (_model, _messages, _signal, emit) => {
      emit({ type: 'meta', model_id: 2, model: 'chat', provider: '模型网关' })
      emit({ type: 'delta', text: '真实流内容' })
      emit({ type: 'done' })
    })
    const wrapper = mount(ChatPage, { global })
    await flushPromises()
    await wrapper.get('textarea').setValue('你好')
    await wrapper.get('form').trigger('submit')
    await flushPromises()
    expect(streamChat).toHaveBeenCalledWith(
      2,
      [{ role: 'user', content: '你好' }],
      expect.any(AbortSignal),
      expect.any(Function),
    )
    expect(wrapper.text()).toContain('真实流内容')
    expect(wrapper.text()).toContain('回复完成')
    wrapper.unmount()
  })
})
