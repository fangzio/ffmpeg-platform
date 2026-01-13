<template>
  <div class="home-page">
    <!-- Hero Section -->
    <section class="hero-section">
      <div class="hero-content">
        <div class="hero-badge">
          <span class="badge-dot"></span>
          FFmpeg-based Video Processing
        </div>
        <h1 class="hero-title">
          专业级<span class="highlight">视频处理</span>工具平台
        </h1>
        <p class="hero-subtitle">
          基于 FFmpeg 构建的强大视频处理平台，支持多种视频生成和编辑功能
        </p>
        <div class="hero-actions">
          <button class="btn-primary" @click="$router.push('/tools/image-audio')">
            <el-icon><VideoCamera /></el-icon>
            开始创作
          </button>
          <button class="btn-secondary" @click="scrollToTools">
            <el-icon><Compass /></el-icon>
            探索工具
          </button>
        </div>
      </div>
      <div class="hero-visual">
        <div class="visual-card">
          <div class="card-header">
            <span class="card-dot"></span>
            <span class="card-dot"></span>
            <span class="card-dot"></span>
          </div>
          <div class="card-body">
            <div class="code-line" v-for="i in 8" :key="i"></div>
          </div>
        </div>
      </div>
    </section>

    <!-- Features Section -->
    <section class="features-section" ref="toolsSection">
      <h2 class="section-title">
        <span class="title-line"></span>
        功能工具
        <span class="title-line"></span>
      </h2>

      <div class="features-grid">
        <div class="feature-card" @click="$router.push('/tools/image-audio')">
          <div class="card-icon">
            <el-icon><Picture /></el-icon>
          </div>
          <h3>图片+音频视频</h3>
          <p>将单张图片与音频合成为视频，支持自定义尺寸、码率、编码格式</p>
          <div class="card-footer">
            <span class="tag">静态图片</span>
            <span class="tag">音频合成</span>
          </div>
          <div class="card-glow"></div>
        </div>

        <div class="feature-card" @click="$router.push('/tools/slideshow')">
          <div class="card-icon">
            <el-icon><Film /></el-icon>
          </div>
          <h3>图片轮播视频</h3>
          <p>将多张图片制作成轮播视频，支持转场效果、背景音乐、自定义时长</p>
          <div class="card-footer">
            <span class="tag">多图轮播</span>
            <span class="tag">转场效果</span>
          </div>
          <div class="card-glow"></div>
        </div>

        <div class="feature-card disabled">
          <div class="card-icon">
            <el-icon><VideoPlay /></el-icon>
          </div>
          <h3>视频编辑</h3>
          <p>视频剪辑、合并、转码等高级功能</p>
          <div class="card-footer">
            <span class="tag">即将推出</span>
          </div>
          <div class="card-coming-soon">COMING SOON</div>
        </div>

        <div class="feature-card disabled">
          <div class="card-icon">
            <el-icon><Microphone /></el-icon>
          </div>
          <h3>音频处理</h3>
          <p>音频提取、转换、混音等功能</p>
          <div class="card-footer">
            <span class="tag">即将推出</span>
          </div>
          <div class="card-coming-soon">COMING SOON</div>
        </div>
      </div>
    </section>

    <!-- Tech Stack Section -->
    <section class="tech-section">
      <h2 class="section-title">
        <span class="title-line"></span>
        技术栈
        <span class="title-line"></span>
      </h2>
      <div class="tech-stack">
        <div class="tech-item">
          <div class="tech-icon">FFmpeg</div>
          <p>核心视频处理引擎</p>
        </div>
        <div class="tech-item">
          <div class="tech-icon">Go</div>
          <p>高性能后端服务</p>
        </div>
        <div class="tech-item">
          <div class="tech-icon">Vue 3</div>
          <p>现代化前端框架</p>
        </div>
        <div class="tech-item">
          <div class="tech-icon">WebSocket</div>
          <p>实时进度推送</p>
        </div>
      </div>
    </section>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import { VideoCamera, Compass, Picture, Film, VideoPlay, Microphone } from '@element-plus/icons-vue'

