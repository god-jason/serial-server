import axios from 'axios'

const api = axios.create({
  baseURL: '/api',
  timeout: 10000,
  withCredentials: true
})

api.interceptors.response.use(response => {
  return response
}, error => {
  if (error.response) {
    if (error.response.status === 401) {
      window.location.href = '/login'
    }
  } else if (error.request) {
    console.error('请求发送失败:', error.message)
  } else {
    console.error('请求配置错误:', error.message)
  }
  return Promise.reject(error)
})

export default api