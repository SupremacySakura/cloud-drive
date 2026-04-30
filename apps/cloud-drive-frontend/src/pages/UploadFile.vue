<script setup lang="ts">
import { Icon } from '@iconify/vue'
import { computed, reactive, ref } from 'vue'
import { getListByFolderIDAndUserID, uploadFile } from '../services/apis/file'
import { detectFileType, formatBytes, iconForFile } from '../utils/file'
import { createId } from '../utils/hash'
import { useUserStore } from '../stores/user'
import LoginRequiredPlaceholder from '../components/bussiness/LoginRequiredPlaceholder.vue'

const userStore = useUserStore()

type QueueItemStatus =
  | 'pending'
  | 'hashing'
  | 'uploading'
  | 'merging'
  | 'success'
  | 'failed'
  | 'canceled'

type QueueItem = {
  id: string
  file: File
  status: QueueItemStatus
  percent: number
  message: string | null
  canceled: boolean
}

const isDragging = ref(false)
const fileInputRef = ref<HTMLInputElement | null>(null)
const items = ref<QueueItem[]>([])
const processingCount = computed(
  () =>
    items.value.filter(i => ['pending', 'hashing', 'uploading', 'merging'].includes(i.status))
      .length,
)
const clampPercent = (value: number) => Math.min(100, Math.max(0, Math.floor(value)))

type BreadcrumbItem = { id: number; name: string }
const selectedUploadFolderId = ref(0)
const selectedUploadFolderPath = ref('root')
const isFolderPickerOpen = ref(false)
const isFolderPickerLoading = ref(false)
const folderPickerCurrentFolderId = ref(0)
const folderPickerBreadcrumbs = ref<BreadcrumbItem[]>([{ id: 0, name: 'root' }])
const folderPickerFolders = ref<{ id: number; name: string }[]>([])
const folderPickerErrorMessage = ref<string | null>(null)

const removeItem = (id: string) => {
  items.value = items.value.filter(x => x.id !== id)
}

const cancelItem = (item: QueueItem) => {
  item.canceled = true
  item.status = 'canceled'
  item.message = '已取消'
}

const retryItem = async (item: QueueItem) => {
  item.status = 'pending'
  item.percent = 0
  item.message = null
  item.canceled = false
  await startUpload(item)
}

const clearCompleted = () => {
  items.value = items.value.filter(x => x.status !== 'success')
}

const openFileDialog = () => {
  fileInputRef.value?.click()
}

const loadFolderPickerFolders = async (folderId: number) => {
  isFolderPickerLoading.value = true
  folderPickerErrorMessage.value = null
  try {
    const list = await getListByFolderIDAndUserID(folderId, 1, 100)
    folderPickerFolders.value = list
      .filter(item => item.type === 'folder')
      .map(item => ({ id: item.id, name: item.name }))
  } catch (error: unknown) {
    folderPickerFolders.value = []
    folderPickerErrorMessage.value = error instanceof Error ? error.message : '加载目录失败'
  } finally {
    isFolderPickerLoading.value = false
  }
}

const openFolderPicker = async () => {
  isFolderPickerOpen.value = true
  folderPickerCurrentFolderId.value = 0
  folderPickerBreadcrumbs.value = [{ id: 0, name: 'root' }]
  await loadFolderPickerFolders(0)
}

const closeFolderPicker = () => {
  isFolderPickerOpen.value = false
}

const goToFolderPickerFolder = async (folder: { id: number; name: string }) => {
  folderPickerCurrentFolderId.value = folder.id
  folderPickerBreadcrumbs.value = [
    ...folderPickerBreadcrumbs.value,
    { id: folder.id, name: folder.name },
  ]
  await loadFolderPickerFolders(folder.id)
}

const goToFolderPickerBreadcrumb = async (index: number) => {
  const next = folderPickerBreadcrumbs.value[index]
  if (!next) return
  folderPickerCurrentFolderId.value = next.id
  folderPickerBreadcrumbs.value = folderPickerBreadcrumbs.value.slice(0, index + 1)
  await loadFolderPickerFolders(next.id)
}

const selectCurrentFolderForUpload = () => {
  selectedUploadFolderId.value = folderPickerCurrentFolderId.value
  selectedUploadFolderPath.value = folderPickerBreadcrumbs.value.map(item => item.name).join(' / ')
  closeFolderPicker()
}

const pushFiles = async (files: File[]) => {
  for (const file of files) {
    const item = reactive<QueueItem>({
      id: createId(),
      file,
      status: 'pending',
      percent: 0,
      message: null,
      canceled: false,
    })
    items.value.unshift(item)
    await startUpload(item)
  }
}

