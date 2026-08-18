import { defineStore } from 'pinia'
import { ref } from 'vue'
import {
  login as loginApi, register as registerApi, getUserInfo,
  logout as logoutApi, type UserInfo, type LoginParams, type RegisterParams
} from '@/api/auth'
import { getToken, setToken, removeToken } from '@/api/request'

export const useUserStore = defineStore('user', () => {
  const token = ref<string>(getToken())
  const userInfo = ref<UserInfo | null>(null)

  async function login(params: LoginParams) {
    const res = await loginApi(params)
    token.value = res.data.token
    setToken(res.data.token)
    return res
  }

  async function register(params: RegisterParams) {
    return await registerApi(params)
  }

  async function fetchUserInfo() {
    const res = await getUserInfo()
    userInfo.value = res.data
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
    removeToken()
  }

  return { token, userInfo, login, register, fetchUserInfo, logout, reset }
})
