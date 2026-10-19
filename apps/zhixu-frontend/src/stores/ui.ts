import { defineStore } from 'pinia'
export type Toast = { id: string; message: string; type: 'success' | 'error' | 'info' }

export const useUiStore = defineStore('ui', {
  state: () => ({
    activeModal: null as string | null,
    toastMessages: [] as Toast[],
    sidebarCollapsed: false,
  }),
  actions: {
    openModal(name: string) {
      this.activeModal = name
    },
    closeModal() {
      this.activeModal = null
    },
    pushToast(message: string, type: 'success' | 'error' | 'info' = 'info') {
      const id = Math.random().toString(36).slice(2)
      this.toastMessages.push({ id, message, type })
      setTimeout(() => {
        this.toastMessages = this.toastMessages.filter(t => t.id !== id)
      }, 3000)
    },
    toggleSidebar() {
      this.sidebarCollapsed = !this.sidebarCollapsed
    },
  },
})
