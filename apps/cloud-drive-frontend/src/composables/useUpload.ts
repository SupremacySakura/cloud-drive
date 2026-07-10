import { computed, getCurrentScope, onScopeDispose, reactive, ref } from 'vue'
import { getListByFolderIDAndUserID, makeDirectory, uploadFile } from '../services/apis/file'
import type { ToastType, UploadTask } from '../types/file'
import { detectFileType } from '../utils/file'
import { createId } from '../utils/hash'

type MaybePromise<T> = T | Promise<T>

export type UseUploadOptions = {
  refresh?: () => MaybePromise<void>
  notify?: (message: string, type: ToastType) => MaybePromise<void>
}

type Deferred = {
  promise: Promise<void>
  resolve: () => void
}

const MAX_CONCURRENT_UPLOADS = 3

const createDeferred = (): Deferred => {
  let resolve!: () => void
  const promise = new Promise<void>(done => {
    resolve = done
  })
  return { promise, resolve }
}

const normalizeFolderPath = (path: string) => {
  return path.replace(/\\/g, '/').replace(/^\/+|\/+$/g, '')
}

const folderPathForFile = (file: File) => {
  const relativePath = normalizeFolderPath(file.webkitRelativePath || file.name)
  const lastSlashIndex = relativePath.lastIndexOf('/')
  return lastSlashIndex > 0 ? relativePath.slice(0, lastSlashIndex) : ''
}

