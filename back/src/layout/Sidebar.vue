<template>
  <div class="sidebar-container">
    <div class="logo">
      <img :src="siteStore.logoSrc || defaultLogo" class="logo-img" alt="logo" />
      <el-tooltip v-if="!collapse" :content="siteStore.siteName" placement="bottom" :disabled="!titleOverflow">
        <span ref="titleRef" class="logo-title" @mouseenter="checkOverflow">{{ siteStore.siteName }}</span>
      </el-tooltip>
    </div>
    <el-menu
      :default-active="activeMenu"
      :collapse="collapse"
      :collapse-transition="false"
      background-color="#304156"
      text-color="#bfcbd9"
      active-text-color="#409EFF"
      router
      unique-opened
    >
      <el-menu-item index="/dashboard/index">
        <el-icon><HomeFilled /></el-icon>
        <template #title>首页</template>
      </el-menu-item>
      <SidebarItem v-for="m in menus" :key="m.id" :item="m" :base-path="m.path" />
    </el-menu>
  </div>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { useRoute } from 'vue-router'
import { useMenuStore } from '@/stores/menu'
import { useSiteStore } from '@/stores/site'
import defaultLogo from '@/assets/logo.svg'
import SidebarItem from './SidebarItem.vue'

defineProps<{ collapse?: boolean }>()

const route = useRoute()
const menuStore = useMenuStore()
const siteStore = useSiteStore()
const menus = computed(() => menuStore.menus)
const activeMenu = computed(() => route.path)

const titleRef = ref<HTMLElement>()
const titleOverflow = ref(false)
function checkOverflow() {
  if (titleRef.value) {
    titleOverflow.value = titleRef.value.scrollWidth > titleRef.value.clientWidth
  }
}
</script>

<style scoped>
.sidebar-container {
  height: 100%;
  display: flex;
  flex-direction: column;
}
.logo {
  height: 60px;
  display: flex;
  align-items: center;
  padding: 0 16px;
  gap: 10px;
  color: #fff;
  background: #2b3a4d;
  overflow: hidden;
}
.logo-img {
  width: 32px;
  height: 32px;
  border-radius: 6px;
  flex-shrink: 0;
}
.logo-title {
  font-size: 16px;
  font-weight: 600;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  min-width: 0;
}
.el-menu {
  border-right: none;
  flex: 1;
}
</style>
