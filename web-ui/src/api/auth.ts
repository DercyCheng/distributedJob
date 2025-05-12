import http from './http'

interface LoginParams {
  username: string
  password: string
}

interface LoginResult {
  accessToken?: string  // 后端返回的是accessToken，不是token
  refreshToken?: string
  userId?: number
  username?: string
  realName?: string
  departmentId?: number
  roleId?: number
  tokenType?: string
  expiresIn?: number
  // 兼容可能的旧代码
  token?: string
  user?: {
    id?: number
    username?: string
    name?: string
    email?: string
    departmentId?: number
    departmentName?: string
    roleId?: number
    roleName?: string
    permissions?: string[]
  }
}

interface RefreshTokenResult {
  accessToken?: string
  refreshToken?: string
  token?: string // 兼容可能的旧代码
}

export function login(data: LoginParams) {
  return http.post<any, LoginResult>('/auth/login', data)
}

export function logout() {
  return http.post<any, any>('/auth/logout')
}

export interface UserInfo {
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

export function getUserInfo() {
  return http.get<any, UserInfo>('/auth/userinfo')
}

/**
 * 刷新令牌
 * @returns 新令牌
 */
export function refreshToken() {
  // 直接发送令牌，让后端从Authorization header中获取
  return http.post<any, RefreshTokenResult>('/auth/refresh', {})
}