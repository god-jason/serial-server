<template>
  <div class="serial">
    <el-card>
      <template #header>
        <span>串口列表</span>
        <el-button type="primary" @click="showAddDialog = true" style="float: right">添加串口</el-button>
      </template>
      <el-table :data="serialPorts" border>
        <el-table-column prop="id" label="ID" />
        <el-table-column prop="name" label="名称" />
        <el-table-column prop="port" label="端口" />
        <el-table-column prop="baud_rate" label="波特率" />
        <el-table-column prop="protocol" label="协议">
          <template #default="{ row }">
            {{ protocolLabels[row.protocol] || row.protocol }}
          </template>
        </el-table-column>
        <el-table-column prop="enabled" label="状态">
          <template #default="{ row }">
            <el-switch :value="row.enabled" />
          </template>
        </el-table-column>
        <el-table-column label="操作">
          <template #default="{ row }">
            <el-button @click="openPort(row)" text>打开</el-button>
            <el-button @click="closePort(row)" text>关闭</el-button>
            <el-button @click="editPort(row)" text>编辑</el-button>
            <el-button @click="deletePort(row)" text type="danger">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <el-dialog v-model="showAddDialog" title="添加串口">
      <el-form :model="portForm" label-width="100px">
        <el-form-item label="名称">
          <el-input v-model="portForm.name" />
        </el-form-item>
        <el-form-item label="端口">
          <el-input v-model="portForm.port" />
        </el-form-item>
        <el-form-item label="波特率">
          <el-select v-model="portForm.baud_rate">
            <el-option label="1200" :value="1200" />
            <el-option label="2400" :value="2400" />
            <el-option label="4800" :value="4800" />
            <el-option label="9600" :value="9600" />
            <el-option label="19200" :value="19200" />
            <el-option label="38400" :value="38400" />
            <el-option label="57600" :value="57600" />
            <el-option label="115200" :value="115200" />
          </el-select>
        </el-form-item>
        <el-form-item label="数据位">
          <el-select v-model="portForm.data_bits">
            <el-option label="5" :value="5" />
            <el-option label="6" :value="6" />
            <el-option label="7" :value="7" />
            <el-option label="8" :value="8" />
          </el-select>
        </el-form-item>
        <el-form-item label="校验位">
          <el-select v-model="portForm.parity">
            <el-option label="无" value="none" />
            <el-option label="奇校验" value="odd" />
            <el-option label="偶校验" value="even" />
          </el-select>
        </el-form-item>
        <el-form-item label="停止位">
          <el-select v-model="portForm.stop_bits">
            <el-option label="1" :value="1" />
            <el-option label="1.5" :value="1" />
            <el-option label="2" :value="2" />
          </el-select>
        </el-form-item>
        <el-form-item label="流控">
          <el-select v-model="portForm.flow_control">
            <el-option label="无" value="none" />
            <el-option label="RTS/CTS" value="rtscts" />
            <el-option label="XON/XOFF" value="xonxoff" />
          </el-select>
        </el-form-item>
        <el-form-item label="延迟封包(ms)">
          <el-input-number v-model="portForm.delay_packaging" />
        </el-form-item>
        <el-form-item label="协议">
          <el-select v-model="portForm.protocol">
            <el-option label="原始" value="raw" />
            <el-option label="Modbus RTU" value="modbus_rtu" />
          </el-select>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showAddDialog = false">取消</el-button>
        <el-button type="primary" @click="savePort">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, onMounted, reactive } from 'vue'
import api from '../utils/api'

const serialPorts = ref([])
const showAddDialog = ref(false)

const protocolLabels = {
  raw: '原始',
  modbus_rtu: 'Modbus RTU'
}

const portForm = reactive({
  id: '',
  name: '',
  port: '',
  baud_rate: 9600,
  data_bits: 8,
  parity: 'none',
  stop_bits: 1,
  flow_control: 'none',
  delay_packaging: 10,
  delay_timeout: 100,
  protocol: 'raw',
  enabled: false
})

onMounted(async () => {
  await loadPorts()
})

const loadPorts = async () => {
  const response = await api.get('/serial/ports')
  serialPorts.value = response.data
}

const savePort = async () => {
  if (portForm.id) {
    await api.put(`/serial/ports/${portForm.id}`, portForm)
  } else {
    await api.post('/serial/ports', portForm)
  }
  showAddDialog.value = false
  await loadPorts()
}

const editPort = (row) => {
  Object.assign(portForm, row)
  showAddDialog.value = true
}

const deletePort = async (row) => {
  if (confirm('确定删除此串口？')) {
    await api.delete(`/serial/ports/${row.id}`)
    await loadPorts()
  }
}

const openPort = async (row) => {
  await api.post(`/serial/ports/${row.id}/open`)
  alert('串口已打开')
}

const closePort = async (row) => {
  await api.post(`/serial/ports/${row.id}/close`)
  alert('串口已关闭')
}
</script>

<style scoped>
.serial {
  padding: 20px;
}
</style>
