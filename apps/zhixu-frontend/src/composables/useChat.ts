import { ref, onScopeDispose } from 'vue'
import { streamChat, type ChatWireMessage } from '../services/apis/chat'
export interface ChatMessage {
  role: 'user' | 'assistant'
  content: string
  model?: string
  status?: 'generating' | 'completed' | 'stopped' | 'failed'
  error?: string
}
export function useChat() {
  const messages = ref<ChatMessage[]>([])
  const running = ref(false)
  let controller: AbortController | null = null
  function stop() {
    controller?.abort()
    const last = messages.value.at(-1)
    if (last?.status === 'generating') last.status = 'stopped'
    running.value = false
  }
  function reset() {
    stop()
    messages.value = []
  }
  async function generate(modelID: number, label: string) {
    const current = new AbortController()
    controller = current
    running.value = true
    const history: ChatWireMessage[] = messages.value
      .filter(m => m.role === 'user' || m.status === 'completed')
      .map(m => ({ role: m.role, content: m.content }))
    const message: ChatMessage = {
      role: 'assistant',
      content: '',
      model: label,
      status: 'generating',
    }
    messages.value.push(message)
    const index = messages.value.length - 1
    try {
      await streamChat(modelID, history, current.signal, event => {
        if (current.signal.aborted) return
        const reply = messages.value[index]!
        if (event.type === 'delta') {
          if (reply.content.length + event.text.length > 200000)
            throw new Error('回复过长，请开始新对话')
          reply.content += event.text
        }
        if (event.type === 'meta') reply.model = `${event.provider} / ${event.model}`
      })
      if (!current.signal.aborted) {
        const reply = messages.value[index]!
        if (!reply.content.trim())
          throw new Error('模型未返回公开文字回复，请检查模型是否支持普通对话')
        reply.status = 'completed'
      }
    } catch (e) {
      if (!current.signal.aborted) {
        const reply = messages.value[index]!
        reply.status = 'failed'
        reply.error = e instanceof Error ? e.message : '模型调用失败，请手动重试'
      }
    } finally {
      if (controller === current) running.value = false
    }
  }
  async function send(text: string, modelID: number, label: string) {
    if (running.value || !text.trim()) return false
    const last = messages.value.at(-1)
    if (last?.role === 'assistant' && last.status !== 'completed') return false
    messages.value.push({ role: 'user', content: text.trim() })
    await generate(modelID, label)
    return true
  }
  async function retry(index: number, modelID: number, label: string) {
    if (running.value || messages.value[index]?.role !== 'assistant') return
    messages.value.splice(index)
    await generate(modelID, label)
  }
  onScopeDispose(stop)
  return { messages, running, stop, reset, send, retry }
}
