import axios, { AxiosError, type InternalAxiosRequestConfig } from 'axios'
import { getToken } from '../services/session'

interface ApiEnvelope<T> {
  code: number
  msg: string
  data: T
}

const request = axios.create({
  baseURL: import.meta.env.VITE_API_BASE_URL || 'http://127.0.0.1:8080/api',
  timeout: 10000,
})

request.interceptors.request.use((config: InternalAxiosRequestConfig) => {
  const token = getToken()
  if (token) {
    config.headers.Authorization = `Bearer ${token}`
  }
  return config
})

request.interceptors.response.use(
  (response) => {
    const result = response.data as ApiEnvelope<unknown>
    if (result && typeof result.code === 'number') {
      if (result.code === 1) return result.data
      return Promise.reject(new Error(result.msg || '请求失败'))
    }
    return response.data
  },
  (error: AxiosError<ApiEnvelope<unknown>>) => {
    if (error.code === 'ECONNABORTED') {
      return Promise.reject(new Error('请求超时，请检查后端服务或网络连接'))
    }

    if (!error.response) {
      return Promise.reject(new Error('无法连接后端服务，请确认后端已启动并监听 http://127.0.0.1:8080'))
    }

    const message = error.response.data?.msg || `请求失败（HTTP ${error.response.status}）`
    return Promise.reject(new Error(message))
  },
)

export default request
