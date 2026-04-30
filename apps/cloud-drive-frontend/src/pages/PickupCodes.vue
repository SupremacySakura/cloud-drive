<script setup lang="ts">
import { ref, computed, onMounted, onBeforeUnmount } from 'vue'
import { useRouter } from 'vue-router'
import { Icon } from '@iconify/vue'
import { getPickupCodeList, getPickupCodeCount, deletePickupCode } from '../services/apis/file'
import type { PickupCodeItem, PickupCodeType } from '../services/types/file'
import CreatePickupCodeModal from '../components/bussiness/CreatePickupCodeModal.vue'
import ConfirmDialog from '../components/ui/ConfirmDialog.vue'
import { sanitizeFileName } from '../utils/file'

const pickupList = ref<PickupCodeItem[]>([])
const loading = ref(false)
const showCreateModal = ref(false)
const showDetailModal = ref(false)
const selectedItem = ref<PickupCodeItem | null>(null)
const router = useRouter()
const openMenuId = ref<string | null>(null)
const menuTargetItem = ref<PickupCodeItem | null>(null)
const menuPosition = ref<{ top: number; left: number } | null>(null)
const deletingCodeId = ref<number | null>(null)
const showDeleteConfirm = ref(false)
const deleteTargetItem = ref<PickupCodeItem | null>(null)

const toastMessage = ref('')
const toastType = ref<'success' | 'error' | 'info'>('info')
const showToast = ref(false)
let toastTimer: ReturnType<typeof setTimeout> | null = null

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
const startIndex = computed(() =>
  totalCount.value === 0 ? 0 : (currentPage.value - 1) * pageSize.value + 1,
)
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
      getPickupCodeCount(),
    ])
    pickupList.value = (list ?? []).map(item => ({
      ...item,
      name: sanitizeFileName(item.name),
    }))
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

const goToFilePickup = () => {
  router.push('/pickup')
}

const closeCreateModal = () => {
  showCreateModal.value = false
}

const handleCreateSuccess = () => {
  showCreateModal.value = false
  fetchData()
}

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

const closeOverlays = () => {
  openMenuId.value = null
  menuTargetItem.value = null
  menuPosition.value = null
}

const onGlobalClick = () => closeOverlays()

const onStopPropagation = (e: MouseEvent) => e.stopPropagation()

const openPickupMenu = (item: PickupCodeItem, event: MouseEvent) => {
  const menuId = String(item.id)
  if (openMenuId.value === menuId) {
    closeOverlays()
    return
  }
  const button = event.currentTarget as HTMLButtonElement
  const rect = button.getBoundingClientRect()
  const menuHeight = 180
  const menuWidth = 192
  const padding = 8
  const spaceBelow = window.innerHeight - rect.bottom
  const spaceAbove = rect.top

  let top = rect.bottom + padding
  if (spaceBelow < menuHeight && spaceAbove > spaceBelow) {
    top = rect.top - menuHeight - padding
  }

  let left = rect.right - menuWidth
  if (left < padding) left = padding

  menuPosition.value = { top, left }
  openMenuId.value = menuId
  menuTargetItem.value = item
}

const handleCopyCode = async (code: string) => {
  if (!code) {
    displayToast('复制失败，请稍后重试', 'error')
    return
  }
  try {
    await navigator.clipboard.writeText(code)
    displayToast(`取件码 ${code} 已复制`, 'success')
    closeOverlays()
  } catch {
    displayToast('复制失败，请手动复制', 'error')
  }
}

const handleViewDetail = (item: PickupCodeItem) => {
  selectedItem.value = {
    ...item,
    name: sanitizeFileName(item.name),
  }
  showDetailModal.value = true
}

const handleViewDetailFromMenu = () => {
  if (!menuTargetItem.value) return
  handleViewDetail(menuTargetItem.value)
  closeOverlays()
}

const handleDeleteFromMenu = async () => {
  if (!menuTargetItem.value) return
  deleteTargetItem.value = menuTargetItem.value
  showDeleteConfirm.value = true
  closeOverlays()
}

