<template>
  <div class="page-container">
    <div class="page-header">
      <h2 class="page-title">SMTP 配置</h2>
      <el-button type="primary" :loading="saving" @click="save">保存配置</el-button>
    </div>

    <el-alert
      v-if="!auth.canConfigSmtp"
      type="warning"
      :closable="false"
      show-icon
      title="当前角色无权修改 SMTP 配置"
      style="margin-bottom: 16px"
    />

    <el-card shadow="never">
      <el-form ref="formRef" :model="form" :rules="rules" label-width="110px" style="max-width: 720px">
        <el-form-item label="SMTP 主机" prop="host">
          <el-input v-model="form.host" placeholder="如 smtp.example.com" />
        </el-form-item>
        <el-form-item label="端口" prop="port">
          <el-input-number v-model="form.port" :min="1" :max="65535" />
        </el-form-item>
        <el-form-item label="用户名" prop="username">
          <el-input v-model="form.username" />
        </el-form-item>
        <el-form-item label="密码" prop="password">
          <el-input v-model="form.password" type="password" show-password />
        </el-form-item>
        <el-form-item label="发件地址" prop="from">
          <el-input v-model="form.from" placeholder="如 noreply@example.com" />
        </el-form-item>
        <el-form-item label="加密方式" prop="encryption">
          <el-select v-model="form.encryption" style="width: 100%">
            <el-option label="不加密" value="none" />
            <el-option label="SSL" value="ssl" />
            <el-option label="STARTTLS" value="starttls" />
          </el-select>
        </el-form-item>
        <el-form-item label="启用">
          <el-switch v-model="form.enabled" />
        </el-form-item>
      </el-form>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { ElMessage, type FormInstance, type FormRules } from 'element-plus'
import { smtpApi } from '@/api/smtp'
import type { SmtpConfig } from '@/api/types'
import { useAuthStore } from '@/stores/auth'

const auth = useAuthStore()
const saving = ref(false)
const formRef = ref<FormInstance>()
const form = reactive<SmtpConfig>({
  host: '',
  port: 465,
  username: '',
  password: '',
  from: '',
  encryption: 'ssl',
  enabled: false,
})

const rules: FormRules = {
  host: [{ required: true, message: '请输入 SMTP 主机', trigger: 'blur' }],
  port: [{ required: true, message: '请输入端口', trigger: 'blur' }],
  username: [{ required: true, message: '请输入用户名', trigger: 'blur' }],
  from: [{ required: true, message: '请输入发件地址', trigger: 'blur' }],
}

function errorText(error: unknown): string {
  return error instanceof Error ? error.message : '操作失败'
}

async function load() {
  try {
    const config = await smtpApi.get()
    if (config) Object.assign(form, config)
  } catch (error) {
    ElMessage.error(errorText(error))
  }
}

onMounted(load)

async function save() {
  if (!auth.canConfigSmtp) {
    ElMessage.warning('当前角色无权修改 SMTP 配置')
    return
  }
  if (!formRef.value) return
  try {
    await formRef.value.validate()
  } catch {
    return
  }
  saving.value = true
  try {
    await smtpApi.update({ ...form })
    ElMessage.success('保存成功')
  } catch (error) {
    ElMessage.error(errorText(error))
  } finally {
    saving.value = false
  }
}
</script>
