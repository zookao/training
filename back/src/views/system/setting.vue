<template>
  <div class="setting-page">
    <el-card>
      <template #header>系统设置</template>
      <el-form :model="form" label-width="100px" style="max-width: 600px">
        <el-form-item label="站点 Logo">
          <el-upload
            :show-file-list="false"
            :before-upload="beforeLogoUpload"
            :http-request="handleLogoUpload"
            accept="image/png,image/jpeg"
          >
            <div v-if="form.logoUrl" class="logo-preview">
              <el-image :src="authUrl(form.logoUrl)" fit="contain" style="width: 96px; height: 96px" />
              <span class="logo-replace">点击替换</span>
            </div>
            <el-button v-else :loading="uploading">+ 选择 Logo</el-button>
          </el-upload>
          <el-button v-if="form.logoUrl" link type="danger" style="margin-left: 12px" @click="form.logoUrl = ''">移除</el-button>
          <div class="logo-tip">建议尺寸 128×128（正方形），支持 PNG/JPG，不超过 2MB</div>
        </el-form-item>
        <el-form-item label="站点名称">
          <el-input v-model="form.siteName" placeholder="如：培训学习平台" maxlength="30" show-word-limit />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" :loading="saving" @click="handleSave">保存设置</el-button>
        </el-form-item>
      </el-form>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { reactive, ref, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { authUrl } from '@/utils/authUrl'
import { getSiteConfig, updateSiteConfig, uploadLogo } from '@/api/setting'
import { useSiteStore } from '@/stores/site'

const siteStore = useSiteStore()
const uploading = ref(false)
const saving = ref(false)
const form = reactive({ logoUrl: '', siteName: '' })

onMounted(async () => {
  try {
    const res = await getSiteConfig()
    form.logoUrl = res.data.logoUrl
    form.siteName = res.data.siteName
  } catch (e) {
    /* handled in interceptor */
  }
})

function beforeLogoUpload(file: File) {
  if (!file.type.startsWith('image/')) {
    ElMessage.error('请选择图片文件')
    return false
  }
  if (file.size > 2 * 1024 * 1024) {
    ElMessage.error('Logo 不能超过 2MB')
    return false
  }
  return true
}

async function handleLogoUpload(option: { file: File }) {
  uploading.value = true
  try {
    const res = await uploadLogo(option.file)
    form.logoUrl = res.data.url
    ElMessage.success('Logo 上传成功')
  } catch (e) {
    /* handled in interceptor */
  } finally {
    uploading.value = false
  }
}

async function handleSave() {
  if (!form.siteName.trim()) {
    ElMessage.error('站点名称不能为空')
    return
  }
  saving.value = true
  try {
    await updateSiteConfig({ logoUrl: form.logoUrl, siteName: form.siteName.trim() })
    ElMessage.success('保存成功')
    siteStore.load()
  } catch (e) {
    /* handled in interceptor */
  } finally {
    saving.value = false
  }
}
</script>

<style scoped>
.setting-page {
  padding: 16px;
}
.logo-preview {
  position: relative;
  width: 96px;
  height: 96px;
  border: 1px dashed #d9d9d9;
  border-radius: 8px;
  overflow: hidden;
}
.logo-replace {
  position: absolute;
  inset: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  background: rgba(0, 0, 0, 0.5);
  color: #fff;
  font-size: 12px;
  opacity: 0;
  transition: opacity 0.2s;
}
.logo-preview:hover .logo-replace {
  opacity: 1;
}
.logo-tip {
  width: 100%;
  margin-top: 6px;
  font-size: 12px;
  color: #909399;
}
</style>