export function useUpload(options: UseUploadOptions = {}) {
  const uploadTasks = ref<UploadTask[]>([])
  const isUploadPanelOpen = ref(false)
  const isUploading = ref(false)
  const fileInputRef = ref<HTMLInputElement | null>(null)
  const folderInputRef = ref<HTMLInputElement | null>(null)
  const activeUploadCount = ref(0)

  const pendingQueue: UploadTask[] = []
  const queuedTaskIds = new Set<string>()
  const canceledTaskIds = new Set<string>()
  const retryAfterCancelTaskIds = new Set<string>()
  const activeControllers = new Map<string, AbortController>()
  const cycleTasks = new Map<string, UploadTask>()
  const folderCache = new Map<string, number>()
  const folderLocks = new Map<string, Promise<number>>()
  const managedTasks = new Map<string, UploadTask>()

  let activeWorkerCount = 0
  let cycleDeferred: Deferred | null = null
  let finalizationPromise: Promise<void> | null = null
  let disposed = false

  const overallProgress = computed(() => {
    if (uploadTasks.value.length === 0) return 0
    const total = uploadTasks.value.reduce((sum, task) => sum + task.percent, 0)
    return Math.floor(total / uploadTasks.value.length)
  })

  const openFileDialog = () => fileInputRef.value?.click()
  const openFolderDialog = () => folderInputRef.value?.click()

  const createTasks = (files: File[], baseFolderId: number, includeFolderPath: boolean) => {
    return files.map<UploadTask>(file => ({
      id: createId(),
      file,
      targetFolderId: baseFolderId,
      relativePath: includeFolderPath ? folderPathForFile(file) : '',
      status: 'pending',
      percent: 0,
      message: '等待中',
    }))
  }

  const createFileTasks = (files: File[], baseFolderId: number) => {
    return createTasks(files, baseFolderId, false)
  }

  const createFolderTasks = (files: File[], baseFolderId: number) => {
    return createTasks(files, baseFolderId, true)
  }

  const resetTaskForQueue = (task: UploadTask) => {
    canceledTaskIds.delete(task.id)
    task.status = 'pending'
    task.percent = 0
    task.message = '等待中'
  }

  const ensureCycle = () => {
    cycleDeferred ??= createDeferred()
    return cycleDeferred
  }

  const summaryForTasks = (tasks: UploadTask[], refreshFailed: boolean) => {
    const successCount = tasks.filter(task => task.status === 'success').length
    const failedCount = tasks.filter(task => task.status === 'failed').length
    const canceledCount = tasks.filter(task => task.status === 'canceled').length

    if (!refreshFailed && failedCount === 0 && canceledCount === 0) {
      return {
        message: `成功上传 ${successCount} 个文件`,
        type: 'success' as const,
      }
    }

    const parts = [`${successCount} 个成功`, `${failedCount} 个失败`]
    if (canceledCount > 0) parts.push(`${canceledCount} 个取消`)
    if (refreshFailed) parts.push('列表刷新失败')
    return {
      message: `上传完成：${parts.join('，')}`,
      type: failedCount > 0 || refreshFailed ? ('error' as const) : ('info' as const),
    }
  }

  const maybeFinalizeCycle = () => {
    if (
      disposed ||
      finalizationPromise ||
      !cycleDeferred ||
      pendingQueue.length > 0 ||
      activeWorkerCount > 0
    ) {
      return
    }

    const deferred = cycleDeferred
    cycleDeferred = null
    const completedTasks = Array.from(cycleTasks.values())
    cycleTasks.clear()
    folderCache.clear()

    const finalize = async () => {
      let refreshFailed = false
      try {
        await options.refresh?.()
      } catch {
        refreshFailed = true
      }

      if (!disposed && completedTasks.length > 0) {
        const summary = summaryForTasks(completedTasks, refreshFailed)
        try {
          await options.notify?.(summary.message, summary.type)
        } catch {
          // Notification errors must not leave the queue in a busy state.
        }
      }

      isUploading.value = false
      deferred.resolve()
    }

    const currentFinalization = finalize()
    const trackedFinalization = currentFinalization.finally(() => {
      if (finalizationPromise === trackedFinalization || disposed) {
        finalizationPromise = null
      }
    })
    finalizationPromise = trackedFinalization
  }

  const findExistingFolder = async (parentFolderId: number, name: string) => {
    const list = await getListByFolderIDAndUserID(parentFolderId, 1, 100)
    return list.find(item => item.type === 'folder' && item.name === name)?.id ?? 0
  }

  const getOrCreateFolder = async (
    cacheKey: string,
    parentFolderId: number,
    name: string,
  ): Promise<number> => {
    const cachedFolderId = folderCache.get(cacheKey)
    if (cachedFolderId !== undefined) return cachedFolderId

    const existingLock = folderLocks.get(cacheKey)
    if (existingLock) return existingLock

    const creation = (async () => {
      const createdFolderId = await makeDirectory({
        folder_id: parentFolderId,
        name,
      })

      const folderId =
        createdFolderId > 0 ? createdFolderId : await findExistingFolder(parentFolderId, name)

      if (folderId <= 0) {
        throw new Error(`创建文件夹 "${name}" 失败`)
      }

      folderCache.set(cacheKey, folderId)
      return folderId
    })()

    folderLocks.set(cacheKey, creation)
    try {
      return await creation
    } finally {
      if (folderLocks.get(cacheKey) === creation) {
        folderLocks.delete(cacheKey)
      }
    }
  }

  const ensureFolderStructure = async (baseFolderId: number, relativePath: string) => {
    const normalizedPath = normalizeFolderPath(relativePath)
    if (!normalizedPath) return baseFolderId

    const parts = normalizedPath.split('/').filter(Boolean)
    if (parts.some(part => part === '..')) {
      throw new Error('上传目录路径不合法')
    }

    let currentFolderId = baseFolderId
    let currentPath = ''
    for (const part of parts) {
      currentPath = currentPath ? `${currentPath}/${part}` : part
      const cacheKey = `${baseFolderId}/${currentPath}`
      currentFolderId = await getOrCreateFolder(cacheKey, currentFolderId, part)
    }
    return currentFolderId
  }

  const processUploadTask = async (task: UploadTask, controller: AbortController) => {
    try {
      let targetFolderId = task.targetFolderId
      if (task.relativePath) {
        targetFolderId = await ensureFolderStructure(task.targetFolderId, task.relativePath)
      }

      if (controller.signal.aborted || canceledTaskIds.has(task.id) || disposed) {
        task.status = 'canceled'
        task.message = '已取消'
        return
      }

      task.status = 'hashing'
      task.message = '计算文件哈希...'
      await uploadFile(
        task.file,
        {
          file_type: detectFileType(task.file),
          folder_id: targetFolderId,
        },
        progress => {
          if (controller.signal.aborted || canceledTaskIds.has(task.id) || disposed) return
          task.percent = Math.min(100, Math.max(0, Math.floor(progress)))
          if (task.percent <= 0) {
            task.status = 'hashing'
            task.message = '计算文件哈希...'
          } else if (task.percent < 100) {
            task.status = 'uploading'
            task.message = '上传中...'
          } else {
            task.status = 'merging'
            task.message = '处理中...'
          }
        },
        controller.signal,
      )

      if (controller.signal.aborted || canceledTaskIds.has(task.id) || disposed) {
        task.status = 'canceled'
        task.message = '已取消'
        return
      }

      task.status = 'success'
      task.percent = 100
      task.message = '上传完成'
    } catch (error: unknown) {
      if (controller.signal.aborted || canceledTaskIds.has(task.id) || disposed) {
        task.status = 'canceled'
        task.message = '已取消'
        return
      }
      task.status = 'failed'
      task.message = error instanceof Error ? error.message : '上传失败'
    }
  }

  const queueTask = (task: UploadTask) => {
    if (queuedTaskIds.has(task.id) || activeControllers.has(task.id)) return false
    resetTaskForQueue(task)
    pendingQueue.push(task)
    queuedTaskIds.add(task.id)
    cycleTasks.set(task.id, task)
    return true
  }

  const runWorker = async () => {
    while (!disposed) {
      const task = pendingQueue.shift()
      if (!task) return
      queuedTaskIds.delete(task.id)

      if (canceledTaskIds.has(task.id)) continue

      const controller = new AbortController()
      activeControllers.set(task.id, controller)
      activeUploadCount.value += 1
      try {
        await processUploadTask(task, controller)
      } finally {
        if (activeControllers.get(task.id) === controller) {
          activeControllers.delete(task.id)
          activeUploadCount.value = Math.max(0, activeUploadCount.value - 1)
        }

        if (retryAfterCancelTaskIds.delete(task.id) && !disposed) {
          queueTask(task)
        }
      }
    }
  }

  const ensureWorkers = () => {
    if (disposed) return
    while (activeWorkerCount < MAX_CONCURRENT_UPLOADS && pendingQueue.length > 0) {
      activeWorkerCount += 1
      void runWorker().finally(() => {
        activeWorkerCount -= 1
        if (pendingQueue.length > 0) ensureWorkers()
        maybeFinalizeCycle()
      })
    }
  }

  const getManagedTask = (task: UploadTask) => {
    const existingTask = managedTasks.get(task.id)
    if (existingTask) return existingTask
    const managedTask = reactive(task) as UploadTask
    managedTasks.set(task.id, managedTask)
    return managedTask
  }

  const addVisibleTasks = (tasks: UploadTask[]) => {
    const existingIds = new Set(uploadTasks.value.map(task => task.id))
    const newTasks = tasks.filter(task => !existingIds.has(task.id))
    if (newTasks.length > 0) uploadTasks.value.unshift(...newTasks)
  }

  const startUploadTasks = async (tasks: UploadTask[]) => {
    if (disposed || tasks.length === 0) return
    if (finalizationPromise) await finalizationPromise
    if (disposed) return

    const tasksToQueue = tasks.map(getManagedTask)
    addVisibleTasks(tasksToQueue)
    const deferred = ensureCycle()
    let added = false
    for (const task of tasksToQueue) {
      if (queueTask(task)) added = true
    }
    if (added) {
      isUploadPanelOpen.value = true
      isUploading.value = true
      ensureWorkers()
    }
    await deferred.promise
  }

  const handleFileInput = (files: File[], baseFolderId: number) => {
    const tasks = createFileTasks(files, baseFolderId)
    void startUploadTasks(tasks)
    return tasks
  }

  const handleFolderInput = (files: File[], baseFolderId: number) => {
    const tasks = createFolderTasks(files, baseFolderId)
    void startUploadTasks(tasks)
    return tasks
  }

  const removePendingTask = (taskId: string) => {
    const index = pendingQueue.findIndex(task => task.id === taskId)
    if (index < 0) return false
    pendingQueue.splice(index, 1)
    queuedTaskIds.delete(taskId)
    return true
  }

  const cancelTaskInternal = (task: UploadTask, preserveRetryRequest = false) => {
    if (!preserveRetryRequest) retryAfterCancelTaskIds.delete(task.id)
    canceledTaskIds.add(task.id)
    task.status = 'canceled'
    task.message = '已取消'
    removePendingTask(task.id)
    activeControllers.get(task.id)?.abort()
    maybeFinalizeCycle()
  }

  const cancelTask = (task: UploadTask) => {
    cancelTaskInternal(getManagedTask(task))
  }

  const retryTask = async (task: UploadTask) => {
    if (disposed) return
    const managedTask = getManagedTask(task)

    if (activeControllers.has(managedTask.id)) {
      retryAfterCancelTaskIds.add(managedTask.id)
      cancelTaskInternal(managedTask, true)
      const deferred = cycleDeferred
      if (deferred) await deferred.promise
      return
    }

    removePendingTask(managedTask.id)
    await startUploadTasks([managedTask])
  }

  const removeTask = (taskId: string) => {
    const task = uploadTasks.value.find(item => item.id === taskId)
    if (task && (queuedTaskIds.has(taskId) || activeControllers.has(taskId))) {
      cancelTask(task)
    }
    uploadTasks.value = uploadTasks.value.filter(item => item.id !== taskId)
    managedTasks.delete(taskId)
    canceledTaskIds.delete(taskId)
    retryAfterCancelTaskIds.delete(taskId)
  }

  const clearCompleted = () => {
    for (const task of uploadTasks.value) {
      if (task.status === 'success') managedTasks.delete(task.id)
    }
    uploadTasks.value = uploadTasks.value.filter(task => task.status !== 'success')
  }

  const clearAllTasks = () => {
    retryAfterCancelTaskIds.clear()
    for (const task of uploadTasks.value) {
      if (queuedTaskIds.has(task.id) || activeControllers.has(task.id)) {
        canceledTaskIds.add(task.id)
        task.status = 'canceled'
        task.message = '已取消'
      }
    }

    pendingQueue.splice(0, pendingQueue.length)
    queuedTaskIds.clear()
    for (const controller of activeControllers.values()) controller.abort()
    uploadTasks.value = []
    managedTasks.clear()
    canceledTaskIds.clear()
    isUploadPanelOpen.value = false
    maybeFinalizeCycle()
  }

  const toggleUploadPanel = () => {
    isUploadPanelOpen.value = !isUploadPanelOpen.value
  }

  const waitForIdle = async () => {
    while (cycleDeferred || finalizationPromise) {
      const waits: Promise<void>[] = []
      if (cycleDeferred) waits.push(cycleDeferred.promise)
      if (finalizationPromise) waits.push(finalizationPromise)
      await Promise.all(waits)
    }
  }

  const dispose = () => {
    if (disposed) return
    disposed = true
    retryAfterCancelTaskIds.clear()
    for (const task of pendingQueue) {
      task.status = 'canceled'
      task.message = '已取消'
    }
    pendingQueue.splice(0, pendingQueue.length)
    queuedTaskIds.clear()
    for (const [taskId, controller] of activeControllers) {
      canceledTaskIds.add(taskId)
      controller.abort()
    }
    folderCache.clear()
    folderLocks.clear()
    activeUploadCount.value = 0
    isUploading.value = false
    isUploadPanelOpen.value = false
    cycleDeferred?.resolve()
    cycleDeferred = null
    cycleTasks.clear()
  }

  if (getCurrentScope()) onScopeDispose(dispose)

  return {
    MAX_CONCURRENT_UPLOADS,
    uploadTasks,
    isUploadPanelOpen,
    isUploading,
    activeUploadCount,
    fileInputRef,
    folderInputRef,
    overallProgress,
    openFileDialog,
    openFolderDialog,
    createFileTasks,
    createFolderTasks,
    handleFileInput,
    handleFolderInput,
    startUploadTasks,
    cancelTask,
    retryTask,
    removeTask,
    clearCompleted,
    clearAllTasks,
    toggleUploadPanel,
    waitForIdle,
  }
}
