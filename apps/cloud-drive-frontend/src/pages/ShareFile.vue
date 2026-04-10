<script setup lang="ts">
import { useUserStore } from '../stores/user'
import LoginRequiredPlaceholder from '../components/bussiness/LoginRequiredPlaceholder.vue'

const userStore = useUserStore()
</script>

<template>
    <div class="flex-1 h-full bg-background-light dark:bg-background-dark">
        <LoginRequiredPlaceholder v-if="!userStore.isLoggedIn" />
        <div v-else class="p-8">
            分享文件
        </div>
    </div>
</template>

<style lang="sass" scoped></style>