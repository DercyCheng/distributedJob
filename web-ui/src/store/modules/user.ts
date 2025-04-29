import { defineStore } from 'pinia'
import { login, getUserInfo, refreshToken as apiRefreshToken } from '@/api/auth'
import { setToken, getToken, removeToken } from '@/utils/token'

interface UserState {
  token: string
  userInfo: {
    id: number
    username: string
    name: string
    email: string
    departmentId: number
    departmentName: string
    roleId: number
    roleName: string
    permissions: string[]
  }
}

export const useUserStore = defineStore('user', {
  state: (): UserState => ({
    token: getToken(),
    userInfo: {
      id: 0,
      username: '',
      name: '',
      email: '',
      departmentId: 0,
      departmentName: '',
      roleId: 0,
      roleName: '',
      permissions: []
    }
  }),
  
  getters: {
    isLoggedIn: (state) => !!state.token,
    hasPermission: (state) => (permission: string) => {
      return state.userInfo.permissions.includes(permission)
    }
  },
  
  actions: {
    // 登录
    async login(username: string, password: string) {
      try {
        const data = await login({ username, password })
        this.token = data.token
        setToken(data.token)
        
        // 获取用户信息
        if (data.user && data.user.id) {
          this.userInfo = data.user
        } else {
          // If user info is not included in login response, fetch it separately
          await this.fetchUserInfo()
        }
        
        return data
      } catch (error) {
        console.error('Login failed:', error)
        throw error
      }
    },
    
    // 获取用户信息
    async fetchUserInfo() {
      try {
        const userInfo = await getUserInfo()
        this.userInfo = userInfo
        return userInfo
      } catch (error) {
        console.error('Failed to fetch user info:', error)
        throw error
      }
    },
    
    // 刷新令牌
    async refreshToken() {
      try {
        const currentToken = this.token || getToken()
        if (!currentToken) {
          throw new Error('No token to refresh')
        }
        
        const response = await apiRefreshToken(currentToken)
        const newToken = response.token
        
        if (newToken) {
          this.token = newToken
          setToken(newToken)
          return true
        }
        
        return false
      } catch (error) {
        console.error('Failed to refresh token:', error)
        throw error
      }
    },
    
    // 登出
    logout() {
      this.token = ''
      this.userInfo = {
        id: 0,
        username: '',
        name: '',
        email: '',
        departmentId: 0,
        departmentName: '',
        roleId: 0,
        roleName: '',
        permissions: []
      }
      removeToken()
    }
  }
})