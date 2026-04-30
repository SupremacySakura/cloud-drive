<script setup lang="ts">
import SideBar, { type NavItem } from '../components/bussiness/SideBar.vue'

const NavItems: NavItem[] = [
  { label: '仪表盘', icon: 'material-symbols:dashboard', to: '/home/dashboard' },
  { label: '文件管理', icon: 'material-symbols:folder-open', to: '/home/files' },
  { label: '取件码', icon: 'material-symbols:key-outline', to: '/home/pickup-codes' },
  { label: '上传', icon: 'material-symbols:cloud-upload', to: '/home/upload' },
]
</script>

<template>
  <div class="flex h-screen">
    <!-- Skip to content link for accessibility -->
    <a
      href="#main-content"
      class="sr-only focus:not-sr-only focus:absolute focus:top-4 focus:left-4 focus:z-50 focus:px-4 focus:py-2 focus:bg-primary focus:text-white focus:rounded-lg focus:font-medium focus:outline-none focus:ring-2 focus:ring-primary/50"
    >
      跳转到主要内容
    </a>
    <SideBar :nav-items="NavItems" :brand-title="'云盘'" :brand-subtitle="''"></SideBar>
    <router-view id="main-content" class="flex-1" tabindex="-1"></router-view>
  </div>
</template>

<style lang="sass" scoped></style>
