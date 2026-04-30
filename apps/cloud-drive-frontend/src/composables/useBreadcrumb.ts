import { ref, type Ref } from 'vue'

export type BreadcrumbItem = { id: number; name: string }

/**
 * 面包屑导航状态容器
 * 目前仅负责维护 breadcrumbs 的结构，实际导航行为由调用方实现
 */
export function useBreadcrumb() {
  const breadcrumbs: Ref<Array<BreadcrumbItem>> = ref([{ id: 0, name: 'root' }])

  // 当前目录名称，依赖 breadcrumbs 最后一个元素
  const currentName = ref<string>(breadcrumbs.value.at(-1)?.name ?? 'root')

  // 更新当前名称的副作用
  const updateCurrentName = () => {
    currentName.value = breadcrumbs.value.at(-1)?.name ?? 'root'
  }

  const pushBreadcrumb = (item: BreadcrumbItem) => {
    breadcrumbs.value = [...breadcrumbs.value, item]
    updateCurrentName()
  }

  const goToBreadcrumb = async (index: number) => {
    if (index < 0 || index >= breadcrumbs.value.length) return
    breadcrumbs.value = breadcrumbs.value.slice(0, index + 1)
    updateCurrentName()
  }

  const goToParent = async () => {
    if (breadcrumbs.value.length > 1) {
      breadcrumbs.value = breadcrumbs.value.slice(0, -1)
      updateCurrentName()
    }
  }

  return {
    breadcrumbs,
    currentName,
    pushBreadcrumb,
    goToBreadcrumb,
    goToParent,
  }
}
