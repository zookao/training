<template>
  <div class="page">
    <el-card shadow="never">
      <div class="toolbar">
        <el-input
          v-model="query.keyword"
          placeholder="搜索课程标题"
          clearable
          style="width: 220px"
          @keyup.enter="loadList"
          @clear="loadList"
        />
        <el-button type="primary" @click="loadList">搜索</el-button>
        <el-button v-permission="'course:add'" type="success" @click="openCreate">新增课程</el-button>
      </div>

      <el-table v-loading="loading" :data="list" border>
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
        <el-table-column prop="title" label="标题" min-width="160" />
        <el-table-column prop="description" label="描述" min-width="200" show-overflow-tooltip />
        <el-table-column label="视频数" width="80" align="center">
          <template #default="{ row }">{{ (row.videos || []).length }}</template>
        </el-table-column>
        <el-table-column prop="sort" label="排序" width="70" align="center" />
        <el-table-column label="状态" width="80" align="center">
          <template #default="{ row }">
            <el-tag size="small" :type="row.status === 1 ? 'success' : 'info'">
              {{ row.status === 1 ? '启用' : '禁用' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="340" fixed="right">
          <template #default="{ row }">
            <el-button v-permission="'course:question'" link type="primary" @click="openQuestion(row)">试题管理</el-button>
            <el-button v-permission="'course:testpaper'" link type="primary" @click="openTestpaper(row)">试卷管理</el-button>
            <el-button v-permission="'course:exam-report'" link type="primary" @click="openExamReport(row)">考试报告</el-button>
            <el-button v-permission="'course:edit'" link type="primary" @click="openEdit(row)">编辑</el-button>
            <el-button v-permission="'course:delete'" link type="danger" @click="handleDelete(row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>

      <el-pagination
        class="pager"
        v-model:current-page="query.page"
        v-model:page-size="query.pageSize"
        :total="total"
        :page-sizes="[10, 20, 50]"
        layout="total, sizes, prev, pager, next"
        @size-change="loadList"
        @current-change="loadList"
      />
    </el-card>

    <!-- 课程表单 -->
    <el-dialog v-model="formVisible" :title="editId ? '编辑课程' : '新增课程'" width="860px" @closed="resetForm">
      <el-form ref="formRef" :model="form" :rules="formRules" label-width="80px">
        <el-form-item label="标题" prop="title"><el-input v-model="form.title" /></el-form-item>
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
          <el-input v-model="form.description" type="textarea" :rows="2" />
        </el-form-item>
        <el-form-item label="排序"><el-input-number v-model="form.sort" :min="0" /></el-form-item>
        <el-form-item label="状态">
          <el-switch v-model="form.status" :active-value="1" :inactive-value="0" />
        </el-form-item>
        <el-form-item label="视频">
          <div class="videos-wrap">
            <el-button v-permission="'course:add'" type="primary" plain @click="openUpload">+ 添加视频</el-button>
            <el-table :data="form.videos" border size="small" class="videos-table" empty-text="暂无视频">
              <el-table-column label="缩略图" width="90" align="center">
                <template #default="{ row }">
                  <el-image
                    v-if="row.thumbnail"
                    :src="authUrl(row.thumbnail)"
                    fit="cover"
                    style="width: 60px; height: 40px"
                    :preview-src-list="[authUrl(row.thumbnail)]"
                    preview-teleported
                  />
                  <span v-else>—</span>
                </template>
              </el-table-column>
              <el-table-column label="标题" min-width="140">
                <template #default="{ row }">
                  <el-input v-model="row.title" size="small" placeholder="视频标题" />
                </template>
              </el-table-column>
              <el-table-column label="时长" width="80" align="center">
                <template #default="{ row }">
                  <span>{{ formatDuration(row.duration) }}</span>
                </template>
              </el-table-column>
              <el-table-column label="课件" width="110" align="center">
                <template #default="{ row }">
                  <template v-if="row.courseware">
                    <el-tag size="small" type="warning">{{ coursewareExt(row.courseware) || '课件' }}</el-tag>
                    <div class="pptx-pages">{{ row.coursewarePageCount || 0 }} 页</div>
                  </template>
                  <span v-else>—</span>
                </template>
              </el-table-column>
              <el-table-column label="描述" min-width="160">
                <template #default="{ row }">
                  <el-input v-model="row.description" size="small" placeholder="视频描述" />
                </template>
              </el-table-column>
              <el-table-column label="排序" width="110" align="center">
                <template #default="{ row }">
                  <el-input-number
                    v-model="row.sort"
                    size="small"
                    :min="0"
                    controls-position="right"
                    style="width: 100px"
                  />
                </template>
              </el-table-column>
              <el-table-column label="操作" width="200" align="center">
                <template #default="{ row, $index }">
                  <el-button
                    v-if="row.courseware"
                    link
                    type="warning"
                    size="small"
                    :disabled="!row.coursewarePageCount"
                    @click="openPageDuration(row)"
                  >课件翻页设置</el-button>
                  <el-button
                    v-if="row.courseware && row.id && !row.coursewarePageCount"
                    link
                    type="primary"
                    size="small"
                    :loading="reparsingId === row.id"
                    @click="handleReparse(row)"
                  >识别页数</el-button>
                  <el-button link type="danger" @click="removeVideo($index)">删除</el-button>
                </template>
              </el-table-column>
            </el-table>
          </div>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="formVisible = false">取消</el-button>
        <el-button type="primary" :loading="saving" @click="submitForm">确定</el-button>
      </template>
    </el-dialog>

    <!-- 上传视频 -->
    <el-dialog v-model="uploadVisible" title="添加视频" width="560px" @closed="resetUpload">
      <el-form label-width="90px">
        <el-form-item label="视频文件" required>
          <input ref="videoInputRef" type="file" accept="video/mp4,video/webm,.mp4,.webm" style="display: none" @change="onVideoPick" />
          <el-button :disabled="uploading" @click="videoInputRef?.click()">选择视频</el-button>
          <span class="file-name">{{ uploadState.videoName }}</span>
          <span v-if="videoSizeText" class="file-size">{{ videoSizeText }}</span>
          <span class="format-tip">仅支持 MP4/WebM（浏览器可播放格式）</span>
        </el-form-item>
        <el-form-item label="缩略图">
          <input ref="thumbInputRef" type="file" accept="image/*" style="display: none" @change="onThumbPick" />
          <el-button :disabled="uploading" @click="thumbInputRef?.click()">选择缩略图</el-button>
          <span class="file-name">{{ uploadState.thumbName }}</span>
          <span class="format-tip">不上传则自动截取视频封面</span>
        </el-form-item>
        <el-form-item label="课件">
          <input ref="coursewareInputRef" type="file" accept=".pptx,.ppt,.odp,.fodp" style="display: none" @change="onCoursewarePick" />
          <el-button :disabled="uploading" @click="coursewareInputRef?.click()">选择课件</el-button>
          <span class="file-name">{{ uploadState.coursewareName }}</span>
          <span v-if="coursewareSizeText" class="file-size">{{ coursewareSizeText }}</span>
          <span class="format-tip">支持 PPTX/PPT/ODP/FODP</span>
        </el-form-item>
        <el-form-item v-if="uploading" label="上传进度">
          <div class="upload-progress">
            <div class="progress-head">
              <span class="progress-label">{{ uploadPhase.label }}</span>
              <span class="progress-percent">
                <template v-if="uploadPhase.phase === 'process'">处理中…</template>
                <template v-else>{{ uploadState.progress }}%</template>
              </span>
            </div>
            <el-progress
              :percentage="uploadState.progress"
              :indeterminate="uploadPhase.phase === 'process'"
              :show-text="false"
              :stroke-width="10"
            />
            <div class="progress-hint">
              <span v-if="uploadPhase.phase === 'upload'">支持断点续传，中断或刷新后可继续上传</span>
              <span v-else>后处理中，请勿关闭页面</span>
            </div>
          </div>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="uploadVisible = false" :disabled="uploading">取消</el-button>
        <el-button v-if="uploading" type="danger" plain @click="cancelUpload">中断上传</el-button>
        <el-button type="primary" :loading="uploading" :disabled="!uploadState.video" @click="doUpload">
          {{ uploading ? '上传中…' : '上传' }}
        </el-button>
      </template>
    </el-dialog>

    <!-- 设置课件每页时长 -->
    <el-dialog v-model="pageDurationVisible" title="设置课件每页时长" width="720px" @closed="resetPageDuration">
      <div v-if="pageDurationTarget" class="page-duration">
        <div class="duration-summary">
          <el-tag type="info">视频时长：{{ formatDuration(pageDurationTarget.duration) }}</el-tag>
          <el-tag type="warning">课件页数：{{ pageDurationTarget.coursewarePageCount }} 页</el-tag>
          <el-tag :type="durationSumType">已分配：{{ formatDuration(durationSum) }}</el-tag>
        </div>
        <div class="duration-actions">
          <el-button size="small" type="primary" plain @click="autoAssignEqual">平均分配视频时长</el-button>
          <el-button size="small" type="success" plain :loading="durationImporting" @click="durationImportRef?.click()">导入打点</el-button>
          <el-button size="small" plain @click="onDownloadDurationTemplate">下载打点模板</el-button>
          <el-button size="small" plain @click="clearDurations">清空</el-button>
          <span class="duration-tip">提示：每页时长表示该页对应的视频播放秒数，播放到对应区间时学员端自动翻页</span>
        </div>
        <input ref="durationImportRef" type="file" accept=".xlsx,.xls" style="display:none" @change="onDurationImportPick" />
        <el-alert
          v-if="pageDurationTarget.duration && durationSum > pageDurationTarget.duration"
          type="warning"
          :closable="false"
          show-icon
          style="margin: 8px 0"
          title="已分配时长超过视频总时长，超出部分的页将不会在播放中触发"
        />
        <el-table :data="pageDurations" border size="small" max-height="380" style="margin-top: 8px">
          <el-table-column label="页码" width="70" align="center" type="index" :index="(i: number) => i + 1" />
          <el-table-column label="本页时长（秒）" width="170">
            <template #default="{ row }">
              <el-input-number
                v-model="row.duration"
                :min="0"
                :max="pageDurationTarget.duration || 99999"
                size="small"
                controls-position="right"
                style="width: 150px"
              />
            </template>
          </el-table-column>
          <el-table-column label="视频打点" min-width="120">
            <template #default="{ $index }">
              <span class="range-text">{{ pageRangeText($index) }}</span>
            </template>
          </el-table-column>
        </el-table>
      </div>
      <template #footer>
        <el-button @click="pageDurationVisible = false">取消</el-button>
        <el-button type="primary" @click="savePageDurations">保存</el-button>
      </template>
    </el-dialog>

    <!-- 试题管理 -->
    <QuestionDialog v-model:visible="questionVisible" :course-id="examCourseId" :course-title="examCourseTitle" />
    <!-- 试卷管理 -->
    <TestpaperDialog v-model:visible="testpaperVisible" :course-id="examCourseId" :course-title="examCourseTitle" />
    <!-- 考试报告 -->
    <ExamReportDialog v-model:visible="examReportVisible" :course-id="examCourseId" :course-title="examCourseTitle" />
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted } from 'vue'
import { ElMessage, ElMessageBox, type FormInstance, type FormRules } from 'element-plus'
import {
  getCourseList,
  getCourseDetail,
  createCourse,
  updateCourse,
  deleteCourse,
  uploadImage,
  reparseCoursewarePages,
  importDurations,
  downloadDurationTemplate,
  type CourseItem,
  type CourseForm,
  type VideoItem,
  type UploadVideoRes
} from '@/api/course'
import { authUrl } from '@/utils/authUrl'
import { uploadFileChunked } from '@/utils/chunkedUpload'
import QuestionDialog from './question-dialog.vue'
import TestpaperDialog from './testpaper-dialog.vue'
import ExamReportDialog from './exam-report-dialog.vue'

const list = ref<CourseItem[]>([])
const loading = ref(false)
const total = ref(0)
const query = reactive({ page: 1, pageSize: 10, keyword: '' })
const coverUploading = ref(false)

async function loadList() {
  loading.value = true
  try {
    const res = await getCourseList({ page: query.page, pageSize: query.pageSize, keyword: query.keyword })
    list.value = res.data.list || []
    total.value = res.data.total
  } finally {
    loading.value = false
  }
}

// 课程表单
const formVisible = ref(false)
const editId = ref(0)
const saving = ref(false)
const formRef = ref<FormInstance>()
const form = reactive<CourseForm>({ title: '', cover: '', description: '', sort: 0, status: 1, videos: [] })
const formRules: FormRules = {
  title: [{ required: true, message: '请输入标题', trigger: 'blur' }]
}

function resetForm() {
  editId.value = 0
  Object.assign(form, { title: '', cover: '', description: '', sort: 0, status: 1, videos: [] })
  formRef.value?.clearValidate()
}
function openCreate() {
  resetForm()
  formVisible.value = true
}
async function openEdit(row: CourseItem) {
  resetForm()
  editId.value = row.id
  const res = await getCourseDetail(row.id)
  const d = res.data
  Object.assign(form, {
    title: d.title,
    cover: d.cover || '',
    description: d.description,
    sort: d.sort,
    status: d.status,
    videos: (d.videos || []).map((v) => ({
      id: v.id,
      url: v.url,
      thumbnail: v.thumbnail,
      courseware: v.courseware || '',
      coursewarePdf: v.coursewarePdf || '',
      coursewarePageCount: v.coursewarePageCount || 0,
      coursewarePages: v.coursewarePages || '',
      title: v.title,
      description: v.description,
      sort: v.sort,
      duration: v.duration
    }))
  })
  formVisible.value = true
}
function removeVideo(idx: number) {
  form.videos.splice(idx, 1)
}
// 时长格式化：秒 → mm:ss 或 hh:mm:ss
function formatDuration(sec?: number) {
  if (!sec || sec <= 0) return '—'
  const h = Math.floor(sec / 3600)
  const m = Math.floor((sec % 3600) / 60)
  const s = sec % 60
  const pad = (n: number) => String(n).padStart(2, '0')
  return h > 0 ? `${h}:${pad(m)}:${pad(s)}` : `${m}:${pad(s)}`
}
async function submitForm() {
  if (!formRef.value) return
  await formRef.value.validate(async (valid) => {
    if (!valid) return
    saving.value = true
    try {
      if (editId.value) await updateCourse(editId.value, form)
      else await createCourse(form)
      ElMessage.success(editId.value ? '更新成功' : '创建成功')
      formVisible.value = false
      loadList()
    } finally {
      saving.value = false
    }
  })
}
async function handleDelete(row: CourseItem) {
  await ElMessageBox.confirm(`确定删除课程「${row.title}」？其下视频将一并删除。`, '提示', { type: 'warning' })
  await deleteCourse(row.id)
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

// 上传视频
const uploadVisible = ref(false)
const uploading = ref(false)
const videoInputRef = ref<HTMLInputElement>()
const thumbInputRef = ref<HTMLInputElement>()
const coursewareInputRef = ref<HTMLInputElement>()
const uploadState = reactive<{
  video: File | null
  videoName: string
  thumb: File | null
  thumbName: string
  courseware: File | null
  coursewareName: string
  progress: number
}>({ video: null, videoName: '', thumb: null, thumbName: '', courseware: null, coursewareName: '', progress: 0 })

// 分片上传阶段：phase 标识上传/后处理，label 描述当前操作
const uploadPhase = ref<{ phase: 'upload' | 'process'; label: string }>({ phase: 'upload', label: '' })
// 取消控制器（中断上传）
let abortController: AbortController | null = null

// 文件大小格式化
function formatSize(bytes: number): string {
  if (!bytes) return '0 B'
  const units = ['B', 'KB', 'MB', 'GB']
  const i = Math.min(units.length - 1, Math.floor(Math.log(bytes) / Math.log(1024)))
  return `${(bytes / Math.pow(1024, i)).toFixed(1)} ${units[i]}`
}
const videoSizeText = computed(() => (uploadState.video ? formatSize(uploadState.video.size) : ''))
const coursewareSizeText = computed(() =>
  uploadState.courseware ? formatSize(uploadState.courseware.size) : ''
)

function resetUpload() {
  uploadState.video = null
  uploadState.videoName = ''
  uploadState.thumb = null
  uploadState.thumbName = ''
  uploadState.courseware = null
  uploadState.coursewareName = ''
  uploadState.progress = 0
  uploadPhase.value = { phase: 'upload', label: '' }
  if (videoInputRef.value) videoInputRef.value.value = ''
  if (thumbInputRef.value) thumbInputRef.value.value = ''
  if (coursewareInputRef.value) coursewareInputRef.value.value = ''
}
function openUpload() {
  resetUpload()
  uploadVisible.value = true
}
function onVideoPick(e: Event) {
  const target = e.target as HTMLInputElement
  const f = target.files?.[0]
  if (!f) return
  // 仅允许浏览器 <video> 原生可播放的容器格式：MP4 / WebM
  if (!/\.(mp4|webm)$/i.test(f.name)) {
    ElMessage.error('不支持的视频格式，仅支持 MP4/WebM（浏览器可播放格式）')
    target.value = ''
    return
  }
  uploadState.video = f
  uploadState.videoName = f.name
}
function onThumbPick(e: Event) {
  const f = (e.target as HTMLInputElement).files?.[0]
  if (f) {
    uploadState.thumb = f
    uploadState.thumbName = f.name
  }
}
function onCoursewarePick(e: Event) {
  const f = (e.target as HTMLInputElement).files?.[0]
  if (f) {
    if (!/\.(pptx|ppt|odp|fodp)$/i.test(f.name)) {
      ElMessage.error('仅支持 PPTX/PPT/ODP/FODP 格式课件')
      return
    }
    uploadState.courseware = f
    uploadState.coursewareName = f.name
  }
}

// 从课件 URL 提取大写扩展名（如 PPTX/PPT/ODP/FODP），用于表格标签展示
function coursewareExt(url: string): string {
  const m = url.match(/\.(pptx|ppt|odp|fodp)$/i)
  return m ? m[1].toUpperCase() : ''
}
async function doUpload() {
  if (!uploadState.video) return
  uploading.value = true
  abortController = new AbortController()

  let videoRes: UploadVideoRes | undefined
  let coursewareRes: UploadVideoRes | undefined

  try {
    // 1. 视频（必需，分片 + 断点续传），缩略图随 merge 一并提交
    uploadPhase.value = { phase: 'upload', label: '上传视频中' }
    uploadState.progress = 0
    videoRes = await uploadFileChunked(uploadState.video, {
      type: 'video',
      thumbnail: uploadState.thumb,
      signal: abortController.signal,
      onProgress: (p) => {
        uploadState.progress = p.percent
        uploadPhase.value = {
          phase: p.phase === 'uploading' ? 'upload' : 'process',
          label:
            p.phase === 'uploading'
              ? '上传视频中'
              : '视频处理中（解析时长 / 生成缩略图）'
        }
      }
    })

    // 2. 课件（可选，分片 + 断点续传）
    if (uploadState.courseware) {
      uploadPhase.value = { phase: 'upload', label: '上传课件中' }
      uploadState.progress = 0
      coursewareRes = await uploadFileChunked(uploadState.courseware, {
        type: 'courseware',
        signal: abortController.signal,
        onProgress: (p) => {
          uploadState.progress = p.percent
          uploadPhase.value = {
            phase: p.phase === 'uploading' ? 'upload' : 'process',
            label:
              p.phase === 'uploading'
                ? '上传课件中'
                : '课件处理中（LibreOffice 转 PDF，可能耗时数分钟）'
          }
        }
      })
    }

    // 3. 合并结果到 form.videos
    const nextSort = form.videos.length ? Math.max(...form.videos.map((v) => v.sort)) + 1 : 1
    const v: VideoItem = {
      url: videoRes.url,
      thumbnail: videoRes.thumbnail,
      courseware: coursewareRes?.courseware || '',
      coursewarePdf: coursewareRes?.coursewarePdf || '',
      coursewarePageCount: coursewareRes?.coursewarePageCount || 0,
      coursewarePages: '',
      title: uploadState.video.name.replace(/\.[^.]+$/, ''),
      description: '',
      sort: nextSort,
      duration: videoRes.duration
    }
    form.videos.push(v)
    ElMessage.success('视频已添加')
    uploadVisible.value = false
  } catch (e: unknown) {
    // 用户主动中断：静默（请求拦截器已忽略 CanceledError 的 toast）
    const name = (e as { name?: string })?.name
    const code = (e as { code?: string })?.code
    if (name === 'AbortError' || name === 'CanceledError' || code === 'ERR_CANCELED') {
      ElMessage.info('已中断上传，可重新点击「上传」续传')
    }
    /* 其他错误由 request 拦截器统一 toast */
  } finally {
    uploading.value = false
    abortController = null
    uploadPhase.value = { phase: 'upload', label: '' }
  }
}

// 中断上传：仅取消在传分片，已传分片保留，下次「上传」可续传
function cancelUpload() {
  abortController?.abort()
}

onMounted(loadList)

// ---------- 考试功能 ----------
const questionVisible = ref(false)
const testpaperVisible = ref(false)
const examReportVisible = ref(false)
const examCourseId = ref(0)
const examCourseTitle = ref('')

function openQuestion(row: CourseItem) {
  examCourseId.value = row.id
  examCourseTitle.value = row.title
  questionVisible.value = true
}
function openTestpaper(row: CourseItem) {
  examCourseId.value = row.id
  examCourseTitle.value = row.title
  testpaperVisible.value = true
}
function openExamReport(row: CourseItem) {
  examCourseId.value = row.id
  examCourseTitle.value = row.title
  examReportVisible.value = true
}

// ---------- 课件每页时长设置 ----------
const pageDurationVisible = ref(false)
const pageDurationTarget = ref<VideoItem | null>(null)
const pageDurations = reactive<{ duration: number }[]>([])
const reparsingId = ref<number>(0)

// 重新识别课件页数（用于旧课件补录页数）
async function handleReparse(row: VideoItem) {
  if (!row.id) return
  reparsingId.value = row.id
  try {
    const res = await reparseCoursewarePages(row.id)
    row.coursewarePageCount = res.data.coursewarePageCount || 0
    ElMessage.success(`已识别 ${row.coursewarePageCount} 页`)
  } catch {
    /* handled in interceptor */
  } finally {
    reparsingId.value = 0
  }
}

const durationSum = computed(() => pageDurations.reduce((s, p) => s + (p.duration || 0), 0))
const durationSumType = computed<'info' | 'success' | 'warning' | 'danger'>(() => {
  if (!pageDurationTarget.value) return 'info'
  const total = pageDurationTarget.value.duration || 0
  if (durationSum.value === 0) return 'info'
  if (total > 0 && durationSum.value > total) return 'danger'
  if (total > 0 && durationSum.value === total) return 'success'
  return 'warning'
})

function openPageDuration(row: VideoItem) {
  pageDurationTarget.value = row
  const count = row.coursewarePageCount || 0
  // 解析已存在的时长配置
  let existing: number[] = []
  if (row.coursewarePages) {
    try {
      const parsed = JSON.parse(row.coursewarePages)
      if (Array.isArray(parsed)) existing = parsed.map((n: unknown) => Number(n) || 0)
    } catch {
      /* ignore */
    }
  }
  pageDurations.splice(0, pageDurations.length)
  for (let i = 0; i < count; i++) {
    pageDurations.push({ duration: existing[i] || 0 })
  }
  pageDurationVisible.value = true
}

function resetPageDuration() {
  pageDurationTarget.value = null
  pageDurations.splice(0, pageDurations.length)
}

// 计算第 idx 页（0-based）对应的视频区间文本
function pageRangeText(idx: number) {
  let start = 0
  for (let i = 0; i < idx; i++) start += pageDurations[i]?.duration || 0
  const dur = pageDurations[idx]?.duration || 0
  if (dur === 0) return '—'
  const end = start + dur
  const h = Math.floor(end / 3600)
  const m = Math.floor((end % 3600) / 60)
  const s = end % 60
  const pad = (n: number) => String(n).padStart(2, '0')
  return `${pad(h)}:${pad(m)}:${pad(s)}`
}

// 平均分配视频时长到所有页（清零后重分）
function autoAssignEqual() {
  if (!pageDurationTarget.value || pageDurations.length === 0) return
  const total = pageDurationTarget.value.duration || 0
  if (total <= 0) {
    ElMessage.warning('视频时长为 0，无法分配')
    return
  }
  const count = pageDurations.length
  const avg = Math.floor(total / count)
  const remainder = total - avg * count
  pageDurations.forEach((p, i) => {
    p.duration = avg + (i < remainder ? 1 : 0)
  })
}

function clearDurations() {
  pageDurations.forEach((p) => {
    p.duration = 0
  })
}

function savePageDurations() {
  if (!pageDurationTarget.value) return
  const arr = pageDurations.map((p) => p.duration || 0)
  const zeroIdx = arr.findIndex((d) => d <= 0)
  if (zeroIdx >= 0) {
    ElMessage.warning(`第 ${zeroIdx + 1} 页翻页时长为 0，请填写大于 0 的值`)
    return
  }
  pageDurationTarget.value.coursewarePages = JSON.stringify(arr)
  ElMessage.success('课件时长已保存，提交课程后生效')
  pageDurationVisible.value = false
}

// ---------- 导入打点 ----------
const durationImportRef = ref<HTMLInputElement | null>(null)
const durationImporting = ref(false)
async function onDownloadDurationTemplate() {
  try {
    await downloadDurationTemplate()
    ElMessage.success('模板已下载')
  } catch {
    /* handled in interceptor */
  }
}
function onDurationImportPick(e: Event) {
  const target = e.target as HTMLInputElement
  if (!target.files || !target.files.length) return
  const file = target.files[0]
  target.value = '' // 重置，允许重复选同一文件
  if (!pageDurationTarget.value) return
  const duration = pageDurationTarget.value.duration || 0
  if (duration <= 0) {
    ElMessage.warning('视频时长为 0，无法导入打点')
    return
  }
  durationImporting.value = true
  importDurations(file, duration)
    .then((res) => {
      const durations = res.data.durations || []
      if (durations.length === 0) {
        ElMessage.warning('未解析到有效打点')
        return
      }
      // 填充到表格：按页码顺序对应，多余忽略，不足补 0
      for (let i = 0; i < pageDurations.length; i++) {
        pageDurations[i].duration = i < durations.length ? durations[i] : 0
      }
      ElMessage.success(`已导入 ${durations.length} 页打点`)
    })
    .catch(() => {
      /* handled in interceptor */
    })
    .finally(() => {
      durationImporting.value = false
    })
}
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
.videos-wrap {
  width: 100%;
}
.videos-table {
  margin-top: 10px;
}
.file-name {
  margin-left: 10px;
  color: #888;
  font-size: 13px;
}
.file-size {
  margin-left: 8px;
  color: #909399;
  font-size: 12px;
}
.format-tip {
  margin-left: 10px;
  color: #e6a23c;
  font-size: 12px;
}
.upload-progress {
  width: 100%;
  .progress-head {
    display: flex;
    align-items: center;
    justify-content: space-between;
    margin-bottom: 6px;
  }
  .progress-label {
    font-size: 13px;
    color: #303133;
  }
  .progress-percent {
    font-size: 13px;
    color: #409eff;
    font-weight: 600;
    font-variant-numeric: tabular-nums;
  }
  .progress-hint {
    margin-top: 6px;
    font-size: 12px;
    color: #909399;
  }
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
.pptx-pages {
  font-size: 11px;
  color: #909399;
  margin-top: 2px;
}
.page-duration {
  .duration-summary {
    display: flex;
    align-items: center;
    gap: 8px;
    flex-wrap: wrap;
    margin-bottom: 8px;
  }
  .duration-actions {
    display: flex;
    align-items: center;
    gap: 8px;
    flex-wrap: wrap;
    margin-bottom: 4px;
  }
  .duration-tip {
    font-size: 12px;
    color: #909399;
  }
  .range-text {
    font-family: 'Menlo', 'Consolas', monospace;
    font-size: 12px;
    color: #409eff;
  }
}
</style>
