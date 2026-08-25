<template>
  <div class="learn-page" v-loading="loading">
    <div class="page-header">
      <el-button link :icon="ArrowLeft" @click="router.back()">返回</el-button>
      <h2 class="page-title">{{ course?.title || '课程学习' }}</h2>
    </div>

    <div v-if="course" class="learn-body">
      <!-- ============ 左侧栏 30% ============ -->
      <div class="left-col">
        <!-- 左上：视频播放 40% -->
        <div class="left-top">
          <div class="player-wrap">
            <video
              v-if="currentVideo"
              ref="videoRef"
              :key="currentVideo.id"
              :src="authUrl(currentVideo.url)"
              :poster="authUrl(currentVideo.thumbnail) || undefined"
              controls
              controlslist="nodownload noremoteplayback"
              preload="metadata"
              class="video-el"
              :class="{ locked: !currentVideo.completed }"
              @loadedmetadata="onLoadedMetadata"
              @timeupdate="onTimeUpdate"
              @seeking="onSeeking"
              @ended="onEnded"
              @play="onPlay"
            />
            <div v-if="currentVideo && !currentVideo.completed" class="locked-tip">
              <el-icon><Lock /></el-icon>
              <span>未完成不可快进</span>
            </div>
            <el-empty v-else-if="!currentVideo" description="暂无视频" :image-size="60" />
          </div>
        </div>

        <!-- 左下：课程目录 60% -->
        <div class="left-bottom">
          <div class="playlist">
            <div class="playlist-header">
              <span>课程目录</span>
              <span class="count">{{ course.videos.length }} 节</span>
            </div>
            <div class="playlist-body">
              <div
                v-for="(v, idx) in course.videos"
                :key="v.id"
                class="playlist-item"
                :class="{ active: v.id === currentVideoId, done: v.completed }"
                @click="switchVideo(v.id)"
              >
                <div class="item-thumb">
                  <img v-if="v.thumbnail" :src="authUrl(v.thumbnail)" :alt="v.title" />
                  <div v-else class="thumb-placeholder">
                    <el-icon :size="18"><VideoCamera /></el-icon>
                  </div>
                  <span class="sort-no">{{ idx + 1 }}</span>
                </div>
                <div class="item-info">
                  <div class="item-title">{{ v.title || '视频 ' + v.sort }}</div>
                  <div class="item-meta">
                    <el-progress :percentage="v.percent" :stroke-width="4" :show-text="false" />
                    <span class="item-pct">{{ v.percent }}%</span>
                  </div>
                </div>
                <el-icon v-if="v.completed" class="done-icon"><CircleCheckFilled /></el-icon>
              </div>
              <!-- 考试入口 -->
              <div
                v-for="tp in testpapers"
                :key="'exam-' + tp.id"
                class="playlist-item exam-item"
                :class="{ disabled: !tp.available }"
                @click="openExam(tp)"
              >
                <div class="item-thumb">
                  <div class="thumb-placeholder exam-thumb">
                    <el-icon :size="18"><DocumentRemove /></el-icon>
                  </div>
                  <span class="sort-no"><el-icon><Lock v-if="!tp.available" /><Timer v-else /></el-icon></span>
                </div>
                <div class="item-info">
                  <div class="item-title">考试：{{ tp.name }}</div>
                  <div class="item-meta">
                    <el-tag size="small" :type="tp.type === 1 ? 'primary' : 'warning'">{{ tp.type === 1 ? '随时考试' : '课程完成后考试' }}</el-tag>
                    <span class="item-pct">{{ tp.duration }}分钟</span>
                    <el-tag v-if="tp.inProgress" size="small" type="success" effect="dark" class="exam-progress-tag">正在考试中</el-tag>
                    <span v-if="tp.hasFinished" class="exam-score" :class="tp.bestPassed ? 'pass' : 'fail'">最高分：{{ tp.bestScore }}</span>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- ============ 右侧栏 70% ============ -->
      <div class="right-col">
        <!-- 右上：课件预览 90% -->
        <div class="right-top">
          <div v-if="currentVideo?.coursewarePdf" class="courseware-viewer">
            <template v-if="coursewareError">
              <div class="viewer-error">
                <el-icon :size="40"><WarningFilled /></el-icon>
                <span>课件加载失败</span>
                <el-button size="small" @click="loadCourseware">重试</el-button>
              </div>
            </template>
            <template v-else>
              <div ref="pptxContainerRef" class="pptx-container">
                <div v-if="coursewareLoading" class="viewer-loading">
                  <el-icon class="is-loading" :size="32"><Loading /></el-icon>
                  <span>课件加载中...</span>
                </div>
                <canvas v-show="!coursewareLoading && totalPages > 0" ref="pdfCanvasRef" class="pdf-canvas" />
              </div>
              <div v-if="totalPages > 0" class="viewer-nav">
                <el-button :icon="ArrowLeftBold" size="small" :disabled="currentPage <= 1" @click="prevPage">上一页</el-button>
                <span class="page-info">{{ currentPage }} / {{ totalPages }}</span>
                <el-button :icon="ArrowRightBold" size="small" :disabled="currentPage >= totalPages" @click="nextPage">下一页</el-button>
                <span v-if="currentPageRange" class="page-range" :title="'当前页对应视频区间'">
                  <el-icon><Timer /></el-icon>
                  {{ currentPageRange }}
                </span>
                <span v-if="linkageActive" class="linkage-badge" title="视频播放自动联动翻页中">
                  <el-icon class="is-loading"><Connection /></el-icon>
                  联动
                </span>
              </div>
            </template>
          </div>
          <el-empty v-else description="本节暂无课件" :image-size="80">
            <template #image>
              <el-icon :size="64" style="color:#c0c4cc"><DocumentRemove /></el-icon>
            </template>
          </el-empty>
        </div>

        <!-- 右下：视频描述 + 课件下载 10% -->
        <div class="right-bottom">
          <div class="desc-header">
            <div class="desc-title">{{ currentVideo?.title || '视频 ' + (currentVideo?.sort || '') }}</div>
            <el-button
              v-if="currentVideo?.courseware"
              type="primary"
              plain
              :icon="Download"
              size="small"
              @click="downloadCourseware"
            >下载课件</el-button>
          </div>
          <div class="desc-text">{{ currentVideo?.description || '暂无描述' }}</div>
        </div>
      </div>
    </div>

    <!-- 考试弹窗 -->
    <ExamDialog v-model:visible="examVisible" :testpaper-id="examTestpaperId" />

    <!-- 学习校验弹窗（滑动验证，不可关闭，必须通过才能继续学习） -->
    <el-dialog
      v-model="checkVisible"
      title="学习校验"
      width="380px"
      :close-on-click-modal="false"
      :close-on-press-escape="false"
      :show-close="false"
      align-center
    >
      <div class="check-body">
        <el-icon :size="36" color="#e6a23c"><WarningFilled /></el-icon>
        <p class="check-tip">请完成滑动校验以继续学习</p>
        <SlideCheck :key="checkSlideKey" @success="onCheckSuccess" />
      </div>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onBeforeUnmount, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ArrowLeft, ArrowLeftBold, ArrowRightBold, VideoCamera, CircleCheckFilled, Lock, Download, DocumentRemove, Loading, WarningFilled, Timer, Connection } from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus'
