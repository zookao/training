import request from './request'
import type { ApiResult } from './request'
import type { QuestionItem } from './question'

export interface TestpaperItem {
  id: number
  courseId: number
  name: string
  description: string
  type: number // 1随时考 2课程完成后考
  totalScore: number
  passScore: number
  duration: number // 分钟
  sort: number
  status: number
  createdAt: string
  updatedAt: string
}

export interface TestpaperForm {
  name: string
  description: string
  type: number
  totalScore: number
  passScore: number
  duration: number
  sort: number
  status: number
}

export interface TestpaperQuestionItem {
  id: number
  testpaperId: number
  questionId: number
  score: number
  sort: number
  question: QuestionItem
}

export interface TestpaperQuestionReq {
  questionId: number
  score: number
}

export const getTestpaperList = (courseId: number, params: { page: number; pageSize: number; keyword?: string }) =>
  request.get<any, ApiResult<{ list: TestpaperItem[]; total: number; page: number; pageSize: number }>>(`/course/${courseId}/testpaper`, { params })

export const createTestpaper = (courseId: number, data: TestpaperForm) =>
  request.post<any, ApiResult>(`/course/${courseId}/testpaper`, data)

export const updateTestpaper = (id: number, data: TestpaperForm) =>
  request.put<any, ApiResult>(`/testpaper/${id}`, data)

export const deleteTestpaper = (id: number) =>
  request.delete<any, ApiResult>(`/testpaper/${id}`)

export const getTestpaperQuestions = (id: number) =>
  request.get<any, ApiResult<TestpaperQuestionItem[]>>(`/testpaper/${id}/questions`)

export const setTestpaperQuestions = (id: number, items: TestpaperQuestionReq[]) =>
  request.put<any, ApiResult>(`/testpaper/${id}/questions`, { items })

// 考试报告
export interface ExamReportItem {
  recordId: number
  testpaperId: number
  testpaperName: string
  userId: number
  username: string
  nickname: string
  studentNo: string
  score: number
  passed: boolean
  duration: number
  startedAt: string
}

export const getExamReport = (courseId: number, keyword?: string) =>
  request.get<any, ApiResult<ExamReportItem[]>>(`/course/${courseId}/exam-report`, { params: { keyword } })

export const getExamRecordDetail = (recordId: number) =>
  request.get<any, ApiResult<any>>(`/exam-record/${recordId}`)
