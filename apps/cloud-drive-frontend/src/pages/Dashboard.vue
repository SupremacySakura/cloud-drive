<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { Icon } from '@iconify/vue'
import { useUserStore } from '../stores/user'
import LoginRequiredPlaceholder from '../components/bussiness/LoginRequiredPlaceholder.vue'
import { getDashboardOverview } from '../services/apis/file'
import type { DashboardFileStatItem, DashboardRecentActivityItem } from '../services/types/file'
import { formatBytes, sanitizeFileName } from '../utils/file'

const userStore = useUserStore()

type FileStatCard = {
    type: string
    title: string
    count: string
    size: string
    icon: string
    colorClass: string
}

type RecentActivityCard = {
    id: number
    name: string
    desc: string
    highlight: string
    time: string
    size: string
    icon: string
}

const storageUsedPercent = ref(0)
const storageUsedBytes = ref(0)
const storageTotalBytes = ref(0)
const storageLeftBytes = ref(0)
const isLoading = ref(false)

const storageUsedStr = computed(() => formatBytes(storageUsedBytes.value))
const storageTotalStr = computed(() => formatBytes(storageTotalBytes.value))
const storageLeftStr = computed(() => formatBytes(storageLeftBytes.value))

const fileStats = ref<FileStatCard[]>([])
const recentActivities = ref<RecentActivityCard[]>([])

const fileTypeMeta: Record<string, { title: string; icon: string; colorClass: string }> = {
    image: { title: '图片', icon: 'material-symbols:image-outline-rounded', colorClass: 'bg-blue-100 text-blue-600' },
    video: { title: '视频', icon: 'material-symbols:videocam-outline-rounded', colorClass: 'bg-red-100 text-red-600' },
    audio: { title: '音频', icon: 'material-symbols:music-note-rounded', colorClass: 'bg-purple-100 text-purple-600' },
    document: { title: '文档', icon: 'material-symbols:description-outline-rounded', colorClass: 'bg-orange-100 text-orange-600' },
    other: { title: '其他', icon: 'material-symbols:insert-drive-file-outline-rounded', colorClass: 'bg-slate-100 text-slate-600' },
}

const dashboardTypeOrder = ['image', 'video', 'audio', 'document', 'other']

const mapFileStats = (stats: DashboardFileStatItem[]): FileStatCard[] => {
    const map = new Map(stats.map((item) => [item.type, item]))
    return dashboardTypeOrder.map((type) => {
        const item = map.get(type)
        const meta = fileTypeMeta[type] ?? fileTypeMeta.other
        return {
            type,
            title: meta.title,
            count: (item?.count ?? 0).toLocaleString('zh-CN'),
            size: formatBytes(item?.size ?? 0),
            icon: meta.icon,
            colorClass: meta.colorClass,
        }
    })
}

const mapActivityIcon = (fileType: string) => {
    if (fileType === 'image') return 'material-symbols:image-outline-rounded'
    if (fileType === 'video') return 'material-symbols:videocam-outline-rounded'
    if (fileType === 'audio') return 'material-symbols:music-note-rounded'
    if (fileType === 'document') return 'material-symbols:description-outline-rounded'
    return 'material-symbols:insert-drive-file-outline-rounded'
}

const formatRelativeTime = (value: string) => {
    const timestamp = new Date(value).getTime()
    if (Number.isNaN(timestamp)) return '-'
    const diff = Date.now() - timestamp
    const minute = 60 * 1000
    const hour = 60 * minute
    const day = 24 * hour
    if (diff < minute) return '刚刚'
    if (diff < hour) return `${Math.floor(diff / minute)} 分钟前`
    if (diff < day) return `${Math.floor(diff / hour)} 小时前`
    return `${Math.floor(diff / day)} 天前`
}

const mapRecentActivities = (activities: DashboardRecentActivityItem[]): RecentActivityCard[] => {
    return activities.map((item) => ({
        id: item.id,
        name: sanitizeFileName(item.name),
        desc: '位于',
        highlight: sanitizeFileName(item.folder_name || '根目录'),
        time: formatRelativeTime(item.updated_at),
        size: formatBytes(item.size),
        icon: mapActivityIcon(item.file_type),
    }))
}

const resetDashboardData = () => {
    storageUsedPercent.value = 0
    storageUsedBytes.value = 0
    storageTotalBytes.value = 0
    storageLeftBytes.value = 0
    fileStats.value = mapFileStats([])
    recentActivities.value = []
}

const loadDashboardData = async () => {
    isLoading.value = true
    try {
        const data = await getDashboardOverview()
        storageUsedPercent.value = Math.max(0, Math.min(100, data.storage_used_percent))
        storageUsedBytes.value = data.storage_used
        storageTotalBytes.value = data.storage_total
        storageLeftBytes.value = data.storage_left
        fileStats.value = mapFileStats(data.file_stats ?? [])
        recentActivities.value = mapRecentActivities(data.recent_activities ?? [])
    } catch (error) {
        resetDashboardData()
        console.error('获取仪表盘数据失败', error)
    } finally {
        isLoading.value = false
    }
}

// Calculate SVG dash offset
const circumference = 364.4
const dashOffset = computed(() => circumference - (circumference * storageUsedPercent.value / 100))