import * as pdfjsLib from 'pdfjs-dist'
import PdfWorker from 'pdfjs-dist/build/pdf.worker.min.mjs?worker'
import { getCourseLearn, reportProgress, checkPass, type CourseLearnRes, type VideoLearnItem } from '@/api/learning'
import { getCourseExam, type ExamTestpaperItem } from '@/api/exam'
import { authUrl } from '@/utils/authUrl'
import ExamDialog from './exam-dialog.vue'
import SlideCheck from './slide-check.vue'

pdfjsLib.GlobalWorkerOptions.workerPort = new PdfWorker()

const route = useRoute()
const router = useRouter()
const courseId = Number(route.params.id)
const classId = Number(route.query.classId) || 0

const loading = ref(false)
const course = ref<CourseLearnRes | null>(null)
const videoRef = ref<HTMLVideoElement | null>(null)
const currentVideoId = ref<number>(0)
let pendingAutoPlay = false    // 切换视频后是否自动播放（autoNext/手动切换置 true，首次加载不自动播放）

// 考试
const testpapers = ref<ExamTestpaperItem[]>([])
const examVisible = ref(false)
const examTestpaperId = ref(0)

// 学习校验（滑动验证）
const checkVisible = ref(false)

// 当前播放视频对象
const currentVideo = computed<VideoLearnItem | undefined>(() => {
  if (!course.value) return undefined
  return course.value.videos.find((v) => v.id === currentVideoId.value)
})

