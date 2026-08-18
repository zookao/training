<template>
  <!-- 有可见子菜单：渲染为子菜单组 -->
  <el-sub-menu v-if="hasVisibleChildren" :index="basePath">
    <template #title>
      <el-icon v-if="item.icon"><component :is="item.icon" /></el-icon>
      <span>{{ item.name }}</span>
    </template>
    <SidebarItem
      v-for="child in visibleChildren"
      :key="child.id"
      :item="child"
      :base-path="resolvePath(child.path)"
    />
  </el-sub-menu>
  <!-- 无子菜单：渲染为单菜单项 -->
  <el-menu-item v-else :index="resolvePath(singleChild ? singleChild.path : '')">
    <el-icon v-if="iconName"><component :is="iconName" /></el-icon>
    <template #title>{{ singleChild ? singleChild.name : item.name }}</template>
  </el-menu-item>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import type { MenuItem } from '@/api/auth'

const props = defineProps<{ item: MenuItem; basePath: string }>()

const visibleChildren = computed(
  () => (props.item.children || []).filter((c) => c.type !== 'F')
)
const hasVisibleChildren = computed(() => visibleChildren.value.length > 1)
const singleChild = computed(() =>
  visibleChildren.value.length === 1 ? visibleChildren.value[0] : null
)
const iconName = computed(() => (singleChild.value ? singleChild.value.icon : props.item.icon))

function resolvePath(childPath: string): string {
  if (childPath.startsWith('/')) return childPath
  return `${props.basePath}/${childPath}`.replace(/\/+/g, '/')
}
</script>
