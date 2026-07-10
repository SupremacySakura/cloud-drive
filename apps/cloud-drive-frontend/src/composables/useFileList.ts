import { computed, onScopeDispose, ref } from 'vue'
import { getListByFolderIDAndUserID, getListCountByFolderIDAndUserID } from '../services/apis/file'
import type { FileListItem } from '../services/types/file'
import type {
  BreadcrumbItem,
  FileFilterKey,
  FileItemKey,
  FileSortKey,
  FileViewMode,
  SortDirection,
} from '../types/file'
import { buildVisibleFileItems, fileItemKey } from '../utils/file-management'

export const fileFilterOptions: Array<{ key: FileFilterKey; label: string }> = [
  { key: 'all', label: '全部类型' },
  { key: 'folder', label: '文件夹' },
  { key: 'image', label: '图片' },
  { key: 'video', label: '视频' },
  { key: 'audio', label: '音频' },
  { key: 'document', label: '文档' },
  { key: 'other', label: '其他' },
]

const isAbortError = (error: unknown) => {
  return error instanceof Error && (error.name === 'CanceledError' || error.name === 'AbortError')
}

export function useFileList() {
  const viewMode = ref<FileViewMode>('list')
  const sortKey = ref<FileSortKey>('name')
  const sortDirection = ref<SortDirection>('asc')
  const activeFilter = ref<FileFilterKey>('all')
  const searchQuery = ref('')
  const debouncedSearchQuery = ref('')

  const currentFolderId = ref(0)
  const breadcrumbs = ref<BreadcrumbItem[]>([{ id: 0, name: 'root' }])
  const page = ref(1)
  const pageSize = ref(10)
  const totalCount = ref(0)
  const rawItems = ref<FileListItem[]>([])
  const isLoading = ref(false)
  const errorMessage = ref<string | null>(null)
  const selectedIds = ref<Set<FileItemKey>>(new Set())

  let listAbortController: AbortController | null = null
  let searchDebounceTimer: ReturnType<typeof setTimeout> | null = null
  let requestSequence = 0

  const sortedFiles = computed(() =>
    buildVisibleFileItems(rawItems.value, {
      query: debouncedSearchQuery.value,
      filter: activeFilter.value,
      sortKey: sortKey.value,
      sortDirection: sortDirection.value,
    }),
  )

  const activeFilterLabel = computed(
    () => fileFilterOptions.find(option => option.key === activeFilter.value)?.label ?? '全部类型',
  )
  const sortLabel = computed(() => {
    if (sortKey.value === 'name') return 'Name'
    if (sortKey.value === 'size') return 'Size'
    return 'Last Modified'
  })
  const currentFolderName = computed(
    () => breadcrumbs.value[breadcrumbs.value.length - 1]?.name ?? 'root',
  )
  const hasParentFolder = computed(() => breadcrumbs.value.length > 1)
  const totalPages = computed(() => Math.max(1, Math.ceil(totalCount.value / pageSize.value)))
  const startIndex = computed(() =>
    totalCount.value === 0 ? 0 : (page.value - 1) * pageSize.value + 1,
  )
  const endIndex = computed(() => Math.min(page.value * pageSize.value, totalCount.value))
  const pageNumbers = computed(() => {
    const total = totalPages.value
    if (total <= 3) return Array.from({ length: total }, (_, index) => index + 1)
    if (page.value <= 1) return [1, 2, 3]
    if (page.value >= total) return [total - 2, total - 1, total]
    return [page.value - 1, page.value, page.value + 1]
  })

  const allSelected = computed(() => {
    return (
      sortedFiles.value.length > 0 &&
      sortedFiles.value.every(item => selectedIds.value.has(item.key))
    )
  })
  const selectedCount = computed(
    () => sortedFiles.value.filter(item => selectedIds.value.has(item.key)).length,
  )

  const clearSelection = () => {
    selectedIds.value = new Set()
  }

  const cancelListRequest = () => {
    requestSequence += 1
    listAbortController?.abort()
    listAbortController = null
    isLoading.value = false
  }

  const fetchFolder = async (folderId = currentFolderId.value) => {
    listAbortController?.abort()
    const controller = new AbortController()
    listAbortController = controller
    const sequence = ++requestSequence

    isLoading.value = true
    errorMessage.value = null
    clearSelection()

    try {
      const count = await getListCountByFolderIDAndUserID(folderId, controller.signal)
      if (sequence !== requestSequence) return

      const safeCount = Number.isFinite(count) ? count : 0
      const pages = Math.max(1, Math.ceil(safeCount / pageSize.value))
      if (page.value > pages) page.value = pages

      const list = await getListByFolderIDAndUserID(
        folderId,
        page.value,
        pageSize.value,
        controller.signal,
      )
      if (sequence !== requestSequence) return

      totalCount.value = safeCount
      rawItems.value = Array.isArray(list) ? list : []
    } catch (error: unknown) {
      if (sequence !== requestSequence || isAbortError(error)) return
      rawItems.value = []
      totalCount.value = 0
      errorMessage.value = error instanceof Error ? error.message : '加载失败'
    } finally {
      if (sequence === requestSequence) {
        isLoading.value = false
        if (listAbortController === controller) listAbortController = null
      }
    }
  }

  const refresh = () => fetchFolder(currentFolderId.value)

  const onSearchInput = (value: string) => {
    searchQuery.value = value
    if (searchDebounceTimer) clearTimeout(searchDebounceTimer)
    searchDebounceTimer = setTimeout(() => {
      debouncedSearchQuery.value = value
      clearSelection()
      if (page.value !== 1) {
        page.value = 1
        void refresh()
      }
    }, 300)
  }

  const clearSearch = () => {
    searchQuery.value = ''
    debouncedSearchQuery.value = ''
    clearSelection()
    if (searchDebounceTimer) {
      clearTimeout(searchDebounceTimer)
      searchDebounceTimer = null
    }
  }

  const setFilter = (key: FileFilterKey) => {
    activeFilter.value = key
    clearSelection()
  }

  const setSort = (key: FileSortKey) => {
    if (sortKey.value === key) {
      sortDirection.value = sortDirection.value === 'asc' ? 'desc' : 'asc'
      return
    }
    sortKey.value = key
    sortDirection.value = 'asc'
  }

  const toggleAll = (checked: boolean) => {
    if (!checked) {
      clearSelection()
      return
    }
    selectedIds.value = new Set(sortedFiles.value.map(item => item.key))
  }

  const toggleOne = (item: Pick<FileListItem, 'id' | 'type'>, checked: boolean) => {
    const key = fileItemKey(item)
    const next = new Set(selectedIds.value)
    if (checked) next.add(key)
    else next.delete(key)
    selectedIds.value = next
  }

  const selectCurrentPage = () => {
    selectedIds.value = new Set(sortedFiles.value.map(item => item.key))
  }

  const goToPage = async (nextPage: number) => {
    const clamped = Math.min(Math.max(1, nextPage), totalPages.value)
    if (clamped === page.value) return
    page.value = clamped
    await refresh()
  }

  const goToFolder = async (folderId: number, folderName: string) => {
    currentFolderId.value = folderId
    page.value = 1
    breadcrumbs.value = [...breadcrumbs.value, { id: folderId, name: folderName }]
    await fetchFolder(folderId)
  }

  const goToBreadcrumb = async (index: number) => {
    const next = breadcrumbs.value[index]
    if (!next) return
    breadcrumbs.value = breadcrumbs.value.slice(0, index + 1)
    currentFolderId.value = next.id
    page.value = 1
    await fetchFolder(next.id)
  }

  const goToParentFolder = async () => {
    if (!hasParentFolder.value) return
    await goToBreadcrumb(breadcrumbs.value.length - 2)
  }

  const onRowClick = async (item: FileListItem) => {
    if (item.type !== 'folder') return
    await goToFolder(item.id, item.name)
  }

  onScopeDispose(() => {
    cancelListRequest()
    if (searchDebounceTimer) clearTimeout(searchDebounceTimer)
  })

  return {
    viewMode,
    sortKey,
    sortDirection,
    activeFilter,
    activeFilterLabel,
    sortLabel,
    searchQuery,
    debouncedSearchQuery,
    currentFolderId,
    breadcrumbs,
    currentFolderName,
    hasParentFolder,
    page,
    pageSize,
    totalCount,
    totalPages,
    startIndex,
    endIndex,
    pageNumbers,
    rawItems,
    sortedFiles,
    isLoading,
    errorMessage,
    selectedIds,
    selectedCount,
    allSelected,
    fetchFolder,
    refresh,
    cancelListRequest,
    onSearchInput,
    clearSearch,
    setFilter,
    setSort,
    clearSelection,
    toggleAll,
    toggleOne,
    selectCurrentPage,
    goToPage,
    goToFolder,
    goToBreadcrumb,
    goToParentFolder,
    onRowClick,
  }
}
