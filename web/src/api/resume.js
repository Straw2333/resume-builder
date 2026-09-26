import axios from 'axios'
import { ElMessage } from 'element-plus'

const http = axios.create({ baseURL: '/api/resumes', timeout: 60000 })
// AI 接口单独实例：润色/测试可能要等模型较久
const ai = axios.create({ baseURL: '/api/ai', timeout: 120000 })

// 统一错误提示
const onError = (err) => {
  const msg = err.response?.data?.error || err.message || '请求失败'
  ElMessage.error(msg)
  return Promise.reject(err)
}
http.interceptors.response.use((res) => res, onError)
ai.interceptors.response.use((res) => res, onError)

export default {
  list: () => http.get('').then((r) => r.data),
  get: (id) => http.get(`/${id}`).then((r) => r.data),
  create: (data) => http.post('', data).then((r) => r.data),
  update: (id, data) => http.put(`/${id}`, data).then((r) => r.data),
  remove: (id) => http.delete(`/${id}`).then((r) => r.data),
  duplicate: (id) => http.post(`/${id}/duplicate`).then((r) => r.data),
  exportUrl: (id, format) => `/api/resumes/${id}/export?format=${format}`,
  upload: (formData) => axios.post('/api/upload', formData).then((r) => r.data),

  // ===== AI 优化 =====
  aiGetSettings: () => ai.get('/settings').then((r) => r.data),
  aiSaveSettings: (data) => ai.put('/settings', data).then((r) => r.data),
  aiTest: (data) => ai.post('/test', data).then((r) => r.data),
  aiPolish: (data) => ai.post('/polish', data).then((r) => r.data)
}
