<template>
  <div class="login-container">
    <div class="brand-side">
      <div class="brand-overlay"></div>
      <div class="brand-inner">
        <img :src="siteStore.logoSrc || defaultLogo" class="brand-logo" alt="logo" />
        <h1>{{ siteStore.siteName }}</h1>
        <p>随时随地，开启你的在线学习之旅</p>
        <div class="brand-features">
          <div class="feature-item">
            <el-icon :size="20"><Reading /></el-icon>
            <span>海量课程</span>
          </div>
          <div class="feature-item">
            <el-icon :size="20"><VideoCamera /></el-icon>
            <span>高清视频</span>
          </div>
        </div>
      </div>
    </div>
    <el-card class="login-card">
      <div class="title">欢迎回来</div>
      <p class="subtitle">学员登录</p>
      <el-form ref="formRef" :model="form" :rules="rules" @keyup.enter="submit">
        <el-form-item prop="username">
          <el-input v-model="form.username" placeholder="账号 / 手机号" :prefix-icon="User" />
        </el-form-item>
        <el-form-item prop="password">
          <el-input v-model="form.password" type="password" placeholder="密码" :prefix-icon="Lock" show-password />
        </el-form-item>
        <el-button type="primary" class="login-btn" :loading="loading" @click="submit">登 录</el-button>
      </el-form>
      <p class="tip">
        还没有账号？<router-link to="/register" class="link">立即注册</router-link>
      </p>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive } from 'vue'
import { useRouter } from 'vue-router'
import { User, Lock, Reading, VideoCamera } from '@element-plus/icons-vue'
import { ElMessage, type FormInstance, type FormRules } from 'element-plus'
import { useUserStore } from '@/stores/user'
import { useSiteStore } from '@/stores/site'
import defaultLogo from '@/assets/logo.svg'

const router = useRouter()
const userStore = useUserStore()
const siteStore = useSiteStore()
siteStore.ensureLoaded()
const formRef = ref<FormInstance>()
const loading = ref(false)
const form = reactive({ username: '', password: '' })
const rules: FormRules = {
  username: [{ required: true, message: '请输入账号或手机号', trigger: 'blur' }],
  password: [{ required: true, message: '请输入密码', trigger: 'blur' }]
}

async function submit() {
  if (!formRef.value) return
  await formRef.value.validate(async (valid) => {
    if (!valid) return
    loading.value = true
    try {
      await userStore.login(form)
      await userStore.fetchUserInfo()
      ElMessage.success('登录成功')
      router.push('/classes')
    } catch (e) {
      /* error handled in interceptor */
    } finally {
      loading.value = false
    }
  })
}
</script>

<style scoped lang="scss">
.login-container {
  height: 100%;
  display: flex;
  align-items: stretch;
  justify-content: center;
  background: #f5f7fa;
}

/* 左侧品牌区 - 背景图 + 渐变遮罩 */
.brand-side {
  position: relative;
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  overflow: hidden;
  background-image: url('');
  background-size: cover;
  background-position: center;
}

.brand-overlay {
  position: absolute;
  inset: 0;
  background: linear-gradient(135deg, rgba(30, 41, 98, 0.82) 0%, rgba(76, 29, 149, 0.78) 100%);
}

.brand-inner {
  position: relative;
  z-index: 1;
  text-align: center;
  color: #fff;
  padding: 40px;
  .brand-logo {
    width: 72px;
    height: 72px;
    border-radius: 14px;
    box-shadow: 0 8px 24px rgba(0, 0, 0, 0.3);
  }
  h1 {
    font-size: 34px;
    margin: 18px 0 8px;
    letter-spacing: 3px;
    font-weight: 700;
    text-shadow: 0 2px 12px rgba(0, 0, 0, 0.3);
  }
  p {
    font-size: 15px;
    opacity: 0.9;
    letter-spacing: 1px;
  }
}

/* 特性标签 */
.brand-features {
  display: flex;
  gap: 28px;
  justify-content: center;
  margin-top: 40px;
  .feature-item {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 8px;
    font-size: 13px;
    opacity: 0.9;
    .el-icon {
      padding: 10px;
      border-radius: 12px;
      background: rgba(255, 255, 255, 0.15);
      backdrop-filter: blur(10px);
    }
  }
}

/* 右侧登录卡片 */
.login-card {
  width: 420px;
  padding: 48px 36px;
  border-radius: 0;
  text-align: center;
  box-shadow: -8px 0 30px rgba(0, 0, 0, 0.1);
  border: none;
}
.title {
  font-size: 26px;
  font-weight: 700;
  margin-bottom: 6px;
  color: #1e293b;
}
.subtitle {
  color: #94a3b8;
  margin-bottom: 32px;
  font-size: 13px;
  letter-spacing: 2px;
}
.login-btn {
  width: 100%;
  height: 44px;
  font-size: 15px;
  letter-spacing: 8px;
  border: none;
  background: linear-gradient(135deg, #4f46e5 0%, #7c3aed 100%);
  box-shadow: 0 4px 16px rgba(79, 70, 229, 0.3);
  transition: all 0.3s;
  &:hover {
    transform: translateY(-1px);
    box-shadow: 0 6px 20px rgba(79, 70, 229, 0.4);
  }
}
.tip {
  margin-top: 20px;
  color: #94a3b8;
  font-size: 13px;
  .link {
    color: #4f46e5;
    text-decoration: none;
    font-weight: 500;
    &:hover {
      text-decoration: underline;
    }
  }
}
@media (max-width: 768px) {
  .brand-side {
    display: none;
  }
  .login-card {
    width: 100%;
  }
}
</style>
