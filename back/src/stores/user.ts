import { defineStore } from 'pinia'
import { ref } from 'vue'
import { login as loginApi, getUserInfo, logout as logoutApi, type UserInfo, type LoginParams } from '@/api/auth'
import { getToken, setToken, removeToken } from '@/utils/auth'

export const useUserStore = defineStore('user', () => {
  const token = ref<string>(getToken())
  const userInfo = ref<UserInfo | null>(null)
  const roles = ref<string[]>([])
  const perms = ref<string[]>([])

  async function login(params: LoginParams) {
    const res = await loginApi(params)
    token.value = res.data.token
    setToken(res.data.token)
    return res
  }

  async function fetchUserInfo() {
    const res = await getUserInfo()
    userInfo.value = res.data
    roles.value = res.data.roles || []
    perms.value = res.data.perms || []
    return res.data
  }

  async function logout() {
    try {
      await logoutApi()
    } catch (e) {
      /* ignore */
    }
    reset()
  }

  function reset() {
    token.value = ''
    userInfo.value = null
    roles.value = []
    perms.value = []
    removeToken()
  }

  return { token, userInfo, roles, perms, login, fetchUserInfo, logout, reset }
})
