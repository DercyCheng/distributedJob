import { createRouter, createWebHistory } from 'vue-router'
import Dashboard from '@/views/Dashboard.vue'
import Jobs from '@/views/Jobs.vue'
import Executions from '@/views/Executions.vue'
import Workers from '@/views/Workers.vue'
import Logs from '@/views/Logs.vue'

const routes = [
    {
        path: '/',
        redirect: '/dashboard'
    },
    {
        path: '/dashboard',
        name: 'Dashboard',
        component: Dashboard
    },
    {
        path: '/jobs',
        name: 'Jobs',
        component: Jobs
    },
    {
        path: '/executions',
        name: 'Executions',
        component: Executions
    },
    {
        path: '/workers',
        name: 'Workers',
        component: Workers
    },
    {
        path: '/logs',
        name: 'Logs',
        component: Logs
    }
]

const router = createRouter({
    history: createWebHistory(),
    routes
})

export default router
