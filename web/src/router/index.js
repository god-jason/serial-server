import { createRouter, createWebHistory } from 'vue-router'
import { isLoggedIn, checkLoginStatus } from '../stores/session'

const routes = [
  {
    path: '/',
    redirect: '/home'
  },
  {
    path: '/login',
    name: 'Login',
    component: () => import('../views/Login.vue'),
    meta: { requiresAuth: false }
  },
  {
    path: '/',
    component: () => import('../views/Layout.vue'),
    meta: { requiresAuth: true },
    children: [
      {
        path: 'home',
        name: 'Home',
        component: () => import('../views/Home.vue'),
        meta: { title: '首页' }
      },
      {
        path: 'serial',
        name: 'Serial',
        component: () => import('../views/Serial.vue'),
        meta: { title: '串口配置' }
      },
      {
        path: 'tcp-client',
        name: 'TcpClient',
        component: () => import('../views/TcpClient.vue'),
        meta: { title: 'TCP客户端' }
      },
      {
        path: 'tcp-server',
        name: 'TcpServer',
        component: () => import('../views/TcpServer.vue'),
        meta: { title: 'TCP服务端' }
      },
      {
        path: 'mqtt',
        name: 'Mqtt',
        component: () => import('../views/Mqtt.vue'),
        meta: { title: 'MQTT' }
      },
      {
        path: 'http',
        name: 'Http',
        component: () => import('../views/Http.vue'),
        meta: { title: 'HTTP' }
      },
      {
        path: 'network',
        name: 'Network',
        component: () => import('../views/Network.vue'),
        meta: { title: '网络配置' }
      },
      {
        path: 'maintenance',
        name: 'Maintenance',
        component: () => import('../views/Maintenance.vue'),
        meta: { title: '系统维护' }
      },
      {
        path: 'stats',
        name: 'Stats',
        component: () => import('../views/Stats.vue'),
        meta: { title: '流量统计' }
      },
      {
        path: 'logs',
        name: 'Logs',
        component: () => import('../views/Logs.vue'),
        meta: { title: '系统日志' }
      },
      {
        path: 'debug',
        name: 'Debug',
        component: () => import('../views/Debug.vue'),
        meta: { title: '串口调试' }
      },
      {
        path: 'terminal',
        name: 'Terminal',
        component: () => import('../views/Terminal.vue'),
        meta: { title: '终端' }
      }
    ]
  }
]

const router = createRouter({
  history: createWebHistory(),
  routes
})

let initCheckDone = false

router.beforeEach(async (to, from, next) => {
  const requiresAuth = to.matched.some(record => record.meta.requiresAuth !== false)

  if (!initCheckDone && requiresAuth) {
    await checkLoginStatus()
    initCheckDone = true
  }

  if (requiresAuth && !isLoggedIn()) {
    next('/login')
  } else if (to.path === '/login' && isLoggedIn()) {
    next('/home')
  } else {
    next()
  }
})

export default router