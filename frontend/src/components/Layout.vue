<template>
  <div class="tech-layout">
    <!-- 顶部导航 -->
    <header class="tech-header">
      <div class="header-content">
        <div class="logo" @click="$router.push('/')">
          <div class="logo-icon">
            <span class="icon-bracket">[</span>
            <span class="icon-text">FP</span>
            <span class="icon-bracket">]</span>
          </div>
          <span class="logo-text">FFmpeg Platform</span>
        </div>

        <nav class="nav-menu">
          <router-link to="/" class="nav-item" active-class="active">
            <el-icon><HomeFilled /></el-icon>
            <span>首页</span>
          </router-link>
          <router-link to="/tools/image-audio" class="nav-item" active-class="active">
            <el-icon><Picture /></el-icon>
            <span>图片+音频</span>
          </router-link>
          <router-link to="/tools/slideshow" class="nav-item" active-class="active">
            <el-icon><Film /></el-icon>
            <span>图片轮播</span>
          </router-link>
        </nav>

        <div class="header-actions">
          <div class="status-indicator">
            <div class="status-dot"></div>
            <span>系统运行中</span>
          </div>
        </div>
      </div>
      <div class="header-glow"></div>
    </header>

    <!-- 主内容区 -->
    <main class="tech-main">
      <div class="tech-container">
        <router-view v-slot="{ Component }">
          <transition name="page-fade" mode="out-in">
            <component :is="Component" />
          </transition>
        </router-view>
      </div>
    </main>

    <!-- 底部 -->
    <footer class="tech-footer">
      <div class="footer-content">
        <p>FFmpeg Platform - 专业视频处理工具平台</p>
        <p class="footer-tech">Powered by FFmpeg & Vue 3</p>
      </div>
    </footer>

    <!-- 背景装饰 -->
    <div class="tech-bg">
      <div class="grid-line"></div>
      <div class="grid-line"></div>
      <div class="grid-line"></div>
    </div>
  </div>
</template>

<script setup>
import { HomeFilled, Picture, Film } from '@element-plus/icons-vue'
</script>

