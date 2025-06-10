import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { login, logout, getCurrentUser, refreshToken } from '@/api/auth'
import { ElMessage } from 'element-plus'

export const useAuthStore = defineStore('auth', () => {
    // State
    const user = ref(null)
    const token = ref(localStorage.getItem('token') || null)
    const isLoading = ref(false)

    // Getters
    const isAuthenticated = computed(() => !!token.value && !!user.value)
    const userInfo = computed(() => user.value || {})

    // Actions
    const setToken = (newToken) => {
        token.value = newToken
        if (newToken) {
            localStorage.setItem('token', newToken)
        } else {
            localStorage.removeItem('token')
        }
    }

    const setUser = (userData) => {
        user.value = userData
    }

    const loginUser = async (credentials) => {
        try {
            isLoading.value = true
            const response = await login(credentials)

            if (response.token) {
                setToken(response.token)
                setUser(response.user)
                ElMessage.success('登录成功')
                return response
            }
        } catch (error) {
            ElMessage.error(error.response?.data?.error || '登录失败')
            throw error
        } finally {
            isLoading.value = false
        }
    }

    const logoutUser = async () => {
        try {
            await logout()
        } catch (error) {
            console.error('Logout error:', error)
        } finally {
            setToken(null)
            setUser(null)
            ElMessage.success('已退出登录')
        }
    }

    const fetchCurrentUser = async () => {
        if (!token.value) return

        try {
            const response = await getCurrentUser()
            setUser(response.user || response)
            return response
        } catch (error) {
            console.error('获取用户信息失败:', error)
            // 如果获取用户信息失败，可能是token过期，清除登录状态
            if (error.response?.status === 401) {
                setToken(null)
                setUser(null)
            }
            throw error
        }
    }

    const refreshUserToken = async () => {
        try {
            const response = await refreshToken()
            if (response.token) {
                setToken(response.token)
                return response
            }
        } catch (error) {
            console.error('刷新token失败:', error)
            setToken(null)
            setUser(null)
            throw error
        }
    }

    const initialize = async () => {
        if (token.value) {
            try {
                await fetchCurrentUser()
            } catch (error) {
                console.error('初始化用户信息失败:', error)
            }
        }
    }

    return {
        // State
        user,
        token,
        isLoading,

        // Getters
        isAuthenticated,
        userInfo,

        // Actions
        setToken,
        setUser,
        loginUser,
        logoutUser,
        fetchCurrentUser,
        refreshUserToken,
        initialize
    }
})
