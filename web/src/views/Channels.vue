<template>
  <div class="channels">
    <el-card>
      <template #header>
        <span>通道列表</span>
        <el-button type="primary" @click="showAddDialog = true" style="float: right">添加通道</el-button>
      </template>
      <el-table :data="channels" border>
        <el-table-column prop="id" label="ID" />
        <el-table-column prop="name" label="名称" />
        <el-table-column prop="type" label="类型">
          <template #default="{ row }">
            {{ typeLabels[row.type] || row.type }}
          </template>
        </el-table-column>
        <el-table-column prop="serial_port" label="关联串口" />
        <el-table-column prop="enabled" label="状态">
          <template #default="{ row }">
            <el-switch :value="row.enabled" @change="toggleChannel(row)" />
          </template>
        </el-table-column>
        <el-table-column label="操作">
          <template #default="{ row }">
            <el-button @click="editChannel(row)" text>编辑</el-button>
            <el-button @click="deleteChannel(row)" text type="danger">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <el-dialog v-model="showAddDialog" title="添加通道">
      <el-form :model="channelForm" label-width="100px">
        <el-form-item label="名称">
          <el-input v-model="channelForm.name" />
        </el-form-item>
        <el-form-item label="类型">
          <el-select v-model="channelForm.type" @change="resetChannelForm">
            <el-option label="TCP客户端" value="tcp_client" />
            <el-option label="TCP服务端" value="tcp_server" />
            <el-option label="MQTT" value="mqtt" />
            <el-option label="HTTP" value="http" />
          </el-select>
        </el-form-item>
        <el-form-item label="关联串口">
          <el-select v-model="channelForm.serial_port">
            <el-option v-for="port in serialPorts" :key="port.id" :label="port.name" :value="port.id" />
          </el-select>
        </el-form-item>
        <template v-if="channelForm.type === 'tcp_client'">
          <el-form-item label="主机">
            <el-input v-model="channelForm.tcp_client.host" />
          </el-form-item>
          <el-form-item label="端口">
            <el-input-number v-model="channelForm.tcp_client.port" />
          </el-form-item>
        </template>
        <template v-if="channelForm.type === 'tcp_server'">
          <el-form-item label="端口">
            <el-input-number v-model="channelForm.tcp_server.port" />
          </el-form-item>
        </template>
        <template v-if="channelForm.type === 'mqtt'">
          <el-form-item label="Broker">
            <el-input v-model="channelForm.mqtt.broker" />
          </el-form-item>
          <el-form-item label="端口">
            <el-input-number v-model="channelForm.mqtt.port" />
          </el-form-item>
          <el-form-item label="用户名">
            <el-input v-model="channelForm.mqtt.username" />
          </el-form-item>
          <el-form-item label="密码">
            <el-input v-model="channelForm.mqtt.password" type="password" />
          </el-form-item>
          <el-form-item label="订阅Topic">
            <el-input v-model="channelForm.mqtt.subscribe_topic" />
          </el-form-item>
          <el-form-item label="发送Topic">
            <el-input v-model="channelForm.mqtt.send_topic" />
          </el-form-item>
        </template>
        <template v-if="channelForm.type === 'http'">
          <el-form-item label="URL">
            <el-input v-model="channelForm.http.url" />
          </el-form-item>
          <el-form-item label="方法">
            <el-select v-model="channelForm.http.method">
              <el-option label="GET" value="GET" />
              <el-option label="POST" value="POST" />
            </el-select>
          </el-form-item>
        </template>
        <el-form-item label="注册包">
          <el-input v-model="channelForm.register_packet" placeholder="十六进制" />
        </el-form-item>
        <el-form-item label="心跳包">
          <el-input v-model="channelForm.heartbeat_packet" placeholder="十六进制" />
        </el-form-item>
        <el-form-item label="心跳间隔(秒)">
          <el-input-number v-model="channelForm.heartbeat_interval" />
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

const channels = ref([])
const serialPorts = ref([])
const showAddDialog = ref(false)

const typeLabels = {
  tcp_client: 'TCP客户端',
  tcp_server: 'TCP服务端',
  mqtt: 'MQTT',
  http: 'HTTP'
}

const channelForm = reactive({
  id: '',
  name: '',
  type: 'tcp_client',
  serial_port: '',
  enabled: true,
  register_packet: '',
  heartbeat_packet: '',
  heartbeat_interval: 30,
  tcp_client: { host: '', port: 8080 },
  tcp_server: { port: 8080 },
  mqtt: { broker: '', port: 1883, username: '', password: '', subscribe_topic: '', send_topic: '' },
  http: { url: '', method: 'POST' }
})

onMounted(async () => {
  await loadChannels()
  await loadSerialPorts()
})

const loadChannels = async () => {
  const response = await api.get('/channels')
  channels.value = response.data
}

const loadSerialPorts = async () => {
  const response = await api.get('/serial/ports')
  serialPorts.value = response.data
}

const resetChannelForm = () => {
  channelForm.tcp_client = { host: '', port: 8080 }
  channelForm.tcp_server = { port: 8080 }
  channelForm.mqtt = { broker: '', port: 1883, username: '', password: '', subscribe_topic: '', send_topic: '' }
  channelForm.http = { url: '', method: 'POST' }
}

const saveChannel = async () => {
  if (channelForm.id) {
    await api.put(`/channels/${channelForm.id}`, channelForm)
  } else {
    await api.post('/channels', channelForm)
  }
  showAddDialog.value = false
  await loadChannels()
}

const editChannel = (row) => {
  Object.assign(channelForm, row)
  showAddDialog.value = true
}

const deleteChannel = async (row) => {
  if (confirm('确定删除此通道？')) {
    await api.delete(`/channels/${row.id}`)
    await loadChannels()
  }
}

const toggleChannel = async (row) => {
  if (row.enabled) {
    await api.post(`/channels/${row.id}/enable`)
  } else {
    await api.post(`/channels/${row.id}/disable`)
  }
}
</script>

<style scoped>
.channels {
  padding: 20px;
}
</style>
