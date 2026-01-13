<template>
  <div class="slideshow-create">
    <el-card class="tech-card">
      <template #header>
        <div class="card-header">
          <el-icon><Film /></el-icon>
          <span>图片轮播视频制作</span>
        </div>
      </template>

      <el-form :model="form" label-width="140px" :rules="rules" ref="formRef">
        <!-- 图片上传 -->
        <el-form-item label="图片列表" prop="imagePaths" class="form-item-images">
          <el-upload
            class="upload-demo"
            :auto-upload="true"
            :show-file-list="true"
            :limit="20"
            accept="image/*"
            :http-request="handleImageUpload"
            :on-remove="handleImageRemove"
            multiple
            drag
          >
            <div class="upload-area">
              <el-icon class="upload-icon"><Plus /></el-icon>
              <div class="upload-text">点击或拖拽上传图片</div>
              <div class="upload-hint">支持JPG、PNG格式，最多20张</div>
            </div>
          </el-upload>
          <div class="form-hint">已上传 {{ form.imagePaths.length }} 张图片</div>
        </el-form-item>

        <!-- 背景音乐 -->
        <el-form-item label="背景音乐（可选）">
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

        <!-- 轮播参数 -->
        <el-divider content-position="left">
          <span class="divider-text">轮播设置</span>
        </el-divider>

        <el-form-item label="每张图片时长">
          <el-input-number v-model="form.imageDuration" :min="1" :max="10" :step="0.5" />
          <span class="input-suffix">秒</span>
          <div class="form-hint">推荐：3-5秒</div>
        </el-form-item>

        <el-form-item label="转场效果">
          <el-radio-group v-model="form.transitionType">
            <el-radio label="fade">淡入淡出</el-radio>
            <el-radio label="none">无转场</el-radio>
          </el-radio-group>
        </el-form-item>

        <el-form-item label="转场时长" v-if="form.transitionType !== 'none'">
          <el-input-number v-model="form.transitionDur" :min="0.1" :max="2" :step="0.1" />
          <span class="input-suffix">秒</span>
          <div class="form-hint">推荐：0.5-1秒</div>
        </el-form-item>

        <!-- 视频参数 -->
        <el-divider content-position="left">
          <span class="divider-text">视频参数</span>
        </el-divider>

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
          <div class="form-hint">默认：1280x720（推荐）</div>
        </el-form-item>

        <el-form-item label="帧率 (FPS)">
          <el-input-number v-model="form.fps" :min="1" :max="60" :step="1" />
          <div class="form-hint">推荐：25-30fps</div>
        </el-form-item>

        <el-form-item label="视频码率">
          <el-select v-model="form.videoBitrate" placeholder="选择码率">
            <el-option label="低质量 (500k)" value="500k" />
            <el-option label="标准质量 (1M)" value="1M" />
            <el-option label="高质量 (2M)" value="2M" />
            <el-option label="超高质量 (5M)" value="5M" />
          </el-select>
        </el-form-item>

        <el-form-item label="输出格式">
          <el-radio-group v-model="form.outputFormat">
            <el-radio label="mp4">MP4 (推荐)</el-radio>
            <el-radio label="mov">MOV</el-radio>
          </el-radio-group>
        </el-form-item>

        <el-form-item>
          <el-button
            type="primary"
            size="large"
            :loading="creating"
            @click="handleSubmit"
            class="submit-btn"
          >
            <el-icon v-if="!creating"><VideoCamera /></el-icon>
            开始生成轮播视频
          </el-button>
        </el-form-item>
      </el-form>
    </el-card>

    <!-- 进度监控 -->
    <ProgressMonitor
      v-if="currentTaskId"
      :task-id="currentTaskId"
      @completed="handleTaskCompleted"
    />

    <!-- 任务历史 -->
    <el-card class="tech-card" v-if="tasks.length > 0" style="margin-top: 20px">
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
          <template #default="{ row }">{{ row.progress.toFixed(1) }}%</template>
        </el-table-column>
        <el-table-column prop="created_at" label="创建时间" width="180">
          <template #default="{ row }">{{ formatTime(row.created_at) }}</template>
        </el-table-column>
        <el-table-column label="操作">
          <template #default="{ row }">
            <el-button link type="primary" @click="viewTask(row)">查看详情</el-button>
            <el-button
              link
              type="success"
              v-if="row.status === 'completed'"
              @click="downloadVideo(row)"
            >
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
          <el-tag :type="getStatusType(selectedTask.status)">
            {{ getStatusText(selectedTask.status) }}
          </el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="进度">
          {{ selectedTask.progress.toFixed(1) }}%
        </el-descriptions-item>
        <el-descriptions-item label="图片数量">
          {{ selectedTask.input_params.image_paths?.length || 0 }} 张
        </el-descriptions-item>
      </el-descriptions>

      <el-divider content-position="left">FFmpeg 命令</el-divider>
      <el-input
        v-model="selectedTask.ffmpeg_command"
        type="textarea"
        :rows="3"
        readonly
        class="code-block"
      />

      <el-divider content-position="left" v-if="selectedTask.filter_graph">Filter Graph</el-divider>
      <el-input
        v-if="selectedTask.filter_graph"
        v-model="selectedTask.filter_graph"
        type="textarea"
        :rows="4"
        readonly
        class="code-block"
      />
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { Upload, Plus, Film, VideoCamera, List } from '@element-plus/icons-vue'
import { taskAPI, uploadFile } from '@/api/task'
import ProgressMonitor from '@/components/ProgressMonitor.vue'

