<script setup lang="ts">
import { Icon } from '@iconify/vue'
import { computed, onBeforeUnmount, onMounted, ref } from 'vue'
import { createPublicShareLink, deletePublicShareLink, getListByFolderIDAndUserID, getListCountByFolderIDAndUserID, getPublicShareLink, makeDirectory, previewFileById, uploadFile } from '../services/apis/file'
import type { FileListItem } from '../services/types/file'
import { formatBytes, formatTime, iconForListItem, typeLabelForListItem, detectFileType, iconForFile } from '../utils/file'
import { useUserStore } from '../stores/user'
import LoginRequiredPlaceholder from '../components/bussiness/LoginRequiredPlaceholder.vue'
import { createId } from '../utils/hash'
import type { UploadFileConfig } from '../types/file'

// 上传任务状态类型
type UploadTaskStatus = 'pending' | 'hashing' | 'uploading' | 'merging' | 'success' | 'failed' | 'canceled'

// 上传任务类型
type UploadTask = {
    id: string
    file: File
    targetFolderId: number
    relativePath: string // 相对于上传根目录的路径
    status: UploadTaskStatus
    percent: number
    message: string | null
    canceled: boolean
}

const userStore = useUserStore()

type ViewMode = 'list' | 'grid'
type SortKey = 'name' | 'size' | 'modified'
type SortDirection = 'asc' | 'desc'
type PreviewKind = 'image' | 'pdf' | 'video' | 'audio' | 'text' | 'unsupported'

type BreadcrumbItem = { id: number; name: string }
type DisplayItem = FileListItem & {
    icon: string
    iconBg: string
    iconFg: string
    typeLabel: string
    lastModifiedText: string
}

const viewMode = ref<ViewMode>('list')
const isSortOpen = ref(false)
const sortKey = ref<SortKey>('name')
const sortDirection = ref<SortDirection>('asc')
const openMenuId = ref<string | null>(null)
const menuTargetFile = ref<DisplayItem | null>(null)
const selectedIds = ref<Set<number>>(new Set())
const menuPosition = ref<{ top: number; left: number } | null>(null)

const currentFolderId = ref(0)
const breadcrumbs = ref<BreadcrumbItem[]>([{ id: 0, name: 'root' }])

const page = ref(1)
const pageSize = ref(10)
const totalCount = ref(0)

const isLoading = ref(false)
const errorMessage = ref<string | null>(null)

// 创建文件夹相关状态
const isCreateFolderModalOpen = ref(false)
const newFolderName = ref('')
const isCreatingFolder = ref(false)

// 上传相关状态
const uploadTasks = ref<UploadTask[]>([])
const isUploadPanelOpen = ref(false)
const isUploading = ref(false)
const fileInputRef = ref<HTMLInputElement | null>(null)
const folderInputRef = ref<HTMLInputElement | null>(null)

// 文件预览相关状态
const isPreviewModalOpen = ref(false)
const previewLoading = ref(false)
const previewError = ref<string | null>(null)
const previewingFile = ref<DisplayItem | null>(null)
const previewBlob = ref<Blob | null>(null)
const previewUrl = ref('')
const previewMimeType = ref('')
const previewTextContent = ref('')
const publicShareLink = ref('')
const shareError = ref<string | null>(null)
const isCreatingShareLink = ref(false)
const isDeletingShareLink = ref(false)

// Toast 提示状态
const toastMessage = ref('')
const toastType = ref<'success' | 'error' | 'info'>('info')
const showToast = ref(false)
let toastTimer: ReturnType<typeof setTimeout> | null = null

// 并发控制：最大同时上传文件数
const MAX_CONCURRENT_UPLOADS = 3

const rawItems = ref<FileListItem[]>([])

const iconForItem = (item: FileListItem) => {
    const meta = iconForListItem(item)
    return { icon: meta.icon, iconBg: meta.bg, iconFg: meta.fg }
}

const typeLabelForItem = (item: FileListItem) => {
    return typeLabelForListItem(item)
}

const fetchFolder = async (folderId: number) => {
    isLoading.value = true
    errorMessage.value = null
    openMenuId.value = null
    selectedIds.value = new Set()
    try {
        const count = await getListCountByFolderIDAndUserID(folderId)
        totalCount.value = Number.isFinite(count) ? count : 0
        const totalPages = Math.max(1, Math.ceil(totalCount.value / pageSize.value))
        if (page.value > totalPages) page.value = totalPages

        const list = await getListByFolderIDAndUserID(folderId, page.value, pageSize.value)
        rawItems.value = Array.isArray(list) ? list : []
    } catch (e: any) {
        rawItems.value = []
        totalCount.value = 0
        errorMessage.value = e?.message || '加载失败'
    } finally {
        isLoading.value = false
    }
}

const sortedFiles = computed(() => {
    const dir = sortDirection.value === 'asc' ? 1 : -1
    const mapped: DisplayItem[] = rawItems.value.map((item) => {
        const meta = iconForItem(item)
        return {
            ...item,
            ...meta,
            typeLabel: typeLabelForItem(item),
            lastModifiedText: formatTime(item.updated_at),
        }
    })
    const data = mapped
    data.sort((a, b) => {
        if (sortKey.value === 'name') return dir * a.name.localeCompare(b.name)
        if (sortKey.value === 'size') return dir * ((a.size ?? 0) - (b.size ?? 0))
        return dir * a.updated_at.localeCompare(b.updated_at)
    })
    return data
})

const currentFolderName = computed(() => breadcrumbs.value[breadcrumbs.value.length - 1]?.name || 'root')

const totalPages = computed(() => Math.max(1, Math.ceil(totalCount.value / pageSize.value)))
const startIndex = computed(() => (totalCount.value === 0 ? 0 : (page.value - 1) * pageSize.value + 1))
const endIndex = computed(() => Math.min(page.value * pageSize.value, totalCount.value))

const pageNumbers = computed(() => {
    const total = totalPages.value
    if (total <= 3) return Array.from({ length: total }, (_, i) => i + 1)
    if (page.value <= 1) return [1, 2, 3]
    if (page.value >= total) return [total - 2, total - 1, total]
    return [page.value - 1, page.value, page.value + 1]
})

const allSelected = computed(() => {
    const total = sortedFiles.value.length
    if (total === 0) return false
    return selectedIds.value.size === total
})

const selectedCount = computed(() => selectedIds.value.size)

const toggleAll = (checked: boolean) => {
    if (!checked) {
        selectedIds.value = new Set()
        return
    }
    selectedIds.value = new Set(sortedFiles.value.map((f) => f.id))
}

const toggleOne = (id: string, checked: boolean) => {
    const next = new Set(selectedIds.value)
    const numeric = Number(id)
    if (!Number.isFinite(numeric)) return
    if (checked) next.add(numeric)
    else next.delete(numeric)
    selectedIds.value = next
}

