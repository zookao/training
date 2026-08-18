<template>
  <div class="page">
    <el-card shadow="never">
      <div class="toolbar">
        <el-input v-model="query.keyword" placeholder="用户名/姓名/手机/学号" clearable style="width:240px" @keyup.enter="loadList" />
        <el-button type="primary" @click="loadList">查询</el-button>
        <el-button v-permission="'user:add'" type="success" @click="openCreate">新增学员</el-button>
        <el-button v-permission="'user:import'" type="warning" plain @click="openImport">批量导入</el-button>
      </div>

      <el-table :data="list" v-loading="loading" border stripe>
        <el-table-column type="index" label="#" width="50" />
        <el-table-column prop="username" label="用户名" width="140" />
        <el-table-column prop="nickname" label="姓名" width="120" />
        <el-table-column prop="studentNo" label="学号" width="120" />
        <el-table-column label="院系" width="140">
          <template #default="{ row }">
            {{ departmentMap[row.departmentId] || '-' }}
          </template>
        </el-table-column>
        <el-table-column prop="phone" label="手机" width="140" />
        <el-table-column prop="email" label="邮箱" min-width="180" />
        <el-table-column label="状态" width="90">
          <template #default="{ row }">
            <el-tag :type="row.status === 1 ? 'success' : 'danger'" size="small">
              {{ row.status === 1 ? '启用' : '禁用' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="lastLoginAt" label="最后登录" width="170" />
        <el-table-column prop="createdAt" label="注册时间" width="170" />
        <el-table-column label="操作" width="240" fixed="right">
          <template #default="{ row }">
            <el-button v-permission="'user:edit'" link type="primary" @click="openEdit(row)">编辑</el-button>
            <el-button v-permission="'user:edit'" link type="warning" @click="openReset(row)">重置密码</el-button>
            <el-button v-permission="'user:delete'" link type="danger" @click="handleDelete(row)">删除</el-button>
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
    <el-dialog v-model="formVisible" :title="editId ? '编辑学员' : '新增学员'" width="520px" @closed="resetForm">
      <el-form ref="formRef" :model="form" :rules="formRules" label-width="80px">
        <el-form-item label="用户名" prop="username">
          <el-input v-model="form.username" :disabled="!!editId" placeholder="登录用户名" />
        </el-form-item>
        <el-form-item v-if="!editId" label="密码" prop="password">
          <el-input v-model="form.password" type="password" show-password placeholder="至少6位" />
        </el-form-item>
        <el-form-item label="姓名"><el-input v-model="form.nickname" /></el-form-item>
        <el-form-item label="学号"><el-input v-model="form.studentNo" /></el-form-item>
        <el-form-item label="院系">
          <el-select v-model="form.departmentId" clearable placeholder="选择院系（选填）" style="width:100%">
            <el-option v-for="d in departments" :key="d.id" :label="d.name" :value="d.id" />
          </el-select>
        </el-form-item>
        <el-form-item label="手机" prop="phone"><el-input v-model="form.phone" /></el-form-item>
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

    <!-- 重置密码 -->
    <el-dialog v-model="pwdVisible" title="重置密码" width="420px">
      <el-input v-model="newPwd" type="password" show-password placeholder="新密码（至少6位）" />
      <template #footer>
        <el-button @click="pwdVisible = false">取消</el-button>
        <el-button type="primary" :loading="saving" @click="submitReset">确定</el-button>
      </template>
    </el-dialog>

    <!-- 批量导入 -->
    <el-dialog v-model="importVisible" title="批量导入学员" width="900px" @closed="resetImport">
      <el-alert
        type="info"
        :closable="false"
        show-icon
        title="Excel 字段：账号、手机号、姓名、学号、院系"
        description="规则：账号和手机号必填，初始密码为手机号，状态默认启用。院系须填写已存在的院系名称。请先下载模板填写。"
        style="margin-bottom: 12px"
      />
      <div class="import-actions">
        <input ref="importInputRef" type="file" accept=".xlsx,.xls" style="display: none" @change="onImportPick" />
        <el-button @click="importInputRef?.click()">选择 Excel 文件</el-button>
        <el-button link type="primary" @click="onDownloadTemplate">下载模板</el-button>
        <span class="file-name">{{ importFile?.name }}</span>
      </div>
      <div v-if="importResult" class="import-summary">
        <el-tag type="success">成功 {{ importResult.success }}</el-tag>
        <el-tag type="danger" style="margin-left: 8px">失败 {{ importResult.failed }}</el-tag>
        <el-tag style="margin-left: 8px">总计 {{ importResult.total }}</el-tag>
      </div>
      <el-table
        v-if="importResult && importResult.rows.length"
        :data="importResult.rows"
        border
        size="small"
        style="margin-top: 12px"
        max-height="280"
      >
        <el-table-column prop="row" label="行号" width="70" align="center" />
        <el-table-column prop="username" label="账号" width="120" />
        <el-table-column prop="phone" label="手机号" width="130" />
        <el-table-column prop="name" label="姓名" width="90" />
        <el-table-column prop="studentNo" label="学号" width="110" />
        <el-table-column prop="department" label="院系" width="120" />
        <el-table-column label="结果" width="80" align="center">
          <template #default="{ row }">
            <el-tag :type="row.success ? 'success' : 'danger'" size="small">
              {{ row.success ? '成功' : '失败' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="reason" label="原因" min-width="140" show-overflow-tooltip />
      </el-table>
      <template #footer>
        <el-button @click="importVisible = false">关闭</el-button>
        <el-button type="primary" :loading="importing" :disabled="!importFile" @click="submitImport">开始导入</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted } from 'vue'
import { ElMessage, ElMessageBox, type FormInstance, type FormRules } from 'element-plus'
import {
  getUserList, createUser, updateUser, resetUserPassword, deleteUser,
  importUsers, downloadUserImportTemplate,
  type UserItem, type UserForm, type UserImportResult
} from '@/api/user'
import { getAllDepartments, type DepartmentItem } from '@/api/department'

const loading = ref(false)
const list = ref<UserItem[]>([])
const total = ref(0)
const query = reactive({ page: 1, pageSize: 10, keyword: '' })

const departments = ref<DepartmentItem[]>([])
const departmentMap = computed(() => {
  const m: Record<number, string> = {}
  departments.value.forEach((d) => { m[d.id] = d.name })
  return m
})

async function loadDepartments() {
  const res = await getAllDepartments()
  departments.value = res.data || []
}

async function loadList() {
  loading.value = true
  try {
    const res = await getUserList(query)
    list.value = res.data.list
    total.value = res.data.total
  } finally {
    loading.value = false
  }
}

// 新增/编辑
const formVisible = ref(false)
const editId = ref(0)
const saving = ref(false)
const formRef = ref<FormInstance>()
const form = reactive<UserForm & { password?: string }>({ username: '', password: '', nickname: '', studentNo: '', departmentId: undefined, phone: '', email: '', status: 1 })
const formRules: FormRules = {
  username: [{ required: true, message: '请输入用户名', trigger: 'blur' }],
  password: [{ required: true, min: 6, message: '至少6位', trigger: 'blur' }],
  phone: [{ required: true, message: '请输入手机号', trigger: 'blur' }]
}

function resetForm() {
  editId.value = 0
  Object.assign(form, { username: '', password: '', nickname: '', studentNo: '', departmentId: undefined, phone: '', email: '', status: 1 })
  formRef.value?.clearValidate()
}

async function openCreate() {
  resetForm()
  await loadDepartments()
  formVisible.value = true
}

async function openEdit(row: UserItem) {
  editId.value = row.id
  Object.assign(form, {
    username: row.username,
    nickname: row.nickname,
    studentNo: row.studentNo,
    departmentId: row.departmentId || undefined,
    phone: row.phone,
    email: row.email,
    status: row.status
  })
  await loadDepartments()
  formVisible.value = true
}

async function submitForm() {
  if (!formRef.value) return
  await formRef.value.validate(async (valid) => {
    if (!valid) return
    saving.value = true
    try {
      if (editId.value) {
        await updateUser(editId.value, { nickname: form.nickname, studentNo: form.studentNo, departmentId: form.departmentId, phone: form.phone, email: form.email, status: form.status })
      } else {
        await createUser(form)
      }
      ElMessage.success(editId.value ? '更新成功' : '创建成功')
      formVisible.value = false
      loadList()
    } finally {
      saving.value = false
    }
  })
}

// 重置密码
const pwdVisible = ref(false)
const newPwd = ref('')
let currentId = 0
function openReset(row: UserItem) {
  currentId = row.id
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
    await resetUserPassword(currentId, newPwd.value)
    ElMessage.success('重置成功')
    pwdVisible.value = false
  } finally {
    saving.value = false
  }
}

async function handleDelete(row: UserItem) {
  await ElMessageBox.confirm(`确定删除学员「${row.username}」？`, '提示', { type: 'warning' })
  await deleteUser(row.id)
  ElMessage.success('删除成功')
  loadList()
}

// 批量导入
const importVisible = ref(false)
const importing = ref(false)
const importInputRef = ref<HTMLInputElement | null>(null)
const importFile = ref<File | null>(null)
const importResult = ref<UserImportResult | null>(null)

function openImport() {
  importFile.value = null
  importResult.value = null
  importVisible.value = true
}
function resetImport() {
  importFile.value = null
  importResult.value = null
}
function onImportPick(e: Event) {
  const target = e.target as HTMLInputElement
  if (target.files && target.files.length) {
    importFile.value = target.files[0]
    importResult.value = null
  }
}
async function onDownloadTemplate() {
  try {
    await downloadUserImportTemplate()
    ElMessage.success('模板已下载')
  } catch (e) {
    /* handled in interceptor */
  }
}
async function submitImport() {
  if (!importFile.value) return
  importing.value = true
  try {
    const res = await importUsers(importFile.value)
    importResult.value = res.data
    if (res.data.success > 0) {
      ElMessage.success(`导入成功 ${res.data.success} 条`)
      loadList()
    } else {
      ElMessage.warning('未导入成功任何记录')
    }
  } catch (e) {
    /* handled in interceptor */
  } finally {
    importing.value = false
  }
}

onMounted(() => {
  loadDepartments()
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
.import-actions {
  display: flex;
  align-items: center;
  gap: 12px;
}
.file-name {
  color: #909399;
  font-size: 13px;
}
.import-summary {
  margin-top: 12px;
}
</style>
