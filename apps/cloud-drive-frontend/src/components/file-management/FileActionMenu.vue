<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, onMounted, ref, watch, type CSSProperties } from 'vue'
import { Icon } from '@iconify/vue'
import type { FileDisplayItem } from '../../types/file'

const props = withDefaults(
  defineProps<{
    visible: boolean
    target: FileDisplayItem | null
    position: { top: number; left: number } | null
    downloading?: boolean
    deleting?: boolean
  }>(),
  {
    downloading: false,
    deleting: false,
  },
)

const emit = defineEmits<{
  preview: []
  download: []
  rename: []
  move: []
  delete: []
  close: []
}>()

const menuRef = ref<HTMLElement | null>(null)
const isOpen = computed(() => props.visible && props.target !== null && props.position !== null)
const menuStyle = computed<CSSProperties>(() => ({
  top: `${props.position?.top ?? 0}px`,
  left: `${props.position?.left ?? 0}px`,
}))
const previewDisabled = computed(() => props.target?.type !== 'file')

const handleKeydown = (event: KeyboardEvent) => {
  if (props.visible && event.key === 'Escape') emit('close')
}

watch(isOpen, async open => {
  if (!open) return
  await nextTick()
  menuRef.value?.focus()
})

onMounted(() => document.addEventListener('keydown', handleKeydown))
onBeforeUnmount(() => document.removeEventListener('keydown', handleKeydown))
</script>

<template>
  <Teleport to="body">
    <div
      v-if="isOpen"
      ref="menuRef"
      class="fixed z-50 w-48 rounded-xl border border-slate-200 bg-white py-2 shadow-xl outline-none dark:border-slate-800 dark:bg-slate-950"
      :style="menuStyle"
      role="menu"
      tabindex="-1"
      :aria-label="`${target?.name ?? '文件'} 操作菜单`"
      @click.stop
    >
      <button
        class="flex w-full items-center gap-3 px-4 py-2 text-sm hover:bg-slate-50 focus:outline-none focus:ring-2 focus:ring-primary/30 disabled:cursor-not-allowed disabled:opacity-50 dark:hover:bg-slate-900"
        :class="previewDisabled ? 'text-slate-400' : 'text-slate-700 dark:text-slate-300'"
        type="button"
        role="menuitem"
        aria-label="预览文件"
        :disabled="previewDisabled"
        @click="emit('preview')"
      >
        <Icon class="text-sm" icon="material-symbols:visibility" />
        预览
      </button>
      <button
        class="flex w-full items-center gap-3 px-4 py-2 text-sm text-slate-700 hover:bg-slate-50 focus:outline-none focus:ring-2 focus:ring-primary/30 disabled:cursor-not-allowed disabled:opacity-50 dark:text-slate-300 dark:hover:bg-slate-900"
        type="button"
        role="menuitem"
        aria-label="下载文件"
        :disabled="downloading"
        @click="emit('download')"
      >
        <Icon class="text-sm" icon="material-symbols:download" />
        下载
      </button>
      <button
        class="flex w-full items-center gap-3 px-4 py-2 text-sm text-slate-700 hover:bg-slate-50 focus:outline-none focus:ring-2 focus:ring-primary/30 dark:text-slate-300 dark:hover:bg-slate-900"
        type="button"
        role="menuitem"
        aria-label="重命名文件"
        @click="emit('rename')"
      >
        <Icon class="text-sm" icon="material-symbols:edit" />
        重命名
      </button>
      <button
        class="flex w-full items-center gap-3 px-4 py-2 text-sm text-slate-700 hover:bg-slate-50 focus:outline-none focus:ring-2 focus:ring-primary/30 dark:text-slate-300 dark:hover:bg-slate-900"
        type="button"
        role="menuitem"
        aria-label="移动文件"
        @click="emit('move')"
      >
        <Icon class="text-sm" icon="material-symbols:drive-file-move" />
        移动到
      </button>
      <div class="my-1 border-t border-slate-100 dark:border-slate-800"></div>
      <button
        class="flex w-full items-center gap-3 px-4 py-2 text-sm text-red-500 hover:bg-red-50 focus:outline-none focus:ring-2 focus:ring-red-400 disabled:cursor-not-allowed disabled:opacity-50 dark:hover:bg-red-900/20"
        type="button"
        role="menuitem"
        aria-label="删除文件"
        :disabled="deleting"
        @click="emit('delete')"
      >
        <Icon class="text-sm" icon="material-symbols:delete" />
        删除
      </button>
    </div>
  </Teleport>
</template>
