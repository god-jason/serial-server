<template>
  <div class="stats">
    <el-card>
      <template #header>
        <span>流量统计</span>
        <el-button @click="refresh" text style="float: right">刷新</el-button>
      </template>
      <el-row :gutter="20">
        <el-col :span="6">
          <el-card shadow="hover">
            <div class="stat-item">
              <span class="stat-label">串口发送</span>
              <span class="stat-value">{{ formatBytes(stats.serial_tx) }}</span>
            </div>
          </el-card>
        </el-col>
        <el-col :span="6">
          <el-card shadow="hover">
            <div class="stat-item">
              <span class="stat-label">串口接收</span>
              <span class="stat-value">{{ formatBytes(stats.serial_rx) }}</span>
            </div>
          </el-card>
        </el-col>
        <el-col :span="6">
          <el-card shadow="hover">
            <div class="stat-item">
              <span class="stat-label">网络发送</span>
              <span class="stat-value">{{ formatBytes(stats.network_tx) }}</span>
            </div>
          </el-card>
        </el-col>
        <el-col :span="6">
          <el-card shadow="hover">
            <div class="stat-item">
              <span class="stat-label">网络接收</span>
              <span class="stat-value">{{ formatBytes(stats.network_rx) }}</span>
            </div>
          </el-card>
        </el-col>
      </el-row>
      <el-row style="margin-top: 20px">
        <el-col :span="6">
          <el-card shadow="hover">
            <div class="stat-item">
              <span class="stat-label">缓存数量</span>
              <span class="stat-value">{{ stats.cache_count }}</span>
            </div>
          </el-card>
        </el-col>
        <el-col :span="6">
          <el-card shadow="hover">
            <div class="stat-item">
              <span class="stat-label">重发次数</span>
              <span class="stat-value">{{ stats.resend_count }}</span>
            </div>
          </el-card>
        </el-col>
        <el-col :span="6">
          <el-card shadow="hover">
            <div class="stat-item">
              <span class="stat-label">在线通道</span>
              <span class="stat-value">{{ stats.connected_channels }}</span>
            </div>
          </el-card>
        </el-col>
        <el-col :span="6">
          <el-card shadow="hover">
            <div class="stat-item">
              <span class="stat-label">打开串口</span>
              <span class="stat-value">{{ stats.opened_ports }}</span>
            </div>
          </el-card>
        </el-col>
      </el-row>
    </el-card>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import api from '../utils/api'

const stats = ref({
  serial_tx: 0,
  serial_rx: 0,
  network_tx: 0,
  network_rx: 0,
  cache_count: 0,
  resend_count: 0,
  connected_channels: 0,
  opened_ports: 0
})

onMounted(async () => {
  await refresh()
})

const refresh = async () => {
  try {
    const response = await api.get('/stats')
    stats.value = response.data
  } catch (error) {
    console.error('获取统计数据失败:', error)
  }
}

const formatBytes = (bytes) => {
  if (bytes === undefined || bytes === null || isNaN(bytes) || bytes < 0) {
    return '-'
  }
  if (bytes === 0) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i]
}
</script>

<style scoped>
.stats {
  padding: 20px;
}

.stat-item {
  display: flex;
  flex-direction: column;
  align-items: center;
}

.stat-label {
  font-size: 14px;
  color: #666;
  margin-bottom: 8px;
}

.stat-value {
  font-size: 24px;
  font-weight: bold;
  color: #333;
}
</style>
