<template>
  <div class="header-wrap">
    <div class="left">
      <el-icon class="collapse-btn" @click="toggle">
        <Fold v-if="!collapse" />
        <Expand v-else />
      </el-icon>
      <el-breadcrumb separator="/">
        <el-breadcrumb-item :to="{ path: '/dashboard/index' }">首页</el-breadcrumb-item>
        <el-breadcrumb-item v-if="currentTitle">{{ currentTitle }}</el-breadcrumb-item>
      </el-breadcrumb>
    </div>
    <div class="right">
      <el-dropdown @command="handleCommand">
        <span class="user-info">
          <el-avatar :size="28" :src="userInfo?.avatar">
            {{ (userInfo?.nickname || 'A').charAt(0) }}
          </el-avatar>
          <span class="username">{{ userInfo?.nickname || userInfo?.username }}</span>
          <el-icon><ArrowDown /></el-icon>
        </span>
        <template #dropdown>
          <el-dropdown-menu>
            <el-dropdown-item command="dashboard">首页</el-dropdown-item>
            <el-dropdown-item command="logout" divided>退出登录</el-dropdown-item>
          </el-dropdown-menu>
        </template>
      </el-dropdown>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useUserStore } from '@/stores/user'
import { useMenuStore } from '@/stores/menu'
import { ElMessageBox } from 'element-plus'

const props = defineProps<{ collapse: boolean }>()
const emit = defineEmits<{ (e: 'update:collapse', v: boolean): void }>()

const route = useRoute()
const router = useRouter()
const userStore = useUserStore()
const menuStore = useMenuStore()
const userInfo = computed(() => userStore.userInfo)
const currentTitle = computed(() => (route.meta?.title as string) || '')

function toggle() {
  emit('update:collapse', !props.collapse)
}

async function handleCommand(cmd: string) {
  if (cmd === 'dashboard') {
    router.push('/dashboard/index')
  } else if (cmd === 'logout') {
    await ElMessageBox.confirm('确定退出登录吗？', '提示', { type: 'warning' })
    await userStore.logout()
    menuStore.reset()
    router.push('/login')
  }
}
</script>

<style scoped lang="scss">
.header-wrap {
  width: 100%;
  display: flex;
  align-items: center;
  justify-content: space-between;
}
.left {
  display: flex;
  align-items: center;
  gap: 16px;
}
.collapse-btn {
  font-size: 20px;
  cursor: pointer;
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
</style>
