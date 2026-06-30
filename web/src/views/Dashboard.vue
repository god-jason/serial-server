<template>
  <div class="dashboard">
    <aside class="sidebar">
      <div class="logo">
        <h2>串口服务器</h2>
      </div>
      <el-menu :default-active="activeMenu" router>
        <el-menu-item index="/dashboard/gateway">
          <el-icon><Location /></el-icon>
          <span>网关配置</span>
        </el-menu-item>
        <el-menu-item index="/dashboard/channels">
          <el-icon><Connection /></el-icon>
          <span>通道配置</span>
        </el-menu-item>
        <el-menu-item index="/dashboard/serial">
          <el-icon><DataLine /></el-icon>
          <span>串口配置</span>
        </el-menu-item>
        <el-menu-item index="/dashboard/device">
          <el-icon><Monitor /></el-icon>
          <span>设备信息</span>
        </el-menu-item>
        <el-menu-item index="/dashboard/maintenance">
          <el-icon><Setting /></el-icon>
          <span>系统维护</span>
        </el-menu-item>
        <el-menu-item index="/dashboard/debug">
          <el-icon><DataAnalysis /></el-icon>
          <span>串口调试</span>
        </el-menu-item>
        <el-menu-item index="/dashboard/terminal">
          <el-icon><DataBoard /></el-icon>
          <span>终端</span>
        </el-menu-item>
        <el-menu-item index="/dashboard/oem">
          <el-icon><OfficeBuilding /></el-icon>
          <span>OEM配置</span>
        </el-menu-item>
        <el-menu-item index="/dashboard/stats">
          <el-icon><PieChart /></el-icon>
          <span>流量统计</span>
        </el-menu-item>
      </el-menu>
      <div class="logout">
        <el-button @click="logout" text>退出登录</el-button>
      </div>
    </aside>
    <main class="main-content">
      <header class="header">
        <router-view v-slot="{ Component }">
          <h1>{{ getTitle(Component) }}</h1>
        </router-view>
      </header>
      <router-view />
    </main>
  </div>
</template>

<script setup>
import { computed } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import {
  Location,
  Connection,
  DataLine,
  Monitor,
  Setting,
  DataAnalysis,
  DataBoard,
  OfficeBuilding,
  PieChart
} from '@element-plus/icons-vue'

const router = useRouter()
const route = useRoute()

const activeMenu = computed(() => route.path)

const getTitle = (Component) => {
  const titles = {
    Gateway: '网关配置',
    Channels: '通道配置',
    Serial: '串口配置',
    Device: '设备信息',
    Maintenance: '系统维护',
    Debug: '串口调试',
    Terminal: '终端',
    OEM: 'OEM配置',
    Stats: '流量统计'
  }
  return titles[Component?.name] || ''
}

const logout = () => {
  localStorage.removeItem('token')
  router.push('/')
}
</script>

<style scoped>
.dashboard {
  display: flex;
  height: 100vh;
}

.sidebar {
  width: 200px;
  background: #263445;
  color: white;
  display: flex;
  flex-direction: column;
}

.logo {
  padding: 20px;
  text-align: center;
  border-bottom: 1px solid #3a4553;
}

.logo h2 {
  margin: 0;
  font-size: 16px;
}

.el-menu {
  border-right: none;
  flex: 1;
}

.el-menu-item {
  color: rgba(255, 255, 255, 0.8);
}

.el-menu-item.is-active {
  background: #1890ff;
  color: white;
}

.logout {
  padding: 20px;
  border-top: 1px solid #3a4553;
}

.main-content {
  flex: 1;
  display: flex;
  flex-direction: column;
  background: #f5f7fa;
  overflow: auto;
}

.header {
  padding: 16px 24px;
  background: white;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.08);
}

.header h1 {
  margin: 0;
  font-size: 20px;
  color: #333;
}
</style>
