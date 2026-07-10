<script setup lang="ts">
import { nextTick, ref, useId, watch } from 'vue'
import { Icon } from '@iconify/vue'
import type { FileListItem } from '../../services/types/file'
import type { BreadcrumbItem, FileDisplayItem } from '../../types/file'

const props = withDefaults(
  defineProps<{
    modelValue: boolean
    target: FileDisplayItem | null
    breadcrumbs: readonly BreadcrumbItem[]
    folders: readonly FileListItem[]
    browserLoading: boolean
    targetFolderId: number
    moving?: boolean
  }>(),
  {
    moving: false,
  },
)

const emit = defineEmits<{
  'update:modelValue': [value: boolean]
  cancel: []
  navigate: [index: number]
  'open-folder': [folder: FileListItem]
  'select-current': []
  confirm: []
}>()

const dialogRef = ref<HTMLElement | null>(null)
const titleId = `move-file-dialog-title-${useId()}`

const handleCancel = () => {
  if (props.moving) return
  emit('update:modelValue', false)
  emit('cancel')
}
const isLastBreadcrumb = (index: number) => index === props.breadcrumbs.length - 1
const isOwnFolder = (folder: FileListItem) =>
  props.target?.type === 'folder' && props.target.key === `folder:${folder.id}`

watch(
  () => props.modelValue,
  async visible => {
    if (!visible) return
    await nextTick()
    dialogRef.value?.focus()
  },
)
</script>

<template>
  <Teleport to="body">
    <div
      v-if="modelValue"
      class="fixed inset-0 z-50 flex items-center justify-center bg-black/50 p-4 backdrop-blur-sm"
      role="dialog"
      aria-modal="true"
      :aria-labelledby="titleId"
      @click.self="handleCancel"
      @keydown.esc="handleCancel"
    >
      <div
        ref="dialogRef"
        class="w-full max-w-2xl rounded-xl border border-slate-200 bg-white p-6 shadow-xl outline-none dark:border-slate-800 dark:bg-slate-950"
        tabindex="-1"
      >
        <h3 :id="titleId" class="mb-2 text-lg font-bold text-slate-900 dark:text-slate-100">
          移动到
        </h3>
        <p class="mb-4 text-sm text-slate-500">当前对象：{{ target?.name ?? '-' }}</p>

        <div class="mb-3 flex items-center justify-between gap-3">
          <nav class="flex items-center gap-2 overflow-x-auto text-xs text-slate-500">
            <template v-for="(breadcrumb, index) in breadcrumbs" :key="`${breadcrumb.id}-${index}`">
              <button
                v-if="!isLastBreadcrumb(index)"
                class="whitespace-nowrap hover:text-primary focus:outline-none focus:ring-2 focus:ring-primary/30"
                type="button"
                @click="emit('navigate', index)"
              >
                {{ breadcrumb.name }}
              </button>
              <span
                v-else
                class="whitespace-nowrap font-semibold text-slate-900 dark:text-slate-100"
              >
                {{ breadcrumb.name }}
              </span>
              <Icon
                v-if="!isLastBreadcrumb(index)"
                icon="material-symbols:chevron-right"
                class="text-xs"
              />
            </template>
          </nav>
          <button
            class="rounded-lg border border-slate-200 px-3 py-1.5 text-xs font-semibold hover:bg-slate-50 focus:outline-none focus:ring-2 focus:ring-primary/30 dark:border-slate-700 dark:hover:bg-slate-900"
            type="button"
            aria-label="选中当前目录作为目标"
            :disabled="moving"
            @click="emit('select-current')"
          >
            选中当前目录
          </button>
        </div>

        <div class="overflow-hidden rounded-lg border border-slate-200 dark:border-slate-800">
          <div class="max-h-72 divide-y divide-slate-100 overflow-y-auto dark:divide-slate-800">
            <div v-if="browserLoading" class="p-4 text-sm text-slate-500">正在加载目录...</div>
            <button
              v-for="folder in folders"
              :key="`folder:${folder.id}`"
              class="flex w-full items-center justify-between px-4 py-3 text-left hover:bg-slate-50 focus:outline-none focus:ring-2 focus:ring-primary/30 disabled:cursor-not-allowed disabled:opacity-50 dark:hover:bg-slate-900"
              type="button"
              :disabled="isOwnFolder(folder) || moving"
              @click="emit('open-folder', folder)"
            >
              <span class="flex min-w-0 items-center gap-2">
                <Icon icon="material-symbols:folder" class="shrink-0 text-primary" />
                <span class="truncate text-sm text-slate-700 dark:text-slate-300">
                  {{ folder.name }}
                </span>
              </span>
              <Icon icon="material-symbols:chevron-right" class="text-slate-400" />
            </button>
            <div v-if="!browserLoading && folders.length === 0" class="p-4 text-sm text-slate-500">
              当前目录下没有子文件夹
            </div>
          </div>
        </div>

        <p class="mt-3 text-xs text-slate-500">
          已选目标目录 ID：
          <span class="font-semibold text-slate-700 dark:text-slate-300">
            {{ targetFolderId }}
          </span>
        </p>

        <div class="mt-4 flex justify-end gap-3">
          <button
            class="rounded-lg px-4 py-2 text-sm font-medium text-slate-600 transition-colors hover:bg-slate-100 focus:outline-none focus:ring-2 focus:ring-slate-400 disabled:cursor-not-allowed disabled:opacity-60 dark:text-slate-400 dark:hover:bg-slate-900"
            type="button"
            :disabled="moving"
            @click="handleCancel"
          >
            取消
          </button>
          <button
            class="flex items-center gap-2 rounded-lg bg-primary px-4 py-2 text-sm font-medium text-white transition-colors hover:bg-primary/90 focus:outline-none focus:ring-2 focus:ring-primary/50 disabled:cursor-not-allowed disabled:opacity-60"
            type="button"
            :disabled="moving"
            @click="emit('confirm')"
          >
            <Icon v-if="moving" icon="material-symbols:progress-activity" class="animate-spin" />
            {{ moving ? '移动中...' : '确认移动' }}
          </button>
        </div>
      </div>
    </div>
  </Teleport>
</template>
