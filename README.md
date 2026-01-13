# FFmpeg Platform - 企业级视频处理平台

一个现代化的 FFmpeg 包装平台，专注于提供清晰的参数语义、可解释的失败信息、实时进度监控和命令回放能力。

## 核心差异点 🎯

与其他 FFmpeg 包装项目相比，本项目的"杀手级特性"：

### 1. 参数语义清晰
- ✅ 提供直观的参数命名（如 `audio_loop: true` 而非复杂的 ffmpeg 参数）
- ✅ 参数分组清晰（视频参数、音频参数、输出设置）
- ✅ 提供默认值和推荐值
- ✅ 前端表单化配置，降低使用门槛

### 2. 失败可解释
- ✅ 保存完整的 stderr 日志
- ✅ 智能提取错误信息
- ✅ 错误原因可追溯
- ✅ 前端友好的错误展示

### 3. 进度可感知
- ✅ WebSocket 实时推送处理进度
- ✅ 显示当前帧/总帧数
- ✅ 实时计算 ETA（预计剩余时间）
- ✅ 显示处理速度（speed multiplier）
- ✅ 实时日志流

### 4. 命令可回放
- ✅ 保存完整的 ffmpeg 命令
- ✅ 保存 filter_complex 图
- ✅ 一键复制命令到剪贴板
- ✅ 可直接在命令行执行验证

## 技术栈

### 后端
- **语言**: Go 1.21
- **框架**: Gin (HTTP Server)
- **任务队列**: Asynq (Redis-based)
- **数据库**: PostgreSQL
- **缓存**: Redis
- **WebSocket**: Gorilla WebSocket
- **FFmpeg**: 内置于 Docker 镜像

### 前端
- **框架**: Vue 3 + Vite
- **UI组件**: Element Plus
- **HTTP客户端**: Axios
- **实时通信**: WebSocket

### 基础设施
- **容器化**: Docker + Docker Compose
- **反向代理**: Nginx
- **数据持久化**: Docker Volumes

## 项目结构

```
fp/
├── backend/                 # Go 后端
│   ├── api/
│   │   ├── handler/        # HTTP handlers
│   │   └── middleware/     # 中间件（CORS等）
│   ├── config/             # 配置管理
│   ├── model/              # 数据模型
│   ├── service/            # 业务逻辑
│   │   ├── task.go        # 任务管理
│   │   └── ffmpeg.go      # FFmpeg服务
│   ├── worker/             # 异步任务处理
│   ├── pkg/
│   │   ├── ffmpeg/        # FFmpeg执行器和解析器
│   │   └── storage/       # 文件存储
│   ├── main.go
│   ├── go.mod
│   └── Dockerfile
├── frontend/               # Vue 3 前端
│   ├── src/
│   │   ├── api/           # API客户端
│   │   ├── views/         # 页面组件
│   │   ├── components/    # 通用组件
│   │   ├── App.vue
│   │   └── main.js
│   ├── package.json
│   ├── vite.config.js
│   ├── Dockerfile
│   └── nginx.conf
├── docker-compose.yml      # Docker编排
└── README.md
```

## 快速开始

### 前置要求
- Docker 20.10+
- Docker Compose 2.0+

### 一键启动

```bash
# 克隆项目
cd fp

# 启动所有服务（会自动构建镜像并安装FFmpeg）
docker-compose up -d

# 查看日志
docker-compose logs -f
```

服务启动后：
- 前端: http://localhost
- 后端API: http://localhost:8008
- 数据库: localhost:5432
- Redis: localhost:6379

### 停止服务

```bash
docker-compose down

# 删除所有数据（包括数据库）
docker-compose down -v
```

## 本地开发

### 后端开发

```bash
cd backend

# 安装依赖
go mod download

# 启动数据库和Redis
docker-compose up -d postgres redis

# 设置环境变量
export DATABASE_DSN="host=localhost user=ffmpeg password=ffmpeg dbname=ffmpeg port=5432 sslmode=disable"
export REDIS_ADDR="localhost:6379"

# 运行
go run main.go
```

### 前端开发

```bash
cd frontend

# 安装依赖
npm install

# 启动开发服务器
npm run dev
```

前端开发服务器会自动代理 API 请求到 `http://localhost:8008`

## 功能演示

### 图片+音频生成视频

1. 上传一张图片（JPG/PNG）
2. 上传一段音频（MP3/WAV）
3. 配置参数：
   - 视频尺寸（默认使用图片原始尺寸）
   - 帧率（推荐 25-30 fps）
   - 视频编码（H.264 或 H.265）
   - 音频编码（AAC 或 MP3）
   - 是否循环播放音频