// ---------- 课件在线预览（PDF.js 翻页模式） ----------
// 后端上传 PPTX 时已用 LibreOffice 转 PDF，前端用 pdfjs-dist 渲染到 canvas。
// 原始 PPTX 保留用于下载，不再剥离任何元素，排版 100% 保真。
const coursewareLoading = ref(false)
const coursewareError = ref(false)
const pptxContainerRef = ref<HTMLElement | null>(null)
const pdfCanvasRef = ref<HTMLCanvasElement | null>(null)
const currentPage = ref(1)
const totalPages = ref(0)
let pdfDoc: pdfjsLib.PDFDocumentProxy | null = null
// 当前在途的渲染任务：新渲染前必须 cancel 旧的，否则同一张 canvas 并发 render
// 会出现第二次渲染重置 canvas.width → 清空并重置 transform → 第一次渲染的后续
// 绘制指令以 identity transform 绘制 PDF（y 轴朝上），导致画面上下颠倒。
let currentRenderTask: pdfjsLib.RenderTask | null = null

async function loadCourseware() {
  const video = currentVideo.value
  // 无 PDF（旧课件未转换）则清空，回退到"本节暂无课件"提示
  if (!video?.coursewarePdf) {
    cancelRender()
    if (pdfDoc) {
      pdfDoc.destroy()
      pdfDoc = null
    }
    totalPages.value = 0
    currentPage.value = 1
    return
  }
  // 释放旧文档前先取消在途渲染，避免旧 doc 的 render 在 destroy 后仍写 canvas
  cancelRender()
  if (pdfDoc) {
    pdfDoc.destroy()
    pdfDoc = null
  }
  coursewareError.value = false
  totalPages.value = 0
  currentPage.value = 1
  coursewareLoading.value = true
  try {
    // PDF URL 带 token（后端 JWTAuth 从 query 读取），加 _t 防缓存
    const url = authUrl(video.coursewarePdf) + '&_t=' + Date.now()
    const loadingTask = pdfjsLib.getDocument(url)
    pdfDoc = await loadingTask.promise
    totalPages.value = pdfDoc.numPages
    // 根据视频当前位置联动到对应页（而非重置为第1页）
    const v = videoRef.value
    const pos = v ? Math.floor(v.currentTime) : 0
    const target = pageForPosition(pos)
    currentPage.value = target > 0 && target <= totalPages.value ? target : 1
    await renderPage(currentPage.value)
  } catch (e) {
    console.error('[courseware] load error:', e)
    coursewareError.value = true
  } finally {
    coursewareLoading.value = false
  }
}

// 渲染指定页到 canvas，按容器尺寸自适应缩放（含 devicePixelRatio 提升清晰度）
async function renderPage(n: number) {
  if (!pdfDoc || !pdfCanvasRef.value) return
  // 取消上一次在途渲染，避免同一 canvas 并发 render 导致画面颠倒/撕裂
  cancelRender()
  const page = await pdfDoc.getPage(n)
  // getPage 是异步的，期间可能已切换视频/销毁 doc，需复检
  if (!pdfDoc || !pdfCanvasRef.value) return
  const canvas = pdfCanvasRef.value
  const ctx = canvas.getContext('2d')
  if (!ctx) return
  const container = pptxContainerRef.value
  const containerW = container ? Math.max(0, container.clientWidth - 16) : 800
  const containerH = container ? Math.max(0, container.clientHeight - 16) : 600
  const viewport1 = page.getViewport({ scale: 1 })
  let scale = Math.min(containerW / viewport1.width, containerH / viewport1.height)
  // 容器尚未布局（宽高为 0）时 scale 退化为按宽度或默认 1，避免 scale<=0
  // 触发 viewport transform 翻转，产生倒立画面
  if (!(scale > 0)) {
    scale = containerW > 0 ? containerW / viewport1.width : (containerH > 0 ? containerH / viewport1.height : 1)
  }
  const dpr = window.devicePixelRatio || 1
  const viewport = page.getViewport({ scale: scale * dpr })
  canvas.width = viewport.width
  canvas.height = viewport.height
  canvas.style.width = viewport1.width * scale + 'px'
  canvas.style.height = viewport1.height * scale + 'px'
  currentRenderTask = page.render({ canvasContext: ctx, viewport })
  try {
    await currentRenderTask.promise
  } catch (e: unknown) {
    // cancel() 会以 RenderingCancelledException reject，属正常中断，忽略
    const name = (e as { name?: string })?.name
    if (name !== 'RenderingCancelledException') throw e
  } finally {
    currentRenderTask = null
  }
}

