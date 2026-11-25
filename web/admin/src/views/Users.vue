<template>
  <div class="page-container">
    <div class="page-header">
      <h2 class="page-title">用户管理</h2>
      <el-button type="primary" @click="openCreate">新建用户</el-button>
    </div>

    <div class="table-toolbar">
      <el-input
        v-model="keyword"
        placeholder="按手机号 / 邮箱 / 姓名搜索"
        clearable
        style="width: 260px"
        @keyup.enter="load"
      />
      <el-button @click="load">搜索</el-button>
    </div>

    <div class="table-wrap">
      <el-table :data="pagedList" border v-loading="loading" min-width="820">
        <el-table-column prop="id" label="ID" width="80" />
        <el-table-column prop="name" label="姓名" min-width="120" />
        <el-table-column prop="phone" label="手机号" min-width="140" />
        <el-table-column prop="email" label="邮箱" min-width="180" />
        <el-table-column v-if="auth.isSuper" prop="tenantName" label="所属租户" min-width="140" />
        <el-table-column v-if="auth.isSuper" prop="role" label="角色" width="130" />
        <el-table-column label="钱包余额" width="130">
          <template #default="{ row }">{{ formatCents(row.balanceCents) }}</template>
        </el-table-column>
        <el-table-column label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="row.status === 'disabled' ? 'danger' : 'success'">
              {{ statusText(row.status) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="180" fixed="right">
          <template #default="{ row }">
            <el-button link type="primary" @click="openEdit(row)">编辑</el-button>
            <el-button
              v-if="row.status !== 'disabled'"
              link
              type="danger"
              @click="disableUser(row)"
            >
              禁用
            </el-button>
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

    <el-dialog v-model="createVisible" title="新建用户" width="480px">
      <el-form ref="createFormRef" :model="createForm" :rules="createRules" label-width="90px">
        <el-form-item label="姓名" prop="name">
          <el-input v-model="createForm.name" />
        </el-form-item>
        <el-form-item label="手机号" prop="phone">
          <el-input v-model="createForm.phone" />
        </el-form-item>
        <el-form-item label="邮箱" prop="email">
          <el-input v-model="createForm.email" />
        </el-form-item>
        <el-form-item label="密码" prop="password">
          <el-input v-model="createForm.password" type="password" show-password />
        </el-form-item>
        <template v-if="auth.isSuper">
          <el-form-item label="角色" prop="role">
            <el-select v-model="createForm.role" style="width: 100%">
              <el-option label="租户管理员" value="tenant_admin" />
              <el-option label="操作员" value="operator" />
              <el-option label="充电用户" value="user" />
            </el-select>
          </el-form-item>
          <el-form-item label="所属租户" prop="tenantId">
            <el-select v-model="createForm.tenantId" style="width: 100%" filterable placeholder="选择租户">
              <el-option v-for="t in tenants" :key="t.id" :label="t.name" :value="String(t.id)" />
            </el-select>
          </el-form-item>
        </template>
      </el-form>
      <template #footer>
        <div class="dialog-footer">
          <el-button @click="createVisible = false">取消</el-button>
          <el-button type="primary" :loading="saving" @click="create">确定</el-button>
        </div>
      </template>
    </el-dialog>

    <el-dialog v-model="editVisible" title="编辑用户" width="460px">
      <el-form ref="editFormRef" :model="editForm" :rules="editRules" label-width="80px">
        <el-form-item label="姓名" prop="name">
          <el-input v-model="editForm.name" />
        </el-form-item>
        <el-form-item label="邮箱" prop="email">
          <el-input v-model="editForm.email" />
        </el-form-item>
        <el-form-item label="状态" prop="status">
          <el-select v-model="editForm.status" style="width: 100%">
            <el-option label="启用" value="active" />
            <el-option label="停用" value="disabled" />
          </el-select>
        </el-form-item>
      </el-form>
      <template #footer>
        <div class="dialog-footer">
          <el-button @click="editVisible = false">取消</el-button>
          <el-button type="primary" :loading="saving" @click="update">保存</el-button>
        </div>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { ElMessage, ElMessageBox, type FormInstance, type FormRules } from 'element-plus'
import { usersApi } from '@/api/users'
import { tenantsApi } from '@/api/tenants'
import type { CreateUserPayload } from '@/api/users'
import type { Role, User } from '@/api/types'
import { formatCents } from '@/utils/format'
import { useAuthStore } from '@/stores/auth'

const auth = useAuthStore()
const loading = ref(false)
const keyword = ref('')
const list = ref<User[]>([])
const tenants = ref<{ id: number; name: string }[]>([])
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
    list.value = await usersApi.list(keyword.value || undefined)
    page.value = 1
  } catch (error) {
    ElMessage.error(errorText(error))
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  load()
  if (auth.isSuper) {
    tenantsApi.list().then((rows) => { tenants.value = rows }).catch(() => {})
  }
})

