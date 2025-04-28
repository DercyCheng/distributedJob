import { defineStore } from 'pinia'
import { login, getUserInfo } from '@/api/auth'

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
    token: localStorage.getItem('token') || '',
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
    async login(username: string, password: string) {
      try {
        const data = await login({ username, password })
        this.token = data.token
        this.userInfo = data.user
        localStorage.setItem('token', data.token)
        return data
      } catch (error) {
        console.error('Login failed:', error)
        throw error
      }
    },
    
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
      localStorage.removeItem('token')
    }
  }
})