// 取消在途渲染任务：cancel 后其 promise 会以 RenderingCancelledException reject
function cancelRender() {
  if (currentRenderTask) {
    try {
      currentRenderTask.cancel()
    } catch {
      /* ignore */
    }
    currentRenderTask = null
  }
}

function prevPage() {
  if (currentPage.value > 1) {
    currentPage.value--
    renderPage(currentPage.value)
    // 手动翻页后短暂禁用联动，便于预览
    manualUntil = Date.now() + MANUAL_LOCK_MS
  }
}

function nextPage() {
  if (currentPage.value < totalPages.value) {
    currentPage.value++
    renderPage(currentPage.value)
    manualUntil = Date.now() + MANUAL_LOCK_MS
  }
}

// ---------- 视频播放联动课件翻页 ----------
// 解析当前视频的课件每页时长配置 [10,60,300]（秒）
const pageDurationsArr = computed<number[]>(() => {
  const video = currentVideo.value
  if (!video?.coursewarePages) return []
  try {
    const parsed = JSON.parse(video.coursewarePages)
    if (Array.isArray(parsed)) return parsed.map((n: unknown) => Number(n) || 0)
  } catch {
    /* ignore */
  }
  return []
})

// 联动是否激活（有页面时长配置 + 课件已渲染）
const linkageActive = computed(() => pageDurationsArr.value.length > 0 && totalPages.value > 0)

// 当前页对应的视频区间文本，如 "10s - 70s"
const currentPageRange = computed(() => {
  const durations = pageDurationsArr.value
  const idx = currentPage.value - 1
  if (idx < 0 || idx >= durations.length) return ''
  let start = 0
  for (let i = 0; i < idx; i++) start += durations[i] || 0
  const end = start + (durations[idx] || 0)
  if (end <= start) return ''
  return `${start}s - ${end}s`
})

// 手动浏览锁定：用户手动翻页后，N 秒内不自动联动，便于预览
let manualUntil = 0
const MANUAL_LOCK_MS = 5000

// 根据视频播放位置计算应显示的课件页码（1-based），返回 0 表示无映射
function pageForPosition(position: number): number {
  const durations = pageDurationsArr.value
  if (durations.length === 0) return 0
  let start = 0
  for (let i = 0; i < durations.length; i++) {
    const end = start + (durations[i] || 0)
    if (position < end) return i + 1
    start = end
  }
  // 超出最后一页时长，停留在最后一页
  return durations.length
}

// 根据视频当前位置自动翻页（手动浏览期间跳过）
function syncPageToPosition(position: number) {
  if (!linkageActive.value) return
  if (Date.now() < manualUntil) return
  const target = pageForPosition(position)
  if (target <= 0) return
  const clamped = Math.min(target, totalPages.value)
  if (clamped !== currentPage.value) {
    currentPage.value = clamped
    renderPage(currentPage.value)
  }
}

// 切换视频时弹出待完成的校验（课件加载改在 onLoadedMetadata，确保 video 元素就绪后再加载）
watch(currentVideo, () => {
  if (currentVideo.value?.checkPending && !checkVisible.value) {
    triggerCheck(false)
  }
})

// 上报状态
let reportTimer: number | null = null
const REPORT_INTERVAL = 10        // 每 10 秒上报一次
let lastReportPosition = -1       // 上次上报的位置，避免重复上报
let reporting = false             // 是否正在上报中，防止并发请求

// 防快进：记录最远自然播放点与拖拽前位置
const maxPlayedTime = ref(0)      // 最远已观看秒数（浮点）
let preSeekTime = 0               // 拖拽前的播放位置

async function loadCourse() {
  loading.value = true
  try {
    const [res, examRes] = await Promise.all([getCourseLearn(courseId), getCourseExam(courseId)])
    course.value = res.data
    testpapers.value = examRes.data || []
    // 选中第一个未完成的视频；若全部已完成则选第一节
    const first = course.value.videos.find((v) => !v.completed) || course.value.videos[0]
    if (first) {
      currentVideoId.value = first.id
    }
  } finally {
    loading.value = false
  }
}

async function refreshExam() {
  try {
    const examRes = await getCourseExam(courseId)
    testpapers.value = examRes.data || []
  } catch {
    /* ignore */
  }
}

