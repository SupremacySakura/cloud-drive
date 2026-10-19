import { effectScope } from 'vue'
import { flushPromises } from '@vue/test-utils'
import { afterEach, describe, expect, it, vi } from 'vitest'
import { useKnowledgeLibrary } from '../useKnowledgeLibrary'
import { useKnowledgeImport } from '../useKnowledgeImport'
import {
  listKnowledge,
  addKnowledge,
  getKnowledge,
  type KnowledgeItem,
} from '../../services/apis/knowledge'
vi.mock('../../services/apis/knowledge', () => ({
  listKnowledge: vi.fn(),
  addKnowledge: vi.fn(),
  getKnowledge: vi.fn(),
  retryKnowledge: vi.fn(),
}))
const item = { id: 1, title: '笔记', status: 'pending' } as KnowledgeItem
afterEach(() => {
  vi.useRealTimers()
  vi.clearAllMocks()
})
describe('knowledge lifecycle', () => {
  it('ignores an older list request after a new load', async () => {
    let resolve!: (value: {
      items: KnowledgeItem[]
      total: number
      page: number
      page_size: number
    }) => void
    vi.mocked(listKnowledge)
      .mockImplementationOnce(
        () =>
          new Promise(r => {
            resolve = r
          }),
      )
      .mockResolvedValueOnce({
        items: [{ ...item, id: 2, status: 'ready' }],
        total: 1,
        page: 1,
        page_size: 20,
      })
    const scope = effectScope()
    const library = scope.run(useKnowledgeLibrary)!
    const first = library.load()
    await library.load()
    resolve({ items: [item], total: 1, page: 1, page_size: 20 })
    await first
    expect(library.items.value[0]?.id).toBe(2)
    scope.stop()
  })
  it('preserves content and shows a visible loading error', async () => {
    vi.mocked(listKnowledge).mockRejectedValue(new Error('网络异常'))
    const scope = effectScope()
    const library = scope.run(useKnowledgeLibrary)!
    await library.load()
    expect(library.error.value).toBe('网络异常')
    expect(library.loading.value).toBe(false)
    scope.stop()
  })
  it('polls to completion even after the panel closes and stops on dispose', async () => {
    vi.useFakeTimers()
    vi.mocked(getKnowledge).mockResolvedValue({ ...item, status: 'ready' })
    const scope = effectScope()
    const flow = scope.run(() => useKnowledgeImport(vi.fn()))!
    flow.track(item)
    flow.panelOpen.value = false
    await vi.advanceTimersByTimeAsync(1500)
    await flushPromises()
    expect(flow.item.value?.status).toBe('ready')
    scope.stop()
    await vi.advanceTimersByTimeAsync(10000)
    expect(getKnowledge).toHaveBeenCalledTimes(1)
  })
  it('validates titles before making a request', async () => {
    const scope = effectScope()
    const flow = scope.run(() => useKnowledgeImport(vi.fn()))!
    flow.target.value = { id: 1, name: 'note.md' } as never
    flow.title.value = '知'.repeat(201)
    await flow.submit()
    expect(flow.error.value).toContain('1–200')
    expect(addKnowledge).not.toHaveBeenCalled()
    scope.stop()
  })
})
