<script setup lang="ts">
import { computed, ref } from 'vue'
import { Icon } from '@iconify/vue'
import type { BreadcrumbItem } from '../../types/file'

const props = withDefaults(
  defineProps<{
    breadcrumbs: readonly BreadcrumbItem[]
    currentFolderName: string
    errorMessage?: string | null
    uploadTaskCount?: number
    completedUploadCount?: number
    isUploading?: boolean
    overallProgress?: number
    isUploadPanelOpen?: boolean
  }>(),
  {
    errorMessage: null,
    uploadTaskCount: 0,
    completedUploadCount: 0,
    isUploading: false,
    overallProgress: 0,
    isUploadPanelOpen: false,
  },
)

const emit = defineEmits<{
  navigate: [index: number]
  'create-folder': []
  'files-selected': [files: File[]]
  'folder-selected': [files: File[]]
  'toggle-upload-panel': []
}>()

const fileInputRef = ref<HTMLInputElement | null>(null)
const folderInputRef = ref<HTMLInputElement | null>(null)

const nestedBreadcrumbs = computed(() => props.breadcrumbs.slice(1))
const normalizedProgress = computed(() => Math.min(100, Math.max(0, props.overallProgress)))
const uploadStatusIcon = computed(() =>
  props.isUploading ? 'material-symbols:progress-activity' : 'material-symbols:check-circle',
)

const navigateToBreadcrumb = (index: number) => emit('navigate', index)
const isLastBreadcrumb = (nestedIndex: number) => nestedIndex === nestedBreadcrumbs.value.length - 1
const openFileDialog = () => fileInputRef.value?.click()
const openFolderDialog = () => folderInputRef.value?.click()

const takeFilesFromInput = (event: Event): File[] => {
  const input = event.currentTarget
  if (!(input instanceof HTMLInputElement) || !input.files?.length) return []
  const files = Array.from(input.files)
  input.value = ''
  return files
}

const handleFileInput = (event: Event) => {
  const files = takeFilesFromInput(event)
  if (files.length > 0) emit('files-selected', files)
}

const handleFolderInput = (event: Event) => {
  const files = takeFilesFromInput(event)
  if (files.length > 0) emit('folder-selected', files)
}

defineExpose({ openFileDialog, openFolderDialog })
</script>

