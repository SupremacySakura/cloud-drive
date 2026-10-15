import { computed, ref } from 'vue'
import { getOperationLogs } from '../services/apis/log'
import type { OperationLogItem } from '../services/types/log'

/** 操作日志首屏条数（首页 feed 一次拉取的量） */
export const OPERATION_LOG_PAGE_SIZE = 20

/**
 * 首页「操作日志」feed 数据源。
 * 加载失败静默回退（与 getOperationLogs 契约一致），不抛出异常。
 * load / loadMore / reset 共享同一个在途控制器：任何新操作都会 abort 前一个未完成的请求，
 * 被 abort 的请求返回后不再提交数据，避免旧数据回灌新列表。
 */
export const useOperationLogs = () => {
  const logs = ref<OperationLogItem[]>([])
  const total = ref(0)
  const isLoading = ref(false)
  const isLoadingMore = ref(false)
  const hasMore = computed(() => logs.value.length < total.value)

  let page = 0
  let inflightController: AbortController | null = null

  const reset = () => {
    inflightController?.abort()
    inflightController = null
    page = 0
    logs.value = []
    total.value = 0
    isLoading.value = false
    isLoadingMore.value = false
  }

  const load = async () => {
    inflightController?.abort()
    const controller = new AbortController()
    inflightController = controller
    isLoading.value = true
    try {
      const data = await getOperationLogs(1, OPERATION_LOG_PAGE_SIZE, controller.signal)
      if (controller.signal.aborted) return
      page = 1
      logs.value = data.list
      total.value = data.total
    } catch {
      // 请求失败静默回退空列表；仅当请求未被更新的 load 取代时才清空
      if (!controller.signal.aborted) {
        logs.value = []
        total.value = 0
      }
    } finally {
      if (inflightController === controller) {
        inflightController = null
        isLoading.value = false
      }
    }
  }

  const loadMore = async () => {
    if (!hasMore.value || isLoading.value || isLoadingMore.value) return
    const controller = new AbortController()
    inflightController = controller
    isLoadingMore.value = true
    try {
      const nextPage = page + 1
      const data = await getOperationLogs(nextPage, OPERATION_LOG_PAGE_SIZE)
      // reset()/load() 可能已把本请求判为过期：被 abort 则不提交，防止旧数据回灌
      if (controller.signal.aborted) return
      page = nextPage
      logs.value = [...logs.value, ...data.list]
      total.value = data.total
    } catch {
      // 加载更多失败静默：保留已加载内容，用户可重试
    } finally {
      // loadMore 同一实例内串行（入口有 isLoadingMore 护栏），无条件复位是安全的
      isLoadingMore.value = false
      if (inflightController === controller) {
        inflightController = null
      }
    }
  }

  return { logs, total, isLoading, isLoadingMore, hasMore, load, loadMore, reset }
}
