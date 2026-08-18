import { defineStore } from 'pinia'
import { ref } from 'vue'
import type { RouteRecordRaw } from 'vue-router'
import { getMenus, type MenuItem } from '@/api/auth'
import Layout from '@/layout/index.vue'

// 动态加载 views 下所有 vue 组件
const modules = import.meta.glob('@/views/**/*.vue')

export const useMenuStore = defineStore('menu', () => {
  const menus = ref<MenuItem[]>([])
  const routes = ref<RouteRecordRaw[]>([])
  const loaded = ref(false)

  async function generateRoutes(): Promise<RouteRecordRaw[]> {
    const res = await getMenus()
    menus.value = res.data || []
    const dynamicRoutes = transformMenus(menus.value)
    routes.value = dynamicRoutes
    loaded.value = true
    return dynamicRoutes
  }

  function reset() {
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
        component: Layout,
        redirect: `${m.path}/${children[0].path}`,
        meta: { title: m.name, icon: m.icon },
        children
      })
    } else {
      // 顶级菜单（无目录）：单菜单挂到 Layout
      result.push({
        path: m.path,
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
