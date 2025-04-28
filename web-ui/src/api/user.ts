import http from './http'
import type { PageResult } from './task'

export interface User {
  id: number
  username: string
  name: string
  email: string
  departmentId: number
  departmentName: string
  roleId: number
  roleName: string
  status: 'active' | 'disabled'
  createdAt: string
  updatedAt: string
}

export interface UserQueryParams {
  page: number
  pageSize: number
  username?: string
  name?: string
  departmentId?: number
  roleId?: number
  status?: string
}

export function getUserList(params: UserQueryParams) {
  return http.get<any, PageResult<User>>('/users', { params })
}

export function getUserById(id: number) {
  return http.get<any, User>(`/users/${id}`)
}

export function createUser(data: Omit<User, 'id' | 'createdAt' | 'updatedAt' | 'departmentName' | 'roleName'> & { password: string }) {
  return http.post<any, User>('/users', data)
}

export function updateUser(id: number, data: Partial<Omit<User, 'id' | 'createdAt' | 'updatedAt' | 'departmentName' | 'roleName' | 'username'>> & { password?: string }) {
  return http.put<any, User>(`/users/${id}`, data)
}

export function deleteUser(id: number) {
  return http.delete<any, void>(`/users/${id}`)
}

export function resetPassword(id: number, password: string) {
  return http.put<any, void>(`/users/${id}/reset-password`, { password })
}

export function updateUserStatus(id: number, status: 'active' | 'disabled') {
  return http.put<any, User>(`/users/${id}/status`, { status })
}