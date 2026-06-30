<template>
  <div class="gateway">
    <el-card>
      <el-form :model="config" label-width="100px">
        <el-form-item label="IP地址">
          <el-input v-model="config.gateway.ip" />
        </el-form-item>
        <el-form-item label="子网掩码">
          <el-input v-model="config.gateway.netmask" />
        </el-form-item>
        <el-form-item label="网关">
          <el-input v-model="config.gateway.gateway" />
        </el-form-item>
        <el-form-item label="DNS">
          <el-input v-model="config.gateway.dns" />
        </el-form-item>
        <el-form-item label="DHCP">
          <el-switch v-model="config.gateway.dhcp" />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="save">保存配置</el-button>
        </el-form-item>
      </el-form>
    </el-card>
    <el-card style="margin-top: 20px">
      <template #header>
        <span>网络接口</span>
      </template>
      <el-table :data="networkInfo" border>
        <el-table-column prop="name" label="名称" />
        <el-table-column prop="hardware_addr" label="MAC地址" />
        <el-table-column prop="ip_addresses" label="IP地址" />
        <el-table-column prop="mtu" label="MTU" />
      </el-table>
    </el-card>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import api from '../utils/api'

const config = ref({
  gateway: {
    ip: '',
    netmask: '',
    gateway: '',
    dns: '',
    dhcp: true
  }
})

const networkInfo = ref([])

onMounted(async () => {
  const response = await api.get('/config')
  config.value = response.data
  const netResponse = await api.get('/system/network')
  networkInfo.value = netResponse.data
})

const save = async () => {
  await api.put('/config', config.value)
  alert('配置已保存')
}
</script>

<style scoped>
.gateway {
  padding: 20px;
}
</style>
