import { onScopeDispose, ref } from 'vue'
import axios from 'axios'
import {
  getAIConfiguration,
  getAIProviderCatalog,
  saveAIProvider,
  deleteAIProvider,
  saveAIDefaults,
  type AIConfiguration,
  type ProviderPreset,
  type ProviderPayload,
  type DefaultsPayload,
} from '../services/apis/aiSettings'

export function aiSettingsError(error: unknown) {
  if (axios.isAxiosError<{ msg?: string }>(error))
    return error.response?.data.msg || '请求未完成，请检查网络并重新加载配置后重试'
  return error instanceof Error ? error.message : '模型设置暂时不可用'
}
export function useAISettings() {
  const configuration = ref<AIConfiguration | null>(null)
  const catalog = ref<ProviderPreset[]>([])
  const loading = ref(false)
  const busy = ref(false)
  const error = ref('')
  const mutationError = ref('')
  const conflict = ref(false)
  let controller: AbortController | null = null
  let disposed = false
  async function load() {
    if (busy.value) return false
    controller?.abort()
    const current = new AbortController()
    controller = current
    loading.value = true
    error.value = ''
    try {
      const [data, presets] = await Promise.all([
        getAIConfiguration(current.signal),
        getAIProviderCatalog(current.signal),
      ])
      if (current.signal.aborted) return false
      configuration.value = data
      catalog.value = presets
      mutationError.value = ''
      conflict.value = false
      return true
    } catch (e) {
      if (!current.signal.aborted) error.value = aiSettingsError(e)
      return false
    } finally {
      if (!current.signal.aborted) loading.value = false
    }
  }
  async function mutate(operation: (signal: AbortSignal) => Promise<AIConfiguration>) {
    if (busy.value || loading.value || disposed) return false
    controller?.abort()
    const current = new AbortController()
    controller = current
    busy.value = true
    mutationError.value = ''
    conflict.value = false
    try {
      const data = await operation(current.signal)
      if (current.signal.aborted) return false
      configuration.value = data
      return true
    } catch (e) {
      if (!current.signal.aborted) {
        mutationError.value = aiSettingsError(e)
        conflict.value = axios.isAxiosError(e) && e.response?.status === 409
      }
      return false
    } finally {
      if (!current.signal.aborted) busy.value = false
    }
  }
  function clearError() {
    mutationError.value = ''
    conflict.value = false
  }
  onScopeDispose(() => {
    disposed = true
    controller?.abort()
  })
  return {
    configuration,
    catalog,
    loading,
    busy,
    error,
    mutationError,
    conflict,
    load,
    clearError,
    saveProvider: (id: number | null, payload: ProviderPayload) =>
      mutate(signal => saveAIProvider(id, payload, signal)),
    removeProvider: (id: number) =>
      mutate(signal => deleteAIProvider(id, configuration.value!.settings.config_version, signal)),
    saveDefaults: (payload: DefaultsPayload) => mutate(signal => saveAIDefaults(payload, signal)),
  }
}
