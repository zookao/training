import request from './request'
import type { ApiResult } from './request'

export interface ClassProgressItem {
  id: number
  name: string
  cover: string
  description: string
  courseCount: number
  percent: number
}

export interface CourseProgressItem {
  id: number
  title: string
  cover: string
  description: string
  videoCount: number
  completedVideos: number
  percent: number
}

export interface ClassDetailRes {
  id: number
  name: string
  description: string
  courses: CourseProgressItem[]
  percent: number
}

export interface VideoLearnItem {
  id: number
  url: string
  thumbnail: string
  courseware: string
  coursewarePdf: string // 课件 PDF URL（由 PPTX 转换生成，用于在线预览）
  coursewarePages: string // 课件每页时长 JSON: [10,60,300]（秒），用于视频播放联动翻页
  title: string
  description: string
  sort: number
  duration: number
  position: number
  maxPosition: number // 当前最大播放位置（秒）
  percent: number
  completed: boolean
  nextCheckPosition: number // 下次校验触发的视频位置（秒），0=不校验
  checkPending: boolean // 是否有待完成的滑动校验
}

export interface CourseLearnRes {
  id: number
  title: string
  cover: string
  description: string
  videos: VideoLearnItem[]
}

export interface ProgressReq {
  videoId: number
  courseId: number
  classId?: number
  position: number
  duration?: number // 兜底用：仅旧视频无服务端时长时使用
  completed: boolean
}

export interface ProgressRes {
  percent: number
  completed: boolean
  maxPosition: number
  checkPending: boolean
  nextCheckPosition: number
}

export const getMyClasses = () =>
  request.get<any, ApiResult<ClassProgressItem[]>>('/user/classes')

export const getClassDetail = (id: number) =>
  request.get<any, ApiResult<ClassDetailRes>>(`/user/classes/${id}`)

export const getCourseLearn = (id: number) =>
  request.get<any, ApiResult<CourseLearnRes>>(`/user/course/${id}`)

export const reportProgress = (data: ProgressReq) =>
  request.post<any, ApiResult<ProgressRes>>('/user/progress', data)

export const checkPass = (videoId: number) =>
  request.post<any, ApiResult<{ nextCheckPosition: number }>>(`/user/video/${videoId}/check`)
