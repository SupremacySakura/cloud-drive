import { computed, onMounted, onScopeDispose, ref } from 'vue'
import type { FileDisplayItem } from '../types/file'

export function useFileActionMenu() {
  const menuTargetFile = ref<FileDisplayItem | null>(null)
  const menuPosition = ref<{ top: number; left: number } | null>(null)
  const isMenuOpen = computed(() => Boolean(menuTargetFile.value && menuPosition.value))

  const closeFileMenu = () => {
    menuTargetFile.value = null
    menuPosition.value = null
  }

  const openFileMenu = (file: FileDisplayItem, button: HTMLElement) => {
    if (menuTargetFile.value?.key === file.key) {
      closeFileMenu()
      return
    }

    const rect = button.getBoundingClientRect()
    const menuHeight = 220
    const menuWidth = 192
    const padding = 8
    const spaceBelow = window.innerHeight - rect.bottom
    const spaceAbove = rect.top
    const preferredTop =
      spaceBelow < menuHeight && spaceAbove > spaceBelow
        ? rect.top - menuHeight - padding
        : rect.bottom + padding

    menuPosition.value = {
      top: Math.max(padding, Math.min(preferredTop, window.innerHeight - menuHeight - padding)),
      left: Math.max(
        padding,
        Math.min(rect.right - menuWidth, window.innerWidth - menuWidth - padding),
      ),
    }
    menuTargetFile.value = file
  }

  const closeOnViewportChange = () => closeFileMenu()

  onMounted(() => {
    document.addEventListener('click', closeFileMenu)
    window.addEventListener('resize', closeOnViewportChange)
    window.addEventListener('scroll', closeOnViewportChange, true)
  })

  onScopeDispose(() => {
    document.removeEventListener('click', closeFileMenu)
    window.removeEventListener('resize', closeOnViewportChange)
    window.removeEventListener('scroll', closeOnViewportChange, true)
  })

  return {
    isMenuOpen,
    menuTargetFile,
    menuPosition,
    openFileMenu,
    closeFileMenu,
  }
}
