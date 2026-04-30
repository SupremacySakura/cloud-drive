import { ref } from 'vue'
import type { FileListItem } from '../services/types/file'
import {
  previewFileById,
  getPublicShareLink,
  deletePublicShareLink,
  createPublicShareLink,
} from '../services/apis/file'

type DisplayItem = FileListItem & {
  icon: string
  iconBg: string
  iconFg: string
  typeLabel: string
  lastModifiedText: string
}

export function useFilePreview() {
  const isPreviewModalOpen = ref(false)
  const previewLoading = ref(false)
  const previewError = ref<string | null>(null)
  const previewingFile = ref<DisplayItem | null>(null)
  const previewBlob = ref<Blob | null>(null)
  const previewUrl = ref<string>('')
  const previewMimeType = ref<string>('')
  const previewTextContent = ref<string>('')
  const publicShareLink = ref<string>('')
  const shareError = ref<string | null>(null)
  const isCreatingShareLink = ref(false)
  const isDeletingShareLink = ref(false)

  const loadPreview = async (file: DisplayItem) => {
    if (file.type !== 'file') return
    isPreviewModalOpen.value = true
    previewLoading.value = true
    previewError.value = null
    previewingFile.value = file
    previewBlob.value = null
    previewMimeType.value = ''
    previewTextContent.value = ''
    publicShareLink.value = ''
    shareError.value = null
    isCreatingShareLink.value = false
    isDeletingShareLink.value = false
    try {
      const { blob, contentType, fileName } = await previewFileById(file.id)
      const finalName = fileName || file.name
      const finalType = (contentType || '').split(';')[0].toLowerCase() || ''
      previewBlob.value = blob
      previewMimeType.value = finalType
      if (
        finalType.startsWith('text/') ||
        [
          'md',
          'json',
          'csv',
          'log',
          'xml',
          'yaml',
          'yml',
          'js',
          'ts',
          'vue',
          'html',
          'css',
        ].includes(finalName.split('.').pop() ?? '')
      ) {
        if (blob.size <= 1024 * 1024) previewTextContent.value = await blob.text()
        else previewTextContent.value = '文本文件较大，已为你展示下载入口，请下载后查看完整内容。'
      }
      previewUrl.value = URL.createObjectURL(blob)
    } catch (e: any) {
      previewError.value = e?.message || '预览失败'
    } finally {
      previewLoading.value = false
    }
  }

  const openPreviewModal = (file: DisplayItem) => loadPreview(file)
  const closePreviewModal = () => {
    isPreviewModalOpen.value = false
    previewLoading.value = false
    previewError.value = null
    previewingFile.value = null
    previewBlob.value = null
    if (previewUrl.value) URL.revokeObjectURL(previewUrl.value)
    previewUrl.value = ''
    previewMimeType.value = ''
    previewTextContent.value = ''
    publicShareLink.value = ''
    shareError.value = null
  }
  const revokePreviewUrl = () => {
    if (previewUrl.value) URL.revokeObjectURL(previewUrl.value)
  }

  const generatePublicShareLink = async () => {
    if (!previewingFile.value?.id) return
    isCreatingShareLink.value = true
    shareError.value = null
    try {
      const { token } = await createPublicShareLink(previewingFile.value.id)
      publicShareLink.value = `${window.location.origin}/api/file/share/open?token=${encodeURIComponent(token)}`
    } catch (e: any) {
      shareError.value = e?.message || '生成分享链接失败'
    } finally {
      isCreatingShareLink.value = false
    }
  }
  const loadExistingPublicShareLink = async (fileId: number) => {
    try {
      const data = await getPublicShareLink(fileId)
      publicShareLink.value = data?.exists && data?.url ? data.url : ''
    } catch (e: any) {
      shareError.value = e?.message || '获取分享链接失败'
    }
  }
  const removePublicShareLink = async () => {
    if (!previewingFile.value?.id) return
    isDeletingShareLink.value = true
    shareError.value = null
    try {
      await deletePublicShareLink(previewingFile.value.id)
      publicShareLink.value = ''
    } catch (e: any) {
      shareError.value = e?.message || '删除分享链接失败'
    } finally {
      isDeletingShareLink.value = false
    }
  }
  const copyPublicShareLink = async () => {
    if (!publicShareLink.value) return
    try {
      await navigator.clipboard.writeText(publicShareLink.value)
    } catch {
      // ignore
    }
  }

  return {
    isPreviewModalOpen,
    previewLoading,
    previewError,
    previewingFile,
    previewBlob,
    previewUrl,
    previewMimeType,
    previewTextContent,
    publicShareLink,
    shareError,
    isCreatingShareLink,
    isDeletingShareLink,
    loadPreview,
    openPreviewModal,
    closePreviewModal,
    revokePreviewUrl,
    loadExistingPublicShareLink,
    generatePublicShareLink,
    removePublicShareLink,
    copyPublicShareLink,
  }
}
