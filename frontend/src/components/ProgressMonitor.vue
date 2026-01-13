<template>
  <el-card class="progress-monitor tech-card" v-if="visible">
    <template #header>
      <div class="card-header">
        <el-icon class="rotating" v-if="task.status === 'processing'"><Loading /></el-icon>
        <el-icon v-else-if="task.status === 'completed'"><CircleCheck /></el-icon>
        <el-icon v-else-if="task.status === 'failed'"><CircleClose /></el-icon>
        <span>实时进度监控 - {{ getStatusText(task.status) }}</span>
      </div>
    </template>

    <div class="progress-content">
      <!-- 进度条 -->
      <el-progress
        :percentage="task.progress"
        :status="getProgressStatus()"
        :stroke-width="20"
      >
        <span class="progress-text">{{ task.progress.toFixed(1) }}%</span>
      </el-progress>

      <!-- 核心差异点：详细进度信息 -->
      <el-descriptions :column="3" border style="margin-top: 20px">
        <el-descriptions-item label="当前帧">
          <el-tag>{{ task.current_frame }}</el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="总帧数">
          <el-tag>{{ task.total_frames }}</el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="预计剩余时间">
          <el-tag :type="task.eta > 60 ? 'warning' : 'success'">
            {{ formatETA(task.eta) }}
          </el-tag>
        </el-descriptions-item>
      </el-descriptions>

      <!-- 实时日志流 -->
      <el-divider content-position="left">处理日志</el-divider>
      <div class="log-container">
        <div v-for="(log, index) in logs" :key="index" class="log-item">
          <span class="log-time">{{ formatLogTime(log.timestamp) }}</span>
          <span class="log-message">{{ log.message }}</span>
        </div>
      </div>

      <!-- 操作按钮 -->
      <div class="actions" style="margin-top: 20px">
        <el-button
          v-if="task.status === 'completed'"
          type="success"
          @click="handleDownload"
        >
          <el-icon><Download /></el-icon>
          下载视频
        </el-button>
        <el-button
          v-if="task.status === 'failed'"
          type="danger"
          @click="handleViewError"
        >
          <el-icon><Warning /></el-icon>
          查看错误详情
        </el-button>
      </div>

      <!-- 错误详情展示 -->
      <el-alert
        v-if="task.status === 'failed' && errorDetails"
        type="error"
        title="错误详情"
        :closable="true"
        style="margin-top: 20px"
      >
        <div class="error-details">
          <p><strong>错误信息：</strong></p>
          <pre>{{ errorDetails.error_message || '未知错误' }}</pre>

          <el-collapse v-if="errorDetails.stderr_log">
            <el-collapse-item title="查看完整日志" name="1">
              <pre class="stderr-log">{{ errorDetails.stderr_log }}</pre>
            </el-collapse-item>
          </el-collapse>
        </div>
      </el-alert>
    </div>
  </el-card>
</template>

<script setup>
import { ref, reactive, onMounted, onUnmounted, watch } from 'vue'
import { ElMessage } from 'element-plus'
import { Loading, CircleCheck, CircleClose, Download, Warning } from '@element-plus/icons-vue'
import { taskAPI } from '@/api/task'

const props = defineProps({
  taskId: {
    type: String,
    required: true
  }
})

const emit = defineEmits(['completed', 'failed'])

const visible = ref(true)
const task = reactive({
  status: 'pending',
  progress: 0,
  current_frame: 0,
  total_frames: 0,
  eta: 0
})

const logs = ref([])
const errorDetails = ref(null)
let ws = null
let pollTimer = null // 轮询定时器