const saving = ref(false)
const createVisible = ref(false)
const createFormRef = ref<FormInstance>()
const createForm = reactive({
  name: '',
  phone: '',
  email: '',
  password: '',
  role: 'tenant_admin' as Role,
  tenantId: '',
})

const createRules: FormRules = {
  name: [{ required: true, message: '请输入姓名', trigger: 'blur' }],
  phone: [{ required: true, message: '请输入手机号', trigger: 'blur' }],
  email: [{ required: true, type: 'email', message: '请输入正确邮箱', trigger: 'blur' }],
  password: [{ required: true, message: '请输入密码', trigger: 'blur' }],
  tenantId: [
    {
      validator: (_rule, value, callback) => {
        if (auth.isSuper && !value) callback(new Error('超管创建用户时需填写租户ID'))
        else callback()
      },
      trigger: 'blur',
    },
  ],
}

function openCreate() {
  Object.assign(createForm, {
    name: '',
    phone: '',
    email: '',
    password: '',
    role: 'tenant_admin' as Role,
    tenantId: '',
  })
  createVisible.value = true
}

async function create() {
  if (!createFormRef.value) return
  try {
    await createFormRef.value.validate()
  } catch {
    return
  }
  saving.value = true
  try {
    const payload: CreateUserPayload = {
      name: createForm.name,
      phone: createForm.phone,
      email: createForm.email,
      password: createForm.password,
    }
    if (auth.isSuper) {
      payload.role = createForm.role
      payload.tenantId = Number(createForm.tenantId)
    }
    await usersApi.create(payload)
    ElMessage.success('创建成功')
    createVisible.value = false
    await load()
  } catch (error) {
    ElMessage.error(errorText(error))
  } finally {
    saving.value = false
  }
}

const editVisible = ref(false)
const editing = ref<User | null>(null)
const editFormRef = ref<FormInstance>()
const editForm = reactive({ name: '', email: '', status: 'active' })

const editRules: FormRules = {
  name: [{ required: true, message: '请输入姓名', trigger: 'blur' }],
  email: [{ required: true, type: 'email', message: '请输入正确邮箱', trigger: 'blur' }],
}

function openEdit(row: User) {
  editing.value = row
  Object.assign(editForm, { name: row.name, email: row.email, status: row.status || 'active' })
  editVisible.value = true
}

async function update() {
  if (!editFormRef.value || !editing.value) return
  try {
    await editFormRef.value.validate()
  } catch {
    return
  }
  saving.value = true
  try {
    await usersApi.update(editing.value.id, {
      name: editForm.name,
      email: editForm.email,
      status: editForm.status,
    })
    ElMessage.success('保存成功')
    editVisible.value = false
    await load()
  } catch (error) {
    ElMessage.error(errorText(error))
  } finally {
    saving.value = false
  }
}

async function disableUser(row: User) {
  try {
    await ElMessageBox.confirm(`确认禁用用户「${row.name}」？`, '提示', { type: 'warning' })
  } catch {
    return
  }
  try {
    await usersApi.update(row.id, { name: row.name, email: row.email, status: 'disabled' })
    ElMessage.success('已禁用')
    await load()
  } catch (error) {
    ElMessage.error(errorText(error))
  }
}

function statusText(status?: string): string {
  if (status === 'disabled') return '停用'
  return '启用'
}
</script>
