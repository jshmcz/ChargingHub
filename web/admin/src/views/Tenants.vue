<template>
  <div class="page-container">
    <div class="page-header">
      <h2 class="page-title">租户管理</h2>
      <el-button type="primary" @click="openCreate">新建租户</el-button>
    </div>

    <div class="table-wrap">
      <el-table :data="pagedList" border v-loading="loading" min-width="600">
        <el-table-column prop="id" label="ID" width="90" />
        <el-table-column prop="name" label="租户名称" min-width="180" />
        <el-table-column label="状态" width="120">
          <template #default="{ row }">
            <el-tag :type="row.status === 'active' ? 'success' : 'info'">
              {{ statusText(row.status) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="创建时间" min-width="180">
          <template #default="{ row }">{{ formatDate(row.createdAt) }}</template>
        </el-table-column>
        <el-table-column label="操作" width="120" fixed="right">
          <template #default="{ row }">
            <el-button link type="primary" @click="openEdit(row)">编辑</el-button>
          </template>
        </el-table-column>
      </el-table>
    </div>

    <div class="pagination-wrapper">
      <el-pagination
        background
        layout="total, prev, pager, next"
        :total="list.length"
        v-model:current-page="page"
        :page-size="pageSize"
      />
    </div>

    <el-dialog v-model="dialogVisible" :title="editing ? '编辑租户' : '新建租户'" width="520px">
      <el-form ref="formRef" :model="form" :rules="rules" label-width="96px">
        <el-form-item label="租户名称" prop="name">
          <el-input v-model="form.name" placeholder="请输入租户名称" />
        </el-form-item>
        <template v-if="!editing">
          <el-divider content-position="left">管理员账号（该租户登录后自动归属）</el-divider>
          <el-form-item label="管理员姓名" prop="adminName">
            <el-input v-model="form.adminName" placeholder="如：张房东" />
          </el-form-item>
          <el-form-item label="登录手机号" prop="adminPhone">
            <el-input v-model="form.adminPhone" placeholder="管理员用此手机号登录" />
          </el-form-item>
          <el-form-item label="邮箱" prop="adminEmail">
            <el-input v-model="form.adminEmail" placeholder="用于接收充电通知（选填）" />
          </el-form-item>
          <el-form-item label="初始密码" prop="adminPassword">
            <el-input v-model="form.adminPassword" type="password" show-password placeholder="请设置初始密码" />
          </el-form-item>
        </template>
        <el-form-item v-else label="状态" prop="status">
          <el-select v-model="form.status" style="width: 100%">
            <el-option label="启用" value="active" />
            <el-option label="停用" value="disabled" />
          </el-select>
        </el-form-item>
      </el-form>
      <template #footer>
        <div class="dialog-footer">
          <el-button @click="dialogVisible = false">取消</el-button>
          <el-button type="primary" :loading="saving" @click="save">保存</el-button>
        </div>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { ElMessage, type FormInstance, type FormRules } from 'element-plus'
import { tenantsApi } from '@/api/tenants'
import type { Tenant } from '@/api/types'
import { formatDate } from '@/utils/format'

const loading = ref(false)
const list = ref<Tenant[]>([])
const page = ref(1)
const pageSize = 10
const pagedList = computed(() =>
  list.value.slice((page.value - 1) * pageSize, page.value * pageSize),
)

function errorText(error: unknown): string {
  return error instanceof Error ? error.message : '操作失败'
}

async function load() {
  loading.value = true
  try {
    list.value = await tenantsApi.list()
  } catch (error) {
    ElMessage.error(errorText(error))
  } finally {
    loading.value = false
  }
}

onMounted(load)

const dialogVisible = ref(false)
const editing = ref<Tenant | null>(null)
const saving = ref(false)
const formRef = ref<FormInstance>()
const form = reactive({
  name: '',
  status: 'active',
  adminName: '',
  adminPhone: '',
  adminEmail: '',
  adminPassword: '',
})

const rules: FormRules = {
  name: [{ required: true, message: '请输入租户名称', trigger: 'blur' }],
  adminName: [{ required: true, message: '请输入管理员姓名', trigger: 'blur' }],
  adminPhone: [{ required: true, message: '请输入管理员登录手机号', trigger: 'blur' }],
  adminPassword: [{ required: true, message: '请设置初始密码', trigger: 'blur' }],
}

function openCreate() {
  editing.value = null
  Object.assign(form, { name: '', status: 'active', adminName: '', adminPhone: '', adminEmail: '', adminPassword: '' })
  dialogVisible.value = true
}

function openEdit(row: Tenant) {
  editing.value = row
  Object.assign(form, { name: row.name, status: row.status || 'active' })
  dialogVisible.value = true
}

async function save() {
  if (!formRef.value) return
  try {
    await formRef.value.validate()
  } catch {
    return
  }
  saving.value = true
  try {
    if (editing.value) {
      await tenantsApi.update(editing.value.id, { name: form.name, status: form.status })
      ElMessage.success('保存成功')
    } else {
      const created = await tenantsApi.create({
        name: form.name,
        status: form.status,
        adminName: form.adminName,
        adminPhone: form.adminPhone,
        adminEmail: form.adminEmail,
        adminPassword: form.adminPassword,
      })
      ElMessage.success(`租户创建成功，管理员 ${created.adminPhone || form.adminPhone} 可登录管理该租户`)
    }
    dialogVisible.value = false
    await load()
  } catch (error) {
    ElMessage.error(errorText(error))
  } finally {
    saving.value = false
  }
}

function statusText(status: string): string {
  if (status === 'active') return '启用'
  if (status === 'disabled') return '停用'
  return status
}
</script>
