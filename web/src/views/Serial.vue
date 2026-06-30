<template>
  <div class="serial">
    <el-card class="main-card">
      <template #header>
        <div class="card-header">
          <el-icon class="card-icon"><DataLine /></el-icon>
          <span>串口列表</span>
          <el-button type="primary" @click="showAddDialog = true" class="add-btn">
            <el-icon><Plus /></el-icon>
            添加串口
          </el-button>
        </div>
      </template>
      <el-table :data="serialPorts" border stripe>
        <el-table-column prop="id" label="ID" width="80" />
        <el-table-column prop="name" label="名称" width="120" />
        <el-table-column prop="port" label="端口" width="150" />
        <el-table-column prop="baud_rate" label="波特率" width="100" />
        <el-table-column prop="data_bits" label="数据位" width="80" />
        <el-table-column prop="parity" label="校验位" width="80">
          <template #default="{ row }">
            {{ parityLabels[row.parity] || row.parity }}
          </template>
        </el-table-column>
        <el-table-column prop="stop_bits" label="停止位" width="80" />
        <el-table-column prop="protocol" label="协议" width="120">
          <template #default="{ row }">
            <el-tag :type="row.protocol === 'modbus_rtu' ? 'warning' : 'info'">
              {{ protocolLabels[row.protocol] || row.protocol }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="enabled" label="状态" width="100">
          <template #default="{ row }">
            <el-switch :value="row.enabled" />
          </template>
        </el-table-column>
        <el-table-column label="操作" width="220">
          <template #default="{ row }">
            <el-button @click="openPort(row)" text type="success">打开</el-button>
            <el-button @click="closePort(row)" text>关闭</el-button>
            <el-button @click="editPort(row)" text type="primary">编辑</el-button>
            <el-button @click="deletePort(row)" text type="danger">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <el-dialog v-model="showAddDialog" :title="portForm.id ? '编辑串口' : '添加串口'" width="600px">
      <el-form :model="portForm" label-width="120px">
        <el-form-item label="名称">
          <el-input v-model="portForm.name" placeholder="请输入串口名称" />
        </el-form-item>
        <el-form-item label="端口路径">
          <el-input v-model="portForm.port" placeholder="例如: /dev/ttyS0 或 COM1" />
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
          <el-input-number v-model="portForm.delay_packaging" :min="0" :max="1000" />
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
import { DataLine, Plus } from '@element-plus/icons-vue'
import { ElMessage, ElMessageBox } from 'element-plus'

const serialPorts = ref([])
const showAddDialog = ref(false)

const parityLabels = {
  none: '无',
  odd: '奇',
  even: '偶'
}

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
  ElMessage.success('保存成功')
}

const editPort = (row) => {
  Object.assign(portForm, row)
  showAddDialog.value = true
}

const deletePort = async (row) => {
  ElMessageBox.confirm('确定删除此串口？', '提示', {
    confirmButtonText: '确定',
    cancelButtonText: '取消',
    type: 'warning'
  }).then(async () => {
    await api.delete(`/serial/ports/${row.id}`)
    await loadPorts()
    ElMessage.success('删除成功')
  }).catch(() => {})
}

const openPort = async (row) => {
  await api.post(`/serial/ports/${row.id}/open`)
  ElMessage.success('串口已打开')
}

const closePort = async (row) => {
  await api.post(`/serial/ports/${row.id}/close`)
  ElMessage.success('串口已关闭')
}
</script>

<style scoped>
.serial {
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

.add-btn {
  margin-left: auto;
}
</style>