const formRef = ref(null)
const form = reactive({
  imagePaths: [],
  backgroundAudio: '',
  imageDuration: 3,
  transitionType: 'fade',
  transitionDur: 0.5,
  width: 1280,
  height: 720,
  fps: 25,
  videoBitrate: '2M',
  audioCodec: 'aac',
  audioBitrate: '128k',
  outputFormat: 'mp4'
})

const rules = {
  imagePaths: [{ type: 'array', required: true, message: '请至少上传一张图片', trigger: 'change' }]
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
    form.imagePaths.push(result.url)
    ElMessage.success('图片上传成功')
  } catch (error) {
    ElMessage.error('图片上传失败')
  }
}

// 移除图片
const handleImageRemove = (file, fileList) => {
  form.imagePaths = form.imagePaths.slice(0, fileList.length)
}

// 上传音频
const handleAudioUpload = async ({ file }) => {
  try {
    const result = await uploadFile(file)
    form.backgroundAudio = result.url
    ElMessage.success('音频上传成功')
  } catch (error) {
    ElMessage.error('音频上传失败')
  }
}

// 提交任务
const handleSubmit = async () => {
  try {
    await formRef.value.validate()

    if (form.imagePaths.length < 1) {
      ElMessage.warning('请至少上传一张图片')
      return
    }

    creating.value = true
    const task = await taskAPI.createTask({
      type: 'image_slideshow',
      input_params: {
        image_paths: form.imagePaths,
        background_audio: form.backgroundAudio,
        image_duration: form.imageDuration,
        transition_type: form.transitionType,
        transition_dur: form.transitionDur,
        width: form.width,
        height: form.height,
        fps: form.fps,
        video_codec: 'libx264',
        audio_codec: form.audioCodec,
        video_bitrate: form.videoBitrate,
        audio_bitrate: form.audioBitrate,
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
  loadTasks()
  setTimeout(() => {
    currentTaskId.value = ''
  }, 2000)
}

// 加载任务列表
const loadTasks = async () => {
  try {
    const result = await taskAPI.listTasks({ page: 1, page_size: 10 })
    tasks.value = (result.tasks || []).filter((t) => t.type === 'image_slideshow')
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

// 辅助函数
const getStatusType = (status) => {
  const map = { pending: 'info', processing: 'warning', completed: 'success', failed: 'danger' }
  return map[status] || 'info'
}

const getStatusText = (status) => {
  const map = { pending: '等待中', processing: '处理中', completed: '已完成', failed: '失败' }
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
.slideshow-create {
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

.divider-text {
  color: #a5b4fc;
  font-weight: 500;
}

.form-item-images :deep(.el-upload-dragger) {
  background: rgba(15, 23, 42, 0.5);
  border: 2px dashed rgba(99, 102, 241, 0.3);
}

.form-item-images :deep(.el-upload-dragger:hover) {
  border-color: #6366f1;
}

.upload-area {
  padding: 40px;
  text-align: center;
}

.upload-icon {
  font-size: 48px;
  color: #6366f1;
  margin-bottom: 16px;
}

.upload-text {
  font-size: 16px;
  color: #e0e7ff;
  margin-bottom: 8px;
}

.upload-hint {
  font-size: 13px;
  color: #a5b4fc;
}

.form-hint {
  font-size: 12px;
  color: #a5b4fc;
  margin-top: 5px;
}

.input-suffix {
  margin-left: 12px;
  color: #a5b4fc;
}

.submit-btn {
  width: 100%;
  height: 48px;
  font-size: 16px;
  background: linear-gradient(135deg, #6366f1, #8b5cf6);
  border: none;
}

.submit-btn:hover {
  transform: translateY(-2px);
  box-shadow: 0 4px 20px rgba(99, 102, 241, 0.5);
}

.code-block :deep(textarea) {
  font-family: 'Courier New', monospace;
  background: rgba(15, 23, 42, 0.8);
  color: #e0e7ff;
  border: 1px solid rgba(99, 102, 241, 0.2);
}
</style>
