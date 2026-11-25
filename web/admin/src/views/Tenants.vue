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

    <el-dialog v-model="dialogVisible" :title="editing ? '编辑租户' : '新建租户'" width="440px">
      <el-form ref="formRef" :model="form" :rules="rules" label-width="80px">
        <el-form-item label="名称" prop="name">
          <el-input v-model="form.name" placeholder="请输入租户名称" />
        </el-form-item>
        <el-form-item label="状态" prop="status">
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
const form = reactive({ name: '', status: 'active' })

const rules: FormRules = {
  name: [{ required: true, message: '请输入租户名称', trigger: 'blur' }],
}

function openCreate() {
  editing.value = null
  Object.assign(form, { name: '', status: 'active' })
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
    } else {
      await tenantsApi.create({ name: form.name, status: form.status })
    }
    ElMessage.success('保存成功')
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
