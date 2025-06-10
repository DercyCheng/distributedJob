import axios from 'axios'
import { ElMessage } from 'element-plus'

// 创建 axios 实例
const api = axios.create({
    baseURL: '/api/v1',
    timeout: 10000
})

// 请求拦截器
api.interceptors.request.use(
    config => {
        // 添加认证token
        const token = localStorage.getItem('token')
        if (token) {
            config.headers.Authorization = `Bearer ${token}`
        }
        return config
    },
    error => {
        return Promise.reject(error)
    }
)

// 响应拦截器
api.interceptors.response.use(
    response => {
        return response.data
    },
    error => {
        const message = error.response?.data?.error || error.message || '请求失败'

        // 处理401未授权错误
        if (error.response?.status === 401) {
            // 清除本地存储的token
            localStorage.removeItem('token')
            // 如果不是登录页面，跳转到登录页面
            if (window.location.pathname !== '/login') {
                window.location.href = '/login'
            }
        } else {
            ElMessage.error(message)
        }

        return Promise.reject(error)
    }
)

export default api
