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
  colorClass: string
}

const storageUsedPercent = ref(0)
const storageUsedBytes = ref(0)
const storageTotalBytes = ref(0)
const storageLeftBytes = ref(0)
const isLoading = ref(false)

const storageUsedStr = computed(() => formatBytes(storageUsedBytes.value))
const storageTotalStr = computed(() => formatBytes(storageTotalBytes.value))
const storageLeftStr = computed(() => formatBytes(storageLeftBytes.value))

// 用量 ≥ 90% 转 warning 警示（DESIGN.md：warning 用于空间不足）
const isStorageWarning = computed(() => storageUsedPercent.value >= 90)

const fileStats = ref<FileStatCard[]>([])
const recentActivities = ref<RecentActivityCard[]>([])

const fileTypeMeta: Record<string, { title: string; icon: string; colorClass: string }> = {
  image: {
    title: '图片',
    icon: 'material-symbols:image-outline-rounded',
    colorClass: 'bg-info-tint text-info',
  },
  video: {
    title: '视频',
    icon: 'material-symbols:videocam-outline-rounded',
    colorClass: 'bg-pink-tint text-pink',
  },
  audio: {
    title: '音频',
    icon: 'material-symbols:music-note-rounded',
    colorClass: 'bg-purple-tint text-purple',
  },
  document: {
    title: '文档',
    icon: 'material-symbols:description-outline-rounded',
    colorClass: 'bg-warning-tint text-warning',
  },
  other: {
    title: '其他',
    icon: 'material-symbols:insert-drive-file-outline-rounded',
    colorClass: 'bg-surface-tertiary text-label-secondary',
  },
}

const dashboardTypeOrder = ['image', 'video', 'audio', 'document', 'other']