const toolsSection = ref(null)

const scrollToTools = () => {
  toolsSection.value?.scrollIntoView({ behavior: 'smooth' })
}
</script>

<style scoped>
.home-page {
  width: 100%;
}

/* Hero Section */
.hero-section {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 60px;
  align-items: center;
  padding: 80px 0;
  min-height: 600px;
}

.hero-content {
  animation: fadeInUp 0.8s ease-out;
}

@keyframes fadeInUp {
  from {
    opacity: 0;
    transform: translateY(30px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

.hero-badge {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  padding: 8px 16px;
  background: rgba(99, 102, 241, 0.1);
  border: 1px solid rgba(99, 102, 241, 0.3);
  border-radius: 20px;
  color: #a5b4fc;
  font-size: 13px;
  font-weight: 500;
  margin-bottom: 24px;
}

.badge-dot {
  width: 6px;
  height: 6px;
  background: #6366f1;
  border-radius: 50%;
  animation: pulse 2s ease-in-out infinite;
}

.hero-title {
  font-size: 56px;
  font-weight: 700;
  line-height: 1.2;
  margin: 0 0 24px 0;
  color: #e0e7ff;
  letter-spacing: -0.5px;
}

.highlight {
  background: linear-gradient(135deg, #6366f1, #8b5cf6, #ec4899);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
}

.hero-subtitle {
  font-size: 18px;
  line-height: 1.8;
  color: #a5b4fc;
  margin: 0 0 40px 0;
}

.hero-actions {
  display: flex;
  gap: 16px;
}

.btn-primary,
.btn-secondary {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  padding: 14px 32px;
  border-radius: 10px;
  font-size: 16px;
  font-weight: 600;
  cursor: pointer;
  border: none;
  transition: all 0.3s ease;
}

.btn-primary {
  background: linear-gradient(135deg, #6366f1, #8b5cf6);
  color: white;
  box-shadow: 0 4px 20px rgba(99, 102, 241, 0.4);
}

.btn-primary:hover {
  transform: translateY(-2px);
  box-shadow: 0 6px 30px rgba(99, 102, 241, 0.6);
}

.btn-secondary {
  background: rgba(99, 102, 241, 0.1);
  color: #e0e7ff;
  border: 1px solid rgba(99, 102, 241, 0.3);
}

.btn-secondary:hover {
  background: rgba(99, 102, 241, 0.2);
  border-color: rgba(99, 102, 241, 0.5);
  transform: translateY(-2px);
}

/* Hero Visual */
.hero-visual {
  animation: fadeInUp 0.8s ease-out 0.2s backwards;
}

.visual-card {
  background: rgba(30, 41, 59, 0.5);
  border: 1px solid rgba(99, 102, 241, 0.2);
  border-radius: 16px;
  padding: 16px;
  box-shadow: 0 8px 32px rgba(0, 0, 0, 0.3);
  backdrop-filter: blur(10px);
}

.card-header {
  display: flex;
  gap: 8px;
  margin-bottom: 16px;
}

.card-dot {
  width: 12px;
  height: 12px;
  border-radius: 50%;
  background: rgba(99, 102, 241, 0.3);
}

.card-dot:nth-child(1) {
  background: #ef4444;
}
.card-dot:nth-child(2) {
  background: #f59e0b;
}
.card-dot:nth-child(3) {
  background: #22c55e;
}

.card-body {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.code-line {
  height: 20px;
  background: linear-gradient(90deg, rgba(99, 102, 241, 0.2), transparent);
  border-radius: 4px;
  animation: codePulse 3s ease-in-out infinite;
}

.code-line:nth-child(odd) {
  width: 80%;
}

.code-line:nth-child(even) {
  width: 60%;
}

@keyframes codePulse {
  0%,
  100% {
    opacity: 0.5;
  }
  50% {
    opacity: 1;
  }
}

/* Features Section */
.features-section,
.tech-section {
  padding: 80px 0;
}

.section-title {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 20px;
  font-size: 36px;
  font-weight: 700;
  color: #e0e7ff;
  margin: 0 0 60px 0;
  text-align: center;
}

.title-line {
  height: 2px;
  width: 60px;
  background: linear-gradient(90deg, transparent, #6366f1, transparent);
}

.features-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(280px, 1fr));
  gap: 24px;
}

.feature-card {
  position: relative;
  padding: 32px;
  background: rgba(30, 41, 59, 0.5);
  border: 1px solid rgba(99, 102, 241, 0.2);
  border-radius: 16px;
  cursor: pointer;
  transition: all 0.3s ease;
  overflow: hidden;
}

.feature-card::before {
  content: '';
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: linear-gradient(135deg, rgba(99, 102, 241, 0.1), rgba(139, 92, 246, 0.1));
  opacity: 0;
  transition: opacity 0.3s ease;
}

.feature-card:hover::before {
  opacity: 1;
}

.feature-card:hover {
  transform: translateY(-4px);
  border-color: rgba(99, 102, 241, 0.5);
  box-shadow: 0 8px 32px rgba(99, 102, 241, 0.3);
}

.feature-card.disabled {
  cursor: not-allowed;
  opacity: 0.6;
}

.feature-card.disabled:hover {
  transform: none;
  border-color: rgba(99, 102, 241, 0.2);
  box-shadow: none;
}

.card-icon {
  width: 60px;
  height: 60px;
  background: linear-gradient(135deg, #6366f1, #8b5cf6);
  border-radius: 12px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 28px;
  color: white;
  margin-bottom: 20px;
}

.feature-card h3 {
  font-size: 22px;
  font-weight: 600;
  color: #e0e7ff;
  margin: 0 0 12px 0;
}

.feature-card p {
  font-size: 15px;
  line-height: 1.6;
  color: #a5b4fc;
  margin: 0 0 20px 0;
}

.card-footer {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
}

.tag {
  padding: 4px 12px;
  background: rgba(99, 102, 241, 0.15);
  border: 1px solid rgba(99, 102, 241, 0.3);
  border-radius: 12px;
  font-size: 12px;
  color: #a5b4fc;
}

.card-glow {
  position: absolute;
  bottom: -50%;
  left: -50%;
  width: 200%;
  height: 200%;
  background: radial-gradient(circle, rgba(99, 102, 241, 0.1), transparent);
  opacity: 0;
  transition: opacity 0.3s ease;
}

.feature-card:hover .card-glow {
  opacity: 1;
}

.card-coming-soon {
  position: absolute;
  top: 16px;
  right: 16px;
  padding: 6px 12px;
  background: rgba(139, 92, 246, 0.2);
  border: 1px solid rgba(139, 92, 246, 0.4);
  border-radius: 8px;
  font-size: 11px;
  font-weight: 600;
  color: #c4b5fd;
  letter-spacing: 0.5px;
}

/* Tech Stack */
.tech-stack {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
  gap: 32px;
}

.tech-item {
  text-align: center;
  padding: 24px;
  background: rgba(30, 41, 59, 0.3);
  border: 1px solid rgba(99, 102, 241, 0.2);
  border-radius: 12px;
  transition: all 0.3s ease;
}

.tech-item:hover {
  border-color: rgba(99, 102, 241, 0.4);
  transform: translateY(-4px);
}

.tech-icon {
  font-size: 24px;
  font-weight: 700;
  color: #6366f1;
  margin-bottom: 12px;
  font-family: 'Courier New', monospace;
}

.tech-item p {
  font-size: 14px;
  color: #a5b4fc;
  margin: 0;
}

/* 响应式 */
@media (max-width: 1024px) {
  .hero-section {
    grid-template-columns: 1fr;
    gap: 40px;
  }

  .hero-title {
    font-size: 42px;
  }
}

@media (max-width: 640px) {
  .hero-section {
    padding: 40px 0;
  }

  .hero-title {
    font-size: 32px;
  }

  .section-title {
    font-size: 28px;
  }

  .features-grid {
    grid-template-columns: 1fr;
  }
}
</style>
