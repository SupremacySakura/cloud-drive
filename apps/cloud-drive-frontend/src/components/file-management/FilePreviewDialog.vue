<script setup lang="ts">
import { computed, nextTick, ref, useId, watch } from 'vue'
import { Icon } from '@iconify/vue'
import type { FileDisplayItem, PreviewKind } from '../../types/file'
import { formatBytes } from '../../utils/file'

const props = withDefaults(
  defineProps<{
    modelValue: boolean
    file: FileDisplayItem | null
    loading: boolean
    error: string | null
    kind: PreviewKind
    url: string
    textContent: string
    canDownload: boolean
    publicShareLink: string
    creatingShareLink?: boolean
    deletingShareLink?: boolean
  }>(),
  {
    creatingShareLink: false,
    deletingShareLink: false,
  },
)

const emit = defineEmits<{
  'update:modelValue': [value: boolean]
  close: []
  download: []
  'create-share': []
  'delete-share': []
  'copy-share': []
}>()

const closeButtonRef = ref<HTMLButtonElement | null>(null)
const titleId = `file-preview-dialog-title-${useId()}`
const fileName = computed(() => props.file?.name ?? '文件预览')
const fileTypeLabel = computed(() => props.file?.typeLabel ?? 'File')
const fileSize = computed(() => formatBytes(props.file?.size ?? 0))
const lastModifiedText = computed(() => props.file?.lastModifiedText ?? '-')
const fileIcon = computed(() => props.file?.icon ?? 'material-symbols:description')

const handleClose = () => {
  emit('update:modelValue', false)
  emit('close')
}

watch(
  () => props.modelValue,
  async visible => {
    if (!visible) return
    await nextTick()
    closeButtonRef.value?.focus()
  },
)
</script>

