import http from './http'

interface LoginParams {
  username: string
  password: string
}

interface LoginResult {
  token: string
  user: {
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

interface RefreshTokenResult {
  token: string
}

export function login(data: LoginParams) {
  return http.post<any, LoginResult>('/auth/login', data)
}

export function logout() {
  return http.post<any, any>('/auth/logout')
}

export function getUserInfo() {
  return http.get<any, LoginResult['user']>('/auth/userinfo')
}

/**
 * 刷新令牌
 * @param token 当前令牌
 * @returns 新令牌
 */
export function refreshToken(token: string) {
  return http.post<any, RefreshTokenResult>('/auth/refresh', { token })
}