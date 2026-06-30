<template>
  <div class="debug">
    <el-card>
      <template #header>
        <span>串口调试</span>
      </template>
      <el-row :gutter="20">
        <el-col :span="8">
          <el-form :model="debugForm" label-width="80px">
            <el-form-item label="选择串口">
              <el-select v-model="debugForm.port">
                <el-option v-for="port in serialPorts" :key="port.id" :label="port.name" :value="port.id" />
              </el-select>
            </el-form-item>
            <el-form-item label="显示格式">
              <el-radio-group v-model="debugForm.format">
                <el-radio label="hex">HEX</el-radio>
                <el-radio label="ascii">ASCII</el-radio>
              </el-radio-group>
            </el-form-item>
            <el-form-item>
              <el-button type="primary" @click="connect">连接</el-button>
              <el-button @click="disconnect">断开</el-button>
            </el-form-item>
          </el-form>
        </el-col>
        <el-col :span="16">
          <el-input
            v-model="sendData"
            type="textarea"
            :rows="4"
            placeholder="输入要发送的数据"
            style="margin-bottom: 10px"
          />
          <el-button type="primary" @click="send">发送</el-button>
        </el-col>
      </el-row>
      <el-row style="margin-top: 20px">
        <el-col :span="24">
          <el-input
            v-model="receivedData"
            type="textarea"
            :rows="10"
            readonly
            style="background: #f5f5f5"
          />
        </el-col>
      </el-row>
    </el-card>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted, onUnmounted } from 'vue'
import api from '../utils/api'
import { ElMessage } from 'element-plus'

const serialPorts = ref([])
const sendData = ref('')
const receivedData = ref('')
let ws = null

const debugForm = reactive({
  port: '',
  format: 'hex'
})

onMounted(async () => {
  const response = await api.get('/serial/ports')
  serialPorts.value = response.data
})

onUnmounted(() => {
  disconnect()
})

const connect = () => {
  if (!debugForm.port) {
    ElMessage.warning('请选择串口')
    return
  }
  const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
  ws = new WebSocket(`${protocol}//${window.location.host}/ws/serial/${debugForm.port}`)
  ws.onmessage = (event) => {
    const data = event.data
    if (debugForm.format === 'hex') {
      receivedData.value += bytesToHex(data) + '\n'
    } else {
      receivedData.value += data + '\n'
    }
  }
  ws.onclose = () => {
    ElMessage.info('连接已断开')
  }
}

const disconnect = () => {
  if (ws) {
    ws.close()
    ws = null
  }
}

const send = () => {
  if (!ws) {
    ElMessage.warning('请先连接')
    return
  }
  if (debugForm.format === 'hex') {
    const bytes = hexToBytes(sendData.value)
    ws.send(bytes)
  } else {
    ws.send(sendData.value)
  }
}

const hexToBytes = (hex) => {
  const bytes = []
  for (let i = 0; i < hex.length; i += 2) {
    bytes.push(parseInt(hex.substr(i, 2), 16))
  }
  return new Uint8Array(bytes)
}

const bytesToHex = (data) => {
  if (typeof data === 'string') {
    return data
  }
  const arr = new Uint8Array(data)
  let hex = ''
  for (let i = 0; i < arr.length; i++) {
    hex += arr[i].toString(16).padStart(2, '0') + ' '
  }
  return hex.trim()
}
</script>

<style scoped>
.debug {
  padding: 20px;
}
</style>
