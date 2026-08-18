<template>
  <div class="page">
    <el-card shadow="never">
      <div class="toolbar">
        <el-input v-model="query.keyword" placeholder="角色名称/标识" clearable style="width:220px" @keyup.enter="loadList" />
        <el-button type="primary" @click="loadList">查询</el-button>
        <el-button v-permission="'role:add'" type="success" @click="openCreate">新增角色</el-button>
      </div>

      <el-table :data="list" v-loading="loading" border stripe>
        <el-table-column type="index" label="#" width="50" />
        <el-table-column prop="title" label="角色名称" width="160" />
        <el-table-column prop="name" label="标识" width="160" />
        <el-table-column prop="guardType" label="端" width="80">
          <template #default="{ row }">
            <el-tag size="small" :type="row.guardType === 'admin' ? 'primary' : 'success'">{{ row.guardType }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="sort" label="排序" width="80" />
        <el-table-column label="状态" width="80">
          <template #default="{ row }">
            <el-tag :type="row.status === 1 ? 'success' : 'danger'" size="small">{{ row.status === 1 ? '启用' : '禁用' }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="remark" label="备注" min-width="120" />
        <el-table-column label="操作" width="280" fixed="right">
          <template #default="{ row }">
            <template v-if="row.id === 1">
              <el-tag type="info" size="small">超级管理员</el-tag>
            </template>
            <template v-else>
              <el-button v-permission="'role:edit'" link type="primary" @click="openEdit(row)">编辑</el-button>
              <el-button link type="primary" @click="openMenus(row)">分配菜单</el-button>
              <el-button link type="primary" @click="openApis(row)">分配接口</el-button>
              <el-button v-permission="'role:delete'" link type="danger" @click="handleDelete(row)">删除</el-button>
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
    <el-dialog v-model="formVisible" :title="editId ? '编辑角色' : '新增角色'" width="500px" @closed="resetForm">
      <el-form ref="formRef" :model="form" :rules="formRules" label-width="80px">
        <el-form-item label="标识" prop="name">
          <el-input v-model="form.name" placeholder="如 editor" />
        </el-form-item>
        <el-form-item label="名称" prop="title">
          <el-input v-model="form.title" placeholder="如 编辑员" />
        </el-form-item>
        <el-form-item label="端">
          <el-radio-group v-model="form.guardType">
            <el-radio value="admin">后台</el-radio>
            <!-- <el-radio value="user">user</el-radio> -->
          </el-radio-group>
        </el-form-item>
        <el-form-item label="排序"><el-input-number v-model="form.sort" :min="0" /></el-form-item>
        <el-form-item label="状态">
          <el-radio-group v-model="form.status">
            <el-radio :value="1">启用</el-radio>
            <el-radio :value="0">禁用</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item label="备注"><el-input v-model="form.remark" type="textarea" /></el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="formVisible = false">取消</el-button>
        <el-button type="primary" :loading="saving" @click="submitForm">确定</el-button>
      </template>
    </el-dialog>

    <!-- 分配菜单 -->
    <el-dialog v-model="menusVisible" title="分配菜单" width="520px">
      <el-tree
        ref="menuTreeRef"
        :data="menuTree"
        node-key="id"
        show-checkbox
        :default-checked-keys="checkedMenuIds"
        :props="{ label: 'name', children: 'children' }"
        check-strictly
      />
      <template #footer>
        <el-button @click="menusVisible = false">取消</el-button>
        <el-button type="primary" :loading="saving" @click="submitMenus">确定</el-button>
      </template>
    </el-dialog>

    <!-- 分配接口 -->
    <el-dialog v-model="apisVisible" title="分配接口" width="640px" @opened="restoreApiSelection">
      <el-input v-model="apiSearch" placeholder="按分组/路径搜索" clearable style="margin-bottom:12px" />
      <el-table ref="apiTableRef" :data="filteredApiList" border max-height="380" @selection-change="onApiSelect" row-key="id">
        <el-table-column type="selection" width="45" :reserve-selection="true" />
        <el-table-column prop="group" label="分组" width="100" />
        <el-table-column prop="method" label="方法" width="80" />
        <el-table-column prop="path" label="路径" />
        <el-table-column prop="description" label="说明" width="140" />
      </el-table>
      <template #footer>
        <el-button @click="apisVisible = false">取消</el-button>
        <el-button type="primary" :loading="saving" @click="submitApis">确定</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted, nextTick, computed } from 'vue'
import { ElMessage, ElMessageBox, type FormInstance, type FormRules, type TreeInstance, type TableInstance } from 'element-plus'
import { getRoleList, createRole, updateRole, deleteRole, getRoleMenuIds, getRoleApiIds, assignRoleMenus, assignRoleApis, type RoleItem, type RoleForm } from '@/api/role'
import { getMenuTree } from '@/api/menu'
import { getAllApis, type ApiItem } from '@/api/api'
import type { MenuItem } from '@/api/auth'

const loading = ref(false)
const list = ref<RoleItem[]>([])
const total = ref(0)
const query = reactive({ page: 1, pageSize: 10, keyword: '', guardType: 'admin' })

async function loadList() {
  loading.value = true
  try {
    const res = await getRoleList(query)
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
const form = reactive<RoleForm>({ name: '', title: '', guardType: 'admin', sort: 0, status: 1, remark: '' })
const formRules: FormRules = {
  name: [{ required: true, message: '请输入标识', trigger: 'blur' }],
  title: [{ required: true, message: '请输入名称', trigger: 'blur' }]
}
function resetForm() {
  editId.value = 0
  Object.assign(form, { name: '', title: '', guardType: 'admin', sort: 0, status: 1, remark: '' })
  formRef.value?.clearValidate()
}
function openCreate() {
  resetForm()
  formVisible.value = true
}
function openEdit(row: RoleItem) {
  editId.value = row.id
  Object.assign(form, { name: row.name, title: row.title, guardType: row.guardType, sort: row.sort, status: row.status, remark: row.remark })
  formVisible.value = true
}
async function submitForm() {
  if (!formRef.value) return
  await formRef.value.validate(async (valid) => {
    if (!valid) return
    saving.value = true
    try {
      if (editId.value) {
        await updateRole(editId.value, form)
      } else {
        await createRole(form)
      }
      ElMessage.success(editId.value ? '更新成功' : '创建成功')
      formVisible.value = false
      loadList()
    } finally {
      saving.value = false
    }
  })
}

async function handleDelete(row: RoleItem) {
  await ElMessageBox.confirm(`确定删除角色「${row.title}」？`, '提示', { type: 'warning' })
  await deleteRole(row.id)
  ElMessage.success('删除成功')
  loadList()
}

// 分配菜单
const menusVisible = ref(false)
const menuTree = ref<MenuItem[]>([])
const checkedMenuIds = ref<number[]>([])
const menuTreeRef = ref<TreeInstance>()
let currentRoleId = 0
async function openMenus(row: RoleItem) {
  currentRoleId = row.id
  const [treeRes, idsRes] = await Promise.all([getMenuTree('admin'), getRoleMenuIds(row.id)])
  menuTree.value = treeRes.data || []
  checkedMenuIds.value = idsRes.data || []
  menusVisible.value = true
  await nextTick()
  menuTreeRef.value?.setCheckedKeys(checkedMenuIds.value, false)
}
async function submitMenus() {
  const ids = menuTreeRef.value?.getCheckedKeys(false) as number[]
  saving.value = true
  try {
    await assignRoleMenus(currentRoleId, ids || [])
    ElMessage.success('分配成功')
    menusVisible.value = false
  } finally {
    saving.value = false
  }
}

// 分配接口
const apisVisible = ref(false)
const apiList = ref<ApiItem[]>([])
const apiSearch = ref('')
const filteredApiList = computed(() => {
  const kw = apiSearch.value.trim().toLowerCase()
  if (!kw) return apiList.value
  return apiList.value.filter(a =>
    a.group.toLowerCase().includes(kw) ||
    a.path.toLowerCase().includes(kw)
  )
})
const selectedApiIds = ref<number[]>([])
const apiTableRef = ref<TableInstance>()
const initialApiIds = ref<number[]>([])
async function openApis(row: RoleItem) {
  currentRoleId = row.id
  apiSearch.value = ''
  const [apiRes, idsRes] = await Promise.all([getAllApis(), getRoleApiIds(row.id)])
  apiList.value = apiRes.data || []
  initialApiIds.value = idsRes.data || []
  selectedApiIds.value = [...initialApiIds.value]
  apisVisible.value = true
}
// dialog 打开后回显已选接口（用 initialApiIds 判断，避免 onApiSelect 覆盖 selectedApiIds）
function restoreApiSelection() {
  apiTableRef.value?.clearSelection()
  apiList.value.forEach((a) => {
    if (initialApiIds.value.includes(a.id)) {
      apiTableRef.value?.toggleRowSelection(a, true)
    }
  })
}
function onApiSelect(rows: ApiItem[]) {
  const visibleSelected = new Set(rows.map((r) => r.id))
  const result = new Set<number>(visibleSelected)
  // 保留被搜索过滤掉的已选行，避免搜索时丢失选择
  apiList.value.forEach(a => {
    if (!filteredApiList.value.some(f => f.id === a.id) && selectedApiIds.value.includes(a.id)) {
      result.add(a.id)
    }
  })
  selectedApiIds.value = [...result]
}
async function submitApis() {
  saving.value = true
  try {
    await assignRoleApis(currentRoleId, selectedApiIds.value)
    ElMessage.success('分配成功')
    apisVisible.value = false
  } finally {
    saving.value = false
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
