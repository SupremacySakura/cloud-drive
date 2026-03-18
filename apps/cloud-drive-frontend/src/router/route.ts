import type { RouteRecordRaw } from 'vue-router'

const routes: RouteRecordRaw[] = [
    {
        path: '/',
        redirect: '/home',
    },
    {
        path: '/home',
        component: () => import('../pages/HomePage.vue'),
        redirect: '/home/dashboard',
        children: []
    },
    {
        path: '/login',
        component: () => import('../pages/Login.vue'),
    },
    {
        path:'/register',
        component: () => import('../pages/Register.vue'),
    }
]
export {
    routes,
}
