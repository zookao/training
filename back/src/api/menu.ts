import request from './request'
import type { ApiResult } from './request'
import type { MenuItem } from './auth'

export interface MenuForm {
  parentId?: number
  name: string
  type: string
  path?: string
  component?: string
  redirect?: string
  icon?: string
  hidden?: boolean
  keepAlive?: boolean
  sort?: number
  guardType?: string
  perms?: string
}

export const getMenuTree = (guardType = 'admin') =>
  request.get<any, ApiResult<MenuItem[]>>('/menu', { params: { guardType } })

export const getMenuDetail = (id: number) =>
  request.get<any, ApiResult<MenuItem>>(`/menu/${id}`)

export const createMenu = (data: MenuForm) =>
  request.post<any, ApiResult>('/menu', data)

export const updateMenu = (id: number, data: MenuForm) =>
  request.put<any, ApiResult>(`/menu/${id}`, data)

export const deleteMenu = (id: number) =>
  request.delete<any, ApiResult>(`/menu/${id}`)
