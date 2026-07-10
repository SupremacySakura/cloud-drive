<script setup lang="ts">
import { computed } from 'vue'
import { Icon } from '@iconify/vue'
import type { UploadTask } from '../../types/file'
import { iconForFile } from '../../utils/file'
import { uploadTaskMeta } from '../../utils/file-management'

const props = defineProps<{
  modelValue: boolean
  tasks: readonly UploadTask[]
  isUploading: boolean
  overallProgress: number
}>()

const emit = defineEmits<{
  'update:modelValue': [value: boolean]
  close: []
  'clear-completed': []
  retry: [task: UploadTask]
  cancel: [task: UploadTask]
  remove: [taskId: string]
}>()

const completedCount = computed(() => props.tasks.filter(task => task.status === 'success').length)
const hasCompletedTasks = computed(() => completedCount.value > 0)
const normalizedOverallProgress = computed(() =>
  Math.min(100, Math.max(0, Math.floor(props.overallProgress))),
)
const taskRows = computed(() =>
  props.tasks.map(task => {
    const meta = uploadTaskMeta(task.status)
    const percent = Math.min(100, Math.max(0, Math.floor(task.percent)))
    return {
      task,
      icon: iconForFile(task.file),
      meta,
      percent,
      statusLabel: meta.label ?? `${percent}%`,
      removable: ['success', 'failed', 'canceled'].includes(task.status),
    }
  }),
)

const closePanel = () => {
  emit('update:modelValue', false)
  emit('close')
}
</script>

<template>
  <div
    v-if="modelValue && tasks.length > 0"
    class="fixed bottom-4 right-4 z-40 flex max-h-[500px] w-[calc(100vw-2rem)] flex-col rounded-xl border border-slate-200 bg-white shadow-2xl dark:border-slate-800 dark:bg-slate-950 sm:w-96"
  >
    <div
      class="flex items-center justify-between border-b border-slate-200 p-4 dark:border-slate-800"
    >
      <div class="flex items-center gap-2">
        <Icon
          v-if="isUploading"
          icon="material-symbols:progress-activity"
          class="animate-spin text-primary"
        />
        <Icon v-else icon="material-symbols:check-circle" class="text-green-500" />
        <span class="font-semibold text-slate-900 dark:text-slate-100">
          上传任务 ({{ completedCount }}/{{ tasks.length }})
        </span>
      </div>
      <div class="flex items-center gap-1">
        <button
          v-if="hasCompletedTasks"
          class="rounded-lg p-1.5 text-slate-400 hover:bg-slate-100 hover:text-slate-600 focus:outline-none focus:ring-2 focus:ring-primary/30 dark:hover:bg-slate-900 dark:hover:text-slate-200"
          type="button"
          aria-label="清理已完成的上传任务"
          title="清理已完成"
          @click="emit('clear-completed')"
        >
          <Icon icon="material-symbols:cleaning-services" />
        </button>
        <button
          class="rounded-lg p-1.5 text-slate-400 hover:bg-slate-100 hover:text-slate-600 focus:outline-none focus:ring-2 focus:ring-primary/30 dark:hover:bg-slate-900 dark:hover:text-slate-200"
          type="button"
          aria-label="关闭上传面板"
          title="关闭面板"
          @click="closePanel"
        >
          <Icon icon="material-symbols:close" />
        </button>
      </div>
    </div>

    <div
      class="border-b border-slate-200 bg-slate-50 px-4 py-2 dark:border-slate-800 dark:bg-slate-900/50"
    >
      <div class="mb-1 flex items-center justify-between text-xs">
        <span class="text-slate-500">总体进度</span>
        <span class="font-medium text-slate-700 dark:text-slate-300">
          {{ normalizedOverallProgress }}%
        </span>
      </div>
      <div class="h-1.5 overflow-hidden rounded-full bg-slate-200 dark:bg-slate-800">
        <div
          class="h-1.5 rounded-full bg-primary transition-all duration-300"
          :style="{ width: `${normalizedOverallProgress}%` }"
        ></div>
      </div>
    </div>

    <div class="max-h-[350px] flex-1 space-y-2 overflow-y-auto p-2">
      <div
        v-for="row in taskRows"
        :key="row.task.id"
        class="flex items-center gap-3 rounded-lg border border-slate-100 bg-slate-50/50 p-3 dark:border-slate-800 dark:bg-slate-900/30"
      >
        <div
          class="flex h-8 w-8 shrink-0 items-center justify-center rounded-lg"
          :class="`${row.icon.bg} ${row.icon.fg}`"
        >
          <Icon :icon="row.icon.icon" class="text-sm" />
        </div>

        <div class="min-w-0 flex-1">
          <div class="mb-1 flex items-center justify-between">
            <p class="truncate text-xs font-medium text-slate-800 dark:text-slate-100">
              {{ row.task.file.name }}
            </p>
            <span class="text-[10px] font-medium uppercase tracking-wider" :class="row.meta.color">
              {{ row.statusLabel }}
            </span>
          </div>
          <div class="h-1 overflow-hidden rounded-full bg-slate-200 dark:bg-slate-800">
            <div
              class="h-1 rounded-full transition-all duration-200"
              :class="row.meta.bar"
              :style="{ width: `${row.percent}%` }"
            ></div>
          </div>
          <p class="mt-1 truncate text-[10px] text-slate-400">{{ row.task.message }}</p>
        </div>

        <div class="flex items-center gap-1">
          <button
            v-if="row.task.status === 'failed'"
            class="rounded p-1 text-slate-400 transition-colors hover:text-primary focus:outline-none focus:ring-2 focus:ring-primary/30"
            type="button"
            aria-label="重试上传"
            @click="emit('retry', row.task)"
          >
            <Icon icon="material-symbols:replay" class="text-sm" />
          </button>
          <button
            v-if="row.meta.active"
            class="rounded p-1 text-slate-400 transition-colors hover:text-red-500 focus:outline-none focus:ring-2 focus:ring-red-400"
            type="button"
            aria-label="取消上传"
            @click="emit('cancel', row.task)"
          >
            <Icon icon="material-symbols:close" class="text-sm" />
          </button>
          <button
            v-if="row.removable"
            class="rounded p-1 text-slate-400 transition-colors hover:text-red-500 focus:outline-none focus:ring-2 focus:ring-red-400"
            type="button"
            aria-label="移除任务"
            @click="emit('remove', row.task.id)"
          >
            <Icon icon="material-symbols:delete" class="text-sm" />
          </button>
        </div>
      </div>
    </div>
  </div>
</template>
