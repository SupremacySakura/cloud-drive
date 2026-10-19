import { useUserStore } from '../../stores/user'

export interface ChatWireMessage {
  role: 'user' | 'assistant'
  content: string
}
export type ChatEvent =
  | { type: 'delta'; text: string }
  | { type: 'meta'; model_id: number; model: string; provider: string }
  | { type: 'done' }
  | { type: 'error'; message: string }

export async function readChatStream(response: Response, onEvent: (event: ChatEvent) => void) {
  if (!response.ok) {
    const data = await response.json().catch(() => null)
    throw new Error(data?.msg || '对话请求失败，请检查登录状态与模型服务')
  }
  if (!response.body || !response.headers.get('Content-Type')?.includes('application/x-ndjson'))
    throw new Error('对话服务返回格式无效')
  const reader = response.body.getReader()
  const decoder = new TextDecoder()
  let buffer = ''
  let completed = false
  function consume(line: string) {
    if (!line.trim()) return
    const event = JSON.parse(line) as ChatEvent
    if (event.type === 'error') throw new Error(event.message)
    if (event.type === 'done') completed = true
    if (event.type === 'delta' && typeof event.text !== 'string') throw new Error('回复格式无效')
    onEvent(event)
  }
  try {
    for (;;) {
      const { value, done } = await reader.read()
      buffer += done ? decoder.decode() : decoder.decode(value, { stream: true })
      let split: number
      while ((split = buffer.indexOf('\n')) !== -1) {
        consume(buffer.slice(0, split))
        buffer = buffer.slice(split + 1)
      }
      if (buffer.length > 1024 * 1024) throw new Error('回复事件过大')
      if (done) break
    }
    if (buffer.trim()) consume(buffer)
    if (!completed) throw new Error('回复连接中断，已收到的内容保留，请手动重试')
  } finally {
    await reader.cancel().catch(() => {})
  }
}
export async function streamChat(
  modelID: number,
  messages: ChatWireMessage[],
  signal: AbortSignal,
  onEvent: (event: ChatEvent) => void,
) {
  const user = useUserStore()
  // Dedicated fetch avoids Axios's automatic retries and 30-second timeout.
  const response = await fetch('/api/chat', {
    method: 'POST',
    signal,
    headers: { 'Content-Type': 'application/json', Authorization: `Bearer ${user.token}` },
    body: JSON.stringify({ model_id: modelID, messages }),
  })
  if (response.status === 401) user.logout()
  await readChatStream(response, onEvent)
}
