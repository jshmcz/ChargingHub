<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import {
  Button,
  Cell,
  CellGroup,
  Empty,
  Field,
  Icon,
  Loading,
  NavBar,
  Popup,
  PullRefresh,
  Tag,
  showToast,
} from 'vant'
import { adminApi, type CreateUserPayload } from '@/api/admin'
import { walletApi, type WalletMutationPayload } from '@/api/wallets'
import type { Role, UserInfo } from '@/api/types'
import { useAuthStore } from '@/stores/auth'
import { formatCents, roleTagType, roleText } from '@/utils/format'

const auth = useAuthStore()

const keyword = ref('')
const users = ref<UserInfo[]>([])
const loading = ref(true)
const refreshing = ref(false)

const createOpen = ref(false)
const creating = ref(false)
const createForm = reactive({
  phone: '',
  email: '',
  password: '',
  name: '',
  role: 'user',
  tenantId: '',
})

const editOpen = ref(false)
const editing = ref(false)
const editForm = reactive({ id: '', status: '', email: '', name: '' })

const rechargeOpen = ref(false)
const recharging = ref(false)
const rechargingUser = ref<UserInfo | null>(null)
const rechargeForm = reactive({
  amountCents: '',
  method: 'cash',
  voucherNo: '',
  remark: '',
})

async function load(): Promise<void> {
  loading.value = true
  try {
    users.value = await adminApi.users(keyword.value.trim() || undefined)
  } finally {
    loading.value = false
  }
}

async function onRefresh(): Promise<void> {
  await load()
  refreshing.value = false
}

function resetCreateForm(): void {
  createForm.phone = ''
  createForm.email = ''
  createForm.password = ''
  createForm.name = ''
  createForm.role = 'user'
  createForm.tenantId = ''
}

async function createUser(): Promise<void> {
  if (!createForm.phone.trim() || !createForm.password || !createForm.name.trim()) {
    showToast('请填写手机号、密码和姓名')
    return
  }

  const payload: CreateUserPayload = {
    phone: createForm.phone.trim(),
    email: createForm.email.trim(),
    password: createForm.password,
    name: createForm.name.trim(),
  }
  if (auth.role === 'super') {
    payload.role = createForm.role as Role
    if (createForm.tenantId.trim()) payload.tenantId = createForm.tenantId.trim()
  }

  creating.value = true
  try {
    await adminApi.createUser(payload)
    showToast({ type: 'success', message: '用户创建成功' })
    createOpen.value = false
    resetCreateForm()
    await load()
  } catch {
    // 统一由 request.ts 提示
  } finally {
    creating.value = false
  }
}

function openEdit(user: UserInfo): void {
  editForm.id = user.id
  editForm.status = user.status || ''
  editForm.email = user.email || ''
  editForm.name = user.name || ''
  editOpen.value = true
}

async function saveUser(): Promise<void> {
  if (!editForm.status.trim()) {
    showToast('请填写用户状态')
    return
  }
  editing.value = true
  try {
    await adminApi.updateUser(editForm.id, {
      status: editForm.status.trim(),
      email: editForm.email.trim(),
      name: editForm.name.trim(),
    })
    showToast({ type: 'success', message: '保存成功' })
    editOpen.value = false
    await load()
  } catch {
    // 统一由 request.ts 提示
  } finally {
    editing.value = false
  }
}

function openRecharge(user: UserInfo): void {
  rechargingUser.value = user
  Object.assign(rechargeForm, { amountCents: '', method: 'cash', voucherNo: '', remark: '' })
  rechargeOpen.value = true
}

async function submitRecharge(): Promise<void> {
  const user = rechargingUser.value
  if (!user) return
  const amount = Number(rechargeForm.amountCents)
  if (!Number.isInteger(amount) || amount <= 0) {
    showToast('金额需为正整数（单位：分）')
    return
  }
  recharging.value = true
  try {
    const payload: WalletMutationPayload = {
      userId: user.id,
      amountCents: amount,
      method: rechargeForm.method.trim() || 'cash',
      voucherNo: rechargeForm.voucherNo.trim(),
      remark: rechargeForm.remark.trim(),
    }
    if (auth.role === 'super' && user.tenantId) {
      payload.tenantId = Number(user.tenantId)
    }
    await walletApi.recharge(payload)
    showToast({ type: 'success', message: '充值成功' })
    rechargeOpen.value = false
    await load()
  } catch {
    // 统一由 request.ts 提示
  } finally {
    recharging.value = false
  }
}

onMounted(load)
</script>

