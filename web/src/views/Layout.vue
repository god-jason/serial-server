<template>
  <div class="layout">
    <aside class="sidebar" :class="{ collapse: isCollapse }">
      <div class="logo">
        <div class="logo-icon">
          <el-icon :size="32"><DataLine /></el-icon>
        </div>
        <h2 v-show="!isCollapse">串口服务器</h2>
      </div>
      <el-menu
        :default-active="activeMenu"
        router
        :collapse="isCollapse"
        :collapse-transition="false"
      >
        <el-menu-item index="/home">
          <el-icon><House /></el-icon>
          <span>首页</span>
        </el-menu-item>

        <el-sub-menu index="gateway">
          <template #title>
            <el-icon><Share /></el-icon>
            <span>网关</span>
          </template>
          <el-menu-item index="/serial">
            <el-icon><DataLine /></el-icon>
            <span>串口</span>
          </el-menu-item>
          <el-menu-item index="/tcp-client">
            <el-icon><Connection /></el-icon>
            <span>TCP客户端</span>
          </el-menu-item>
          <el-menu-item index="/tcp-server">
            <el-icon><Box /></el-icon>
            <span>TCP服务端</span>
          </el-menu-item>
          <el-menu-item index="/mqtt">
            <el-icon><Cloudy /></el-icon>
            <span>MQTT</span>
          </el-menu-item>
          <el-menu-item index="/http">
            <el-icon><Link /></el-icon>
            <span>HTTP</span>
          </el-menu-item>
        </el-sub-menu>

        <el-sub-menu index="system">
          <template #title>
            <el-icon><Monitor /></el-icon>
            <span>系统</span>
          </template>
          <el-menu-item index="/network">
            <el-icon><Location /></el-icon>
            <span>网络</span>
          </el-menu-item>
          <el-menu-item index="/maintenance">
            <el-icon><Setting /></el-icon>
            <span>维护</span>
          </el-menu-item>
          <el-menu-item index="/stats">
            <el-icon><PieChart /></el-icon>
            <span>流量</span>
          </el-menu-item>
          <el-menu-item index="/logs">
            <el-icon><Document /></el-icon>
            <span>日志</span>
          </el-menu-item>
        </el-sub-menu>

        <el-sub-menu index="tools">
          <template #title>
            <el-icon><Tools /></el-icon>
            <span>工具</span>
          </template>
          <el-menu-item index="/debug">
            <el-icon><DataAnalysis /></el-icon>
            <span>串口调试</span>
          </el-menu-item>
          <el-menu-item index="/terminal">
            <el-icon><Cpu /></el-icon>
            <span>终端</span>
          </el-menu-item>
        </el-sub-menu>

        <el-menu-item @click="logout">
          <el-icon><SwitchButton /></el-icon>
          <span>退出登录</span>
        </el-menu-item>
      </el-menu>
    </aside>
    <main class="main-content" :class="{ collapse: isCollapse }">
      <header class="header">
        <div class="header-left">
          <el-button @click="toggleCollapse" circle size="small">
            <el-icon><Menu /></el-icon>
          </el-button>
          <h1>{{ currentTitle }}</h1>
        </div>
        <div class="header-right">
          <div class="system-info">
            <el-icon :size="16"><Monitor /></el-icon>
            <span>串口服务器 v1.0.0</span>
          </div>
        </div>
      </header>
      <div class="content-wrapper">
        <router-view />
      </div>
    </main>
  </div>
</template>

<script setup>
import { ref, computed } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { setLoggedIn } from '../stores/session'
import {
  DataLine,
  House,
  Connection,
  Monitor,
  Setting,
  PieChart,
  Document,
  Location,
  DataAnalysis,
  Tools,
  Menu,
  SwitchButton,
  Link,
  Box,
  Share,
  Cloudy,
  Cpu
} from '@element-plus/icons-vue'

const router = useRouter()
const route = useRoute()
const isCollapse = ref(false)

const menuTitles = {
  '/home': '首页',
  '/serial': '串口配置',
  '/tcp-client': 'TCP客户端',
  '/tcp-server': 'TCP服务端',
  '/mqtt': 'MQTT',
  '/http': 'HTTP通道',
  '/network': '网络配置',
  '/maintenance': '系统维护',
  '/stats': '流量统计',
  '/logs': '系统日志',
  '/debug': '串口调试',
  '/terminal': '终端'
}

const activeMenu = computed(() => {
  return route.path
})

const currentTitle = computed(() => {
  return menuTitles[route.path] || '首页'
})

const toggleCollapse = () => {
  isCollapse.value = !isCollapse.value
}

const logout = () => {
  setLoggedIn(false)
  router.push('/login')
}
</script>

<style scoped>
.layout {
  display: flex;
  height: 100vh;
  background: #f0f2f5;
}

.sidebar {
  width: 240px;
  background: #fff;
  display: flex;
  flex-direction: column;
  transition: width 0.3s ease;
  box-shadow: 2px 0 8px rgba(0, 0, 0, 0.1);
  position: fixed;
  left: 0;
  top: 0;
  bottom: 0;
  z-index: 100;
}

.sidebar.collapse {
  width: 64px;
}

.logo {
  padding: 20px;
  display: flex;
  align-items: center;
  gap: 12px;
  border-bottom: 1px solid #e6e6e6;
}

.logo-icon {
  width: 40px;
  height: 40px;
  background: linear-gradient(135deg, #409eff 0%, #667eea 100%);
  border-radius: 10px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: white;
}

.logo h2 {
  margin: 0;
  font-size: 18px;
  font-weight: 600;
  color: #303133;
}

.el-menu {
  border-right: none;
  flex: 1;
}

.el-menu-item.is-active {
  color: #409eff;
}

.main-content {
  flex: 1;
  display: flex;
  flex-direction: column;
  margin-left: 240px;
  transition: margin-left 0.3s ease;
}

.main-content.collapse {
  margin-left: 64px;
}

.header {
  padding: 16px 24px;
  background: white;
  display: flex;
  justify-content: space-between;
  align-items: center;
  box-shadow: 0 2px 12px rgba(0, 0, 0, 0.08);
}

.header-left {
  display: flex;
  align-items: center;
  gap: 16px;
}

.header h1 {
  margin: 0;
  font-size: 20px;
  font-weight: 600;
  color: #303133;
}

.header-right {
  display: flex;
  align-items: center;
  gap: 16px;
}

.system-info {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 14px;
  color: #909399;
  padding: 8px 16px;
  background: #f5f7fa;
  border-radius: 20px;
}

.content-wrapper {
  flex: 1;
  padding: 24px;
  overflow: auto;
}
</style>