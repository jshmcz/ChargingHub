<template>
  <div class="page-container">
    <div class="page-header">
      <h2 class="page-title">{{ auth.isEndUser ? '我的钱包' : '钱包管理' }}</h2>
    </div>

    <div v-if="!auth.isEndUser" class="table-toolbar">
      <span class="text-muted">按用户 ID 筛选</span>
      <el-input-number v-model="userIdFilter" :min="0" :controls="false" placeholder="用户ID" />
      <el-button @click="load">查询</el-button>
      <el-button @click="userIdFilter = 0; load()">全部</el-button>
    </div>

    <div class="table-wrap">
      <el-table :data="pagedList" border v-loading="loading" min-width="760">
        <el-table-column prop="id" label="ID" width="90" />
        <el-table-column prop="userId" label="用户ID" width="110" />
        <el-table-column label="余额" min-width="140">
          <template #default="{ row }">{{ formatCents(row.balanceCents) }}</template>
        </el-table-column>
        <el-table-column label="冻结" width="100">
          <template #default="{ row }">
            <el-tag :type="row.frozen ? 'danger' : 'success'">{{ row.frozen ? '是' : '否' }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column v-if="!auth.isEndUser" label="操作" width="300" fixed="right">
          <template #default="{ row }">
            <el-button link type="primary" @click="openOp('recharge', row.userId)">线下充值</el-button>
            <el-button
              v-if="auth.canRefund"
              link
              type="warning"
              @click="openOp('refund', row.userId)"
            >
              线下退款
            </el-button>
            <el-button link @click="toggleFreeze(row)">
              {{ row.frozen ? '解冻' : '冻结' }}
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

    <el-dialog v-model="dialogVisible" :title="opType === 'recharge' ? '线下充值' : '线下退款'" width="480px">
      <el-alert
        v-if="opType === 'refund'"
        type="warning"
        :closable="false"
        show-icon
        title="退款前请确认：余额充足，且用户无充电中/待结算订单、无待处理/生效中预约。"
        style="margin-bottom: 16px"
      />
      <el-form ref="formRef" :model="form" :rules="rules" label-width="90px">
        <el-form-item label="用户ID" prop="userId">
          <el-input-number v-model="form.userId" :min="1" :controls="false" style="width: 100%" />
        </el-form-item>
        <el-form-item label="金额(元)" prop="amountYuan">
          <el-input-number v-model="form.amountYuan" :min="0.01" :precision="2" style="width: 100%" />
        </el-form-item>
        <el-form-item label="方式" prop="method">
          <el-select v-model="form.method" style="width: 100%">
            <el-option label="现金" value="cash" />
            <el-option label="银行转账" value="bank" />
            <el-option label="微信" value="wechat" />
            <el-option label="支付宝" value="alipay" />
            <el-option label="其他" value="other" />
          </el-select>
        </el-form-item>
        <el-form-item label="凭证号" prop="voucherNo">
          <el-input v-model="form.voucherNo" placeholder="选填" />
        </el-form-item>
        <el-form-item label="备注" prop="remark">
          <el-input v-model="form.remark" type="textarea" :rows="2" placeholder="选填" />
        </el-form-item>
      </el-form>
      <template #footer>
        <div class="dialog-footer">
          <el-button @click="dialogVisible = false">取消</el-button>
          <el-button type="primary" :loading="saving" @click="submit">确定</el-button>
        </div>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { ElMessage, type FormInstance, type FormRules } from 'element-plus'
import { walletsApi } from '@/api/wallets'
import type { WalletOpPayload } from '@/api/wallets'
import type { Wallet } from '@/api/types'
import { formatCents, yuanToCents } from '@/utils/format'
import { useAuthStore } from '@/stores/auth'

const auth = useAuthStore()
const loading = ref(false)
const list = ref<Wallet[]>([])
const userIdFilter = ref(0)
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
    if (auth.isEndUser) {
      const wallet = await walletsApi.my()
      list.value = wallet ? [wallet] : []
    } else {
      list.value = await walletsApi.list(userIdFilter.value || undefined)
    }
    page.value = 1
  } catch (error) {
    ElMessage.error(errorText(error))
  } finally {
    loading.value = false
  }
}

onMounted(load)

const dialogVisible = ref(false)
const opType = ref<'recharge' | 'refund'>('recharge')
const saving = ref(false)
const formRef = ref<FormInstance>()
const form = reactive({ userId: 0, amountYuan: 0, method: 'cash', voucherNo: '', remark: '' })

const rules: FormRules = {
  userId: [{ required: true, message: '请输入用户ID', trigger: 'blur' }],
  amountYuan: [{ required: true, message: '请输入金额', trigger: 'blur' }],
}

function openOp(type: 'recharge' | 'refund', userId: number) {
  opType.value = type
  Object.assign(form, {
    userId: userId || userIdFilter.value || 0,
    amountYuan: 0,
    method: 'cash',
    voucherNo: '',
    remark: '',
  })
  dialogVisible.value = true
}

async function submit() {
  if (!formRef.value) return
  try {
    await formRef.value.validate()
  } catch {
    return
  }
  const payload: WalletOpPayload = {
    userId: form.userId,
    amountCents: yuanToCents(form.amountYuan),
    method: form.method,
    voucherNo: form.voucherNo,
    remark: form.remark,
  }
  saving.value = true
  try {
    if (opType.value === 'recharge') await walletsApi.recharge(payload)
    else await walletsApi.refund(payload)
    ElMessage.success(opType.value === 'recharge' ? '充值成功' : '退款成功')
    dialogVisible.value = false
    await load()
  } catch (error) {
    ElMessage.error(errorText(error))
  } finally {
    saving.value = false
  }
}

async function toggleFreeze(row: Wallet) {
  try {
    await walletsApi.freeze(row.userId, !row.frozen)
    row.frozen = !row.frozen
    ElMessage.success(row.frozen ? '已冻结' : '已解冻')
  } catch (error) {
    ElMessage.error(errorText(error))
  }
}
</script>
