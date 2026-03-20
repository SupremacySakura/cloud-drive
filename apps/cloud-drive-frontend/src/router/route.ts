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
        children: [
            {
                path: '/home/dashboard',
                component: () => import('../pages/Dashboard.vue'),
            },
            {
                path: '/home/files',
                component: () => import('../pages/FileManagement.vue'),
            },
            {
                path: '/home/pickup-codes',
                component: () => import('../pages/PickupCodes.vue'),
            },
            {
                path: '/home/upload',
                component: () => import('../pages/UploadFile.vue'),
            },
            {
                path: '/home/share',
                component: () => import('../pages/ShareFile.vue'),
            },
        ]
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
