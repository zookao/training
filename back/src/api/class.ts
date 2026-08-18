import request from './request'
import type { ApiResult } from './request'
import type { CourseItem } from './course'
import type { UserItem } from './user'

export interface ClassItem {
  id: number
  name: string
  cover: string
  description: string
  sort: number
  status: number
  courses: CourseItem[]
  users: UserItem[]
  createdAt: string
  updatedAt: string
}

export interface ClassForm {
  name: string
  cover: string
  description: string
  sort: number
  status: number
}

export const getClassList = (params: { page: number; pageSize: number; keyword?: string }) =>
  request.get<any, ApiResult<{ list: ClassItem[]; total: number; page: number; pageSize: number }>>('/class', { params })

export const getClassDetail = (id: number) =>
  request.get<any, ApiResult<ClassItem>>(`/class/${id}`)

export const createClass = (data: ClassForm) =>
  request.post<any, ApiResult>('/class', data)

export const updateClass = (id: number, data: ClassForm) =>
  request.put<any, ApiResult>(`/class/${id}`, data)

export const deleteClass = (id: number) =>
  request.delete<any, ApiResult>(`/class/${id}`)

export const getClassCourseIds = (id: number) =>
  request.get<any, ApiResult<number[]>>(`/class/${id}/courseIds`)

export const getClassUserIds = (id: number) =>
  request.get<any, ApiResult<number[]>>(`/class/${id}/userIds`)

export const assignClassCourses = (id: number, ids: number[]) =>
  request.put<any, ApiResult>(`/class/${id}/courses`, { ids })

export const assignClassUsers = (id: number, ids: number[]) =>
  request.put<any, ApiResult>(`/class/${id}/users`, { ids })

// 班级学习报告
export interface StudentReportItem {
  userId: number
  username: string
  nickname: string
  studentNo: string
  percent: number
}
export interface ClassLearningReport {
  classId: number
  className: string
  students: StudentReportItem[]
}
export interface VideoReportItem {
  videoId: number
  title: string
  percent: number
  completed: boolean
  duration: number
  maxPosition: number
}
export interface CourseReportItem {
  courseId: number
  title: string
  percent: number
  completedVideos: number
  videoCount: number
  videos: VideoReportItem[]
}
export interface StudentLearningDetail {
  userId: number
  username: string
  nickname: string
  courses: CourseReportItem[]
}
export const getClassLearningReport = (id: number) =>
  request.get<any, ApiResult<ClassLearningReport>>(`/class/${id}/learning-report`)
export const getStudentLearningDetail = (id: number, userId: number) =>
  request.get<any, ApiResult<StudentLearningDetail>>(`/class/${id}/learning-report/${userId}`)
