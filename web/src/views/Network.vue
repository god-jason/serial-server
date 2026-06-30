<template>
  <div class="network">
    <el-card class="main-card">
      <template #header>
        <div class="card-header">
          <el-icon class="card-icon"><Location /></el-icon>
          <span>网络配置</span>
        </div>
      </template>

      <el-row :gutter="20">
        <el-col :xs="24" :sm="24" :md="12" :lg="12">
          <el-card class="config-card">
            <template #header>
              <div class="card-header small">
                <el-icon class="card-icon small"><DataBoard /></el-icon>
                <span>IP地址配置</span>
              </div>
            </template>
            <el-form :model="config.gateway" label-width="120px" class="config-form">
              <el-form-item label="IP地址">
                <el-input v-model="config.gateway.ip" placeholder="例如: 192.168.1.100" />
              </el-form-item>
              <el-form-item label="子网掩码">
                <el-input v-model="config.gateway.netmask" placeholder="例如: 255.255.255.0" />
              </el-form-item>
              <el-form-item label="默认网关">
                <el-input v-model="config.gateway.gateway" placeholder="例如: 192.168.1.1" />
              </el-form-item>
              <el-form-item label="DNS服务器">
                <el-input v-model="config.gateway.dns" placeholder="例如: 8.8.8.8" />
              </el-form-item>
              <el-form-item label="启用DHCP">
                <el-switch v-model="config.gateway.dhcp" active-text="是" inactive-text="否" />
              </el-form-item>
              <el-form-item>
                <el-button type="primary" @click="saveGateway">
                  <el-icon><Check /></el-icon>
                  保存配置
                </el-button>
              </el-form-item>
            </el-form>
          </el-card>
        </el-col>

        <el-col :xs="24" :sm="24" :md="12" :lg="12">
          <el-card class="config-card">
            <template #header>
              <div class="card-header small">
                <el-icon class="card-icon small"><Monitor /></el-icon>
                <span>网络接口信息</span>
              </div>
            </template>
            <el-table :data="networkInfo" border stripe size="small">
              <el-table-column prop="name" label="接口名称" width="100" />
              <el-table-column prop="hardware_addr" label="MAC地址" width="140" />
              <el-table-column prop="ip_addresses" label="IP地址" min-width="150">
                <template #default="{ row }">
                  <span v-for="(ip, idx) in row.ip_addresses" :key="idx" class="ip-tag">{{ ip }}</span>
                </template>
              </el-table-column>
              <el-table-column prop="status" label="状态" width="80">
                <template #default="{ row }">
                  <el-tag :type="row.status === 'up' ? 'success' : 'danger'" size="small">
                    {{ row.status === 'up' ? '启用' : '禁用' }}
                  </el-tag>
                </template>
              </el-table-column>
            </el-table>
            <el-button @click="refreshNetwork" text size="small" style="margin-top: 12px">刷新</el-button>
          </el-card>
        </el-col>
      </el-row>

      <el-row :gutter="20" style="margin-top: 20px">
        <el-col :xs="24" :sm="24" :md="12" :lg="12">
          <el-card class="config-card">
            <template #header>
              <div class="card-header small">
                <el-icon class="card-icon small"><DataBoard /></el-icon>
                <span>WiFi配置</span>
              </div>
            </template>
            <el-form :model="wifi" label-width="100px" class="config-form">
              <el-form-item label="启用">
                <el-switch v-model="wifi.enabled" />
              </el-form-item>
              <el-form-item label="SSID">
                <el-input v-model="wifi.ssid" />
              </el-form-item>
              <el-form-item label="密码">
                <el-input v-model="wifi.password" type="password" />
              </el-form-item>
              <el-form-item>
                <el-button type="primary" @click="saveWiFi">
                  <el-icon><Check /></el-icon>
                  保存配置
                </el-button>
              </el-form-item>
            </el-form>
          </el-card>
        </el-col>

        <el-col :xs="24" :sm="24" :md="12" :lg="12">
          <el-card class="config-card">
            <template #header>
              <div class="card-header small">
                <el-icon class="card-icon small"><Monitor /></el-icon>
                <span>4G模组配置</span>
              </div>
            </template>
            <el-form :model="module4G" label-width="100px" class="config-form">
              <el-form-item label="启用">
                <el-switch v-model="module4G.enabled" />
              </el-form-item>
              <el-form-item label="APN">
                <el-input v-model="module4G.apn" />
              </el-form-item>
              <el-form-item>
                <el-button type="primary" @click="save4G">
                  <el-icon><Check /></el-icon>
                  保存配置
                </el-button>
              </el-form-item>
            </el-form>
          </el-card>
        </el-col>
      </el-row>
    </el-card>
  </div>