watch(
    () => userStore.isLoggedIn,
    (isLoggedIn) => {
        if (!isLoggedIn) {
            resetDashboardData()
            return
        }
        loadDashboardData()
    },
    { immediate: true },
)
</script>

<template>
    <div
        class="flex-1 h-full bg-background-light dark:bg-background-dark font-display text-slate-900 dark:text-slate-100 overflow-y-auto">
        <LoginRequiredPlaceholder v-if="!userStore.isLoggedIn" />

        <main v-else class="p-8 space-y-8">
            <!-- Welcome and Quick Actions -->
            <div class="flex flex-col md:flex-row md:items-center justify-between gap-4">
                <div>
                    <h2 class="text-2xl font-bold">仪表盘</h2>
                    <p class="text-slate-500 text-sm">欢迎回来，这是您的存储情况。</p>
                </div>
            </div>

            <!-- Storage and Stats Grid -->
            <div class="grid grid-cols-1 lg:grid-cols-12 gap-6">
                <!-- Storage Card -->
                <div
                    class="lg:col-span-4 bg-white dark:bg-slate-900 p-6 rounded-xl border border-slate-200 dark:border-slate-800 shadow-sm">
                    <div class="flex justify-between items-start mb-6">
                        <h3 class="font-bold">存储使用情况</h3>
                        <Icon icon="material-symbols:info-outline-rounded" class="text-2xl text-slate-400" />
                    </div>

                    <div class="flex items-center justify-center py-6">
                        <div class="relative flex items-center justify-center">
                            <svg class="w-32 h-32 transform -rotate-90">
                                <circle class="text-slate-100 dark:text-slate-800" cx="64" cy="64" fill="transparent"
                                    r="58" stroke="currentColor" stroke-width="8"></circle>
                                <circle class="text-primary transition-all duration-1000 ease-out" cx="64" cy="64"
                                    fill="transparent" r="58" stroke="currentColor" :stroke-dasharray="circumference"
                                    :stroke-dashoffset="dashOffset" stroke-width="8" stroke-linecap="round"></circle>
                            </svg>
                            <div class="absolute inset-0 flex flex-col items-center justify-center">
                                <span class="text-2xl font-bold">{{ storageUsedPercent }}%</span>
                                <span class="text-[10px] uppercase font-bold text-slate-400">已使用</span>
                            </div>
                        </div>
                    </div>

                    <div class="space-y-4 mt-4">
                        <div class="flex justify-between text-sm">
                            <span class="text-slate-500">已使用 {{ storageUsedStr }} / 共 {{ storageTotalStr }}</span>
                            <span class="font-semibold">剩余 {{ storageLeftStr }}</span>
                        </div>
                    </div>
                </div>

                <!-- File Type Stats -->
                <div class="lg:col-span-8 grid grid-cols-1 sm:grid-cols-2 xl:grid-cols-3 gap-6">
                    <div v-for="stat in fileStats" :key="stat.type"
                        class="bg-white dark:bg-slate-900 p-6 rounded-xl border border-slate-200 dark:border-slate-800 shadow-sm flex flex-col justify-between hover:border-primary/30 transition-colors">
                        <div :class="['w-10 h-10 rounded-lg flex items-center justify-center mb-4', stat.colorClass]">
                            <Icon :icon="stat.icon" class="text-2xl" />
                        </div>
                        <div>
                            <h4 class="text-sm font-semibold text-slate-500">{{ stat.title }}</h4>
                            <p class="text-2xl font-bold mt-1">{{ stat.count }}</p>
                            <p class="text-xs text-slate-400 mt-2">共 {{ stat.size }}</p>
                        </div>
                    </div>
                </div>
            </div>

            <!-- Recent Activity -->
            <div
                class="bg-white dark:bg-slate-900 rounded-xl border border-slate-200 dark:border-slate-800 shadow-sm overflow-hidden">
                <div class="p-6 border-b border-slate-200 dark:border-slate-800 flex justify-between items-center">
                    <h3 class="font-bold">最近活动</h3>
                    <span class="text-sm text-slate-500">{{ isLoading ? '加载中...' : `${recentActivities.length} 条`
                        }}</span>
                </div>
                <div class="divide-y divide-slate-100 dark:divide-slate-800">
                    <div v-if="!isLoading && recentActivities.length === 0" class="p-6 text-sm text-slate-500">
                        暂无最近活动
                    </div>
                    <div v-for="activity in recentActivities" :key="activity.id"
                        class="p-4 flex items-center justify-between hover:bg-slate-50 dark:hover:bg-slate-800/50 transition-colors">
                        <div class="flex items-center gap-4">
                            <div
                                class="w-10 h-10 rounded bg-slate-100 dark:bg-slate-800 flex items-center justify-center">
                                <Icon :icon="activity.icon" class="text-2xl text-slate-500" />
                            </div>
                            <div>
                                <p class="text-sm font-semibold">{{ activity.name }}</p>
                                <p class="text-xs text-slate-500">{{ activity.desc }} <span class="text-primary">{{
                                    activity.highlight }}</span></p>
                            </div>
                        </div>
                        <div class="text-right">
                            <p class="text-xs text-slate-400">{{ activity.time }}</p>
                            <p class="text-xs font-medium text-slate-600 dark:text-slate-400 mt-1">{{ activity.size }}
                            </p>
                        </div>
                    </div>
                </div>
            </div>
        </main>
    </div>
</template>
