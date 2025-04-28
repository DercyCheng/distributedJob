import http from './http'

export interface Task {
  id: number
  name: string
  description: string
  type: 'http' | 'grpc'
  status: 'active' | 'paused' | 'deleted'
  cronExpression: string
  departmentId: number
  departmentName: string
  createdBy: number
  creatorName: string
  createdAt: string
  updatedAt: string
  config: HttpTaskConfig | GrpcTaskConfig
}

export interface HttpTaskConfig {
  url: string
  method: 'GET' | 'POST' | 'PUT' | 'DELETE'
  headers: Record<string, string>
  body: string
  timeout: number
  successCodes: number[]
}

export interface GrpcTaskConfig {
  host: string
  port: number
  service: string
  method: string
  request: string
  timeout: number
}

export interface TaskQueryParams {
  page: number
  pageSize: number
  name?: string
  type?: string
  status?: string
  departmentId?: number
  startTime?: string
  endTime?: string
}

export interface PageResult<T> {
  list: T[]
  total: number
  page: number
  pageSize: number
}

export function getTaskList(params: TaskQueryParams) {
  return http.get<any, PageResult<Task>>('/tasks', { params })
}

export function getTaskById(id: number) {
  return http.get<any, Task>(`/tasks/${id}`)
}

export function createTask(data: Omit<Task, 'id' | 'createdAt' | 'updatedAt' | 'createdBy' | 'creatorName' | 'departmentName'>) {
  return http.post<any, Task>('/tasks', data)
}

export function updateTask(id: number, data: Partial<Omit<Task, 'id' | 'createdAt' | 'updatedAt' | 'createdBy' | 'creatorName' | 'departmentName'>>) {
  return http.put<any, Task>(`/tasks/${id}`, data)
}

export function deleteTask(id: number) {
  return http.delete<any, void>(`/tasks/${id}`)
}

export function pauseTask(id: number) {
  return http.put<any, Task>(`/tasks/${id}/pause`, {})
}

export function resumeTask(id: number) {
  return http.put<any, Task>(`/tasks/${id}/resume`, {})
}

export function executeTask(id: number) {
  return http.post<any, void>(`/tasks/${id}/execute`, {})
}