<template>
  <el-dialog :model-value="visible" :title="exam?.name || '考试'" width="800px" top="5vh" :before-close="handleBeforeClose" @update:model-value="onVisibleChange" @open="onOpen" @closed="onClosed">
    <!-- 考试信息 + 计时器 -->
    <div v-if="exam && !submitted" class="exam-header">
      <div class="exam-info">
        <span>总分：{{ exam.totalScore }}分</span>
        <span>及格：{{ exam.passScore }}分</span>
        <span>题量：{{ exam.questions.length }}题</span>
      </div>
      <div class="exam-timer" :class="{ urgent: remainSec < 60 }">
        <el-icon><Timer /></el-icon>
        <span>{{ formatTime(remainSec) }}</span>
      </div>
    </div>

    <!-- 答题区 -->
    <div v-if="exam && !submitted" class="exam-body">
      <div v-for="(q, idx) in parsedQuestions" :key="q.id" class="question-item">
        <div class="q-title">
          <span class="q-num">{{ idx + 1 }}.</span>
          <el-tag size="small" style="margin-right:6px">{{ typeText(q.type) }}</el-tag>
          <span>{{ q.title }}</span>
          <span class="q-score">（{{ q.score }}分）</span>
        </div>
        <div class="q-options">
          <template v-if="q.type === 1 || q.type === 3">
            <el-radio-group v-model="answers[q.id]">
              <el-radio v-for="opt in q.parsedOptions" :key="opt.label" :value="opt.label" class="option-item">
                {{ opt.label }}. {{ opt.content }}
              </el-radio>
            </el-radio-group>
          </template>
          <template v-else>
            <el-checkbox-group v-model="answers[q.id]">
              <el-checkbox v-for="opt in q.parsedOptions" :key="opt.label" :value="opt.label" class="option-item">
                {{ opt.label }}. {{ opt.content }}
              </el-checkbox>
            </el-checkbox-group>
          </template>
        </div>
      </div>
    </div>

    <!-- 结果区 -->
    <div v-if="submitted && result" class="result-body">
      <el-result :icon="result.passed ? 'success' : 'error'" :title="result.passed ? '恭喜通过！' : '未通过'" :sub-title="`得分：${result.score} / ${result.total}（及格线：${result.passLine}）`" />
      <div class="result-details">
        <div v-for="(d, idx) in result.details" :key="idx" class="detail-item">
          <div class="detail-q">
            <el-tag size="small" :type="d.correct ? 'success' : 'danger'">{{ d.correct ? '✓' : '✗' }}</el-tag>
            <span>{{ idx + 1 }}. {{ d.title }}</span>
            <span style="margin-left:8px;color:#909399">{{ d.score }}/{{ d.maxScore }}分</span>
          </div>
          <div class="detail-a">
            <span>你的答案：</span>
            <span :class="d.correct ? 'ok' : 'no'">{{ d.userAnswer.join(', ') || '未作答' }}</span>
            <!-- 暂时隐藏正确答案，如需恢复将 v-if 改为 !d.correct -->
            <template v-if="false">
              <span style="margin-left:12px">正确答案：</span>
              <span class="ok">{{ d.correctAnswer.join(', ') }}</span>
            </template>
          </div>
        </div>
      </div>
    </div>

    <template #footer>
      <template v-if="exam && !submitted">
        <el-button @click="abandon">放弃</el-button>
        <el-button type="primary" :loading="submitting" @click="doSubmit">交卷</el-button>
      </template>
      <template v-else-if="submitted">
        <el-button type="primary" @click="$emit('update:visible', false)">关闭</el-button>
      </template>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { ref, computed, onUnmounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Timer } from '@element-plus/icons-vue'
import { getExam, submitExam, saveExamDraft, type ExamDetail, type SubmitResult, type DraftAnswer } from '@/api/exam'

const props = defineProps<{ visible: boolean; testpaperId: number }>()
const emit = defineEmits<{ 'update:visible': [val: boolean] }>()

const exam = ref<ExamDetail | null>(null)
const answers = ref<Record<number, string | string[]>>({})
const submitted = ref(false)
const result = ref<SubmitResult | null>(null)
const submitting = ref(false)
const remainSec = ref(0)
let timer: ReturnType<typeof setInterval> | null = null
let draftTimer: ReturnType<typeof setInterval> | null = null
const DRAFT_INTERVAL = 10000 // 10秒自动保存草稿（断点续考）

interface ParsedQuestion {
  id: number
  type: number
  title: string
  score: number
  parsedOptions: { label: string; content: string }[]
}

const parsedQuestions = computed<ParsedQuestion[]>(() => {
  if (!exam.value) return []
  return exam.value.questions.map(q => {
    let opts: { label: string; content: string }[] = []
    try { opts = JSON.parse(q.options) } catch { opts = [] }
    return { id: q.id, type: q.type, title: q.title, score: q.score, parsedOptions: opts }
  })
})

function typeText(t: number) { return t === 1 ? '单选' : t === 2 ? '多选' : '判断' }
function formatTime(sec: number) {
  const m = Math.floor(sec / 60)
  const s = sec % 60
  return `${String(m).padStart(2, '0')}:${String(s).padStart(2, '0')}`
}

