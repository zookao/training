import request from './request'
import type { ApiResult } from './request'

export interface UserItem {
  id: number
  username: string
  nickname: string
  studentNo: string
  departmentId: number
  avatar: string
  email: string
  phone: string
  status: number
  lastLoginIp: string
  lastLoginAt: string
  createdAt: string
  updatedAt: string
}

export interface UserForm {
  username?: string
  password?: string
  nickname?: string
  studentNo?: string
  departmentId?: number
  email?: string
  phone?: string
  status?: number
}

export const getUserList = (params: { page: number; pageSize: number; keyword?: string; departmentId?: number }) =>
  request.get<any, ApiResult<{ list: UserItem[]; total: number; page: number; pageSize: number }>>('/user', { params })

export const getAllUsers = () =>
  request.get<any, ApiResult<UserItem[]>>('/user/all')

export const getUserDetail = (id: number) =>
  request.get<any, ApiResult<UserItem>>(`/user/${id}`)

export const createUser = (data: UserForm) =>
  request.post<any, ApiResult>('/user', data)

export const updateUser = (id: number, data: UserForm) =>
  request.put<any, ApiResult>(`/user/${id}`, data)

export const resetUserPassword = (id: number, password: string) =>
  request.put<any, ApiResult>(`/user/${id}/password`, { password })

export const deleteUser = (id: number) =>
  request.delete<any, ApiResult>(`/user/${id}`)

export interface UserImportRow {
  row: number
  username: string
  phone: string
  name: string
  studentNo: string
  department: string
  success: boolean
  reason: string
}

export interface UserImportResult {
  total: number
  success: number
  failed: number
  rows: UserImportRow[]
}

export const importUsers = (file: File) => {
  const form = new FormData()
  form.append('file', file)
  return request.post<any, ApiResult<UserImportResult>>('/user/import', form, {
    headers: { 'Content-Type': 'multipart/form-data' },
    timeout: 120000
  })
}

export const downloadUserImportTemplate = async () => {
  const res = await request.get('/template/user-import', {
    responseType: 'blob'
  })
  const blob = new Blob([res as unknown as BlobPart], {
    type: 'application/vnd.openxmlformats-officedocument.spreadsheetml.sheet'
  })
  const url = URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url
  a.download = '学员导入模板.xlsx'
  a.click()
  URL.revokeObjectURL(url)
}
