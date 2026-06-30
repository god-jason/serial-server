<template>
  <div class="maintenance">
    <el-card>
      <template #header>
        <span>系统维护</span>
      </template>
      <el-row :gutter="20">
        <el-col :span="6">
          <el-button type="warning" @click="restart" style="width: 100%">重启系统</el-button>
        </el-col>
        <el-col :span="6">
          <el-button type="danger" @click="reset" style="width: 100%">系统还原</el-button>
        </el-col>
        <el-col :span="6">
          <el-button @click="showUpgradeDialog = true" style="width: 100%">远程升级</el-button>
        </el-col>
        <el-col :span="6">
          <el-button @click="showPasswordDialog = true" style="width: 100%">修改密码</el-button>
        </el-col>
      </el-row>
    </el-card>

    <el-card style="margin-top: 20px">
      <template #header>
        <span>系统日志</span>
      </template>
      <el-button @click="loadLogs" text>刷新日志</el-button>
      <pre class="log-content">{{ logs }}</pre>
    </el-card>

    <el-dialog v-model="showUpgradeDialog" title="远程升级">
      <el-upload
        :action="uploadUrl"
        :headers="uploadHeaders"
        :on-success="onUploadSuccess"
        :on-error="onUploadError"
        accept=".bin,.tar.gz"
        :show-file-list="false"
      >
        <el-button type="primary">选择升级文件</el-button>
      </el-upload>
      <template #footer>
        <el-button @click="showUpgradeDialog = false">取消</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="showPasswordDialog" title="修改密码">
      <el-form :model="passwordForm" label-width="100px">
        <el-form-item label="旧密码">
          <el-input v-model="passwordForm.old_password" type="password" />
        </el-form-item>
        <el-form-item label="新密码">
          <el-input v-model="passwordForm.new_password" type="password" />
        </el-form-item>
        <el-form-item label="确认新密码">
          <el-input v-model="passwordForm.confirm_password" type="password" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showPasswordDialog = false">取消</el-button>
        <el-button type="primary" @click="changePassword">确定</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive } from 'vue'
import api from '../utils/api'

const logs = ref('')
const showUpgradeDialog = ref(false)
const showPasswordDialog = ref(false)

const passwordForm = reactive({
  old_password: '',
  new_password: '',
  confirm_password: ''
})

const uploadUrl = '/api/system/upgrade'
const uploadHeaders = { Authorization: localStorage.getItem('token') }

const loadLogs = async () => {
  const response = await api.get('/logs')
  logs.value = typeof response.data === 'string' ? response.data : JSON.stringify(response.data, null, 2)
}

const restart = () => {
  if (confirm('确定重启系统？')) {
    api.post('/system/restart')
    alert('系统正在重启...')
  }
}

const reset = () => {
  if (confirm('确定恢复系统默认设置？此操作将清除所有配置！')) {
    api.post('/system/reset')
    alert('系统正在重置...')
  }
}

const onUploadSuccess = () => {
  alert('升级文件上传成功，系统正在升级...')
  showUpgradeDialog.value = false
}

const onUploadError = () => {
  alert('升级文件上传失败')
}

const changePassword = async () => {
  if (passwordForm.new_password !== passwordForm.confirm_password) {
    alert('两次输入的密码不一致')
    return
  }
  await api.put('/password', {
    old_password: passwordForm.old_password,
    new_password: passwordForm.new_password
  })
  alert('密码修改成功')
  showPasswordDialog.value = false
  passwordForm.old_password = ''
  passwordForm.new_password = ''
  passwordForm.confirm_password = ''
}
</script>

<style scoped>
.maintenance {
  padding: 20px;
}

.log-content {
  background: #f5f5f5;
  padding: 16px;
  border-radius: 4px;
  max-height: 400px;
  overflow: auto;
  white-space: pre-wrap;
}
</style>
