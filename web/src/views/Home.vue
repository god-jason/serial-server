<template>
  <div class="dashboard-home">
    <el-row :gutter="20">
      <el-col :span="6">
        <el-card class="status-card" shadow="hover">
          <div class="status-header">
            <el-icon class="status-icon status-online"><CircleCheck /></el-icon>
            <span class="status-label">网关状态</span>
          </div>
          <div class="status-value">运行中</div>
          <div class="status-desc">系统正常运行</div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card class="status-card" shadow="hover">
          <div class="status-header">
            <el-icon class="status-icon status-cpu"><Cpu /></el-icon>
            <span class="status-label">CPU使用率</span>
          </div>
          <div class="status-value">{{ formatPercent(info.cpu_percent) }}</div>
          <div class="status-desc">{{ info.cpu_cores || '-' }} 核心</div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card class="status-card" shadow="hover">
          <div class="status-header">
            <el-icon class="status-icon status-mem"><DataBoard /></el-icon>
            <span class="status-label">内存使用</span>
          </div>
          <div class="status-value">{{ formatBytes(info.mem_used || 0) }}</div>
          <div class="status-desc">总计 {{ formatBytes(info.mem_total || 0) }}</div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card class="status-card" shadow="hover">
          <div class="status-header">
            <el-icon class="status-icon status-disk"><Monitor /></el-icon>
            <span class="status-label">磁盘使用</span>
          </div>
          <div class="status-value">{{ formatBytes(info.disk_used || 0) }}</div>
          <div class="status-desc">总计 {{ formatBytes(info.disk_total || 0) }}</div>
        </el-card>
      </el-col>
    </el-row>

    <el-row :gutter="20" style="margin-top: 20px">
      <el-col :span="24">
        <el-card class="info-card">
          <template #header>
            <div class="card-header">
              <el-icon class="card-icon"><Monitor /></el-icon>
              <span>系统信息</span>
              <el-button @click="refreshInfo" text size="small" style="margin-left: auto">刷新</el-button>
            </div>
          </template>
          <el-descriptions :column="4" border size="small">
            <el-descriptions-item label="主机名">{{ info.hostname || '-' }}</el-descriptions-item>
            <el-descriptions-item label="操作系统">{{ info.os || '-' }}</el-descriptions-item>
            <el-descriptions-item label="内核版本">{{ info.kernel_version || '-' }}</el-descriptions-item>
            <el-descriptions-item label="Go版本">{{ info.go_version || '-' }}</el-descriptions-item>
            <el-descriptions-item label="运行时间">{{ formatUptime(info.uptime) }}</el-descriptions-item>
            <el-descriptions-item label="架构">{{ info.architecture || '-' }}</el-descriptions-item>
            <el-descriptions-item label="CPU型号">{{ info.cpu_model || '-' }}</el-descriptions-item>
            <el-descriptions-item label="CPU核心">{{ info.cpu_cores || '-' }} 核心</el-descriptions-item>
            <el-descriptions-item label="内存总量">{{ formatBytes(info.mem_total || 0) }}</el-descriptions-item>
            <el-descriptions-item label="内存使用">{{ formatPercent(info.mem_percent) }}</el-descriptions-item>
            <el-descriptions-item label="磁盘总量">{{ formatBytes(info.disk_total || 0) }}</el-descriptions-item>
            <el-descriptions-item label="磁盘使用">{{ formatPercent(info.disk_percent) }}</el-descriptions-item>
            <el-descriptions-item label="软件版本">{{ info.version || 'v1.0.0' }}</el-descriptions-item>
            <el-descriptions-item label="编译时间">{{ info.build_time || '-' }}</el-descriptions-item>
            <el-descriptions-item label="启动时间">{{ info.start_time || '-' }}</el-descriptions-item>
            <el-descriptions-item label="平台">{{ info.platform || '-' }}</el-descriptions-item>
          </el-descriptions>
        </el-card>
      </el-col>
    </el-row>

    <el-row :gutter="20" style="margin-top: 20px">
      <el-col :span="24">
        <el-card class="info-card">
          <template #header>
            <div class="card-header">
              <el-icon class="card-icon"><DataLine /></el-icon>
              <span>流量统计</span>
            </div>
          </template>
          <el-row :gutter="20">
            <el-col :span="6">
              <div class="stat-box">
                <span class="stat-label">串口发送</span>
                <span class="stat-num">{{ formatBytes(stats.serial_tx || 0) }}</span>
              </div>
            </el-col>
            <el-col :span="6">
              <div class="stat-box">
                <span class="stat-label">串口接收</span>
                <span class="stat-num">{{ formatBytes(stats.serial_rx || 0) }}</span>
              </div>
            </el-col>
            <el-col :span="6">
              <div class="stat-box">
                <span class="stat-label">网络发送</span>
                <span class="stat-num">{{ formatBytes(stats.network_tx || 0) }}</span>
              </div>
            </el-col>
            <el-col :span="6">
              <div class="stat-box">
                <span class="stat-label">网络接收</span>
                <span class="stat-num">{{ formatBytes(stats.network_rx || 0) }}</span>
              </div>
            </el-col>
          </el-row>
          <el-row :gutter="20" style="margin-top: 12px">
            <el-col :span="6">
              <div class="stat-box">
                <span class="stat-label">缓存数量</span>
                <span class="stat-num">{{ stats.cache_count || 0 }}</span>
              </div>
            </el-col>
            <el-col :span="6">
              <div class="stat-box">
                <span class="stat-label">重发数量</span>
                <span class="stat-num">{{ stats.resend_count || 0 }}</span>
              </div>
            </el-col>
            <el-col :span="6">
              <div class="stat-box">
                <span class="stat-label">已连接通道</span>
                <span class="stat-num">{{ stats.connected_channels || 0 }}</span>
              </div>
            </el-col>
            <el-col :span="6">
              <div class="stat-box">
                <span class="stat-label">已打开串口</span>
                <span class="stat-num">{{ stats.opened_ports || 0 }}</span>
              </div>
            </el-col>
          </el-row>
        </el-card>
      </el-col>
    </el-row>

    <el-row :gutter="20" style="margin-top: 20px">
      <el-col :span="24">
        <el-card class="info-card">
          <template #header>
            <div class="card-header">
              <el-icon class="card-icon"><Location /></el-icon>
              <span>网络信息</span>
              <el-button @click="refreshNetwork" text size="small" style="margin-left: auto">刷新</el-button>
            </div>
          </template>
          <el-descriptions :column="2" border size="small">
            <el-descriptions-item label="IP地址">
              <span v-for="(iface, idx) in networkInfo" :key="idx">
                <span v-if="iface.ip_addresses && iface.ip_addresses.length">
                  <span v-for="(ip, i) in iface.ip_addresses" :key="i" class="ip-tag">{{ ip }}</span>
                  <span class="iface-name">{{ iface.name }}</span>
                </span>
              </span>
              <span v-if="!networkInfo.length || !networkInfo.some(n => n.ip_addresses && n.ip_addresses.length)">-</span>
            </el-descriptions-item>
            <el-descriptions-item label="接口状态">
              <el-tag
                v-for="(iface, idx) in networkInfo"
                :key="idx"
                :type="iface.status === 'up' ? 'success' : 'danger'"
                class="status-tag"
              >
                {{ iface.name }}: {{ iface.status === 'up' ? '已启用' : '已禁用' }}
              </el-tag>
            </el-descriptions-item>
            <el-descriptions-item label="MAC地址">
              <span v-for="(iface, idx) in networkInfo" :key="idx" class="ip-tag">{{ iface.mac_address || '-' }}</span>
            </el-descriptions-item>
            <el-descriptions-item label="网关地址">{{ gatewayInfo.gateway || '-' }}</el-descriptions-item>
          </el-descriptions>
        </el-card>
      </el-col>
    </el-row>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import api from '../utils/api'
