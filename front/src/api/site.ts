import request from './request'
import type { ApiResult } from './request'

export interface SiteConfig {
  logoUrl: string
  siteName: string
}

export const fetchSiteConfig = async (): Promise<SiteConfig> => {
  const res = await request.get<any, ApiResult<SiteConfig>>('/user/site-config')
  return res.data || { logoUrl: '', siteName: '培训学习平台' }
}
