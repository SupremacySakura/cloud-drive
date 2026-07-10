<script setup lang="ts">
import { computed, nextTick, ref, useId, watch } from 'vue'
import { Icon } from '@iconify/vue'

const props = withDefaults(
  defineProps<{
    modelValue: boolean
    title: string
    value: string
    label: string
    placeholder?: string
    description?: string | null
    loading?: boolean
    confirmText?: string
    loadingText?: string
    cancelText?: string
  }>(),
  {
    placeholder: '',
    description: null,
    loading: false,
    confirmText: '确认',
    loadingText: '处理中...',
    cancelText: '取消',
  },
)

const emit = defineEmits<{
  'update:modelValue': [value: boolean]
  'update:value': [value: string]
  confirm: []
  cancel: []
}>()

const inputRef = ref<HTMLInputElement | null>(null)
const titleId = `file-name-dialog-title-${useId()}`
const isConfirmDisabled = computed(() => props.loading || props.value.trim().length === 0)

const handleInput = (event: Event) => {
  const input = event.currentTarget
  if (input instanceof HTMLInputElement) emit('update:value', input.value)
}

const handleCancel = () => {
  if (props.loading) return
  emit('update:modelValue', false)
  emit('cancel')
}

const handleConfirm = () => {
  if (isConfirmDisabled.value) return
  emit('confirm')
}

watch(
  () => props.modelValue,
  async visible => {
    if (!visible) return
    await nextTick()
    inputRef.value?.focus()
    inputRef.value?.select()
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
        class="w-full max-w-md rounded-xl border border-slate-200 bg-white p-6 shadow-xl dark:border-slate-800 dark:bg-slate-950"
      >
        <h3 :id="titleId" class="mb-2 text-lg font-bold text-slate-900 dark:text-slate-100">
          {{ title }}
        </h3>
        <slot name="description">
          <p v-if="description" class="mb-4 text-sm text-slate-500">{{ description }}</p>
        </slot>

        <div class="mb-4" :class="description ? '' : 'mt-4'">
          <label class="mb-2 block text-sm font-medium text-slate-700 dark:text-slate-300">
            {{ label }}
          </label>
          <input
            ref="inputRef"
            :value="value"
            type="text"
            class="w-full rounded-lg border border-slate-300 bg-white px-4 py-2 text-slate-900 focus:outline-none focus:ring-2 focus:ring-primary/50 dark:border-slate-700 dark:bg-slate-900 dark:text-slate-100"
            :placeholder="placeholder"
            :disabled="loading"
            @input="handleInput"
            @keyup.enter="handleConfirm"
          />
        </div>

        <div class="flex justify-end gap-3">
          <button
            class="rounded-lg px-4 py-2 text-sm font-medium text-slate-600 transition-colors hover:bg-slate-100 focus:outline-none focus:ring-2 focus:ring-slate-400 disabled:cursor-not-allowed disabled:opacity-60 dark:text-slate-400 dark:hover:bg-slate-900"
            type="button"
            :disabled="loading"
            @click="handleCancel"
          >
            {{ cancelText }}
          </button>
          <button
            class="flex items-center gap-2 rounded-lg bg-primary px-4 py-2 text-sm font-medium text-white transition-colors hover:bg-primary/90 focus:outline-none focus:ring-2 focus:ring-primary/50 disabled:cursor-not-allowed disabled:opacity-60"
            type="button"
            :disabled="isConfirmDisabled"
            @click="handleConfirm"
          >
            <Icon v-if="loading" icon="material-symbols:progress-activity" class="animate-spin" />
            {{ loading ? loadingText : confirmText }}
          </button>
        </div>
      </div>
    </div>
  </Teleport>
</template>
