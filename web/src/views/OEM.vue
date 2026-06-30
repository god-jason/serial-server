<template>
  <div class="oem">
    <el-card>
      <template #header>
        <span>OEM配置</span>
      </template>
      <el-form :model="oemConfig" label-width="100px">
        <el-form-item label="名称">
          <el-input v-model="oemConfig.name" />
        </el-form-item>
        <el-form-item label="Logo URL">
          <el-input v-model="oemConfig.logo" />
        </el-form-item>
        <el-form-item label="主题颜色">
          <el-color-picker v-model="oemConfig.theme_color" />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="save">保存配置</el-button>
        </el-form-item>
      </el-form>
    </el-card>

    <el-card style="margin-top: 20px">
      <template #header>
        <span>4G模组配置</span>
      </template>
      <el-form :model="module4G" label-width="100px">
        <el-form-item label="启用">
          <el-switch v-model="module4G.enabled" />
        </el-form-item>
        <el-form-item label="APN">
          <el-input v-model="module4G.apn" />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="save">保存配置</el-button>
        </el-form-item>
      </el-form>
    </el-card>

    <el-card style="margin-top: 20px">
      <template #header>
        <span>WiFi配置</span>
      </template>
      <el-form :model="wifi" label-width="100px">
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
          <el-button type="primary" @click="save">保存配置</el-button>
        </el-form-item>
      </el-form>
    </el-card>
  </div>
</template>

<script setup>
import { ref, onMounted, reactive } from 'vue'
import api from '../utils/api'

const oemConfig = reactive({
  name: '',
  logo: '',
  theme_color: '#409EFF'
})

const module4G = reactive({
  enabled: false,
  apn: ''
})

const wifi = reactive({
  enabled: false,
  ssid: '',
  password: ''
})

onMounted(async () => {
  const response = await api.get('/config')
  Object.assign(oemConfig, response.data.system.oem)
  Object.assign(module4G, response.data.system.module_4g)
  Object.assign(wifi, response.data.system.wifi)
})

const save = async () => {
  const config = await api.get('/config')
  config.data.system.oem = oemConfig
  config.data.system.module_4g = module4G
  config.data.system.wifi = wifi
  await api.put('/config', config.data)
  alert('配置已保存')
}
</script>

<style scoped>
.oem {
  padding: 20px;
}
</style>
