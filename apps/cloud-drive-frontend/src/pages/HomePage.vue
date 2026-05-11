<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { Icon } from '@iconify/vue'
import { useRoute } from 'vue-router'
import SideBar, { type NavItem } from '../components/bussiness/SideBar.vue'

const NavItems: NavItem[] = [
  { label: '仪表盘', icon: 'material-symbols:dashboard', to: '/home/dashboard' },
  { label: '文件管理', icon: 'material-symbols:folder-open', to: '/home/files' },
  { label: '取件码', icon: 'material-symbols:key-outline', to: '/home/pickup-codes' },
  { label: '上传', icon: 'material-symbols:cloud-upload', to: '/home/upload' },
]

const route = useRoute()
const isMobileMenuOpen = ref(false)

const currentNavLabel = computed(() => {
  const matched = NavItems.find(
    item => route.path === item.to || route.path.startsWith(`${item.to}/`),
  )
  return matched?.label || '云盘'
})

watch(
  () => route.fullPath,
  () => {
    isMobileMenuOpen.value = false
  },
)
</script>

<template>
  <div class="flex min-h-screen bg-background-light dark:bg-background-dark lg:h-screen">
    <!-- Skip to content link for accessibility -->
    <a
      href="#main-content"
      class="sr-only focus:not-sr-only focus:absolute focus:top-4 focus:left-4 focus:z-50 focus:px-4 focus:py-2 focus:bg-primary focus:text-white focus:rounded-lg focus:font-medium focus:outline-none focus:ring-2 focus:ring-primary/50"
    >
      跳转到主要内容
    </a>
    <div class="hidden h-full lg:block">
      <SideBar :nav-items="NavItems" :brand-title="'云盘'" :brand-subtitle="''" />
    </div>

    <div class="flex min-h-screen min-w-0 flex-1 flex-col lg:min-h-0">
      <header
        class="sticky top-0 z-40 flex items-center justify-between border-b border-slate-200 bg-white/95 px-4 py-3 backdrop-blur dark:border-slate-800 dark:bg-slate-950/95 lg:hidden"
      >
        <div class="min-w-0">
          <p class="text-base font-bold text-slate-900 dark:text-slate-100">云盘</p>
          <p class="truncate text-xs text-slate-500 dark:text-slate-400">{{ currentNavLabel }}</p>
        </div>
        <button
          class="inline-flex h-10 w-10 items-center justify-center rounded-lg border border-slate-200 text-slate-700 transition-colors hover:bg-slate-50 focus:outline-none focus:ring-2 focus:ring-primary/30 dark:border-slate-800 dark:text-slate-200 dark:hover:bg-slate-900"
          type="button"
          aria-label="打开导航菜单"
          @click="isMobileMenuOpen = true"
        >
          <Icon icon="material-symbols:menu-rounded" class="text-2xl" />
        </button>
      </header>

      <router-view id="main-content" class="min-h-0 min-w-0 flex-1" tabindex="-1" />
    </div>

    <Transition
      enter-active-class="transition duration-200 ease-out"
      enter-from-class="opacity-0"
      enter-to-class="opacity-100"
      leave-active-class="transition duration-150 ease-in"
      leave-from-class="opacity-100"
      leave-to-class="opacity-0"
    >
      <div
        v-if="isMobileMenuOpen"
        class="fixed inset-0 z-50 flex lg:hidden"
        @click="isMobileMenuOpen = false"
      >
        <div class="absolute inset-0 bg-slate-950/40 backdrop-blur-sm"></div>
        <div
          class="relative flex h-full w-[min(18rem,86vw)] flex-col bg-white shadow-2xl dark:bg-slate-950"
          @click.stop
        >
          <div
            class="flex items-center justify-end border-b border-slate-200 bg-white px-4 py-3 dark:border-slate-800 dark:bg-slate-950"
          >
            <button
              class="inline-flex h-10 w-10 items-center justify-center rounded-lg text-slate-500 transition-colors hover:bg-slate-100 hover:text-slate-700 focus:outline-none focus:ring-2 focus:ring-primary/30 dark:hover:bg-slate-900 dark:hover:text-slate-200"
              type="button"
              aria-label="关闭导航菜单"
              @click="isMobileMenuOpen = false"
            >
              <Icon icon="material-symbols:close-rounded" class="text-2xl" />
            </button>
          </div>
          <div class="min-h-0 flex-1">
            <SideBar
              :nav-items="NavItems"
              :brand-title="'云盘'"
              :brand-subtitle="''"
              :compact="true"
            />
          </div>
        </div>
      </div>
    </Transition>
  </div>
</template>

<style lang="sass" scoped></style>