<style scoped>
.tech-layout {
  min-height: 100vh;
  background: linear-gradient(135deg, #0a0e27 0%, #1a1f3a 50%, #0f1423 100%);
  color: #e0e7ff;
  position: relative;
  overflow-x: hidden;
}

/* 背景网格装饰 */
.tech-bg {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  pointer-events: none;
  z-index: 0;
  overflow: hidden;
}

.grid-line {
  position: absolute;
  background: linear-gradient(
    90deg,
    transparent 0%,
    rgba(99, 102, 241, 0.1) 50%,
    transparent 100%
  );
  height: 1px;
  width: 100%;
  animation: gridMove 20s linear infinite;
}

.grid-line:nth-child(1) {
  top: 20%;
  animation-delay: 0s;
}

.grid-line:nth-child(2) {
  top: 50%;
  animation-delay: -7s;
}

.grid-line:nth-child(3) {
  top: 80%;
  animation-delay: -14s;
}

@keyframes gridMove {
  from {
    transform: translateX(-100%);
  }
  to {
    transform: translateX(100%);
  }
}

/* 顶部导航 */
.tech-header {
  position: sticky;
  top: 0;
  z-index: 1000;
  background: rgba(10, 14, 39, 0.8);
  backdrop-filter: blur(20px);
  border-bottom: 1px solid rgba(99, 102, 241, 0.2);
  box-shadow: 0 4px 30px rgba(99, 102, 241, 0.1);
}

.header-glow {
  position: absolute;
  bottom: -2px;
  left: 0;
  right: 0;
  height: 2px;
  background: linear-gradient(
    90deg,
    transparent,
    #6366f1,
    #8b5cf6,
    #6366f1,
    transparent
  );
  animation: headerGlow 3s ease-in-out infinite;
}

@keyframes headerGlow {
  0%,
  100% {
    opacity: 0.5;
  }
  50% {
    opacity: 1;
  }
}

.header-content {
  max-width: 1400px;
  margin: 0 auto;
  padding: 0 24px;
  height: 70px;
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.logo {
  display: flex;
  align-items: center;
  gap: 12px;
  cursor: pointer;
  transition: all 0.3s ease;
}

.logo:hover {
  transform: translateY(-2px);
}

.logo-icon {
  font-family: 'Courier New', monospace;
  font-size: 24px;
  font-weight: bold;
  color: #6366f1;
  display: flex;
  align-items: center;
}

.icon-bracket {
  color: #8b5cf6;
  animation: bracket-pulse 2s ease-in-out infinite;
}

.icon-text {
  margin: 0 2px;
  background: linear-gradient(135deg, #6366f1, #8b5cf6);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
}

@keyframes bracket-pulse {
  0%,
  100% {
    opacity: 1;
  }
  50% {
    opacity: 0.5;
  }
}

.logo-text {
  font-size: 18px;
  font-weight: 600;
  background: linear-gradient(135deg, #e0e7ff, #c7d2fe);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
  letter-spacing: 0.5px;
}

.nav-menu {
  display: flex;
  gap: 8px;
}

.nav-item {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 10px 20px;
  border-radius: 8px;
  color: #a5b4fc;
  text-decoration: none;
  font-size: 15px;
  font-weight: 500;
  transition: all 0.3s ease;
  position: relative;
  overflow: hidden;
}

.nav-item::before {
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

.nav-item:hover::before {
  opacity: 1;
}

.nav-item:hover {
  color: #e0e7ff;
  transform: translateY(-2px);
}

.nav-item.active {
  color: #e0e7ff;
  background: linear-gradient(135deg, rgba(99, 102, 241, 0.2), rgba(139, 92, 246, 0.2));
  box-shadow: 0 0 20px rgba(99, 102, 241, 0.3);
}

.header-actions {
  display: flex;
  align-items: center;
  gap: 16px;
}

.status-indicator {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 16px;
  background: rgba(34, 197, 94, 0.1);
  border: 1px solid rgba(34, 197, 94, 0.3);
  border-radius: 20px;
  font-size: 13px;
  color: #86efac;
}

.status-dot {
  width: 8px;
  height: 8px;
  background: #22c55e;
  border-radius: 50%;
  animation: pulse 2s ease-in-out infinite;
}

@keyframes pulse {
  0%,
  100% {
    box-shadow: 0 0 0 0 rgba(34, 197, 94, 0.7);
  }
  50% {
    box-shadow: 0 0 0 6px rgba(34, 197, 94, 0);
  }
}

/* 主内容区 */
.tech-main {
  position: relative;
  z-index: 1;
  min-height: calc(100vh - 140px);
  padding: 40px 0;
}

.tech-container {
  max-width: 1400px;
  margin: 0 auto;
  padding: 0 24px;
}

/* 页面切换动画 */
.page-fade-enter-active,
.page-fade-leave-active {
  transition: all 0.3s ease;
}

.page-fade-enter-from {
  opacity: 0;
  transform: translateY(20px);
}

.page-fade-leave-to {
  opacity: 0;
  transform: translateY(-20px);
}

/* 底部 */
.tech-footer {
  position: relative;
  z-index: 1;
  padding: 30px 0;
  border-top: 1px solid rgba(99, 102, 241, 0.2);
  background: rgba(10, 14, 39, 0.6);
  backdrop-filter: blur(10px);
}

.footer-content {
  max-width: 1400px;
  margin: 0 auto;
  padding: 0 24px;
  text-align: center;
}

.footer-content p {
  margin: 8px 0;
  color: #a5b4fc;
  font-size: 14px;
}

.footer-tech {
  font-size: 12px;
  color: #6b7280;
  font-family: 'Courier New', monospace;
}

/* 响应式 */
@media (max-width: 768px) {
  .header-content {
    padding: 0 16px;
  }

  .nav-menu {
    gap: 4px;
  }

  .nav-item {
    padding: 8px 12px;
    font-size: 13px;
  }

  .nav-item span {
    display: none;
  }

  .status-indicator span {
    display: none;
  }

  .tech-container {
    padding: 0 16px;
  }
}
</style>
