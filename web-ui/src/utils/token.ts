// Token Utility
import { refreshToken } from '@/api/auth'
import router from '@/router'
import { ElMessage } from 'element-plus'

// 令牌在localStorage中的键名
const TOKEN_KEY = 'token'

// 刷新令牌的阈值时间（毫秒）
// 当令牌剩余有效期小于这个值时，将自动刷新令牌
const REFRESH_THRESHOLD = 10 * 60 * 1000 // 10分钟

// 刷新令牌的冷却时间（毫秒）
// 防止频繁刷新令牌
const REFRESH_COOLDOWN = 5 * 60 * 1000 // 5分钟

// 上次刷新令牌的时间
let lastRefreshTime = 0

/**
 * 解析JWT令牌
 * @param token JWT令牌
 * @returns 解析后的令牌内容
 */
export function parseToken(token: string) {
  try {
    const base64Url = token.split('.')[1]
    const base64 = base64Url.replace(/-/g, '+').replace(/_/g, '/')
    const jsonPayload = decodeURIComponent(
      window
        .atob(base64)
        .split('')
        .map(c => '%' + ('00' + c.charCodeAt(0).toString(16)).slice(-2))
        .join('')
    )
    return JSON.parse(jsonPayload)
  } catch (error) {
    console.error('Failed to parse token:', error)
    return null
  }
}

/**
 * 获取令牌
 * @returns 当前令牌
 */
export function getToken(): string {
  return localStorage.getItem(TOKEN_KEY) || ''
}

/**
 * 设置令牌
 * @param token 令牌
 */
export function setToken(token: string): void {
  localStorage.setItem(TOKEN_KEY, token)
}

/**
 * 移除令牌
 */
export function removeToken(): void {
  localStorage.removeItem(TOKEN_KEY)
}

/**
 * 获取令牌过期时间
 * @param token JWT令牌
 * @returns 过期时间戳（毫秒）
 */
export function getTokenExpiration(token: string): number {
  const payload = parseToken(token)
  if (!payload || !payload.exp) return 0
  return payload.exp * 1000 // 转换为毫秒
}

/**
 * 检查令牌是否需要刷新
 * @param token JWT令牌
 * @returns 是否需要刷新
 */
export function shouldRefreshToken(token: string): boolean {
  if (!token) return false
  
  const now = Date.now()
  const expiration = getTokenExpiration(token)
  
  // 如果令牌已过期，直接返回true
  if (expiration <= now) return true
  
  // 如果距离上次刷新时间不足冷却时间，不刷新
  if (now - lastRefreshTime < REFRESH_COOLDOWN) return false
  
  // 如果令牌剩余有效期小于阈值，需要刷新
  return expiration - now < REFRESH_THRESHOLD
}

/**
 * 自动刷新令牌
 * @returns Promise<boolean> 是否成功刷新
 */
export async function autoRefreshToken(): Promise<boolean> {
  const token = getToken()
  
  if (!token || !shouldRefreshToken(token)) {
    return false
  }
  
  try {
    const response = await refreshToken()
    // 后端可能返回token或accessToken
    const newToken = response.accessToken || response.token || ''
    
    if (newToken) {
      setToken(newToken)
      lastRefreshTime = Date.now()
      return true
    }
    
    return false
  } catch (error) {
    console.error('Failed to refresh token:', error)
    
    // 如果令牌已过期，强制登出
    if (getTokenExpiration(token) <= Date.now()) {
      handleTokenExpired()
    }
    
    return false
  }
}

/**
 * 处理令牌过期
 */
export function handleTokenExpired(): void {
  removeToken()
  ElMessage.error('登录已过期，请重新登录')
  
  // 跳转到登录页
  if (router.currentRoute.value.path !== '/login') {
    router.push({
      path: '/login',
      query: { redirect: router.currentRoute.value.fullPath }
    })
  }
}