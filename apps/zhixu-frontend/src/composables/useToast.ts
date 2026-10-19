import { onScopeDispose, ref } from 'vue'
import type { ToastType } from '../types/file'

export function useToast(duration = 3000) {
  const toastMessage = ref('')
  const toastType = ref<ToastType>('info')
  const showToast = ref(false)
  let timer: ReturnType<typeof setTimeout> | null = null

  const closeToast = () => {
    showToast.value = false
    if (timer) {
      clearTimeout(timer)
      timer = null
    }
  }

  const displayToast = (message: string, type: ToastType = 'info') => {
    if (timer) clearTimeout(timer)
    toastMessage.value = message
    toastType.value = type
    showToast.value = true
    timer = setTimeout(closeToast, duration)
  }

  onScopeDispose(closeToast)

  return {
    toastMessage,
    toastType,
    showToast,
    displayToast,
    closeToast,
  }
}
