import axios from 'axios'

export interface SiteConfig {
  logoUrl: string
  siteName: string
}

// 公开接口（登录页/注册页未登录时也可调用）
export const fetchSiteConfig = async (): Promise<SiteConfig> => {
  const res = await axios.get('/api/user/site-config')
  return res.data.data || { logoUrl: '', siteName: '培训学习平台' }
}