// WebSocket连接
const connectWebSocket = () => {
  const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
  const wsUrl = `${protocol}//${window.location.host}/api/tasks/${props.taskId}/progress`
  console.log('[WebSocket] Connecting to:', wsUrl)

  ws = new WebSocket(wsUrl)

  ws.onopen = () => {
    console.log('[WebSocket] Connected successfully')
    addLog('连接成功，开始监听进度...')
  }

  ws.onmessage = (event) => {
    try {
      console.log('[WebSocket] Received message:', event.data)
      const progress = JSON.parse(event.data)
      updateProgress(progress)
    } catch (error) {
      console.error('[WebSocket] Parse progress failed', error)
    }
  }

  ws.onerror = (error) => {
    console.error('[WebSocket] Error:', error)
    addLog('连接错误，尝试重连...')
  }

  ws.onclose = (event) => {
    console.log('[WebSocket] Connection closed, code:', event.code, 'reason:', event.reason)
    addLog('连接已关闭')
  }
}

// 更新进度
const updateProgress = (progress) => {
  console.log('[Progress] Updating:', progress)
  task.status = progress.status
  task.progress = progress.progress || 0
  task.current_frame = progress.current_frame || 0
  task.total_frames = progress.total_frames || 0
  task.eta = progress.eta || 0

  if (progress.message) {
    addLog(progress.message)
  }

  // 任务完成或失败
  if (progress.status === 'completed') {
    console.log('[Progress] Task completed!')
    ElMessage.success('视频生成成功！')
    stopPolling() // 停止轮询
    emit('completed', progress)
  } else if (progress.status === 'failed') {
    console.log('[Progress] Task failed!')
    ElMessage.error('视频生成失败')
    stopPolling() // 停止轮询
    // 自动加载错误详情
    loadErrorDetails()
    emit('failed', progress)
  }
}

// 轮询任务状态（作为WebSocket的备份）
const pollTaskStatus = async () => {
  try {
    const taskDetail = await taskAPI.getTask(props.taskId)
    console.log('[Poll] Task status:', taskDetail.status, 'Progress:', taskDetail.progress)

    // 只在状态变化时更新
    if (taskDetail.status !== task.status || Math.abs(taskDetail.progress - task.progress) > 0.1) {
      updateProgress({
        status: taskDetail.status,
        progress: taskDetail.progress,
        current_frame: taskDetail.current_frame,
        total_frames: taskDetail.total_frames,
        eta: taskDetail.eta,
        message: `Polling update: ${taskDetail.progress.toFixed(1)}%`
      })
    }

    // 如果任务已完成或失败，停止轮询
    if (taskDetail.status === 'completed' || taskDetail.status === 'failed') {
      stopPolling()
    }
  } catch (error) {
    console.error('[Poll] Failed to get task status:', error)
  }
}

// 启动轮询
const startPolling = () => {
  // 每2秒轮询一次
  pollTimer = setInterval(pollTaskStatus, 2000)
  console.log('[Poll] Started polling')
}

// 停止轮询
const stopPolling = () => {
  if (pollTimer) {
    clearInterval(pollTimer)
    pollTimer = null
    console.log('[Poll] Stopped polling')
  }
}

// 加载错误详情
const loadErrorDetails = async () => {
  try {
    const taskDetail = await taskAPI.getTask(props.taskId)
    errorDetails.value = taskDetail
  } catch (error) {
    console.error('Failed to load error details:', error)
  }
}

// 添加日志
const addLog = (message) => {
  logs.value.push({
    timestamp: new Date(),
    message
  })

  // 限制日志数量
  if (logs.value.length > 50) {
    logs.value.shift()
  }
}

// 格式化ETA
const formatETA = (seconds) => {
  if (seconds === 0) return '计算中...'
  if (seconds < 60) return `${seconds}秒`
  const minutes = Math.floor(seconds / 60)
  const secs = seconds % 60
  return `${minutes}分${secs}秒`
}

// 格式化日志时间
const formatLogTime = (time) => {
  return time.toLocaleTimeString('zh-CN')
}

// 获取状态文本
const getStatusText = (status) => {
  const map = {
    pending: '等待中',
    processing: '处理中',
    completed: '已完成',
    failed: '失败'
  }
  return map[status] || status
}

