<script setup lang="ts">
import { computed } from 'vue'
import { Icon } from '@iconify/vue'
import type { ToastType } from '../../types/file'

const props = defineProps<{
  visible: boolean
  message: string
  type: ToastType
}>()

const icon = computed(() => {
  if (props.type === 'success') return 'material-symbols:check-circle'
  if (props.type === 'error') return 'material-symbols:error'
  return 'material-symbols:info'
})
const containerClass = computed(() => {
  if (props.type === 'success')
    return 'border-green-200 bg-green-50 dark:border-green-800 dark:bg-green-900/20'
  if (props.type === 'error')
    return 'border-red-200 bg-red-50 dark:border-red-800 dark:bg-red-900/20'
  return 'border-blue-200 bg-blue-50 dark:border-blue-800 dark:bg-blue-900/20'
})
const iconClass = computed(() => {
  if (props.type === 'success') return 'text-green-500'
  if (props.type === 'error') return 'text-red-500'
  return 'text-blue-500'
})
const textClass = computed(() => {
  if (props.type === 'success') return 'text-green-800 dark:text-green-200'
  if (props.type === 'error') return 'text-red-800 dark:text-red-200'
  return 'text-blue-800 dark:text-blue-200'
})
</script>

<template>
  <Transition
    enter-active-class="transform ease-out duration-300 transition"
    enter-from-class="translate-y-2 opacity-0"
    enter-to-class="translate-y-0 opacity-100"
    leave-active-class="transition ease-in duration-200"
    leave-from-class="opacity-100"
    leave-to-class="opacity-0"
  >
    <div
      v-if="visible"
      class="fixed right-4 top-4 z-50 flex items-center gap-3 rounded-xl border px-6 py-4 shadow-2xl"
      :class="containerClass"
      :role="type === 'error' ? 'alert' : 'status'"
      aria-live="polite"
    >
      <Icon :icon="icon" class="text-2xl" :class="iconClass" />
      <span class="font-medium" :class="textClass">{{ message }}</span>
    </div>
  </Transition>
</template>
