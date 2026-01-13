<template>
  <div class="task-create">
    <el-card class="tech-card">
      <template #header>
        <div class="card-header">
          <el-icon><VideoPlay /></el-icon>
          <span>图片+音频生成视频</span>
        </div>
      </template>

      <el-form :model="form" label-width="140px" :rules="rules" ref="formRef">
        <!-- 文件上传 -->
        <el-form-item label="图片文件" prop="imagePath">
          <el-upload
            class="upload-demo"
            :auto-upload="true"
            :show-file-list="true"
            :limit="1"
            accept="image/*"
            :http-request="handleImageUpload"
          >
            <el-button type="primary" :icon="Upload">选择图片</el-button>
            <template #tip>
              <div class="el-upload__tip">支持 JPG、PNG 等格式</div>
            </template>
          </el-upload>
        </el-form-item>

        <el-form-item label="音频文件" prop="audioPath">
          <el-upload
            class="upload-demo"
            :auto-upload="true"
            :show-file-list="true"
            :limit="1"
            accept="audio/*"
            :http-request="handleAudioUpload"
          >
            <el-button type="primary" :icon="Upload">选择音频</el-button>
            <template #tip>
              <div class="el-upload__tip">支持 MP3、WAV、AAC 等格式</div>
            </template>
          </el-upload>
        </el-form-item>

        <!-- 语义化参数 -->
        <el-divider content-position="left">视频参数</el-divider>

        <el-form-item label="视频尺寸">
          <el-row :gutter="10">
            <el-col :span="11">
              <el-input v-model.number="form.width" placeholder="宽度">
                <template #append>px</template>
              </el-input>
            </el-col>
            <el-col :span="2" style="text-align: center">×</el-col>
            <el-col :span="11">
              <el-input v-model.number="form.height" placeholder="高度">
                <template #append>px</template>
              </el-input>
            </el-col>
          </el-row>
          <div class="form-hint">默认使用图片原始尺寸</div>
        </el-form-item>

        <el-form-item label="帧率 (FPS)">
          <el-input-number v-model="form.fps" :min="1" :max="60" :step="1" />
          <div class="form-hint">推荐：25-30fps</div>
        </el-form-item>

        <el-form-item label="视频编码">
          <el-select v-model="form.videoCodec" placeholder="选择编码器">
            <el-option label="H.264 (兼容性好)" value="libx264" />
            <el-option label="H.265 (体积更小)" value="libx265" />
          </el-select>
        </el-form-item>

        <el-form-item label="视频码率">
          <el-select v-model="form.videoBitrate" placeholder="选择码率">
            <el-option label="低质量 (500k)" value="500k" />
            <el-option label="标准质量 (1M)" value="1M" />
            <el-option label="高质量 (2M)" value="2M" />
            <el-option label="超高质量 (5M)" value="5M" />
          </el-select>
        </el-form-item>

        <el-divider content-position="left">音频参数</el-divider>

        <el-form-item label="音频编码">
          <el-select v-model="form.audioCodec" placeholder="选择编码器">
            <el-option label="AAC (推荐)" value="aac" />
            <el-option label="MP3" value="libmp3lame" />
          </el-select>
        </el-form-item>

        <el-form-item label="音频码率">
          <el-select v-model="form.audioBitrate" placeholder="选择码率">
            <el-option label="低质量 (96k)" value="96k" />
            <el-option label="标准质量 (128k)" value="128k" />
            <el-option label="高质量 (192k)" value="192k" />
            <el-option label="超高质量 (320k)" value="320k" />
          </el-select>
        </el-form-item>

        <el-form-item label="音频循环">
          <el-switch
            v-model="form.audioLoop"
            active-text="开启（背景音乐循环播放）"
            inactive-text="关闭（音频结束即停止）"
          />
        </el-form-item>

        <el-divider content-position="left">输出设置</el-divider>

        <el-form-item label="输出格式">
          <el-radio-group v-model="form.outputFormat">
            <el-radio label="mp4">MP4 (推荐)</el-radio>
            <el-radio label="mov">MOV</el-radio>
            <el-radio label="avi">AVI</el-radio>
          </el-radio-group>
        </el-form-item>

        <el-form-item>
          <el-button type="primary" size="large" :loading="creating" @click="handleSubmit">
            <el-icon v-if="!creating"><VideoCamera /></el-icon>
            开始生成视频
          </el-button>
        </el-form-item>
      </el-form>
    </el-card>

    <!-- 进度监控组件 -->
    <ProgressMonitor v-if="currentTaskId" :task-id="currentTaskId" @completed="handleTaskCompleted" />

    <!-- 任务历史 -->
    <el-card class="tech-card" style="margin-top: 20px" v-if="tasks.length > 0">
      <template #header>
        <div class="card-header">
          <el-icon><List /></el-icon>
          <span>任务历史</span>
        </div>
      </template>

      <el-table :data="tasks" style="width: 100%">
        <el-table-column prop="id" label="任务ID" width="280" />
        <el-table-column prop="status" label="状态" width="120">
          <template #default="{ row }">
            <el-tag :type="getStatusType(row.status)">{{ getStatusText(row.status) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="progress" label="进度" width="120">
          <template #default="{ row }">
            {{ row.progress.toFixed(1) }}%
          </template>
        </el-table-column>
        <el-table-column prop="created_at" label="创建时间" width="180">
          <template #default="{ row }">
            {{ formatTime(row.created_at) }}
          </template>
        </el-table-column>
        <el-table-column label="操作">
          <template #default="{ row }">
            <el-button link type="primary" @click="viewTask(row)">查看详情</el-button>
            <el-button link type="success" v-if="row.status === 'completed'" @click="downloadVideo(row)">
              下载视频
            </el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <!-- 任务详情对话框 -->
    <el-dialog v-model="detailVisible" title="任务详情" width="80%">
      <el-descriptions :column="2" border v-if="selectedTask">
        <el-descriptions-item label="任务ID">{{ selectedTask.id }}</el-descriptions-item>
        <el-descriptions-item label="状态">
          <el-tag :type="getStatusType(selectedTask.status)">{{ getStatusText(selectedTask.status) }}</el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="进度">{{ selectedTask.progress.toFixed(1) }}%</el-descriptions-item>
        <el-descriptions-item label="当前帧/总帧数">
          {{ selectedTask.current_frame }} / {{ selectedTask.total_frames }}
        </el-descriptions-item>
        <el-descriptions-item label="预计剩余时间">{{ selectedTask.eta }} 秒</el-descriptions-item>
        <el-descriptions-item label="创建时间">{{ formatTime(selectedTask.created_at) }}</el-descriptions-item>
      </el-descriptions>

      <!-- 核心差异点展示 -->
      <el-divider content-position="left">FFmpeg 命令（可回放）</el-divider>
      <el-input
        v-model="selectedTask.ffmpeg_command"
        type="textarea"
        :rows="3"
        readonly
        class="code-block"
      />
      <el-button size="small" @click="copyCommand" style="margin-top: 10px">
        <el-icon><CopyDocument /></el-icon> 复制命令
      </el-button>

      <el-divider content-position="left" v-if="selectedTask.filter_graph">Filter Graph</el-divider>
      <el-input
        v-if="selectedTask.filter_graph"
        v-model="selectedTask.filter_graph"
        type="textarea"
        :rows="2"
        readonly
        class="code-block"
      />

      <el-divider content-position="left" v-if="selectedTask.status === 'failed'">错误信息（失败可解释）</el-divider>
      <el-alert
        v-if="selectedTask.status === 'failed'"
        :title="selectedTask.error_message"
        type="error"
        :closable="false"
        show-icon
      />

      <el-divider content-position="left" v-if="selectedTask.stderr_log">完整日志</el-divider>
      <el-input
        v-if="selectedTask.stderr_log"
        v-model="selectedTask.stderr_log"
        type="textarea"
        :rows="10"
        readonly
        class="code-block"
      />
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { Upload, VideoPlay, VideoCamera, List, CopyDocument } from '@element-plus/icons-vue'
import { taskAPI, uploadFile } from '@/api/task'
import ProgressMonitor from '@/components/ProgressMonitor.vue'

const formRef = ref(null)
const form = reactive({
  imagePath: '',
  audioPath: '',
  width: 0,
  height: 0,
  fps: 25,
  videoCodec: 'libx264',
  audioBitrate: '128k',
  videoBitrate: '1M',
  audioCodec: 'aac',
  audioLoop: false,
  outputFormat: 'mp4'
})

const rules = {
  imagePath: [{ required: true, message: '请上传图片文件', trigger: 'change' }],
  audioPath: [{ required: true, message: '请上传音频文件', trigger: 'change' }]
}

const creating = ref(false)
const currentTaskId = ref('')
const tasks = ref([])
const detailVisible = ref(false)
const selectedTask = ref(null)

// 上传图片
const handleImageUpload = async ({ file }) => {
  try {
    const result = await uploadFile(file)
    form.imagePath = result.url
    ElMessage.success('图片上传成功')
  } catch (error) {
    ElMessage.error('图片上传失败')
  }
}

// 上传音频
const handleAudioUpload = async ({ file }) => {
  try {
    const result = await uploadFile(file)
    form.audioPath = result.url
    ElMessage.success('音频上传成功')
  } catch (error) {
    ElMessage.error('音频上传失败')
  }
}

// 提交任务
const handleSubmit = async () => {
  try {
    await formRef.value.validate()

    creating.value = true
    const task = await taskAPI.createTask({
      type: 'image_audio_to_video',
      input_params: {
        image_path: form.imagePath,
        audio_path: form.audioPath,
        width: form.width,
        height: form.height,
        fps: form.fps,
        video_codec: form.videoCodec,
        audio_codec: form.audioCodec,
        video_bitrate: form.videoBitrate,
        audio_bitrate: form.audioBitrate,
        audio_loop: form.audioLoop,
        output_format: form.outputFormat
      }
    })

    ElMessage.success('任务创建成功，开始处理...')
    currentTaskId.value = task.id
    loadTasks()
  } catch (error) {
    ElMessage.error('任务创建失败: ' + (error.message || '未知错误'))
  } finally {
    creating.value = false
  }
}

// 任务完成回调
const handleTaskCompleted = () => {
  console.log('[TaskCreate] Task completed, reloading task list')
  loadTasks() // 刷新任务列表
  // 短暂延迟后清除当前任务ID，隐藏进度监控
  setTimeout(() => {
    currentTaskId.value = ''
  }, 2000)
}

// 加载任务列表
const loadTasks = async () => {
  try {
    const result = await taskAPI.listTasks({ page: 1, page_size: 10 })
    tasks.value = result.tasks || []
  } catch (error) {
    console.error('加载任务列表失败', error)
  }
}

// 查看任务详情
const viewTask = async (task) => {
  try {
    selectedTask.value = await taskAPI.getTask(task.id)
    detailVisible.value = true
  } catch (error) {
    ElMessage.error('获取任务详情失败')
  }
}

// 下载视频
const downloadVideo = (task) => {
  window.open(task.output_url, '_blank')
}

// 复制命令
const copyCommand = () => {
  navigator.clipboard.writeText(selectedTask.value.ffmpeg_command)
  ElMessage.success('命令已复制到剪贴板')
}

// 辅助函数
const getStatusType = (status) => {
  const map = {
    pending: 'info',
    processing: 'warning',
    completed: 'success',
    failed: 'danger'
  }
  return map[status] || 'info'
}

const getStatusText = (status) => {
  const map = {
    pending: '等待中',
    processing: '处理中',
    completed: '已完成',
    failed: '失败'
  }
  return map[status] || status
}

const formatTime = (time) => {
  return new Date(time).toLocaleString('zh-CN')
}

onMounted(() => {
  loadTasks()
})
</script>

<style scoped>
.task-create {
  max-width: 100%;
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

.card-header {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 16px;
  font-weight: 600;
  color: #e0e7ff;
}

.form-hint {
  font-size: 12px;
  color: #a5b4fc;
  margin-top: 5px;
}

.code-block {
  font-family: 'Courier New', monospace;
  font-size: 13px;
}

.code-block :deep(textarea) {
  font-family: 'Courier New', monospace;
  background: rgba(15, 23, 42, 0.8);
  color: #e0e7ff;
  border: 1px solid rgba(99, 102, 241, 0.2);
}
</style>
