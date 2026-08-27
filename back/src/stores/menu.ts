import { defineStore } from 'pinia'
import { ref } from 'vue'
import type { RouteRecordRaw } from 'vue-router'
import { getMenus, type MenuItem } from '@/api/auth'
import router from '@/router'
import Layout from '@/layout/index.vue'

// 动态加载 views 下所有 vue 组件
const modules = import.meta.glob('@/views/**/*.vue')

// 兜底 404 路由命名（命名后重复添加会自动替换，避免多次登录累积重复路由）
const CATCH_ALL_NAME = 'CatchAllRedirect'
// 已注册的动态路由名（仅记录顶级路由；removeRoute 顶级路由时会级联删除其子路由）
let addedRouteNames: (string | symbol)[] = []

function removeDynamicRoutes() {
  addedRouteNames.forEach((name) => {
    if (router.hasRoute(name)) router.removeRoute(name)
  })
  addedRouteNames = []
}

export const useMenuStore = defineStore('menu', () => {
  const menus = ref<MenuItem[]>([])
  const routes = ref<RouteRecordRaw[]>([])
  const loaded = ref(false)

  async function generateRoutes(): Promise<RouteRecordRaw[]> {
    // 先移除上一次会话注册的动态路由，避免旧路由残留（同名父路由阴影）导致页面空白
    removeDynamicRoutes()
    const res = await getMenus()
    menus.value = res.data || []
    const dynamicRoutes = transformMenus(menus.value)
    routes.value = dynamicRoutes
    dynamicRoutes.forEach((r) => {
      router.addRoute(r)
      if (r.name) addedRouteNames.push(r.name)
    })
    // 兜底 404（必须最后添加）
    router.addRoute({ name: CATCH_ALL_NAME, path: '/:pathMatch(.*)*', redirect: '/404' })
    addedRouteNames.push(CATCH_ALL_NAME)
    loaded.value = true
    return dynamicRoutes
  }

  function reset() {
    removeDynamicRoutes()
    menus.value = []
    routes.value = []
    loaded.value = false
  }

  return { menus, routes, loaded, generateRoutes, reset }
})

/** 将后端菜单树转换为路由（每个顶级菜单包裹 Layout） */
function transformMenus(menus: MenuItem[]): RouteRecordRaw[] {
  const result: RouteRecordRaw[] = []
  for (const m of menus) {
    if (m.type === 'F') continue
    const visibleChildren = (m.children || []).filter((c) => c.type !== 'F')
    if (m.type === 'M' && visibleChildren.length) {
      const children: RouteRecordRaw[] = visibleChildren.map((c) => ({
        path: c.path,
        name: `M${c.id}`,
        component: loadComponent(c.component),
        meta: { title: c.name, icon: c.icon, keepAlive: c.keepAlive }
      }))
      result.push({
        path: m.path,
        name: `L${m.id}`,
        component: Layout,
        redirect: `${m.path}/${children[0].path}`,
        meta: { title: m.name, icon: m.icon },
        children
      })
    } else {
      // 顶级菜单（无目录）：单菜单挂到 Layout
      result.push({
        path: m.path,
        name: `L${m.id}`,
        component: Layout,
        children: [
          {
            path: '',
            name: `M${m.id}`,
            component: loadComponent(m.component),
            meta: { title: m.name, icon: m.icon, keepAlive: m.keepAlive }
          }
        ]
      })
    }
  }
  return result
}

function loadComponent(component: string) {
  if (!component) return () => import('@/views/error/404.vue')
  const key = `/src/views/${component}.vue`
  return modules[key] || (() => import('@/views/error/404.vue'))
}