<template>
  <div class="page--plain">
    <NavBar title="用户管理">
      <template #right>
        <Icon name="plus" size="20" @click="createOpen = true" />
      </template>
    </NavBar>

    <div class="search-bar">
      <Field v-model="keyword" placeholder="搜索手机号 / 姓名" clearable />
      <Button size="small" type="primary" @click="load">查询</Button>
    </div>

    <PullRefresh v-model="refreshing" @refresh="onRefresh">
      <div v-if="loading" class="page-loading">
        <Loading type="spinner" />
      </div>
      <Empty v-else-if="users.length === 0" image="search" description="暂无用户" />
      <CellGroup v-else inset>
        <Cell
          v-for="user in users"
          :key="user.id"
          :title="`${user.name || '未命名'} · ${user.phone || '--'}`"
        >
          <template #value>
            <div class="col user-value">
              <Tag :type="roleTagType(user.role)" size="medium">
                {{ roleText(user.role) }}
              </Tag>
              <span class="money">¥{{ formatCents(user.balanceCents) }}</span>
            </div>
          </template>
          <template #label>
            <div class="col user-label">
              <span>{{ user.email || '未填写邮箱' }}</span>
              <span v-if="user.status" class="muted">状态：{{ user.status }}</span>
            </div>
          </template>
          <template #right-icon>
            <div class="row-actions">
              <Button
                v-if="auth.canRefund && user.role === 'user'"
                size="mini"
                plain
                type="success"
                @click="openRecharge(user)"
              >
                充值
              </Button>
              <Button size="mini" plain type="primary" @click="openEdit(user)">编辑</Button>
            </div>
          </template>
        </Cell>
      </CellGroup>
    </PullRefresh>

    <Popup v-model:show="createOpen" position="bottom" round>
      <div class="popup-title">新建用户</div>
      <div class="popup-form">
        <Field v-model="createForm.name" label="姓名" placeholder="请输入姓名" />
        <Field v-model="createForm.phone" label="手机号" placeholder="请输入手机号" />
        <Field v-model="createForm.email" label="邮箱" placeholder="请输入邮箱" />
        <Field v-model="createForm.password" label="密码" type="password" placeholder="请输入初始密码" />
        <template v-if="auth.role === 'super'">
          <Field v-model="createForm.role" label="角色" placeholder="user / tenant_admin / operator" />
          <Field v-model="createForm.tenantId" label="租户 ID" placeholder="创建租户管理员时填写" />
        </template>
      </div>
      <div class="form-actions">
        <Button block @click="createOpen = false">取消</Button>
        <Button block type="primary" :loading="creating" @click="createUser">创建</Button>
      </div>
    </Popup>

    <Popup v-model:show="editOpen" position="bottom" round>
      <div class="popup-title">编辑用户</div>
      <div class="popup-form">
        <Field v-model="editForm.name" label="姓名" placeholder="请输入姓名" />
        <Field v-model="editForm.email" label="邮箱" placeholder="请输入邮箱" />
        <Field v-model="editForm.status" label="状态" placeholder="如 active / disabled" />
      </div>
      <div class="form-actions">
        <Button block @click="editOpen = false">取消</Button>
        <Button block type="primary" :loading="editing" @click="saveUser">保存</Button>
      </div>
    </Popup>

    <Popup v-model:show="rechargeOpen" position="bottom" round>
      <div class="popup-title">线下充值</div>
      <div class="popup-form">
        <Field readonly label="用户" :model-value="rechargingUser ? `${rechargingUser.name} · ${rechargingUser.phone}` : ''" />
        <Field v-model="rechargeForm.amountCents" label="金额(分)" type="digit" placeholder="请输入整数分" />
        <Field v-model="rechargeForm.method" label="方式" placeholder="cash / wechat / alipay / bank" />
        <Field v-model="rechargeForm.voucherNo" label="凭证号" placeholder="选填" />
        <Field v-model="rechargeForm.remark" label="备注" placeholder="选填" />
      </div>
      <div class="form-actions">
        <Button block @click="rechargeOpen = false">取消</Button>
        <Button block type="primary" :loading="recharging" @click="submitRecharge">确认充值</Button>
      </div>
    </Popup>
  </div>
</template>

<style scoped>
.search-bar {
  display: flex;
  align-items: center;
  gap: 8px;
  margin: 12px;
}

.search-bar :deep(.van-cell) {
  padding: 8px 12px;
  border-radius: 8px;
}

.user-value {
  align-items: flex-end;
  gap: 4px;
}

.user-label {
  align-items: flex-start;
}

.row-actions {
  display: flex;
  gap: 4px;
  align-items: center;
}

.popup-title {
  padding: 16px;
  font-size: 15px;
  font-weight: 600;
  text-align: center;
}

.popup-form {
  max-height: 55vh;
  overflow-y: auto;
}

.form-actions {
  padding-bottom: calc(16px + env(safe-area-inset-bottom));
}
</style>
