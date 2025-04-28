import http from './http'

export interface Department {
  id: number
  name: string
  parentId: number | null
  parentName: string
  status: 'active' | 'disabled'
  createdAt: string
  updatedAt: string
  children?: Department[]
}

export interface DepartmentQueryParams {
  name?: string
  status?: string
}

export function getDepartmentTree(params?: DepartmentQueryParams) {
  return http.get<any, Department[]>('/departments/tree', { params })
}

export function getDepartmentList(params?: DepartmentQueryParams) {
  return http.get<any, Department[]>('/departments', { params })
}

export function getDepartmentById(id: number) {
  return http.get<any, Department>(`/departments/${id}`)
}

export function createDepartment(data: Pick<Department, 'name' | 'parentId' | 'status'>) {
  return http.post<any, Department>('/departments', data)
}

export function updateDepartment(id: number, data: Pick<Department, 'name' | 'parentId' | 'status'>) {
  return http.put<any, Department>(`/departments/${id}`, data)
}

export function deleteDepartment(id: number) {
  return http.delete<any, void>(`/departments/${id}`)
}

export function batchUpdateDepartmentStatus(ids: number[], status: 'active' | 'disabled') {
  return http.put<any, void>('/departments/batch-status', { ids, status })
}