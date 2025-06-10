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
        // 可以在这里添加认证 token
        // config.headers.Authorization = `Bearer ${getToken()}`
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
        ElMessage.error(message)
        return Promise.reject(error)
    }
)

export default api
