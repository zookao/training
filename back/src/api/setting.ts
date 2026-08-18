import request from './request'
import type { ApiResult } from './request'

export interface SiteConfig {
  logoUrl: string
  siteName: string
}

export const getSiteConfig = () =>
  request.get<any, ApiResult<SiteConfig>>('/setting/site')

export const updateSiteConfig = (data: SiteConfig) =>
  request.put<any, ApiResult>('/setting/site', data)

export const uploadLogo = (file: File) => {
  const form = new FormData()
  form.append('file', file)
  return request.post<any, ApiResult<{ url: string }>>('/upload/image', form, {
    headers: { 'Content-Type': 'multipart/form-data' },
    timeout: 60000
  })
}