import { CircleCheck, Cpu, DataBoard, Monitor, Location, DataLine } from '@element-plus/icons-vue'

const info = ref({})
const networkInfo = ref([])
const gatewayInfo = ref({})
const stats = ref({})

onMounted(async () => {
  await refreshInfo()
  await refreshNetwork()
  await refreshStats()
})

const refreshInfo = async () => {
  try {
    const response = await api.get('/system/info')
    info.value = response.data || {}
  } catch (error) {
    console.error('获取系统信息失败:', error)
  }
}

const refreshNetwork = async () => {
  try {
    const response = await api.get('/system/network')
    networkInfo.value = response.data || []
  } catch (error) {
    console.error('获取网络信息失败:', error)
  }
}

const refreshStats = async () => {
  try {
    const response = await api.get('/stats')
    stats.value = response.data || {}
  } catch (error) {
    console.error('获取统计信息失败:', error)
  }
}

const formatUptime = (seconds) => {
  if (seconds === undefined || seconds === null || isNaN(seconds)) {
    return '-'
  }
  const days = Math.floor(seconds / 86400)
  const hours = Math.floor((seconds % 86400) / 3600)
  const minutes = Math.floor((seconds % 3600) / 60)
  return `${days}天 ${hours}时 ${minutes}分`
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
.dashboard-home {
  padding: 0;
}

.status-card {
  border-radius: 12px;
  box-shadow: 0 2px 12px rgba(0, 0, 0, 0.06);
}

.status-header {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 12px;
}

.status-icon {
  font-size: 18px;
}

.status-online {
  color: #67c23a;
}

.status-cpu {
  color: #409eff;
}

.status-mem {
  color: #e6a23c;
}

.status-disk {
  color: #f56c6c;
}

.status-label {
  font-size: 14px;
  color: #909399;
}

.status-value {
  font-size: 28px;
  font-weight: 700;
  color: #303133;
  margin-bottom: 4px;
}

.status-desc {
  font-size: 12px;
  color: #909399;
}

.info-card {
  border-radius: 12px;
  box-shadow: 0 2px 12px rgba(0, 0, 0, 0.06);
}

.card-header {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 16px;
  font-weight: 600;
  color: #303133;
}

.card-icon {
  color: #409eff;
  font-size: 20px;
}

.ip-tag {
  display: inline-block;
  padding: 4px 12px;
  background: #ecf5ff;
  color: #409eff;
  border-radius: 4px;
  font-size: 13px;
  margin-right: 8px;
  margin-bottom: 4px;
}

.iface-name {
  font-size: 12px;
  color: #909399;
  margin-left: 8px;
}

.status-tag {
  margin-right: 8px;
  margin-bottom: 4px;
}

.stat-box {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 16px;
  background: #f5f7fa;
  border-radius: 8px;
}

.stat-label {
  font-size: 14px;
  color: #606266;
  margin-bottom: 8px;
}

.stat-num {
  font-size: 20px;
  font-weight: 600;
  color: #303133;
}
</style>
