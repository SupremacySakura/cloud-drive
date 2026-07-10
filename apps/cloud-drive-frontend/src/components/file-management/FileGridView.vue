<script setup lang="ts">
import { computed } from 'vue'
import { Icon } from '@iconify/vue'
import type { FileDisplayItem, FileItemKey } from '../../types/file'
import { formatBytes } from '../../utils/file'
import { ownerInitials } from '../../utils/file-management'
import FileEmptyState from './FileEmptyState.vue'

const props = withDefaults(
  defineProps<{
    files: readonly FileDisplayItem[]
    loading: boolean
    hasParentFolder: boolean
    selectedIds: ReadonlySet<FileItemKey>
    query?: string
    ownerName?: string
  }>(),
  {
    query: '',
    ownerName: 'Me',
  },
)

const emit = defineEmits<{
  'go-parent': []
  'open-item': [file: FileDisplayItem]
  'toggle-select': [file: FileDisplayItem, checked: boolean]
  'open-menu': [file: FileDisplayItem, anchor: HTMLElement]
  upload: []
}>()

const ownerBadge = computed(() => ownerInitials(props.ownerName))

const handleItemClick = (file: FileDisplayItem) => emit('open-item', file)
const handleToggleSelect = (file: FileDisplayItem, event: Event) => {
  const input = event.currentTarget
  if (input instanceof HTMLInputElement) emit('toggle-select', file, input.checked)
}
const handleOpenMenu = (file: FileDisplayItem, event: MouseEvent) => {
  const anchor = event.currentTarget
  if (anchor instanceof HTMLElement) emit('open-menu', file, anchor)
}
</script>

<template>
  <div>
    <div class="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3">
      <div
        v-if="hasParentFolder"
        class="cursor-pointer rounded-xl border border-slate-200 bg-white p-4 shadow-sm transition-colors hover:bg-slate-50 dark:border-slate-800 dark:bg-slate-950 dark:hover:bg-slate-900/40"
        role="button"
        tabindex="0"
        @click="emit('go-parent')"
        @keyup.enter.self="emit('go-parent')"
      >
        <div class="flex items-start justify-between gap-3">
          <div class="flex min-w-0 items-center gap-3">
            <div
              class="flex h-11 w-11 shrink-0 items-center justify-center rounded-lg bg-amber-100 text-amber-700 dark:bg-amber-900/30 dark:text-amber-300"
            >
              <Icon class="text-[22px]" icon="material-symbols:folder" />
            </div>
            <div class="min-w-0">
              <div class="truncate text-sm font-semibold text-slate-900 dark:text-slate-100">
                ..
              </div>
              <div class="text-xs text-slate-500">返回上一级目录</div>
            </div>
          </div>
        </div>
      </div>

      <div
        v-for="file in files"
        :key="file.key"
        class="rounded-xl border border-slate-200 bg-white p-4 shadow-sm transition-colors hover:bg-slate-50 dark:border-slate-800 dark:bg-slate-950 dark:hover:bg-slate-900/40"
        :class="file.type === 'folder' ? 'cursor-pointer' : ''"
        :role="file.type === 'folder' ? 'button' : undefined"
        :tabindex="file.type === 'folder' ? 0 : undefined"
        @click="handleItemClick(file)"
        @keyup.enter.self="handleItemClick(file)"
      >
        <div class="flex items-start justify-between gap-3">
          <div class="flex min-w-0 items-center gap-3">
            <div
              class="flex h-11 w-11 shrink-0 items-center justify-center rounded-lg"
              :class="`${file.iconBg} ${file.iconFg}`"
            >
              <Icon class="text-[22px]" :icon="file.icon" />
            </div>
            <div class="min-w-0">
              <div class="truncate text-sm font-semibold text-slate-900 dark:text-slate-100">
                {{ file.name }}
              </div>
              <div class="text-xs text-slate-500">{{ file.typeLabel }}</div>
            </div>
          </div>
          <input
            class="rounded border-slate-300 text-primary focus:ring-primary/20"
            type="checkbox"
            :aria-label="`选择 ${file.name}`"
            :checked="selectedIds.has(file.key)"
            @click.stop
            @change="handleToggleSelect(file, $event)"
          />
        </div>

        <div class="mt-4 flex items-center justify-between text-xs text-slate-500">
          <span>{{ file.type === 'folder' ? '-' : formatBytes(file.size) }}</span>
          <span class="whitespace-nowrap">{{ file.lastModifiedText }}</span>
        </div>

        <div class="mt-3 flex items-center justify-between">
          <div class="flex items-center gap-2">
            <div
              class="flex h-6 w-6 items-center justify-center rounded-full bg-slate-200 text-[10px] font-bold text-slate-600 dark:bg-slate-800 dark:text-slate-300"
            >
              {{ ownerBadge }}
            </div>
            <span class="text-xs text-slate-500">{{ ownerName }}</span>
          </div>
          <button
            class="rounded-lg p-2 text-slate-400 hover:bg-white/50 hover:text-slate-700 focus:outline-none focus:ring-2 focus:ring-primary/30 dark:hover:bg-slate-900 dark:hover:text-slate-200"
            type="button"
            :aria-label="`${file.name} 操作菜单`"
            @click.stop="handleOpenMenu(file, $event)"
          >
            <Icon icon="material-symbols:more-vert" />
          </button>
        </div>
      </div>

      <div v-if="loading" class="col-span-full py-20">
        <div class="flex flex-col items-center justify-center text-center">
          <Icon
            icon="material-symbols:progress-activity"
            class="mb-3 animate-spin text-4xl text-primary"
          />
          <p class="text-sm text-slate-400">加载中...</p>
        </div>
      </div>
      <div v-else-if="files.length === 0" class="col-span-full py-20">
        <FileEmptyState :query="query" @upload="emit('upload')" />
      </div>
    </div>
    <slot name="footer" />
  </div>
</template>
