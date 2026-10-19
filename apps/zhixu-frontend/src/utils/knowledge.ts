import axios from 'axios'
import type { KnowledgeStatus } from '../services/apis/knowledge'

export const knowledgeLabels: Record<KnowledgeStatus, string> = {
  pending: '排队中',
  running: '处理中',
  ready: '已加入',
  failed: '加入失败',
}
export const isMarkdownFile = (file: { type: string; name: string }) =>
  file.type === 'file' && /\.(md|markdown)$/i.test(file.name)
export const isKnowledgeActive = (status: KnowledgeStatus) =>
  status === 'pending' || status === 'running'
export const knowledgeError = (error: unknown) => {
  if (axios.isAxiosError<{ msg?: string }>(error))
    return error.response?.data.msg || '请求失败，请检查网络后重试'
  return error instanceof Error ? error.message : '请求失败，请稍后重试'
}
export const knowledgeDate = (date: string | null) => {
  if (!date) return '尚未完成'
  const value = new Date(date)
  return Number.isNaN(value.getTime()) ? '未知时间' : value.toLocaleString('zh-CN')
}
