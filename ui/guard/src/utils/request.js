import axios from 'axios'
import { getAppConfig, getPagePath } from './config'

const request = axios.create({
  baseURL: getAppConfig().baseURL,
  timeout: 5000
})

request.interceptors.request.use(config => {
  const token = localStorage.getItem('token')
  if (token) {
    config.headers['Defender-Authorization'] = `Bearer ${token}`
  }
  return config
})

request.interceptors.response.use(
  response => {
    const { data } = response
    if (data && typeof data === 'object' && 'code' in data) {
      if (data.code === 0) {
        return data.data
      } else {
        return Promise.reject(new Error(data.error || data.message || 'Request failed'))
      }
    }
    return data
  },
  error => {
    if (error.response?.status === 401 && !error.config?.url?.includes('/login')) {
      localStorage.removeItem('token')
      window.location.href = `${getPagePath('guard')}/login`
    }
    
    const errorMessage = error.response?.data?.error 
      || error.response?.data?.message 
      || error.message 
      || 'Network error'
    
    return Promise.reject(new Error(errorMessage))
  }
)

export default request
