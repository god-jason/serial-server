<template>
  <div class="logs">
    <el-card class="main-card">
      <template #header>
        <div class="card-header">
          <el-icon class="card-icon"><Document /></el-icon>
          <span>系统日志</span>
          <div class="header-actions">
            <el-select v-model="logLevel" style="width: 120px; margin-right: 12px">
              <el-option label="全部" value="all" />
              <el-option label="DEBUG" value="DEBUG" />
              <el-option label="INFO" value="INFO" />
              <el-option label="WARN" value="WARN" />
              <el-option label="ERROR" value="ERROR" />
            </el-select>
            <el-button @click="loadLogs" type="primary" size="small">
              <el-icon><Refresh /></el-icon>
              刷新
            </el-button>
            <el-button @click="clearLogs" type="danger" size="small">
              <el-icon><Delete /></el-icon>
              清空
            </el-button>
          </div>
        </div>
      </template>
      <div class="log-container">
        <pre class="log-content" ref="logContentRef">{{ filteredLogs }}</pre>
      </div>
    </el-card>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, nextTick, watch } from 'vue'
import api from '../utils/api'
import { Document, Refresh, Delete } from '@element-plus/icons-vue'
import { ElMessage, ElMessageBox } from 'element-plus'

const logs = ref('')
const logLevel = ref('all')
const logContentRef = ref(null)

const filteredLogs = computed(() => {
  if (logLevel.value === 'all' || !logs.value) {
    return logs.value
  }
  const lines = logs.value.split('\n')
  return lines.filter(line => line.includes(logLevel.value)).join('\n')
})

onMounted(async () => {
  await loadLogs()
})

const loadLogs = async () => {
  try {
    const response = await api.get('/logs')
    logs.value = typeof response.data === 'string' ? response.data : JSON.stringify(response.data, null, 2)
    await nextTick()
    scrollToBottom()
  } catch (error) {
    console.error('加载日志失败:', error)
    ElMessage.error('加载日志失败')
  }
}

const clearLogs = () => {
  ElMessageBox.confirm('确定清空系统日志？', '提示', {
    confirmButtonText: '确定',
    cancelButtonText: '取消',
    type: 'warning'
  }).then(async () => {
    try {
      await api.delete('/logs')
      logs.value = ''
      ElMessage.success('日志已清空')
    } catch (error) {
      console.error('清空日志失败:', error)
      ElMessage.error('清空日志失败')
    }
  }).catch(() => {})
}

const scrollToBottom = () => {
  if (logContentRef.value) {
    logContentRef.value.scrollTop = logContentRef.value.scrollHeight
  }
}

watch(logLevel, () => {
  nextTick(scrollToBottom)
})
</script>

<style scoped>
.logs {
  width: 100%;
}

.main-card {
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

.header-actions {
  margin-left: auto;
  display: flex;
  align-items: center;
}

.log-container {
  max-height: 500px;
  overflow: hidden;
  border-radius: 8px;
}

.log-content {
  background: #1a1a2e;
  color: #bfc9d4;
  padding: 16px;
  margin: 0;
  max-height: 500px;
  overflow-y: auto;
  white-space: pre-wrap;
  font-family: 'Consolas', 'Monaco', 'Courier New', monospace;
  font-size: 13px;
  line-height: 1.6;
}

.log-content::-webkit-scrollbar {
  width: 6px;
}

.log-content::-webkit-scrollbar-track {
  background: rgba(255, 255, 255, 0.05);
}

.log-content::-webkit-scrollbar-thumb {
  background: rgba(255, 255, 255, 0.2);
  border-radius: 3px;
}

.log-content::-webkit-scrollbar-thumb:hover {
  background: rgba(255, 255, 255, 0.3);
}
</style>
