import { effectScope } from 'vue'
import { describe, it, expect, vi, afterEach } from 'vitest'
import { useChat } from '../useChat'
import { streamChat } from '../../services/apis/chat'
vi.mock('../../services/apis/chat', () => ({ streamChat: vi.fn() }))
afterEach(() => vi.clearAllMocks())
describe('temporary chat', () => {
  it('sends completed multi-turn context and never duplicates the user on retry', async () => {
    vi.mocked(streamChat).mockImplementation(async (_id, _messages, _signal, emit) => {
      emit({ type: 'delta', text: '你好' })
      emit({ type: 'done' })
    })
    const scope = effectScope()
    const chat = scope.run(useChat)!
    await chat.send('第一轮', 2, 'provider/model')
    await chat.send('第二轮', 2, 'provider/model')
    expect(vi.mocked(streamChat).mock.calls[1]![1]).toEqual([
      { role: 'user', content: '第一轮' },
      { role: 'assistant', content: '你好' },
      { role: 'user', content: '第二轮' },
    ])
    vi.mocked(streamChat).mockRejectedValueOnce(new Error('认证失败'))
    await chat.send('第三轮', 2, 'provider/model')
    expect(chat.messages.value.at(-1)?.status).toBe('failed')
    await chat.retry(5, 2, 'provider/model')
    expect(chat.messages.value).toHaveLength(6)
    scope.stop()
  })
  it('aborts on reset and prevents stale deltas from reviving cleared messages', async () => {
    let finish!: () => void
    vi.mocked(streamChat).mockImplementation(
      (_id, _messages, _signal, emit) =>
        new Promise<void>(resolve => {
          finish = () => {
            emit({ type: 'delta', text: '旧回复' })
            resolve()
          }
        }),
    )
    const scope = effectScope()
    const chat = scope.run(useChat)!
    const task = chat.send('你好', 2, 'provider/model')
    const signal = vi.mocked(streamChat).mock.calls[0]![2]
    chat.reset()
    expect(signal.aborted).toBe(true)
    finish()
    await task
    expect(chat.messages.value).toEqual([])
    scope.stop()
  })
})
