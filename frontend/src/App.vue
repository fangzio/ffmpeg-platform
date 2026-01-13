<template>
  <router-view />
</template>

<script setup>
// App容器，使用路由视图
</script>

<style>
* {
  margin: 0;
  padding: 0;
  box-sizing: border-box;
}

body {
  font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial,
    sans-serif;
  -webkit-font-smoothing: antialiased;
  -moz-osx-font-smoothing: grayscale;
}

/* 覆盖 Element Plus 默认主题 */
:root {
  --el-color-primary: #6366f1;
  --el-color-primary-light-3: #8b5cf6;
  --el-color-primary-light-5: #a78bfa;
  --el-color-primary-light-7: #c4b5fd;
  --el-color-primary-light-9: #e0e7ff;
  --el-bg-color: rgba(30, 41, 59, 0.5);
  --el-bg-color-overlay: rgba(15, 23, 42, 0.8);
  --el-text-color-primary: #e0e7ff;
  --el-text-color-regular: #a5b4fc;
  --el-border-color: rgba(99, 102, 241, 0.2);
  --el-border-color-light: rgba(99, 102, 241, 0.1);
}

/* Element Plus 组件样式覆盖 */
.el-card {
  background: rgba(30, 41, 59, 0.5) !important;
  border-color: rgba(99, 102, 241, 0.2) !important;
  color: #e0e7ff !important;
}

.el-card__header {
  background: rgba(15, 23, 42, 0.5) !important;
  border-color: rgba(99, 102, 241, 0.2) !important;
  color: #e0e7ff !important;
}

.el-form-item__label {
  color: #a5b4fc !important;
}

.el-input__wrapper {
  background: rgba(15, 23, 42, 0.5) !important;
  border-color: rgba(99, 102, 241, 0.3) !important;
  box-shadow: 0 0 0 1px rgba(99, 102, 241, 0.3) inset !important;
}

.el-input__wrapper:hover {
  border-color: rgba(99, 102, 241, 0.5) !important;
}

.el-input__wrapper.is-focus {
  border-color: #6366f1 !important;
  box-shadow: 0 0 0 1px #6366f1 inset !important;
}

.el-input__inner {
  color: #e0e7ff !important;
}

.el-input__inner::placeholder {
  color: #6b7280 !important;
}

.el-select__wrapper {
  background: rgba(15, 23, 42, 0.5) !important;
  border-color: rgba(99, 102, 241, 0.3) !important;
}

.el-textarea__inner {
  background: rgba(15, 23, 42, 0.5) !important;
  border-color: rgba(99, 102, 241, 0.3) !important;
  color: #e0e7ff !important;
}

.el-button--primary {
  background: linear-gradient(135deg, #6366f1, #8b5cf6);
  border: none;
}

.el-table {
  background: transparent !important;
  color: #e0e7ff !important;
}

.el-table th.el-table__cell {
  background: rgba(15, 23, 42, 0.5) !important;
  color: #a5b4fc !important;
  border-color: rgba(99, 102, 241, 0.2) !important;
}

.el-table tr {
  background: transparent !important;
}

.el-table td.el-table__cell {
  border-color: rgba(99, 102, 241, 0.1) !important;
}

.el-table--enable-row-hover .el-table__body tr:hover > td {
  background: rgba(99, 102, 241, 0.1) !important;
}

.el-dialog {
  background: rgba(15, 23, 42, 0.95) !important;
  border: 1px solid rgba(99, 102, 241, 0.3) !important;
  backdrop-filter: blur(10px);
}

.el-dialog__header {
  border-color: rgba(99, 102, 241, 0.2) !important;
}

.el-dialog__title {
  color: #e0e7ff !important;
}

.el-descriptions__label {
  background: rgba(99, 102, 241, 0.1) !important;
  color: #a5b4fc !important;
}

.el-descriptions__content {
  background: rgba(15, 23, 42, 0.3) !important;
  color: #e0e7ff !important;
}

/* 滚动条样式 */
::-webkit-scrollbar {
  width: 8px;
  height: 8px;
}

::-webkit-scrollbar-track {
  background: rgba(15, 23, 42, 0.5);
}

::-webkit-scrollbar-thumb {
  background: rgba(99, 102, 241, 0.5);
  border-radius: 4px;
}

::-webkit-scrollbar-thumb:hover {
  background: rgba(99, 102, 241, 0.7);
}
</style>
