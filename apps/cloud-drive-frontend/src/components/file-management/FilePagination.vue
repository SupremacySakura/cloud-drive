<script setup lang="ts">
import { computed, ref } from 'vue'
import { Icon } from '@iconify/vue'

const props = defineProps<{
  page: number
  totalPages: number
  startIndex: number
  endIndex: number
  totalCount: number
  loading: boolean
}>()

const emit = defineEmits<{
  change: [page: number]
}>()

const jumpToPageInput = ref('')
const safeTotalPages = computed(() => Math.max(1, props.totalPages))
const pageNumbers = computed(() => {
  const total = safeTotalPages.value
  if (total <= 3) return Array.from({ length: total }, (_, index) => index + 1)
  if (props.page <= 1) return [1, 2, 3]
  if (props.page >= total) return [total - 2, total - 1, total]
  return [props.page - 1, props.page, props.page + 1]
})

const requestPage = (page: number) => {
  const target = Math.min(Math.max(1, page), safeTotalPages.value)
  if (target !== props.page) emit('change', target)
}

const submitJump = () => {
  const target = Number.parseInt(jumpToPageInput.value, 10)
  if (!Number.isFinite(target) || target < 1 || target > safeTotalPages.value) return
  jumpToPageInput.value = ''
  requestPage(target)
}
</script>

<template>
  <div
    class="flex flex-col gap-3 border-t border-slate-200 bg-white px-4 py-4 dark:border-slate-800 dark:bg-slate-950 sm:flex-row sm:items-center sm:justify-between sm:px-6"
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
    <div class="flex flex-wrap items-center gap-2">
      <button
        class="rounded-lg border border-slate-200 p-1.5 text-slate-400 hover:bg-slate-50 focus:outline-none focus:ring-2 focus:ring-primary/30 dark:border-slate-800 dark:hover:bg-slate-900"
        type="button"
        aria-label="上一页"
        :disabled="page <= 1 || loading"
        @click="requestPage(page - 1)"
      >
        <Icon class="text-sm" icon="material-symbols:chevron-left" />
      </button>
      <button
        v-for="pageNumber in pageNumbers"
        :key="pageNumber"
        class="flex h-8 w-8 items-center justify-center rounded-lg text-xs font-medium focus:outline-none focus:ring-2 focus:ring-primary/30"
        :class="
          pageNumber === page
            ? 'bg-primary font-bold text-white'
            : 'text-slate-700 hover:bg-slate-50 dark:text-slate-200 dark:hover:bg-slate-900'
        "
        type="button"
        :aria-label="`第 ${pageNumber} 页`"
        :aria-current="pageNumber === page ? 'page' : undefined"
        :disabled="loading"
        @click="requestPage(pageNumber)"
      >
        {{ pageNumber }}
      </button>
      <span v-if="safeTotalPages > 3" class="px-1 text-slate-400">...</span>
      <button
        class="rounded-lg border border-slate-200 p-1.5 text-slate-400 hover:bg-slate-50 focus:outline-none focus:ring-2 focus:ring-primary/30 dark:border-slate-800 dark:hover:bg-slate-900"
        type="button"
        aria-label="下一页"
        :disabled="page >= safeTotalPages || loading"
        @click="requestPage(page + 1)"
      >
        <Icon class="text-sm" icon="material-symbols:chevron-right" />
      </button>
      <div v-if="safeTotalPages > 1" class="ml-2 flex items-center gap-1">
        <span class="text-xs text-slate-400">跳至</span>
        <input
          v-model="jumpToPageInput"
          type="number"
          :min="1"
          :max="safeTotalPages"
          aria-label="跳转页码"
          class="w-12 rounded border border-slate-200 bg-white px-1.5 py-1 text-center text-xs text-slate-900 focus:outline-none focus:ring-1 focus:ring-primary/30 dark:border-slate-700 dark:bg-slate-900 dark:text-slate-100"
          @keyup.enter="submitJump"
        />
        <span class="text-xs text-slate-400">页</span>
      </div>
    </div>
  </div>
</template>
