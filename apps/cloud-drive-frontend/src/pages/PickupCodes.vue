<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { Icon } from '@iconify/vue'
import { getPickupCodeList, getPickupCodeCount } from '../services/apis/file'
import type { PickupCodeItem, PickupCodeType } from '../services/types/file'
import CreatePickupCodeModal from '../components/bussiness/CreatePickupCodeModal.vue'

const pickupList = ref<PickupCodeItem[]>([])
const loading = ref(false)
const showCreateModal = ref(false)
const showDetailModal = ref(false)
const selectedItem = ref<PickupCodeItem | null>(null)
const copySuccess = ref<string | null>(null)

const currentPage = ref(1)
const pageSize = ref(10)
const totalCount = ref(0)

const stats = computed(() => {
    const now = new Date()
    const weekLater = new Date(now.getTime() + 7 * 24 * 60 * 60 * 1000)

    return {
        activeCount: pickupList.value.filter(item => item.status === 'Active').length,
        totalDownloads: pickupList.value.reduce((sum, item) => sum + item.download, 0),
        expiringSoon: pickupList.value.filter(item => {
            const expireDate = new Date(item.expire_time)
            return expireDate >= now && expireDate <= weekLater
        }).length,
    }
})

const totalPages = computed(() => Math.max(1, Math.ceil(totalCount.value / pageSize.value)))
const startIndex = computed(() => (totalCount.value === 0 ? 0 : (currentPage.value - 1) * pageSize.value + 1))
const endIndex = computed(() => Math.min(currentPage.value * pageSize.value, totalCount.value))

const pageNumbers = computed(() => {
    const total = totalPages.value
    if (total <= 3) return Array.from({ length: total }, (_, i) => i + 1)
    if (currentPage.value <= 1) return [1, 2, 3]
    if (currentPage.value >= total) return [total - 2, total - 1, total]
    return [currentPage.value - 1, currentPage.value, currentPage.value + 1]
})

const progressPercent = (item: PickupCodeItem) => {
    if (item.max_download <= 0) return 0
    return Math.min(100, Math.round((item.download / item.max_download) * 100))
}

const typeIcon = (type: PickupCodeType) => {
    if (type === 'folder') return 'material-symbols:folder-zip-outline-rounded'
    return 'material-symbols:description-outline-rounded'
}

const statusLabel = (status: string) => {
    return status === 'Active' ? 'Active' : 'Expired'
}

const formatDate = (dateStr: string) => {
    const date = new Date(dateStr)
    return date.toISOString().split('T')[0]
}

const fetchData = async () => {
    loading.value = true
    try {
        const [list, count] = await Promise.all([
            getPickupCodeList(currentPage.value, pageSize.value),
            getPickupCodeCount()
        ])
        pickupList.value = list ?? []
        totalCount.value = count ?? 0
    } catch (error) {
        console.error('Failed to fetch pickup codes:', error)
        pickupList.value = []
        totalCount.value = 0
    } finally {
        loading.value = false
    }
}

const handlePageChange = (page: number) => {
    if (page < 1 || page > totalPages.value) return
    currentPage.value = page
    fetchData()
}

const openCreateModal = () => {
    showCreateModal.value = true
}

const closeCreateModal = () => {
    showCreateModal.value = false
}

const handleCreateSuccess = () => {
    showCreateModal.value = false
    fetchData()
}

const handleCopyCode = async (code: string) => {
    try {
        await navigator.clipboard.writeText(code)
        copySuccess.value = code
        setTimeout(() => {
            copySuccess.value = null
        }, 2000)
    } catch {
        console.error('Failed to copy code')
    }
}

const handleViewDetail = (item: PickupCodeItem) => {
    selectedItem.value = item
    showDetailModal.value = true
}

const handleCloseDetail = () => {
    showDetailModal.value = false
    selectedItem.value = null
}

onMounted(() => {
    fetchData()
})
</script>

