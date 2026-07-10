<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref } from 'vue'
import { Icon } from '@iconify/vue'
import type { FileFilterKey, FileSortKey, FileViewMode, SortDirection } from '../../types/file'

export type FileFilterOption = {
  key: FileFilterKey
  label: string
}

const props = defineProps<{
  searchQuery: string
  viewMode: FileViewMode
  filterOptions: readonly FileFilterOption[]
  activeFilter: FileFilterKey
  selectedCount: number
  sortKey: FileSortKey
  sortDirection: SortDirection
}>()

const emit = defineEmits<{
  'update:searchQuery': [value: string]
  'update:viewMode': [value: FileViewMode]
  'clear-search': []
  'select-filter': [key: FileFilterKey]
  'select-sort': [key: FileSortKey]
  'select-current-page': []
  'clear-selection': []
}>()

const rootRef = ref<HTMLElement | null>(null)
const isFilterOpen = ref(false)
const isSortOpen = ref(false)

const activeFilterLabel = computed(
  () => props.filterOptions.find(option => option.key === props.activeFilter)?.label ?? '全部类型',
)
const sortLabel = computed(() => {
  if (props.sortKey === 'name') return 'Name'
  if (props.sortKey === 'size') return 'Size'
  return 'Last Modified'
})

const closeMenus = () => {
  isFilterOpen.value = false
  isSortOpen.value = false
}

const handleDocumentClick = (event: MouseEvent) => {
  const target = event.target
  if (target instanceof Node && rootRef.value && !rootRef.value.contains(target)) closeMenus()
}

const handleSearchInput = (event: Event) => {
  const input = event.currentTarget
  if (input instanceof HTMLInputElement) emit('update:searchQuery', input.value)
}

const clearSearch = () => {
  emit('update:searchQuery', '')
  emit('clear-search')
}

const selectViewMode = (viewMode: FileViewMode) => emit('update:viewMode', viewMode)
const toggleFilterMenu = () => {
  isFilterOpen.value = !isFilterOpen.value
  isSortOpen.value = false
}
const toggleSortMenu = () => {
  isSortOpen.value = !isSortOpen.value
  isFilterOpen.value = false
}
const selectFilter = (key: FileFilterKey) => {
  emit('select-filter', key)
  isFilterOpen.value = false
}
const selectSort = (key: FileSortKey) => {
  emit('select-sort', key)
  isSortOpen.value = false
}

onMounted(() => document.addEventListener('click', handleDocumentClick))
onBeforeUnmount(() => document.removeEventListener('click', handleDocumentClick))
</script>

