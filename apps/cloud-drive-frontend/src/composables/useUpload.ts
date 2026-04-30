import { ref, computed } from 'vue'

type UploadTaskStatus =
  | 'pending'
  | 'hashing'
  | 'uploading'
  | 'merging'
  | 'success'
  | 'failed'
  | 'canceled'
type UploadTask = {
  id: string
  file: File
  targetFolderId: number
  relativePath: string
  status: UploadTaskStatus
  percent: number
  message: string | null
  canceled: boolean
}

export function useUpload() {
  const MAX_CONCURRENT_UPLOADS = 3
  const uploadTasks = ref<UploadTask[]>([])
  const isUploadPanelOpen = ref(false)
  const isUploading = ref(false)
  const fileInputRef = ref<HTMLInputElement | null>(null)
  const folderInputRef = ref<HTMLInputElement | null>(null)

  const overallProgress = computed(() => {
    if (uploadTasks.value.length === 0) return 0
    const total = uploadTasks.value.reduce((sum, t) => sum + t.percent, 0)
    return Math.floor(total / uploadTasks.value.length)
  })

  // 公开的入口：打开选择文件
  const openFileDialog = () => fileInputRef.value?.click()
  const openFolderDialog = () => folderInputRef.value?.click()

  // 上传任务入口：简单实现，保持 API 与原实现兼容性
  const startUploadTasks = async (tasks: UploadTask[]) => {
    uploadTasks.value.unshift(...tasks)
    isUploadPanelOpen.value = true
    isUploading.value = true
    // 本简化版本仅执行任务的基础状态更新，实际分片/上传逻辑在真实环境中实现
    for (const t of tasks) {
      t.status = 'pending'
      t.percent = 0
      t.message = '等待中'
    }
    // 模拟上传完成
    for (const t of tasks) {
      if (t.canceled) continue
      t.status = 'success'
      t.percent = 100
      t.message = '上传完成'
    }
    isUploading.value = false
    await Promise.resolve()
  }

  // 取消、重试、删除等操作在后续实现中完善
  const cancelTask = (task: UploadTask) => {
    task.canceled = true
    task.status = 'canceled'
  }
  const retryTask = async (task: UploadTask) => {
    task.status = 'pending'
    task.percent = 0
    task.message = '等待中'
    await Promise.resolve()
  }
  const removeTask = (taskId: string) => {
    uploadTasks.value = uploadTasks.value.filter(t => t.id !== taskId)
  }
  const clearCompleted = () => {
    uploadTasks.value = uploadTasks.value.filter(t => t.status !== 'success')
  }
  const clearAllTasks = () => {
    uploadTasks.value = []
  }

  return {
    MAX_CONCURRENT_UPLOADS,
    uploadTasks,
    isUploadPanelOpen,
    isUploading,
    fileInputRef,
    folderInputRef,
    overallProgress,
    openFileDialog,
    openFolderDialog,
    startUploadTasks,
    cancelTask,
    retryTask,
    removeTask,
    clearCompleted,
    clearAllTasks,
  }
}
