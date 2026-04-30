<script setup lang="ts">
import { RouterLink, useRoute } from 'vue-router'

const route = useRoute()
const redirectPath = typeof route.query.redirect === 'string' ? route.query.redirect : '/home'
</script>

<template>
  <div
    class="bg-background-light dark:bg-background-dark font-display text-slate-900 dark:text-slate-100 min-h-screen flex flex-col"
  >
    <header
      class="w-full flex items-center justify-between px-6 py-4 lg:px-10 border-b border-primary/10 bg-white dark:bg-background-dark/50 backdrop-blur-sm sticky top-0 z-50"
    >
      <div class="flex items-center gap-2">
        <div class="text-primary flex items-center justify-center">
          <svg class="w-8 h-8" fill="none" viewBox="0 0 48 48" xmlns="http://www.w3.org/2000/svg">
            <path
              d="M13.8261 17.4264C16.7203 18.1174 20.2244 18.5217 24 18.5217C27.7756 18.5217 31.2797 18.1174 34.1739 17.4264C36.9144 16.7722 39.9967 15.2331 41.3563 14.1648L24.8486 40.6391C24.4571 41.267 23.5429 41.267 23.1514 40.6391L6.64374 14.1648C8.00331 15.2331 11.0856 16.7722 13.8261 17.4264Z"
              fill="currentColor"
            />
            <path
              clip-rule="evenodd"
              d="M39.998 12.236C39.9944 12.2537 39.9875 12.2845 39.9748 12.3294C39.9436 12.4399 39.8949 12.5741 39.8346 12.7175C39.8168 12.7597 39.7989 12.8007 39.7813 12.8398C38.5103 13.7113 35.9788 14.9393 33.7095 15.4811C30.9875 16.131 27.6413 16.5217 24 16.5217C20.3587 16.5217 17.0125 16.131 14.2905 15.4811C12.0012 14.9346 9.44505 13.6897 8.18538 12.8168C8.17384 12.7925 8.16216 12.767 8.15052 12.7408C8.09919 12.6249 8.05721 12.5114 8.02977 12.411C8.00356 12.3152 8.00039 12.2667 8.00004 12.2612C8.00004 12.261 8 12.2607 8.00004 12.2612C8.00004 12.2359 8.0104 11.9233 8.68485 11.3686C9.34546 10.8254 10.4222 10.2469 11.9291 9.72276C14.9242 8.68098 19.1919 8 24 8C28.8081 8 33.0758 8.68098 36.0709 9.72276C37.5778 10.2469 38.6545 10.8254 39.3151 11.3686C39.9006 11.8501 39.9857 12.1489 39.998 12.236ZM4.95178 15.2312L21.4543 41.6973C22.6288 43.5809 25.3712 43.5809 26.5457 41.6973L43.0534 15.223C43.0709 15.1948 43.0878 15.1662 43.104 15.1371L41.3563 14.1648C43.104 15.1371 43.1038 15.1374 43.104 15.1371L43.1051 15.135L43.1065 15.1325L43.1101 15.1261L43.1199 15.1082C43.1276 15.094 43.1377 15.0754 43.1497 15.0527C43.1738 15.0075 43.2062 14.9455 43.244 14.8701C43.319 14.7208 43.4196 14.511 43.5217 14.2683C43.6901 13.8679 44 13.0689 44 12.2609C44 10.5573 43.003 9.22254 41.8558 8.2791C40.6947 7.32427 39.1354 6.55361 37.385 5.94477C33.8654 4.72057 29.133 4 24 4C18.867 4 14.1346 4.72057 10.615 5.94478C8.86463 6.55361 7.30529 7.32428 6.14419 8.27911C4.99695 9.22255 3.99999 10.5573 3.99999 12.2609C3.99999 13.1275 4.29264 13.9078 4.49321 14.3607C4.60375 14.6102 4.71348 14.8196 4.79687 14.9689C4.83898 15.0444 4.87547 15.1065 4.9035 15.1529C4.91754 15.1762 4.92954 15.1957 4.93916 15.2111L4.94662 15.223L4.95178 15.2312ZM35.9868 18.996L24 38.22L12.0131 18.996C12.4661 19.1391 12.9179 19.2658 13.3617 19.3718C16.4281 20.1039 20.0901 20.5217 24 20.5217C27.9099 20.5217 31.5719 20.1039 34.6383 19.3718C35.082 19.2658 35.5339 19.1391 35.9868 18.996Z"
              fill="currentColor"
              fill-rule="evenodd"
            />
          </svg>
        </div>
        <span class="text-xl font-bold tracking-tight text-slate-900 dark:text-slate-100"
          >云盘 -by supremacy</span
        >
      </div>
    </header>

    <main class="flex-1 flex flex-col items-center justify-center px-4 py-12">
      <div
        class="w-full max-w-md bg-white dark:bg-slate-900/50 rounded-xl shadow-xl shadow-primary/5 border border-primary/10 overflow-hidden"
      >
        <div class="p-8">
          <div class="flex flex-col items-center mb-8">
            <!-- 锁图标 -->
            <div class="w-16 h-16 mb-4 text-primary">
              <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="currentColor">
                <path
                  fill-rule="evenodd"
                  d="M12 2C9.23858 2 7 4.23858 7 7V8H5C3.34315 8 2 9.34315 2 11V19C2 20.6569 3.34315 22 5 22H19C20.6569 22 22 20.6569 22 19V11C22 9.34315 20.6569 8 19 8H17V7C17 4.23858 14.7614 2 12 2ZM9 7C9 5.34315 10.3431 4 12 4C13.6569 4 15 5.34315 15 7V8H9V7ZM5 10H19C19.5523 10 20 10.4477 20 11V19C20 19.5523 19.5523 20 19 20H5C4.44772 20 4 19.5523 4 19V11C4 10.4477 4.44772 10 5 10Z"
                  clip-rule="evenodd"
                />
              </svg>
            </div>
            <h1 class="text-2xl font-black text-slate-900 dark:text-slate-100 mb-2">需要登录</h1>
            <p class="text-slate-500 dark:text-slate-400 text-center">此页面需要登录后才能访问。</p>
          </div>

          <div class="space-y-4">
            <RouterLink
              :to="{ path: '/login', query: { redirect: redirectPath } }"
              class="w-full h-12 bg-primary hover:bg-primary/90 text-white font-bold rounded-lg shadow-lg shadow-primary/20 transition-all transform active:scale-[0.98] flex items-center justify-center"
            >
              立即登录
            </RouterLink>

            <RouterLink
              to="/register"
              class="w-full h-12 border-2 border-primary text-primary hover:bg-primary/10 font-bold rounded-lg transition-all transform active:scale-[0.98] flex items-center justify-center"
            >
              注册新账号
            </RouterLink>
          </div>
        </div>

        <div
          class="bg-slate-50 dark:bg-slate-800/50 p-6 text-center border-t border-slate-100 dark:border-slate-800"
        >
          <p class="text-slate-600 dark:text-slate-400 text-sm">
            没有账号？注册后即可访问所有功能。
          </p>
        </div>
      </div>
    </main>
  </div>
</template>

<style scoped></style>
