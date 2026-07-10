<script setup lang="ts">
import { computed } from 'vue'
import { Icon } from '@iconify/vue'

const props = withDefaults(
  defineProps<{
    query?: string
  }>(),
  {
    query: '',
  },
)

const emit = defineEmits<{
  upload: []
}>()

const hasSearchQuery = computed(() => props.query.trim().length > 0)
const illustrationIcon = computed(() =>
  hasSearchQuery.value ? 'material-symbols:search-off' : 'material-symbols:cloud-upload-outline',
)
const title = computed(() => (hasSearchQuery.value ? '没有找到匹配的文件' : '还没有文件'))
const description = computed(() =>
  hasSearchQuery.value ? '试试其他关键词' : '点击上传按钮，开始管理你的文件',
)
</script>

<template>
  <div class="flex flex-col items-center justify-center text-center">
    <div class="relative mb-6">
      <div class="flex h-28 w-28 items-center justify-center rounded-full bg-primary/5">
        <div class="flex h-20 w-20 items-center justify-center rounded-full bg-primary/10">
          <Icon class="text-4xl text-primary" :icon="illustrationIcon" />
        </div>
      </div>
      <div class="absolute right-1 top-1 h-3 w-3 rounded-full bg-primary/20"></div>
      <div class="absolute bottom-3 left-0 h-2 w-2 rounded-full bg-primary/15"></div>
    </div>
    <p class="mb-1 text-lg font-semibold text-slate-700 dark:text-slate-200">{{ title }}</p>
    <p class="mb-5 text-sm text-slate-400 dark:text-slate-500">{{ description }}</p>
    <button
      v-if="!hasSearchQuery"
      class="flex items-center gap-2 rounded-lg bg-primary px-5 py-2.5 text-sm font-bold text-white shadow-lg shadow-primary/20 transition-all hover:bg-primary/90 focus:outline-none focus:ring-2 focus:ring-primary/50"
      type="button"
      @click="emit('upload')"
    >
      <Icon class="text-[20px]" icon="material-symbols:upload" />
      上传文件
    </button>
  </div>
</template>
