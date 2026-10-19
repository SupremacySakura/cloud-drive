import { computed, getCurrentScope, onScopeDispose, ref, shallowRef } from 'vue'
import {
  createPublicShareLink,
  deletePublicShareLink,
  getPublicShareLink,
  previewFileById,
} from '../services/apis/file'
import type { FileDisplayItem, PreviewKind, ToastType } from '../types/file'
import { inferPreviewMimeType, isTextPreviewable, normalizeMimeType } from '../utils/file'

const MAX_INLINE_TEXT_SIZE = 1024 * 1024

type Notify = (message: string, type?: ToastType) => void

export type UseFilePreviewOptions = {
  notify?: Notify
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

export function useFilePreview(options: UseFilePreviewOptions = {}) {
  const notify = options.notify ?? (() => undefined)

  const isPreviewModalOpen = ref(false)
  const previewLoading = ref(false)
  const previewError = ref<string | null>(null)
  const previewingFile = ref<FileDisplayItem | null>(null)
  const previewBlob = shallowRef<Blob | null>(null)
  const previewUrl = ref('')
  const previewMimeType = ref('')
  const previewTextContent = ref('')

  const publicShareLink = ref('')
  const shareError = ref<string | null>(null)
  const isLoadingShareLink = ref(false)
  const isCreatingShareLink = ref(false)
  const isDeletingShareLink = ref(false)

  let previewRequestSequence = 0
  let previewAbortController: AbortController | null = null
  let shareRequestSequence = 0
  let shareAbortController: AbortController | null = null

  const previewKind = computed<PreviewKind>(() => {
    const mime = normalizeMimeType(previewMimeType.value)
    const name = previewingFile.value?.name ?? ''
    if (!mime) return 'unsupported'
    if (isTextPreviewable(mime, name)) return 'text'
    if (mime.startsWith('image/')) return 'image'
    if (mime === 'application/pdf' || mime.includes('pdf')) return 'pdf'
    if (mime.startsWith('video/')) return 'video'
    if (mime.startsWith('audio/')) return 'audio'
    return 'unsupported'
  })

  const revokePreviewUrl = () => {
    if (!previewUrl.value) return
    URL.revokeObjectURL(previewUrl.value)
    previewUrl.value = ''
  }

  const cancelPreviewRequest = () => {
    previewRequestSequence += 1
    previewAbortController?.abort()
    previewAbortController = null
  }

  const cancelShareRequest = () => {
    shareRequestSequence += 1
    shareAbortController?.abort()
    shareAbortController = null
  }

  const beginPreviewRequest = () => {
    previewAbortController?.abort()
    const controller = new AbortController()
    previewAbortController = controller
    const sequence = ++previewRequestSequence
    return { controller, sequence }
  }

  const beginShareRequest = () => {
    shareAbortController?.abort()
    const controller = new AbortController()
    shareAbortController = controller
    const sequence = ++shareRequestSequence
    return { controller, sequence }
  }

  const isCurrentPreviewRequest = (
    sequence: number,
    controller: AbortController,
    fileId: number,
  ) => {
    return (
      sequence === previewRequestSequence &&
      previewAbortController === controller &&
      !controller.signal.aborted &&
      isPreviewModalOpen.value &&
      previewingFile.value?.type === 'file' &&
      previewingFile.value.id === fileId
    )
  }

  const isCurrentShareRequest = (sequence: number, controller: AbortController, fileId: number) => {
    return (
      sequence === shareRequestSequence &&
      shareAbortController === controller &&
      !controller.signal.aborted &&
      isPreviewModalOpen.value &&
      previewingFile.value?.type === 'file' &&
      previewingFile.value.id === fileId
    )
  }

  const resetPreviewState = () => {
    previewLoading.value = false
    previewError.value = null
    previewBlob.value = null
    previewMimeType.value = ''
    previewTextContent.value = ''
    revokePreviewUrl()
  }

  const resetShareState = () => {
    publicShareLink.value = ''
    shareError.value = null
    isLoadingShareLink.value = false
    isCreatingShareLink.value = false
    isDeletingShareLink.value = false
  }

  const loadExistingPublicShareLink = async (fileId: number) => {
    const { controller, sequence } = beginShareRequest()
    isLoadingShareLink.value = true
    isCreatingShareLink.value = false
    isDeletingShareLink.value = false
    shareError.value = null

    try {
      const data = await getPublicShareLink(fileId, controller.signal)
      if (!isCurrentShareRequest(sequence, controller, fileId)) return
      publicShareLink.value = data?.exists && data.url ? data.url : ''
    } catch (error: unknown) {
      if (isAbortError(error) || !isCurrentShareRequest(sequence, controller, fileId)) return
      shareError.value = errorMessage(error, '获取分享链接失败')
    } finally {
      if (isCurrentShareRequest(sequence, controller, fileId)) {
        isLoadingShareLink.value = false
        shareAbortController = null
      }
    }
  }

  const openPreviewModal = async (file: FileDisplayItem) => {
    if (file.type !== 'file') return

    cancelPreviewRequest()
    cancelShareRequest()
    resetPreviewState()
    resetShareState()

    isPreviewModalOpen.value = true
    previewingFile.value = file
    previewLoading.value = true

    const { controller, sequence } = beginPreviewRequest()
    void loadExistingPublicShareLink(file.id)

    try {
      const { blob, contentType, fileName } = await previewFileById(file.id, controller.signal)
      if (!isCurrentPreviewRequest(sequence, controller, file.id)) return

      const finalName = fileName || file.name
      const finalType = inferPreviewMimeType(finalName, contentType)
      let textContent = ''
      if (isTextPreviewable(finalType, finalName)) {
        textContent =
          blob.size > MAX_INLINE_TEXT_SIZE
            ? '文本文件较大，已为你展示下载入口，请下载后查看完整内容。'
            : await blob.text()
      }

      if (!isCurrentPreviewRequest(sequence, controller, file.id)) return
      previewBlob.value = blob
      previewMimeType.value = finalType
      previewTextContent.value = textContent
      previewUrl.value = URL.createObjectURL(blob)
    } catch (error: unknown) {
      if (isAbortError(error) || !isCurrentPreviewRequest(sequence, controller, file.id)) return
      previewError.value = errorMessage(error, '预览失败')
    } finally {
      if (isCurrentPreviewRequest(sequence, controller, file.id)) {
        previewLoading.value = false
        previewAbortController = null
      }
    }
  }

  const loadPreview = openPreviewModal

  const closePreviewModal = () => {
    cancelPreviewRequest()
    cancelShareRequest()
    isPreviewModalOpen.value = false
    previewingFile.value = null
    resetPreviewState()
    resetShareState()
  }

  const triggerPreviewDownload = () => {
    if (!previewBlob.value || !previewingFile.value) return
    const url = URL.createObjectURL(previewBlob.value)
    const anchor = document.createElement('a')
    try {
      anchor.href = url
      anchor.download = previewingFile.value.name
      document.body.appendChild(anchor)
      anchor.click()
    } finally {
      anchor.remove()
      URL.revokeObjectURL(url)
    }
  }

  const generatePublicShareLink = async () => {
    const file = previewingFile.value
    if (!file || file.type !== 'file') return

    const { controller, sequence } = beginShareRequest()
    isLoadingShareLink.value = false
    isCreatingShareLink.value = true
    isDeletingShareLink.value = false
    shareError.value = null

    try {
      const { url } = await createPublicShareLink(file.id, controller.signal)
      if (!isCurrentShareRequest(sequence, controller, file.id)) return
      publicShareLink.value = url
      notify('公网链接已生成', 'success')
    } catch (error: unknown) {
      if (isAbortError(error) || !isCurrentShareRequest(sequence, controller, file.id)) return
      const message = errorMessage(error, '生成分享链接失败')
      shareError.value = message
      notify(message, 'error')
    } finally {
      if (isCurrentShareRequest(sequence, controller, file.id)) {
        isCreatingShareLink.value = false
        shareAbortController = null
      }
    }
  }

  const removePublicShareLink = async () => {
    const file = previewingFile.value
    if (!file || file.type !== 'file') return

    const previousLink = publicShareLink.value
    const { controller, sequence } = beginShareRequest()
    isLoadingShareLink.value = false
    isCreatingShareLink.value = false
    isDeletingShareLink.value = true
    shareError.value = null

    try {
      await deletePublicShareLink(file.id, controller.signal)
      if (!isCurrentShareRequest(sequence, controller, file.id)) return
      publicShareLink.value = ''
      notify('公网链接已删除', 'success')
    } catch (error: unknown) {
      if (isAbortError(error) || !isCurrentShareRequest(sequence, controller, file.id)) return
      const message = errorMessage(error, '删除分享链接失败')
      publicShareLink.value = previousLink
      shareError.value = message
      notify(message, 'error')
    } finally {
      if (isCurrentShareRequest(sequence, controller, file.id)) {
        isDeletingShareLink.value = false
        shareAbortController = null
      }
    }
  }

  const copyPublicShareLink = async () => {
    const link = publicShareLink.value
    if (!link) return
    try {
      await navigator.clipboard.writeText(link)
      notify('链接已复制', 'success')
    } catch {
      notify('复制失败，请手动复制', 'error')
    }
  }

  const dispose = () => {
    cancelPreviewRequest()
    cancelShareRequest()
    revokePreviewUrl()
  }

  if (getCurrentScope()) onScopeDispose(dispose)

  return {
    isPreviewModalOpen,
    previewLoading,
    previewError,
    previewingFile,
    previewBlob,
    previewUrl,
    previewMimeType,
    previewTextContent,
    previewKind,
    publicShareLink,
    shareError,
    isLoadingShareLink,
    isCreatingShareLink,
    isDeletingShareLink,
    loadPreview,
    openPreviewModal,
    closePreviewModal,
    revokePreviewUrl,
    triggerPreviewDownload,
    loadExistingPublicShareLink,
    generatePublicShareLink,
    removePublicShareLink,
    copyPublicShareLink,
  }
}
