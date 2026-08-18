import { createRouter, createWebHistory, type RouteRecordRaw } from 'vue-router'
import { getToken } from '@/api/request'

const routes: RouteRecordRaw[] = [
  {
    path: '/',
    redirect: '/classes'
  },
  {
    path: '/login',
    name: 'Login',
    component: () => import('@/views/login/index.vue'),
    meta: { title: '登录' }
  },
  {
    path: '/register',
    name: 'Register',
    component: () => import('@/views/register/index.vue'),
    meta: { title: '注册' }
  },
  {
    path: '/',
    component: () => import('@/layout/index.vue'),
    meta: { requireAuth: true },
    children: [
      {
        path: 'classes',
        name: 'Classes',
        component: () => import('@/views/classes/index.vue'),
        meta: { title: '我的班级', requireAuth: true }
      },
      {
        path: 'classes/:id',
        name: 'ClassDetail',
        component: () => import('@/views/classes/detail.vue'),
        meta: { title: '班级详情', requireAuth: true }
      },
      {
        path: 'course/:id',
        name: 'CourseLearn',
        component: () => import('@/views/course/learn.vue'),
        meta: { title: '课程学习', requireAuth: true }
      },
      {
        path: 'profile',
        name: 'Profile',
        component: () => import('@/views/profile/index.vue'),
        meta: { title: '个人中心', requireAuth: true }
      }
    ]
  },
  {
    path: '/:pathMatch(.*)*',
    redirect: '/classes'
  }
]

const router = createRouter({
  history: createWebHistory(),
  routes
})

router.beforeEach((to, _from, next) => {
  if (to.meta?.requireAuth && !getToken()) {
    next('/login')
  } else {
    next()
  }
})

export default router
