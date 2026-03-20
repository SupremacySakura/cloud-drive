<script setup lang="ts">
import SideBar, { type NavItem } from '../components/bussiness/SideBar.vue'

const NavItems: NavItem[] = [
    { label: '仪表盘', icon:'material-symbols:dashboard', to: '/home/dashboard' },
    { label: '文件管理', icon: 'material-symbols:folder-open', to: '/home/files' },
    { label: '取件码', icon: 'material-symbols:key-outline', to: '/home/pickup-codes' },
    { label: '上传', icon: 'material-symbols:cloud-upload', to: '/home/upload' },
    { label: '分享', icon: 'material-symbols:share', to: '/home/share' },
]
</script>

<template>
    <div class="flex h-screen">
        <SideBar :nav-items="NavItems" :brand-title="'云盘'" :brand-subtitle="''"></SideBar>
        <router-view></router-view>
    </div>
</template>

<style lang="sass" scoped></style>