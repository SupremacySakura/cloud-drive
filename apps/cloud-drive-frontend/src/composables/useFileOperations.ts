import { ref } from 'vue'
import type { FileListItem } from '../services/types/file'
import { renameFile, moveFile, deleteFile } from '../services/apis/file'

type DisplayItem = FileListItem & {
  icon: string
  iconBg: string
  iconFg: string
  typeLabel: string
  lastModifiedText: string
}

export function useFileOperations() {
  const isDeleteConfirmModalOpen = ref(false)
  const deleteConfirmTarget = ref<DisplayItem | null>(null)
  const isRenaming = ref(false)
  const isMoving = ref(false)
  const renamingTarget = ref<DisplayItem | null>(null)
  const isRenameModalOpen = ref(false)
  const renameName = ref('')
  const isMoveModalOpen = ref(false)
  const moveTargetFolderId = ref(0)
  const movingTarget = ref<DisplayItem | null>(null)

  const handleRename = async (target?: DisplayItem, nextName?: string) => {
    const t = target ?? renamingTarget.value
    const name = (nextName ?? renameName.value).trim()
    if (!t || !name) return
    isRenaming.value = true
    try {
      await renameFile({
        file_id: t.type === 'file' ? t.id : 0,
        folder_id: t.type === 'folder' ? t.id : 0,
        name,
      })
      closeRenameModal()
    } finally {
      isRenaming.value = false
    }
  }

  const handleMove = async (target?: DisplayItem, targetFolderId?: number) => {
    const t = target ?? movingTarget.value
    const dest = targetFolderId ?? moveTargetFolderId.value
    if (!t) return
    isMoving.value = true
    try {
      await moveFile({
        file_id: t.type === 'file' ? t.id : 0,
        folder_id: t.type === 'folder' ? t.id : 0,
        target_folder_id: dest,
      })
      closeMoveModal()
    } finally {
      isMoving.value = false
    }
  }

  const handleDeleteFromModal = async () => {
    if (!deleteConfirmTarget.value) return
    const t = deleteConfirmTarget.value
    try {
      await deleteFile({
        file_id: t.type === 'file' ? t.id : 0,
        folder_id: t.type === 'folder' ? t.id : 0,
      })
    } finally {
      isDeleteConfirmModalOpen.value = false
      deleteConfirmTarget.value = null
    }
  }

  const openRenameModal = () => {
    if (!renamingTarget.value) return
    renameName.value = renamingTarget.value.name
    isRenameModalOpen.value = true
  }
  const closeRenameModal = () => {
    isRenameModalOpen.value = false
    renameName.value = ''
    renamingTarget.value = null
  }
  const closeMoveModal = () => {
    isMoveModalOpen.value = false
    movingTarget.value = null
    moveTargetFolderId.value = 0
  }

  const openMoveModal = () => {
    if (!movingTarget.value) return
    isMoveModalOpen.value = true
  }

  return {
    isDeleteConfirmModalOpen,
    deleteConfirmTarget,
    isRenaming,
    isMoving,
    renamingTarget,
    isRenameModalOpen,
    renameName,
    isMoveModalOpen,
    moveTargetFolderId,
    movingTarget,
    openRenameModal,
    closeRenameModal,
    handleRename,
    openMoveModal,
    closeMoveModal,
    handleMove,
    handleDeleteFromModal,
  }
}
