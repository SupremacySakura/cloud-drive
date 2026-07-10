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
    allSelected: boolean
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
  'toggle-all': [checked: boolean]
  'toggle-select': [file: FileDisplayItem, checked: boolean]
  'open-menu': [file: FileDisplayItem, anchor: HTMLElement]
  upload: []
}>()

const ownerBadge = computed(() => ownerInitials(props.ownerName))

const handleItemClick = (file: FileDisplayItem) => emit('open-item', file)
const handleToggleAll = (event: Event) => {
  const input = event.currentTarget
  if (input instanceof HTMLInputElement) emit('toggle-all', input.checked)
}
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
  <div
    class="overflow-hidden rounded-xl border border-slate-200 bg-white shadow-sm dark:border-slate-800 dark:bg-slate-950"
  >
    <div class="overflow-x-auto">
      <table class="w-full min-w-[720px] border-collapse text-left">
        <thead
          class="border-b border-slate-200 bg-slate-50 dark:border-slate-800 dark:bg-slate-900/50"
        >
          <tr>
            <th scope="col" class="w-10 px-6 py-4">
              <input
                class="rounded border-slate-300 text-primary focus:ring-primary/20"
                type="checkbox"
                aria-label="选择当前页全部文件"
                :checked="allSelected"
                @change="handleToggleAll"
              />
            </th>
            <th
              scope="col"
              class="px-4 py-4 text-xs font-bold uppercase tracking-wider text-slate-500"
            >
              Name
            </th>
            <th
              scope="col"
              class="hidden px-4 py-4 text-xs font-bold uppercase tracking-wider text-slate-500 md:table-cell"
            >
              Type
            </th>
            <th
              scope="col"
              class="hidden px-4 py-4 text-xs font-bold uppercase tracking-wider text-slate-500 sm:table-cell"
            >
              Size
            </th>
            <th
              scope="col"
              class="hidden px-4 py-4 text-xs font-bold uppercase tracking-wider text-slate-500 lg:table-cell"
            >
              Owner
            </th>
            <th
              scope="col"
              class="px-4 py-4 text-xs font-bold uppercase tracking-wider text-slate-500"
            >
              Last Modified
            </th>
            <th
              scope="col"
              class="px-6 py-4 text-right text-xs font-bold uppercase tracking-wider text-slate-500"
            >
              Actions
            </th>
          </tr>
        </thead>

        <tbody class="divide-y divide-slate-100 dark:divide-slate-800">
          <tr
            v-if="hasParentFolder"
            class="group cursor-pointer transition-colors hover:bg-slate-50 dark:hover:bg-slate-900/40"
            role="button"
            tabindex="0"
            @click="emit('go-parent')"
            @keyup.enter.self="emit('go-parent')"
          >
            <td class="px-6 py-4"></td>
            <td class="px-4 py-4">
              <div class="flex items-center gap-3">
                <div
                  class="flex h-10 w-10 items-center justify-center rounded bg-amber-100 text-amber-700 dark:bg-amber-900/30 dark:text-amber-300"
                >
                  <Icon icon="material-symbols:folder" />
                </div>
                <span
                  class="max-w-[150px] truncate text-sm font-semibold text-slate-900 dark:text-slate-100 md:max-w-xs"
                >
                  ..
                </span>
              </div>
            </td>
            <td class="hidden px-4 py-4 text-sm text-slate-500 md:table-cell">文件夹</td>
            <td class="hidden px-4 py-4 text-sm text-slate-500 sm:table-cell">-</td>
            <td class="hidden px-4 py-4 text-sm text-slate-500 lg:table-cell">-</td>
            <td class="whitespace-nowrap px-4 py-4 text-sm text-slate-500">返回上一级目录</td>
            <td class="px-6 py-4"></td>
          </tr>

          <tr
            v-for="file in files"
            :key="file.key"
            class="group transition-colors hover:bg-slate-50 dark:hover:bg-slate-900/40"
            :class="file.type === 'folder' ? 'cursor-pointer' : ''"
            :role="file.type === 'folder' ? 'button' : undefined"
            :tabindex="file.type === 'folder' ? 0 : undefined"
            @click="handleItemClick(file)"
            @keyup.enter.self="handleItemClick(file)"
          >
            <td class="px-6 py-4">
              <input
                class="rounded border-slate-300 text-primary focus:ring-primary/20"
                type="checkbox"
                :aria-label="`选择 ${file.name}`"
                :checked="selectedIds.has(file.key)"
                @click.stop
                @change="handleToggleSelect(file, $event)"
              />
            </td>
            <td class="px-4 py-4">
              <div class="flex items-center gap-3">
                <div
                  class="flex h-10 w-10 items-center justify-center rounded"
                  :class="`${file.iconBg} ${file.iconFg}`"
                >
                  <Icon :icon="file.icon" />
                </div>
                <span
                  class="max-w-[150px] truncate text-sm font-semibold text-slate-900 dark:text-slate-100 md:max-w-xs"
                >
                  {{ file.name }}
                </span>
              </div>
            </td>
            <td class="hidden px-4 py-4 text-sm text-slate-500 md:table-cell">
              {{ file.typeLabel }}
            </td>
            <td class="hidden px-4 py-4 text-sm text-slate-500 sm:table-cell">
              {{ file.type === 'folder' ? '-' : formatBytes(file.size) }}
            </td>
            <td class="hidden px-4 py-4 lg:table-cell">
              <div class="flex items-center gap-2">
                <div
                  class="flex h-6 w-6 items-center justify-center rounded-full bg-slate-200 text-[10px] font-bold text-slate-600 dark:bg-slate-800 dark:text-slate-300"
                >
                  {{ ownerBadge }}
                </div>
                <span class="text-sm text-slate-600 dark:text-slate-400">{{ ownerName }}</span>
              </div>
            </td>
            <td class="whitespace-nowrap px-4 py-4 text-sm text-slate-500">
              {{ file.lastModifiedText }}
            </td>
            <td class="px-6 py-4 text-right">
              <button
                class="rounded-lg p-2 text-slate-400 group-hover:bg-white hover:text-slate-600 focus:outline-none focus:ring-2 focus:ring-primary/30 dark:group-hover:bg-slate-900 dark:hover:text-slate-200"
                type="button"
                :aria-label="`${file.name} 操作菜单`"
                @click.stop="handleOpenMenu(file, $event)"
              >
                <Icon icon="material-symbols:more-vert" aria-hidden="true" />
              </button>
            </td>
          </tr>

          <tr v-if="loading">
            <td colspan="7" class="py-16">
              <div class="flex flex-col items-center justify-center text-center">
                <Icon
                  icon="material-symbols:progress-activity"
                  class="mb-3 animate-spin text-4xl text-primary"
                />
                <p class="text-sm text-slate-400">加载中...</p>
              </div>
            </td>
          </tr>
          <tr v-if="files.length === 0 && !loading">
            <td colspan="7" class="py-20">
              <FileEmptyState :query="query" @upload="emit('upload')" />
            </td>
          </tr>
        </tbody>
      </table>
    </div>
    <slot name="footer" />
  </div>
</template>
