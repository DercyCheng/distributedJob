import { defineStore } from 'pinia'
import { login, getUserInfo, refreshToken as apiRefreshToken, UserInfo } from '@/api/auth'
import { setToken, getToken, removeToken } from '@/utils/token'

interface UserState {
  token: string
  userInfo: UserInfo
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
        // 后端返回 accessToken，而不是 token
        const accessToken = data.accessToken || data.token || ''
        this.token = accessToken
        setToken(accessToken)
        
        // 获取用户信息
        const userId = data.userId ?? 0
        const userName = data.username ?? ''
        if (userId) {
          // 如果登录响应包含用户基本信息
          this.userInfo = {
            ...this.userInfo,
            id: userId,
            username: userName,
            name: data.realName || '',
            departmentId: data.departmentId || 0,
            roleId: data.roleId || 0
          }
          // 获取完整用户信息（包括权限等）
          await this.fetchUserInfo()
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
        if (userInfo) {
          // Convert possibly undefined values to their default types
          this.userInfo = {
            id: userInfo.id ?? 0,
            username: userInfo.username ?? '',
            name: userInfo.name ?? '',
            email: userInfo.email ?? '',
            departmentId: userInfo.departmentId ?? 0,
            departmentName: userInfo.departmentName ?? '',
            roleId: userInfo.roleId ?? 0,
            roleName: userInfo.roleName ?? '',
            permissions: userInfo.permissions ?? []
          }
          return this.userInfo
        }
        return this.userInfo
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
        
        const response = await apiRefreshToken()
        const newToken = response.accessToken || response.token || ''
        
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