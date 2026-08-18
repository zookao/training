<template>
  <div class="classes-page">
    <div class="page-header">
      <h2 class="page-title">我的班级</h2>
      <p class="page-desc">在这里查看你已加入的班级，并学习课程</p>
    </div>

    <div v-loading="loading" class="class-grid">
      <el-empty v-if="!loading && list.length === 0" description="还没有加入任何班级" />
      <el-card
        v-for="item in list"
        :key="item.id"
        shadow="hover"
        class="class-card"
        @click="goDetail(item.id)"
      >
        <div class="card-body">
          <div class="class-cover">
            <img v-if="item.cover" :src="authUrl(item.cover)" :alt="item.name" />
            <div v-else class="cover-placeholder">
              <el-icon :size="28"><School /></el-icon>
            </div>
          </div>
          <div class="class-name">{{ item.name }}</div>
          <div class="class-desc">{{ item.description || '暂无描述' }}</div>
          <div class="class-meta">
            <span>课程数：{{ item.courseCount }}</span>
          </div>
          <div class="progress-wrap">
            <div class="progress-label">
              <span>学习进度</span>
              <span class="percent">{{ item.percent }}%</span>
            </div>
            <el-progress :percentage="item.percent" :stroke-width="8" :show-text="false" />
          </div>
        </div>
      </el-card>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { School } from '@element-plus/icons-vue'
import { getMyClasses, type ClassProgressItem } from '@/api/learning'
import { authUrl } from '@/utils/authUrl'

const router = useRouter()
const loading = ref(false)
const list = ref<ClassProgressItem[]>([])

async function loadList() {
  loading.value = true
  try {
    const res = await getMyClasses()
    list.value = res.data || []
  } finally {
    loading.value = false
  }
}

function goDetail(id: number) {
  router.push(`/classes/${id}`)
}

onMounted(() => {
  loadList()
})
</script>

<style scoped lang="scss">
.classes-page {
  max-width: 1100px;
  margin: 0 auto;
}
.page-header {
  margin-bottom: 20px;
  .page-title {
    margin: 0 0 6px;
    font-size: 22px;
  }
  .page-desc {
    margin: 0;
    color: #909399;
    font-size: 13px;
  }
}
.class-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(300px, 1fr));
  gap: 16px;
}
.class-card {
  cursor: pointer;
  transition: transform 0.2s;
  &:hover {
    transform: translateY(-2px);
  }
}
.card-body {
  display: flex;
  flex-direction: column;
  gap: 10px;
}
.class-cover {
  width: 100%;
  aspect-ratio: 16 / 9;
  border-radius: 4px;
  overflow: hidden;
  background: #f0f2f5;
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
.class-name {
  font-size: 17px;
  font-weight: 600;
  color: #303133;
}
.class-desc {
  color: #909399;
  font-size: 13px;
  min-height: 40px;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}
.class-meta {
  font-size: 13px;
  color: #606266;
}
.progress-wrap {
  .progress-label {
    display: flex;
    justify-content: space-between;
    font-size: 12px;
    color: #606266;
    margin-bottom: 4px;
    .percent {
      color: #409eff;
      font-weight: 600;
    }
  }
}
</style>
