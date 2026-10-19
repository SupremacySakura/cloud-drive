import { describe, it, expect } from 'vitest'
import { readChatStream } from './chat'
function response(parts: string[]) {
  const encoder = new TextEncoder()
  return new Response(
    new ReadableStream({
      start(controller) {
        for (const part of parts) controller.enqueue(encoder.encode(part))
        controller.close()
      },
    }),
    { headers: { 'Content-Type': 'application/x-ndjson' } },
  )
}
describe('chat stream', () => {
  it('handles frames split across chunks and Unicode', async () => {
    const events: unknown[] = []
    await readChatStream(
      response(['{"type":"del', 'ta","text":"你好"}\n{"type":"done"}', '\n']),
      e => events.push(e),
    )
    expect(events).toEqual([{ type: 'delta', text: '你好' }, { type: 'done' }])
  })
  it('does not treat a truncated stream as completed', async () => {
    await expect(
      readChatStream(response(['{"type":"delta","text":"部分回复"}\n']), () => {}),
    ).rejects.toThrow('中断')
  })
  it('surfaces sanitized server errors without automatic retries', async () => {
    await expect(
      readChatStream(response(['{"type":"error","message":"模型调用失败"}\n']), () => {}),
    ).rejects.toThrow('模型调用失败')
  })
})
