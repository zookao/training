import type { App, Directive } from 'vue'
import { useUserStore } from '@/stores/user'

/** v-permission 按钮权限指令：v-permission="'admin:add'" 或 v-permission="['a','b']" */
const permission: Directive = {
  mounted(el, binding) {
    const store = useUserStore()
    const value = binding.value
    if (!value) return
    // 超管拥有全部权限
    if (store.perms.includes('*')) return
    const codes = Array.isArray(value) ? value : [value]
    const has = codes.some((c: string) => store.perms.includes(c))
    if (!has) {
      el.parentNode?.removeChild(el)
    }
  }
}

export function setupDirectives(app: App) {
  app.directive('permission', permission)
}
