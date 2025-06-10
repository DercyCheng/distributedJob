import api from './index'

// 用户登录
export const login = async (credentials) => {
    return await api.post('/auth/login', credentials)
}

// 用户登出
export const logout = async () => {
    return await api.post('/auth/logout')
}

// 获取当前用户信息
export const getCurrentUser = async () => {
    return await api.get('/auth/me')
}

// 更新用户个人信息
export const updateProfile = async (profileData) => {
    return await api.put('/auth/profile', profileData)
}

// 修改密码
export const changePassword = async (passwordData) => {
    return await api.put('/auth/password', passwordData)
}

// 刷新token
export const refreshToken = async () => {
    return await api.post('/auth/refresh')
}