const goToPage = async (nextPage: number) => {
    const total = totalPages.value
    const clamped = Math.min(Math.max(1, nextPage), total)
    if (clamped === page.value) return
    page.value = clamped
    await fetchFolder(currentFolderId.value)
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

const onRowClick = async (file: FileListItem) => {
    if (file.type !== 'folder') return
    await goToFolder(file.id, file.name)
}

const setSort = (key: SortKey) => {
    if (sortKey.value === key) {
        sortDirection.value = sortDirection.value === 'asc' ? 'desc' : 'asc'
        isSortOpen.value = false
        return
    }
    sortKey.value = key
    sortDirection.value = 'asc'
    isSortOpen.value = false
}

const sortLabel = computed(() => {
    if (sortKey.value === 'name') return 'Name'
    if (sortKey.value === 'size') return 'Size'
    return 'Last Modified'
})

const ownerInitials = (name: string) => {
    const trimmed = name.trim()
    if (!trimmed) return '?'
    if (trimmed.toLowerCase() === 'me') return 'ME'
    const parts = trimmed.split(/\s+/).filter(Boolean)
    if (parts.length >= 2) return `${parts[0][0]}${parts[1][0]}`.toUpperCase()
    return trimmed.slice(0, 2).toUpperCase()
}

const getFileExt = (name: string) => {
    const idx = name.lastIndexOf('.')
    if (idx < 0 || idx === name.length - 1) return ''
    return name.slice(idx + 1).toLowerCase()
}

const normalizeMimeType = (mimeType: string) => {
    return (mimeType || '').split(';')[0].trim().toLowerCase()
}

const inferMimeType = (name: string, mimeType: string) => {
    const normalized = normalizeMimeType(mimeType)
    if (normalized && normalized !== 'application/octet-stream') return normalized
    const ext = getFileExt(name)
    if (['txt', 'md', 'json', 'csv', 'log', 'xml', 'yaml', 'yml', 'html', 'css', 'js', 'ts', 'vue'].includes(ext)) return 'text/plain'
    if (ext === 'pdf') return 'application/pdf'
    if (['png', 'jpg', 'jpeg', 'gif', 'webp', 'bmp', 'svg'].includes(ext)) return `image/${ext === 'jpg' ? 'jpeg' : ext}`
    if (['mp4', 'webm', 'ogg', 'mov', 'm4v', 'avi', 'mkv', 'mpeg'].includes(ext)) return `video/${ext === 'mov' ? 'mp4' : ext}`
    if (['mp3', 'wav', 'ogg', 'aac', 'flac', 'm4a'].includes(ext)) return `audio/${ext}`
    return 'application/octet-stream'
}

const isTextPreviewable = (mimeType: string, name: string) => {
    if (mimeType.startsWith('text/')) return true
    return ['md', 'json', 'csv', 'log', 'xml', 'yaml', 'yml', 'js', 'ts', 'vue', 'html', 'css'].includes(getFileExt(name))
}

const previewKind = computed<PreviewKind>(() => {
    const mime = normalizeMimeType(previewMimeType.value)
    if (!mime) return 'unsupported'
    if (mime.startsWith('image/')) return 'image'
    if (mime === 'application/pdf' || mime.includes('pdf')) return 'pdf'
    if (mime.startsWith('video/')) return 'video'
    if (mime.startsWith('audio/')) return 'audio'
    if (isTextPreviewable(mime, previewingFile.value?.name || '')) return 'text'
    return 'unsupported'
})

const revokePreviewUrl = () => {
    if (!previewUrl.value) return
    URL.revokeObjectURL(previewUrl.value)
    previewUrl.value = ''
}

const closePreviewModal = () => {
    isPreviewModalOpen.value = false
    previewLoading.value = false
    previewError.value = null
    previewingFile.value = null
    previewBlob.value = null
    previewMimeType.value = ''
    previewTextContent.value = ''
    publicShareLink.value = ''
    shareError.value = null
    isCreatingShareLink.value = false
    isDeletingShareLink.value = false
    revokePreviewUrl()
}

const triggerPreviewDownload = () => {
    if (!previewBlob.value || !previewingFile.value) return
    const url = URL.createObjectURL(previewBlob.value)
    const a = document.createElement('a')
    a.href = url
    a.download = previewingFile.value.name
    document.body.appendChild(a)
    a.click()
    a.remove()
    URL.revokeObjectURL(url)
}

const openPreviewModal = async (file: DisplayItem) => {
    if (file.type !== 'file') return
    closeOverlays()
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
    revokePreviewUrl()
    void loadExistingPublicShareLink(file.id)
    try {
        const { blob, contentType, fileName } = await previewFileById(file.id)
        const finalName = fileName || file.name
        const finalType = inferMimeType(finalName, contentType)
        previewBlob.value = blob
        previewMimeType.value = finalType
        if (isTextPreviewable(finalType, finalName)) {
            const maxTextSize = 1024 * 1024
            if (blob.size > maxTextSize) {
                previewTextContent.value = '文本文件较大，已为你展示下载入口，请下载后查看完整内容。'
            } else {
                previewTextContent.value = await blob.text()
            }
        }
        previewUrl.value = URL.createObjectURL(blob)
    } catch (e: any) {
        previewError.value = e?.message || '预览失败'
    } finally {
        previewLoading.value = false
    }
}

const handlePreviewFromMenu = async () => {
    if (!menuTargetFile.value) return
    await openPreviewModal(menuTargetFile.value)
}

const generatePublicShareLink = async () => {
    if (!previewingFile.value || previewingFile.value.type !== 'file') return
    isCreatingShareLink.value = true
    shareError.value = null
    try {
        const { token } = await createPublicShareLink(previewingFile.value.id)
        // 始终使用前端构造的 URL，确保包含 /api 前缀以便 Vite 代理正确转发
        publicShareLink.value = `${window.location.origin}/api/file/share/open?token=${encodeURIComponent(token)}`
        displayToast('分享链接已生成', 'success')
    } catch (e: any) {
        shareError.value = e?.message || '生成分享链接失败'
    } finally {
        isCreatingShareLink.value = false
    }
}

const loadExistingPublicShareLink = async (fileId: number) => {
    shareError.value = null
    try {
        const data = await getPublicShareLink(fileId)
        publicShareLink.value = data?.exists && data?.url ? data.url : ''
    } catch (e: any) {
        shareError.value = e?.message || '获取分享链接失败'
    }
}

const removePublicShareLink = async () => {
    if (!previewingFile.value || previewingFile.value.type !== 'file') return
    isDeletingShareLink.value = true
    shareError.value = null
    try {
        await deletePublicShareLink(previewingFile.value.id)
        publicShareLink.value = ''
        displayToast('分享链接已删除', 'success')
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
        displayToast('链接已复制', 'success')
    } catch (e) {
        displayToast('复制失败，请手动复制', 'error')
    }
}

const closeOverlays = () => {
    isSortOpen.value = false
    openMenuId.value = null
    menuTargetFile.value = null
    menuPosition.value = null
}

const onGlobalClick = () => closeOverlays()

const onStopPropagation = (e: MouseEvent) => e.stopPropagation()

const openFileMenu = (file: DisplayItem, event: MouseEvent) => {
    const menuId = `${file.type}-${file.id}`
    if (openMenuId.value === menuId) {
        openMenuId.value = null
        menuTargetFile.value = null
        menuPosition.value = null
        return
    }

    const button = event.currentTarget as HTMLButtonElement
    const rect = button.getBoundingClientRect()
    const menuHeight = 220 // 菜单大致高度
    const menuWidth = 192 // w-48 = 12rem = 192px
    const padding = 8

    // 计算可用空间
    const spaceBelow = window.innerHeight - rect.bottom
    const spaceAbove = rect.top

    // 决定菜单显示在上方还是下方
    let top: number
    if (spaceBelow < menuHeight && spaceAbove > spaceBelow) {
        // 下方空间不足，显示在上方
        top = rect.top - menuHeight - padding
    } else {
        // 默认显示在下方
        top = rect.bottom + padding
    }

    // 水平位置：右对齐
    let left = rect.right - menuWidth

    // 确保不超出视口左侧
    if (left < padding) {
        left = padding
    }

    menuPosition.value = { top, left }
    openMenuId.value = menuId
    menuTargetFile.value = file
}

const openCreateFolderModal = () => {
    newFolderName.value = ''
    isCreateFolderModalOpen.value = true
}

const closeCreateFolderModal = () => {
    isCreateFolderModalOpen.value = false
    newFolderName.value = ''
}

const handleCreateFolder = async () => {
    const name = newFolderName.value.trim()
    if (!name) {
        errorMessage.value = '文件夹名称不能为空'
        return
    }

    isCreatingFolder.value = true
    errorMessage.value = null

    try {
        const folderId = await makeDirectory({
            folder_id: currentFolderId.value,
            name: name,
        })

        if (folderId > 0) {
            closeCreateFolderModal()
            await fetchFolder(currentFolderId.value)
        } else {
            errorMessage.value = '创建文件夹失败'
        }
    } catch (e: any) {
        errorMessage.value = e?.message || '创建文件夹失败'
    } finally {
        isCreatingFolder.value = false
    }
}

// ============ 上传功能 ============

// 计算总体进度
const overallProgress = computed(() => {
    if (uploadTasks.value.length === 0) return 0
    const total = uploadTasks.value.reduce((sum, t) => sum + t.percent, 0)
    return Math.floor(total / uploadTasks.value.length)
})

// 打开文件选择对话框
const openFileDialog = () => {
    fileInputRef.value?.click()
}

// 打开文件夹选择对话框
const openFolderDialog = () => {
    folderInputRef.value?.click()
}

// 处理文件选择
const onFileInputChange = async (e: Event) => {
    const el = e.target as HTMLInputElement
    const fileList = el.files
    if (!fileList?.length) return

    const files = Array.from(fileList)
    el.value = '' // 清空输入，允许重复选择相同文件

    // 创建上传任务，所有文件都上传到当前文件夹
    // relativePath 设为空字符串，表示直接上传到目标文件夹，不需要创建子文件夹
    const tasks: UploadTask[] = files.map((file) => ({
        id: createId(),
        file,
        targetFolderId: currentFolderId.value,
        relativePath: '',
        status: 'pending',
        percent: 0,
        message: '等待中',
        canceled: false,
    }))

    await startUploadTasks(tasks)
}

// 处理文件夹选择
const onFolderInputChange = async (e: Event) => {
    const el = e.target as HTMLInputElement
    const fileList = el.files
    if (!fileList?.length) return

    const files = Array.from(fileList)
    el.value = '' // 清空输入

    // 解析文件夹结构
    // webkitRelativePath 格式: "folder/subfolder/file.txt"
    const tasks: UploadTask[] = files.map((file) => {
        const relativePath = file.webkitRelativePath || file.name
        // 获取文件所在的文件夹路径
        const lastSlashIndex = relativePath.lastIndexOf('/')
        const folderPath = lastSlashIndex > 0 ? relativePath.slice(0, lastSlashIndex) : ''

        return {
            id: createId(),
            file,
            targetFolderId: currentFolderId.value, // 初始目标文件夹，会在递归中更新
            relativePath: folderPath, // 相对于上传根目录的文件夹路径
            status: 'pending',
            percent: 0,
            message: '等待中',
            canceled: false,
        }
    })

    await startUploadTasks(tasks)
}

// 开始上传任务队列
const startUploadTasks = async (tasks: UploadTask[]) => {
    uploadTasks.value.unshift(...tasks)
    isUploadPanelOpen.value = true
    isUploading.value = true

    // 使用队列控制并发
    const queue = [...tasks]
    let activeCount = 0
    const results: Promise<void>[] = []

    const processTask = async (task: UploadTask): Promise<void> => {
        activeCount++
        try {
            await processUploadTask(task)
        } finally {
            activeCount--
            // 当这个任务完成时，尝试启动下一个
            processNext()
        }
    }

    const processNext = (): void => {
        while (activeCount < MAX_CONCURRENT_UPLOADS && queue.length > 0) {
            const task = queue.shift()
            if (task) {
                results.push(processTask(task))
            }
        }
    }

    try {
        // 启动初始任务
        processNext()

        // 等待所有任务完成
        await Promise.all(results)

        // 统计上传结果
        const successCount = tasks.filter(t => t.status === 'success').length
        const failedCount = tasks.filter(t => t.status === 'failed').length

        if (failedCount === 0) {
            displayToast(`成功上传 ${successCount} 个文件`, 'success')
        } else if (successCount === 0) {
            displayToast(`上传失败，${failedCount} 个文件未能上传`, 'error')
        } else {
            displayToast(`上传完成：${successCount} 个成功，${failedCount} 个失败`, 'info')
        }
    } finally {
        isUploading.value = false
        // 清除当前文件夹的缓存，确保能获取最新数据
        folderCache.clear()
        // 刷新当前文件夹列表 - 使用 finally 确保一定会执行
        await fetchFolder(currentFolderId.value)
    }
}

// 处理单个上传任务
const processUploadTask = async (task: UploadTask): Promise<void> => {
    if (task.canceled) return

    try {
        // 如果是文件夹上传，需要先创建文件夹结构
        let targetFolderId = task.targetFolderId

        if (task.relativePath) {
            targetFolderId = await ensureFolderStructure(task.targetFolderId, task.relativePath)
        }

        task.status = 'uploading'
        task.message = '上传中...'

        const fileType = detectFileType(task.file)
        const config: UploadFileConfig = {
            file_type: fileType,
            folder_id: targetFolderId,
        }

        // 使用 Promise 包装上传过程
        let lastProgress = 0
        await new Promise<void>((resolve, reject) => {
            uploadFile(
                task.file,
                config,
                (progress) => {
                    lastProgress = progress
                    if (task.canceled) {
                        reject(new Error('已取消'))
                        return
                    }
                    task.percent = Math.min(100, Math.max(0, Math.floor(progress)))
                    task.message = progress >= 100 ? '处理中...' : '上传中...'
                }
            ).then(() => {
                if (task.canceled) {
                    reject(new Error('已取消'))
                } else {
                    // 检查进度是否达到100%，uploadFile在失败时会重置为0
                    if (lastProgress >= 100) {
                        resolve()
                    } else {
                        reject(new Error('上传过程中断'))
                    }
                }
            }).catch((error) => {
                reject(error)
            })
        })

        task.status = 'success'
        task.percent = 100
        task.message = '上传完成'
    } catch (e: any) {
        if (task.canceled) return
        task.status = 'failed'
        task.message = e?.message || '上传失败'
        // 重新抛出错误，让上层知道失败了
        throw e
    }
}

// 文件夹缓存：路径 -> folderId
const folderCache = new Map<string, number>()
// 正在创建的文件夹锁：路径 -> Promise<void>
const creatingFolders = new Map<string, Promise<void>>()

// 确保文件夹结构存在，返回最终文件夹ID
const ensureFolderStructure = async (baseFolderId: number, relativePath: string): Promise<number> => {
    // 标准化路径（处理不同操作系统的分隔符）
    const normalizedPath = relativePath.replace(/\\/g, '/').replace(/^\//, '').replace(/\/$/, '')

    if (!normalizedPath) return baseFolderId

    // 检查缓存
    const cacheKey = `${baseFolderId}/${normalizedPath}`
    if (folderCache.has(cacheKey)) {
        return folderCache.get(cacheKey)!
    }

    // 如果正在创建中，等待完成
    if (creatingFolders.has(cacheKey)) {
        await creatingFolders.get(cacheKey)
        return folderCache.get(cacheKey) || baseFolderId
    }

    const parts = normalizedPath.split('/').filter(Boolean)
    let currentFolderId = baseFolderId
    let currentPath = ''

    for (const part of parts) {
        currentPath = currentPath ? `${currentPath}/${part}` : part
        const key = `${baseFolderId}/${currentPath}`
        // 捕获当前的 parentFolderId，避免闭包问题
        const parentFolderId = currentFolderId

        if (folderCache.has(key)) {
            currentFolderId = folderCache.get(key)!
        } else if (creatingFolders.has(key)) {
            // 等待其他任务创建完成
            await creatingFolders.get(key)
            currentFolderId = folderCache.get(key) || currentFolderId
        } else {
            // 创建文件夹锁
            const createPromise = (async () => {
                try {
                    // 使makeDirectory 返回新创建的文件夹 ID
                    const newFolderId = await makeDirectory({
                        folder_id: parentFolderId,
                        name: part,
                    })

                    if (newFolderId > 0) {
                        // 成功创建，使用返回的 ID
                        currentFolderId = newFolderId
                        folderCache.set(key, currentFolderId)
                    } else {
                        // 创建失败，尝试从列表中查找是否已存在
                        const list = await getListByFolderIDAndUserID(parentFolderId, 1, 100)
                        const existingFolder = list.find((item) => item.name === part && item.type === 'folder')
                        if (existingFolder) {
                            currentFolderId = existingFolder.id
                            folderCache.set(key, currentFolderId)
                        } else {
                            throw new Error(`创建文件夹 "${part}" 失败`)
                        }
                    }
                } catch (e) {
                    console.error(`创建文件夹 "${part}" 失败:`, e)
                    // 创建失败时，尝试从列表中查找是否已存在
                    try {
                        const list = await getListByFolderIDAndUserID(parentFolderId, 1, 100)
                        const existingFolder = list.find((item) => item.name === part && item.type === 'folder')
                        if (existingFolder) {
                            currentFolderId = existingFolder.id
                            folderCache.set(key, currentFolderId)
                        }
                    } catch (fetchError) {
                        console.error('获取文件夹列表失败:', fetchError)
                    }
                }
            })()

            creatingFolders.set(key, createPromise)
            await createPromise
            creatingFolders.delete(key)
        }
    }

    return currentFolderId
}

// 取消上传任务
const cancelTask = (task: UploadTask) => {
    task.canceled = true
    task.status = 'canceled'
    task.message = '已取消'
}

// 重试上传任务
const retryTask = async (task: UploadTask) => {
    task.status = 'pending'
    task.percent = 0
    task.message = '等待中'
    task.canceled = false
    await processUploadTask(task)
}

// 移除已完成或失败的任务
const removeTask = (taskId: string) => {
    uploadTasks.value = uploadTasks.value.filter((t) => t.id !== taskId)
}

// 清空已完成的任务
const clearCompleted = () => {
    uploadTasks.value = uploadTasks.value.filter((t) => t.status !== 'success')
}

// 清空所有任务
const clearAllTasks = () => {
    uploadTasks.value = []
}

// 切换上传面板
const toggleUploadPanel = () => {
    isUploadPanelOpen.value = !isUploadPanelOpen.value
}

// 显示 Toast 提示
const displayToast = (message: string, type: 'success' | 'error' | 'info' = 'info') => {
    if (toastTimer) {
        clearTimeout(toastTimer)
    }
    toastMessage.value = message
    toastType.value = type
    showToast.value = true
    toastTimer = setTimeout(() => {
        showToast.value = false
    }, 3000)
}

onMounted(() => {
    document.addEventListener('click', onGlobalClick)
    fetchFolder(0)
})

onBeforeUnmount(() => {
    document.removeEventListener('click', onGlobalClick)
    revokePreviewUrl()
})
</script>

<template>
    <div
        class="flex-1 flex flex-col min-w-0 bg-background-light dark:bg-background-dark font-display text-slate-900 dark:text-slate-100">
        <LoginRequiredPlaceholder v-if="!userStore.isLoggedIn" />
        <template v-else>
            <main class="flex-1 flex flex-col min-w-0 overflow-hidden">
                <div class="flex-1 overflow-y-auto p-8">
                    <div class="flex flex-col md:flex-row md:items-center justify-between gap-4 mb-8">
                        <div>
                            <nav class="flex items-center gap-2 text-sm text-slate-500 mb-2">
                                <button class="hover:text-primary flex items-center" type="button"
                                    @click="goToBreadcrumb(0)">
                                    <Icon class="text-sm mr-1" icon="material-symbols:home" />
                                    root
                                </button>
                                <template v-for="(bc, idx) in breadcrumbs.slice(1)" :key="bc.id">
                                    <Icon class="text-sm" icon="material-symbols:chevron-right" />
                                    <button v-if="idx + 1 < breadcrumbs.length - 1"
                                        class="hover:text-primary flex items-center" type="button"
                                        @click="goToBreadcrumb(idx + 1)">
                                        {{ bc.name }}
                                    </button>
                                    <span v-else class="text-slate-900 dark:text-slate-100 font-medium">{{ bc.name
                                    }}</span>
                                </template>
                            </nav>
                            <h2 class="text-2xl font-bold text-slate-900 dark:text-slate-100">{{ currentFolderName }}
                            </h2>
                            <p v-if="errorMessage" class="mt-2 text-sm text-red-500">{{ errorMessage }}</p>
                        </div>

                        <div class="flex items-center gap-3">
                            <!-- 隐藏的文件输入框 -->
                            <input ref="fileInputRef" type="file" class="hidden" multiple @change="onFileInputChange" />
                            <input ref="folderInputRef" type="file" class="hidden" webkitdirectory directory
                                @change="onFolderInputChange" />

                            <button
                                class="flex items-center gap-2 px-4 py-2 bg-white dark:bg-slate-950 border border-slate-200 dark:border-slate-800 text-slate-700 dark:text-slate-200 rounded-lg text-sm font-semibold hover:bg-slate-50 dark:hover:bg-slate-900 transition-all"
                                type="button" @click="openCreateFolderModal">
                                <Icon class="text-[20px]" icon="material-symbols:create-new-folder" />
                                新建文件夹
                            </button>

                            <!-- 上传下拉菜单 -->
                            <div class="relative group">
                                <button
                                    class="flex items-center gap-2 px-6 py-2 bg-primary text-white rounded-lg text-sm font-bold shadow-lg shadow-primary/20 hover:bg-primary/90 transition-all"
                                    type="button">
                                    <Icon class="text-[20px]" icon="material-symbols:upload" />
                                    上传
                                </button>
                                <!-- 下拉选项 -->
                                <div
                                    class="absolute right-0 top-full mt-2 w-48 bg-white dark:bg-slate-950 border border-slate-200 dark:border-slate-800 rounded-xl shadow-xl opacity-0 invisible group-hover:opacity-100 group-hover:visible transition-all z-20 py-2">
                                    <button
                                        class="w-full flex items-center gap-3 px-4 py-2 text-sm text-slate-700 dark:text-slate-300 hover:bg-slate-50 dark:hover:bg-slate-900"
                                        type="button" @click="openFileDialog">
                                        <Icon icon="material-symbols:description" />
                                        上传文件
                                    </button>
                                    <button
                                        class="w-full flex items-center gap-3 px-4 py-2 text-sm text-slate-700 dark:text-slate-300 hover:bg-slate-50 dark:hover:bg-slate-900"
                                        type="button" @click="openFolderDialog">
                                        <Icon icon="material-symbols:folder" />
                                        上传文件夹
                                    </button>
                                </div>
                            </div>

                            <!-- 上传进度按钮（当有任务时显示） -->
                            <button v-if="uploadTasks.length > 0"
                                class="flex items-center gap-2 px-4 py-2 bg-white dark:bg-slate-950 border border-slate-200 dark:border-slate-800 text-slate-700 dark:text-slate-200 rounded-lg text-sm font-semibold hover:bg-slate-50 dark:hover:bg-slate-900 transition-all"
                                :class="isUploadPanelOpen ? 'bg-slate-50 dark:bg-slate-900' : ''" type="button"
                                @click="toggleUploadPanel">
                                <Icon class="text-[20px]"
                                    :icon="isUploading ? 'material-symbols:progress-activity' : 'material-symbols:check-circle'"
                                    :class="isUploading ? 'animate-spin text-primary' : 'text-green-500'" />
                                <span v-if="isUploading">{{ overallProgress }}%</span>
                                <span v-else>{{uploadTasks.filter(t => t.status === 'success').length}}/{{
                                    uploadTasks.length }}</span>
                            </button>
                        </div>
                    </div>

                    <div
                        class="bg-white dark:bg-slate-950 rounded-xl border border-slate-200 dark:border-slate-800 mb-6 p-3 flex flex-wrap items-center justify-between gap-4 shadow-sm">
                        <div class="flex items-center gap-2">
                            <button class="p-2 rounded-lg"
                                :class="viewMode === 'list' ? 'bg-primary/10 text-primary' : 'text-slate-400 hover:bg-slate-100 dark:hover:bg-slate-900'"
                                title="List View" type="button" @click="viewMode = 'list'">
                                <Icon icon="material-symbols:list" />
                            </button>
                            <button class="p-2 rounded-lg"
                                :class="viewMode === 'grid' ? 'bg-primary/10 text-primary' : 'text-slate-400 hover:bg-slate-100 dark:hover:bg-slate-900'"
                                title="Grid View" type="button" @click="viewMode = 'grid'">
                                <Icon icon="material-symbols:grid-view" />
                            </button>

                            <div class="h-6 w-px bg-slate-200 dark:bg-slate-800 mx-2"></div>

                            <button
                                class="flex items-center gap-2 px-3 py-1.5 text-sm font-medium text-slate-600 dark:text-slate-400 hover:bg-slate-50 dark:hover:bg-slate-900 rounded-lg border border-transparent hover:border-slate-200 dark:hover:border-slate-800"
                                type="button">
                                <Icon class="text-[18px]" icon="material-symbols:filter-list" />
                                筛选
                            </button>

                            <div v-if="selectedCount > 0" class="ml-2 flex items-center gap-2 text-sm text-slate-500">
                                <span>已选择 {{ selectedCount }} 项</span>
                                <button class="text-primary font-semibold hover:underline" type="button"
                                    @click="toggleAll(false)">
                                    清空
                                </button>
                            </div>
                        </div>

                        <div class="relative" @click="onStopPropagation">
                            <div class="flex items-center gap-2 text-xs font-medium text-slate-400">
                                <span>Sorted by</span>
                                <button
                                    class="flex items-center gap-1 text-slate-900 dark:text-slate-100 hover:text-primary"
                                    type="button" @click="isSortOpen = !isSortOpen">
                                    {{ sortLabel }}
                                    <Icon class="text-sm" icon="material-symbols:expand-more" />
                                </button>
                            </div>

                            <div v-if="isSortOpen"
                                class="absolute right-0 mt-2 w-48 bg-white dark:bg-slate-950 border border-slate-200 dark:border-slate-800 rounded-xl shadow-xl z-20 py-2">
                                <button
                                    class="w-full text-left px-4 py-2 text-sm text-slate-700 dark:text-slate-300 hover:bg-slate-50 dark:hover:bg-slate-900"
                                    type="button" @click="setSort('name')">
                                    Name
                                </button>
                                <button
                                    class="w-full text-left px-4 py-2 text-sm text-slate-700 dark:text-slate-300 hover:bg-slate-50 dark:hover:bg-slate-900"
                                    type="button" @click="setSort('modified')">
                                    Last Modified
                                </button>
                                <button
                                    class="w-full text-left px-4 py-2 text-sm text-slate-700 dark:text-slate-300 hover:bg-slate-50 dark:hover:bg-slate-900"
                                    type="button" @click="setSort('size')">
                                    Size
                                </button>
                                <div class="border-t border-slate-100 dark:border-slate-800 my-1"></div>
                                <div class="px-4 py-2 text-xs text-slate-400">
                                    {{ sortDirection === 'asc' ? 'Ascending' : 'Descending' }}
                                </div>
                            </div>
                        </div>
                    </div>

                    <div v-if="viewMode === 'list'"
                        class="bg-white dark:bg-slate-950 rounded-xl border border-slate-200 dark:border-slate-800 shadow-sm overflow-hidden">
                        <table class="w-full text-left border-collapse">
                            <thead
                                class="bg-slate-50 dark:bg-slate-900/50 border-b border-slate-200 dark:border-slate-800">
                                <tr>
                                    <th class="py-4 px-6 w-10">
                                        <input class="rounded border-slate-300 text-primary focus:ring-primary/20"
                                            type="checkbox" :checked="allSelected"
                                            @change="toggleAll(($event.target as HTMLInputElement).checked)" />
                                    </th>
                                    <th class="py-4 px-4 text-xs font-bold uppercase tracking-wider text-slate-500">Name
                                    </th>
                                    <th
                                        class="py-4 px-4 text-xs font-bold uppercase tracking-wider text-slate-500 hidden md:table-cell">
                                        Type</th>
                                    <th
                                        class="py-4 px-4 text-xs font-bold uppercase tracking-wider text-slate-500 hidden sm:table-cell">
                                        Size</th>
                                    <th
                                        class="py-4 px-4 text-xs font-bold uppercase tracking-wider text-slate-500 hidden lg:table-cell">
                                        Owner</th>
                                    <th class="py-4 px-4 text-xs font-bold uppercase tracking-wider text-slate-500">Last
                                        Modified</th>
                                    <th
                                        class="py-4 px-6 text-right text-xs font-bold uppercase tracking-wider text-slate-500">
                                        Actions</th>
                                </tr>
                            </thead>

                            <tbody class="divide-y divide-slate-100 dark:divide-slate-800">
                                <tr v-for="file in sortedFiles" :key="file.id"
                                    class="group hover:bg-slate-50 dark:hover:bg-slate-900/40 transition-colors"
                                    :class="file.type === 'folder' ? 'cursor-pointer' : ''" @click="onRowClick(file)">
                                    <td class="py-4 px-6">
                                        <input class="rounded border-slate-300 text-primary focus:ring-primary/20"
                                            type="checkbox" :checked="selectedIds.has(file.id)" @click.stop
                                            @change="toggleOne(String(file.id), ($event.target as HTMLInputElement).checked)" />
                                    </td>
                                    <td class="py-4 px-4">
                                        <div class="flex items-center gap-3">
                                            <div class="w-10 h-10 rounded flex items-center justify-center"
                                                :class="`${file.iconBg} ${file.iconFg}`">
                                                <Icon :icon="file.icon" />
                                            </div>
                                            <span
                                                class="text-sm font-semibold text-slate-900 dark:text-slate-100 truncate max-w-[150px] md:max-w-xs">
                                                {{ file.name }}
                                            </span>
                                        </div>
                                    </td>
                                    <td class="py-4 px-4 text-sm text-slate-500 hidden md:table-cell">{{ file.typeLabel
                                    }}
                                    </td>
                                    <td class="py-4 px-4 text-sm text-slate-500 hidden sm:table-cell">
                                        {{ file.type === 'folder' ? '-' : formatBytes(file.size) }}
                                    </td>
                                    <td class="py-4 px-4 hidden lg:table-cell">
                                        <div class="flex items-center gap-2">
                                            <div
                                                class="w-6 h-6 rounded-full bg-slate-200 dark:bg-slate-800 text-slate-600 dark:text-slate-300 flex items-center justify-center text-[10px] font-bold">
                                                {{ ownerInitials('Me') }}
                                            </div>
                                            <span class="text-sm text-slate-600 dark:text-slate-400">Me</span>
                                        </div>
                                    </td>
                                    <td class="py-4 px-4 text-sm text-slate-500 whitespace-nowrap">{{
                                        file.lastModifiedText
                                    }}
                                    </td>
                                    <td class="py-4 px-6 text-right" @click="onStopPropagation">
                                        <button
                                            class="p-2 text-slate-400 hover:text-slate-600 dark:hover:text-slate-200 rounded-lg group-hover:bg-white dark:group-hover:bg-slate-900"
                                            type="button" @click="(e) => openFileMenu(file, e)">
                                            <Icon icon="material-symbols:more-vert" />
                                        </button>
                                    </td>
                                </tr>

                                <tr v-if="sortedFiles.length === 0" class="opacity-60">
                                    <td colspan="7" class="p-10 text-center text-slate-500 dark:text-slate-400">
                                        暂无文件
                                    </td>
                                </tr>
                            </tbody>
                        </table>

                        <div
                            class="px-6 py-4 bg-white dark:bg-slate-950 border-t border-slate-200 dark:border-slate-800 flex items-center justify-between">
                            <p class="text-xs text-slate-500">
                                Showing <span class="font-bold text-slate-900 dark:text-slate-100">{{ startIndex }}-{{
                                    endIndex }}</span>
                                of <span class="font-bold text-slate-900 dark:text-slate-100">{{ totalCount }}</span>
                                items
                            </p>
                            <div class="flex items-center gap-2">
                                <button
                                    class="p-1.5 border border-slate-200 dark:border-slate-800 rounded-lg text-slate-400 hover:bg-slate-50 dark:hover:bg-slate-900"
                                    type="button" :disabled="page <= 1 || isLoading" @click="goToPage(page - 1)">
                                    <Icon class="text-sm" icon="material-symbols:chevron-left" />
                                </button>
                                <button v-for="p in pageNumbers" :key="p"
                                    class="w-8 h-8 flex items-center justify-center rounded-lg text-xs font-medium"
                                    :class="p === page ? 'bg-primary text-white font-bold' : 'hover:bg-slate-50 dark:hover:bg-slate-900 text-slate-700 dark:text-slate-200'"
                                    type="button" :disabled="isLoading" @click="goToPage(p)">
                                    {{ p }}
                                </button>
                                <span v-if="totalPages > 3" class="text-slate-400 px-1">...</span>
                                <button
                                    class="p-1.5 border border-slate-200 dark:border-slate-800 rounded-lg text-slate-400 hover:bg-slate-50 dark:hover:bg-slate-900"
                                    type="button" :disabled="page >= totalPages || isLoading"
                                    @click="goToPage(page + 1)">
                                    <Icon class="text-sm" icon="material-symbols:chevron-right" />
                                </button>
                            </div>
                        </div>
                    </div>

                    <div v-else>
                        <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
                            <div v-for="file in sortedFiles" :key="file.id"
                                class="bg-white dark:bg-slate-950 rounded-xl border border-slate-200 dark:border-slate-800 shadow-sm p-4 hover:bg-slate-50 dark:hover:bg-slate-900/40 transition-colors"
                                :class="file.type === 'folder' ? 'cursor-pointer' : ''" @click="onRowClick(file)">
                                <div class="flex items-start justify-between gap-3">
                                    <div class="flex items-center gap-3 min-w-0">
                                        <div class="w-11 h-11 rounded-lg flex items-center justify-center shrink-0"
                                            :class="`${file.iconBg} ${file.iconFg}`">
                                            <Icon class="text-[22px]" :icon="file.icon" />
                                        </div>
                                        <div class="min-w-0">
                                            <div
                                                class="text-sm font-semibold text-slate-900 dark:text-slate-100 truncate">
                                                {{
                                                    file.name }}</div>
                                            <div class="text-xs text-slate-500">{{ file.typeLabel }}</div>
                                        </div>
                                    </div>
                                    <input class="rounded border-slate-300 text-primary focus:ring-primary/20"
                                        type="checkbox" :checked="selectedIds.has(file.id)" @click.stop
                                        @change="toggleOne(String(file.id), ($event.target as HTMLInputElement).checked)" />
                                </div>

                                <div class="mt-4 flex items-center justify-between text-xs text-slate-500">
                                    <span>{{ file.type === 'folder' ? '-' : formatBytes(file.size) }}</span>
                                    <span class="whitespace-nowrap">{{ file.lastModifiedText }}</span>
                                </div>

                                <div class="mt-3 flex items-center justify-between">
                                    <div class="flex items-center gap-2">
                                        <div
                                            class="w-6 h-6 rounded-full bg-slate-200 dark:bg-slate-800 text-slate-600 dark:text-slate-300 flex items-center justify-center text-[10px] font-bold">
                                            {{ ownerInitials('Me') }}
                                        </div>
                                        <span class="text-xs text-slate-500">Me</span>
                                    </div>
                                    <button
                                        class="p-2 text-slate-400 hover:text-slate-700 dark:hover:text-slate-200 rounded-lg hover:bg-white/50 dark:hover:bg-slate-900"
                                        type="button" @click.stop="(e) => openFileMenu(file, e)">
                                        <Icon icon="material-symbols:more-vert" />
                                    </button>
                                </div>
                            </div>

                            <div v-if="sortedFiles.length === 0"
                                class="col-span-full p-10 text-center text-slate-500 dark:text-slate-400">
                                暂无文件
                            </div>
                        </div>

                        <!-- 网格视图分页 -->
                        <div v-if="sortedFiles.length > 0"
                            class="mt-6 px-6 py-4 bg-white dark:bg-slate-950 border border-slate-200 dark:border-slate-800 rounded-xl flex items-center justify-between">
                            <p class="text-xs text-slate-500">
                                Showing <span class="font-bold text-slate-900 dark:text-slate-100">{{ startIndex }}-{{
                                    endIndex }}</span>
                                of <span class="font-bold text-slate-900 dark:text-slate-100">{{ totalCount }}</span>
                                items
                            </p>
                            <div class="flex items-center gap-2">
                                <button
                                    class="p-1.5 border border-slate-200 dark:border-slate-800 rounded-lg text-slate-400 hover:bg-slate-50 dark:hover:bg-slate-900"
                                    type="button" :disabled="page <= 1 || isLoading" @click="goToPage(page - 1)">
                                    <Icon class="text-sm" icon="material-symbols:chevron-left" />
                                </button>
                                <button v-for="p in pageNumbers" :key="p"
                                    class="w-8 h-8 flex items-center justify-center rounded-lg text-xs font-medium"
                                    :class="p === page ? 'bg-primary text-white font-bold' : 'hover:bg-slate-50 dark:hover:bg-slate-900 text-slate-700 dark:text-slate-200'"
                                    type="button" :disabled="isLoading" @click="goToPage(p)">
                                    {{ p }}
                                </button>
                                <span v-if="totalPages > 3" class="text-slate-400 px-1">...</span>
                                <button
                                    class="p-1.5 border border-slate-200 dark:border-slate-800 rounded-lg text-slate-400 hover:bg-slate-50 dark:hover:bg-slate-900"
                                    type="button" :disabled="page >= totalPages || isLoading"
                                    @click="goToPage(page + 1)">
                                    <Icon class="text-sm" icon="material-symbols:chevron-right" />
                                </button>
                            </div>
                        </div>
                    </div>
                </div>

                <!-- 创建文件夹模态框 -->
                <div v-if="isCreateFolderModalOpen"
                    class="fixed inset-0 bg-black/50 backdrop-blur-sm z-50 flex items-center justify-center p-4"
                    @click="closeCreateFolderModal">
                    <div class="bg-white dark:bg-slate-950 rounded-xl border border-slate-200 dark:border-slate-800 shadow-xl w-full max-w-md p-6"
                        @click.stop>
                        <h3 class="text-lg font-bold text-slate-900 dark:text-slate-100 mb-4">新建文件夹</h3>

                        <div class="mb-4">
                            <label class="block text-sm font-medium text-slate-700 dark:text-slate-300 mb-2">
                                文件夹名称
                            </label>
                            <input v-model="newFolderName" type="text"
                                class="w-full px-4 py-2 border border-slate-300 dark:border-slate-700 rounded-lg bg-white dark:bg-slate-900 text-slate-900 dark:text-slate-100 focus:outline-none focus:ring-2 focus:ring-primary/50"
                                placeholder="请输入文件夹名称" @keyup.enter="handleCreateFolder" autofocus />
                        </div>

                        <div class="flex justify-end gap-3">
                            <button
                                class="px-4 py-2 text-sm font-medium text-slate-600 dark:text-slate-400 hover:bg-slate-100 dark:hover:bg-slate-900 rounded-lg transition-colors"
                                type="button" @click="closeCreateFolderModal" :disabled="isCreatingFolder">
                                取消
                            </button>
                            <button
                                class="px-4 py-2 text-sm font-medium text-white bg-primary hover:bg-primary/90 rounded-lg transition-colors flex items-center gap-2"
                                type="button" @click="handleCreateFolder"
                                :disabled="isCreatingFolder || !newFolderName.trim()">
                                <Icon v-if="isCreatingFolder" icon="material-symbols:progress-activity"
                                    class="animate-spin" />
                                {{ isCreatingFolder ? '创建中...' : '创建' }}
                            </button>
                        </div>
                    </div>
                </div>

                <!-- 文件预览模态框 -->
                <div v-if="isPreviewModalOpen"
                    class="fixed inset-0 z-50 flex items-center justify-center p-4 sm:p-6 md:p-10 bg-slate-900/60 backdrop-blur-sm"
                    @click="closePreviewModal">
                    <div class="bg-white dark:bg-slate-950 w-full max-w-6xl h-full max-h-[850px] rounded-xl shadow-2xl flex flex-col overflow-hidden relative border border-slate-200 dark:border-slate-800"
                        @click.stop>
                        <div
                            class="px-6 py-4 border-b border-slate-100 dark:border-slate-800 flex items-center justify-between shrink-0">
                            <div class="flex items-center gap-3">
                                <div
                                    class="w-10 h-10 bg-primary/10 rounded-lg flex items-center justify-center text-primary">
                                    <Icon class="text-[22px]"
                                        :icon="previewingFile?.icon || 'material-symbols:description'" />
                                </div>
                                <div>
                                    <h2
                                        class="text-lg font-bold text-slate-900 dark:text-slate-100 tracking-tight truncate max-w-[60vw]">
                                        {{ previewingFile?.name || '文件预览' }}
                                    </h2>
                                    <p class="text-xs text-slate-400">
                                        {{ previewingFile?.typeLabel || 'File' }} • {{ formatBytes(previewingFile?.size
                                            || 0) }}
                                    </p>
                                </div>
                            </div>
                            <button
                                class="w-10 h-10 flex items-center justify-center rounded-full hover:bg-slate-100 dark:hover:bg-slate-900 text-slate-400 hover:text-slate-600 dark:hover:text-slate-200 transition-colors"
                                type="button" @click="closePreviewModal">
                                <Icon icon="material-symbols:close" />
                            </button>
                        </div>

                        <div class="flex flex-1 overflow-hidden">
                            <div class="flex-1 bg-slate-50 dark:bg-slate-900/30 relative flex flex-col overflow-hidden">
                                <div class="flex-1 overflow-y-auto p-6 md:p-8 pb-24">
                                    <div v-if="previewLoading"
                                        class="min-h-full flex flex-col items-center justify-center gap-3 text-slate-500">
                                        <Icon icon="material-symbols:progress-activity"
                                            class="text-3xl animate-spin text-primary" />
                                        <span class="text-sm">正在加载预览...</span>
                                    </div>

                                    <div v-else-if="previewError"
                                        class="min-h-full flex flex-col items-center justify-center text-center text-red-500 px-6">
                                        <Icon icon="material-symbols:error" class="text-4xl mb-2" />
                                        <p class="font-semibold">{{ previewError }}</p>
                                    </div>

                                    <div v-else-if="previewKind === 'image' && previewUrl"
                                        class="w-full min-h-full flex items-start justify-center">
                                        <img :src="previewUrl" :alt="previewingFile?.name"
                                            class="max-w-full h-auto object-contain rounded-lg shadow-sm border border-slate-200 dark:border-slate-700" />
                                    </div>

                                    <iframe v-else-if="previewKind === 'pdf' && previewUrl" :src="previewUrl"
                                        class="w-full min-h-[520px] h-[calc(100vh-300px)] rounded-lg border border-slate-200 dark:border-slate-700 bg-white"
                                        title="PDF 预览"></iframe>

                                    <div v-else-if="previewKind === 'video' && previewUrl"
                                        class="w-full min-h-full flex items-start justify-center">
                                        <video controls :src="previewUrl"
                                            class="max-w-full rounded-lg bg-black"></video>
                                    </div>

                                    <div v-else-if="previewKind === 'audio' && previewUrl"
                                        class="min-h-full flex items-center justify-center">
                                        <audio controls :src="previewUrl" class="w-full max-w-2xl"></audio>
                                    </div>

                                    <pre v-else-if="previewKind === 'text'"
                                        class="w-full h-full overflow-auto rounded-lg border border-slate-200 dark:border-slate-700 bg-white dark:bg-slate-950 p-4 text-sm text-slate-700 dark:text-slate-200 whitespace-pre-wrap break-words">
                        {{ previewTextContent || '文件为空' }}</pre>

                                    <div v-else
                                        class="min-h-full flex flex-col items-center justify-center text-center text-slate-500 px-6">
                                        <Icon icon="material-symbols:description" class="text-4xl mb-2" />
                                        <p class="font-semibold">当前文件类型暂不支持在线预览</p>
                                        <p class="text-xs mt-1">你可以使用下方按钮下载后查看</p>
                                    </div>
                                </div>

                                <div
                                    class="absolute bottom-6 left-1/2 -translate-x-1/2 bg-white dark:bg-slate-950 rounded-xl shadow-lg border border-slate-200 dark:border-slate-800 p-2 flex items-center gap-1 z-20">
                                    <button
                                        class="flex items-center gap-2 px-4 py-2 rounded-lg bg-primary hover:bg-primary/90 text-white font-semibold text-sm transition-all shadow-md shadow-primary/20 active:scale-95 disabled:opacity-60 disabled:cursor-not-allowed"
                                        type="button" :disabled="!previewBlob" @click="triggerPreviewDownload">
                                        <Icon class="text-[20px]" icon="material-symbols:download" />
                                        下载
                                    </button>
                                    <button
                                        class="flex items-center gap-2 px-4 py-2 rounded-lg hover:bg-primary/10 text-primary font-semibold text-sm transition-all active:scale-95 disabled:opacity-60 disabled:cursor-not-allowed"
                                        type="button" :disabled="isCreatingShareLink || !previewingFile"
                                        @click="generatePublicShareLink">
                                        <Icon class="text-[20px]"
                                            :icon="isCreatingShareLink ? 'material-symbols:progress-activity' : 'material-symbols:share'"
                                            :class="isCreatingShareLink ? 'animate-spin' : ''" />
                                        {{ isCreatingShareLink ? '生成中' : '分享' }}
                                    </button>
                                </div>

                                <div v-if="publicShareLink || shareError"
                                    class="absolute bottom-24 left-1/2 -translate-x-1/2 w-[min(90%,680px)] bg-white dark:bg-slate-950 rounded-xl shadow-lg border border-slate-200 dark:border-slate-800 p-3 z-20">
                                    <p class="text-xs text-slate-500 mb-2">公网分享链接（免鉴权访问）</p>
                                    <div v-if="publicShareLink" class="flex items-center gap-2">
                                        <input :value="publicShareLink" readonly
                                            class="flex-1 px-3 py-2 text-sm rounded-lg border border-slate-200 dark:border-slate-700 bg-slate-50 dark:bg-slate-900 text-slate-700 dark:text-slate-200" />
                                        <button
                                            class="px-3 py-2 text-sm font-semibold rounded-lg bg-primary text-white hover:bg-primary/90"
                                            type="button" @click="copyPublicShareLink">
                                            复制
                                        </button>
                                        <button
                                            class="px-3 py-2 text-sm font-semibold rounded-lg bg-red-500 text-white hover:bg-red-600 disabled:opacity-60 disabled:cursor-not-allowed"
                                            type="button" :disabled="isDeletingShareLink" @click="removePublicShareLink">
                                            {{ isDeletingShareLink ? '删除中' : '删除' }}
                                        </button>
                                    </div>
                                    <p v-else-if="shareError" class="text-sm text-red-500">{{ shareError }}</p>
                                </div>
                            </div>

                            <div
                                class="w-80 border-l border-slate-100 dark:border-slate-800 p-6 flex flex-col bg-white dark:bg-slate-950 overflow-y-auto">
                                <h3 class="text-xs font-bold text-slate-400 uppercase tracking-widest mb-6">文件详情</h3>
                                <div class="space-y-6">
                                    <div>
                                        <label class="text-[11px] text-slate-400 block mb-1">文件名</label>
                                        <p class="text-sm font-semibold text-slate-800 dark:text-slate-100 break-all">
                                            {{ previewingFile?.name || '-' }}</p>
                                    </div>
                                    <div>
                                        <label class="text-[11px] text-slate-400 block mb-1">文件大小</label>
                                        <p class="text-sm font-semibold text-slate-800 dark:text-slate-100">{{
                                            formatBytes(previewingFile?.size || 0) }}</p>
                                    </div>
                                    <div>
                                        <label class="text-[11px] text-slate-400 block mb-1">文件类型</label>
                                        <p class="text-sm font-semibold text-slate-800 dark:text-slate-100">{{
                                            previewingFile?.typeLabel || '-' }}</p>
                                    </div>
                                    <div>
                                        <label class="text-[11px] text-slate-400 block mb-1">最后修改</label>
                                        <p class="text-sm font-semibold text-slate-800 dark:text-slate-100">{{
                                            previewingFile?.lastModifiedText || '-' }}</p>
                                    </div>
                                </div>
                            </div>
                        </div>
                    </div>
                </div>

                <!-- 上传任务面板 -->
                <div v-if="uploadTasks.length > 0 && isUploadPanelOpen"
                    class="fixed bottom-4 right-4 w-96 bg-white dark:bg-slate-950 border border-slate-200 dark:border-slate-800 rounded-xl shadow-2xl z-40 flex flex-col max-h-[500px]">
                    <!-- 面板头部 -->
                    <div class="flex items-center justify-between p-4 border-b border-slate-200 dark:border-slate-800">
                        <div class="flex items-center gap-2">
                            <Icon v-if="isUploading" icon="material-symbols:progress-activity"
                                class="animate-spin text-primary" />
                            <Icon v-else icon="material-symbols:check-circle" class="text-green-500" />
                            <span class="font-semibold text-slate-900 dark:text-slate-100">
                                上传任务 ({{uploadTasks.filter(t => t.status === 'success').length}}/{{ uploadTasks.length
                                }})
                            </span>
                        </div>
                        <div class="flex items-center gap-1">
                            <button v-if="uploadTasks.some(t => t.status === 'success')"
                                class="p-1.5 text-slate-400 hover:text-slate-600 dark:hover:text-slate-200 rounded-lg hover:bg-slate-100 dark:hover:bg-slate-900"
                                type="button" @click="clearCompleted" title="清理已完成">
                                <Icon icon="material-symbols:cleaning-services" />
                            </button>
                            <button
                                class="p-1.5 text-slate-400 hover:text-slate-600 dark:hover:text-slate-200 rounded-lg hover:bg-slate-100 dark:hover:bg-slate-900"
                                type="button" @click="clearAllTasks" title="关闭面板">
                                <Icon icon="material-symbols:close" />
                            </button>
                        </div>
                    </div>

                    <!-- 总体进度 -->
                    <div
                        class="px-4 py-2 bg-slate-50 dark:bg-slate-900/50 border-b border-slate-200 dark:border-slate-800">
                        <div class="flex items-center justify-between text-xs mb-1">
                            <span class="text-slate-500">总体进度</span>
                            <span class="font-medium text-slate-700 dark:text-slate-300">{{ overallProgress }}%</span>
                        </div>
                        <div class="h-1.5 bg-slate-200 dark:bg-slate-800 rounded-full overflow-hidden">
                            <div class="h-1.5 bg-primary rounded-full transition-all duration-300"
                                :style="{ width: `${overallProgress}%` }"></div>
                        </div>
                    </div>

                    <!-- 任务列表 -->
                    <div class="flex-1 overflow-y-auto p-2 space-y-2 max-h-[350px]">
                        <div v-for="task in uploadTasks" :key="task.id"
                            class="flex items-center gap-3 p-3 rounded-lg border border-slate-100 dark:border-slate-800 bg-slate-50/50 dark:bg-slate-900/30">
                            <!-- 文件图标 -->
                            <div class="w-8 h-8 rounded-lg flex items-center justify-center shrink-0"
                                :class="iconForFile(task.file).bg + ' ' + iconForFile(task.file).fg">
                                <Icon :icon="iconForFile(task.file).icon" class="text-sm" />
                            </div>

                            <!-- 文件信息 -->
                            <div class="flex-1 min-w-0">
                                <div class="flex items-center justify-between mb-1">
                                    <p class="text-xs font-medium text-slate-800 dark:text-slate-100 truncate">
                                        {{ task.file.name }}
                                    </p>
                                    <span class="text-[10px] font-medium uppercase tracking-wider" :class="{
                                        'text-green-500': task.status === 'success',
                                        'text-red-500': task.status === 'failed',
                                        'text-slate-400': task.status === 'canceled',
                                        'text-primary': ['pending', 'hashing', 'uploading', 'merging'].includes(task.status)
                                    }">
                                        {{ task.status === 'success' ? '成功' : task.status === 'failed' ? '失败' :
                                            task.status
                                                === 'canceled' ? '取消' : task.percent + '%' }}
                                    </span>
                                </div>

                                <!-- 进度条 -->
                                <div class="h-1 bg-slate-200 dark:bg-slate-800 rounded-full overflow-hidden">
                                    <div class="h-1 rounded-full transition-all duration-200" :class="task.status === 'failed'
                                        ? 'bg-red-500' : task.status === 'success' ? 'bg-green-500' : 'bg-primary'"
                                        :style="{ width: task.percent + '%' }"></div>
                                </div>

                                <p class="text-[10px] text-slate-400 mt-1 truncate">{{ task.message }}</p>
                            </div>

                            <!-- 操作按钮 -->
                            <div class="flex items-center gap-1">
                                <button v-if="task.status === 'failed'"
                                    class="p-1 text-slate-400 hover:text-primary rounded transition-colors"
                                    type="button" @click="retryTask(task)">
                                    <Icon icon="material-symbols:replay" class="text-sm" />
                                </button>
                                <button v-if="['uploading', 'hashing', 'merging', 'pending'].includes(task.status)"
                                    class="p-1 text-slate-400 hover:text-red-500 rounded transition-colors"
                                    type="button" @click="cancelTask(task)">
                                    <Icon icon="material-symbols:close" class="text-sm" />
                                </button>
                                <button v-if="['success', 'failed', 'canceled'].includes(task.status)"
                                    class="p-1 text-slate-400 hover:text-red-500 rounded transition-colors"
                                    type="button" @click="removeTask(task.id)">
                                    <Icon icon="material-symbols:delete" class="text-sm" />
                                </button>
                            </div>
                        </div>
                    </div>
                </div>

                <!-- Toast 提示 -->
                <transition enter-active-class="transform ease-out duration-300 transition"
                    enter-from-class="translate-y-2 opacity-0" enter-to-class="translate-y-0 opacity-100"
                    leave-active-class="transition ease-in duration-200" leave-from-class="opacity-100"
                    leave-to-class="opacity-0">
                    <div v-if="showToast"
                        class="fixed top-4 right-4 z-50 flex items-center gap-3 px-6 py-4 rounded-xl shadow-2xl border"
                        :class="{
                            'bg-green-50 dark:bg-green-900/20 border-green-200 dark:border-green-800': toastType === 'success',
                            'bg-red-50 dark:bg-red-900/20 border-red-200 dark:border-red-800': toastType === 'error',
                            'bg-blue-50 dark:bg-blue-900/20 border-blue-200 dark:border-blue-800': toastType === 'info'
                        }">
                        <Icon v-if="toastType === 'success'" icon="material-symbols:check-circle"
                            class="text-2xl text-green-500" />
                        <Icon v-else-if="toastType === 'error'" icon="material-symbols:error"
                            class="text-2xl text-red-500" />
                        <Icon v-else icon="material-symbols:info" class="text-2xl text-blue-500" />
                        <span class="font-medium" :class="{
                            'text-green-800 dark:text-green-200': toastType === 'success',
                            'text-red-800 dark:text-red-200': toastType === 'error',
                            'text-blue-800 dark:text-blue-200': toastType === 'info'
                        }"> {{ toastMessage }}</span>
                    </div>
                </transition>

                <!-- 文件操作菜单 - 使用 Teleport 挂载到 body 避免被遮挡 -->
                <Teleport to="body">
                    <div v-if="openMenuId && menuPosition"
                        class="fixed w-48 bg-white dark:bg-slate-950 border border-slate-200 dark:border-slate-800 rounded-xl shadow-xl z-50 py-2"
                        :style="{ top: `${menuPosition.top}px`, left: `${menuPosition.left}px` }"
                        @click="onStopPropagation">
                        <button
                            class="w-full flex items-center gap-3 px-4 py-2 text-sm hover:bg-slate-50 dark:hover:bg-slate-900 disabled:opacity-50 disabled:cursor-not-allowed"
                            :class="menuTargetFile?.type === 'file' ? 'text-slate-700 dark:text-slate-300' : 'text-slate-400'"
                            type="button" :disabled="menuTargetFile?.type !== 'file'" @click="handlePreviewFromMenu">
                            <Icon class="text-sm" icon="material-symbols:visibility" />
                            预览
                        </button>
                        <button
                            class="w-full flex items-center gap-3 px-4 py-2 text-sm text-slate-700 dark:text-slate-300 hover:bg-slate-50 dark:hover:bg-slate-900"
                            type="button">
                            <Icon class="text-sm" icon="material-symbols:download" />
                            下载
                        </button>
                        <button
                            class="w-full flex items-center gap-3 px-4 py-2 text-sm text-slate-700 dark:text-slate-300 hover:bg-slate-50 dark:hover:bg-slate-900"
                            type="button">
                            <Icon class="text-sm" icon="material-symbols:edit" />
                            重命名
                        </button>
                        <button
                            class="w-full flex items-center gap-3 px-4 py-2 text-sm text-slate-700 dark:text-slate-300 hover:bg-slate-50 dark:hover:bg-slate-900"
                            type="button">
                            <Icon class="text-sm" icon="material-symbols:drive-file-move" />
                            移动到
                        </button>
                        <div class="border-t border-slate-100 dark:border-slate-800 my-1"></div>
                        <button
                            class="w-full flex items-center gap-3 px-4 py-2 text-sm text-red-500 hover:bg-red-50 dark:hover:bg-red-900/20"
                            type="button">
                            <Icon class="text-sm" icon="material-symbols:delete" />
                            删除
                        </button>
                    </div>
                </Teleport>
            </main>
        </template>
    </div>
</template>

<style lang="sass" scoped></style>
