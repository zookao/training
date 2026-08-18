<template>
  <el-dialog :model-value="visible" :title="`试题管理 - ${courseTitle}`" width="900px" @update:model-value="$emit('update:visible', $event)" @open="loadList">
    <div class="toolbar">
      <el-input v-model="query.keyword" placeholder="搜索题干" clearable style="width:200px" @keyup.enter="loadList" />
      <el-button type="primary" @click="loadList">查询</el-button>
      <el-button type="success" @click="openCreate">新增试题</el-button>
      <el-button type="warning" plain @click="openImport">批量导入</el-button>
    </div>
    <el-table :data="list" v-loading="loading" border stripe size="small">
      <el-table-column type="index" label="#" width="50" />
      <el-table-column label="题型" width="80">
        <template #default="{ row }">
          <el-tag :type="typeTag(row.type)" size="small">{{ typeText(row.type) }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="title" label="题干" min-width="200" show-overflow-tooltip />
      <el-table-column label="答案" width="80">
        <template #default="{ row }">{{ formatAnswer(row.answer) }}</template>
      </el-table-column>
      <el-table-column prop="sort" label="排序" width="60" />
      <el-table-column label="状态" width="60">
        <template #default="{ row }">
          <el-tag :type="row.status === 1 ? 'success' : 'danger'" size="small">{{ row.status === 1 ? '启用' : '禁用' }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column label="操作" width="130" fixed="right">
        <template #default="{ row }">
          <el-button link type="primary" @click="openEdit(row)">编辑</el-button>
          <el-button link type="danger" @click="handleDelete(row)">删除</el-button>
        </template>
      </el-table-column>
    </el-table>
    <el-pagination class="pager" v-model:current-page="query.page" v-model:page-size="query.pageSize" :total="total" :page-sizes="[10,20,50]" layout="total,sizes,prev,pager,next" @size-change="loadList" @current-change="loadList" />

    <!-- 新增/编辑 -->
    <el-dialog v-model="formVisible" :title="editId ? '编辑试题' : '新增试题'" width="700px" append-to-body>
      <el-form ref="formRef" :model="form" :rules="formRules" label-width="80px">
        <el-form-item label="题型" prop="type">
          <el-radio-group v-model="form.type" @change="onTypeChange">
            <el-radio :value="1">单选</el-radio>
            <el-radio :value="2">多选</el-radio>
            <el-radio :value="3">判断</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item label="题干" prop="title">
          <el-input v-model="form.title" type="textarea" :rows="3" placeholder="请输入题干" />
        </el-form-item>
        <el-form-item label="选项" v-if="form.type !== 3">
          <div v-for="(opt, idx) in options" :key="idx" class="option-row">
            <el-input v-model="opt.label" style="width:60px" placeholder="A" />
            <el-input v-model="opt.content" style="flex:1" placeholder="选项内容" />
            <el-button v-if="options.length > 2" link type="danger" @click="removeOption(idx)">删除</el-button>
          </div>
          <el-button size="small" @click="addOption">+ 添加选项</el-button>
        </el-form-item>
        <el-form-item label="答案" prop="answerArr">
          <template v-if="form.type === 1 || form.type === 3">
            <el-radio-group v-model="answerArr[0]">
              <el-radio v-for="opt in answerOptions" :key="opt.label" :value="opt.label">{{ opt.label }}. {{ opt.content }}</el-radio>
            </el-radio-group>
          </template>
          <template v-else>
            <el-checkbox-group v-model="answerArr">
              <el-checkbox v-for="opt in answerOptions" :key="opt.label" :value="opt.label">{{ opt.label }}. {{ opt.content }}</el-checkbox>
            </el-checkbox-group>
          </template>
        </el-form-item>
        <el-form-item label="解析"><el-input v-model="form.analysis" type="textarea" :rows="2" /></el-form-item>
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

    <!-- 批量导入 -->
    <el-dialog v-model="importVisible" title="批量导入试题" width="800px" append-to-body @closed="resetImport">
      <el-alert
        type="info"
        :closable="false"
        show-icon
        title="Excel 字段：题型、题干、选项A、选项B、选项C、选项D、正确答案、解析"
        description="规则：题型/题干/答案必填；单选多选至少 2 个选项，答案填字母（如 A 或 AC）；判断题选项默认为 A正确、B错误（模板已填），答案填 A 或 B。请先下载模板填写。"
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
        <el-table-column prop="type" label="题型" width="80" />
        <el-table-column prop="title" label="题干" min-width="180" show-overflow-tooltip />
        <el-table-column prop="answer" label="答案" width="80" />
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
  </el-dialog>
</template>

<script setup lang="ts">
import { ref, reactive, computed, watch } from 'vue'
import { ElMessage, ElMessageBox, type FormInstance, type FormRules } from 'element-plus'
import { getQuestionList, createQuestion, updateQuestion, deleteQuestion, importQuestions, downloadQuestionImportTemplate, type QuestionItem, type QuestionOption, type QuestionImportResult } from '@/api/question'

const props = defineProps<{ visible: boolean; courseId: number; courseTitle?: string }>()
defineEmits<{ 'update:visible': [val: boolean] }>()

const loading = ref(false)
const list = ref<QuestionItem[]>([])
const total = ref(0)
const query = reactive({ page: 1, pageSize: 10, keyword: '' })

async function loadList() {
  if (!props.courseId) return
  loading.value = true
  try {
    const res = await getQuestionList(props.courseId, query)
    list.value = res.data.list
    total.value = res.data.total
  } finally {
    loading.value = false
  }
}

function typeText(t: number) { return t === 1 ? '单选' : t === 2 ? '多选' : '判断' }
function typeTag(t: number) { return t === 1 ? 'primary' : t === 2 ? 'warning' : 'success' }
function formatAnswer(ans: string) {
  try { return JSON.parse(ans).join(',') } catch { return ans }
}

// 新增/编辑
const formVisible = ref(false)
const editId = ref(0)
const saving = ref(false)
const formRef = ref<FormInstance>()
const form = reactive({ type: 1, title: '', analysis: '', sort: 0, status: 1 })
const options = ref<QuestionOption[]>([])
const answerArr = ref<string[]>([])
const formRules: FormRules = {
  type: [{ required: true, message: '请选择题型', trigger: 'change' }],
  title: [{ required: true, message: '请输入题干', trigger: 'blur' }]
}

const answerOptions = computed(() => {
  if (form.type === 3) return [{ label: 'A', content: '正确' }, { label: 'B', content: '错误' }]
  return options.value
})

function onTypeChange() {
  answerArr.value = []
  if (form.type === 3) {
    options.value = [{ label: 'A', content: '正确' }, { label: 'B', content: '错误' }]
  } else if (options.value.length < 2) {
    options.value = [
      { label: 'A', content: '' },
      { label: 'B', content: '' }
    ]
  }
}

function addOption() {
  const labels = 'ABCDEFGH'
  options.value.push({ label: labels[options.value.length] || '', content: '' })
}
function removeOption(idx: number) {
  options.value.splice(idx, 1)
  answerArr.value = answerArr.value.filter(a => options.value.some(o => o.label === a))
}

function resetForm() {
  editId.value = 0
  Object.assign(form, { type: 1, title: '', analysis: '', sort: 0, status: 1 })
  options.value = [{ label: 'A', content: '' }, { label: 'B', content: '' }]
  answerArr.value = []
  formRef.value?.clearValidate()
}

function openCreate() {
  resetForm()
  formVisible.value = true
}

function openEdit(row: QuestionItem) {
  editId.value = row.id
  Object.assign(form, { type: row.type, title: row.title, analysis: row.analysis, sort: row.sort, status: row.status })
  try { options.value = JSON.parse(row.options) } catch { options.value = [] }
  try { answerArr.value = JSON.parse(row.answer) } catch { answerArr.value = [] }
  if (form.type === 3) {
    options.value = [{ label: 'A', content: '正确' }, { label: 'B', content: '错误' }]
  }
  formVisible.value = true
}

async function submitForm() {
  if (!formRef.value) return
  if (form.type !== 3 && options.value.some(o => !o.content)) {
    ElMessage.warning('请填写所有选项内容')
    return
  }
  if (answerArr.value.length === 0 || (form.type !== 2 && !answerArr.value[0])) {
    ElMessage.warning('请选择正确答案')
    return
  }
  await formRef.value.validate(async (valid) => {
    if (!valid) return
    saving.value = true
    try {
      const opts = form.type === 3
        ? [{ label: 'A', content: '正确' }, { label: 'B', content: '错误' }]
        : options.value
      const data = {
        type: form.type,
        title: form.title,
        options: JSON.stringify(opts),
        answer: JSON.stringify(form.type === 2 ? answerArr.value : [answerArr.value[0]]),
        analysis: form.analysis,
        sort: form.sort,
        status: form.status
      }
      if (editId.value) {
        await updateQuestion(editId.value, data)
      } else {
        await createQuestion(props.courseId, data)
      }
      ElMessage.success(editId.value ? '更新成功' : '创建成功')
      formVisible.value = false
      loadList()
    } finally {
      saving.value = false
    }
  })
}

async function handleDelete(row: QuestionItem) {
  await ElMessageBox.confirm('确定删除该试题？已组卷的试卷将自动移除该题。', '提示', { type: 'warning' })
  await deleteQuestion(row.id)
  ElMessage.success('删除成功')
  loadList()
}

// 批量导入
const importVisible = ref(false)
const importing = ref(false)
const importInputRef = ref<HTMLInputElement | null>(null)
const importFile = ref<File | null>(null)
const importResult = ref<QuestionImportResult | null>(null)

function openImport() {
  importFile.value = null
  importResult.value = null
  importVisible.value = true
}
function resetImport() {
  importFile.value = null
  importResult.value = null
  if (importInputRef.value) importInputRef.value.value = ''
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
    await downloadQuestionImportTemplate()
    ElMessage.success('模板已下载')
  } catch {
    /* handled in interceptor */
  }
}
async function submitImport() {
  if (!importFile.value) return
  importing.value = true
  try {
    const res = await importQuestions(props.courseId, importFile.value)
    importResult.value = res.data
    if (res.data.success > 0) {
      ElMessage.success(`导入成功 ${res.data.success} 条`)
      loadList()
    } else {
      ElMessage.warning('未导入成功任何记录')
    }
  } catch {
    /* handled in interceptor */
  } finally {
    importing.value = false
    if (importInputRef.value) importInputRef.value.value = ''
  }
}

watch(() => props.visible, (v) => { if (v) loadList() })
</script>

<style scoped>
.toolbar { display: flex; gap: 8px; margin-bottom: 12px; }
.pager { margin-top: 12px; justify-content: flex-end; }
.option-row { display: flex; gap: 8px; align-items: center; margin-bottom: 8px; }
.import-actions { display: flex; align-items: center; gap: 8px; }
.file-name { color: #909399; font-size: 13px; margin-left: 4px; }
.import-summary { margin-top: 12px; }
</style>
