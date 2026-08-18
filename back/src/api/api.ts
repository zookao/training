import request from './request'
import type { ApiResult } from './request'

export interface ApiItem {
  id: number
  group: string
  method: string
  path: string
  description: string
}

export interface ApiForm {
  group?: string
  method?: string
  path?: string
  description?: string
}

export const getApiList = (params: { page: number; pageSize: number; keyword?: string }) =>
  request.get<any, ApiResult<{ list: ApiItem[]; total: number; page: number; pageSize: number }>>('/api', { params })

export const getAllApis = () =>
  request.get<any, ApiResult<ApiItem[]>>('/api/all')

export const createApi = (data: ApiForm) =>
  request.post<any, ApiResult>('/api', data)

export const updateApi = (id: number, data: ApiForm) =>
  request.put<any, ApiResult>(`/api/${id}`, data)

export const deleteApi = (id: number) =>
  request.delete<any, ApiResult>(`/api/${id}`)
