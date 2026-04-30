import { ref, computed } from 'vue'
import type { FileListItem } from '../services/types/file'
import { getListByFolderIDAndUserID, getListCountByFolderIDAndUserID } from '../services/apis/file'
import { useFileStore } from '../stores/file'
import { iconForListItem, typeLabelForListItem, sanitizeFileName, formatTime } from '../utils/file'

type ViewMode = 'list' | 'grid'

type DisplayItem = FileListItem & {
  icon: string
  iconBg: string
  iconFg: string
  typeLabel: string
  lastModifiedText: string
}

export function useFileList() {
  const _fs = useFileStore()
  void _fs
  const viewMode = ref<ViewMode>('list')
  const currentFolderId = ref<number>(0)
  const breadcrumbs = ref<{ id: number; name: string }[]>([{ id: 0, name: 'root' }])

  const page = ref<number>(1)
  const pageSize = ref<number>(10)
  const totalCount = ref<number>(0)
  const isLoading = ref(false)
  const errorMessage = ref<string | null>(null)

  const rawItems = ref<FileListItem[]>([])

  const listAbortController = ref<AbortController | null>(null)

  const sortedFiles = computed<DisplayItem[]>(() => {
    const mapped = rawItems.value.map(item => {
      const meta = iconForListItem(item)
      return {
        ...item,
        name: sanitizeFileName(item.name),
        icon: meta.icon,
        iconBg: meta.bg,
        iconFg: meta.fg,
        typeLabel: typeLabelForListItem(item),
        lastModifiedText: formatTime(item.updated_at),
      } as DisplayItem
    })
    // 默认排序：名称升序，如需外部设定排序请在外部覆盖
    mapped.sort((a, b) => a.name.localeCompare(b.name))
    return mapped
  })

  const totalPages = computed(() => Math.max(1, Math.ceil(totalCount.value / pageSize.value)))
  const startIndex = computed(() =>
    totalCount.value === 0 ? 0 : (page.value - 1) * pageSize.value + 1,
  )
  const endIndex = computed(() => Math.min(page.value * pageSize.value, totalCount.value))
  const pageNumbers = computed<number[]>(() => {
    const total = totalPages.value
    if (total <= 3) return Array.from({ length: total }, (_, i) => i + 1)
    if (page.value <= 1) return [1, 2, 3]
    if (page.value >= total) return [total - 2, total - 1, total]
    return [page.value - 1, page.value, page.value + 1]
  })

  const fetchFolder = async (folderId: number) => {
    if (listAbortController.value) {
      listAbortController.value.abort()
    }
    listAbortController.value = new AbortController()
    const signal = listAbortController.value.signal

    isLoading.value = true
    errorMessage.value = null
    currentFolderId.value = folderId
    try {
      const count = await getListCountByFolderIDAndUserID(folderId, signal)
      totalCount.value = Number.isFinite(count) ? count : 0
      const list = await getListByFolderIDAndUserID(folderId, page.value, pageSize.value, signal)
      rawItems.value = Array.isArray(list) ? list : []
    } catch (e: any) {
      if (e.name === 'CanceledError' || e.name === 'AbortError') {
        return
      }
      rawItems.value = []
      totalCount.value = 0
      errorMessage.value = e?.message || '加载失败'
    } finally {
      isLoading.value = false
    }
  }

  const goToFolder = async (folderId: number, folderName: string) => {
    currentFolderId.value = folderId
    breadcrumbs.value = [...breadcrumbs.value, { id: folderId, name: folderName }]
    page.value = 1
    await fetchFolder(folderId)
  }

  const goToBreadcrumb = async (index: number) => {
    const bc = breadcrumbs.value[index]
    if (!bc) return
    breadcrumbs.value = breadcrumbs.value.slice(0, index + 1)
    currentFolderId.value = bc.id
    page.value = 1
    await fetchFolder(bc.id)
  }

  const goToParentFolder = async () => {
    if (breadcrumbs.value.length > 1) {
      await goToBreadcrumb(breadcrumbs.value.length - 2)
    }
  }

  const onRowClick = async (item: FileListItem) => {
    if (item.type !== 'folder') return
    await goToFolder(item.id, item.name)
  }

  const setSort = () => {
    // 简单占位，复杂排序留给上层实现
  }

  const goToPage = async (nextPage: number) => {
    const total = totalPages.value
    const clamped = Math.min(Math.max(1, nextPage), total)
    if (clamped === page.value) return
    page.value = clamped
    await fetchFolder(currentFolderId.value)
  }

  // 选择相关的简单实现，供外层模板绑定
  const selectedIds = ref<Set<number>>(new Set())
  const allSelected = computed(() => selectedIds.value.size === sortedFiles.value.length)
  const toggleAll = (checked: boolean) => {
    if (!checked) {
      selectedIds.value = new Set()
      return
    }
    selectedIds.value = new Set(sortedFiles.value.map(f => f.id))
  }
  const toggleOne = (id: string, checked: boolean) => {
    const next = new Set(selectedIds.value)
    const numeric = Number(id)
    if (!Number.isFinite(numeric)) return
    if (checked) next.add(numeric)
    else next.delete(numeric)
    selectedIds.value = next
  }

  const cancelListRequest = () => {
    if (listAbortController.value) {
      listAbortController.value.abort()
      listAbortController.value = null
    }
  }

  return {
    viewMode,
    currentFolderId,
    breadcrumbs,
    page,
    pageSize,
    totalCount,
    isLoading,
    errorMessage,
    rawItems,
    sortedFiles,
    totalPages,
    startIndex,
    endIndex,
    pageNumbers,
    fetchFolder,
    goToFolder,
    goToBreadcrumb,
    goToParentFolder,
    onRowClick,
    setSort,
    goToPage,
    selectedIds,
    allSelected,
    toggleAll,
    toggleOne,
    cancelListRequest,
  }
}
