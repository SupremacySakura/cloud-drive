<script setup lang="ts">
import { Icon } from '@iconify/vue'
import { useRouter } from 'vue-router'

const router = useRouter()

const goToLogin = () => {
    router.push({
        path: '/login',
        query: { redirect: router.currentRoute.value.fullPath }
    })
}
</script>

<template>
    <div class="flex flex-col items-center justify-center h-full min-h-[400px] p-8">
        <div class="flex flex-col items-center max-w-md text-center">
            <div class="w-20 h-20 mb-6 rounded-full bg-primary/10 flex items-center justify-center">
                <Icon icon="material-symbols:lock-outline" class="w-10 h-10 text-primary" />
            </div>
            <h2 class="text-2xl font-semibold text-gray-800 dark:text-gray-100 mb-2">
                登录后可用
            </h2>
            <p class="text-gray-500 dark:text-gray-400 mb-8">
                此功能需要登录后才能使用，登录后即可享受完整的云盘服务
            </p>
            <button
                @click="goToLogin"
                class="px-8 py-3 bg-primary hover:bg-primary/90 text-white font-medium rounded-lg transition-colors duration-200 focus:outline-none focus:ring-2 focus:ring-primary/50"
            >
                立即登录
            </button>
        </div>
    </div>
</template>
