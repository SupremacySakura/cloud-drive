<script setup lang="ts">
import { onScopeDispose, ref, watch } from 'vue'
import axios from 'axios'
import type { FileDisplayItem } from '../../types/file'
import { getKnowledgeByFile, type KnowledgeItem } from '../../services/apis/knowledge'
import { isKnowledgeActive, isMarkdownFile } from '../../utils/knowledge'
import KnowledgeBadge from './KnowledgeBadge.vue'
const props = defineProps<{ file: FileDisplayItem; revision?: number }>()
const emit = defineEmits<{ action: [file: FileDisplayItem, knowledge: KnowledgeItem | null] }>()
const item = ref<KnowledgeItem | null>(null)
const state = ref<'loading' | 'loaded' | 'error'>('loading')
let controller: AbortController | null = null
let timer: ReturnType<typeof setTimeout> | undefined
async function load() {
  controller?.abort()
  clearTimeout(timer)
  if (!isMarkdownFile(props.file)) return
  const current = new AbortController()
  controller = current
  state.value = 'loading'
  try {
    const data = await getKnowledgeByFile(props.file.id, current.signal)
    if (current.signal.aborted) return
    item.value = data
    state.value = 'loaded'
    if (isKnowledgeActive(data.status)) timer = setTimeout(load, 3000)
  } catch (error) {
    if (current.signal.aborted) return
    if (axios.isAxiosError(error) && error.response?.status === 404) {
      item.value = null
      state.value = 'loaded'
    } else state.value = 'error'
  }
}
watch(() => [props.file.id, props.file.name, props.revision], load, { immediate: true })
onScopeDispose(() => {
  controller?.abort()
  clearTimeout(timer)
})
</script>
<template>
  <div v-if="file.type === 'file'" class="flex flex-wrap items-center gap-2" @click.stop>
    <template v-if="isMarkdownFile(file)">
      <KnowledgeBadge v-if="item" :status="item.status" />
      <span v-else class="text-caption text-label-secondary">{{
        state === 'loading' ? '查询状态…' : state === 'error' ? '状态查询失败' : '未加入'
      }}</span>
      <button v-if="state === 'error'" type="button" class="knowledge-text-action" @click="load">
        重新查询
      </button>
      <button
        v-else
        type="button"
        class="knowledge-text-action"
        :disabled="state === 'loading' || (!item && file.size > 20 * 1024 * 1024)"
        @click="emit('action', file, item)"
      >
        {{
          item?.status === 'ready'
            ? '查看知识'
            : item?.status === 'failed'
              ? '查看失败原因'
              : item
                ? '查看进度'
                : file.size > 20 * 1024 * 1024
                  ? '超过 20 MB'
                  : '加入知识库'
        }}
      </button>
    </template>
    <span v-else class="text-caption text-label-secondary">知识化暂不支持此格式</span>
  </div>
</template>
