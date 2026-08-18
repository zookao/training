import request from './request'
import type { ApiResult } from './request'

export interface LoginParams {
  username: string
  password: string
}

export interface UserInfo {
  id: number
  username: string
  nickname: string
  avatar: string
  roles: string[]
  perms: string[]
}

export interface MenuItem {
  id: number
  parentId: number
  name: string
  type: string
  path: string
  component: string
  redirect: string
  icon: string
  hidden: boolean
  keepAlive: boolean
  sort: number
  guardType: string
  perms: string
  children?: MenuItem[]
}

export const login = (data: LoginParams) =>
  request.post<any, ApiResult<{ token: string; expires: number }>>('/auth/login', data)

export const getUserInfo = () =>
  request.get<any, ApiResult<UserInfo>>('/auth/userinfo')

export const getMenus = () =>
  request.get<any, ApiResult<MenuItem[]>>('/auth/menus')

export const logout = () => request.post<any, ApiResult>('/auth/logout')
