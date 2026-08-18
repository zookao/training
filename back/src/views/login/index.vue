<template>
  <div class="login-container">
    <div class="bg-orbs">
      <span class="orb orb-1"></span>
      <span class="orb orb-2"></span>
      <span class="orb orb-3"></span>
    </div>
    <div class="login-card">
      <div class="login-header">
        <div class="logo-wrap">
          <img :src="siteStore.logoSrc || defaultLogo" alt="logo" class="logo-img" />
        </div>
        <div class="login-title">{{ siteStore.siteName }}</div>
        <div class="login-subtitle">Training Management System</div>
      </div>
      <el-form ref="formRef" :model="form" :rules="rules" size="large" @keyup.enter="handleLogin">
        <el-form-item prop="username">
          <el-input v-model="form.username" placeholder="请输入用户名" :prefix-icon="User" />
        </el-form-item>
        <el-form-item prop="password">
          <el-input v-model="form.password" type="password" placeholder="请输入密码" :prefix-icon="Lock" show-password />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" class="login-btn" :loading="loading" @click="handleLogin">登 录</el-button>
        </el-form-item>
      </el-form>
      <div class="login-footer">© 2026 培训管理系统 · All Rights Reserved</div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { User, Lock } from '@element-plus/icons-vue'
import { ElMessage, type FormInstance, type FormRules } from 'element-plus'
import { useUserStore } from '@/stores/user'
import { useSiteStore } from '@/stores/site'
import defaultLogo from '@/assets/logo.svg'

const router = useRouter()
const route = useRoute()
const userStore = useUserStore()
const siteStore = useSiteStore()
siteStore.ensureLoaded()

const formRef = ref<FormInstance>()
const loading = ref(false)
const form = reactive({ username: '', password: '' })
const rules: FormRules = {
  username: [{ required: true, message: '请输入用户名', trigger: 'blur' }],
  password: [{ required: true, message: '请输入密码', trigger: 'blur' }]
}

async function handleLogin() {
  if (!formRef.value) return
  await formRef.value.validate(async (valid) => {
    if (!valid) return
    loading.value = true
    try {
      await userStore.login(form)
      ElMessage.success('登录成功')
      const redirect = (route.query.redirect as string) || '/dashboard/index'
      router.push(redirect)
    } catch (e) {
      /* handled by interceptor */
    } finally {
      loading.value = false
    }
  })
}
</script>

<style scoped lang="scss">
.login-container {
  position: relative;
  height: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
  overflow: hidden;
  background: linear-gradient(135deg, #0f172a 0%, #1e293b 50%, #312e81 100%);
}

/* 背景光球装饰 */
.bg-orbs {
  position: absolute;
  inset: 0;
  pointer-events: none;
  .orb {
    position: absolute;
    border-radius: 50%;
    filter: blur(80px);
    opacity: 0.5;
    animation: float 12s ease-in-out infinite;
  }
  .orb-1 {
    width: 400px;
    height: 400px;
    background: #4f46e5;
    top: -100px;
    left: -100px;
  }
  .orb-2 {
    width: 350px;
    height: 350px;
    background: #0ea5e9;
    bottom: -80px;
    right: -80px;
    animation-delay: -4s;
  }
  .orb-3 {
    width: 300px;
    height: 300px;
    background: #8b5cf6;
    top: 50%;
    left: 50%;
    transform: translate(-50%, -50%);
    animation-delay: -8s;
  }
}
@keyframes float {
  0%, 100% { transform: translate(0, 0) scale(1); }
  33% { transform: translate(30px, -30px) scale(1.05); }
  66% { transform: translate(-20px, 20px) scale(0.95); }
}

/* 登录卡片 - 玻璃拟态 */
.login-card {
  position: relative;
  z-index: 1;
  width: 420px;
  padding: 40px 36px 28px;
  border-radius: 16px;
  background: rgba(255, 255, 255, 0.95);
  backdrop-filter: blur(20px);
  box-shadow:
    0 20px 60px rgba(0, 0, 0, 0.3),
    0 0 0 1px rgba(255, 255, 255, 0.1);
}

.login-header {
  text-align: center;
  margin-bottom: 32px;
  .logo-wrap {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    width: 64px;
    height: 64px;
    border-radius: 16px;
    overflow: hidden;
    margin-bottom: 16px;
    box-shadow: 0 8px 24px rgba(79, 70, 229, 0.4);
  }
  .logo-img {
    width: 100%;
    height: 100%;
    object-fit: cover;
  }
  .login-title {
    font-size: 24px;
    font-weight: 700;
    color: #1e293b;
    letter-spacing: 1px;
  }
  .login-subtitle {
    font-size: 12px;
    color: #94a3b8;
    margin-top: 6px;
    letter-spacing: 2px;
    text-transform: uppercase;
  }
}

.login-btn {
  width: 100%;
  height: 44px;
  font-size: 15px;
  letter-spacing: 8px;
  border: none;
  background: linear-gradient(135deg, #4f46e5 0%, #7c3aed 100%);
  box-shadow: 0 4px 16px rgba(79, 70, 229, 0.4);
  transition: all 0.3s;
  &:hover {
    transform: translateY(-1px);
    box-shadow: 0 6px 20px rgba(79, 70, 229, 0.5);
  }
}

.login-footer {
  text-align: center;
  color: #cbd5e1;
  font-size: 12px;
  margin-top: 24px;
}
</style>
