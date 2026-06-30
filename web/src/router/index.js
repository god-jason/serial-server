import { createRouter, createWebHistory } from 'vue-router'

const routes = [
  {
    path: '/',
    name: 'Login',
    component: () => import('../views/Login.vue')
  },
  {
    path: '/dashboard',
    name: 'Dashboard',
    component: () => import('../views/Dashboard.vue'),
    children: [
      { path: 'gateway', name: 'Gateway', component: () => import('../views/Gateway.vue') },
      { path: 'channels', name: 'Channels', component: () => import('../views/Channels.vue') },
      { path: 'serial', name: 'Serial', component: () => import('../views/Serial.vue') },
      { path: 'device', name: 'Device', component: () => import('../views/Device.vue') },
      { path: 'maintenance', name: 'Maintenance', component: () => import('../views/Maintenance.vue') },
      { path: 'debug', name: 'Debug', component: () => import('../views/Debug.vue') },
      { path: 'terminal', name: 'Terminal', component: () => import('../views/Terminal.vue') },
      { path: 'oem', name: 'OEM', component: () => import('../views/OEM.vue') },
      { path: 'stats', name: 'Stats', component: () => import('../views/Stats.vue') }
    ]
  }
]

const router = createRouter({
  history: createWebHistory(),
  routes
})

router.beforeEach((to, from, next) => {
  const token = localStorage.getItem('token')
  if (to.name !== 'Login' && !token) {
    next({ name: 'Login' })
  } else {
    next()
  }
})

export default router
