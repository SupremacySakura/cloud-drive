<script setup lang="tsx">
import { Icon } from "@iconify/vue";
import { computed } from 'vue'
import { RouterLink, useRoute } from 'vue-router'

export type NavItem = {
    label: string
    icon: string
    to: string
}

const props = withDefaults(
    defineProps<{
        brandTitle?: string
        brandSubtitle?: string
        navItems?: NavItem[]
        storagePercent?: number
        storageDetail?: string
    }>(),
    {
        brandTitle: 'CloudDrive',
        brandSubtitle: 'Cloud Storage Plan',
        storagePercent: 65,
        storageDetail: '85.4 GB of 128 GB used',
    },
)

const route = useRoute()

const defaultNavItems: NavItem[] = [
    { label: 'Dashboard', icon:'material-symbols:dashboard', to: '/home/dashboard' },
    { label: 'File Manager', icon: 'material-symbols:folder-open', to: '/home/files' },
    { label: 'Pickup Codes', icon: 'material-symbols:key-outline', to: '/home/pickup-codes' },
    { label: 'Analytics', icon: 'material-symbols:bar-chart', to: '/home/analytics' },
    { label: 'Settings', icon: 'material-symbols:settings', to: '/home/settings' },
]

const items = computed(() => props.navItems ?? defaultNavItems)

const storageWidth = computed(() => {
    const percent = Number.isFinite(props.storagePercent) ? props.storagePercent : 0
    const clamped = Math.min(100, Math.max(0, percent))
    return `${clamped}%`
})

const normalizePath = (path: string) => {
    const normalized = path.replace(/\/+$/, '')
    return normalized || '/'
}

const isActive = (to: string) => {
    const current = normalizePath(route.path)
    const target = normalizePath(to)
    return current === target || (target !== '/' && current.startsWith(`${target}/`))
}

const linkClass = (to: string) => {
    const base =
        'flex items-center gap-3 px-3 py-2 rounded-lg transition-all group'
    return isActive(to)
        ? `${base} bg-primary/10 text-primary font-semibold`
        : `${base} text-slate-600 dark:text-slate-400 hover:bg-slate-50 dark:hover:bg-slate-900`
}

const iconClass = (to: string) => {
    const base = 'material-symbols-outlined text-[20px]'
    return isActive(to) ? `${base} fill-[1]` : `${base} group-hover:text-primary`
}
</script>

<template>
    <aside
        class="w-full lg:w-64 border-r border-slate-200 dark:border-slate-800 bg-white dark:bg-slate-950 p-6 flex flex-col justify-between">
        <div class="flex flex-col gap-6">
            <div class="flex flex-col">
                <h1 class="text-slate-900 dark:text-slate-100 text-base font-bold">
                    {{ brandTitle }}
                </h1>
                <p class="text-primary text-xs font-semibold uppercase tracking-wider">
                    {{ brandSubtitle }}
                </p>
            </div>

            <nav class="flex flex-col gap-1">
                <RouterLink v-for="item in items" :key="item.to" :to="item.to" :class="linkClass(item.to)">
                    <span :class="iconClass(item.to)"><Icon :icon="item.icon" /></span>
                    <span class="text-sm font-medium">{{ item.label }}</span>
                </RouterLink>
            </nav>
        </div>

        <div class="mt-10 p-4 bg-slate-50 dark:bg-slate-900 rounded-xl border border-slate-200 dark:border-slate-800">
            <p class="text-xs text-slate-500 mb-2">存储使用情况</p>
            <div class="h-2 w-full bg-slate-200 dark:bg-slate-700 rounded-full overflow-hidden">
                <div class="h-full bg-primary" :style="{ width: storageWidth }"></div>
            </div>
            <p class="text-[10px] mt-2 text-slate-500 font-medium">
                {{ storageDetail }}
            </p>
        </div>
    </aside>
</template>
