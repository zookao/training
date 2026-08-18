<template>
  <div class="profile-page">
    <div class="page-header">
      <h2 class="page-title">个人中心</h2>
      <p class="page-desc">管理你的资料与登录密码</p>
    </div>
    <el-row :gutter="20">
      <el-col :span="14">
        <el-card shadow="never">
          <template #header>
            <div class="card-header">基本资料</div>
          </template>
          <el-form ref="profileRef" :model="profileForm" :rules="profileRules" label-width="80px">
            <el-form-item label="账号">
              <el-input :model-value="userInfo?.username" disabled />
            </el-form-item>
            <el-form-item label="姓名" prop="nickname">
              <el-input v-model="profileForm.nickname" />
            </el-form-item>
            <el-form-item label="学号">
              <el-input v-model="profileForm.studentNo" />
            </el-form-item>
            <el-form-item label="院系">
              <el-input :model-value="departmentName" disabled placeholder="由管理员分配" />
            </el-form-item>
            <el-form-item label="手机" prop="phone">
              <el-input v-model="profileForm.phone" placeholder="手机号（可用于登录）" />
            </el-form-item>
            <el-form-item label="邮箱" prop="email">
              <el-input v-model="profileForm.email" />
            </el-form-item>
            <el-form-item>
              <el-button type="primary" :loading="savingProfile" @click="submitProfile">保存修改</el-button>
            </el-form-item>
          </el-form>
        </el-card>
      </el-col>
      <el-col :span="10">
        <el-card shadow="never">
          <template #header>
            <div class="card-header">修改密码</div>
          </template>
          <el-form ref="pwdRef" :model="pwdForm" :rules="pwdRules" label-width="80px">
            <el-form-item label="原密码" prop="oldPassword">
              <el-input v-model="pwdForm.oldPassword" type="password" show-password />
            </el-form-item>
            <el-form-item label="新密码" prop="newPassword">
              <el-input v-model="pwdForm.newPassword" type="password" show-password />
            </el-form-item>
            <el-form-item label="确认密码" prop="confirmPassword">
              <el-input v-model="pwdForm.confirmPassword" type="password" show-password />
            </el-form-item>
            <el-form-item>
              <el-button type="primary" :loading="savingPwd" @click="submitPwd">修改密码</el-button>
            </el-form-item>
          </el-form>
        </el-card>
      </el-col>
    </el-row>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted, computed } from 'vue'
import { ElMessage, type FormInstance, type FormRules } from 'element-plus'
import { useUserStore } from '@/stores/user'
import { updateProfile, changePassword, getAllDepartments, type DepartmentItem } from '@/api/auth'

const userStore = useUserStore()
const userInfo = computed(() => userStore.userInfo)

const profileRef = ref<FormInstance>()
const pwdRef = ref<FormInstance>()
const savingProfile = ref(false)
const savingPwd = ref(false)
const departments = ref<DepartmentItem[]>([])

const profileForm = reactive({ nickname: '', studentNo: '', phone: '', email: '' })
// 院系由管理员分配，学员仅可查看不可修改
const departmentName = computed(() => {
  const id = userInfo.value?.departmentId
  if (!id) return ''
  return departments.value.find((d) => d.id === id)?.name || ''
})
const profileRules: FormRules = {
  nickname: [{ required: true, message: '请输入姓名', trigger: 'blur' }],
  phone: [{ required: true, message: '请输入手机号', trigger: 'blur' }]
}

const pwdForm = reactive({ oldPassword: '', newPassword: '', confirmPassword: '' })
const validateConfirm = (_rule: any, value: string, callback: any) => {
  if (value !== pwdForm.newPassword) {
    callback(new Error('两次输入的密码不一致'))
  } else {
    callback()
  }
}
const pwdRules: FormRules = {
  oldPassword: [{ required: true, message: '请输入原密码', trigger: 'blur' }],
  newPassword: [
    { required: true, message: '请输入新密码', trigger: 'blur' },
    { min: 6, message: '密码至少6位', trigger: 'blur' }
  ],
  confirmPassword: [
    { required: true, message: '请确认新密码', trigger: 'blur' },
    { validator: validateConfirm, trigger: 'blur' }
  ]
}

async function submitProfile() {
  if (!profileRef.value) return
  await profileRef.value.validate(async (valid) => {
    if (!valid) return
    savingProfile.value = true
    try {
      await updateProfile(profileForm)
      ElMessage.success('保存成功')
      await userStore.fetchUserInfo()
    } finally {
      savingProfile.value = false
    }
  })
}

async function submitPwd() {
  if (!pwdRef.value) return
  await pwdRef.value.validate(async (valid) => {
    if (!valid) return
    savingPwd.value = true
    try {
      await changePassword({ oldPassword: pwdForm.oldPassword, newPassword: pwdForm.newPassword })
      ElMessage.success('修改成功')
      pwdRef.value?.resetFields()
    } finally {
      savingPwd.value = false
    }
  })
}

onMounted(async () => {
  if (!userStore.userInfo) {
    await userStore.fetchUserInfo()
  }
  const info = userStore.userInfo
  if (info) {
    profileForm.nickname = info.nickname
    profileForm.studentNo = info.studentNo
    profileForm.phone = info.phone
    profileForm.email = info.email
  }
  try {
    const res = await getAllDepartments()
    departments.value = res.data || []
  } catch (e) {
    /* 院系加载失败不阻塞页面 */
  }
})
</script>

<style scoped lang="scss">
.profile-page {
  max-width: 1100px;
  margin: 0 auto;
}
.page-header {
  margin-bottom: 20px;
  .page-title {
    margin: 0 0 6px;
    font-size: 22px;
  }
  .page-desc {
    margin: 0;
    color: #909399;
    font-size: 13px;
  }
}
.card-header {
  font-weight: 600;
}
</style>
