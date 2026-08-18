import router from './router'
import { useUserStore } from './stores/user'
import { useMenuStore } from './stores/menu'
import { getToken } from './utils/auth'
import { ElMessage } from 'element-plus'

const whiteList = ['/login', '/404']

router.beforeEach(async (to, _from, next) => {
  const hasToken = getToken()
  if (hasToken) {
    if (to.path === '/login') {
      next({ path: '/' })
      return
    }
    const userStore = useUserStore()
    const menuStore = useMenuStore()
    if (userStore.roles.length === 0) {
      try {
        await userStore.fetchUserInfo()
        const accessRoutes = await menuStore.generateRoutes()
        // 动态注入路由
        accessRoutes.forEach((r) => router.addRoute(r))
        // 兜底 404（必须最后添加）
        router.addRoute({ path: '/:pathMatch(.*)*', redirect: '/404' })
        next({ ...to, replace: true })
      } catch (e: any) {
        await userStore.logout()
        menuStore.reset()
        ElMessage.error(e.message || '获取用户信息失败，请重新登录')
        next(`/login?redirect=${to.path}`)
      }
    } else {
      next()
    }
  } else {
    if (whiteList.includes(to.path)) {
      next()
    } else {
      next(`/login?redirect=${to.path}`)
    }
  }
})