</template>

<script setup>
import { ref, onMounted, reactive } from 'vue'
import api from '../utils/api'
import { Location, DataBoard, Monitor, Check } from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus'

const config = reactive({
  gateway: {
    ip: '',
    netmask: '',
    gateway: '',
    dns: '',
    dhcp: true
  }
})

const wifi = reactive({
  enabled: false,
  ssid: '',
  password: ''
})

const module4G = reactive({
  enabled: false,
  apn: ''
})

const networkInfo = ref([])

onMounted(async () => {
  await loadConfig()
  await refreshNetwork()
})

const loadConfig = async () => {
  try {
    const response = await api.get('/config')
    const data = response.data || {}
    
    if (data.gateway) {
      Object.assign(config.gateway, data.gateway)
    }
    
    const system = data.system || {}
    if (system.wifi) {
      Object.assign(wifi, system.wifi)
    }
    if (system.module_4g) {
      Object.assign(module4G, system.module_4g)
    }
  } catch (error) {
    console.error('加载配置失败:', error)
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

const saveGateway = async () => {
  try {
    const response = await api.get('/config')
    const data = response.data || {}
    data.gateway = config.gateway
    await api.put('/config', data)
    ElMessage.success('网关配置已保存')
  } catch (error) {
    console.error('保存配置失败:', error)
    ElMessage.error('保存失败')
  }
}

const saveWiFi = async () => {
  try {
    const response = await api.get('/config')
    const data = response.data || {}
    if (!data.system) data.system = {}
    data.system.wifi = wifi
    await api.put('/config', data)
    ElMessage.success('WiFi配置已保存')
  } catch (error) {
    console.error('保存配置失败:', error)
    ElMessage.error('保存失败')
  }
}

const save4G = async () => {
  try {
    const response = await api.get('/config')
    const data = response.data || {}
    if (!data.system) data.system = {}
    data.system.module_4g = module4G
    await api.put('/config', data)
    ElMessage.success('4G配置已保存')
  } catch (error) {
    console.error('保存配置失败:', error)
    ElMessage.error('保存失败')
  }
}
</script>

<style scoped>
.network {
  width: 100%;
}

.main-card {
  border-radius: 12px;
  box-shadow: 0 2px 12px rgba(0, 0, 0, 0.06);
}

.config-card {
  border-radius: 10px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.04);
}

.card-header {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 16px;
  font-weight: 600;
  color: #303133;
}

.card-header.small {
  font-size: 14px;
  font-weight: 500;
}

.card-icon {
  color: #409eff;
  font-size: 20px;
}

.card-icon.small {
  font-size: 16px;
}

.config-form {
  padding: 8px 0;
}

.config-form .el-form-item {
  margin-bottom: 20px;
}

.config-form .el-form-item:last-child {
  margin-bottom: 0;
}

.ip-tag {
  display: inline-block;
  padding: 3px 10px;
  background: #ecf5ff;
  color: #409eff;
  border-radius: 4px;
  font-size: 12px;
  margin-right: 6px;
  margin-bottom: 3px;
}
</style>
