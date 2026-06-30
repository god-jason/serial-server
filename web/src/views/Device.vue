<template>
  <div class="device">
    <el-card>
      <template #header>
        <span>设备信息</span>
        <el-button @click="refresh" text style="float: right">刷新</el-button>
      </template>
      <el-descriptions :column="2" border>
        <el-descriptions-item label="主机名">{{ info.hostname || '-' }}</el-descriptions-item>
        <el-descriptions-item label="操作系统">{{ info.os || '-' }}</el-descriptions-item>
        <el-descriptions-item label="平台">{{ info.platform || '-' }}</el-descriptions-item>
        <el-descriptions-item label="内核版本">{{ info.kernel_version || '-' }}</el-descriptions-item>
        <el-descriptions-item label="运行时间">{{ formatUptime(info.uptime) }}</el-descriptions-item>
        <el-descriptions-item label="架构">{{ info.architecture || '-' }}</el-descriptions-item>
        <el-descriptions-item label="CPU型号">{{ info.cpu_model || '-' }}</el-descriptions-item>
        <el-descriptions-item label="CPU核心数">{{ info.cpu_cores || '-' }}</el-descriptions-item>
        <el-descriptions-item label="CPU使用率">{{ formatPercent(info.cpu_percent) }}</el-descriptions-item>
        <el-descriptions-item label="内存总量">{{ formatBytes(info.mem_total) }}</el-descriptions-item>
        <el-descriptions-item label="已用内存">{{ formatBytes(info.mem_used) }}</el-descriptions-item>
        <el-descriptions-item label="可用内存">{{ formatBytes(info.mem_free) }}</el-descriptions-item>
        <el-descriptions-item label="磁盘总量">{{ formatBytes(info.disk_total) }}</el-descriptions-item>
        <el-descriptions-item label="已用磁盘">{{ formatBytes(info.disk_used) }}</el-descriptions-item>
        <el-descriptions-item label="可用磁盘">{{ formatBytes(info.disk_free) }}</el-descriptions-item>
        <el-descriptions-item label="Go版本">{{ info.go_version || '-' }}</el-descriptions-item>
      </el-descriptions>
    </el-card>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import api from '../utils/api'

const info = ref({})

onMounted(async () => {
  await refresh()
})

const refresh = async () => {
  try {
    const response = await api.get('/system/info')
    info.value = response.data || {}
  } catch (error) {
    console.error('获取设备信息失败:', error)
    info.value = {}
  }
}

const formatUptime = (seconds) => {
  if (seconds === undefined || seconds === null || isNaN(seconds)) {
    return '-'
  }
  const days = Math.floor(seconds / 86400)
  const hours = Math.floor((seconds % 86400) / 3600)
  const minutes = Math.floor((seconds % 3600) / 60)
  const secs = Math.floor(seconds % 60)
  return `${days}天 ${hours}时 ${minutes}分 ${secs}秒`
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

const formatPercent = (value) => {
  if (value === undefined || value === null || isNaN(value)) {
    return '-'
  }
  return `${Number(value).toFixed(1)}%`
}
</script>

<style scoped>
.device {
  padding: 20px;
}
</style>