function openExam(tp: ExamTestpaperItem) {
  if (!tp.available) {
    ElMessage.warning('课程未完成，暂不能考试')
    return
  }
  examTestpaperId.value = tp.id
  examVisible.value = true
}

// 考试弹窗关闭后刷新试卷列表：更新最高分、移除/新增"正在考试中"状态
watch(examVisible, (val) => {
  if (!val) refreshExam()
})

// 仅切换视频（不触发上报），用于自动播放下一节等已上报场景
function selectVideo(videoId: number) {
  if (videoId === currentVideoId.value) return
  currentVideoId.value = videoId
}

// 手动切换视频：先上报当前进度，再切换
function switchVideo(videoId: number) {
  if (videoId === currentVideoId.value) return
  flushReport(true)
  pendingAutoPlay = true
  selectVideo(videoId)
}

// ---------- 视频事件 ----------

function onLoadedMetadata() {
  const v = videoRef.value
  if (!v) return
  const rec = currentVideo.value
  // 防快进基准：先以服务端记录的最远点初始化，必须早于 currentTime 赋值
  const startMax = rec?.maxPosition || 0
  maxPlayedTime.value = Math.max(startMax, v.currentTime)
  preSeekTime = v.currentTime
  // 恢复上次播放位置（用 position，非 maxPosition，避免跳到最远）
  if (rec && rec.position > 0 && rec.position < (v.duration - 2)) {
    v.currentTime = rec.position
    preSeekTime = rec.position
  }
  // 切换/加载后重置上次上报位置基准
  lastReportPosition = Math.floor(v.currentTime)
  // 视频元素就绪后加载课件（watch(currentVideo) 在 DOM 重建前触发会取到旧 videoRef，
  // 改在此处加载确保 videoRef 指向新元素，避免与新 video 的 loadedmetadata 时序交叉）
  loadCourseware()
  syncPageToPosition(Math.floor(v.currentTime))
  // 自动播放下一节 / 手动切换后自动播放（首次加载不自动播放，等用户主动点）
  if (pendingAutoPlay) {
    pendingAutoPlay = false
    v.play().catch(() => {})
  }
}

// 记录自然播放进度（仅本地，不触发上报）
function onTimeUpdate() {
  const v = videoRef.value
  if (!v) return
  preSeekTime = v.currentTime
  if (v.currentTime > maxPlayedTime.value) {
    maxPlayedTime.value = v.currentTime
  }
  // 视频播放位置变化时联动课件翻页
  syncPageToPosition(Math.floor(v.currentTime))
  // 学习校验：到达校验点触发滑动弹窗（未弹出时）
  const rec = currentVideo.value
  if (rec && rec.nextCheckPosition > 0 && !checkVisible.value && Math.floor(v.currentTime) >= rec.nextCheckPosition) {
    triggerCheck(true)
  }
}

// 防快进：未观看完成时，禁止拖动到未观看区域
function onSeeking() {
  const v = videoRef.value
  const rec = currentVideo.value
  if (!v || !rec) return
  if (rec.completed) return // 已完成允许自由拖动
  if (v.currentTime > maxPlayedTime.value + 0.5) {
    v.currentTime = preSeekTime
  }
}

// ---------- 学习校验（滑动验证） ----------
const checkSlideKey = ref(0)       // 滑块组件 key，失败时自增以重置
let checkNeedsSync = false         // 本次触发是否需要先同步服务端（首次到达校验点时为 true）

// 触发校验：暂停视频并弹出滑动验证
function triggerCheck(needsSync: boolean) {
  const v = videoRef.value
  if (v) v.pause()
  checkNeedsSync = needsSync
  checkSlideKey.value++ // 每次弹出重置滑块
  checkVisible.value = true
}

// 阻止校验弹窗打开期间播放
function onPlay() {
  if (checkVisible.value) {
    const v = videoRef.value
    if (v) v.pause()
  }
}

// 滑动成功：同步进度 → 通过校验 → 恢复播放
async function onCheckSuccess() {
  const rec = currentVideo.value
  if (!rec) return
  // 首次到达校验点触发时，先上报使服务端置为待校验（CheckPass 依赖该状态）
  if (checkNeedsSync) {
    await flushReport(true)
  }
  try {
    const res = await checkPass(rec.id)
    if (course.value) {
      const target = course.value.videos.find((x) => x.id === rec.id)
      if (target) {
        target.checkPending = false
        target.nextCheckPosition = res.data.nextCheckPosition
      }
    }
    checkVisible.value = false
    const v = videoRef.value
    if (v) v.play().catch(() => {})
  } catch (e) {
    // 校验失败：重置滑块，保持弹窗
    ElMessage.error('校验失败，请重试')
    checkSlideKey.value++
  }
}