const mapFileStats = (stats: DashboardFileStatItem[]): FileStatCard[] => {
  const map = new Map(stats.map(item => [item.type, item]))
  return dashboardTypeOrder.map(type => {
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

const mapActivityMeta = (fileType: string) => {
  const meta = fileTypeMeta[fileType] ?? fileTypeMeta.other
  return { icon: meta.icon, colorClass: meta.colorClass }
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
  return activities.map(item => ({
    id: item.id,
    name: sanitizeFileName(item.name),
    desc: '位于',
    highlight: sanitizeFileName(item.folder_name || '根目录'),
    time: formatRelativeTime(item.updated_at),
    size: formatBytes(item.size),
    ...mapActivityMeta(item.file_type),
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
const dashOffset = computed(() => circumference - (circumference * storageUsedPercent.value) / 100)

watch(
  () => userStore.isLoggedIn,
  isLoggedIn => {
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
  <div class="h-full flex-1 overflow-y-auto">
    <LoginRequiredPlaceholder v-if="!userStore.isLoggedIn" />

    <main v-else class="space-y-6 p-4 sm:p-6 lg:space-y-8 lg:p-8">
      <!-- 页面标题 -->
      <div class="flex flex-col justify-between gap-4 md:flex-row md:items-center">
        <div>
          <h2 class="text-title-2">仪表盘</h2>
          <p class="text-subhead text-label-secondary">欢迎回来，这是您的存储情况。</p>
        </div>
      </div>

      <!-- 存储用量 + 类型统计 -->
      <div class="grid grid-cols-1 items-start gap-6 lg:grid-cols-12">
        <!-- 存储卡：品牌绿环形进度（≥90% 转 warning） -->
        <div class="rounded-md bg-surface p-6 shadow-card lg:col-span-4 lg:self-start">
          <div class="mb-3 flex items-start justify-between gap-4">
            <h3 class="text-title-3">存储使用情况</h3>
            <span
              class="flex h-9 w-9 flex-none items-center justify-center rounded-sm text-label-tertiary"
            >
              <Icon
                icon="material-symbols:info-outline-rounded"
                class="text-[20px]"
                aria-hidden="true"
              />
            </span>
          </div>

          <div class="relative flex items-center justify-center py-3">
            <svg class="h-32 w-32 -rotate-90">
              <circle
                class="text-surface-tertiary"
                cx="64"
                cy="64"
                fill="transparent"
                r="58"
                stroke="currentColor"
                stroke-width="8"
              />
              <circle
                :class="isStorageWarning ? 'text-warning' : 'text-primary'"
                class="transition-all duration-1000 ease-out"
                cx="64"
                cy="64"
                fill="transparent"
                r="58"
                stroke="currentColor"
                :stroke-dasharray="circumference"
                :stroke-dashoffset="dashOffset"
                stroke-width="8"
                stroke-linecap="round"
              />
            </svg>
            <div class="absolute inset-0 flex flex-col items-center justify-center">
              <span class="text-title-1" :class="isStorageWarning && 'text-danger'">
                {{ storageUsedPercent }}%
              </span>
              <span
                class="text-caption font-semibold uppercase tracking-[0.08em] text-label-tertiary"
              >
                已使用
              </span>
            </div>
          </div>

          <div class="mt-4 space-y-2.5 border-t border-hairline pt-4 text-subhead">
            <div class="flex items-center justify-between gap-4">
              <span class="text-label-secondary">已使用</span>
              <span class="font-semibold text-label">{{ storageUsedStr }}</span>
            </div>
            <div class="flex items-center justify-between gap-4">
              <span class="text-label-secondary">总容量</span>
              <span class="font-semibold text-label">{{ storageTotalStr }}</span>
            </div>
            <div class="flex items-center justify-between gap-4">
              <span class="text-label-secondary">剩余</span>
              <span
                class="font-semibold"
                :class="isStorageWarning ? 'text-danger' : 'text-primary'"
                >{{ storageLeftStr }}</span
              >
            </div>
          </div>
        </div>

        <!-- 文件类型统计：hover 微浮起 + 阴影加深 -->
        <div class="grid grid-cols-1 gap-6 sm:grid-cols-2 lg:col-span-8 xl:grid-cols-3">
          <div
            v-for="stat in fileStats"
            :key="stat.type"
            class="flex flex-col justify-between rounded-md bg-surface p-5 shadow-card transition-[transform,box-shadow] duration-200 ease-out hover:-translate-y-0.5 hover:shadow-popover"
          >
            <div
              :class="[
                'mb-4 flex h-10 w-10 items-center justify-center rounded-sm',
                stat.colorClass,
              ]"
            >
              <Icon :icon="stat.icon" class="text-[20px]" />
            </div>
            <div>
              <h4 class="text-subhead text-label-secondary">{{ stat.title }}</h4>
              <p class="text-title-1 mt-0.5">{{ stat.count }}</p>
              <p class="text-caption mt-1.5 text-label-tertiary">共 {{ stat.size }}</p>
            </div>
          </div>
        </div>
      </div>

      <!-- 最近活动：iCloud Drive 风格 48px 行 -->
      <div class="overflow-hidden rounded-md bg-surface shadow-card">
        <div class="flex items-center justify-between border-b border-hairline p-6">
          <h3 class="text-title-3">最近活动</h3>
          <span class="text-subhead text-label-secondary">{{
            isLoading ? '加载中...' : `${recentActivities.length} 条`
          }}</span>
        </div>

        <!-- 加载骨架 -->
        <div v-if="isLoading" class="space-y-3.5 p-6">
          <div v-for="n in 2" :key="n" class="flex animate-pulse items-center gap-3">
            <div class="h-8 w-8 flex-none rounded-sm bg-surface-secondary"></div>
            <div class="flex-1 space-y-1.5">
              <div class="h-3 w-1/2 rounded-full bg-surface-secondary"></div>
              <div class="h-3 w-1/3 rounded-full bg-surface-secondary"></div>
            </div>
          </div>
        </div>

        <template v-else>
          <p
            v-if="recentActivities.length === 0"
            class="py-8 text-center text-subhead text-label-secondary"
          >
            暂无最近活动
          </p>

          <template v-for="(activity, index) in recentActivities" :key="activity.id">
            <div
              class="flex min-h-12 items-center gap-3 px-6 transition-colors duration-150 hover:bg-surface-secondary"
            >
              <div
                class="flex h-8 w-8 flex-none items-center justify-center rounded-sm"
                :class="activity.colorClass"
              >
                <Icon :icon="activity.icon" class="text-[16px]" />
              </div>
              <div class="min-w-0 flex-1">
                <p class="truncate text-[15px] font-semibold leading-[22px]">
                  {{ activity.name }}
                </p>
                <p class="truncate text-subhead text-label-secondary">
                  {{ activity.desc }} <span class="text-primary">{{ activity.highlight }}</span>
                </p>
              </div>
              <div class="flex-none text-right">
                <p class="text-caption text-label-tertiary">{{ activity.time }}</p>
                <p class="text-caption mt-0.5 font-medium text-label-secondary">
                  {{ activity.size }}
                </p>
              </div>
            </div>
            <div v-if="index < recentActivities.length - 1" class="ml-17 h-px bg-hairline" />
          </template>
        </template>
      </div>
    </main>
  </div>
</template>