const onFileInputChange = async (e: Event) => {
  const el = e.target as HTMLInputElement
  const fileList = el.files
  if (!fileList?.length) return
  const files = Array.from(fileList)
  el.value = ''
  await pushFiles(files)
}

const onDragOver = (e: DragEvent) => {
  e.preventDefault()
  isDragging.value = true
}

const onDragLeave = () => {
  isDragging.value = false
}

const onDrop = async (e: DragEvent) => {
  e.preventDefault()
  isDragging.value = false
  const fileList = e.dataTransfer?.files
  if (!fileList?.length) return
  await pushFiles(Array.from(fileList))
}

const startUpload = async (item: QueueItem) => {
  try {
    item.status = 'uploading'
    item.percent = 0
    item.message = '上传中...'
    const fileType = detectFileType(item.file)
    await uploadFile(
      item.file,
      {
        file_type: fileType,
        folder_id: selectedUploadFolderId.value,
      },
      progress => {
        if (item.canceled) return
        item.percent = clampPercent(progress)
        item.message = progress >= 100 ? '上传完成' : '上传中...'
      },
    )

    if (item.canceled) return
    item.status = 'success'
    item.percent = 100
    item.message = '上传完成'
  } catch (error: unknown) {
    if (item.canceled) return
    item.status = 'failed'
    item.message = error instanceof Error ? error.message : '上传失败'
  }
}

const badgeText = (item: QueueItem) => {
  if (item.status === 'success') return 'Success'
  if (item.status === 'failed') return 'Failed'
  if (item.status === 'canceled') return 'Canceled'
  if (item.status === 'merging') return 'Merging'
  if (item.status === 'hashing') return 'Hashing'
  if (item.status === 'pending') return 'Pending'
  return `${clampPercent(item.percent)}%`
}

const badgeClass = (item: QueueItem) => {
  if (item.status === 'failed') return 'text-red-500'
  if (item.status === 'success') return 'text-primary'
  if (item.status === 'canceled') return 'text-slate-400'
  return 'text-primary'
}
</script>