<template>
  <Teleport to="body">
    <div
      v-if="modelValue"
      class="fixed inset-0 z-50 flex items-center justify-center bg-slate-900/60 p-4 backdrop-blur-sm sm:p-6 md:p-10"
      role="dialog"
      aria-modal="true"
      :aria-labelledby="titleId"
      @click.self="handleClose"
      @keydown.esc="handleClose"
    >
      <div
        class="relative flex h-full max-h-[850px] w-full max-w-6xl flex-col overflow-hidden rounded-xl border border-slate-200 bg-white shadow-2xl dark:border-slate-800 dark:bg-slate-950"
      >
        <div
          class="flex shrink-0 items-center justify-between border-b border-slate-100 px-6 py-4 dark:border-slate-800"
        >
          <div class="flex min-w-0 items-center gap-3">
            <div
              class="flex h-10 w-10 shrink-0 items-center justify-center rounded-lg bg-primary/10 text-primary"
            >
              <Icon class="text-[22px]" :icon="fileIcon" />
            </div>
            <div class="min-w-0">
              <h2
                :id="titleId"
                class="max-w-[60vw] truncate text-lg font-bold tracking-tight text-slate-900 dark:text-slate-100"
              >
                {{ fileName }}
              </h2>
              <p class="text-xs text-slate-400">{{ fileTypeLabel }} • {{ fileSize }}</p>
            </div>
          </div>
          <button
            ref="closeButtonRef"
            class="flex h-10 w-10 shrink-0 items-center justify-center rounded-full text-slate-400 transition-colors hover:bg-slate-100 hover:text-slate-600 focus:outline-none focus:ring-2 focus:ring-slate-400 dark:hover:bg-slate-900 dark:hover:text-slate-200"
            type="button"
            aria-label="关闭预览"
            @click="handleClose"
          >
            <Icon icon="material-symbols:close" />
          </button>
        </div>

        <div class="flex min-h-0 flex-1 flex-col overflow-hidden md:flex-row">
          <div
            class="relative flex min-h-0 flex-1 flex-col overflow-hidden bg-slate-50 dark:bg-slate-900/30"
          >
            <div class="flex-1 overflow-y-auto p-6 pb-24 md:p-8 md:pb-24">
              <div
                v-if="loading"
                class="flex min-h-full flex-col items-center justify-center gap-3 text-slate-500"
              >
                <Icon
                  icon="material-symbols:progress-activity"
                  class="animate-spin text-3xl text-primary"
                />
                <span class="text-sm">正在加载预览...</span>
              </div>

              <div
                v-else-if="error"
                class="flex min-h-full flex-col items-center justify-center px-6 text-center text-red-500"
              >
                <Icon icon="material-symbols:error" class="mb-2 text-4xl" />
                <p class="font-semibold">{{ error }}</p>
              </div>

              <div
                v-else-if="kind === 'image' && url"
                class="flex min-h-full w-full items-start justify-center"
              >
                <img
                  :src="url"
                  :alt="fileName"
                  class="h-auto max-w-full rounded-lg border border-slate-200 object-contain shadow-sm dark:border-slate-700"
                />
              </div>

              <iframe
                v-else-if="kind === 'pdf' && url"
                :src="url"
                class="h-[calc(100vh-300px)] min-h-[520px] w-full rounded-lg border border-slate-200 bg-white dark:border-slate-700"
                title="PDF 预览"
              ></iframe>

              <div
                v-else-if="kind === 'video' && url"
                class="flex min-h-full w-full items-start justify-center"
              >
                <video controls :src="url" class="max-w-full rounded-lg bg-black"></video>
              </div>

              <div
                v-else-if="kind === 'audio' && url"
                class="flex min-h-full items-center justify-center"
              >
                <audio controls :src="url" class="w-full max-w-2xl"></audio>
              </div>

              <pre
                v-else-if="kind === 'text'"
                class="h-full w-full overflow-auto whitespace-pre-wrap break-words rounded-lg border border-slate-200 bg-white p-4 text-sm text-slate-700 dark:border-slate-700 dark:bg-slate-950 dark:text-slate-200"
                >{{ textContent || '文件为空' }}</pre
              >

              <div
                v-else
                class="flex min-h-full flex-col items-center justify-center px-6 text-center text-slate-500"
              >
                <slot name="unsupported" :file="file">
                  <Icon icon="material-symbols:description" class="mb-2 text-4xl" />
                  <p class="font-semibold">当前文件类型暂不支持在线预览</p>
                  <p class="mt-1 text-xs">你可以使用下方按钮下载后查看</p>
                </slot>
              </div>
            </div>

            <div
              class="absolute bottom-6 left-1/2 z-20 flex -translate-x-1/2 items-center gap-1 rounded-xl border border-slate-200 bg-white p-2 shadow-lg dark:border-slate-800 dark:bg-slate-950"
            >
              <button
                class="flex items-center gap-2 rounded-lg bg-primary px-4 py-2 text-sm font-semibold text-white shadow-md shadow-primary/20 transition-all hover:bg-primary/90 active:scale-95 focus:outline-none focus:ring-2 focus:ring-primary/50 disabled:cursor-not-allowed disabled:opacity-60"
                type="button"
                aria-label="下载文件"
                :disabled="!canDownload"
                @click="emit('download')"
              >
                <Icon class="text-[20px]" icon="material-symbols:download" />
                下载
              </button>
            </div>
          </div>

          <aside
            class="flex max-h-72 w-full shrink-0 flex-col overflow-y-auto border-t border-slate-100 bg-white p-6 dark:border-slate-800 dark:bg-slate-950 md:max-h-none md:w-80 md:border-l md:border-t-0"
          >
            <h3 class="mb-6 text-xs font-bold uppercase tracking-widest text-slate-400">
              文件详情
            </h3>
            <div class="space-y-6">
              <div>
                <span class="mb-1 block text-[11px] text-slate-400">文件名</span>
                <p class="break-all text-sm font-semibold text-slate-800 dark:text-slate-100">
                  {{ fileName }}
                </p>
              </div>
              <div>
                <span class="mb-1 block text-[11px] text-slate-400">文件大小</span>
                <p class="text-sm font-semibold text-slate-800 dark:text-slate-100">
                  {{ fileSize }}
                </p>
              </div>
              <div>
                <span class="mb-1 block text-[11px] text-slate-400">文件类型</span>
                <p class="text-sm font-semibold text-slate-800 dark:text-slate-100">
                  {{ fileTypeLabel }}
                </p>
              </div>
              <div>
                <span class="mb-1 block text-[11px] text-slate-400">最后修改</span>
                <p class="text-sm font-semibold text-slate-800 dark:text-slate-100">
                  {{ lastModifiedText }}
                </p>
              </div>
              <div class="border-t border-slate-100 pt-5 dark:border-slate-800">
                <span class="mb-2 block text-[11px] text-slate-400"
                  >公网分享链接（免鉴权访问）</span
                >
                <div v-if="publicShareLink" class="space-y-2">
                  <input
                    :value="publicShareLink"
                    readonly
                    aria-label="公网分享链接"
                    class="w-full rounded-lg border border-slate-200 bg-slate-50 px-3 py-2 text-sm text-slate-700 dark:border-slate-700 dark:bg-slate-900 dark:text-slate-200"
                  />
                  <div class="grid grid-cols-2 gap-2">
                    <button
                      class="inline-flex items-center justify-center gap-1.5 rounded-lg bg-primary px-3 py-2 text-sm font-semibold text-white hover:bg-primary/90 focus:outline-none focus:ring-2 focus:ring-primary/50"
                      type="button"
                      aria-label="复制分享链接"
                      @click="emit('copy-share')"
                    >
                      <Icon class="text-[18px]" icon="material-symbols:content-copy-outline" />
                      复制
                    </button>
                    <button
                      class="inline-flex items-center justify-center gap-1.5 rounded-lg bg-red-500 px-3 py-2 text-sm font-semibold text-white hover:bg-red-600 focus:outline-none focus:ring-2 focus:ring-red-400 disabled:cursor-not-allowed disabled:opacity-60"
                      type="button"
                      aria-label="删除分享链接"
                      :disabled="deletingShareLink"
                      @click="emit('delete-share')"
                    >
                      <Icon
                        class="text-[18px]"
                        :icon="
                          deletingShareLink
                            ? 'material-symbols:progress-activity'
                            : 'material-symbols:delete-outline'
                        "
                        :class="deletingShareLink ? 'animate-spin' : ''"
                      />
                      {{ deletingShareLink ? '删除中' : '删除' }}
                    </button>
                  </div>
                </div>
                <button
                  v-else
                  class="inline-flex w-full items-center justify-center gap-2 rounded-lg bg-primary px-4 py-2 text-sm font-semibold text-white shadow-md shadow-primary/20 transition-all hover:bg-primary/90 active:scale-95 focus:outline-none focus:ring-2 focus:ring-primary/50 disabled:cursor-not-allowed disabled:opacity-60"
                  type="button"
                  aria-label="生成公网链接"
                  :disabled="creatingShareLink || !file"
                  @click="emit('create-share')"
                >
                  <Icon
                    class="text-[20px]"
                    :icon="
                      creatingShareLink
                        ? 'material-symbols:progress-activity'
                        : 'material-symbols:share'
                    "
                    :class="creatingShareLink ? 'animate-spin' : ''"
                  />
                  {{ creatingShareLink ? '生成中' : '生成公网链接' }}
                </button>
              </div>
            </div>
          </aside>
        </div>
      </div>
    </div>
  </Teleport>
</template>
