import request from './request'
import type { ApiResult } from './request'

export interface AdminItem {
  id: number
  username: string
  nickname: string
  avatar: string
  email: string
  phone: string
  status: number
  lastLoginIp: string
  lastLoginAt: string
  roles: { id: number; name: string; title: string }[]
}

export interface AdminForm {
  username?: string
  password?: string
  nickname?: string
  email?: string
  phone?: string
  status?: number
  roleIds?: number[]
}

export const getAdminList = (params: { page: number; pageSize: number; keyword?: string }) =>
  request.get<any, ApiResult<{ list: AdminItem[]; total: number; page: number; pageSize: number }>>('/admin', { params })

export const getAdminDetail = (id: number) =>
  request.get<any, ApiResult<AdminItem>>(`/admin/${id}`)

export const createAdmin = (data: AdminForm) =>
  request.post<any, ApiResult>('/admin', data)

export const updateAdmin = (id: number, data: AdminForm) =>
  request.put<any, ApiResult>(`/admin/${id}`, data)

export const assignAdminRoles = (id: number, ids: number[]) =>
  request.put<any, ApiResult>(`/admin/${id}/roles`, { ids })

export const resetAdminPassword = (id: number, password: string) =>
  request.put<any, ApiResult>(`/admin/${id}/password`, { password })

export const deleteAdmin = (id: number) =>
  request.delete<any, ApiResult>(`/admin/${id}`)
