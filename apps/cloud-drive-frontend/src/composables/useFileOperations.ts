import { computed, getCurrentScope, onScopeDispose, ref, type Ref } from 'vue'
import {
  deleteFile,
  downloadById,
  getListByFolderIDAndUserID,
  makeDirectory,
  moveFile,
  renameFile,
} from '../services/apis/file'
import type { FileListItem } from '../services/types/file'
import type { BreadcrumbItem, FileDisplayItem, ToastType } from '../types/file'
import { toFileTarget } from '../utils/file-management'

type Notify = (message: string, type?: ToastType) => void

export type UseFileOperationsOptions = {
  currentFolderId: Readonly<Ref<number>>
  breadcrumbs: Ref<BreadcrumbItem[]>
  refresh: (folderId: number) => Promise<unknown> | unknown
  notify?: Notify
}

type RequestSlot = {
  sequence: number
  controller: AbortController | null
}

const createRequestSlot = (): RequestSlot => ({ sequence: 0, controller: null })

const beginRequest = (slot: RequestSlot) => {
  slot.controller?.abort()
  const controller = new AbortController()
  slot.controller = controller
  const sequence = ++slot.sequence
  return { controller, sequence }
}

const cancelRequest = (slot: RequestSlot) => {
  slot.sequence += 1
  slot.controller?.abort()
  slot.controller = null
}

const isCurrentRequest = (slot: RequestSlot, sequence: number, controller: AbortController) => {
  return slot.sequence === sequence && slot.controller === controller && !controller.signal.aborted
}

const isAbortError = (error: unknown) => {
  if (!(error instanceof Error)) return false
  return (
    error.name === 'AbortError' ||
    error.name === 'CanceledError' ||
    (error as Error & { code?: string }).code === 'ERR_CANCELED'
  )
}

const errorMessage = (error: unknown, fallback: string) => {
  return error instanceof Error && error.message ? error.message : fallback
}

