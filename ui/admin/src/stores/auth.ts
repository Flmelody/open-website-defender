import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import request from '@/utils/request'

interface UserInfo {
  id: number
  username: string
}

interface LoginResponse {
  token: string
  user: UserInfo
}

export const useAuthStore = defineStore('auth', () => {
  const token = ref(localStorage.getItem('token') || '')
  const user = ref<UserInfo | null>(null)

  // Initialize user from localStorage if available
  const storedUser = localStorage.getItem('user')
  if (storedUser) {
    try {
      user.value = JSON.parse(storedUser)
    } catch (e) {
      console.error('Failed to parse stored user data', e)
      localStorage.removeItem('user')
    }
  }
  
  const isLoggedIn = computed(() => !!token.value)

  async function login(username: string, password: string): Promise<void> {
    try {
      const res = await request.post<any, LoginResponse>('/admin-login', { username, password })
      // Since our request interceptor returns data.data directly on success
      // we should expect res to be LoginResponse directly
      
      if (res.token) {
        token.value = res.token
        user.value = res.user
        localStorage.setItem('token', res.token)
        localStorage.setItem('user', JSON.stringify(res.user))
      }
    } catch (error) {
      throw error
    }
  }

  function logout() {
    token.value = ''
    user.value = null
    localStorage.removeItem('token')
    localStorage.removeItem('user')
  }

  return { token, user, isLoggedIn, login, logout }
})

