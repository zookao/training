<template>
  <div class="dashboard">
    <el-row :gutter="16">
      <el-col :span="6" v-for="c in cards" :key="c.title">
        <el-card shadow="hover" class="stat-card">
          <div class="stat-icon" :style="{ background: c.color }">
            <el-icon :size="28"><component :is="c.icon" /></el-icon>
          </div>
          <div class="stat-info">
            <div class="stat-value">{{ c.value }}</div>
            <div class="stat-title">{{ c.title }}</div>
            <div v-if="c.sub" class="stat-sub">{{ c.sub }}</div>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <el-row :gutter="16" style="margin-top: 16px">
      <el-col :span="10">
        <el-card shadow="never" class="panel">
          <template #header>
            <div class="panel-header">学习概览</div>
          </template>
          <div class="learn-metrics">
            <div class="metric">
              <div class="metric-value">{{ data?.learnerCount ?? 0 }}</div>
              <div class="metric-label">学习人数</div>
            </div>
            <div class="metric">
              <div class="metric-value">{{ data?.recordCount ?? 0 }}</div>
              <div class="metric-label">学习记录数</div>
            </div>
            <div class="metric">
              <div class="metric-value">{{ data?.completedCount ?? 0 }}</div>
              <div class="metric-label">已完成视频数</div>
            </div>
            <div class="metric">
              <div class="metric-value">{{ data?.avgProgress ?? 0 }}%</div>
              <div class="metric-label">平均学习进度</div>
            </div>
            <div class="metric">
              <div class="metric-value">{{ data?.completionRate ?? 0 }}%</div>
              <div class="metric-label">完成率</div>
            </div>
            <div class="metric">
              <div class="metric-value">{{ data?.todayActive ?? 0 }}</div>
              <div class="metric-label">今日活跃学员</div>
            </div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="14">
        <el-card shadow="never" class="panel">
          <template #header>
            <div class="panel-header">最近学习记录</div>
          </template>
          <el-table
            :data="data?.recentRecords || []"
            size="small"
            border
            max-height="320"
            empty-text="暂无学习记录"
          >
            <el-table-column label="学员" min-width="120">
              <template #default="{ row }">{{ row.nickname || row.username || '—' }}</template>
            </el-table-column>
            <el-table-column prop="courseTitle" label="课程" min-width="160" show-overflow-tooltip />
            <el-table-column label="进度" width="170">
              <template #default="{ row }">
                <el-progress :percentage="row.percent" :stroke-width="8" :status="row.completed ? 'success' : ''" />
              </template>
            </el-table-column>
            <el-table-column label="状态" width="80" align="center">
              <template #default="{ row }">
                <el-tag :type="row.completed ? 'success' : 'warning'" size="small">
                  {{ row.completed ? '已完成' : '学习中' }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="lastAt" label="最近学习时间" width="160" />
          </el-table>
        </el-card>
      </el-col>
    </el-row>

    <el-card class="welcome" shadow="never" style="margin-top: 16px">
      <h2>欢迎使用培训管理后台</h2>
      <p>当前登录：<strong>{{ userInfo?.nickname || userInfo?.username }}</strong></p>
      <p>角色：<el-tag v-for="r in userStore.roles" :key="r" size="small" style="margin-right:8px">{{ r }}</el-tag></p>
      <p style="color:#909399;font-size:13px">
        通过左侧菜单管理「课程 / 班级 / 学员 / 管理员 / 角色 / 菜单 / 接口权限」。
      </p>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { School, Reading, User, TrendCharts } from '@element-plus/icons-vue'
import { useUserStore } from '@/stores/user'
import { getDashboard, type DashboardData } from '@/api/dashboard'

const userStore = useUserStore()
const userInfo = computed(() => userStore.userInfo)
const data = ref<DashboardData | null>(null)

const cards = computed(() => [
  { title: '班级总数', value: data.value?.classCount ?? '-', sub: data.value ? `启用 ${data.value.activeClass}` : '', icon: School, color: '#409EFF' },
  { title: '课程总数', value: data.value?.courseCount ?? '-', sub: data.value ? `启用 ${data.value.activeCourse}` : '', icon: Reading, color: '#67C23A' },
  { title: '学员总数', value: data.value?.studentCount ?? '-', sub: data.value ? `启用 ${data.value.activeStudent}` : '', icon: User, color: '#E6A23C' },
  { title: '学习记录', value: data.value?.recordCount ?? '-', sub: data.value ? `已完成 ${data.value.completedCount}` : '', icon: TrendCharts, color: '#F56C6C' }
])

async function loadData() {
  try {
    const res = await getDashboard()
    data.value = res.data
  } catch (e) {
    /* handled in interceptor */
  }
}

onMounted(() => {
  loadData()
})
</script>

<style scoped lang="scss">
.stat-card {
  :deep(.el-card__body) {
    display: flex;
    align-items: center;
    gap: 16px;
  }
}
.stat-icon {
  width: 56px;
  height: 56px;
  border-radius: 8px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #fff;
  flex-shrink: 0;
}
.stat-info {
  min-width: 0;
}
.stat-value {
  font-size: 26px;
  font-weight: 600;
  line-height: 1.2;
}
.stat-title {
  color: #909399;
  font-size: 13px;
}
.stat-sub {
  color: #c0c4cc;
  font-size: 12px;
  margin-top: 2px;
}
.panel {
  height: 100%;
}
.panel-header {
  font-weight: 600;
}
.learn-metrics {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 18px 0;
  padding: 12px 0 6px;
}
.metric {
  text-align: center;
  min-width: 0;
  .metric-value {
    font-size: 28px;
    font-weight: 600;
    color: #409eff;
    line-height: 1.2;
    // 数字超长时防溢出：优先换行兜底，超长降字号
    overflow-wrap: break-word;
    @media (max-width: 1400px) {
      font-size: 24px;
    }
  }
  .metric-label {
    font-size: 13px;
    color: #909399;
    margin-top: 4px;
  }
}
.welcome {
  h2 {
    margin-top: 0;
  }
}
</style>
