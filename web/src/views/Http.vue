<template>
  <div class="http">
    <el-card class="main-card">
      <template #header>
        <div class="card-header">
          <el-icon class="card-icon"><Link /></el-icon>
          <span>HTTP通道</span>
          <el-button type="primary" @click="resetForm(); showAddDialog = true" class="add-btn">
            <el-icon><Plus /></el-icon>
            添加通道
          </el-button>
        </div>
      </template>
      <el-table :data="channels" border stripe>
        <el-table-column prop="id" label="ID" width="80" />
        <el-table-column prop="name" label="名称" width="120" />
        <el-table-column prop="serial_port" label="关联串口" width="120">
          <template #default="{ row }">
            {{ getSerialName(row.serial_port) }}
          </template>
        </el-table-column>
        <el-table-column prop="http.url" label="请求URL" width="250">
          <template #default="{ row }">
            {{ row.http?.url || '-' }}
          </template>
        </el-table-column>
        <el-table-column prop="http.method" label="请求方法" width="100">
          <template #default="{ row }">
            <el-tag :type="row.http?.method === 'POST' ? 'danger' : 'info'">
              {{ row.http?.method || '-' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="enabled" label="状态" width="100">
          <template #default="{ row }">
            <el-switch :value="row.enabled" @change="toggleChannel(row)" />
          </template>
        </el-table-column>
        <el-table-column label="操作" width="180">
          <template #default="{ row }">
            <el-button @click="editChannel(row)" text type="primary">编辑</el-button>
            <el-button @click="deleteChannel(row)" text type="danger">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <el-dialog v-model="showAddDialog" :title="isEdit ? '编辑通道' : '添加HTTP通道'" width="600px">
      <el-form :model="channelForm" label-width="120px">
        <el-form-item label="ID">
          <el-input v-model="channelForm.id" :disabled="isEdit" />
        </el-form-item>
        <el-form-item label="名称">
          <el-input v-model="channelForm.name" placeholder="请输入通道名称" />
        </el-form-item>
        <el-form-item label="关联串口">
          <el-select v-model="channelForm.serial_port">
            <el-option v-for="port in serialPorts" :key="port.id" :label="port.name" :value="port.id" />
          </el-select>
        </el-form-item>
        <el-form-item label="请求URL">
          <el-input v-model="channelForm.http.url" placeholder="例如: http://api.example.com/data" />
        </el-form-item>
        <el-form-item label="请求方法">
          <el-select v-model="channelForm.http.method">
            <el-option label="GET" value="GET" />
            <el-option label="POST" value="POST" />
            <el-option label="PUT" value="PUT" />
            <el-option label="DELETE" value="DELETE" />
          </el-select>
        </el-form-item>
        <el-form-item label="认证Token">
          <el-input v-model="channelForm.http.token" placeholder="可选，用于Authorization头" />
        </el-form-item>
        <el-form-item label="Content-Type">
          <el-input v-model="channelForm.http.content_type" placeholder="默认: application/json" />
        </el-form-item>
        <el-form-item label="注册包">
          <div class="packet-input-group">
            <el-input v-model="channelForm.register_packet" :placeholder="registerPacketMode === 'hex' ? '十六进制格式，如：010203' : 'ASCII格式'" />
            <el-button-group class="mode-switch">
              <el-button :type="registerPacketMode === 'hex' ? 'primary' : ''" @click="registerPacketMode = 'hex'">HEX</el-button>
              <el-button :type="registerPacketMode === 'ascii' ? 'primary' : ''" @click="registerPacketMode = 'ascii'">ASCII</el-button>
            </el-button-group>
          </div>
        </el-form-item>
        <el-form-item label="心跳包">
          <div class="packet-input-group">
            <el-input v-model="channelForm.heartbeat_packet" :placeholder="heartbeatPacketMode === 'hex' ? '十六进制格式，如：010203' : 'ASCII格式'" />
            <el-button-group class="mode-switch">
              <el-button :type="heartbeatPacketMode === 'hex' ? 'primary' : ''" @click="heartbeatPacketMode = 'hex'">HEX</el-button>
              <el-button :type="heartbeatPacketMode === 'ascii' ? 'primary' : ''" @click="heartbeatPacketMode = 'ascii'">ASCII</el-button>
            </el-button-group>
          </div>
        </el-form-item>
        <el-form-item label="心跳间隔(秒)">
          <el-input-number v-model="channelForm.heartbeat_interval" :min="1" :max="3600" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showAddDialog = false">取消</el-button>
        <el-button type="primary" @click="saveChannel">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, onMounted, reactive } from 'vue'
import api from '../utils/api'
import { Link, Plus } from '@element-plus/icons-vue'
import { ElMessage, ElMessageBox } from 'element-plus'

const channels = ref([])
const serialPorts = ref([])
const showAddDialog = ref(false)
const isEdit = ref(false)
const registerPacketMode = ref('hex')
const heartbeatPacketMode = ref('hex')

const channelForm = reactive({
  id: '',
  name: '',
  type: 'http',
  serial_port: '',
  enabled: true,
  register_packet: '',
  heartbeat_packet: '',
  heartbeat_interval: 30,
  http: { url: '', method: 'POST', token: '', content_type: 'application/json' }
})

onMounted(async () => {
  await loadChannels()
  await loadSerialPorts()
})

const loadChannels = async () => {
  try {
    const response = await api.get('/channels/http')
    channels.value = response.data || []
  } catch (error) {
    console.error('获取通道失败:', error)
  }
}

const generateHttpId = () => {
  let n = 1
  while (channels.value.some(c => c.id === `http${n}`)) {
    n++
  }
  return `http${n}`
}

const asciiToHex = (str) => {
  let hex = ''
  for (let i = 0; i < str.length; i++) {
    hex += str.charCodeAt(i).toString(16).padStart(2, '0').toUpperCase()
  }
  return hex
}

const prepareChannelData = () => {
  const data = { ...channelForm }
  if (registerPacketMode.value === 'ascii' && data.register_packet) {
    data.register_packet = asciiToHex(data.register_packet)
  }
  if (heartbeatPacketMode.value === 'ascii' && data.heartbeat_packet) {
    data.heartbeat_packet = asciiToHex(data.heartbeat_packet)
  }
  return data
}

const resetForm = () => {
  isEdit.value = false
  channelForm.id = generateHttpId()
  channelForm.name = ''
  channelForm.type = 'http'
  channelForm.serial_port = ''
  channelForm.enabled = true
  channelForm.register_packet = ''
  channelForm.heartbeat_packet = ''
  channelForm.heartbeat_interval = 30
  channelForm.http = { url: '', method: 'POST', token: '', content_type: 'application/json' }
}

const loadSerialPorts = async () => {
  try {
    const response = await api.get('/serial/ports')
    serialPorts.value = response.data || []
  } catch (error) {
    console.error('获取串口列表失败:', error)
  }
}

const getSerialName = (id) => {
  const port = serialPorts.value.find(p => p.id === id)
  return port ? port.name : id
}

const saveChannel = async () => {
  try {
    const data = prepareChannelData()
    if (isEdit.value) {
      await api.put(`/channels/http/${channelForm.id}`, data)
    } else {
      await api.post('/channels/http', data)
    }
    showAddDialog.value = false
    await loadChannels()
    ElMessage.success('保存成功')
  } catch (error) {
    console.error('保存失败:', error)
    ElMessage.error('保存失败')
  }
}

const editChannel = (row) => {
  isEdit.value = true
  Object.assign(channelForm, {
    id: row.id,
    name: row.name,
    type: 'http',
    serial_port: row.serial_port,
    enabled: row.enabled,
    register_packet: row.register_packet || '',
    heartbeat_packet: row.heartbeat_packet || '',
    heartbeat_interval: row.heartbeat_interval || 30,
    http: {
      url: row.http?.url || '',
      method: row.http?.method || 'POST',
      token: row.http?.token || '',
      content_type: row.http?.content_type || 'application/json'
    }
  })
  showAddDialog.value = true
}

const deleteChannel = async (row) => {
  ElMessageBox.confirm('确定删除此通道？', '提示', {
    confirmButtonText: '确定',
    cancelButtonText: '取消',
    type: 'warning'
  }).then(async () => {
    try {
      await api.delete(`/channels/http/${row.id}`)
      await loadChannels()
      ElMessage.success('删除成功')
    } catch (error) {
      ElMessage.error('删除失败')
    }
  }).catch(() => {})
}

const toggleChannel = async (row) => {
  try {
    if (row.enabled) {
      await api.post(`/channels/${row.id}/enable`)
    } else {
      await api.post(`/channels/${row.id}/disable`)
    }
  } catch (error) {
    console.error('切换状态失败:', error)
    row.enabled = !row.enabled
  }
}
</script>

<style scoped>
.http {
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
  color: #909399;
  font-size: 20px;
}

.add-btn {
  margin-left: auto;
}

.packet-input-group {
  display: flex;
  gap: 10px;
  align-items: center;
}

.packet-input-group .el-input {
  flex: 1;
}

.mode-switch {
  flex-shrink: 0;
}
</style>
