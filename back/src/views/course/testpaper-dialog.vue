<template>
  <el-dialog :model-value="visible" :title="`试卷管理 - ${courseTitle}`" width="960px" @update:model-value="$emit('update:visible', $event)" @open="loadList">
    <el-alert title="一个课程只能同时启用一个试卷，启用新试卷将自动禁用该课程下其他已启用的试卷" type="warning" :closable="false" show-icon style="margin-bottom: 12px" />
    <div class="toolbar">
      <el-input v-model="query.keyword" placeholder="搜索试卷名称" clearable style="width:200px" @keyup.enter="loadList" />
      <el-button type="primary" @click="loadList">查询</el-button>
      <el-button type="success" @click="openCreate">新增试卷</el-button>
    </div>
    <el-table :data="list" v-loading="loading" border stripe size="small">
      <el-table-column type="index" label="#" width="50" />
      <el-table-column prop="name" label="试卷名称" min-width="160" />
      <el-table-column label="类型" width="120">
        <template #default="{ row }">
          <el-tag :type="row.type === 1 ? 'primary' : 'warning'" size="small">{{ row.type === 1 ? '随时考试' : '课程完成后考试' }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="totalScore" label="总分" width="70" />
      <el-table-column prop="passScore" label="及格分" width="70" />
      <el-table-column label="时长" width="70">
        <template #default="{ row }">{{ row.duration }}分钟</template>
      </el-table-column>
      <el-table-column label="状态" width="60">
        <template #default="{ row }">
          <el-tag :type="row.status === 1 ? 'success' : 'danger'" size="small">{{ row.status === 1 ? '启用' : '禁用' }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column label="操作" width="200" fixed="right">
        <template #default="{ row }">
          <el-button link type="primary" @click="openQuestions(row)">组卷</el-button>
          <el-button link type="primary" @click="openEdit(row)">编辑</el-button>
          <el-button link type="danger" @click="handleDelete(row)">删除</el-button>
        </template>
      </el-table-column>
    </el-table>
    <el-pagination class="pager" v-model:current-page="query.page" v-model:page-size="query.pageSize" :total="total" :page-sizes="[10,20,50]" layout="total,sizes,prev,pager,next" @size-change="loadList" @current-change="loadList" />

    <!-- 新增/编辑试卷 -->
    <el-dialog v-model="formVisible" :title="editId ? '编辑试卷' : '新增试卷'" width="560px" append-to-body>
      <el-form ref="formRef" :model="form" :rules="formRules" label-width="80px">
        <el-form-item label="名称" prop="name"><el-input v-model="form.name" /></el-form-item>
        <el-form-item label="描述"><el-input v-model="form.description" type="textarea" :rows="2" /></el-form-item>
        <el-form-item label="类型">
          <el-radio-group v-model="form.type">
            <el-radio :value="2">课程完成后考试</el-radio>
            <el-radio :value="1">随时考试</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item label="总分"><el-input-number v-model="form.totalScore" :min="1" disabled /> <span style="color:#909399">固定100分</span></el-form-item>
        <el-form-item label="及格分"><el-input-number v-model="form.passScore" :min="1" :max="100" /></el-form-item>
        <el-form-item label="时长(分)"><el-input-number v-model="form.duration" :min="1" :max="300" /></el-form-item>
        <el-form-item label="排序"><el-input-number v-model="form.sort" :min="0" /></el-form-item>
        <el-form-item label="状态">
          <el-radio-group v-model="form.status">
            <el-radio :value="1">启用</el-radio>
            <el-radio :value="0">禁用</el-radio>
          </el-radio-group>
          <span v-if="form.status === 1" style="color:#e6a23c;font-size:12px;margin-left:8px">启用后将自动禁用该课程下其他已启用的试卷</span>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="formVisible = false">取消</el-button>
        <el-button type="primary" :loading="saving" @click="submitForm">确定</el-button>
      </template>
    </el-dialog>

    <!-- 组卷 -->
    <el-dialog v-model="questionsVisible" :title="`组卷 - ${currentTpName}`" width="800px" append-to-body @opened="onQuestionsOpened">
      <div class="toolbar">
        <span style="color:#909399">从课程试题中选择题目组卷，总分需等于100</span>
        <el-button size="small" @click="avgScores">平均分配分值</el-button>
        <span style="margin-left:auto;color:#409eff">当前总分：{{ totalSelScore }}</span>
      </div>
      <el-table ref="questionTableRef" :data="allQuestions" border size="small" max-height="450" @selection-change="onSelChange">
        <el-table-column type="selection" width="45" />
        <el-table-column label="题型" width="70">
          <template #default="{ row }"><el-tag size="small">{{ typeText(row.type) }}</el-tag></template>
        </el-table-column>
        <el-table-column prop="title" label="题干" min-width="200" show-overflow-tooltip />
        <el-table-column label="分值" width="100">
          <template #default="{ row }">
            <el-input-number v-model="scoreMap[row.id]" :min="0" :max="100" size="small" controls-position="right" style="width:90px" />
          </template>
        </el-table-column>
      </el-table>
      <template #footer>
        <el-button @click="questionsVisible = false">取消</el-button>
        <el-button type="primary" :loading="saving" @click="submitQuestions">保存组卷</el-button>
      </template>
    </el-dialog>
  </el-dialog>
</template>

<script setup lang="ts">
import { ref, reactive, computed, watch, nextTick } from 'vue'
import { ElMessage, ElMessageBox, type FormInstance, type FormRules } from 'element-plus'
import { getTestpaperList, createTestpaper, updateTestpaper, deleteTestpaper, getTestpaperQuestions, setTestpaperQuestions, type TestpaperItem, type TestpaperForm, type TestpaperQuestionReq } from '@/api/testpaper'
import { getAllQuestions, type QuestionItem } from '@/api/question'

const props = defineProps<{ visible: boolean; courseId: number; courseTitle?: string }>()
defineEmits<{ 'update:visible': [val: boolean] }>()

const loading = ref(false)
const list = ref<TestpaperItem[]>([])
const total = ref(0)
const query = reactive({ page: 1, pageSize: 10, keyword: '' })

async function loadList() {
  if (!props.courseId) return
  loading.value = true
  try {
    const res = await getTestpaperList(props.courseId, query)
    list.value = res.data.list
    total.value = res.data.total
  } finally {
    loading.value = false
  }
}

function typeText(t: number) { return t === 1 ? '单选' : t === 2 ? '多选' : '判断' }

// 新增/编辑
const formVisible = ref(false)
const editId = ref(0)
const saving = ref(false)
const formRef = ref<FormInstance>()
const form = reactive<TestpaperForm>({ name: '', description: '', type: 2, totalScore: 100, passScore: 60, duration: 60, sort: 0, status: 1 })
const formRules: FormRules = { name: [{ required: true, message: '请输入名称', trigger: 'blur' }] }

function resetForm() {
  editId.value = 0
  Object.assign(form, { name: '', description: '', type: 2, totalScore: 100, passScore: 60, duration: 60, sort: 0, status: 1 })
  formRef.value?.clearValidate()
}
function openCreate() { resetForm(); formVisible.value = true }
function openEdit(row: TestpaperItem) {
  editId.value = row.id
  Object.assign(form, { name: row.name, description: row.description, type: row.type, totalScore: row.totalScore, passScore: row.passScore, duration: row.duration, sort: row.sort, status: row.status })
  formVisible.value = true
}
async function submitForm() {
  if (!formRef.value) return
  await formRef.value.validate(async (valid) => {
    if (!valid) return
    saving.value = true
    try {
      if (editId.value) await updateTestpaper(editId.value, form)
      else await createTestpaper(props.courseId, form)
      ElMessage.success(editId.value ? '更新成功' : '创建成功')
      formVisible.value = false
      loadList()
    } finally { saving.value = false }
  })
}
async function handleDelete(row: TestpaperItem) {
  await ElMessageBox.confirm('删除试卷将同时删除所有考试记录，确定删除？', '提示', { type: 'warning' })
  await deleteTestpaper(row.id)
  ElMessage.success('删除成功')
  loadList()
}

// 组卷
const questionsVisible = ref(false)
const currentTpId = ref(0)
const currentTpName = ref('')
const allQuestions = ref<QuestionItem[]>([])
const selectedQuestions = ref<QuestionItem[]>([])
const scoreMap = ref<Record<number, number>>({})
const questionTableRef = ref()
const initialSelIds = ref<number[]>([])

const totalSelScore = computed(() => selectedQuestions.value.reduce((sum, q) => sum + (scoreMap.value[q.id] || 0), 0))

async function openQuestions(row: TestpaperItem) {
  currentTpId.value = row.id
  currentTpName.value = row.name
  const [qRes, tpqRes] = await Promise.all([getAllQuestions(props.courseId), getTestpaperQuestions(row.id)])
  allQuestions.value = qRes.data || []
  scoreMap.value = {}
  const selIds: number[] = []
  for (const item of (tpqRes.data || [])) {
    selIds.push(item.questionId)
    scoreMap.value[item.questionId] = item.score
  }
  selectedQuestions.value = allQuestions.value.filter(q => selIds.includes(q.id))
  initialSelIds.value = selIds
  questionsVisible.value = true
}

// 弹窗打开后再回显勾选：此时表格已渲染，调用 toggleRowSelection 才生效
function onQuestionsOpened() {
  nextTick(() => {
    const table = questionTableRef.value
    if (!table) return
    table.clearSelection()
    allQuestions.value.forEach(q => {
      if (initialSelIds.value.includes(q.id)) {
        table.toggleRowSelection(q, true)
      }
    })
  })
}

function onSelChange(rows: QuestionItem[]) {
  selectedQuestions.value = rows
  rows.forEach(r => { if (scoreMap.value[r.id] === undefined) scoreMap.value[r.id] = 0 })
}

function avgScores() {
  if (selectedQuestions.value.length === 0) return
  const avg = Math.floor(100 / selectedQuestions.value.length)
  const remainder = 100 - avg * selectedQuestions.value.length
  // 余数均匀分散到前 remainder 道题，避免全堆第一题造成极端不均
  selectedQuestions.value.forEach((q, i) => {
    scoreMap.value[q.id] = avg + (i < remainder ? 1 : 0)
  })
}

async function submitQuestions() {
  if (totalSelScore.value !== 100) {
    ElMessage.warning(`当前总分 ${totalSelScore.value}，需等于 100`)
    return
  }
  const items: TestpaperQuestionReq[] = selectedQuestions.value.map(q => ({ questionId: q.id, score: scoreMap.value[q.id] || 0 }))
  saving.value = true
  try {
    await setTestpaperQuestions(currentTpId.value, items)
    ElMessage.success('组卷成功')
    questionsVisible.value = false
  } finally { saving.value = false }
}

watch(() => props.visible, (v) => { if (v) loadList() })
</script>

<style scoped>
.toolbar { display: flex; gap: 8px; align-items: center; margin-bottom: 12px; }
.pager { margin-top: 12px; justify-content: flex-end; }
</style>
