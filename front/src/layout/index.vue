<template>
  <el-container class="app-wrapper">
    <el-header class="header">
      <div class="logo" @click="router.push('/classes')">
        <img :src="siteStore.logoSrc || defaultLogo" class="logo-img" alt="logo" />
        <span class="title">{{ siteStore.siteName }}</span>
      </div>
      <div class="nav">
        <router-link to="/classes" class="nav-item">我的班级</router-link>
        <router-link to="/profile" class="nav-item">个人中心</router-link>
      </div>
      <el-dropdown @command="handleCommand">
        <span class="user-info">
          <el-avatar :size="30" :src="userInfo?.avatar">{{ (userInfo?.nickname || 'U').charAt(0) }}</el-avatar>
          <span class="username">{{ userInfo?.nickname || userInfo?.username }}</span>
          <el-icon><ArrowDown /></el-icon>
        </span>
        <template #dropdown>
          <el-dropdown-menu>
            <el-dropdown-item command="profile">个人中心</el-dropdown-item>
            <el-dropdown-item command="logout" divided>退出登录</el-dropdown-item>
          </el-dropdown-menu>
        </template>
      </el-dropdown>
    </el-header>
    <el-main class="app-main">
      <router-view />
    </el-main>
  </el-container>
</template>

<script setup lang="ts">
import { computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useUserStore } from '@/stores/user'
import { useSiteStore } from '@/stores/site'
import { ArrowDown } from '@element-plus/icons-vue'
import { ElMessageBox } from 'element-plus'
import defaultLogo from '@/assets/logo.svg'

const router = useRouter()
const userStore = useUserStore()
const siteStore = useSiteStore()
const userInfo = computed(() => userStore.userInfo)

async function handleCommand(cmd: string) {
  if (cmd === 'profile') {
    router.push('/profile')
  } else if (cmd === 'logout') {
    await ElMessageBox.confirm('确定退出登录吗？', '提示', { type: 'warning' })
    await userStore.logout()
    router.push('/login')
  }
}

onMounted(async () => {
  if (!userStore.userInfo) {
    try {
      await userStore.fetchUserInfo()
    } catch (e) {
      /* ignore */
    }
  }
})
</script>

<style scoped lang="scss">
.app-wrapper {
  height: 100%;
  flex-direction: column;
}
.header {
  background: #fff;
  border-bottom: 1px solid #ebeef5;
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 24px;
  height: 56px;
}
.logo {
  display: flex;
  align-items: center;
  gap: 8px;
  cursor: pointer;
  color: #409eff;
  .logo-img {
    width: 32px;
    height: 32px;
    border-radius: 6px;
  }
  .title {
    font-size: 18px;
    font-weight: 600;
  }
}
.nav {
  flex: 1;
  display: flex;
  justify-content: center;
  gap: 24px;
  .nav-item {
    color: #303133;
    text-decoration: none;
    font-size: 14px;
    padding: 6px 12px;
    border-radius: 4px;
    transition: all 0.2s;
    &:hover {
      color: #409eff;
      background: #ecf5ff;
    }
    &.router-link-active {
      color: #409eff;
      font-weight: 500;
    }
  }
}
.user-info {
  display: flex;
  align-items: center;
  gap: 8px;
  cursor: pointer;
}
.username {
  font-size: 14px;
}
.app-main {
  background: #f5f7fa;
  padding: 20px;
}
</style>
