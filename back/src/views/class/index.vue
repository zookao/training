<template>
  <div class="page">
    <el-card shadow="never">
      <div class="toolbar">
        <el-input v-model="query.keyword" placeholder="班级名称" clearable style="width:220px" @keyup.enter="loadList" />
        <el-button type="primary" @click="loadList">查询</el-button>
        <el-button v-permission="'class:add'" type="success" @click="openCreate">新增班级</el-button>
      </div>

      <el-table :data="list" v-loading="loading" border stripe>
        <el-table-column type="index" label="#" width="50" />
        <el-table-column label="封面" width="90" align="center">
          <template #default="{ row }">
            <el-image
              v-if="row.cover"
              :src="authUrl(row.cover)"
              fit="cover"
              style="width: 60px; height: 40px"
              :preview-src-list="[authUrl(row.cover)]"
              preview-teleported
            />
            <span v-else>—</span>
          </template>
        </el-table-column>
        <el-table-column prop="name" label="班级名称" width="180" />
        <el-table-column prop="description" label="描述" min-width="200" show-overflow-tooltip />
        <el-table-column label="课程" width="80">
          <template #default="{ row }">{{ row.courses?.length || 0 }}</template>
        </el-table-column>
        <el-table-column label="学员" width="80">
          <template #default="{ row }">{{ row.users?.length || 0 }}</template>
        </el-table-column>
        <el-table-column prop="sort" label="排序" width="80" />
        <el-table-column label="状态" width="90">
          <template #default="{ row }">
            <el-tag :type="row.status === 1 ? 'success' : 'danger'" size="small">
              {{ row.status === 1 ? '启用' : '禁用' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="创建时间" width="160">
          <template #default="{ row }">
            {{ row.createdAt || '—' }}
          </template>
        </el-table-column>
        <el-table-column label="操作" width="340" fixed="right">
          <template #default="{ row }">
            <el-button v-permission="'class:edit'" link type="primary" @click="openEdit(row)">编辑</el-button>
            <el-button v-permission="'class:edit'" link type="primary" @click="openCourses(row)">分配课程</el-button>
            <el-button v-permission="'class:edit'" link type="primary" @click="openUsers(row)">分配学员</el-button>
            <el-button v-permission="'class:report'" link type="primary" @click="openReport(row)">学习报告</el-button>
            <el-button v-permission="'class:delete'" link type="danger" @click="handleDelete(row)">删除</el-button>
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
    <el-dialog v-model="formVisible" :title="editId ? '编辑班级' : '新增班级'" width="520px" @closed="resetForm">
      <el-form ref="formRef" :model="form" :rules="formRules" label-width="80px">
        <el-form-item label="名称" prop="name">
          <el-input v-model="form.name" placeholder="班级名称" />
        </el-form-item>
        <el-form-item label="封面图">
          <el-upload
            :show-file-list="false"
            :before-upload="beforeCoverUpload"
            :http-request="handleCoverUpload"
            accept="image/*"
          >
            <div v-if="form.cover" class="cover-preview">
              <el-image :src="authUrl(form.cover)" fit="cover" style="width: 160px; height: 90px" />
              <span class="cover-replace">点击替换</span>
            </div>
            <el-button v-else :loading="coverUploading">+ 选择封面图</el-button>
          </el-upload>
          <el-button v-if="form.cover" link type="danger" style="margin-left: 12px" @click="form.cover = ''">移除</el-button>
          <div class="cover-tip">建议尺寸 640×360（16:9），支持 JPG/PNG，不超过 5MB</div>
        </el-form-item>
        <el-form-item label="描述">
          <el-input v-model="form.description" type="textarea" :rows="3" />
        </el-form-item>
        <el-form-item label="排序">
          <el-input-number v-model="form.sort" :min="0" />
        </el-form-item>
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

    <!-- 分配课程（服务端搜索 + 分页，跨页选择） -->
    <el-dialog v-model="coursesVisible" title="分配课程" width="640px" @closed="onCoursesClosed">
      <div class="assign-toolbar">
        <el-input v-model="courseSearch.keyword" placeholder="搜索课程标题" clearable style="width:220px" @keyup.enter="searchCourses" />
        <el-button type="primary" @click="searchCourses">查询</el-button>
        <span class="selected-tip">已选 {{ assignCourseIds.length }} 个</span>
      </div>
      <el-table :data="courses" border size="small" row-key="id">
        <el-table-column width="50" align="center">
          <template #default="{ row }">
            <el-checkbox
              :model-value="assignCourseIds.includes(row.id)"
              :disabled="row.status !== 1"
              @change="(val: boolean) => toggleCourse(row.id, val)"
            />
          </template>
        </el-table-column>
        <el-table-column prop="title" label="课程标题" min-width="160" />
        <el-table-column prop="description" label="描述" min-width="180" show-overflow-tooltip />
        <el-table-column label="状态" width="80" align="center">
          <template #default="{ row }">
            <el-tag :type="row.status === 1 ? 'success' : 'info'" size="small">
              {{ row.status === 1 ? '启用' : '禁用' }}
            </el-tag>
          </template>
        </el-table-column>
      </el-table>
      <el-pagination
        class="pager"
        v-model:current-page="courseSearch.page"
        v-model:page-size="courseSearch.pageSize"
        :total="courseTotal"
        :page-sizes="[10, 20, 50]"
        layout="total, sizes, prev, pager, next"
        small
        @size-change="loadCourses"
        @current-change="loadCourses"
      />
      <template #footer>
        <el-button @click="coursesVisible = false">取消</el-button>
        <el-button type="primary" :loading="saving" @click="submitCourses">确定</el-button>
      </template>
    </el-dialog>

    <!-- 分配学员（服务端搜索 + 分页，跨页选择） -->
    <el-dialog v-model="usersVisible" title="分配学员" width="760px" @closed="onUsersClosed">
      <div class="assign-toolbar">
        <el-input v-model="userSearch.keyword" placeholder="搜索用户名/姓名/手机" clearable style="width:200px" @keyup.enter="searchUsers" />
        <el-select v-model="userSearch.departmentId" clearable placeholder="院系筛选" style="width:160px" @change="searchUsers">
          <el-option v-for="d in allDepartments" :key="d.id" :label="d.name" :value="d.id" />
        </el-select>
        <el-button type="primary" @click="searchUsers">查询</el-button>
        <span class="selected-tip">已选 {{ assignUserIds.length }} 个</span>
      </div>
      <el-table :data="users" border size="small" row-key="id">
        <el-table-column width="50" align="center">
          <template #header>
            <el-checkbox
              :model-value="isAllSelected"
              :indeterminate="isIndeterminate"
              @change="toggleAllPage"
            />
          </template>
          <template #default="{ row }">
            <el-checkbox
              :model-value="assignUserIds.includes(row.id)"
              :disabled="row.status !== 1"
              @change="(val: boolean) => toggleUser(row.id, val)"
            />
          </template>
        </el-table-column>
        <el-table-column prop="username" label="用户名" width="130" />
        <el-table-column prop="nickname" label="姓名" width="100" />
        <el-table-column prop="studentNo" label="学号" width="110" />
        <el-table-column prop="phone" label="手机" width="130" />
        <el-table-column label="院系" width="120">
          <template #default="{ row }">{{ deptName(row.departmentId) }}</template>
        </el-table-column>
        <el-table-column label="状态" width="80" align="center">
          <template #default="{ row }">
            <el-tag :type="row.status === 1 ? 'success' : 'info'" size="small">
              {{ row.status === 1 ? '启用' : '禁用' }}
            </el-tag>
          </template>
        </el-table-column>
      </el-table>
      <el-pagination
        class="pager"
        v-model:current-page="userSearch.page"
        v-model:page-size="userSearch.pageSize"
        :total="userTotal"
        :page-sizes="[10, 20, 50]"
        layout="total, sizes, prev, pager, next"
        small
        @size-change="loadUsers"
        @current-change="loadUsers"
      />
      <template #footer>
        <el-button @click="usersVisible = false">取消</el-button>
        <el-button type="primary" :loading="saving" @click="submitUsers">确定</el-button>
      </template>
    </el-dialog>

    <!-- 学习报告 -->
    <el-dialog v-model="reportVisible" :title="`班级学习报告 - ${reportData?.className || ''}`" width="820px" @closed="reportData = null">
      <el-table :data="reportData?.students || []" v-loading="reportLoading" border stripe>
        <el-table-column type="index" label="#" width="50" />
        <el-table-column prop="username" label="用户名" width="140" />
        <el-table-column prop="nickname" label="姓名" width="120" />
        <el-table-column prop="studentNo" label="学号" width="120" />
        <el-table-column label="学习完成度" min-width="200">
          <template #default="{ row }">
            <el-progress :percentage="row.percent" :stroke-width="14" :text-inside="true" :status="row.percent >= 100 ? 'success' : ''" />
          </template>
        </el-table-column>
        <el-table-column label="操作" width="110" fixed="right">
          <template #default="{ row }">
            <el-button link type="primary" @click="openStudentDetail(row.userId)">查看详情</el-button>
          </template>
        </el-table-column>
      </el-table>
      <template #footer>
        <el-button @click="reportVisible = false">关闭</el-button>
      </template>
    </el-dialog>

    <!-- 学员学习详情 -->
    <el-dialog v-model="detailVisible" :title="`学员学习详情 - ${detailData?.username || ''}`" width="920px" @closed="detailData = null">
      <el-table :data="detailData?.courses || []" v-loading="detailLoading" border stripe row-key="courseId">
        <el-table-column type="expand">
          <template #default="{ row }">
            <div style="padding: 8px 16px">
              <el-table :data="row.videos" border size="small">
                <el-table-column prop="title" label="视频名称" min-width="180" show-overflow-tooltip />
                <el-table-column label="视频学习进度" min-width="240">
                  <template #default="{ row: v }">
                    <el-progress :percentage="v.percent" :stroke-width="12" :text-inside="true" :status="v.completed ? 'success' : ''" style="width: 160px; display: inline-flex; vertical-align: middle" />
                    <span style="margin-left: 8px; font-size: 12px; color: #909399">{{ v.completed ? '已完成' : '学习中' }}</span>
                  </template>
                </el-table-column>
              </el-table>
            </div>
          </template>
        </el-table-column>
        <el-table-column prop="title" label="课程名称" min-width="180" show-overflow-tooltip />
        <el-table-column label="课程学习进度" min-width="260">
          <template #default="{ row }">
            <el-progress :percentage="row.percent" :stroke-width="14" :text-inside="true" :status="row.percent >= 100 ? 'success' : ''" style="width: 180px; display: inline-flex; vertical-align: middle" />
            <span style="margin-left: 8px; font-size: 12px; color: #909399">{{ row.completedVideos }}/{{ row.videoCount }} 视频</span>
          </template>
        </el-table-column>
      </el-table>
      <template #footer>
        <el-button @click="detailVisible = false">关闭</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted } from 'vue'
import { ElMessage, ElMessageBox, type FormInstance, type FormRules } from 'element-plus'
import {
  getClassList, getClassDetail, createClass, updateClass, deleteClass,
  getClassCourseIds, getClassUserIds, assignClassCourses, assignClassUsers,
  getClassLearningReport, getStudentLearningDetail,
  type ClassItem, type ClassForm, type ClassLearningReport, type StudentLearningDetail
} from '@/api/class'
import { getCourseList, uploadImage, type CourseItem } from '@/api/course'
import { getUserList, type UserItem } from '@/api/user'
import { getAllDepartments, type DepartmentItem } from '@/api/department'
import { authUrl } from '@/utils/authUrl'

const loading = ref(false)
const list = ref<ClassItem[]>([])
const total = ref(0)
const query = reactive({ page: 1, pageSize: 10, keyword: '' })

async function loadList() {
  loading.value = true
  try {
    const res = await getClassList(query)
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
const coverUploading = ref(false)
const form = reactive<ClassForm>({ name: '', cover: '', description: '', sort: 0, status: 1 })
const formRules: FormRules = {
  name: [{ required: true, message: '请输入班级名称', trigger: 'blur' }]
}

function resetForm() {
  editId.value = 0
  Object.assign(form, { name: '', cover: '', description: '', sort: 0, status: 1 })
  formRef.value?.clearValidate()
}

function openCreate() {
  resetForm()
  formVisible.value = true
}

async function openEdit(row: ClassItem) {
  editId.value = row.id
  const res = await getClassDetail(row.id)
  Object.assign(form, {
    name: res.data.name,
    cover: res.data.cover || '',
    description: res.data.description,
    sort: res.data.sort,
    status: res.data.status
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
        await updateClass(editId.value, form)
      } else {
        await createClass(form)
      }
      ElMessage.success(editId.value ? '更新成功' : '创建成功')
      formVisible.value = false
      loadList()
    } finally {
      saving.value = false
    }
  })
}

async function handleDelete(row: ClassItem) {
  await ElMessageBox.confirm(`确定删除班级「${row.name}」？`, '提示', { type: 'warning' })
  await deleteClass(row.id)
  ElMessage.success('删除成功')
  loadList()
}

// 封面图上传
function beforeCoverUpload(file: File) {
  if (!file.type.startsWith('image/')) {
    ElMessage.error('请选择图片文件')
    return false
  }
  if (file.size > 5 * 1024 * 1024) {
    ElMessage.error('封面图不能超过 5MB')
    return false
  }
  return true
}
async function handleCoverUpload(option: { file: File }) {
  coverUploading.value = true
  try {
    const res = await uploadImage(option.file)
    form.cover = res.data.url
    ElMessage.success('封面上传成功')
  } catch (e) {
    /* handled in interceptor */
  } finally {
    coverUploading.value = false
  }
}

// 分配课程（服务端搜索 + 分页 + 跨页选择）
const coursesVisible = ref(false)
const courses = ref<CourseItem[]>([])
const courseTotal = ref(0)
const assignCourseIds = ref<number[]>([])
const courseSearch = reactive({ keyword: '', page: 1, pageSize: 10 })
let currentId = 0

async function loadCourses() {
  const res = await getCourseList({ page: courseSearch.page, pageSize: courseSearch.pageSize, keyword: courseSearch.keyword })
  courses.value = res.data.list || []
  courseTotal.value = res.data.total || 0
}
function searchCourses() {
  courseSearch.page = 1
  loadCourses()
}
function toggleCourse(id: number, checked: boolean) {
  const idx = assignCourseIds.value.indexOf(id)
  if (checked && idx === -1) assignCourseIds.value.push(id)
  if (!checked && idx !== -1) assignCourseIds.value.splice(idx, 1)
}
function onCoursesClosed() {
  courseSearch.keyword = ''
  courseSearch.page = 1
  courses.value = []
  courseTotal.value = 0
}

async function openCourses(row: ClassItem) {
  currentId = row.id
  courseSearch.keyword = ''
  courseSearch.page = 1
  await loadCourses()
  const ids = await getClassCourseIds(row.id)
  assignCourseIds.value = ids.data || []
  coursesVisible.value = true
}

async function submitCourses() {
  saving.value = true
  try {
    await assignClassCourses(currentId, assignCourseIds.value)
    ElMessage.success('分配成功')
    coursesVisible.value = false
    loadList()
  } finally {
    saving.value = false
  }
}

// 分配学员（服务端搜索 + 分页 + 跨页选择）
const usersVisible = ref(false)
const users = ref<UserItem[]>([])
const userTotal = ref(0)
const assignUserIds = ref<number[]>([])
const allDepartments = ref<DepartmentItem[]>([])
const userSearch = reactive({ keyword: '', departmentId: undefined as number | undefined, page: 1, pageSize: 10 })

async function loadUsers() {
  const res = await getUserList({ page: userSearch.page, pageSize: userSearch.pageSize, keyword: userSearch.keyword, departmentId: userSearch.departmentId })
  users.value = res.data.list || []
  userTotal.value = res.data.total || 0
}
function searchUsers() {
  userSearch.page = 1
  loadUsers()
}
function toggleUser(id: number, checked: boolean) {
  const idx = assignUserIds.value.indexOf(id)
  if (checked && idx === -1) assignUserIds.value.push(id)
  if (!checked && idx !== -1) assignUserIds.value.splice(idx, 1)
}
// 当前页启用的学员ID（全选基于当前页）
const pageEnabledIds = computed(() => users.value.filter((u) => u.status === 1).map((u) => u.id))
const isAllSelected = computed(() => {
  if (pageEnabledIds.value.length === 0) return false
  return pageEnabledIds.value.every((id) => assignUserIds.value.includes(id))
})
const isIndeterminate = computed(() => {
  const selected = pageEnabledIds.value.filter((id) => assignUserIds.value.includes(id))
  return selected.length > 0 && selected.length < pageEnabledIds.value.length
})
function toggleAllPage(val: boolean) {
  if (val) {
    const set = new Set(assignUserIds.value)
    pageEnabledIds.value.forEach((id) => set.add(id))
    assignUserIds.value = [...set]
  } else {
    const removeSet = new Set(pageEnabledIds.value)
    assignUserIds.value = assignUserIds.value.filter((id) => !removeSet.has(id))
  }
}
function deptName(id: number) {
  return allDepartments.value.find((d) => d.id === id)?.name || '-'
}
function onUsersClosed() {
  userSearch.keyword = ''
  userSearch.departmentId = undefined
  userSearch.page = 1
  users.value = []
  userTotal.value = 0
}

async function openUsers(row: ClassItem) {
  currentId = row.id
  userSearch.keyword = ''
  userSearch.departmentId = undefined
  userSearch.page = 1
  const [deptRes] = await Promise.all([getAllDepartments(), loadUsers()])
  allDepartments.value = deptRes.data || []
  const ids = await getClassUserIds(row.id)
  assignUserIds.value = ids.data || []
  usersVisible.value = true
}

async function submitUsers() {
  saving.value = true
  try {
    await assignClassUsers(currentId, assignUserIds.value)
    ElMessage.success('分配成功')
    usersVisible.value = false
    loadList()
  } finally {
    saving.value = false
  }
}

// 学习报告
const reportVisible = ref(false)
const reportLoading = ref(false)
const reportData = ref<ClassLearningReport | null>(null)

async function openReport(row: ClassItem) {
  reportVisible.value = true
  reportLoading.value = true
  reportData.value = null
  try {
    const res = await getClassLearningReport(row.id)
    reportData.value = res.data
  } finally {
    reportLoading.value = false
  }
}

// 学员学习详情
const detailVisible = ref(false)
const detailLoading = ref(false)
const detailData = ref<StudentLearningDetail | null>(null)

async function openStudentDetail(userId: number) {
  if (!reportData.value) return
  detailVisible.value = true
  detailLoading.value = true
  detailData.value = null
  try {
    const res = await getStudentLearningDetail(reportData.value.classId, userId)
    detailData.value = res.data
  } finally {
    detailLoading.value = false
  }
}

onMounted(() => {
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
.assign-toolbar {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 12px;
}
.selected-tip {
  margin-left: auto;
  font-size: 13px;
  color: #409eff;
}
.cover-tip {
  width: 100%;
  margin-top: 6px;
  font-size: 12px;
  color: #909399;
}
.cover-preview {
  position: relative;
  width: 160px;
  height: 90px;
  border-radius: 4px;
  overflow: hidden;
  cursor: pointer;
}
.cover-replace {
  position: absolute;
  inset: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  background: rgba(0, 0, 0, 0.5);
  color: #fff;
  font-size: 13px;
  opacity: 0;
  transition: opacity 0.2s;
}
.cover-preview:hover .cover-replace {
  opacity: 1;
}
</style>
