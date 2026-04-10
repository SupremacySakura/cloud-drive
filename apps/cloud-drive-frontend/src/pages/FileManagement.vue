<script setup lang="ts">
import { Icon } from '@iconify/vue'
import { computed, onBeforeUnmount, onMounted, ref } from 'vue'
import { getListByFolderIDAndUserID, getListCountByFolderIDAndUserID } from '../services/apis/file'
import type { FileListItem } from '../services/types/file'
import { formatBytes, formatTime, iconForListItem, typeLabelForListItem } from '../utils/file'
import { useUserStore } from '../stores/user'
import LoginRequiredPlaceholder from '../components/bussiness/LoginRequiredPlaceholder.vue'

const userStore = useUserStore()

type ViewMode = 'list' | 'grid'
type SortKey = 'name' | 'size' | 'modified'
type SortDirection = 'asc' | 'desc'

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
const selectedIds = ref<Set<number>>(new Set())

const currentFolderId = ref(0)
const breadcrumbs = ref<BreadcrumbItem[]>([{ id: 0, name: 'root' }])

const page = ref(1)
const pageSize = ref(10)
const totalCount = ref(0)

const isLoading = ref(false)
const errorMessage = ref<string | null>(null)

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

const closeOverlays = () => {
    isSortOpen.value = false
    openMenuId.value = null
}

const onGlobalClick = () => closeOverlays()

const onStopPropagation = (e: MouseEvent) => e.stopPropagation()

onMounted(() => {
    document.addEventListener('click', onGlobalClick)
    fetchFolder(0)
})

onBeforeUnmount(() => {
    document.removeEventListener('click', onGlobalClick)
})
</script>

<template>
    <LoginRequiredPlaceholder v-if="!userStore.isLoggedIn" />
    <div v-else
        class="flex-1 flex flex-col min-w-0 bg-background-light dark:bg-background-dark font-display text-slate-900 dark:text-slate-100">
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
                                <span v-else class="text-slate-900 dark:text-slate-100 font-medium">{{ bc.name }}</span>
                            </template>
                        </nav>
                        <h2 class="text-2xl font-bold text-slate-900 dark:text-slate-100">{{ currentFolderName }}</h2>
                        <p v-if="errorMessage" class="mt-2 text-sm text-red-500">{{ errorMessage }}</p>
                    </div>

                    <div class="flex items-center gap-3">
                        <button
                            class="flex items-center gap-2 px-4 py-2 bg-white dark:bg-slate-950 border border-slate-200 dark:border-slate-800 text-slate-700 dark:text-slate-200 rounded-lg text-sm font-semibold hover:bg-slate-50 dark:hover:bg-slate-900 transition-all"
                            type="button">
                            <Icon class="text-[20px]" icon="material-symbols:create-new-folder" />
                            新建文件夹
                        </button>
                        <button
                            class="flex items-center gap-2 px-6 py-2 bg-primary text-white rounded-lg text-sm font-bold shadow-lg shadow-primary/20 hover:bg-primary/90 transition-all"
                            type="button">
                            <Icon class="text-[20px]" icon="material-symbols:upload" />
                            上传
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
                        <thead class="bg-slate-50 dark:bg-slate-900/50 border-b border-slate-200 dark:border-slate-800">
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
                                <td class="py-4 px-4 text-sm text-slate-500 hidden md:table-cell">{{ file.typeLabel }}
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
                                <td class="py-4 px-4 text-sm text-slate-500 whitespace-nowrap">{{ file.lastModifiedText
                                }}
                                </td>
                                <td class="py-4 px-6 text-right relative" @click="onStopPropagation">
                                    <button
                                        class="p-2 text-slate-400 hover:text-slate-600 dark:hover:text-slate-200 rounded-lg group-hover:bg-white dark:group-hover:bg-slate-900"
                                        type="button"
                                        @click="openMenuId = openMenuId === String(file.id) ? null : String(file.id)">
                                        <Icon icon="material-symbols:more-vert" />
                                    </button>

                                    <div v-if="openMenuId === String(file.id)"
                                        class="absolute right-6 mt-2 w-48 bg-white dark:bg-slate-950 border border-slate-200 dark:border-slate-800 rounded-xl shadow-xl z-20 py-2">
                                        <button
                                            class="w-full flex items-center gap-3 px-4 py-2 text-sm text-slate-700 dark:text-slate-300 hover:bg-slate-50 dark:hover:bg-slate-900"
                                            type="button">
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
                            of <span class="font-bold text-slate-900 dark:text-slate-100">{{ totalCount }}</span> items
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
                                type="button" :disabled="page >= totalPages || isLoading" @click="goToPage(page + 1)">
                                <Icon class="text-sm" icon="material-symbols:chevron-right" />
                            </button>
                        </div>
                    </div>
                </div>

                <div v-else class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
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
                                    <div class="text-sm font-semibold text-slate-900 dark:text-slate-100 truncate">{{
                                        file.name }}</div>
                                    <div class="text-xs text-slate-500">{{ file.typeLabel }}</div>
                                </div>
                            </div>
                            <input class="rounded border-slate-300 text-primary focus:ring-primary/20" type="checkbox"
                                :checked="selectedIds.has(file.id)" @click.stop
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
                                type="button" @click.stop>
                                <Icon icon="material-symbols:more-vert" />
                            </button>
                        </div>
                    </div>

                    <div v-if="sortedFiles.length === 0"
                        class="col-span-full p-10 text-center text-slate-500 dark:text-slate-400">
                        暂无文件
                    </div>
                </div>
            </div>
        </main>
    </div>
</template>

<style lang="sass" scoped></style>
