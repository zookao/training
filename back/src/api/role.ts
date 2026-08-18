import request from './request'
import type { ApiResult } from './request'

export interface RoleItem {
  id: number
  name: string
  title: string
  guardType: string
  sort: number
  status: number
  remark: string
}

export interface RoleForm {
  name: string
  title: string
  guardType?: string
  sort?: number
  status?: number
  remark?: string
}

export const getRoleList = (params: { page: number; pageSize: number; keyword?: string; guardType?: string }) =>
  request.get<any, ApiResult<{ list: RoleItem[]; total: number; page: number; pageSize: number }>>('/role', { params })

export const getAllRoles = (guardType = 'admin') =>
  request.get<any, ApiResult<RoleItem[]>>('/role/all', { params: { guardType } })

export const getRoleDetail = (id: number) =>
  request.get<any, ApiResult<RoleItem>>(`/role/${id}`)

export const getRoleMenuIds = (id: number) =>
  request.get<any, ApiResult<number[]>>(`/role/${id}/menuIds`)

export const getRoleApiIds = (id: number) =>
  request.get<any, ApiResult<number[]>>(`/role/${id}/apiIds`)

export const createRole = (data: RoleForm) =>
  request.post<any, ApiResult>('/role', data)

export const updateRole = (id: number, data: RoleForm) =>
  request.put<any, ApiResult>(`/role/${id}`, data)

export const deleteRole = (id: number) =>
  request.delete<any, ApiResult>(`/role/${id}`)

export const assignRoleMenus = (id: number, ids: number[]) =>
  request.put<any, ApiResult>(`/role/${id}/menus`, { ids })

export const assignRoleApis = (id: number, ids: number[]) =>
  request.put<any, ApiResult>(`/role/${id}/apis`, { ids })
