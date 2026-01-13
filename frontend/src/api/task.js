import axios from 'axios'

const request = axios.create({
  baseURL: '/api',
  timeout: 30000
})

// 请求拦截器
request.interceptors.request.use(
  config => {
    return config
  },
  error => {
    return Promise.reject(error)
  }
)

// 响应拦截器
request.interceptors.response.use(
  response => {
    return response.data
  },
  error => {
    return Promise.reject(error)
  }
)

// 任务相关API
export const taskAPI = {
  // 创建任务
  createTask(data) {
    return request.post('/tasks', data)
  },

  // 获取任务详情
  getTask(taskId) {
    return request.get(`/tasks/${taskId}`)
  },

  // 获取任务列表
  listTasks(params) {
    return request.get('/tasks', { params })
  },

  // WebSocket连接（实时进度）
  connectProgress(taskId) {
    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
    const wsUrl = `${protocol}//${window.location.host}/api/tasks/${taskId}/progress`
    return new WebSocket(wsUrl)
  }
}

// 上传文件
export const uploadFile = (file) => {
  const formData = new FormData()
  formData.append('file', file)

  return request.post('/upload', formData, {
    headers: {
      'Content-Type': 'multipart/form-data'
    }
  })
}

export default request
