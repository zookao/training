import request from './request'
import type { ApiResult } from './request'

export interface DepartmentItem {
  id: number
  name: string
  description: string
  sort: number
  status: number
  createdAt: string
  updatedAt: string
}

export interface DepartmentForm {
  name: string
  description: string
  sort: number
  status: number
}

export const getDepartmentList = (params: { page: number; pageSize: number; keyword?: string }) =>
  request.get<any, ApiResult<{ list: DepartmentItem[]; total: number; page: number; pageSize: number }>>('/department', { params })

export const getAllDepartments = () =>
  request.get<any, ApiResult<DepartmentItem[]>>('/department/all')

export const createDepartment = (data: DepartmentForm) =>
  request.post<any, ApiResult>('/department', data)

export const updateDepartment = (id: number, data: DepartmentForm) =>
  request.put<any, ApiResult>(`/department/${id}`, data)

export const deleteDepartment = (id: number) =>
  request.delete<any, ApiResult>(`/department/${id}`)

// 院系学员
export interface DeptStudentItem {
  id: number
  username: string
  nickname: string
  studentNo: string
  phone: string
  status: number
}

export const getDepartmentStudents = (id: number) =>
  request.get<any, ApiResult<DeptStudentItem[]>>(`/department/${id}/students`)

export const removeDepartmentStudent = (id: number, userId: number) =>
  request.delete<any, ApiResult>(`/department/${id}/students/${userId}`)

export interface DeptImportRow {
  row: number
  phone: string
  name: string
  result: string
  reason: string
}

export interface DeptImportResult {
  total: number
  success: number
  failed: number
  rows: DeptImportRow[]
}

export const importDepartmentStudents = (id: number, file: File) => {
  const fd = new FormData()
  fd.append('file', file)
  return request.post<any, ApiResult<DeptImportResult>>(`/department/${id}/students/import`, fd)
}

export const downloadDeptStudentTemplate = async () => {
  const res = await request.get('/template/department-student-import', {
    responseType: 'blob'
  })
  const blob = new Blob([res as unknown as BlobPart], {
    type: 'application/vnd.openxmlformats-officedocument.spreadsheetml.sheet'
  })
  const url = URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url
  a.download = '院系学员导入模板.xlsx'
  a.click()
  URL.revokeObjectURL(url)
}
