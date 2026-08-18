import request from './request'
import type { ApiResult } from './request'

export interface LoginParams {
  username: string
  password: string
}

export interface RegisterParams {
  username: string
  password: string
  phone: string
  nickname?: string
  studentNo?: string
  departmentId?: number
}

export interface DepartmentItem {
  id: number
  name: string
  description: string
  sort: number
  status: number
}

export interface UserInfo {
  id: number
  username: string
  nickname: string
  studentNo: string
  departmentId: number
  avatar: string
  email: string
  phone: string
}

export interface UpdateProfileParams {
  nickname?: string
  studentNo?: string
  avatar?: string
  email?: string
  phone?: string
}

export interface ChangePwdParams {
  oldPassword: string
  newPassword: string
}

export const register = (data: RegisterParams) =>
  request.post<any, ApiResult>('/user/auth/register', data)

export const login = (data: LoginParams) =>
  request.post<any, ApiResult<{ token: string; expires: number }>>('/user/auth/login', data)

export const getUserInfo = () =>
  request.get<any, ApiResult<UserInfo>>('/user/auth/userinfo')

export const updateProfile = (data: UpdateProfileParams) =>
  request.put<any, ApiResult>('/user/auth/profile', data)

export const changePassword = (data: ChangePwdParams) =>
  request.put<any, ApiResult>('/user/auth/password', data)

export const logout = () =>
  request.post<any, ApiResult>('/user/auth/logout')

export const getAllDepartments = () =>
  request.get<any, ApiResult<DepartmentItem[]>>('/user/department/all')
