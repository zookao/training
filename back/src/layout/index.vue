<template>
  <el-container class="app-wrapper">
    <el-aside :width="isCollapse ? '64px' : '210px'" class="sidebar">
      <Sidebar :collapse="isCollapse" />
    </el-aside>
    <el-container>
      <el-header class="header">
        <Header v-model:collapse="isCollapse" />
      </el-header>
      <el-main class="app-main">
        <AppMain />
      </el-main>
    </el-container>
  </el-container>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import Sidebar from './Sidebar.vue'
import Header from './Header.vue'
import AppMain from './AppMain.vue'

const isCollapse = ref(false)
</script>

<style scoped lang="scss">
.app-wrapper {
  height: 100%;
}
.sidebar {
  background: #304156;
  transition: width 0.28s;
  overflow: hidden;
}
.header {
  background: #fff;
  border-bottom: 1px solid #f0f0f0;
  display: flex;
  align-items: center;
  padding: 0 16px;
}
.app-main {
  background: #f0f2f5;
  padding: 16px;
}
</style>
