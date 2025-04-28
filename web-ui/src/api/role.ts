import http from './http'
import type { PageResult } from './task'

export interface Role {
  id: number
  name: string
  description: string
  status: 'active' | 'disabled'
  permissions: Permission[]
  createdAt: string
  updatedAt: string
}

export interface Permission {
  id: number
  code: string
  name: string
  description: string
  parentId: number | null
  children?: Permission[]
}

export interface RoleQueryParams {
  page: number
  pageSize: number
  name?: string
  status?: string
}

export function getRoleList(params: RoleQueryParams) {
  return http.get<any, PageResult<Role>>('/roles', { params })
}

export function getAllRoles() {
  return http.get<any, Role[]>('/roles/all')
}

export function getRoleById(id: number) {
  return http.get<any, Role>(`/roles/${id}`)
}

export function createRole(data: Omit<Role, 'id' | 'createdAt' | 'updatedAt'>) {
  return http.post<any, Role>('/roles', data)
}

export function updateRole(id: number, data: Partial<Omit<Role, 'id' | 'createdAt' | 'updatedAt'>>) {
  return http.put<any, Role>(`/roles/${id}`, data)
}

export function deleteRole(id: number) {
  return http.delete<any, void>(`/roles/${id}`)
}

export function getAllPermissions() {
  return http.get<any, Permission[]>('/permissions')
}

export function getPermissionTree() {
  return http.get<any, Permission[]>('/permissions/tree')
}