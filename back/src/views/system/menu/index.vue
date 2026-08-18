<template>
  <div class="page">
    <el-card shadow="never">
      <div class="toolbar">
        <el-button @click="toggleExpand">展开/折叠</el-button>
      </div>

      <el-table
        v-if="list.length"
        ref="tableRef"
        :data="list"
        row-key="id"
        border
        :default-expand-all="expandAll"
        :tree-props="{ children: 'children' }"
      >
        <el-table-column prop="name" label="名称" min-width="180" />
        <el-table-column label="图标" width="80">
          <template #default="{ row }">
            <el-icon v-if="row.icon"><component :is="row.icon" /></el-icon>
          </template>
        </el-table-column>
        <el-table-column label="类型" width="80">
          <template #default="{ row }">
            <el-tag size="small" :type="typeTag(row.type)">{{ typeText(row.type) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="path" label="路径" width="160" />
        <el-table-column prop="component" label="组件" width="200" />
        <el-table-column prop="perms" label="权限码" width="140" />
        <el-table-column prop="sort" label="排序" width="70" />
        <el-table-column label="状态" width="70">
          <template #default="{ row }">
            <el-tag size="small" :type="row.hidden ? 'info' : 'success'">{{ row.hidden ? '隐藏' : '显示' }}</el-tag>
          </template>
        </el-table-column>
      </el-table>
      <el-empty v-else description="暂无菜单" />
    </el-card>

    <el-dialog v-model="formVisible" :title="editId ? '编辑菜单' : '新增菜单'" width="600px" @closed="resetForm">
      <el-form ref="formRef" :model="form" :rules="formRules" label-width="90px">
        <el-form-item label="上级菜单">
          <el-tree-select
            v-model="form.parentId"
            :data="parentOptions"
            :props="{ label: 'name', value: 'id', children: 'children' }"
            check-strictly
            clearable
            placeholder="不选则为顶级"
            style="width:100%"
          />
        </el-form-item>
        <el-form-item label="类型" prop="type">
          <el-radio-group v-model="form.type">
            <el-radio value="M">目录</el-radio>
            <el-radio value="C">菜单</el-radio>
            <el-radio value="F">按钮</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item label="名称" prop="name"><el-input v-model="form.name" /></el-form-item>
        <template v-if="form.type !== 'F'">
          <el-form-item label="路由路径"><el-input v-model="form.path" placeholder="如 system 或 /system" /></el-form-item>
          <el-form-item label="组件路径"><el-input v-model="form.component" placeholder="如 system/admin/index" /></el-form-item>
          <el-form-item label="图标"><el-input v-model="form.icon" placeholder="Element Plus 图标名" /></el-form-item>
          <el-form-item label="排序"><el-input-number v-model="form.sort" :min="0" /></el-form-item>
          <el-form-item label="是否隐藏">
            <el-switch v-model="form.hidden" />
          </el-form-item>
        </template>
        <template v-else>
          <el-form-item label="权限码" prop="perms"><el-input v-model="form.perms" placeholder="如 admin:add" /></el-form-item>
        </template>
        <el-form-item label="端">
          <el-radio-group v-model="form.guardType">
            <el-radio value="admin">admin</el-radio>
            <el-radio value="user">user</el-radio>
          </el-radio-group>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="formVisible = false">取消</el-button>
        <el-button type="primary" :loading="saving" @click="submitForm">确定</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted, computed } from 'vue'
import { ElMessage, ElMessageBox, type FormInstance, type FormRules, type TableInstance } from 'element-plus'
import { getMenuTree, createMenu, updateMenu, deleteMenu, type MenuForm } from '@/api/menu'
import type { MenuItem } from '@/api/auth'

const list = ref<MenuItem[]>([])
const expandAll = ref(true)
const tableRef = ref<TableInstance>()

async function loadList() {
  const res = await getMenuTree('admin')
  list.value = res.data || []
}

function toggleExpand() {
  expandAll.value = !expandAll.value
  // 触发重新渲染展开
  const walk = (rows: MenuItem[]) => {
    rows.forEach((r) => {
      tableRef.value?.toggleRowExpansion(r, expandAll.value)
      if (r.children) walk(r.children)
    })
  }
  walk(list.value)
}

function typeText(t: string) {
  return { M: '目录', C: '菜单', F: '按钮' }[t] || t
}
function typeTag(t: string): any {
  return { M: 'info', C: 'success', F: 'warning' }[t] || 'info'
}

// 新增/编辑
const formVisible = ref(false)
const editId = ref(0)
const saving = ref(false)
const formRef = ref<FormInstance>()
const form = reactive<MenuForm>({ parentId: 0, name: '', type: 'C', path: '', component: '', icon: '', sort: 0, hidden: false, keepAlive: false, guardType: 'admin', perms: '' })
const formRules: FormRules = {
  name: [{ required: true, message: '请输入名称', trigger: 'blur' }],
  type: [{ required: true, message: '请选择类型', trigger: 'change' }]
}

const parentOptions = computed(() => {
  // 仅目录/菜单可作为父级
  const filterTree = (nodes: MenuItem[]): MenuItem[] =>
    nodes
      .filter((n) => n.type !== 'F')
      .map((n) => ({ ...n, children: n.children ? filterTree(n.children) : [] }))
  return filterTree(list.value)
})

function resetForm() {
  editId.value = 0
  Object.assign(form, { parentId: 0, name: '', type: 'C', path: '', component: '', icon: '', sort: 0, hidden: false, keepAlive: false, guardType: 'admin', perms: '' })
  formRef.value?.clearValidate()
}
function openCreate(parentId: number) {
  resetForm()
  form.parentId = parentId
  formVisible.value = true
}
async function openEdit(row: MenuItem) {
  editId.value = row.id
  Object.assign(form, {
    parentId: row.parentId,
    name: row.name,
    type: row.type,
    path: row.path,
    component: row.component,
    icon: row.icon,
    sort: row.sort,
    hidden: row.hidden,
    keepAlive: row.keepAlive,
    guardType: row.guardType,
    perms: row.perms
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
        await updateMenu(editId.value, form)
      } else {
        await createMenu(form)
      }
      ElMessage.success(editId.value ? '更新成功' : '创建成功')
      formVisible.value = false
      loadList()
    } finally {
      saving.value = false
    }
  })
}
async function handleDelete(row: MenuItem) {
  await ElMessageBox.confirm(`确定删除菜单「${row.name}」？`, '提示', { type: 'warning' })
  await deleteMenu(row.id)
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
</style>
