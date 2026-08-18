import request from './request'
import type { ApiResult } from './request'

export interface QuestionItem {
  id: number
  courseId: number
  type: number // 1单选 2多选 3判断
  title: string
  options: string // JSON: [{"label":"A","content":"..."}]
  answer: string // JSON: ["A"] 或 ["A","C"]
  analysis: string
  sort: number
  status: number
  createdAt: string
  updatedAt: string
}

export interface QuestionForm {
  type: number
  title: string
  options: string
  answer: string
  analysis: string
  sort: number
  status: number
}

export interface QuestionOption {
  label: string
  content: string
}

export const getQuestionList = (courseId: number, params: { page: number; pageSize: number; keyword?: string }) =>
  request.get<any, ApiResult<{ list: QuestionItem[]; total: number; page: number; pageSize: number }>>(`/course/${courseId}/question`, { params })

export const getAllQuestions = (courseId: number) =>
  request.get<any, ApiResult<QuestionItem[]>>(`/course/${courseId}/question/all`)

export const createQuestion = (courseId: number, data: QuestionForm) =>
  request.post<any, ApiResult>(`/course/${courseId}/question`, data)

export const updateQuestion = (id: number, data: QuestionForm) =>
  request.put<any, ApiResult>(`/question/${id}`, data)

export const deleteQuestion = (id: number) =>
  request.delete<any, ApiResult>(`/question/${id}`)

export interface QuestionImportRow {
  row: number
  type: string
  title: string
  answer: string
  success: boolean
  reason: string
}

export interface QuestionImportResult {
  total: number
  success: number
  failed: number
  rows: QuestionImportRow[]
}

export const importQuestions = (courseId: number, file: File) => {
  const form = new FormData()
  form.append('file', file)
  return request.post<any, ApiResult<QuestionImportResult>>(`/course/${courseId}/question/import`, form, {
    headers: { 'Content-Type': 'multipart/form-data' },
    timeout: 120000
  })
}

export const downloadQuestionImportTemplate = async () => {
  const res = await request.get('/template/question-import', {
    responseType: 'blob'
  })
  const blob = new Blob([res as unknown as BlobPart], {
    type: 'application/vnd.openxmlformats-officedocument.spreadsheetml.sheet'
  })
  const url = URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url
  a.download = '试题导入模板.xlsx'
  a.click()
  URL.revokeObjectURL(url)
}
