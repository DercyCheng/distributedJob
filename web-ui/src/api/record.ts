import http from './http'
import type { PageResult } from './task'

export interface Record {
  id: number
  taskId: number
  taskName: string
  status: 'success' | 'fail'
  startTime: string
  endTime: string
  duration: number
  request: string
  response: string
  error: string
  departmentId: number
  departmentName: string
}

export interface RecordQueryParams {
  page: number
  pageSize: number
  taskId?: number
  status?: string
  departmentId?: number
  startTime?: string
  endTime?: string
}

export function getRecordList(params: RecordQueryParams) {
  return http.get<any, PageResult<Record>>('/records', { params })
}

export function getRecordById(id: number) {
  return http.get<any, Record>(`/records/${id}`)
}

export function getTaskRecords(taskId: number, params: Omit<RecordQueryParams, 'taskId'>) {
  return http.get<any, PageResult<Record>>(`/tasks/${taskId}/records`, { params })
}

export function exportRecords(params: RecordQueryParams) {
  return http.get<any, Blob>('/records/export', { 
    params,
    responseType: 'blob'
  })
}