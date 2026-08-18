<template>
  <div class="slide-check">
    <div class="slide-track" ref="trackRef">
      <div class="slide-fill" :style="{ width: offset + 'px' }"></div>
      <span class="slide-tip">{{ done ? '验证通过' : '请按住滑块，拖动到最右端' }}</span>
      <div
        class="slide-btn"
        :class="{ done }"
        :style="{ left: offset + 'px' }"
        @pointerdown="onDown"
      >
        <el-icon v-if="!done"><DArrowRight /></el-icon>
        <el-icon v-else><Check /></el-icon>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onBeforeUnmount } from 'vue'
import { DArrowRight, Check } from '@element-plus/icons-vue'

const emit = defineEmits<{ success: [] }>()

const trackRef = ref<HTMLElement | null>(null)
const offset = ref(0)
const done = ref(false)
let dragging = false
let startX = 0
let startOffset = 0
let maxX = 0
let activeBtn: HTMLElement | null = null

function onMove(e: PointerEvent) {
  if (!dragging) return
  let x = startOffset + (e.clientX - startX)
  if (x < 0) x = 0
  if (x > maxX) x = maxX
  offset.value = x
}

function onUp() {
  if (!dragging) return
  dragging = false
  cleanup()
  // 到达右端（容差 4px）即视为成功
  if (offset.value >= maxX - 4) {
    done.value = true
    offset.value = maxX
    emit('success')
  } else {
    offset.value = 0
  }
}

function cleanup() {
  if (activeBtn) {
    activeBtn.removeEventListener('pointermove', onMove)
    activeBtn.removeEventListener('pointerup', onUp)
    activeBtn.removeEventListener('pointercancel', onUp)
    activeBtn = null
  }
}

function onDown(e: PointerEvent) {
  if (done.value) return
  const track = trackRef.value
  const btn = e.currentTarget as HTMLElement
  if (!track || !btn) return
  btn.setPointerCapture(e.pointerId)
  dragging = true
  startX = e.clientX
  startOffset = offset.value
  maxX = track.clientWidth - btn.offsetWidth
  activeBtn = btn
  btn.addEventListener('pointermove', onMove)
  btn.addEventListener('pointerup', onUp)
  btn.addEventListener('pointercancel', onUp)
}

onBeforeUnmount(cleanup)
</script>

<style scoped lang="scss">
.slide-check {
  user-select: none;
  width: 100%;
}
.slide-track {
  position: relative;
  height: 40px;
  background: #f0f2f5;
  border-radius: 4px;
  overflow: hidden;
  border: 1px solid #dcdfe6;
}
.slide-fill {
  position: absolute;
  left: 0;
  top: 0;
  height: 100%;
  background: linear-gradient(90deg, #67c23a, #95d475);
  transition: none;
}
.slide-tip {
  position: absolute;
  inset: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 13px;
  color: #909399;
  pointer-events: none;
}
.slide-btn {
  position: absolute;
  top: 0;
  left: 0;
  width: 50px;
  height: 100%;
  background: #fff;
  border: 1px solid #dcdfe6;
  border-radius: 4px;
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: grab;
  color: #606266;
  box-shadow: 0 0 4px rgba(0, 0, 0, 0.1);
  touch-action: none;
}
.slide-btn:active {
  cursor: grabbing;
}
.slide-btn.done {
  color: #67c23a;
  border-color: #67c23a;
}
</style>
