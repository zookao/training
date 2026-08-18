<template>
  <div class="page">
    <el-card shadow="never">
      <div class="toolbar">
        <el-input v-model="query.keyword" placeholder="路径/说明/分组" clearable style="width:240px" @keyup.enter="loadList" />
        <el-button type="primary" @click="loadList">查询</el-button>
      </div>

      <el-table :data="list" v-loading="loading" border stripe>
        <el-table-column type="index" label="#" width="50" />
        <el-table-column prop="group" label="分组" width="100" />
        <el-table-column label="方法" width="90">
          <template #default="{ row }">
            <el-tag size="small" :type="methodTag(row.method)">{{ row.method }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="path" label="路径" min-width="200" />
        <el-table-column prop="description" label="说明" min-width="140" />
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

    <el-dialog v-model="formVisible" :title="editId ? '编辑接口' : '新增接口'" width="500px" @closed="resetForm">
      <el-form ref="formRef" :model="form" :rules="formRules" label-width="80px">
        <el-form-item label="分组" prop="group"><el-input v-model="form.group" placeholder="如 管理员" /></el-form-item>
        <el-form-item label="方法" prop="method">
          <el-select v-model="form.method" style="width:100%">
            <el-option v-for="m in ['GET', 'POST', 'PUT', 'DELETE', 'PATCH']" :key="m" :label="m" :value="m" />
          </el-select>
        </el-form-item>
        <el-form-item label="路径" prop="path"><el-input v-model="form.path" placeholder="如 /api/admin/admin" /></el-form-item>
        <el-form-item label="说明"><el-input v-model="form.description" /></el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="formVisible = false">取消</el-button>
        <el-button type="primary" :loading="saving" @click="submitForm">确定</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox, type FormInstance, type FormRules } from 'element-plus'
import { getApiList, createApi, updateApi, deleteApi, type ApiItem, type ApiForm } from '@/api/api'

const loading = ref(false)
const list = ref<ApiItem[]>([])
const total = ref(0)
const query = reactive({ page: 1, pageSize: 10, keyword: '' })

async function loadList() {
  loading.value = true
  try {
    const res = await getApiList(query)
    list.value = res.data.list
    total.value = res.data.total
  } finally {
    loading.value = false
  }
}

function methodTag(m: string): any {
  return { GET: 'success', POST: 'primary', PUT: 'warning', DELETE: 'danger', PATCH: 'info' }[m] || ''
}

const formVisible = ref(false)
const editId = ref(0)
const saving = ref(false)
const formRef = ref<FormInstance>()
const form = reactive<ApiForm>({ group: '', method: 'GET', path: '', description: '' })
const formRules: FormRules = {
  group: [{ required: true, message: '请输入分组', trigger: 'blur' }],
  method: [{ required: true, message: '请选择方法', trigger: 'change' }],
  path: [{ required: true, message: '请输入路径', trigger: 'blur' }]
}
function resetForm() {
  editId.value = 0
  Object.assign(form, { group: '', method: 'GET', path: '', description: '' })
  formRef.value?.clearValidate()
}
function openCreate() {
  resetForm()
  formVisible.value = true
}
function openEdit(row: ApiItem) {
  editId.value = row.id
  Object.assign(form, { group: row.group, method: row.method, path: row.path, description: row.description })
  formVisible.value = true
}
async function submitForm() {
  if (!formRef.value) return
  await formRef.value.validate(async (valid) => {
    if (!valid) return
    saving.value = true
    try {
      if (editId.value) {
        await updateApi(editId.value, form)
      } else {
        await createApi(form)
      }
      ElMessage.success(editId.value ? '更新成功' : '创建成功')
      formVisible.value = false
      loadList()
    } finally {
      saving.value = false
    }
  })
}
async function handleDelete(row: ApiItem) {
  await ElMessageBox.confirm(`确定删除接口「${row.description || row.path}」？`, '提示', { type: 'warning' })
  await deleteApi(row.id)
  ElMessage.success('删除成功')
  loadList()
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
