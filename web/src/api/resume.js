import axios from 'axios'
import { ElMessage } from 'element-plus'

const http = axios.create({ baseURL: '/api/resumes', timeout: 60000 })

// 统一错误提示
http.interceptors.response.use(
  (res) => res,
  (err) => {
    const msg = err.response?.data?.error || err.message || '请求失败'
    ElMessage.error(msg)
    return Promise.reject(err)
  }
)

export default {
  list: () => http.get('').then((r) => r.data),
  get: (id) => http.get(`/${id}`).then((r) => r.data),
  create: (data) => http.post('', data).then((r) => r.data),
  update: (id, data) => http.put(`/${id}`, data).then((r) => r.data),
  remove: (id) => http.delete(`/${id}`).then((r) => r.data),
  duplicate: (id) => http.post(`/${id}/duplicate`).then((r) => r.data),
  exportUrl: (id, format) => `/api/resumes/${id}/export?format=${format}`
}