// 获取进度条状态
const getProgressStatus = () => {
  if (task.status === 'completed') return 'success'
  if (task.status === 'failed') return 'exception'
  return undefined
}

// 下载视频
const handleDownload = async () => {
  try {
    const taskDetail = await taskAPI.getTask(props.taskId)
    window.open(taskDetail.output_url, '_blank')
  } catch (error) {
    ElMessage.error('获取下载链接失败')
  }
}

// 查看错误详情
const handleViewError = async () => {
  if (!errorDetails.value) {
    await loadErrorDetails()
  }
  // 错误详情已经显示在页面上
}

onMounted(() => {
  console.log('[Component] Mounted, taskId:', props.taskId)
  connectWebSocket()
  // 启动轮询作为备份
  startPolling()
  // 初始化时检查任务状态，如果已失败则加载错误详情
  checkInitialStatus()
})

// 检查初始状态
const checkInitialStatus = async () => {
  try {
    const taskDetail = await taskAPI.getTask(props.taskId)
    console.log('[Init] Initial task status:', taskDetail.status, 'Progress:', taskDetail.progress)

    // 立即同步初始状态
    updateProgress({
      status: taskDetail.status,
      progress: taskDetail.progress,
      current_frame: taskDetail.current_frame,
      total_frames: taskDetail.total_frames,
      eta: taskDetail.eta,
      message: 'Initial status loaded'
    })

    if (taskDetail.status === 'failed') {
      errorDetails.value = taskDetail
    }
  } catch (error) {
    console.error('Failed to check initial status:', error)
  }
}

onUnmounted(() => {
  console.log('[Component] Unmounting, cleaning up')
  if (ws) {
    ws.close()
  }
  stopPolling()
})
</script>

<style scoped>
.progress-monitor {
  margin-top: 20px;
  animation: slideIn 0.3s ease-out;
}

.tech-card {
  background: rgba(30, 41, 59, 0.5);
  border: 1px solid rgba(99, 102, 241, 0.2);
  backdrop-filter: blur(10px);
}

.tech-card :deep(.el-card__header) {
  background: rgba(15, 23, 42, 0.5);
  border-bottom: 1px solid rgba(99, 102, 241, 0.2);
}

@keyframes slideIn {
  from {
    opacity: 0;
    transform: translateY(-20px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

.card-header {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 16px;
  font-weight: 600;
  color: #e0e7ff;
}

.rotating {
  animation: rotate 1s linear infinite;
}

@keyframes rotate {
  from {
    transform: rotate(0deg);
  }
  to {
    transform: rotate(360deg);
  }
}

.progress-text {
  font-size: 14px;
  font-weight: 600;
}

.log-container {
  max-height: 200px;
  overflow-y: auto;
  background: rgba(15, 23, 42, 0.5);
  border-radius: 4px;
  padding: 10px;
  font-family: 'Courier New', monospace;
  font-size: 12px;
  border: 1px solid rgba(99, 102, 241, 0.2);
}

.log-item {
  margin-bottom: 5px;
  line-height: 1.6;
}

.log-time {
  color: #6b7280;
  margin-right: 10px;
}

.log-message {
  color: #a5b4fc;
}

.actions {
  display: flex;
  gap: 10px;
}

.error-details pre {
  background: rgba(15, 23, 42, 0.8);
  padding: 10px;
  border-radius: 4px;
  overflow-x: auto;
  font-size: 12px;
  line-height: 1.5;
  margin: 10px 0;
  color: #e0e7ff;
  border: 1px solid rgba(99, 102, 241, 0.2);
}

.stderr-log {
  max-height: 400px;
  overflow-y: auto;
  background: #1e293b;
  color: #f8f8f2;
  padding: 15px;
  border-radius: 4px;
  font-family: 'Courier New', monospace;
  font-size: 11px;
  line-height: 1.6;
  border: 1px solid rgba(99, 102, 241, 0.3);
}
</style>