function onEnded() {
  // 播放结束上报一次，标记 completed
  flushReport(true)
  ElMessage.success('本节视频已学完')
  // 自动切换到下一节（onEnded 已上报，直接 selectVideo，避免重复上报）
  autoNext()
}

function autoNext() {
  if (!course.value) return
  const videos = course.value.videos
  const idx = videos.findIndex((v) => v.id === currentVideoId.value)
  if (idx === -1) return
  const next = videos[idx + 1]
  if (next) {
    ElMessage.info('即将播放下一节')
    pendingAutoPlay = true
    setTimeout(() => selectVideo(next.id), 800)
  }
}

// ---------- 进度上报 ----------

async function flushReport(force: boolean) {
  const v = videoRef.value
  const rec = currentVideo.value
  if (!v || !rec) return
  const position = Math.floor(v.currentTime)
  // 兜底：优先用服务端返回的 rec.duration，旧视频无值时才取视频元素时长
  const duration = rec.duration || Math.floor(v.duration || 0)
  // 非强制：位置未变化则跳过；正在上报中也跳过，避免并发重复请求
  if (!force) {
    if (reporting) return
    if (position === lastReportPosition) return
  }
  reporting = true
  lastReportPosition = position
  try {
    const res = await reportProgress({
      videoId: rec.id,
      courseId,
      classId,
      position,
      duration,
      completed: false
    })
    // 同步本地状态（不覆盖 duration，保留服务端返回的值）
    if (course.value) {
      const target = course.value.videos.find((x) => x.id === rec.id)
      if (target) {
        const wasCompleted = target.completed
        target.position = position
        target.maxPosition = res.data.maxPosition
        target.percent = res.data.percent
        target.completed = res.data.completed
        target.checkPending = res.data.checkPending
        target.nextCheckPosition = res.data.nextCheckPosition
        // 视频刚完成时刷新考试可用状态
        if (!wasCompleted && res.data.completed) {
          refreshExam()
        }
      }
    }
  } catch (e) {
    /* 上报失败不影响播放 */
  } finally {
    reporting = false
  }
}

// 唯一的周期性上报机制：每 10 秒由定时器触发一次（播放中才上报）
// 注意：不再使用 timeupdate 触发上报，避免与定时器叠加导致重复请求
function startTimer() {
  stopTimer()
  reportTimer = window.setInterval(() => {
    const v = videoRef.value
    if (v && !v.paused && !v.ended) {
      flushReport(false)
    }
  }, REPORT_INTERVAL * 1000)
}
function stopTimer() {
  if (reportTimer !== null) {
    clearInterval(reportTimer)
    reportTimer = null
  }
}

// 切换视频时重置上次上报位置基准与防快进基准（新视频 onLoadedMetadata 会重新校准）
watch(currentVideoId, () => {
  lastReportPosition = -1
  maxPlayedTime.value = 0
  preSeekTime = 0
  // 切换视频时清除手动锁定，新视频立即恢复联动
  manualUntil = 0
})

// 页面离开前上报 + 关闭定时器（注意：beforeunload 用 sendBeacon 更可靠，这里用普通请求）
function handleBeforeUnload() {
  flushReport(true)
}
function handleVisibilityChange() {
  if (document.hidden) {
    flushReport(true)
  }
}

function handleResize() {
  renderPage(currentPage.value)
}

// 下载课件（PPTX）
function downloadCourseware() {
  if (!currentVideo.value?.courseware) return
  const url = authUrl(currentVideo.value.courseware)
  const a = document.createElement('a')
  a.href = url
  a.download = (currentVideo.value.title || '课件') + '.pptx'
  a.target = '_blank'
  document.body.appendChild(a)
  a.click()
  document.body.removeChild(a)
}

onMounted(async () => {
  await loadCourse()
  startTimer()
  window.addEventListener('beforeunload', handleBeforeUnload)
  document.addEventListener('visibilitychange', handleVisibilityChange)
  window.addEventListener('resize', handleResize)
})

