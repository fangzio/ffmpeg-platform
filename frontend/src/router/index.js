import { createRouter, createWebHistory } from 'vue-router'
import Layout from '@/components/Layout.vue'
import Home from '@/views/Home.vue'
import TaskCreate from '@/views/TaskCreate.vue'
import ImageSlideshow from '@/views/ImageSlideshow.vue'

const routes = [
  {
    path: '/',
    component: Layout,
    children: [
      {
        path: '',
        name: 'Home',
        component: Home,
        meta: { title: '首页' }
      },
      {
        path: '/tools/image-audio',
        name: 'ImageAudio',
        component: TaskCreate,
        meta: { title: '图片+音频视频' }
      },
      {
        path: '/tools/slideshow',
        name: 'ImageSlideshow',
        component: ImageSlideshow,
        meta: { title: '图片轮播视频' }
      }
    ]
  }
]

const router = createRouter({
  history: createWebHistory(),
  routes
})

export default router
