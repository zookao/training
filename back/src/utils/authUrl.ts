import { getToken } from '@/utils/auth'

/**
 * 为需要鉴权的静态资源 URL（/upload/*）附加 token 查询参数。
 * <video>/<img> 标签无法设置 Authorization Header，需通过 ?token=xxx 传递。
 */
export function authUrl(url: string | undefined | null): string {
  if (!url) return ''
  if (!url.startsWith('/upload')) return url
  const token = getToken()
  if (!token) return url
  const sep = url.includes('?') ? '&' : '?'
  return `${url}${sep}token=${encodeURIComponent(token)}`
}