onBeforeUnmount(() => {
  flushReport(true)
  stopTimer()
  cancelRender()
  if (pdfDoc) {
    pdfDoc.destroy()
    pdfDoc = null
  }
  window.removeEventListener('beforeunload', handleBeforeUnload)
  document.removeEventListener('visibilitychange', handleVisibilityChange)
  window.removeEventListener('resize', handleResize)
})
</script>

<style scoped lang="scss">
.learn-page {
  height: 100%;
  display: flex;
  flex-direction: column;
  padding: 16px;
  gap: 12px;
  box-sizing: border-box;
  background: #f5f7fa;
}

/* ---- 顶部标题栏 ---- */
.page-header {
  display: flex;
  align-items: center;
  gap: 12px;
  flex-shrink: 0;
  padding: 10px 16px;
  background: #fff;
  border: 1px solid #ebeef5;
  border-radius: 6px;
  box-shadow: 0 1px 2px rgba(0, 0, 0, 0.03);
  .page-title {
    margin: 0;
    font-size: 18px;
    font-weight: 600;
    color: #303133;
  }
}

/* ============ 三分屏主体 ============ */
.learn-body {
  flex: 1;
  display: flex;
  gap: 12px;
  min-height: 0; /* 关键：允许子元素滚动 */
}

/* ---- 左侧栏 30% ---- */
.left-col {
  width: 30%;
  display: flex;
  flex-direction: column;
  gap: 12px;
  min-height: 0;
}
/* 左上：视频 40% */
.left-top {
  flex: 4;
  min-height: 0;
}
.player-wrap {
  position: relative;
  background: #000;
  border: 1px solid #ebeef5;
  border-radius: 6px;
  box-shadow: 0 1px 2px rgba(0, 0, 0, 0.03);
  overflow: hidden;
  width: 100%;
  height: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
}
.video-el {
  width: 100%;
  height: 100%;
  object-fit: contain;
  display: block;
  // 未观看完成时隐藏进度条与当前时间，仅保留剩余时间显示
  &.locked::-webkit-media-controls-timeline {
    display: none !important;
  }
  &.locked::-webkit-media-controls-current-time-display {
    display: none !important;
  }
}
.locked-tip {
  position: absolute;
  top: 10px;
  left: 10px;
  background: rgba(0, 0, 0, 0.65);
  color: #fff;
  font-size: 11px;
  padding: 4px 10px;
  border-radius: 12px;
  display: flex;
  align-items: center;
  gap: 4px;
  pointer-events: none;
  z-index: 2;
}