const closeDeleteConfirm = () => {
  if (deletingCodeId.value !== null) return
  showDeleteConfirm.value = false
  deleteTargetItem.value = null
}

const confirmDeletePickupCode = async () => {
  if (!deleteTargetItem.value) return
  const target = deleteTargetItem.value
  deletingCodeId.value = target.id
  try {
    await deletePickupCode(target.id)
    displayToast('取件码删除成功', 'success')
    showDeleteConfirm.value = false
    deleteTargetItem.value = null
    await fetchData()
  } catch (e: any) {
    displayToast(e?.message || '删除取件码失败', 'error')
  } finally {
    deletingCodeId.value = null
  }
}

const handleCloseDetail = () => {
  showDetailModal.value = false
  selectedItem.value = null
}

onMounted(() => {
  document.addEventListener('click', onGlobalClick)
  fetchData()
})

onBeforeUnmount(() => {
  document.removeEventListener('click', onGlobalClick)
  if (toastTimer) {
    clearTimeout(toastTimer)
  }
})
</script>

<template>
  <div
    class="flex-1 flex flex-col min-w-0 bg-background-light dark:bg-background-dark font-display text-slate-900 dark:text-slate-100"
  >
    <main class="flex-1 flex flex-col min-w-0 overflow-hidden">
      <div class="flex-1 overflow-y-auto p-6 lg:p-10 space-y-8">
        <div class="flex flex-col sm:flex-row justify-between items-start sm:items-center gap-4">
          <div class="space-y-1">
            <h2 class="text-3xl font-black tracking-tight text-slate-900 dark:text-white">
              取件码管理
            </h2>
            <p class="text-slate-500 text-sm">创建、监控并撤销安全的文件提取码。</p>
          </div>
          <div class="flex items-center gap-3">
            <button
              class="flex items-center gap-2 cursor-pointer border border-primary/30 text-primary hover:bg-primary/5 rounded-lg h-11 px-6 font-bold transition-all"
              type="button"
              @click="goToFilePickup"
            >
              <Icon icon="material-symbols:download-rounded" class="text-xl" />
              <span>去取件</span>
            </button>
            <button
              class="flex items-center gap-2 cursor-pointer bg-primary hover:bg-primary/90 text-white rounded-lg h-11 px-6 font-bold transition-all shadow-lg shadow-primary/20"
              type="button"
              @click="openCreateModal"
            >
              <Icon icon="material-symbols:add-rounded" class="text-xl" />
              <span>创建新取件码</span>
            </button>
          </div>
        </div>

        <div class="grid grid-cols-1 md:grid-cols-3 gap-6">
          <div
            class="bg-white dark:bg-slate-900 rounded-xl p-6 border border-slate-200 dark:border-slate-800 shadow-sm"
          >
            <div class="flex items-center justify-between mb-4">
              <span class="text-slate-500 text-sm font-medium uppercase tracking-wider"
                >活跃取件码</span
              >
              <div class="bg-primary/10 p-2 rounded-lg text-primary">
                <Icon icon="material-symbols:lock-open-outline-rounded" class="text-xl" />
              </div>
            </div>
            <div class="flex items-baseline gap-2">
              <p class="text-3xl font-black text-slate-900 dark:text-white">
                {{ stats.activeCount }}
              </p>
              <span class="text-primary text-xs font-bold bg-primary/10 px-2 py-0.5 rounded-full"
                >↑ 12%</span
              >
            </div>
          </div>

          <div
            class="bg-white dark:bg-slate-900 rounded-xl p-6 border border-slate-200 dark:border-slate-800 shadow-sm"
          >
            <div class="flex items-center justify-between mb-4">
              <span class="text-slate-500 text-sm font-medium uppercase tracking-wider"
                >总下载量</span
              >
              <div class="bg-blue-500/10 p-2 rounded-lg text-blue-500">
                <Icon icon="material-symbols:download-rounded" class="text-xl" />
              </div>
            </div>
            <div class="flex items-baseline gap-2">
              <p class="text-3xl font-black text-slate-900 dark:text-white">
                {{ stats.totalDownloads.toLocaleString() }}
              </p>
              <span class="text-blue-500 text-xs font-bold bg-blue-500/10 px-2 py-0.5 rounded-full"
                >↑ 5%</span
              >
            </div>
          </div>

          <div
            class="bg-white dark:bg-slate-900 rounded-xl p-6 border border-slate-200 dark:border-slate-800 shadow-sm"
          >
            <div class="flex items-center justify-between mb-4">
              <span class="text-slate-500 text-sm font-medium uppercase tracking-wider"
                >即将过期取件码</span
              >
              <div class="bg-amber-500/10 p-2 rounded-lg text-amber-500">
                <Icon icon="material-symbols:timer-outline-rounded" class="text-xl" />
              </div>
            </div>
            <div class="flex items-baseline gap-2">
              <p class="text-3xl font-black text-slate-900 dark:text-white">
                {{ stats.expiringSoon }}
              </p>
              <span
                class="text-slate-400 text-xs font-medium border border-slate-200 dark:border-slate-700 px-2 py-0.5 rounded-full"
                >Next 7 days</span
              >
            </div>
          </div>
        </div>

        <div
          class="bg-white dark:bg-slate-950 rounded-xl border border-slate-200 dark:border-slate-800 shadow-sm overflow-hidden"
        >
          <div
            class="px-6 py-4 border-b border-slate-200 dark:border-slate-800 flex items-center justify-between bg-white dark:bg-slate-950"
          >
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
                class="bg-slate-50 dark:bg-slate-900/50 border-b border-slate-200 dark:border-slate-800"
              >
                <tr>
                  <th class="py-4 px-6 text-xs font-bold uppercase tracking-wider text-slate-500">
                    Pickup Code
                  </th>
                  <th class="py-4 px-4 text-xs font-bold uppercase tracking-wider text-slate-500">
                    Associated File
                  </th>
                  <th
                    class="py-4 px-4 text-xs font-bold uppercase tracking-wider text-slate-500 hidden sm:table-cell"
                  >
                    Usage Progress
                  </th>
                  <th
                    class="py-4 px-4 text-xs font-bold uppercase tracking-wider text-slate-500 hidden md:table-cell"
                  >
                    Downloads
                  </th>
                  <th
                    class="py-4 px-4 text-xs font-bold uppercase tracking-wider text-slate-500 hidden lg:table-cell"
                  >
                    Expiration
                  </th>
                  <th class="py-4 px-4 text-xs font-bold uppercase tracking-wider text-slate-500">
                    Status
                  </th>
                  <th
                    class="py-4 px-6 text-right text-xs font-bold uppercase tracking-wider text-slate-500"
                  >
                    Actions
                  </th>
                </tr>
              </thead>

              <tbody class="divide-y divide-slate-100 dark:divide-slate-800">
                <tr v-if="loading" class="opacity-60">
                  <td colspan="7" class="p-10 text-center text-slate-500 dark:text-slate-400">
                    Loading...
                  </td>
                </tr>
                <tr v-else-if="pickupList.length === 0" class="opacity-60">
                  <td colspan="7" class="p-10 text-center text-slate-500 dark:text-slate-400">
                    No pickup codes found. Create your first one!
                  </td>
                </tr>
                <tr
                  v-else
                  v-for="item in pickupList"
                  :key="item.code"
                  class="group hover:bg-slate-50 dark:hover:bg-slate-900/40 transition-colors"
                >
                  <td class="py-4 px-6">
                    <div
                      class="font-mono text-lg font-bold text-slate-900 dark:text-white tracking-widest bg-slate-100 dark:bg-slate-800 px-3 py-1 rounded-lg inline-block border border-slate-200 dark:border-slate-700"
                    >
                      {{ item.code }}
                    </div>
                  </td>
                  <td class="py-4 px-4">
                    <div class="flex items-center gap-3">
                      <div
                        class="w-10 h-10 rounded flex items-center justify-center bg-slate-100 dark:bg-slate-800 text-slate-500"
                      >
                        <Icon :icon="typeIcon(item.type)" class="text-xl" />
                      </div>
                      <div class="truncate max-w-[150px] md:max-w-xs">
                        <p
                          class="text-sm font-semibold text-slate-900 dark:text-slate-100 truncate"
                        >
                          {{ item.name }}
                        </p>
                        <p class="text-xs text-slate-500 truncate">{{ item.type }}</p>
                      </div>
                    </div>
                  </td>
                  <td class="py-4 px-4 hidden sm:table-cell">
                    <div class="w-32">
                      <div
                        class="h-1.5 w-full bg-slate-100 dark:bg-slate-700 rounded-full overflow-hidden"
                      >
                        <div
                          class="h-full"
                          :class="item.status === 'Active' ? 'bg-primary' : 'bg-red-500'"
                          :style="{ width: `${progressPercent(item)}%` }"
                        ></div>
                      </div>
                    </div>
                  </td>
                  <td class="py-4 px-4 text-sm text-slate-500 hidden md:table-cell">
                    {{ item.download }} / {{ item.max_download }}
                  </td>
                  <td
                    class="py-4 px-4 text-sm text-slate-500 whitespace-nowrap hidden lg:table-cell"
                  >
                    {{ formatDate(item.expire_time) }}
                  </td>
                  <td class="py-4 px-4">
                    <span
                      class="inline-flex items-center px-2.5 py-1 rounded-full text-xs font-bold border"
                      :class="
                        item.status === 'Active'
                          ? 'bg-primary/10 text-primary border-primary/20'
                          : 'bg-red-100 text-red-600 border-red-200 dark:bg-red-900/20 dark:text-red-400 dark:border-red-800'
                      "
                    >
                      <span
                        class="size-1.5 rounded-full mr-2"
                        :class="
                          item.status === 'Active' ? 'bg-primary' : 'bg-red-600 dark:bg-red-400'
                        "
                      ></span>
                      {{ statusLabel(item.status) }}
                    </span>
                  </td>
                  <td class="py-4 px-6 text-right" @click="onStopPropagation">
                    <button
                      class="p-2 text-slate-400 hover:text-slate-700 dark:hover:text-slate-200 rounded-lg group-hover:bg-white dark:group-hover:bg-slate-900"
                      type="button"
                      @click="e => openPickupMenu(item, e)"
                    >
                      <Icon icon="material-symbols:more-vert" />
                    </button>
                  </td>
                </tr>
              </tbody>
            </table>
          </div>

          <div
            class="px-6 py-4 bg-white dark:bg-slate-950 border-t border-slate-200 dark:border-slate-800 flex items-center justify-between"
          >
            <p class="text-xs text-slate-500">
              Showing
              <span class="font-bold text-slate-900 dark:text-slate-100"
                >{{ startIndex }}-{{ endIndex }}</span
              >
              of
              <span class="font-bold text-slate-900 dark:text-slate-100">{{ totalCount }}</span>
              items
            </p>
            <div class="flex items-center gap-2">
              <button
                class="p-1.5 border border-slate-200 dark:border-slate-800 rounded-lg text-slate-400 hover:bg-slate-50 dark:hover:bg-slate-900"
                type="button"
                :disabled="currentPage <= 1 || loading"
                @click="handlePageChange(currentPage - 1)"
              >
                <Icon class="text-sm" icon="material-symbols:chevron-left" />
              </button>
              <button
                v-for="p in pageNumbers"
                :key="p"
                class="w-8 h-8 flex items-center justify-center rounded-lg text-xs font-medium"
                :class="
                  p === currentPage
                    ? 'bg-primary text-white font-bold'
                    : 'hover:bg-slate-50 dark:hover:bg-slate-900 text-slate-700 dark:text-slate-200'
                "
                type="button"
                :disabled="loading"
                @click="handlePageChange(p)"
              >
                {{ p }}
              </button>
              <span v-if="totalPages > 3" class="text-slate-400 px-1">...</span>
              <button
                class="p-1.5 border border-slate-200 dark:border-slate-800 rounded-lg text-slate-400 hover:bg-slate-50 dark:hover:bg-slate-900"
                type="button"
                :disabled="currentPage >= totalPages || loading"
                @click="handlePageChange(currentPage + 1)"
              >
                <Icon class="text-sm" icon="material-symbols:chevron-right" />
              </button>
            </div>
          </div>
        </div>
      </div>
    </main>

    <!-- Create Pickup Code Modal -->
    <Teleport to="body">
      <div
        v-if="showCreateModal"
        class="fixed inset-0 z-50 flex items-center justify-center bg-black/50"
        @click.self="closeCreateModal"
      >
        <CreatePickupCodeModal @close="closeCreateModal" @success="handleCreateSuccess" />
      </div>
    </Teleport>

    <!-- Detail Modal -->
    <Teleport to="body">
      <div
        v-if="showDetailModal && selectedItem"
        class="fixed inset-0 z-50 flex items-center justify-center bg-black/50"
        @click.self="handleCloseDetail"
      >
        <div
          class="bg-white dark:bg-slate-900 rounded-xl shadow-2xl w-full max-w-md mx-4 overflow-hidden"
        >
          <div
            class="px-6 py-4 border-b border-slate-200 dark:border-slate-800 flex items-center justify-between"
          >
            <h3 class="text-lg font-bold text-slate-900 dark:text-white">Pickup Code Details</h3>
            <button
              class="p-2 text-slate-400 hover:text-slate-600 dark:hover:text-slate-300 transition-colors"
              type="button"
              @click="handleCloseDetail"
            >
              <Icon icon="material-symbols:close-rounded" class="text-xl" />
            </button>
          </div>

          <div class="p-6 space-y-4">
            <div class="flex items-center justify-center">
              <div
                class="font-mono text-2xl font-bold text-slate-900 dark:text-white tracking-widest bg-slate-100 dark:bg-slate-800 px-6 py-3 rounded-lg border border-slate-200 dark:border-slate-700"
              >
                {{ selectedItem.code }}
              </div>
            </div>

            <div class="space-y-3">
              <div class="flex justify-between">
                <span class="text-slate-500">File Name</span>
                <span class="font-medium text-slate-900 dark:text-white">{{
                  selectedItem.name
                }}</span>
              </div>
              <div class="flex justify-between">
                <span class="text-slate-500">Type</span>
                <span class="font-medium text-slate-900 dark:text-white">{{
                  selectedItem.type
                }}</span>
              </div>
              <div class="flex justify-between">
                <span class="text-slate-500">Downloads</span>
                <span class="font-medium text-slate-900 dark:text-white"
                  >{{ selectedItem.download }} / {{ selectedItem.max_download }}</span
                >
              </div>
              <div class="flex justify-between">
                <span class="text-slate-500">Expiration</span>
                <span class="font-medium text-slate-900 dark:text-white">{{
                  formatDate(selectedItem.expire_time)
                }}</span>
              </div>
              <div class="flex justify-between">
                <span class="text-slate-500">Status</span>
                <span
                  class="inline-flex items-center px-2.5 py-1 rounded-full text-xs font-bold border"
                  :class="
                    selectedItem.status === 'Active'
                      ? 'bg-primary/10 text-primary border-primary/20'
                      : 'bg-red-100 text-red-600 border-red-200 dark:bg-red-900/20 dark:text-red-400 dark:border-red-800'
                  "
                >
                  {{ selectedItem.status === 'Active' ? 'Active' : 'Expired' }}
                </span>
              </div>
            </div>
          </div>

          <div
            class="px-6 py-4 border-t border-slate-200 dark:border-slate-800 flex justify-end gap-3"
          >
            <button
              type="button"
              class="px-4 py-2 rounded-lg text-sm font-medium text-slate-600 dark:text-slate-300 hover:bg-slate-100 dark:hover:bg-slate-800 transition-colors"
              @click="handleCloseDetail"
            >
              Close
            </button>
            <button
              type="button"
              class="px-6 py-2 rounded-lg text-sm font-medium bg-primary text-white hover:bg-primary/90 transition-colors"
              @click="handleCopyCode(selectedItem.code)"
            >
              Copy Code
            </button>
          </div>
        </div>
      </div>
    </Teleport>

    <ConfirmDialog
      v-model="showDeleteConfirm"
      title="确认删除取件码"
      :message="`将删除取件码「${deleteTargetItem?.code || '-'}」，此操作不可撤销。`"
      confirm-text="确认删除"
      cancel-text="取消"
      :loading="deletingCodeId !== null"
      :danger="true"
      @cancel="closeDeleteConfirm"
      @confirm="confirmDeletePickupCode"
    >
      <template #confirm-icon>
        <Icon
          v-if="deletingCodeId !== null"
          icon="material-symbols:progress-activity"
          class="animate-spin"
        />
      </template>
    </ConfirmDialog>

    <transition
      enter-active-class="transform ease-out duration-300 transition"
      enter-from-class="translate-y-2 opacity-0"
      enter-to-class="translate-y-0 opacity-100"
      leave-active-class="transition ease-in duration-200"
      leave-from-class="opacity-100"
      leave-to-class="opacity-0"
    >
      <div
        v-if="showToast"
        class="fixed top-4 right-4 z-50 flex items-center gap-3 px-6 py-4 rounded-xl shadow-2xl border"
        :class="{
          'bg-green-50 dark:bg-green-900/20 border-green-200 dark:border-green-800':
            toastType === 'success',
          'bg-red-50 dark:bg-red-900/20 border-red-200 dark:border-red-800': toastType === 'error',
          'bg-blue-50 dark:bg-blue-900/20 border-blue-200 dark:border-blue-800':
            toastType === 'info',
        }"
      >
        <Icon
          v-if="toastType === 'success'"
          icon="material-symbols:check-circle"
          class="text-2xl text-green-500"
        />
        <Icon
          v-else-if="toastType === 'error'"
          icon="material-symbols:error"
          class="text-2xl text-red-500"
        />
        <Icon v-else icon="material-symbols:info" class="text-2xl text-blue-500" />
        <span
          class="font-medium"
          :class="{
            'text-green-800 dark:text-green-200': toastType === 'success',
            'text-red-800 dark:text-red-200': toastType === 'error',
            'text-blue-800 dark:text-blue-200': toastType === 'info',
          }"
          >{{ toastMessage }}</span
        >
      </div>
    </transition>

    <Teleport to="body">
      <div
        v-if="openMenuId && menuPosition"
        class="fixed w-48 bg-white dark:bg-slate-950 border border-slate-200 dark:border-slate-800 rounded-xl shadow-xl z-50 py-2"
        :style="{ top: `${menuPosition.top}px`, left: `${menuPosition.left}px` }"
        @click="onStopPropagation"
      >
        <button
          class="w-full flex items-center gap-3 px-4 py-2 text-sm text-slate-700 dark:text-slate-300 hover:bg-slate-50 dark:hover:bg-slate-900"
          type="button"
          @click="handleCopyCode(menuTargetItem?.code || '')"
        >
          <Icon class="text-sm" icon="material-symbols:content-copy" />
          复制取件码
        </button>
        <button
          class="w-full flex items-center gap-3 px-4 py-2 text-sm text-slate-700 dark:text-slate-300 hover:bg-slate-50 dark:hover:bg-slate-900"
          type="button"
          @click="handleViewDetailFromMenu"
        >
          <Icon class="text-sm" icon="material-symbols:visibility" />
          查看详情
        </button>
        <div class="border-t border-slate-100 dark:border-slate-800 my-1"></div>
        <button
          class="w-full flex items-center gap-3 px-4 py-2 text-sm text-red-500 hover:bg-red-50 dark:hover:bg-red-900/20 disabled:opacity-50 disabled:cursor-not-allowed"
          type="button"
          :disabled="deletingCodeId !== null"
          @click="handleDeleteFromMenu"
        >
          <Icon
            class="text-sm"
            :icon="
              deletingCodeId !== null
                ? 'material-symbols:progress-activity'
                : 'material-symbols:delete'
            "
            :class="deletingCodeId !== null ? 'animate-spin' : ''"
          />
          删除
        </button>
      </div>
    </Teleport>
  </div>
</template>

<style lang="sass" scoped></style>