<template>
  <div
    class="flex-1 flex flex-col min-w-0 bg-background-light dark:bg-background-dark font-display text-slate-900 dark:text-slate-100"
  >
    <LoginRequiredPlaceholder v-if="!userStore.isLoggedIn" />
    <template v-else>
      <main class="flex-1 overflow-y-auto p-8 space-y-8">
        <section>
          <div
            class="mb-4 flex items-center justify-between gap-3 rounded-xl border border-primary/15 bg-white/70 dark:bg-slate-900/60 px-4 py-3"
          >
            <div class="min-w-0">
              <p class="text-xs text-slate-500 dark:text-slate-400">上传目录</p>
              <p class="text-sm font-semibold text-slate-800 dark:text-slate-100 truncate">
                {{ selectedUploadFolderPath }}
              </p>
            </div>
            <button
              class="shrink-0 inline-flex items-center gap-1.5 rounded-lg border border-primary/30 px-3 py-1.5 text-sm font-semibold text-primary hover:bg-primary/10 transition-colors focus:ring-2 focus:ring-primary/30 focus:outline-none"
              type="button"
              aria-label="选择上传目录"
              @click="openFolderPicker"
            >
              <Icon icon="material-symbols:folder-open-outline" />
              选择目录
            </button>
          </div>
          <input
            ref="fileInputRef"
            type="file"
            class="hidden"
            multiple
            @change="onFileInputChange"
          />
          <div
            class="w-full border-2 border-dashed border-primary/30 rounded-xl bg-primary/5 hover:bg-primary/[0.08] transition-all group flex flex-col items-center justify-center py-16 px-4 cursor-pointer"
            :class="isDragging ? 'ring-2 ring-primary/40 bg-primary/[0.10]' : ''"
            @click="openFileDialog"
            @dragover="onDragOver"
            @dragleave="onDragLeave"
            @drop="onDrop"
          >
            <div
              class="size-16 rounded-full bg-white dark:bg-background-dark shadow-sm flex items-center justify-center text-primary mb-4 group-hover:scale-110 transition-transform"
            >
              <Icon class="text-4xl" icon="material-symbols:upload-file" />
            </div>
            <h3 class="text-xl font-bold text-slate-800 dark:text-slate-100 mb-2">
              拖拽文件到此处上传
            </h3>
            <p class="text-slate-500 dark:text-slate-400 text-sm mb-6 max-w-sm text-center">
              支持图片、视频、文档与压缩包等常见格式。
            </p>
            <button
              class="bg-white dark:bg-slate-800 border border-primary/30 text-primary font-bold px-8 py-2.5 rounded-lg hover:bg-primary hover:text-white transition-all shadow-sm focus:ring-2 focus:ring-primary/30 focus:outline-none"
              type="button"
              aria-label="选择文件"
            >
              选择文件
            </button>
          </div>
        </section>

        <div
          v-if="isFolderPickerOpen"
          class="fixed inset-0 bg-black/50 backdrop-blur-sm z-50 flex items-center justify-center p-4"
          @click="closeFolderPicker"
        >
          <div
            class="w-full max-w-2xl rounded-xl border border-slate-200 dark:border-slate-800 bg-white dark:bg-slate-950 p-6 shadow-xl"
            @click.stop
          >
            <h3 class="text-lg font-bold text-slate-900 dark:text-slate-100">选择上传目录</h3>
            <p class="mt-1 text-sm text-slate-500">默认上传到 root，可切换到任意已有文件夹。</p>

            <div class="mt-4 flex items-center justify-between gap-3">
              <nav class="flex items-center gap-2 overflow-x-auto text-xs text-slate-500">
                <template v-for="(bc, idx) in folderPickerBreadcrumbs" :key="`${bc.id}-${idx}`">
                  <button
                    v-if="idx < folderPickerBreadcrumbs.length - 1"
                    class="whitespace-nowrap hover:text-primary focus:ring-2 focus:ring-primary/30 focus:outline-none rounded"
                    type="button"
                    :aria-label="`导航到 ${bc.name}`"
                    @click="goToFolderPickerBreadcrumb(idx)"
                  >
                    {{ bc.name }}
                  </button>
                  <span
                    v-else
                    class="whitespace-nowrap font-semibold text-slate-900 dark:text-slate-100"
                  >
                    {{ bc.name }}
                  </span>
                  <Icon
                    v-if="idx < folderPickerBreadcrumbs.length - 1"
                    icon="material-symbols:chevron-right"
                    class="text-xs"
                  />
                </template>
              </nav>
              <button
                class="shrink-0 rounded-lg border border-slate-200 px-3 py-1.5 text-xs font-semibold hover:bg-slate-50 dark:border-slate-700 dark:hover:bg-slate-900 focus:ring-2 focus:ring-primary/30 focus:outline-none"
                type="button"
                aria-label="选择当前目录"
                @click="selectCurrentFolderForUpload"
              >
                选择当前目录
              </button>
            </div>

            <div
              class="mt-4 rounded-lg border border-slate-200 dark:border-slate-800 overflow-hidden"
            >
              <div class="max-h-72 overflow-y-auto divide-y divide-slate-100 dark:divide-slate-800">
                <div v-if="isFolderPickerLoading" class="p-4 text-sm text-slate-500">
                  正在加载目录...
                </div>
                <div v-else-if="folderPickerErrorMessage" class="p-4 text-sm text-red-500">
                  {{ folderPickerErrorMessage }}
                </div>
                <button
                  v-for="folder in folderPickerFolders"
                  :key="folder.id"
                  class="w-full flex items-center justify-between px-4 py-3 text-left hover:bg-slate-50 dark:hover:bg-slate-900 focus:ring-2 focus:ring-primary/30 focus:outline-none"
                  type="button"
                  :aria-label="`打开文件夹 ${folder.name}`"
                  @click="goToFolderPickerFolder(folder)"
                >
                  <span class="flex min-w-0 items-center gap-2">
                    <Icon icon="material-symbols:folder" class="shrink-0 text-primary" />
                    <span class="truncate text-sm text-slate-700 dark:text-slate-300">{{
                      folder.name
                    }}</span>
                  </span>
                  <Icon icon="material-symbols:chevron-right" class="text-slate-400" />
                </button>
                <div
                  v-if="
                    !isFolderPickerLoading &&
                    !folderPickerErrorMessage &&
                    folderPickerFolders.length === 0
                  "
                  class="p-4 text-sm text-slate-500"
                >
                  当前目录下没有子文件夹
                </div>
              </div>
            </div>

            <div class="mt-4 flex justify-end gap-3">
              <button
                class="rounded-lg px-4 py-2 text-sm font-medium text-slate-600 hover:bg-slate-100 dark:text-slate-400 dark:hover:bg-slate-900 focus:ring-2 focus:ring-slate-400 focus:outline-none"
                type="button"
                aria-label="取消选择"
                @click="closeFolderPicker"
              >
                取消
              </button>
              <button
                class="rounded-lg bg-primary px-4 py-2 text-sm font-semibold text-white hover:bg-primary/90 focus:ring-2 focus:ring-primary/50 focus:outline-none"
                type="button"
                aria-label="确认选择当前目录"
                @click="selectCurrentFolderForUpload"
              >
                确认
              </button>
            </div>
          </div>
        </div>

        <section class="space-y-4">
          <div class="flex items-center justify-between">
            <h3 class="text-lg font-bold text-slate-800 dark:text-slate-100">上传队列</h3>
            <div class="flex items-center gap-4">
              <span class="text-sm text-slate-500">{{ processingCount }} 项处理中</span>
              <button
                class="text-sm font-bold text-primary hover:underline focus:ring-2 focus:ring-primary/30 focus:outline-none rounded px-1"
                type="button"
                aria-label="清理已完成的上传任务"
                @click="clearCompleted"
              >
                清理已完成
              </button>
            </div>
          </div>

          <div
            class="bg-white dark:bg-slate-900 border border-primary/10 rounded-xl overflow-hidden shadow-sm"
          >
            <div
              v-if="items.length === 0"
              class="p-8 text-center text-slate-500 dark:text-slate-400"
            >
              暂无上传任务
            </div>

            <div
              v-for="item in items"
              :key="item.id"
              class="flex items-center p-4 border-b border-primary/5 last:border-0 hover:bg-slate-50 dark:hover:bg-slate-800/50 transition-colors"
            >
              <div
                class="size-12 rounded-lg flex items-center justify-center shrink-0"
                :class="`${iconForFile(item.file).bg} ${iconForFile(item.file).fg}`"
              >
                <Icon :icon="iconForFile(item.file).icon" />
              </div>

              <div class="ml-4 flex-1 min-w-0">
                <div class="flex items-center justify-between mb-1">
                  <p class="text-sm font-bold text-slate-800 dark:text-slate-100 truncate">
                    {{ item.file.name }}
                  </p>
                  <span
                    class="text-xs font-bold uppercase tracking-wider"
                    :class="badgeClass(item)"
                  >
                    {{ badgeText(item) }}
                  </span>
                </div>

                <div class="flex items-center gap-3">
                  <div class="flex-1 bg-primary/10 h-1.5 rounded-full overflow-hidden">
                    <div
                      class="h-1.5 rounded-full"
                      :class="item.status === 'failed' ? 'bg-red-500' : 'bg-primary'"
                      :style="{ width: `${clampPercent(item.percent)}%` }"
                    ></div>
                  </div>
                  <span class="text-xs text-slate-500 whitespace-nowrap">
                    {{ formatBytes(item.file.size) }} • {{ item.message || '等待中' }}
                  </span>
                </div>
              </div>

              <div class="ml-6 flex items-center gap-2">
                <button
                  v-if="item.status === 'failed'"
                  class="flex items-center gap-1.5 bg-primary/10 hover:bg-primary/20 text-primary font-bold px-3 py-1.5 rounded-lg text-xs transition-colors focus:ring-2 focus:ring-primary/30 focus:outline-none"
                  type="button"
                  aria-label="重试上传"
                  @click="retryItem(item)"
                >
                  <Icon class="text-sm" icon="material-symbols:replay" />
                  重试
                </button>

                <div
                  v-if="item.status === 'success'"
                  class="p-2 text-primary"
                  aria-label="上传成功"
                >
                  <Icon class="text-xl" icon="material-symbols:check-circle" />
                </div>

                <button
                  v-if="
                    item.status === 'uploading' ||
                    item.status === 'hashing' ||
                    item.status === 'merging'
                  "
                  class="p-2 text-slate-400 hover:text-red-500 rounded-lg transition-colors focus:ring-2 focus:ring-red-400 focus:outline-none"
                  type="button"
                  aria-label="取消上传"
                  @click="cancelItem(item)"
                >
                  <Icon class="text-xl" icon="material-symbols:close" />
                </button>

                <button
                  v-else
                  class="p-2 text-slate-400 hover:text-red-500 rounded-lg transition-colors focus:ring-2 focus:ring-red-400 focus:outline-none"
                  type="button"
                  aria-label="移除任务"
                  @click="removeItem(item.id)"
                >
                  <Icon class="text-xl" icon="material-symbols:close" />
                </button>
              </div>
            </div>
          </div>
        </section>
      </main>
    </template>
  </div>
</template>

<style lang="sass" scoped></style>
