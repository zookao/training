import axios, { type AxiosInstance, type InternalAxiosRequestConfig, type AxiosResponse } from 'axios'
import { ElMessage } from 'element-plus'
import { getToken, removeToken } from '@/utils/auth'

export interface ApiResult<T = any> {
  code: number
  msg: string
  data: T
}

const service: AxiosInstance = axios.create({
  baseURL: '/api/admin',
  timeout: 15000
})

// 请求拦截：附 token
service.interceptors.request.use(
  (config: InternalAxiosRequestConfig) => {
    const token = getToken()
    if (token) {
      config.headers.Authorization = `Bearer ${token}`
    }
    return config
  },
  (error) => Promise.reject(error)
)

// 响应拦截：统一处理
service.interceptors.response.use(
  (response: AxiosResponse<ApiResult>) => {
    // 文件流（如 Excel 模板下载）直接返回原始数据
    if (response.config.responseType === 'blob') {
      return response.data as any
    }
    const res = response.data
    if (res.code === 0) {
      return res as any
    }
    // 401 未登录
    if (res.code === 401) {
      ElMessage.error(res.msg || '登录已过期')
      removeToken()
      // 避免循环跳转
      if (!location.hash.includes('/login') && !location.pathname.includes('/login')) {
        location.href = '/'
      }
      return Promise.reject(new Error(res.msg))
    }
    // 403 无权限
    if (res.code === 403) {
      ElMessage.error(res.msg || '无权限')
      return Promise.reject(new Error(res.msg))
    }
    ElMessage.error(res.msg || '请求失败')
    return Promise.reject(new Error(res.msg))
  },
  (error) => {
    // 用户主动取消（分片上传中断等）：静默，不弹 toast
    if (error?.code === 'ERR_CANCELED' || error?.name === 'CanceledError') {
      return Promise.reject(error)
    }
    ElMessage.error(error.message || '网络异常')
    return Promise.reject(error)
  }
)

export default service
