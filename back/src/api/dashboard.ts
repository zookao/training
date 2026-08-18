import request from './request'
import type { ApiResult } from './request'

export interface RecentLearnItem {
  userId: number
  username: string
  nickname: string
  courseId: number
  courseTitle: string
  percent: number
  completed: boolean
  lastAt: string
}

export interface DashboardData {
  classCount: number
  activeClass: number
  courseCount: number
  activeCourse: number
  studentCount: number
  activeStudent: number
  recordCount: number
  completedCount: number
  avgProgress: number
  recentRecords: RecentLearnItem[]
}

export const getDashboard = () =>
  request.get<any, ApiResult<DashboardData>>('/dashboard')
