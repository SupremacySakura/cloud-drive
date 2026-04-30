import { defineStore } from 'pinia'
import type { FileListItem } from '../services/types/file'

// Breadcrumb item type
export type Breadcrumb = { id: number; name: string }

export const useFileStore = defineStore('files', {
  state: () => ({
    currentFolderId: 0,
    breadcrumbs: [] as Breadcrumb[],
    rawItems: [] as FileListItem[],
    selectedItems: new Set<number>() as Set<number>,
  }),
  getters: {
    // 便于组件读取：当前文件列表已排序后直接展示
    // 这里不实现复杂映射，保持简单
    getSelectedCount(state) {
      return state.selectedItems.size
    },
  },
  actions: {
    setCurrentFolderId(id: number) {
      this.currentFolderId = id
    },
    pushBreadcrumb(item: Breadcrumb) {
      this.breadcrumbs = [...this.breadcrumbs, item]
    },
    setBreadcrumbs(list: Breadcrumb[]) {
      this.breadcrumbs = list
    },
    setRawItems(items: FileListItem[]) {
      this.rawItems = items
    },
    setSelectedItems(ids: Set<number>) {
      this.selectedItems = ids
    },
    toggleSelectItem(id: number) {
      const next = new Set(this.selectedItems)
      if (next.has(id)) {
        next.delete(id)
      } else {
        next.add(id)
      }
      this.selectedItems = next
    },
    clearSelection() {
      this.selectedItems = new Set<number>()
    },
  },
})
