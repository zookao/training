<template>
  <div class="class-detail-page" v-loading="loading">
    <div class="page-header">
      <el-button link :icon="ArrowLeft" @click="router.back()">返回</el-button>
      <h2 class="page-title">{{ detail?.name || '班级详情' }}</h2>
    </div>

    <el-card v-if="detail" shadow="never" class="info-card">
      <div class="class-info">
        <div class="info-text">
          <p class="desc">{{ detail.description || '暂无描述' }}</p>
          <div class="meta">
            <span>课程总数：{{ detail.courses.length }}</span>
          </div>
        </div>
        <div class="overall-progress">
          <el-progress type="dashboard" :percentage="detail.percent" :width="100" />
          <span class="overall-label">总完成度</span>
        </div>
      </div>
    </el-card>

    <div class="course-list" v-if="detail">
      <el-empty v-if="detail.courses.length === 0" description="该班级暂无课程" />
      <el-card
        v-for="c in detail.courses"
        :key="c.id"
        shadow="hover"
        class="course-card"
      >
        <div class="course-body">
          <div class="course-cover">
            <img v-if="c.cover" :src="authUrl(c.cover)" :alt="c.title" />
            <div v-else class="cover-placeholder">
              <el-icon :size="24"><Film /></el-icon>
            </div>
          </div>
          <div class="course-main">
            <div class="course-title">{{ c.title }}</div>
            <div class="course-desc">{{ c.description || '暂无描述' }}</div>
            <div class="course-meta">
              <el-tag size="small" type="info">视频数：{{ c.videoCount }}</el-tag>
              <el-tag size="small" :type="c.completedVideos === c.videoCount && c.videoCount > 0 ? 'success' : 'warning'">
                已完成：{{ c.completedVideos }}/{{ c.videoCount }}
              </el-tag>
            </div>
          </div>
          <div class="course-side">
            <div class="side-progress">
              <span class="pct">{{ c.percent }}%</span>
              <el-progress :percentage="c.percent" :stroke-width="6" :show-text="false" />
            </div>
            <el-button type="primary" @click="goLearn(c.id)">进入学习</el-button>
          </div>
        </div>
      </el-card>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ArrowLeft, Film } from '@element-plus/icons-vue'
import { getClassDetail, type ClassDetailRes } from '@/api/learning'
import { authUrl } from '@/utils/authUrl'

const route = useRoute()
const router = useRouter()
const loading = ref(false)
const detail = ref<ClassDetailRes | null>(null)
const classId = Number(route.params.id)

async function loadDetail() {
  loading.value = true
  try {
    const res = await getClassDetail(classId)
    detail.value = res.data
  } finally {
    loading.value = false
  }
}

function goLearn(courseId: number) {
  router.push({ path: `/course/${courseId}`, query: { classId: String(classId) } })
}

onMounted(() => {
  loadDetail()
})
</script>

<style scoped lang="scss">
.class-detail-page {
  max-width: 1100px;
  margin: 0 auto;
}
.page-header {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 16px;
  .page-title {
    margin: 0;
    font-size: 22px;
  }
}
.info-card {
  margin-bottom: 20px;
}
.class-info {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 24px;
}
.info-text {
  flex: 1;
  .desc {
    margin: 0 0 8px;
    color: #606266;
  }
  .meta {
    font-size: 13px;
    color: #909399;
  }
}
.overall-progress {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 6px;
  .overall-label {
    font-size: 13px;
    color: #606266;
  }
}
.course-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
}
.course-card {
  .course-body {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 20px;
  }
}
.course-cover {
  width: 160px;
  height: 90px;
  border-radius: 4px;
  overflow: hidden;
  background: #f0f2f5;
  flex-shrink: 0;
  img {
    width: 100%;
    height: 100%;
    object-fit: cover;
    display: block;
  }
  .cover-placeholder {
    width: 100%;
    height: 100%;
    display: flex;
    align-items: center;
    justify-content: center;
    color: #c0c4cc;
  }
}
.course-main {
  flex: 1;
  min-width: 0;
  .course-title {
    font-size: 16px;
    font-weight: 600;
    margin-bottom: 6px;
  }
  .course-desc {
    color: #909399;
    font-size: 13px;
    margin-bottom: 10px;
  }
  .course-meta {
    display: flex;
    gap: 8px;
  }
}
.course-side {
  display: flex;
  flex-direction: column;
  align-items: flex-end;
  gap: 10px;
  min-width: 160px;
  .side-progress {
    width: 160px;
    text-align: right;
    .pct {
      display: block;
      font-size: 13px;
      color: #409eff;
      font-weight: 600;
      margin-bottom: 4px;
    }
  }
}
</style>