4. 点击"开始生成视频"
5. 实时查看进度：
   - 当前帧/总帧数
   - 百分比进度
   - 预计剩余时间（ETA）
   - 处理日志流
6. 完成后下载视频

### 查看任务详情

- 任务历史列表
- 完整的 FFmpeg 命令（可复制）
- Filter graph（如果有）
- 完整的 stderr 日志
- 错误信息（如果失败）

## API 文档

### 创建任务

```bash
POST /api/tasks
Content-Type: application/json

{
  "type": "image_audio_to_video",
  "input_params": {
    "image_path": "/path/to/image.jpg",
    "audio_path": "/path/to/audio.mp3",
    "width": 1920,
    "height": 1080,
    "fps": 25,
    "video_codec": "libx264",
    "audio_codec": "aac",
    "video_bitrate": "2M",
    "audio_bitrate": "192k",
    "audio_loop": false,
    "output_format": "mp4"
  }
}
```

### 获取任务详情

```bash
GET /api/tasks/:id
```

### 实时监听进度（WebSocket）

```bash
GET /api/tasks/:id/progress
Upgrade: websocket
```

返回格式：
```json
{
  "task_id": "xxx",
  "status": "processing",
  "progress": 45.5,
  "current_frame": 1137,
  "total_frames": 2500,
  "eta": 120,
  "message": "Processing: 45.5% (Frame 1137/2500, Speed: 1.2x)",
  "timestamp": "2024-01-01T12:00:00Z"
}
```

### 上传文件

```bash
POST /api/upload
Content-Type: multipart/form-data

file: <binary>
```

## 环境变量

| 变量名 | 说明 | 默认值 |
|--------|------|--------|
| `SERVER_PORT` | 服务端口 | `8008` |
| `DATABASE_DSN` | PostgreSQL连接串 | `host=postgres user=ffmpeg password=ffmpeg dbname=ffmpeg port=5432 sslmode=disable` |
| `REDIS_ADDR` | Redis地址 | `redis:6379` |
| `UPLOAD_DIR` | 上传目录 | `./storage/uploads` |
| `OUTPUT_DIR` | 输出目录 | `./storage/outputs` |
| `FFMPEG_PATH` | FFmpeg路径 | `ffmpeg` |
| `FFMPEG_LOG_LEVEL` | 日志级别 | `info` |

## 数据模型

### Task（任务）

```go
type Task struct {
    ID            string    // 任务ID
    Type          string    // 任务类型
    Status        string    // 状态：pending/processing/completed/failed
    Progress      float64   // 进度 0-100
    CurrentFrame  int       // 当前帧
    TotalFrames   int       // 总帧数
    ETA           int       // 预计剩余秒数

    // 核心差异点字段
    FFmpegCommand string    // 完整命令（可回放）
    FilterGraph   string    // Filter graph
    StderrLog     string    // 完整日志（失败可解释）
    ErrorMessage  string    // 错误摘要

    InputParams   TaskInputParams  // 语义化参数
    OutputFile    string          // 输出文件路径
    OutputURL     string          // 下载URL

    CreatedAt     time.Time
    UpdatedAt     time.Time
}
```

## 性能优化建议

1. **并发处理**: 调整 Asynq worker 的 `Concurrency` 参数
2. **硬件加速**: 使用 FFmpeg 的硬件编码器（如 `h264_nvenc`）
3. **预设模板**: 为常用场景创建预设参数组合
4. **分布式部署**: 多个 Worker 节点处理任务
5. **对象存储**: 使用 OSS/S3 替代本地存储

## 扩展功能建议

- [ ] 视频拼接
- [ ] 添加字幕
- [ ] 视频转码
- [ ] 视频缩放
- [ ] 添加水印
- [ ] 转场效果
- [ ] 滤镜应用
- [ ] 批量处理
- [ ] 模板管理
- [ ] 用户权限管理

## 故障排查

### FFmpeg 执行失败

1. 查看任务详情中的 `stderr_log`
2. 复制 `ffmpeg_command` 在容器内手动执行：
   ```bash
   docker exec -it ffmpeg-backend sh
   # 粘贴命令执行
   ```

### WebSocket 连接失败

- 检查 nginx 配置的 WebSocket upgrade
- 查看浏览器 Network 面板的 WS 连接

### 数据库连接失败

```bash
docker-compose logs postgres
```

## 许可证

MIT License

## 贡献

欢迎提交 Issue 和 Pull Request！

---

**记住：这个平台的核心价值不是功能多，而是让每个操作都清晰可追溯！**