<template>
  <div
    ref="rootRef"
    class="mb-6 flex flex-wrap items-center justify-between gap-4 rounded-xl border border-slate-200 bg-white p-3 shadow-sm dark:border-slate-800 dark:bg-slate-950"
  >
    <div class="flex w-full flex-wrap items-center gap-2 lg:w-auto">
      <div class="relative w-full sm:w-auto">
        <Icon
          class="absolute left-3 top-1/2 -translate-y-1/2 text-sm text-slate-400"
          icon="material-symbols:search"
        />
        <input
          :value="searchQuery"
          type="text"
          placeholder="搜索文件..."
          aria-label="搜索文件"
          class="w-full rounded-lg border border-slate-200 bg-white py-1.5 pl-9 pr-8 text-sm text-slate-900 placeholder:text-slate-400 focus:border-primary focus:outline-none focus:ring-2 focus:ring-primary/30 dark:border-slate-700 dark:bg-slate-900 dark:text-slate-100 sm:w-48"
          @input="handleSearchInput"
        />
        <button
          v-if="searchQuery"
          class="absolute right-2 top-1/2 -translate-y-1/2 rounded text-slate-400 transition-colors hover:text-slate-600 focus:outline-none focus:ring-2 focus:ring-primary/30 dark:hover:text-slate-200"
          type="button"
          aria-label="清空搜索"
          @click="clearSearch"
        >
          <Icon class="text-sm" icon="material-symbols:close" />
        </button>
      </div>

      <div class="mx-1 hidden h-6 w-px bg-slate-200 dark:bg-slate-800 sm:block"></div>

      <button
        class="rounded-lg p-2"
        :class="
          viewMode === 'list'
            ? 'bg-primary/10 text-primary'
            : 'text-slate-400 hover:bg-slate-100 dark:hover:bg-slate-900'
        "
        aria-label="列表视图"
        type="button"
        @click="selectViewMode('list')"
      >
        <Icon icon="material-symbols:list" aria-hidden="true" />
      </button>
      <button
        class="rounded-lg p-2"
        :class="
          viewMode === 'grid'
            ? 'bg-primary/10 text-primary'
            : 'text-slate-400 hover:bg-slate-100 dark:hover:bg-slate-900'
        "
        aria-label="网格视图"
        type="button"
        @click="selectViewMode('grid')"
      >
        <Icon icon="material-symbols:grid-view" aria-hidden="true" />
      </button>

      <div class="mx-2 hidden h-6 w-px bg-slate-200 dark:bg-slate-800 sm:block"></div>

      <div class="relative">
        <button
          class="flex items-center gap-2 rounded-lg border border-transparent px-3 py-1.5 text-sm font-medium text-slate-600 hover:border-slate-200 hover:bg-slate-50 focus:outline-none focus:ring-2 focus:ring-primary/30 dark:text-slate-400 dark:hover:border-slate-800 dark:hover:bg-slate-900"
          :class="
            activeFilter !== 'all'
              ? 'border-primary/20 bg-primary/10 text-primary hover:border-primary/20 hover:bg-primary/10 dark:hover:border-primary/20'
              : ''
          "
          type="button"
          aria-label="筛选文件"
          :aria-expanded="isFilterOpen"
          @click="toggleFilterMenu"
        >
          <Icon class="text-[18px]" icon="material-symbols:filter-list" />
          {{ activeFilter === 'all' ? '筛选' : activeFilterLabel }}
        </button>

        <div
          v-if="isFilterOpen"
          class="absolute left-0 top-full z-20 mt-2 w-44 rounded-xl border border-slate-200 bg-white py-2 shadow-xl dark:border-slate-800 dark:bg-slate-950"
        >
          <button
            v-for="option in filterOptions"
            :key="option.key"
            class="flex w-full items-center justify-between px-4 py-2 text-left text-sm text-slate-700 hover:bg-slate-50 dark:text-slate-300 dark:hover:bg-slate-900"
            type="button"
            @click="selectFilter(option.key)"
          >
            <span>{{ option.label }}</span>
            <Icon
              v-if="activeFilter === option.key"
              icon="material-symbols:check-rounded"
              class="text-primary"
            />
          </button>
        </div>
      </div>

      <div v-if="selectedCount > 0" class="ml-2 flex items-center gap-2 text-sm text-slate-500">
        <span>已选择 {{ selectedCount }} 项</span>
        <button
          class="font-semibold text-primary hover:underline"
          type="button"
          @click="emit('select-current-page')"
        >
          选择当前页
        </button>
        <button
          class="font-semibold text-primary hover:underline"
          type="button"
          @click="emit('clear-selection')"
        >
          清空
        </button>
      </div>
    </div>

    <div class="relative w-full sm:w-auto">
      <div class="flex items-center gap-2 text-xs font-medium text-slate-400 sm:justify-end">
        <span>Sorted by</span>
        <button
          class="flex items-center gap-1 text-slate-900 hover:text-primary dark:text-slate-100"
          type="button"
          :aria-expanded="isSortOpen"
          @click="toggleSortMenu"
        >
          {{ sortLabel }}
          <Icon class="text-sm" icon="material-symbols:expand-more" />
        </button>
      </div>

      <div
        v-if="isSortOpen"
        class="absolute right-0 z-20 mt-2 w-48 rounded-xl border border-slate-200 bg-white py-2 shadow-xl dark:border-slate-800 dark:bg-slate-950"
      >
        <button
          class="w-full px-4 py-2 text-left text-sm text-slate-700 hover:bg-slate-50 dark:text-slate-300 dark:hover:bg-slate-900"
          type="button"
          @click="selectSort('name')"
        >
          Name
        </button>
        <button
          class="w-full px-4 py-2 text-left text-sm text-slate-700 hover:bg-slate-50 dark:text-slate-300 dark:hover:bg-slate-900"
          type="button"
          @click="selectSort('modified')"
        >
          Last Modified
        </button>
        <button
          class="w-full px-4 py-2 text-left text-sm text-slate-700 hover:bg-slate-50 dark:text-slate-300 dark:hover:bg-slate-900"
          type="button"
          @click="selectSort('size')"
        >
          Size
        </button>
        <div class="my-1 border-t border-slate-100 dark:border-slate-800"></div>
        <div class="px-4 py-2 text-xs text-slate-400">
          {{ sortDirection === 'asc' ? 'Ascending' : 'Descending' }}
        </div>
      </div>
    </div>
  </div>
</template>
