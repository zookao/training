<template>
  <div class="register-container">
    <div class="brand-side">
      <div class="brand-overlay"></div>
      <div class="brand-inner">
        <img :src="siteStore.logoSrc || defaultLogo" class="brand-logo" alt="logo" />
        <h1>{{ siteStore.siteName }}</h1>
        <p>注册账号，开启你的学习之旅</p>
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
    <el-card class="register-card">
      <div class="title">学员注册</div>
      <p class="subtitle">创建你的账号</p>
      <el-form ref="formRef" :model="form" :rules="rules" @keyup.enter="submit">
        <el-form-item prop="username">
          <el-input v-model="form.username" placeholder="账号（3-50位）" :prefix-icon="User" />
        </el-form-item>
        <el-form-item prop="phone">
          <el-input v-model="form.phone" placeholder="手机号（将作为登录凭证）" :prefix-icon="Iphone" maxlength="11" />
        </el-form-item>
        <el-form-item prop="password">
          <el-input v-model="form.password" type="password" placeholder="密码（至少6位）" :prefix-icon="Lock" show-password />
        </el-form-item>
        <el-form-item prop="confirmPassword">
          <el-input v-model="form.confirmPassword" type="password" placeholder="确认密码" :prefix-icon="Lock" show-password />
        </el-form-item>
        <el-form-item prop="nickname">
          <el-input v-model="form.nickname" placeholder="姓名（选填）" :prefix-icon="UserFilled" />
        </el-form-item>
        <el-form-item>
          <el-input v-model="form.studentNo" placeholder="学号（选填）" :prefix-icon="Postcard" />
        </el-form-item>
        <el-form-item>
          <el-select v-model="form.departmentId" clearable placeholder="院系（选填）" style="width:100%">
            <template #prefix>
              <el-icon><OfficeBuilding /></el-icon>
            </template>
            <el-option v-for="d in departments" :key="d.id" :label="d.name" :value="d.id" />
          </el-select>
        </el-form-item>
        <el-button type="primary" class="register-btn" :loading="loading" @click="submit">注 册</el-button>
      </el-form>
      <p class="tip">
        已有账号？<router-link to="/login" class="link">返回登录</router-link>
      </p>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { User, Lock, UserFilled, Iphone, Postcard, OfficeBuilding, Reading, VideoCamera } from '@element-plus/icons-vue'
import { ElMessage, type FormInstance, type FormRules } from 'element-plus'
import { useUserStore } from '@/stores/user'
import { useSiteStore } from '@/stores/site'
import { getAllDepartments, type DepartmentItem } from '@/api/auth'
import defaultLogo from '@/assets/logo.svg'

const router = useRouter()
const userStore = useUserStore()
const siteStore = useSiteStore()
siteStore.ensureLoaded()
const formRef = ref<FormInstance>()
const loading = ref(false)
const departments = ref<DepartmentItem[]>([])
const form = reactive({ username: '', phone: '', password: '', confirmPassword: '', nickname: '', studentNo: '', departmentId: undefined as number | undefined })

const validateConfirm = (_rule: any, value: string, callback: any) => {
  if (value !== form.password) {
    callback(new Error('两次输入的密码不一致'))
  } else {
    callback()
  }
}
const validatePhone = (_rule: any, value: string, callback: any) => {
  if (!value) {
    callback(new Error('请输入手机号'))
  } else if (!/^1\d{10}$/.test(value)) {
    callback(new Error('请输入正确的手机号'))
  } else {
    callback()
  }
}

const rules: FormRules = {
  username: [
    { required: true, message: '请输入账号', trigger: 'blur' },
    { min: 3, max: 50, message: '账号长度3-50位', trigger: 'blur' }
  ],
  phone: [{ required: true, validator: validatePhone, trigger: 'blur' }],
  password: [
    { required: true, message: '请输入密码', trigger: 'blur' },
    { min: 6, message: '密码至少6位', trigger: 'blur' }
  ],
  confirmPassword: [
    { required: true, message: '请确认密码', trigger: 'blur' },
    { validator: validateConfirm, trigger: 'blur' }
  ]
}

async function submit() {
  if (!formRef.value) return
  await formRef.value.validate(async (valid) => {
    if (!valid) return
    loading.value = true
    try {
      await userStore.register({
        username: form.username,
        phone: form.phone,
        password: form.password,
        nickname: form.nickname,
        studentNo: form.studentNo,
        departmentId: form.departmentId
      })
      ElMessage.success('注册成功，请登录')
      router.push('/login')
    } catch (e) {
      /* error handled in interceptor */
    } finally {
      loading.value = false
    }
  })
}

onMounted(async () => {
  try {
    const res = await getAllDepartments()
    departments.value = res.data || []
  } catch (e) {
    /* 院系加载失败不阻塞注册 */
  }
})
</script>

<style scoped lang="scss">
.register-container {
  height: 100%;
  display: flex;
  align-items: stretch;
  justify-content: center;
  background: #f5f7fa;
}
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
.register-card {
  width: 420px;
  padding: 32px 28px;
  border-radius: 0;
  text-align: center;
  box-shadow: -8px 0 30px rgba(0, 0, 0, 0.15);
}
.title {
  font-size: 24px;
  font-weight: 600;
  margin-bottom: 6px;
  color: #303133;
}
.subtitle {
  color: #909399;
  margin-bottom: 24px;
  font-size: 13px;
}
.register-btn {
  width: 100%;
  height: 40px;
  font-size: 15px;
  letter-spacing: 4px;
}
.tip {
  margin-top: 16px;
  color: #909399;
  font-size: 13px;
  .link {
    color: #409eff;
    text-decoration: none;
    &:hover {
      text-decoration: underline;
    }
  }
}
@media (max-width: 768px) {
  .brand-side {
    display: none;
  }
  .register-card {
    width: 100%;
  }
}
</style>