<template>
    <div
        class="flex-1 flex flex-col min-w-0 bg-background-light dark:bg-background-dark font-display text-slate-900 dark:text-slate-100">
        <main class="flex-1 flex flex-col min-w-0 overflow-hidden">
            <div class="flex-1 overflow-y-auto p-6 lg:p-10 space-y-8">
                <div class="flex flex-col sm:flex-row justify-between items-start sm:items-center gap-4">
                    <div class="space-y-1">
                        <h2 class="text-3xl font-black tracking-tight text-slate-900 dark:text-white">取件码管理</h2>
                        <p class="text-slate-500 text-sm">创建、监控并撤销安全的文件提取码。</p>
                    </div>
                    <button
                        class="flex items-center gap-2 cursor-pointer bg-primary hover:bg-primary/90 text-white rounded-lg h-11 px-6 font-bold transition-all shadow-lg shadow-primary/20"
                        type="button" @click="openCreateModal">
                        <Icon icon="material-symbols:add-rounded" class="text-xl" />
                        <span>创建新取件码</span>
                    </button>
                </div>

                <div class="grid grid-cols-1 md:grid-cols-3 gap-6">
                    <div
                        class="bg-white dark:bg-slate-900 rounded-xl p-6 border border-slate-200 dark:border-slate-800 shadow-sm">
                        <div class="flex items-center justify-between mb-4">
                            <span class="text-slate-500 text-sm font-medium uppercase tracking-wider">活跃取件码</span>
                            <div class="bg-primary/10 p-2 rounded-lg text-primary">
                                <Icon icon="material-symbols:lock-open-outline-rounded" class="text-xl" />
                            </div>
                        </div>
                        <div class="flex items-baseline gap-2">
                            <p class="text-3xl font-black text-slate-900 dark:text-white">{{ stats.activeCount }}</p>
                            <span class="text-primary text-xs font-bold bg-primary/10 px-2 py-0.5 rounded-full">↑
                                12%</span>
                        </div>
                    </div>

                    <div
                        class="bg-white dark:bg-slate-900 rounded-xl p-6 border border-slate-200 dark:border-slate-800 shadow-sm">
                        <div class="flex items-center justify-between mb-4">
                            <span class="text-slate-500 text-sm font-medium uppercase tracking-wider">总下载量</span>
                            <div class="bg-blue-500/10 p-2 rounded-lg text-blue-500">
                                <Icon icon="material-symbols:download-rounded" class="text-xl" />
                            </div>
                        </div>
                        <div class="flex items-baseline gap-2">
                            <p class="text-3xl font-black text-slate-900 dark:text-white">{{
                                stats.totalDownloads.toLocaleString() }}</p>
                            <span class="text-blue-500 text-xs font-bold bg-blue-500/10 px-2 py-0.5 rounded-full">↑
                                5%</span>
                        </div>
                    </div>

                    <div
                        class="bg-white dark:bg-slate-900 rounded-xl p-6 border border-slate-200 dark:border-slate-800 shadow-sm">
                        <div class="flex items-center justify-between mb-4">
                            <span class="text-slate-500 text-sm font-medium uppercase tracking-wider">即将过期取件码</span>
                            <div class="bg-amber-500/10 p-2 rounded-lg text-amber-500">
                                <Icon icon="material-symbols:timer-outline-rounded" class="text-xl" />
                            </div>
                        </div>
                        <div class="flex items-baseline gap-2">
                            <p class="text-3xl font-black text-slate-900 dark:text-white">{{ stats.expiringSoon }}</p>
                            <span
                                class="text-slate-400 text-xs font-medium border border-slate-200 dark:border-slate-700 px-2 py-0.5 rounded-full">Next
                                7 days</span>
                        </div>
                    </div>
                </div>

                <div
                    class="bg-white dark:bg-slate-950 rounded-xl border border-slate-200 dark:border-slate-800 shadow-sm overflow-hidden">
                    <div
                        class="px-6 py-4 border-b border-slate-200 dark:border-slate-800 flex items-center justify-between bg-white dark:bg-slate-950">
                        <h3 class="font-bold text-slate-900 dark:text-white">活跃取件码仓库</h3>
                        <div class="flex items-center gap-2">
                            <button class="p-2 text-slate-500 hover:text-primary transition-colors" type="button">
                                <Icon icon="material-symbols:filter-list-rounded" class="text-xl" />
                            </button>
                            <button class="p-2 text-slate-500 hover:text-primary transition-colors" type="button">
                                <Icon icon="material-symbols:sort-rounded" class="text-xl" />
                            </button>
                        </div>
                    </div>

                    <div class="overflow-x-auto">
                        <table class="w-full text-left border-collapse">
                            <thead
                                class="bg-slate-50 dark:bg-slate-900/50 border-b border-slate-200 dark:border-slate-800">
                                <tr>
                                    <th class="py-4 px-6 text-xs font-bold uppercase tracking-wider text-slate-500">
                                        Pickup
                                        Code</th>
                                    <th class="py-4 px-4 text-xs font-bold uppercase tracking-wider text-slate-500">
                                        Associated File</th>
                                    <th
                                        class="py-4 px-4 text-xs font-bold uppercase tracking-wider text-slate-500 hidden sm:table-cell">
                                        Usage Progress</th>
                                    <th
                                        class="py-4 px-4 text-xs font-bold uppercase tracking-wider text-slate-500 hidden md:table-cell">
                                        Downloads</th>
                                    <th
                                        class="py-4 px-4 text-xs font-bold uppercase tracking-wider text-slate-500 hidden lg:table-cell">
                                        Expiration</th>
                                    <th class="py-4 px-4 text-xs font-bold uppercase tracking-wider text-slate-500">
                                        Status
                                    </th>
                                    <th
                                        class="py-4 px-6 text-right text-xs font-bold uppercase tracking-wider text-slate-500">
                                        Actions</th>
                                </tr>
                            </thead>

                            <tbody class="divide-y divide-slate-100 dark:divide-slate-800">
                                <tr v-if="loading" class="opacity-60">
                                    <td colspan="7" class="p-10 text-center text-slate-500 dark:text-slate-400">
                                        Loading...
                                    </td>
                                </tr>
                                <tr v-else-if="pickupList.length === 0" class="opacity-60">
                                    <td colspan="7" class="p-10 text-center text-slate-500 dark:text-slate-400">No
                                        pickup
                                        codes found. Create your first one!</td>
                                </tr>
                                <tr v-else v-for="item in pickupList" :key="item.code"
                                    class="group hover:bg-slate-50 dark:hover:bg-slate-900/40 transition-colors">
                                    <td class="py-4 px-6">
                                        <div
                                            class="font-mono text-lg font-bold text-slate-900 dark:text-white tracking-widest bg-slate-100 dark:bg-slate-800 px-3 py-1 rounded-lg inline-block border border-slate-200 dark:border-slate-700">
                                            {{ item.code }}
                                        </div>
                                    </td>
                                    <td class="py-4 px-4">
                                        <div class="flex items-center gap-3">
                                            <div
                                                class="w-10 h-10 rounded flex items-center justify-center bg-slate-100 dark:bg-slate-800 text-slate-500">
                                                <Icon :icon="typeIcon(item.type)" class="text-xl" />
                                            </div>
                                            <div class="truncate max-w-[150px] md:max-w-xs">
                                                <p
                                                    class="text-sm font-semibold text-slate-900 dark:text-slate-100 truncate">
                                                    {{ item.name }}</p>
                                                <p class="text-xs text-slate-500 truncate">{{ item.type }}</p>
                                            </div>
                                        </div>
                                    </td>
                                    <td class="py-4 px-4 hidden sm:table-cell">
                                        <div class="w-32">
                                            <div
                                                class="h-1.5 w-full bg-slate-100 dark:bg-slate-700 rounded-full overflow-hidden">
                                                <div class="h-full"
                                                    :class="item.status === 'Active' ? 'bg-primary' : 'bg-red-500'"
                                                    :style="{ width: `${progressPercent(item)}%` }"></div>
                                            </div>
                                        </div>
                                    </td>
                                    <td class="py-4 px-4 text-sm text-slate-500 hidden md:table-cell">{{ item.download
                                        }} /
                                        {{ item.max_download }}</td>
                                    <td class="py-4 px-4 text-sm text-slate-500 whitespace-nowrap hidden lg:table-cell">
                                        {{
                                            formatDate(item.expire_time) }}</td>
                                    <td class="py-4 px-4">
                                        <span
                                            class="inline-flex items-center px-2.5 py-1 rounded-full text-xs font-bold border"
                                            :class="item.status === 'Active'
                                                ? 'bg-primary/10 text-primary border-primary/20'
                                                : 'bg-red-100 text-red-600 border-red-200 dark:bg-red-900/20 dark:text-red-400 dark:border-red-800'">
                                            <span class="size-1.5 rounded-full mr-2"
                                                :class="item.status === 'Active' ? 'bg-primary' : 'bg-red-600 dark:bg-red-400'"></span>
                                            {{ statusLabel(item.status) }}
                                        </span>
                                    </td>
                                    <td class="py-4 px-6 text-right">
                                        <div
                                            class="flex items-center justify-end gap-2 opacity-0 group-hover:opacity-100 transition-opacity">
                                            <button class="p-2 rounded-lg transition-all"
                                                :class="copySuccess === item.code ? 'text-primary bg-primary/10' : 'text-slate-400 hover:text-primary hover:bg-primary/10'"
                                                :title="copySuccess === item.code ? 'Copied!' : 'Copy Code'"
                                                type="button" @click="handleCopyCode(item.code)">
                                                <Icon icon="material-symbols:content-copy-outline-rounded"
                                                    class="text-xl" />
                                            </button>
                                            <button
                                                class="p-2 text-slate-400 hover:text-primary hover:bg-primary/10 rounded-lg transition-all"
                                                title="View Details" type="button" @click="handleViewDetail(item)">
                                                <Icon icon="material-symbols:visibility-outline-rounded"
                                                    class="text-xl" />
                                            </button>
                                            <button
                                                class="p-2 text-slate-400 hover:text-red-500 hover:bg-red-50 rounded-lg transition-all opacity-50 cursor-not-allowed"
                                                title="Delete (coming soon)" type="button" disabled>
                                                <Icon icon="material-symbols:delete-outline-rounded"
                                                    class="text-xl text-red-500/80" />
                                            </button>
                                        </div>
                                    </td>
                                </tr>
                            </tbody>
                        </table>
                    </div>

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
                                type="button" :disabled="currentPage <= 1 || loading"
                                @click="handlePageChange(currentPage - 1)">
                                <Icon class="text-sm" icon="material-symbols:chevron-left" />
                            </button>
                            <button v-for="p in pageNumbers" :key="p"
                                class="w-8 h-8 flex items-center justify-center rounded-lg text-xs font-medium"
                                :class="p === currentPage ? 'bg-primary text-white font-bold' : 'hover:bg-slate-50 dark:hover:bg-slate-900 text-slate-700 dark:text-slate-200'"
                                type="button" :disabled="loading" @click="handlePageChange(p)">
                                {{ p }}
                            </button>
                            <span v-if="totalPages > 3" class="text-slate-400 px-1">...</span>
                            <button
                                class="p-1.5 border border-slate-200 dark:border-slate-800 rounded-lg text-slate-400 hover:bg-slate-50 dark:hover:bg-slate-900"
                                type="button" :disabled="currentPage >= totalPages || loading"
                                @click="handlePageChange(currentPage + 1)">
                                <Icon class="text-sm" icon="material-symbols:chevron-right" />
                            </button>
                        </div>
                    </div>
                </div>
            </div>
        </main>

        <!-- Create Pickup Code Modal -->
        <Teleport to="body">
            <div v-if="showCreateModal" class="fixed inset-0 z-50 flex items-center justify-center bg-black/50"
                @click.self="closeCreateModal">
                <CreatePickupCodeModal @close="closeCreateModal" @success="handleCreateSuccess" />
            </div>
        </Teleport>

        <!-- Detail Modal -->
        <Teleport to="body">
            <div v-if="showDetailModal && selectedItem"
                class="fixed inset-0 z-50 flex items-center justify-center bg-black/50" @click.self="handleCloseDetail">
                <div class="bg-white dark:bg-slate-900 rounded-xl shadow-2xl w-full max-w-md mx-4 overflow-hidden">
                    <div
                        class="px-6 py-4 border-b border-slate-200 dark:border-slate-800 flex items-center justify-between">
                        <h3 class="text-lg font-bold text-slate-900 dark:text-white">Pickup Code Details</h3>
                        <button
                            class="p-2 text-slate-400 hover:text-slate-600 dark:hover:text-slate-300 transition-colors"
                            type="button" @click="handleCloseDetail">
                            <Icon icon="material-symbols:close-rounded" class="text-xl" />
                        </button>
                    </div>

                    <div class="p-6 space-y-4">
                        <div class="flex items-center justify-center">
                            <div
                                class="font-mono text-2xl font-bold text-slate-900 dark:text-white tracking-widest bg-slate-100 dark:bg-slate-800 px-6 py-3 rounded-lg border border-slate-200 dark:border-slate-700">
                                {{ selectedItem.code }}
                            </div>
                        </div>

                        <div class="space-y-3">
                            <div class="flex justify-between">
                                <span class="text-slate-500">File Name</span>
                                <span class="font-medium text-slate-900 dark:text-white">{{ selectedItem.name }}</span>
                            </div>
                            <div class="flex justify-between">
                                <span class="text-slate-500">Type</span>
                                <span class="font-medium text-slate-900 dark:text-white">{{ selectedItem.type }}</span>
                            </div>
                            <div class="flex justify-between">
                                <span class="text-slate-500">Downloads</span>
                                <span class="font-medium text-slate-900 dark:text-white">{{ selectedItem.download }} /
                                    {{ selectedItem.max_download }}</span>
                            </div>
                            <div class="flex justify-between">
                                <span class="text-slate-500">Expiration</span>
                                <span class="font-medium text-slate-900 dark:text-white">{{
                                    formatDate(selectedItem.expire_time) }}</span>
                            </div>
                            <div class="flex justify-between">
                                <span class="text-slate-500">Status</span>
                                <span class="inline-flex items-center px-2.5 py-1 rounded-full text-xs font-bold border"
                                    :class="selectedItem.status === 'Active'
                                        ? 'bg-primary/10 text-primary border-primary/20'
                                        : 'bg-red-100 text-red-600 border-red-200 dark:bg-red-900/20 dark:text-red-400 dark:border-red-800'">
                                    {{ selectedItem.status === 'Active' ? 'Active' : 'Expired' }}
                                </span>
                            </div>
                        </div>
                    </div>

                    <div class="px-6 py-4 border-t border-slate-200 dark:border-slate-800 flex justify-end gap-3">
                        <button type="button"
                            class="px-4 py-2 rounded-lg text-sm font-medium text-slate-600 dark:text-slate-300 hover:bg-slate-100 dark:hover:bg-slate-800 transition-colors"
                            @click="handleCloseDetail">
                            Close
                        </button>
                        <button type="button"
                            class="px-6 py-2 rounded-lg text-sm font-medium bg-primary text-white hover:bg-primary/90 transition-colors"
                            @click="handleCopyCode(selectedItem.code); handleCloseDetail()">
                            Copy Code
                        </button>
                    </div>
                </div>
            </div>
        </Teleport>
    </div>
</template>

<style lang="sass" scoped></style>
