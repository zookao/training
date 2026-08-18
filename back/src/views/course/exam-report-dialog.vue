<template>
  <el-dialog :model-value="visible" :title="`考试报告 - ${courseTitle}`" width="900px" @update:model-value="$emit('update:visible', $event)" @open="loadList">
    <div class="toolbar">
      <el-input v-model="keyword" placeholder="试卷名称/账号/姓名/学号" clearable style="width:240px" @keyup.enter="loadList" @clear="loadList" />
      <el-button type="primary" @click="loadList">查询</el-button>
    </div>
    <el-table :data="list" v-loading="loading" border stripe size="small" max-height="500">
      <el-table-column type="index" label="#" width="50" />
      <el-table-column prop="testpaperName" label="试卷" min-width="140" show-overflow-tooltip />
      <el-table-column prop="username" label="账号" width="110" />
      <el-table-column prop="nickname" label="姓名" width="90" />
      <el-table-column prop="studentNo" label="学号" width="100" />
      <el-table-column label="分数" width="70" align="center">
        <template #default="{ row }">
          <span :style="{ color: row.passed ? '#67c23a' : '#f56c6c', fontWeight: 'bold' }">{{ row.score }}</span>
        </template>
      </el-table-column>
      <el-table-column label="结果" width="70" align="center">
        <template #default="{ row }">
          <el-tag :type="row.passed ? 'success' : 'danger'" size="small">{{ row.passed ? '通过' : '未过' }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column label="用时" width="70" align="center">
        <template #default="{ row }">{{ formatDuration(row.duration) }}</template>
      </el-table-column>
      <el-table-column prop="startedAt" label="考试开始时间" width="160" />
      <el-table-column label="操作" width="80" fixed="right">
        <template #default="{ row }">
          <el-button link type="primary" @click="viewDetail(row)">详情</el-button>
        </template>
      </el-table-column>
    </el-table>

    <!-- 答题详情 -->
    <el-dialog v-model="detailVisible" title="答题详情" width="700px" append-to-body>
      <div v-if="detail" v-loading="detailLoading">
        <el-descriptions :column="3" border size="small" style="margin-bottom:12px">
          <el-descriptions-item label="试卷">{{ detail.testpaper?.name }}</el-descriptions-item>
          <el-descriptions-item label="学员">{{ detail.user?.nickname || detail.user?.username }}</el-descriptions-item>
          <el-descriptions-item label="分数">
            <span :style="{ color: detail.record?.passed ? '#67c23a' : '#f56c6c', fontWeight: 'bold' }">{{ detail.record?.score }}</span>
          </el-descriptions-item>
        </el-descriptions>
        <div v-for="(ans, idx) in parsedAnswers" :key="idx" class="answer-item">
          <div class="answer-q">
            <el-tag size="small" :type="ans.correct ? 'success' : 'danger'">{{ ans.correct ? '✓' : '✗' }}</el-tag>
            <span class="q-title">{{ idx + 1 }}. {{ ans.title }}</span>
            <el-tag size="small" style="margin-left:8px">{{ typeText(ans.type) }}</el-tag>
            <span style="margin-left:8px;color:#909399">{{ ans.score }}/{{ ans.maxScore }}分</span>
          </div>
          <div class="answer-row">
            <span>你的答案：</span>
            <span :class="ans.correct ? 'correct' : 'wrong'">{{ formatArr(ans.userAnswer) }}</span>
          </div>
          <div class="answer-row" v-if="!ans.correct">
            <span>正确答案：</span>
            <span class="correct">{{ formatArr(ans.correctAnswer) }}</span>
          </div>
        </div>
      </div>
      <template #footer><el-button @click="detailVisible = false">关闭</el-button></template>
    </el-dialog>
  </el-dialog>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'
import { getExamReport, getExamRecordDetail, type ExamReportItem } from '@/api/testpaper'

const props = defineProps<{ visible: boolean; courseId: number; courseTitle?: string }>()
defineEmits<{ 'update:visible': [val: boolean] }>()

const loading = ref(false)
const list = ref<ExamReportItem[]>([])
const keyword = ref('')

async function loadList() {
  if (!props.courseId) return
  loading.value = true
  try {
    const res = await getExamReport(props.courseId, keyword.value)
    list.value = res.data || []
  } finally { loading.value = false }
}

function formatDuration(sec: number) {
  if (!sec) return '-'
  const m = Math.floor(sec / 60)
  const s = sec % 60
  return m > 0 ? `${m}分${s}秒` : `${s}秒`
}

function typeText(t: number) { return t === 1 ? '单选' : t === 2 ? '多选' : '判断' }
function formatArr(arr: string[]) { return (arr || []).join(', ') }

// 详情
const detailVisible = ref(false)
const detailLoading = ref(false)
const detail = ref<any>(null)
const parsedAnswers = ref<any[]>([])

async function viewDetail(row: ExamReportItem) {
  detailVisible.value = true
  detailLoading.value = true
  detail.value = null
  try {
    const res = await getExamRecordDetail(row.recordId)
    detail.value = res.data
    try { parsedAnswers.value = JSON.parse(res.data.record.answers || '[]') } catch { parsedAnswers.value = [] }
  } finally { detailLoading.value = false }
}

watch(() => props.visible, (v) => { if (v) { keyword.value = ''; loadList() } })
</script>

<style scoped>
.toolbar { display: flex; gap: 8px; margin-bottom: 12px; }
.answer-item { padding: 10px 0; border-bottom: 1px solid #ebeef5; }
.answer-q { display: flex; align-items: center; gap: 4px; margin-bottom: 6px; }
.q-title { font-weight: 500; }
.answer-row { padding-left: 28px; font-size: 13px; margin-top: 4px; }
.correct { color: #67c23a; font-weight: 500; }
.wrong { color: #f56c6c; font-weight: 500; }
</style>
