<template>
  <div class="page">
    <el-card shadow="never">
      <div class="toolbar">
        <el-input v-model="query.keyword" placeholder="用户名/昵称" clearable style="width:220px" @keyup.enter="loadList" />
        <el-button type="primary" @click="loadList">查询</el-button>
        <el-button v-permission="'admin:add'" type="success" @click="openCreate">新增管理员</el-button>
      </div>

      <el-table :data="list" v-loading="loading" border stripe>
        <el-table-column type="index" label="#" width="50" />
        <el-table-column prop="username" label="用户名" width="120" />
        <el-table-column prop="nickname" label="昵称" width="120" />
        <el-table-column label="角色" min-width="180">
          <template #default="{ row }">
            <el-tag v-for="r in row.roles" :key="r.id" size="small" style="margin-right:6px">{{ r.title }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="phone" label="手机" width="130" />
        <el-table-column prop="email" label="邮箱" width="180" />
        <el-table-column label="状态" width="80">
          <template #default="{ row }">
            <el-tag :type="row.status === 1 ? 'success' : 'danger'" size="small">
              {{ row.status === 1 ? '启用' : '禁用' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="lastLoginAt" label="最后登录" width="170" />
        <el-table-column label="操作" width="240" fixed="right">
          <template #default="{ row }">
            <template v-if="row.username === 'admin'">
              <el-button v-permission="'admin:edit'" link type="warning" @click="openReset(row)">重置密码</el-button>
            </template>
            <template v-else>
              <el-button v-permission="'admin:edit'" link type="primary" @click="openEdit(row)">编辑</el-button>
              <el-button v-permission="'admin:edit'" link type="primary" @click="openRoles(row)">分配角色</el-button>
              <el-button v-permission="'admin:edit'" link type="warning" @click="openReset(row)">重置密码</el-button>
              <el-button v-permission="'admin:delete'" link type="danger" @click="handleDelete(row)">删除</el-button>
            </template>
          </template>
        </el-table-column>
      </el-table>

      <el-pagination
        class="pager"
        v-model:current-page="query.page"
        v-model:page-size="query.pageSize"
        :total="total"
        :page-sizes="[10, 20, 50]"
        layout="total, sizes, prev, pager, next, jumper"
        @size-change="loadList"
        @current-change="loadList"
      />
    </el-card>

    <!-- 新增/编辑 -->
    <el-dialog v-model="formVisible" :title="editId ? '编辑管理员' : '新增管理员'" width="520px" @closed="resetForm">
      <el-form ref="formRef" :model="form" :rules="formRules" label-width="80px">
        <el-form-item label="用户名" prop="username">
          <el-input v-model="form.username" :disabled="!!editId" placeholder="登录用户名" />
        </el-form-item>
        <el-form-item v-if="!editId" label="密码" prop="password">
          <el-input v-model="form.password" type="password" show-password placeholder="至少6位" />
        </el-form-item>
        <el-form-item label="昵称"><el-input v-model="form.nickname" /></el-form-item>
        <el-form-item label="角色">
          <el-select v-model="form.roleIds" multiple style="width:100%" placeholder="选择角色">
            <el-option v-for="r in roles" :key="r.id" :label="r.title" :value="r.id" />
          </el-select>
        </el-form-item>
        <el-form-item label="手机"><el-input v-model="form.phone" /></el-form-item>
        <el-form-item label="邮箱"><el-input v-model="form.email" /></el-form-item>
        <el-form-item label="状态">
          <el-radio-group v-model="form.status">
            <el-radio :value="1">启用</el-radio>
            <el-radio :value="0">禁用</el-radio>
          </el-radio-group>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="formVisible = false">取消</el-button>
        <el-button type="primary" :loading="saving" @click="submitForm">确定</el-button>
      </template>
    </el-dialog>

    <!-- 分配角色 -->
    <el-dialog v-model="rolesVisible" title="分配角色" width="420px">
      <el-checkbox-group v-model="assignRoleIds">
        <el-checkbox v-for="r in roles" :key="r.id" :label="r.title" :value="r.id" style="display:block;margin-bottom:8px" />
      </el-checkbox-group>
      <template #footer>
        <el-button @click="rolesVisible = false">取消</el-button>
        <el-button type="primary" :loading="saving" @click="submitRoles">确定</el-button>
      </template>
    </el-dialog>

    <!-- 重置密码 -->
    <el-dialog v-model="pwdVisible" title="重置密码" width="420px">
      <el-input v-model="newPwd" type="password" show-password placeholder="新密码（至少6位）" />
      <template #footer>
        <el-button @click="pwdVisible = false">取消</el-button>
        <el-button type="primary" :loading="saving" @click="submitReset">确定</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox, type FormInstance, type FormRules } from 'element-plus'
import { getAdminList, createAdmin, updateAdmin, assignAdminRoles, resetAdminPassword, deleteAdmin, type AdminItem, type AdminForm } from '@/api/admin'
import { getAllRoles, type RoleItem } from '@/api/role'

const loading = ref(false)
const list = ref<AdminItem[]>([])
const total = ref(0)
const query = reactive({ page: 1, pageSize: 10, keyword: '' })

const roles = ref<RoleItem[]>([])

async function loadList() {
  loading.value = true
  try {
    const res = await getAdminList(query)
    list.value = res.data.list
    total.value = res.data.total
  } finally {
    loading.value = false
  }
}

async function loadRoles() {
  const res = await getAllRoles('admin')
  roles.value = res.data || []
}

// 新增/编辑
const formVisible = ref(false)
const editId = ref(0)
const saving = ref(false)
const formRef = ref<FormInstance>()
const form = reactive<AdminForm & { password?: string }>({ username: '', password: '', nickname: '', phone: '', email: '', status: 1, roleIds: [] })
const formRules: FormRules = {
  username: [{ required: true, message: '请输入用户名', trigger: 'blur' }],
  password: [{ required: true, min: 6, message: '至少6位', trigger: 'blur' }]
}

function resetForm() {
  editId.value = 0
  Object.assign(form, { username: '', password: '', nickname: '', phone: '', email: '', status: 1, roleIds: [] })
  formRef.value?.clearValidate()
}

async function openCreate() {
  resetForm()
  await loadRoles()
  formVisible.value = true
}

async function openEdit(row: AdminItem) {
  editId.value = row.id
  Object.assign(form, {
    username: row.username,
    nickname: row.nickname,
    phone: row.phone,
    email: row.email,
    status: row.status,
    roleIds: row.roles.map((r) => r.id)
  })
  await loadRoles()
  formVisible.value = true
}

async function submitForm() {
  if (!formRef.value) return
  await formRef.value.validate(async (valid) => {
    if (!valid) return
    saving.value = true
    try {
      if (editId.value) {
        await updateAdmin(editId.value, { nickname: form.nickname, phone: form.phone, email: form.email, status: form.status, roleIds: form.roleIds })
      } else {
        await createAdmin(form)
      }
      ElMessage.success(editId.value ? '更新成功' : '创建成功')
      formVisible.value = false
      loadList()
    } finally {
      saving.value = false
    }
  })
}

// 分配角色
const rolesVisible = ref(false)
const assignRoleIds = ref<number[]>([])
let currentAdminId = 0
async function openRoles(row: AdminItem) {
  currentAdminId = row.id
  assignRoleIds.value = row.roles.map((r) => r.id)
  await loadRoles()
  rolesVisible.value = true
}
async function submitRoles() {
  saving.value = true
  try {
    await assignAdminRoles(currentAdminId, assignRoleIds.value)
    ElMessage.success('分配成功')
    rolesVisible.value = false
    loadList()
  } finally {
    saving.value = false
  }
}

// 重置密码
const pwdVisible = ref(false)
const newPwd = ref('')
function openReset(row: AdminItem) {
  currentAdminId = row.id
  newPwd.value = ''
  pwdVisible.value = true
}
async function submitReset() {
  if (!newPwd.value || newPwd.value.length < 6) {
    ElMessage.warning('密码至少6位')
    return
  }
  saving.value = true
  try {
    await resetAdminPassword(currentAdminId, newPwd.value)
    ElMessage.success('重置成功')
    pwdVisible.value = false
  } finally {
    saving.value = false
  }
}

async function handleDelete(row: AdminItem) {
  await ElMessageBox.confirm(`确定删除管理员「${row.username}」？`, '提示', { type: 'warning' })
  await deleteAdmin(row.id)
  ElMessage.success('删除成功')
  loadList()
}

onMounted(() => {
  loadRoles()
  loadList()
})
</script>

<style scoped>
.toolbar {
  display: flex;
  gap: 8px;
  margin-bottom: 16px;
}
.pager {
  margin-top: 16px;
  justify-content: flex-end;
}
</style>
