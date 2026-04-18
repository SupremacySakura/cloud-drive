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
        meta: { requiresAuth: true },
        children: [
            {
                path: '/home/dashboard',
                component: () => import('../pages/Dashboard.vue'),
                meta: { requiresAuth: true },
            },
            {
                path: '/home/files',
                component: () => import('../pages/FileManagement.vue'),
                meta: { requiresAuth: true },
            },
            {
                path: '/home/pickup-codes',
                component: () => import('../pages/PickupCodes.vue'),
                meta: { requiresAuth: true },
            },
            {
                path: '/home/upload',
                component: () => import('../pages/UploadFile.vue'),
                meta: { requiresAuth: true },
            }
        ]
    },
    {
        path: '/login',
        component: () => import('../pages/Login.vue'),
    },
    {
        path: '/register',
        component: () => import('../pages/Register.vue'),
    },
    {
        path: '/require-login',
        component: () => import('../pages/RequireLogin.vue'),
    },
    {
        path: '/pickup',
        component: () => import('../pages/FilePickup.vue'),
    }
]
export {
    routes,
}
