<template>
  <div class="page-container">
    <div class="page-header">
      <h2 class="page-title">平台配置</h2>
    </div>

    <el-alert
      type="info"
      :closable="false"
      show-icon
      style="margin-bottom: 20px"
      title="对外域名用于生成充电桩二维码。二维码一旦打印，改域名后旧二维码将失效，请先用正式域名配置再下载二维码。"
    />

    <div class="config-card">
      <el-form label-width="140px">
        <el-form-item label="对外域名">
          <el-input v-model="baseUrl" placeholder="如 https://ff.dadaf.com:9010" style="max-width: 460px" />
          <div class="text-muted" style="font-size: 12px; margin-top: 4px">
            二维码内容 = 域名 + /m/#/charge?token=枪令牌。留空则不能下载二维码。
          </div>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" :loading="saving" @click="save">保存</el-button>
        </el-form-item>
      </el-form>
    </div>
  </div>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { ElMessage } from 'element-plus'
import { platformApi } from '@/api/platform'

const baseUrl = ref('')
const saving = ref(false)

onMounted(() => {
  platformApi.get().then((cfg) => { baseUrl.value = cfg.baseUrl || '' }).catch(() => {})
})

async function save() {
  saving.value = true
  try {
    const cfg = await platformApi.update({ baseUrl: baseUrl.value.trim() })
    baseUrl.value = cfg.baseUrl
    ElMessage.success('平台配置已保存')
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : '保存失败')
  } finally {
    saving.value = false
  }
}
</script>

<style scoped>
.config-card {
  background: var(--el-bg-color, #fff);
  border: 1px solid var(--el-border-color-lighter, #ebeef5);
  border-radius: 8px;
  padding: 24px;
  max-width: 720px;
}
</style>