export function useFileOperations(options: UseFileOperationsOptions) {
  const notify = options.notify ?? (() => undefined)

  const isCreateFolderModalOpen = ref(false)
  const newFolderName = ref('')
  const isCreatingFolder = ref(false)
  const createFolderError = ref<string | null>(null)

  const isRenameModalOpen = ref(false)
  const renameName = ref('')
  const isRenaming = ref(false)
  const renamingTarget = ref<FileDisplayItem | null>(null)
  const renameError = ref<string | null>(null)

  const isMoveModalOpen = ref(false)
  const isMoving = ref(false)
  const movingTarget = ref<FileDisplayItem | null>(null)
  const moveTargetFolderId = ref(0)
  const moveBrowserFolderId = ref(0)
  const moveBrowserBreadcrumbs = ref<BreadcrumbItem[]>([{ id: 0, name: 'root' }])
  const moveBrowserFolders = ref<FileListItem[]>([])
  const isMoveBrowserLoading = ref(false)
  const moveError = ref<string | null>(null)
  const moveBrowserError = ref<string | null>(null)

  const isDeleteConfirmModalOpen = ref(false)
  const deleteConfirmTarget = ref<FileDisplayItem | null>(null)
  const deletingMenuTargetId = ref<string | null>(null)
  const deleteError = ref<string | null>(null)

  const downloadingMenuTargetId = ref<string | null>(null)
  const downloadError = ref<string | null>(null)

  const operationError = computed(
    () =>
      createFolderError.value ??
      renameError.value ??
      moveError.value ??
      moveBrowserError.value ??
      deleteError.value ??
      downloadError.value,
  )

  const createFolderRequest = createRequestSlot()
  const renameRequest = createRequestSlot()
  const moveRequest = createRequestSlot()
  const moveBrowserRequest = createRequestSlot()
  const deleteRequest = createRequestSlot()
  const downloadRequest = createRequestSlot()

  const refreshCurrentFolder = async () => {
    try {
      await options.refresh(options.currentFolderId.value)
    } catch (error: unknown) {
      notify(errorMessage(error, '刷新文件列表失败'), 'error')
    }
  }

  const resetCreateFolderModal = () => {
    isCreateFolderModalOpen.value = false
    newFolderName.value = ''
    createFolderError.value = null
  }

  const openCreateFolderModal = () => {
    cancelRequest(createFolderRequest)
    isCreatingFolder.value = false
    createFolderError.value = null
    newFolderName.value = ''
    isCreateFolderModalOpen.value = true
  }

  const closeCreateFolderModal = () => {
    cancelRequest(createFolderRequest)
    isCreatingFolder.value = false
    resetCreateFolderModal()
  }

  const handleCreateFolder = async () => {
    const name = newFolderName.value.trim()
    if (!name) {
      createFolderError.value = '文件夹名称不能为空'
      notify(createFolderError.value, 'error')
      return false
    }

    const parentFolderId = options.currentFolderId.value
    const { controller, sequence } = beginRequest(createFolderRequest)
    isCreatingFolder.value = true
    createFolderError.value = null
    let succeeded = false

    try {
      const folderId = await makeDirectory({ folder_id: parentFolderId, name }, controller.signal)
      if (!isCurrentRequest(createFolderRequest, sequence, controller)) return false
      if (!(folderId > 0)) throw new Error('创建文件夹失败')

      succeeded = true
      resetCreateFolderModal()
      notify('文件夹创建成功', 'success')
    } catch (error: unknown) {
      if (isAbortError(error) || !isCurrentRequest(createFolderRequest, sequence, controller)) {
        return false
      }
      const message = errorMessage(error, '创建文件夹失败')
      createFolderError.value = message
      notify(message, 'error')
    } finally {
      if (isCurrentRequest(createFolderRequest, sequence, controller)) {
        isCreatingFolder.value = false
        createFolderRequest.controller = null
      }
    }

    if (succeeded) await refreshCurrentFolder()
    return succeeded
  }

  const resetRenameModal = () => {
    isRenameModalOpen.value = false
    renameName.value = ''
    renamingTarget.value = null
    renameError.value = null
  }

  const openRenameModal = (target?: FileDisplayItem | null) => {
    cancelRequest(renameRequest)
    isRenaming.value = false
    if (target) renamingTarget.value = target
    if (!renamingTarget.value) return
    renameError.value = null
    renameName.value = renamingTarget.value.name
    isRenameModalOpen.value = true
  }

  const closeRenameModal = () => {
    cancelRequest(renameRequest)
    isRenaming.value = false
    resetRenameModal()
  }

  const handleRename = async (target?: FileDisplayItem, nextName?: string) => {
    const item = target ?? renamingTarget.value
    const name = (nextName ?? renameName.value).trim()
    if (!item) return false
    if (!name) {
      renameError.value = '名称不能为空'
      notify(renameError.value, 'error')
      return false
    }
    if (name === item.name) {
      closeRenameModal()
      return true
    }

    const { controller, sequence } = beginRequest(renameRequest)
    isRenaming.value = true
    renameError.value = null
    let succeeded = false

    try {
      await renameFile({ ...toFileTarget(item), name }, controller.signal)
      if (!isCurrentRequest(renameRequest, sequence, controller)) return false

      if (item.type === 'folder') {
        options.breadcrumbs.value = options.breadcrumbs.value.map(breadcrumb =>
          breadcrumb.id === item.id ? { ...breadcrumb, name } : breadcrumb,
        )
      }
      succeeded = true
      resetRenameModal()
      notify('重命名成功', 'success')
    } catch (error: unknown) {
      if (isAbortError(error) || !isCurrentRequest(renameRequest, sequence, controller))
        return false
      const message = errorMessage(error, '重命名失败')
      renameError.value = message
      notify(message, 'error')
    } finally {
      if (isCurrentRequest(renameRequest, sequence, controller)) {
        isRenaming.value = false
        renameRequest.controller = null
      }
    }

    if (succeeded) await refreshCurrentFolder()
    return succeeded
  }

  const loadMoveBrowserFolders = async (folderId: number) => {
    const { controller, sequence } = beginRequest(moveBrowserRequest)
    isMoveBrowserLoading.value = true
    moveBrowserError.value = null

    try {
      const list = await getListByFolderIDAndUserID(folderId, 1, 100, controller.signal)
      if (!isCurrentRequest(moveBrowserRequest, sequence, controller)) return
      moveBrowserFolders.value = Array.isArray(list)
        ? list.filter(item => item.type === 'folder')
        : []
    } catch (error: unknown) {
      if (isAbortError(error) || !isCurrentRequest(moveBrowserRequest, sequence, controller)) return
      const message = errorMessage(error, '加载目录失败')
      moveBrowserFolders.value = []
      moveBrowserError.value = message
      notify(message, 'error')
    } finally {
      if (isCurrentRequest(moveBrowserRequest, sequence, controller)) {
        isMoveBrowserLoading.value = false
        moveBrowserRequest.controller = null
      }
    }
  }

  const resetMoveModal = () => {
    isMoveModalOpen.value = false
    movingTarget.value = null
    moveTargetFolderId.value = 0
    moveBrowserFolderId.value = 0
    moveBrowserBreadcrumbs.value = [{ id: 0, name: 'root' }]
    moveBrowserFolders.value = []
    moveError.value = null
    moveBrowserError.value = null
  }

  const openMoveModal = async (target?: FileDisplayItem | null) => {
    cancelRequest(moveRequest)
    cancelRequest(moveBrowserRequest)
    isMoving.value = false
    isMoveBrowserLoading.value = false
    if (target) movingTarget.value = target
    if (!movingTarget.value) return

    moveError.value = null
    moveBrowserError.value = null
    moveTargetFolderId.value = options.currentFolderId.value
    moveBrowserFolderId.value = options.currentFolderId.value
    moveBrowserBreadcrumbs.value = [...options.breadcrumbs.value]
    isMoveModalOpen.value = true
    await loadMoveBrowserFolders(moveBrowserFolderId.value)
  }

  const closeMoveModal = () => {
    cancelRequest(moveRequest)
    cancelRequest(moveBrowserRequest)
    isMoving.value = false
    isMoveBrowserLoading.value = false
    resetMoveModal()
  }

  const goToMoveBrowserFolder = async (folder: FileListItem) => {
    if (folder.type !== 'folder') return
    moveBrowserFolderId.value = folder.id
    moveTargetFolderId.value = folder.id
    moveBrowserBreadcrumbs.value = [
      ...moveBrowserBreadcrumbs.value,
      { id: folder.id, name: folder.name },
    ]
    await loadMoveBrowserFolders(folder.id)
  }

  const goToMoveBrowserBreadcrumb = async (index: number) => {
    const next = moveBrowserBreadcrumbs.value[index]
    if (!next) return
    moveBrowserBreadcrumbs.value = moveBrowserBreadcrumbs.value.slice(0, index + 1)
    moveBrowserFolderId.value = next.id
    moveTargetFolderId.value = next.id
    await loadMoveBrowserFolders(next.id)
  }

  const selectMoveTargetCurrent = () => {
    moveTargetFolderId.value = moveBrowserFolderId.value
  }

  const handleMove = async (target?: FileDisplayItem, targetFolderId?: number) => {
    const item = target ?? movingTarget.value
    const destination = targetFolderId ?? moveTargetFolderId.value
    if (!item) return false
    if (item.type === 'folder' && destination === item.id) {
      moveError.value = '不能移动到自身目录'
      notify(moveError.value, 'error')
      return false
    }

    const { controller, sequence } = beginRequest(moveRequest)
    isMoving.value = true
    moveError.value = null
    let succeeded = false

    try {
      await moveFile({ ...toFileTarget(item), target_folder_id: destination }, controller.signal)
      if (!isCurrentRequest(moveRequest, sequence, controller)) return false

      succeeded = true
      cancelRequest(moveBrowserRequest)
      isMoveBrowserLoading.value = false
      resetMoveModal()
      notify('移动成功', 'success')
    } catch (error: unknown) {
      if (isAbortError(error) || !isCurrentRequest(moveRequest, sequence, controller)) return false
      const message = errorMessage(error, '移动失败')
      moveError.value = message
      notify(message, 'error')
    } finally {
      if (isCurrentRequest(moveRequest, sequence, controller)) {
        isMoving.value = false
        moveRequest.controller = null
      }
    }

    if (succeeded) await refreshCurrentFolder()
    return succeeded
  }

  const resetDeleteConfirmModal = () => {
    isDeleteConfirmModalOpen.value = false
    deleteConfirmTarget.value = null
    deleteError.value = null
  }

  const handleDeleteFromMenu = (target?: FileDisplayItem | null) => {
    cancelRequest(deleteRequest)
    deletingMenuTargetId.value = null
    if (target) deleteConfirmTarget.value = target
    if (!deleteConfirmTarget.value) return
    deleteError.value = null
    isDeleteConfirmModalOpen.value = true
  }

  const closeDeleteConfirmModal = (force = false) => {
    if (deletingMenuTargetId.value && !force) return
    cancelRequest(deleteRequest)
    deletingMenuTargetId.value = null
    resetDeleteConfirmModal()
  }

  const confirmDeleteFromModal = async () => {
    const item = deleteConfirmTarget.value
    if (!item) return false

    const { controller, sequence } = beginRequest(deleteRequest)
    deletingMenuTargetId.value = `${item.type}-${item.id}`
    deleteError.value = null
    let succeeded = false

    try {
      await deleteFile(toFileTarget(item), controller.signal)
      if (!isCurrentRequest(deleteRequest, sequence, controller)) return false

      succeeded = true
      resetDeleteConfirmModal()
      notify('删除成功', 'success')
    } catch (error: unknown) {
      if (isAbortError(error) || !isCurrentRequest(deleteRequest, sequence, controller))
        return false
      const message = errorMessage(error, '删除失败')
      deleteError.value = message
      notify(message, 'error')
    } finally {
      if (isCurrentRequest(deleteRequest, sequence, controller)) {
        deletingMenuTargetId.value = null
        deleteRequest.controller = null
      }
    }

    if (succeeded) await refreshCurrentFolder()
    return succeeded
  }

  const handleDeleteFromModal = confirmDeleteFromModal

  const handleDownloadFromMenu = async (target?: FileDisplayItem | null) => {
    if (!target) return false

    const { controller, sequence } = beginRequest(downloadRequest)
    downloadingMenuTargetId.value = `${target.type}-${target.id}`
    downloadError.value = null

    try {
      const targetIds = toFileTarget(target)
      await downloadById(targetIds.file_id, targetIds.folder_id, undefined, controller.signal)
      if (!isCurrentRequest(downloadRequest, sequence, controller)) return false
      notify(`开始下载：${target.name}`, 'success')
      return true
    } catch (error: unknown) {
      if (isAbortError(error) || !isCurrentRequest(downloadRequest, sequence, controller))
        return false
      const message = errorMessage(error, '下载失败')
      downloadError.value = message
      notify(message, 'error')
      return false
    } finally {
      if (isCurrentRequest(downloadRequest, sequence, controller)) {
        downloadingMenuTargetId.value = null
        downloadRequest.controller = null
      }
    }
  }

  const dispose = () => {
    cancelRequest(createFolderRequest)
    cancelRequest(renameRequest)
    cancelRequest(moveRequest)
    cancelRequest(moveBrowserRequest)
    cancelRequest(deleteRequest)
    cancelRequest(downloadRequest)
    isCreatingFolder.value = false
    isRenaming.value = false
    isMoving.value = false
    isMoveBrowserLoading.value = false
    deletingMenuTargetId.value = null
    downloadingMenuTargetId.value = null
  }

  if (getCurrentScope()) onScopeDispose(dispose)

  return {
    operationError,
    isCreateFolderModalOpen,
    newFolderName,
    isCreatingFolder,
    createFolderError,
    openCreateFolderModal,
    closeCreateFolderModal,
    handleCreateFolder,
    isRenameModalOpen,
    renameName,
    isRenaming,
    renamingTarget,
    renameError,
    openRenameModal,
    closeRenameModal,
    handleRename,
    isMoveModalOpen,
    isMoving,
    movingTarget,
    moveTargetFolderId,
    moveBrowserFolderId,
    moveBrowserBreadcrumbs,
    moveBrowserFolders,
    isMoveBrowserLoading,
    moveError,
    moveBrowserError,
    openMoveModal,
    closeMoveModal,
    loadMoveBrowserFolders,
    goToMoveBrowserFolder,
    goToMoveBrowserBreadcrumb,
    selectMoveTargetCurrent,
    handleMove,
    isDeleteConfirmModalOpen,
    deleteConfirmTarget,
    deletingMenuTargetId,
    deleteError,
    handleDeleteFromMenu,
    closeDeleteConfirmModal,
    confirmDeleteFromModal,
    handleDeleteFromModal,
    downloadingMenuTargetId,
    downloadError,
    handleDownloadFromMenu,
  }
}
