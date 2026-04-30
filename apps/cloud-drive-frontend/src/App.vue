<script setup lang="ts">
// Skip to content 功能：按下 Enter 跳转到主内容区域
const skipToContent = () => {
  const mainContent = document.querySelector('main')
  if (mainContent) {
    mainContent.setAttribute('tabindex', '-1')
    mainContent.focus()
  }
}
</script>

<template>
  <div>
    <!-- Skip to content 链接 - 为键盘用户提供快速导航到主内容 -->
    <a
      href="#main-content"
      class="sr-only focus:not-sr-only focus:absolute focus:top-2 focus:left-2 focus:z-50 focus:bg-primary focus:text-white focus:px-4 focus:py-2 focus:rounded-lg focus:font-semibold"
      @click.prevent="skipToContent"
    >
      跳转到主内容
    </a>
    <router-view></router-view>
  </div>
</template>
