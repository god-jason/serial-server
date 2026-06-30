<template>
  <div class="maintenance">
    <el-card class="main-card">
      <template #header>
        <div class="card-header">
          <el-icon class="card-icon"><Setting /></el-icon>
          <span>系统维护</span>
        </div>
      </template>
      <el-row :gutter="20">
        <el-col :xs="24" :sm="12" :md="12" :lg="6">
          <el-card class="action-card" shadow="hover">
            <div class="action-icon restart">
              <el-icon :size="32"><RefreshLeft /></el-icon>
            </div>
            <h3 class="action-title">重启系统</h3>
            <p class="action-desc">重启整个系统服务</p>
            <el-button type="warning" @click="restart" style="width: 100%">执行重启</el-button>
          </el-card>
        </el-col>
        <el-col :xs="24" :sm="12" :md="12" :lg="6">
          <el-card class="action-card" shadow="hover">
            <div class="action-icon reset">
              <el-icon :size="32"><RefreshLeft /></el-icon>
            </div>
            <h3 class="action-title">系统还原</h3>
            <p class="action-desc">恢复系统默认设置</p>
            <el-button type="danger" @click="reset" style="width: 100%">执行还原</el-button>
          </el-card>
        </el-col>
        <el-col :xs="24" :sm="12" :md="12" :lg="6">
          <el-card class="action-card" shadow="hover">
            <div class="action-icon upgrade">
              <el-icon :size="32"><Upload /></el-icon>
            </div>
            <h3 class="action-title">远程升级</h3>
            <p class="action-desc">上传固件文件进行升级</p>
            <el-button @click="showUpgradeDialog = true" style="width: 100%">选择文件</el-button>
          </el-card>
        </el-col>
        <el-col :xs="24" :sm="12" :md="12" :lg="6">
          <el-card class="action-card" shadow="hover">
            <div class="action-icon password">
              <el-icon :size="32"><Lock /></el-icon>
            </div>
            <h3 class="action-title">修改密码</h3>
            <p class="action-desc">修改登录密码</p>
            <el-button @click="showPasswordDialog = true" style="width: 100%">修改密码</el-button>
          </el-card>
        </el-col>
      </el-row>
    </el-card>

    <el-dialog v-model="showUpgradeDialog" title="远程升级" width="500px">
      <el-upload
        :action="uploadUrl"
        :headers="uploadHeaders"
        :on-success="onUploadSuccess"
        :on-error="onUploadError"
        accept=".bin,.tar.gz"
        :show-file-list="false"
        drag
      >
        <el-icon class="upload-icon"><Upload /></el-icon>
        <div class="upload-text">点击或拖拽文件到此处上传</div>
        <div class="upload-tip">支持 .bin 和 .tar.gz 格式</div>
      </el-upload>
      <template #footer>
        <el-button @click="showUpgradeDialog = false">取消</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="showPasswordDialog" title="修改密码" width="450px">
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
import { Setting, RefreshLeft, Upload, Lock } from '@element-plus/icons-vue'
import { ElMessage, ElMessageBox } from 'element-plus'

const showUpgradeDialog = ref(false)
const showPasswordDialog = ref(false)

const passwordForm = reactive({
  old_password: '',
  new_password: '',
  confirm_password: ''
})

const uploadUrl = '/api/system/upgrade'
const uploadHeaders = { Authorization: localStorage.getItem('token') }

const restart = () => {
  ElMessageBox.confirm('确定重启系统？', '提示', {
    confirmButtonText: '确定',
    cancelButtonText: '取消',
    type: 'warning'
  }).then(async () => {
    try {
      await api.post('/system/restart')
      ElMessage.info('系统正在重启...')
    } catch (error) {
      ElMessage.error('重启失败')
    }
  }).catch(() => {})
}

const reset = () => {
  ElMessageBox.confirm('确定恢复系统默认设置？此操作将清除所有配置！', '警告', {
    confirmButtonText: '确定',
    cancelButtonText: '取消',
    type: 'error'
  }).then(async () => {
    try {
      await api.post('/system/reset')
      ElMessage.info('系统正在重置...')
    } catch (error) {
      ElMessage.error('重置失败')
    }
  }).catch(() => {})
}

const onUploadSuccess = () => {
  ElMessage.info('升级文件上传成功，系统正在升级...')
  showUpgradeDialog.value = false
}

const onUploadError = () => {
  ElMessage.error('升级文件上传失败')
}

const changePassword = async () => {
  if (passwordForm.new_password !== passwordForm.confirm_password) {
    ElMessage.error('两次输入的密码不一致')
    return
  }
  try {
    await api.put('/password', {
      old_password: passwordForm.old_password,
      new_password: passwordForm.new_password
    })
    ElMessage.success('密码修改成功')
    showPasswordDialog.value = false
    passwordForm.old_password = ''
    passwordForm.new_password = ''
    passwordForm.confirm_password = ''
  } catch (error) {
    ElMessage.error('密码修改失败')
  }
}
</script>

<style scoped>
.maintenance {
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

.action-card {
  border-radius: 12px;
  text-align: center;
  padding: 24px;
  transition: transform 0.2s, box-shadow 0.2s;
}

.action-card:hover {
  transform: translateY(-4px);
  box-shadow: 0 8px 24px rgba(0, 0, 0, 0.12);
}

.action-icon {
  width: 64px;
  height: 64px;
  border-radius: 16px;
  display: flex;
  align-items: center;
  justify-content: center;
  margin: 0 auto 16px;
  color: white;
}

.action-icon.restart {
  background: linear-gradient(135deg, #e6a23c 0%, #d4912f 100%);
}

.action-icon.reset {
  background: linear-gradient(135deg, #f56c6c 0%, #e64949 100%);
}

.action-icon.upgrade {
  background: linear-gradient(135deg, #67c23a 0%, #5eb833 100%);
}

.action-icon.password {
  background: linear-gradient(135deg, #409eff 0%, #3688ff 100%);
}

.action-title {
  font-size: 16px;
  font-weight: 600;
  color: #303133;
  margin: 0 0 8px;
}

.action-desc {
  font-size: 13px;
  color: #909399;
  margin: 0 0 16px;
}

.upload-icon {
  font-size: 48px;
  color: #409eff;
}

.upload-text {
  margin: 16px 0 8px;
  font-size: 14px;
  color: #606266;
}

.upload-tip {
  font-size: 12px;
  color: #c0c4cc;
}
</style>
