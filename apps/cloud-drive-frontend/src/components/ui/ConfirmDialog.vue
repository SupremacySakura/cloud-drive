<script setup lang="ts">
import { ref, watch, nextTick } from 'vue'

const props = withDefaults(
  defineProps<{
    modelValue: boolean
    title?: string
    message?: string
    confirmText?: string
    cancelText?: string
    loading?: boolean
    danger?: boolean
  }>(),
  {
    title: '确认操作',
    message: '',
    confirmText: '确认',
    cancelText: '取消',
    loading: false,
    danger: false,
  },
)

const emit = defineEmits<{
  (e: 'update:modelValue', value: boolean): void
  (e: 'cancel'): void
  (e: 'confirm'): void
}>()

const confirmButtonRef = ref<HTMLButtonElement | null>(null)

const handleClose = () => {
  if (props.loading) return
  emit('update:modelValue', false)
  emit('cancel')
}

const handleConfirm = () => {
  if (props.loading) return
  emit('confirm')
}

// 弹窗打开时聚焦确认按钮
watch(
  () => props.modelValue,
  visible => {
    if (visible) {
      nextTick(() => {
        confirmButtonRef.value?.focus()
      })
    }
  },
)
</script>

<template>
  <Teleport to="body">
    <div
      v-if="modelValue"
      role="dialog"
      aria-modal="true"
      :aria-labelledby="'dialog-title-' + title"
      class="fixed inset-0 bg-black/50 backdrop-blur-sm z-50 flex items-center justify-center p-4"
      @click="handleClose"
    >
      <div
        class="bg-white dark:bg-slate-950 rounded-xl border border-slate-200 dark:border-slate-800 shadow-xl w-full max-w-md p-6"
        @click.stop
      >
        <h3
          :id="'dialog-title-' + title"
          class="text-lg font-bold text-slate-900 dark:text-slate-100 mb-2"
        >
          {{ title }}
        </h3>
        <p v-if="message" class="text-sm text-slate-600 dark:text-slate-300 mb-5">{{ message }}</p>

        <div class="flex justify-end gap-3">
          <button
            class="px-4 py-2 text-sm font-medium text-slate-600 dark:text-slate-400 hover:bg-slate-100 dark:hover:bg-slate-900 rounded-lg transition-colors disabled:opacity-60 disabled:cursor-not-allowed focus:ring-2 focus:ring-slate-400 focus:outline-none"
            type="button"
            :disabled="loading"
            @click="handleClose"
          >
            {{ cancelText }}
          </button>
          <button
            ref="confirmButtonRef"
            class="px-4 py-2 text-sm font-medium text-white rounded-lg transition-colors flex items-center gap-2 disabled:opacity-60 disabled:cursor-not-allowed focus:ring-2 focus:outline-none"
            :class="
              danger
                ? 'bg-red-500 hover:bg-red-600 focus:ring-red-400'
                : 'bg-primary hover:bg-primary/90 focus:ring-primary/50'
            "
            type="button"
            :disabled="loading"
            @click="handleConfirm"
          >
            <slot name="confirm-icon" />
            {{ loading ? '处理中...' : confirmText }}
          </button>
        </div>
      </div>
    </div>
  </Teleport>
</template>
