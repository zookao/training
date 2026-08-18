import request from './request'
import type { ApiResult } from './request'

export interface ExamTestpaperItem {
  id: number
  name: string
  description: string
  type: number // 1随时考 2课程完成后考
  totalScore: number
  passScore: number
  duration: number // 分钟
  available: boolean
  hasFinished?: boolean // 是否有已完成的考试记录
  bestScore?: number // 最高分（hasFinished 为 true 时有效）
  bestPassed?: boolean // 最高分那次是否及格
  inProgress?: boolean // 是否有进行中的考试（未交卷）
}

export interface ExamQuestion {
  id: number
  type: number // 1单选 2多选 3判断
  title: string
  options: string // JSON
  score: number
  sort: number
}

export interface DraftAnswer {
  questionId: number
  userAnswer: string[]
}

export interface ExamDetail {
  id: number
  courseId: number
  name: string
  description: string
  type: number
  totalScore: number
  passScore: number
  duration: number
  remainSec: number // 剩余秒数
  questions: ExamQuestion[]
  draftAnswers: DraftAnswer[] // 草稿答案（断点续考用，新建时为空）
}

export interface AnswerDetail {
  questionId: number
  title: string
  type: number
  userAnswer: string[]
  correctAnswer: string[]
  correct: boolean
  score: number
  maxScore: number
}

export interface SubmitResult {
  score: number
  passed: boolean
  total: number
  passLine: number
  details: AnswerDetail[]
}

export interface ExamRecordItem {
  recordId: number
  testpaperId: number
  testpaperName: string
  score: number
  passed: boolean
  duration: number
  submittedAt: string
}

export const getCourseExam = (courseId: number) =>
  request.get<any, ApiResult<ExamTestpaperItem[]>>(`/user/course/${courseId}/exam`)

export const getExam = (testpaperId: number) =>
  request.get<any, ApiResult<ExamDetail>>(`/user/testpaper/${testpaperId}/exam`)

export const saveExamDraft = (testpaperId: number, answers: DraftAnswer[]) =>
  request.post<any, ApiResult<null>>(`/user/testpaper/${testpaperId}/draft`, { answers })

export const submitExam = (testpaperId: number, answers: { questionId: number; userAnswer: string[] }[]) =>
  request.post<any, ApiResult<SubmitResult>>(`/user/testpaper/${testpaperId}/submit`, { answers })

export const getExamRecords = (courseId: number) =>
  request.get<any, ApiResult<ExamRecordItem[]>>(`/user/course/${courseId}/exam/records`)
