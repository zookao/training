import request from './request'
import type { ApiResult } from './request'

export interface VideoItem {
  id?: number
  url: string
  thumbnail: string
  courseware?: string
  coursewarePdf?: string // 课件 PDF URL（由 PPTX 转换生成，用于在线预览）
  coursewarePageCount?: number // 课件幻灯片页数
  coursewarePages?: string // 课件每页时长 JSON: [10,60,300]（秒）
  title: string
  description: string
  sort: number
  duration?: number
}

export interface CourseItem {
  id: number
  title: string
  cover: string
  description: string
  sort: number
  status: number
  videos: VideoItem[]
}

export interface CourseForm {
  title: string
  cover: string
  description: string
  sort: number
  status: number
  videos: VideoItem[]
}

export interface UploadVideoRes {
  url: string
  thumbnail: string
  courseware: string
  coursewarePdf: string
  coursewarePageCount: number
  filename: string
  duration: number
}

export interface UploadImageRes {
  url: string
}

export const getCourseList = (params: { page: number; pageSize: number; keyword?: string }) =>
  request.get<any, ApiResult<{ list: CourseItem[]; total: number; page: number; pageSize: number }>>('/course', { params })

export const getAllCourses = () =>
  request.get<any, ApiResult<CourseItem[]>>('/course/all')

export const getCourseDetail = (id: number) =>
  request.get<any, ApiResult<CourseItem>>(`/course/${id}`)

export const createCourse = (data: CourseForm) =>
  request.post<any, ApiResult>('/course', data)

export const updateCourse = (id: number, data: CourseForm) =>
  request.put<any, ApiResult>(`/course/${id}`, data)

export const deleteCourse = (id: number) =>
  request.delete<any, ApiResult>(`/course/${id}`)

export const deleteVideo = (courseId: number, videoId: number) =>
  request.delete<any, ApiResult>(`/course/${courseId}/video/${videoId}`)

// 重新识别课件页数（用于旧课件补录页数）
export const reparseCoursewarePages = (videoId: number) =>
  request.post<any, ApiResult<{ coursewarePageCount: number }>>(`/course/video/${videoId}/reparse`)

// 上传视频（含可选缩略图 + 可选课件）。大文件单独放宽超时 + 进度回调
export const uploadVideo = (
  video: File,
  thumbnail: File | null,
  courseware: File | null,
  onProgress?: (percent: number) => void
) => {
  const form = new FormData()
  form.append('video', video)
  if (thumbnail) form.append('thumbnail', thumbnail)
  if (courseware) form.append('courseware', courseware)
  return request.post<any, ApiResult<UploadVideoRes>>('/upload/video', form, {
    headers: { 'Content-Type': 'multipart/form-data' },
    timeout: 0, // 大文件上传不限时（有进度条可取消）
    onUploadProgress: (e) => {
      if (onProgress && e.total) {
        onProgress(Math.round((e.loaded * 100) / e.total))
      }
    }
  })
}

// 上传图片（课程/班级封面）
export const uploadImage = (file: File) => {
  const form = new FormData()
  form.append('file', file)
  return request.post<any, ApiResult<UploadImageRes>>('/upload/image', form, {
    headers: { 'Content-Type': 'multipart/form-data' },
    timeout: 60000
  })
}

// 导入课件打点（Excel），返回每页时长数组
export const importDurations = (file: File, duration: number) => {
  const form = new FormData()
  form.append('file', file)
  form.append('duration', String(duration))
  return request.post<any, ApiResult<{ durations: number[] }>>('/course/video/import-durations', form, {
    headers: { 'Content-Type': 'multipart/form-data' },
    timeout: 30000
  })
}

// 下载课件打点导入模板
export const downloadDurationTemplate = async () => {
  const res = await request.get('/course/video/duration-template', {
    responseType: 'blob'
  })
  const blob = new Blob([res as unknown as BlobPart], {
    type: 'application/vnd.openxmlformats-officedocument.spreadsheetml.sheet'
  })
  const url = URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url
  a.download = '课件打点模板.xlsx'
  a.click()
  URL.revokeObjectURL(url)
}