<template>
  <div class="mb-6 flex flex-col gap-4 lg:mb-8 lg:flex-row lg:items-center lg:justify-between">
    <div>
      <nav class="mb-2 flex items-center gap-2 overflow-x-auto text-sm text-slate-500">
        <button
          class="flex items-center rounded hover:text-primary focus:outline-none focus:ring-2 focus:ring-primary/30"
          type="button"
          aria-label="返回根目录"
          @click="navigateToBreadcrumb(0)"
        >
          <Icon class="mr-1 text-sm" icon="material-symbols:home" />
          root
        </button>
        <template v-for="(breadcrumb, index) in nestedBreadcrumbs" :key="breadcrumb.id">
          <Icon class="text-sm" icon="material-symbols:chevron-right" />
          <button
            v-if="!isLastBreadcrumb(index)"
            class="flex items-center rounded hover:text-primary focus:outline-none focus:ring-2 focus:ring-primary/30"
            type="button"
            :aria-label="`导航到 ${breadcrumb.name} 文件夹`"
            @click="navigateToBreadcrumb(index + 1)"
          >
            {{ breadcrumb.name }}
          </button>
          <span v-else class="font-medium text-slate-900 dark:text-slate-100">
            {{ breadcrumb.name }}
          </span>
        </template>
      </nav>
      <h2 class="text-2xl font-bold text-slate-900 dark:text-slate-100">
        {{ currentFolderName }}
      </h2>
      <p v-if="errorMessage" class="mt-2 text-sm text-red-500">{{ errorMessage }}</p>
    </div>

    <div class="flex flex-wrap items-center gap-3">
      <input ref="fileInputRef" type="file" class="hidden" multiple @change="handleFileInput" />
      <input
        ref="folderInputRef"
        type="file"
        class="hidden"
        webkitdirectory
        directory
        @change="handleFolderInput"
      />

      <button
        class="flex items-center gap-2 rounded-lg border border-slate-200 bg-white px-4 py-2 text-sm font-semibold text-slate-700 transition-all hover:bg-slate-50 focus:outline-none focus:ring-2 focus:ring-primary/30 dark:border-slate-800 dark:bg-slate-950 dark:text-slate-200 dark:hover:bg-slate-900"
        type="button"
        aria-label="新建文件夹"
        @click="emit('create-folder')"
      >
        <Icon class="text-[20px]" icon="material-symbols:create-new-folder" />
        新建文件夹
      </button>

      <div class="flex flex-wrap items-center gap-3 sm:hidden">
        <button
          class="flex items-center gap-2 rounded-lg bg-primary px-4 py-2 text-sm font-bold text-white shadow-lg shadow-primary/20 transition-all hover:bg-primary/90 focus:outline-none focus:ring-2 focus:ring-primary/50"
          type="button"
          aria-label="上传文件"
          @click="openFileDialog"
        >
          <Icon class="text-[20px]" icon="material-symbols:upload" />
          上传文件
        </button>
        <button
          class="flex items-center gap-2 rounded-lg border border-slate-200 bg-white px-4 py-2 text-sm font-semibold text-slate-700 transition-all hover:bg-slate-50 focus:outline-none focus:ring-2 focus:ring-primary/30 dark:border-slate-800 dark:bg-slate-950 dark:text-slate-200 dark:hover:bg-slate-900"
          type="button"
          aria-label="上传文件夹"
          @click="openFolderDialog"
        >
          <Icon icon="material-symbols:folder" />
          上传文件夹
        </button>
      </div>

      <div class="group relative hidden sm:block">
        <button
          class="flex items-center gap-2 rounded-lg bg-primary px-6 py-2 text-sm font-bold text-white shadow-lg shadow-primary/20 transition-all hover:bg-primary/90 focus:outline-none focus:ring-2 focus:ring-primary/50"
          type="button"
          aria-label="上传文件或文件夹"
        >
          <Icon class="text-[20px]" icon="material-symbols:upload" />
          上传
        </button>
        <div
          class="invisible absolute right-0 top-full z-20 mt-2 w-48 rounded-xl border border-slate-200 bg-white py-2 opacity-0 shadow-xl transition-all group-hover:visible group-hover:opacity-100 dark:border-slate-800 dark:bg-slate-950"
        >
          <button
            class="flex w-full items-center gap-3 px-4 py-2 text-sm text-slate-700 hover:bg-slate-50 focus:outline-none focus:ring-2 focus:ring-primary/30 dark:text-slate-300 dark:hover:bg-slate-900"
            type="button"
            aria-label="上传文件"
            @click="openFileDialog"
          >
            <Icon icon="material-symbols:description" />
            上传文件
          </button>
          <button
            class="flex w-full items-center gap-3 px-4 py-2 text-sm text-slate-700 hover:bg-slate-50 focus:outline-none focus:ring-2 focus:ring-primary/30 dark:text-slate-300 dark:hover:bg-slate-900"
            type="button"
            aria-label="上传文件夹"
            @click="openFolderDialog"
          >
            <Icon icon="material-symbols:folder" />
            上传文件夹
          </button>
        </div>
      </div>

      <button
        v-if="uploadTaskCount > 0"
        class="flex items-center gap-2 rounded-lg border border-slate-200 bg-white px-4 py-2 text-sm font-semibold text-slate-700 transition-all hover:bg-slate-50 focus:outline-none focus:ring-2 focus:ring-primary/30 dark:border-slate-800 dark:bg-slate-950 dark:text-slate-200 dark:hover:bg-slate-900"
        :class="isUploadPanelOpen ? 'bg-slate-50 dark:bg-slate-900' : ''"
        type="button"
        aria-label="查看上传进度"
        @click="emit('toggle-upload-panel')"
      >
        <Icon
          class="text-[20px]"
          :icon="uploadStatusIcon"
          :class="isUploading ? 'animate-spin text-primary' : 'text-green-500'"
        />
        <span v-if="isUploading">{{ normalizedProgress }}%</span>
        <span v-else>{{ completedUploadCount }}/{{ uploadTaskCount }}</span>
      </button>
    </div>
  </div>
</template>
