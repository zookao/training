import { defineStore } from 'pinia'
import { ref } from 'vue'
import { fetchSiteConfig } from '@/api/site'

export const useSiteStore = defineStore('site', () => {
  const logoSrc = ref('')
  const siteName = ref('培训学习平台')
  let loaded = false

  async function load() {
    try {
      const cfg = await fetchSiteConfig()
      siteName.value = cfg.siteName || '培训学习平台'
      logoSrc.value = cfg.logoUrl ? `/api/user/site-logo?t=${Date.now()}` : ''
      loaded = true
    } catch (e) {
      /* 用默认值 */
    }
  }

  function ensureLoaded() {
    if (!loaded) load()
  }

  return { logoSrc, siteName, load, ensureLoaded }
})
