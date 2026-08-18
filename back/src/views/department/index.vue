<template>
  <div class="page">
    <el-card shadow="never">
      <div class="toolbar">
        <el-input v-model="query.keyword" placeholder="院系名称" clearable style="width:220px" @keyup.enter="loadList" />
        <el-button type="primary" @click="loadList">查询</el-button>
        <el-button v-permission="'department:add'" type="success" @click="openCreate">新增院系</el-button>
      </div>

      <el-table :data="list" v-loading="loading" border stripe>
        <el-table-column type="index" label="#" width="50" />
        <el-table-column prop="name" label="院系名称" min-width="160" />
        <el-table-column prop="description" label="描述" min-width="200" show-overflow-tooltip />
        <el-table-column prop="sort" label="排序" width="80" />
        <el-table-column label="状态" width="80">
          <template #default="{ row }">
            <el-tag :type="row.status === 1 ? 'success' : 'danger'" size="small">
              {{ row.status === 1 ? '启用' : '禁用' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="createdAt" label="创建时间" width="170" />
        <el-table-column label="操作" width="280" fixed="right">
          <template #default="{ row }">
            <el-button v-permission="'department:edit'" link type="primary" @click="openStudents(row)">学员列表</el-button>
            <el-button v-permission="'department:edit'" link type="primary" @click="openImport(row)">导入学员</el-button>
            <el-button v-permission="'department:edit'" link type="primary" @click="openEdit(row)">编辑</el-button>
            <el-button v-permission="'department:delete'" link type="danger" @click="handleDelete(row)">删除</el-button>
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
    <el-dialog v-model="formVisible" :title="editId ? '编辑院系' : '新增院系'" width="520px" @closed="resetForm">
      <el-form ref="formRef" :model="form" :rules="formRules" label-width="80px">
        <el-form-item label="名称" prop="name">
          <el-input v-model="form.name" placeholder="如 计算机学院" />
        </el-form-item>
        <el-form-item label="描述">
          <el-input v-model="form.description" type="textarea" :rows="3" placeholder="院系描述（选填）" />
        </el-form-item>
        <el-form-item label="排序"><el-input-number v-model="form.sort" :min="0" /></el-form-item>
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

    <!-- 院系学员列表 -->
    <el-dialog v-model="studentVisible" :title="`${currentDeptName} - 学员列表`" width="820px">
      <el-table :data="students" v-loading="studentLoading" border stripe max-height="500">
        <el-table-column type="index" label="#" width="50" />
        <el-table-column prop="username" label="账号" min-width="120" />
        <el-table-column prop="nickname" label="姓名" width="100" />
        <el-table-column prop="studentNo" label="学号" width="120" />
        <el-table-column prop="phone" label="手机号" width="140" />
        <el-table-column label="操作" width="100">
          <template #default="{ row }">
            <el-button v-permission="'department:edit'" link type="danger" @click="handleRemoveStudent(row)">解除绑定</el-button>
          </template>
        </el-table-column>
      </el-table>
      <template #footer>
        <el-button @click="studentVisible = false">关闭</el-button>
      </template>
    </el-dialog>

    <!-- 导入学员到院系 -->
    <el-dialog v-model="importVisible" :title="`导入学员到「${currentDeptName}」`" width="780px" @closed="resetImport">
      <el-alert
        type="info"
        :closable="false"
        show-icon
        title="Excel 字段：手机号"
        description="规则：仅支持已存在的学员；已绑定其他院系的学员将提示失败。请先下载模板填写。"
        style="margin-bottom: 12px"
      />
      <div style="display:flex; gap:8px; align-items:center; margin-bottom:12px;">
        <el-upload :show-file-list="false" :auto-upload="false" accept=".xlsx,.xls" :on-change="onImportFile">
          <el-button type="primary">选择文件</el-button>
        </el-upload>
        <el-button @click="onDownloadTemplate">下载模板</el-button>
        <span v-if="importFile" style="color:#909399; font-size:13px;">{{ importFile.name }}</span>
      </div>
      <div v-if="importResult" style="margin-top:12px;">
        <el-alert type="success" :closable="false" show-icon style="margin-bottom:12px;">
          共 {{ importResult.total }} 条，成功 {{ importResult.success }} 条，失败 {{ importResult.failed }} 条
        </el-alert>
        <el-table :data="importResult.rows" border max-height="380">
          <el-table-column prop="row" label="行号" width="70" align="center" />
          <el-table-column prop="phone" label="手机号" width="140" />
          <el-table-column prop="name" label="姓名" width="100" />
          <el-table-column prop="result" label="结果" width="80" align="center">
            <template #default="{ row }">
              <el-tag :type="row.result === '成功' ? 'success' : 'danger'" size="small">{{ row.result }}</el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="reason" label="原因" min-width="140" show-overflow-tooltip />
        </el-table>
      </div>
      <template #footer>
        <el-button @click="importVisible = false">取消</el-button>
        <el-button type="primary" :loading="importing" :disabled="!importFile" @click="submitImport">开始导入</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox, type FormInstance, type FormRules } from 'element-plus'
import {
  getDepartmentList, createDepartment, updateDepartment, deleteDepartment,
  getDepartmentStudents, removeDepartmentStudent, importDepartmentStudents, downloadDeptStudentTemplate,
  type DepartmentItem, type DepartmentForm, type DeptStudentItem, type DeptImportResult
} from '@/api/department'

const loading = ref(false)
const list = ref<DepartmentItem[]>([])
const total = ref(0)
const query = reactive({ page: 1, pageSize: 10, keyword: '' })

async function loadList() {
  loading.value = true
  try {
    const res = await getDepartmentList(query)
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
const form = reactive<DepartmentForm>({ name: '', description: '', sort: 0, status: 1 })
const formRules: FormRules = {
  name: [{ required: true, message: '请输入院系名称', trigger: 'blur' }]
}

function resetForm() {
  editId.value = 0
  Object.assign(form, { name: '', description: '', sort: 0, status: 1 })
  formRef.value?.clearValidate()
}

function openCreate() {
  resetForm()
  formVisible.value = true
}

function openEdit(row: DepartmentItem) {
  editId.value = row.id
  Object.assign(form, {
    name: row.name,
    description: row.description,
    sort: row.sort,
    status: row.status
  })
  formVisible.value = true
}

async function submitForm() {
  if (!formRef.value) return
  await formRef.value.validate(async (valid) => {
    if (!valid) return
    saving.value = true
    try {
      if (editId.value) {
        await updateDepartment(editId.value, form)
      } else {
        await createDepartment(form)
      }
      ElMessage.success(editId.value ? '更新成功' : '创建成功')
      formVisible.value = false
      loadList()
    } finally {
      saving.value = false
    }
  })
}

async function handleDelete(row: DepartmentItem) {
  await ElMessageBox.confirm(`确定删除院系「${row.name}」？`, '提示', { type: 'warning' })
  await deleteDepartment(row.id)
  ElMessage.success('删除成功')
  loadList()
}

// 院系学员列表
const studentVisible = ref(false)
const studentLoading = ref(false)
const students = ref<DeptStudentItem[]>([])
const currentDeptId = ref(0)
const currentDeptName = ref('')

async function openStudents(row: DepartmentItem) {
  currentDeptId.value = row.id
  currentDeptName.value = row.name
  studentVisible.value = true
  studentLoading.value = true
  try {
    const res = await getDepartmentStudents(row.id)
    students.value = res.data || []
  } finally {
    studentLoading.value = false
  }
}

async function handleRemoveStudent(row: DeptStudentItem) {
  await ElMessageBox.confirm(`确定解除「${row.nickname || row.username}」与该院系的绑定？`, '提示', { type: 'warning' })
  await removeDepartmentStudent(currentDeptId.value, row.id)
  ElMessage.success('已解除绑定')
  students.value = students.value.filter(s => s.id !== row.id)
}

// 导入学员到院系
const importVisible = ref(false)
const importing = ref(false)
const importFile = ref<File | null>(null)
const importResult = ref<DeptImportResult | null>(null)

function openImport(row: DepartmentItem) {
  currentDeptId.value = row.id
  currentDeptName.value = row.name
  importFile.value = null
  importResult.value = null
  importVisible.value = true
}

function resetImport() {
  importFile.value = null
  importResult.value = null
}

function onImportFile(file: any) {
  importFile.value = file.raw
}

async function onDownloadTemplate() {
  try {
    await downloadDeptStudentTemplate()
    ElMessage.success('模板已下载')
  } catch (e) {
    /* handled in interceptor */
  }
}

async function submitImport() {
  if (!importFile.value) return
  importing.value = true
  try {
    const res = await importDepartmentStudents(currentDeptId.value, importFile.value)
    importResult.value = res.data
    ElMessage.success('导入完成')
  } finally {
    importing.value = false
  }
}

onMounted(loadList)
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