/* 左下：课程目录 60% */
.left-bottom {
  flex: 6;
  min-height: 0;
}
.playlist {
  height: 100%;
  background: #fff;
  border: 1px solid #ebeef5;
  border-radius: 6px;
  box-shadow: 0 1px 2px rgba(0, 0, 0, 0.03);
  display: flex;
  flex-direction: column;
  overflow: hidden;
}
.playlist-header {
  padding: 12px 16px;
  border-bottom: 1px solid #ebeef5;
  font-weight: 600;
  font-size: 14px;
  color: #303133;
  display: flex;
  justify-content: space-between;
  align-items: center;
  flex-shrink: 0;
  .count {
    font-size: 12px;
    color: #909399;
    font-weight: normal;
    background: #f5f7fa;
    border-radius: 10px;
    padding: 1px 8px;
  }
}
.playlist-body {
  flex: 1;
  overflow-y: auto;
}
.playlist-item {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 10px 14px;
  cursor: pointer;
  border-bottom: 1px solid #f5f7fa;
  transition: background 0.15s;
  &:last-child {
    border-bottom: none;
  }
  &:hover {
    background: #f5f7fa;
  }
  &.active {
    background: #ecf5ff;
    box-shadow: inset 3px 0 0 #409eff;
    .item-title {
      color: #409eff;
    }
  }
  &.disabled {
    opacity: 0.5;
    cursor: not-allowed;
    &:hover { background: transparent; }
  }
  &.exam-item {
    border-top: 2px solid #e4e7ed;
    .exam-thumb {
      background: #fdf6ec;
      color: #e6a23c;
    }
  }
}
.item-thumb {
  position: relative;
  width: 72px;
  height: 42px;
  border-radius: 4px;
  overflow: hidden;
  background: #f0f2f5;
  flex-shrink: 0;
  img {
    width: 100%;
    height: 100%;
    object-fit: cover;
  }
  .thumb-placeholder {
    width: 100%;
    height: 100%;
    display: flex;
    align-items: center;
    justify-content: center;
    color: #c0c4cc;
  }
  .sort-no {
    position: absolute;
    left: 3px;
    top: 3px;
    background: rgba(0, 0, 0, 0.6);
    color: #fff;
    font-size: 10px;
    line-height: 1;
    border-radius: 2px;
    padding: 2px 4px;
  }
}
.item-info {
  flex: 1;
  min-width: 0;
  .item-title {
    font-size: 13px;
    color: #303133;
    margin-bottom: 6px;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
  .item-meta {
    display: flex;
    align-items: center;
    gap: 6px;
    :deep(.el-progress) {
      flex: 1;
    }
    .item-pct {
      font-size: 11px;
      color: #909399;
      white-space: nowrap;
      flex-shrink: 0;
    }
    .exam-progress-tag {
      flex-shrink: 0;
    }
    .exam-score {
      font-size: 11px;
      font-weight: 600;
      white-space: nowrap;
      flex-shrink: 0;
      &.pass { color: #67c23a; }
      &.fail { color: #f56c6c; }
    }
  }
}
.done-icon {
  color: #67c23a;
  font-size: 16px;
  flex-shrink: 0;
}

/* ---- 右侧栏 70% ---- */
.right-col {
  width: 70%;
  display: flex;
  flex-direction: column;
  gap: 12px;
  min-height: 0;
}
/* 右上：课件预览 90% */
.right-top {
  flex: 9;
  min-height: 0;
  background: #fff;
  border: 1px solid #ebeef5;
  border-radius: 6px;
  box-shadow: 0 1px 2px rgba(0, 0, 0, 0.03);
  display: flex;
  align-items: center;
  justify-content: center;
  overflow: hidden;
}
.courseware-viewer {
  width: 100%;
  height: 100%;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}
.viewer-nav {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  padding: 8px 16px;
  border-top: 1px solid #ebeef5;
  background: #fafafa;
  flex-shrink: 0;
}
.page-info {
  font-size: 13px;
  color: #606266;
  min-width: 60px;
  text-align: center;
  font-variant-numeric: tabular-nums;
}
.page-range {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  font-size: 12px;
  color: #409eff;
  font-family: 'Menlo', 'Consolas', monospace;
  background: #ecf5ff;
  border-radius: 10px;
  padding: 2px 8px;
}
.linkage-badge {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  font-size: 11px;
  color: #67c23a;
  background: #f0f9eb;
  border: 1px solid #e1f3d8;
  border-radius: 10px;
  padding: 2px 8px;
}
.pptx-container {
  flex: 1;
  overflow: hidden;
  position: relative;
  background: #f5f5f5;
}
.pdf-canvas {
  position: absolute;
  left: 50%;
  top: 50%;
  transform: translate(-50%, -50%);
}
.viewer-loading {
  position: absolute;
  inset: 0;
  background: #f5f5f5;
  z-index: 10;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 12px;
  color: #909399;
  font-size: 14px;
}
.viewer-error {
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 12px;
  color: #909399;
  font-size: 14px;
}

/* 右下：视频描述 + 课件下载 10% */
.right-bottom {
  flex: 1;
  min-height: 0;
  background: #fff;
  border: 1px solid #ebeef5;
  border-radius: 6px;
  box-shadow: 0 1px 2px rgba(0, 0, 0, 0.03);
  padding: 12px 16px;
  display: flex;
  flex-direction: column;
  gap: 8px;
  overflow: hidden;
}
.desc-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  min-width: 0;
  .desc-title {
    font-size: 15px;
    font-weight: 600;
    color: #303133;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    min-width: 0;
  }
}
.desc-text {
  font-size: 12px;
  line-height: 1.6;
  color: #909399;
  overflow: hidden;
  text-overflow: ellipsis;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
}
.check-body {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 12px;
  padding: 8px 0 4px;
  .check-tip {
    margin: 0;
    font-size: 14px;
    color: #606266;
  }
}

/* ---- 响应式：小屏堆叠 ---- */
@media (max-width: 900px) {
  .learn-body {
    flex-direction: column;
  }
  .left-col,
  .right-col {
    width: 100%;
  }
  .left-top {
    flex: none;
    aspect-ratio: 16 / 9;
  }
  .left-bottom {
    flex: none;
    height: 300px;
  }
  .right-top {
    flex: none;
    height: 200px;
  }
  .right-bottom {
    flex: none;
    height: auto;
  }
}
</style>