async function onOpen() {
  exam.value = null
  submitted.value = false
  result.value = null
  answers.value = {}
  if (!props.testpaperId) return
  try {
    const res = await getExam(props.testpaperId)
    exam.value = res.data
    // 初始化答案（默认空）
    for (const q of res.data.questions) {
      answers.value[q.id] = q.type === 2 ? [] : ''
    }
    // 恢复草稿答案（断点续考）：服务端返回未完成记录中保存的草稿
    if (res.data.draftAnswers && res.data.draftAnswers.length > 0) {
      const draftMap = new Map<number, string[]>()
      for (const d of res.data.draftAnswers) draftMap.set(d.questionId, d.userAnswer)
      for (const q of res.data.questions) {
        const d = draftMap.get(q.id)
        if (!d) continue
        answers.value[q.id] = q.type === 2 ? [...d] : (d.length ? d[0] : '')
      }
    }
    // 启动倒计时（使用服务端返回的剩余时间）
    remainSec.value = res.data.remainSec
    timer = setInterval(() => {
      remainSec.value--
      if (remainSec.value <= 0) {
        clearTimer()
        ElMessage.warning('考试时间到，自动交卷')
        doSubmit()
      }
    }, 1000)
    // 启动草稿自动保存（断点续考，每 10 秒存一次，防意外断电/刷新丢答案）
    draftTimer = setInterval(saveDraft, DRAFT_INTERVAL)
  } catch (e) {
    /* handled in interceptor */
  }
}

function onClosed() {
  clearTimer()
  exam.value = null
  submitted.value = false
  result.value = null
}

function clearTimer() {
  if (timer) { clearInterval(timer); timer = null }
  if (draftTimer) { clearInterval(draftTimer); draftTimer = null }
}

// buildDraftAnswers 将当前答案组装为草稿载荷（同步快照，避免关闭时状态被重置）
function buildDraftAnswers(): DraftAnswer[] {
  return parsedQuestions.value.map(q => {
    const a = answers.value[q.id]
    if (q.type === 2) {
      return { questionId: q.id, userAnswer: (a as string[]) || [] }
    }
    return { questionId: q.id, userAnswer: a ? [a as string] : [] }
  })
}

// saveDraft 保存草稿（断点续考）。失败静默，不打扰用户
async function saveDraft() {
  if (!exam.value || submitted.value || !props.testpaperId) return
  const payload = buildDraftAnswers() // 同步捕获，避免 onClosed 重置后丢失
  try {
    await saveExamDraft(props.testpaperId, payload)
  } catch {
    /* 草稿保存失败静默处理 */
  }
}

// abandon 放弃：先存草稿再关闭，保证已答内容不丢
function abandon() {
  saveDraft()
  emit('update:visible', false)
}

// handleBeforeClose 点击 X / ESC / 遮罩关闭前：先存草稿再放行关闭
// before-close 仅在用户主动关闭时触发（不会因外部改 visible 而触发），是 X 号存草稿的唯一可靠入口
function handleBeforeClose(done: () => void) {
  if (exam.value && !submitted.value) {
    saveDraft()
  }
  done()
}

// onVisibleChange 透传 el-dialog 可见性变更（done() 关闭时会触发 update:modelValue，需转发给父组件）
function onVisibleChange(val: boolean) {
  emit('update:visible', val)
}

async function doSubmit() {
  if (!exam.value) return
  // 检查未答
  const unanswered = parsedQuestions.value.filter(q => {
    const a = answers.value[q.id]
    return q.type === 2 ? (a as string[]).length === 0 : !a
  })
  if (unanswered.length > 0 && remainSec.value > 0) {
    try {
      await ElMessageBox.confirm(`还有 ${unanswered.length} 题未作答，确定交卷？`, '提示', { type: 'warning' })
    } catch { return }
  }
  submitting.value = true
  clearTimer()
  try {
    const ansArr = parsedQuestions.value.map(q => ({
      questionId: q.id,
      userAnswer: q.type === 2 ? (answers.value[q.id] as string[]) : [answers.value[q.id] as string]
    }))
    const res = await submitExam(props.testpaperId, ansArr)
    result.value = res.data
    submitted.value = true
  } finally {
    submitting.value = false
  }
}

onUnmounted(clearTimer)
</script>

<style scoped lang="scss">
.exam-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 12px 16px;
  background: #f5f7fa;
  border-radius: 8px;
  margin-bottom: 16px;
  .exam-info { display: flex; gap: 20px; color: #606266; font-size: 14px; }
  .exam-timer {
    display: flex; align-items: center; gap: 6px;
    font-size: 18px; font-weight: bold; color: #409eff;
    &.urgent { color: #f56c6c; animation: blink 1s infinite; }
  }
}
@keyframes blink { 50% { opacity: 0.6; } }
.exam-body { max-height: 55vh; overflow-y: auto; }
.question-item {
  padding: 12px 0;
  border-bottom: 1px solid #ebeef5;
  .q-title { margin-bottom: 10px; line-height: 1.6; }
  .q-num { font-weight: bold; margin-right: 4px; }
  .q-score { color: #909399; font-size: 13px; }
  .q-options { padding-left: 20px; }
  .option-item { display: block; margin: 8px 0; }
}
.result-body { max-height: 60vh; overflow-y: auto; }
.result-details { margin-top: 12px; }
.detail-item { padding: 10px 0; border-bottom: 1px solid #ebeef5; }
.detail-q { display: flex; align-items: center; gap: 6px; margin-bottom: 6px; }
.detail-a { padding-left: 28px; font-size: 13px; }
.ok { color: #67c23a; font-weight: 500; }
.no { color: #f56c6c; font-weight: 500; }
</style>
