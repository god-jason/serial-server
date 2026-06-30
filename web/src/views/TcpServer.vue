<template>
  <div class="tcp-server">
    <el-card class="main-card">
      <template #header>
        <div class="card-header">
          <el-icon class="card-icon"><Connection /></el-icon>
          <span>TCP服务端通道</span>
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
        <el-table-column prop="tcp_server.port" label="监听端口" width="100">
          <template #default="{ row }">
            {{ row.tcp_server?.port || '-' }}
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

    <el-dialog v-model="showAddDialog" :title="isEdit ? '编辑通道' : '添加TCP服务端通道'" width="600px">
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
        <el-form-item label="监听端口">
          <el-input-number v-model="channelForm.tcp_server.port" :min="1" :max="65535" />
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
import { Connection, Plus } from '@element-plus/icons-vue'
import { ElMessage, ElMessageBox } from 'element-plus'

const channels = ref([])
const serialPorts = ref([])
const showAddDialog = ref(false)
const isEdit = ref(false)

const channelForm = reactive({
  id: '',
  name: '',
  type: 'tcp_server',
  serial_port: '',
  enabled: true,
  tcp_server: { port: 8080 }
})

onMounted(async () => {
  await loadChannels()
  await loadSerialPorts()
})

const loadChannels = async () => {
  try {
    const response = await api.get('/channels/tcp-server')
    channels.value = response.data || []
  } catch (error) {
    console.error('获取通道失败:', error)
  }
}

const generateTcpsId = () => {
  let n = 1
  while (channels.value.some(c => c.id === `tcps${n}`)) {
    n++
  }
  return `tcps${n}`
}

const resetForm = () => {
  isEdit.value = false
  channelForm.id = generateTcpsId()
  channelForm.name = ''
  channelForm.type = 'tcp_server'
  channelForm.serial_port = ''
  channelForm.enabled = true
  channelForm.tcp_server = { port: 8080 }
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
    if (isEdit.value) {
      await api.put(`/channels/tcp-server/${channelForm.id}`, channelForm)
    } else {
      await api.post('/channels/tcp-server', channelForm)
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
    type: 'tcp_server',
    serial_port: row.serial_port,
    enabled: row.enabled,
    tcp_server: {
      port: row.tcp_server?.port || 8080
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
      await api.delete(`/channels/tcp-server/${row.id}`)
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
.tcp-server {
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
  color: #67c23a;
  font-size: 20px;
}

.add-btn {
  margin-left: auto;
}
</style>
