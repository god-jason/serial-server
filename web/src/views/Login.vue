<template>
  <div class="login-container">
    <div class="login-bg"></div>
    <div class="login-box">
      <div class="login-header">
        <div class="logo-circle">
          <el-icon :size="36"><DataLine /></el-icon>
        </div>
        <h2>串口服务器</h2>
        <p>Serial Server Management</p>
      </div>
      <el-form :model="form" :rules="rules" ref="formRef" label-width="0" class="login-form" @submit.prevent="login">
        <el-form-item prop="password">
          <el-input
            v-model="form.password"
            type="password"
            placeholder="请输入密码"
            @keyup.enter="login"
            :prefix-icon="Lock"
          />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="login" class="login-btn">
            登录
            <el-icon><ArrowRight /></el-icon>
          </el-button>
        </el-form-item>
      </el-form>
      <div class="login-footer">
        <span>版本: v1.0.0</span>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive } from 'vue'
import api from '../utils/api'
import { useRouter } from 'vue-router'
import { setLoggedIn } from '../stores/session'
import { DataLine, Lock, ArrowRight } from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus'

const router = useRouter()
const formRef = ref()
const form = reactive({
  password: ''
})

const rules = {
  password: [{ required: true, message: '请输入密码', trigger: 'blur' }]
}

const login = async () => {
  try {
    const response = await api.post('/login', { password: form.password })
    if (response.data.success) {
      setLoggedIn(true)
      router.push('/home')
    } else {
      ElMessage.error('密码错误')
    }
  } catch (error) {
    console.error('登录失败:', error)
    ElMessage.error('登录失败')
  }
}
</script>

<style scoped>
.login-container {
  display: flex;
  justify-content: center;
  align-items: center;
  height: 100vh;
  position: relative;
  overflow: hidden;
}

.login-bg {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: linear-gradient(135deg, #1a1a2e 0%, #16213e 50%, #0f3460 100%);
}

.login-bg::before {
  content: '';
  position: absolute;
  top: -50%;
  left: -50%;
  width: 200%;
  height: 200%;
  background: radial-gradient(circle, rgba(64, 158, 255, 0.1) 0%, transparent 50%);
  animation: rotate 20s linear infinite;
}

.login-bg::after {
  content: '';
  position: absolute;
  top: 20%;
  right: 10%;
  width: 300px;
  height: 300px;
  background: radial-gradient(circle, rgba(102, 126, 234, 0.15) 0%, transparent 70%);
  border-radius: 50%;
}

@keyframes rotate {
  from { transform: rotate(0deg); }
  to { transform: rotate(360deg); }
}

.login-box {
  position: relative;
  z-index: 10;
  background: rgba(255, 255, 255, 0.95);
  backdrop-filter: blur(10px);
  padding: 60px 50px;
  border-radius: 24px;
  box-shadow: 0 20px 60px rgba(0, 0, 0, 0.3);
  width: 460px;
  max-width: 90%;
}

.login-header {
  text-align: center;
  margin-bottom: 36px;
}

.logo-circle {
  width: 100px;
  height: 100px;
  background: linear-gradient(135deg, #409eff 0%, #667eea 100%);
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  margin: 0 auto 24px;
  color: white;
  box-shadow: 0 10px 30px rgba(64, 158, 255, 0.4);
}

.login-header h2 {
  margin: 0;
  font-size: 28px;
  font-weight: 700;
  color: #1a1a2e;
}

.login-header p {
  margin: 10px 0 0;
  font-size: 16px;
  color: #909399;
  letter-spacing: 2px;
}

.login-form {
  margin-bottom: 24px;
}

.login-form .el-input__wrapper {
  border-radius: 12px;
  height: 52px;
  transition: all 0.3s ease;
}

.login-form .el-input__inner {
  font-size: 16px;
  height: 52px;
  line-height: 52px;
}

.login-form .el-input__wrapper:hover {
  box-shadow: 0 0 0 2px rgba(64, 158, 255, 0.2);
}

.login-form .el-input__wrapper.is-focus {
  box-shadow: 0 0 0 2px rgba(64, 158, 255, 0.4);
  border-color: #409eff;
}

.login-btn {
  width: 100%;
  height: 48px;
  font-size: 16px;
  font-weight: 600;
  border-radius: 12px;
  background: linear-gradient(135deg, #409eff 0%, #667eea 100%);
  border: none;
  transition: all 0.3s ease;
}

.login-btn:hover {
  transform: translateY(-2px);
  box-shadow: 0 8px 20px rgba(64, 158, 255, 0.4);
}

.login-btn:active {
  transform: translateY(0);
}

.login-footer {
  text-align: center;
  padding-top: 20px;
  border-top: 1px solid #f0f0f0;
}

.login-footer span {
  font-size: 12px;
  color: #c0c4cc;
}
</style>
