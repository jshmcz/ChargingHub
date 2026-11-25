<template>
  <div class="page-container">
    <div class="page-header">
      <h2 class="page-title">{{ auth.isEndUser ? '我的流水' : '资金流水' }}</h2>
    </div>

    <div v-if="!auth.isEndUser" class="table-toolbar">
      <span class="text-muted">按用户 ID 筛选</span>
      <el-input-number v-model="userIdFilter" :min="0" :controls="false" placeholder="用户ID" />
      <el-button @click="load">查询</el-button>
      <el-button @click="userIdFilter = 0; load()">全部</el-button>
    </div>

    <div class="table-wrap">
      <el-table :data="pagedList" border v-loading="loading" min-width="980">
        <el-table-column prop="id" label="ID" width="90" />
        <el-table-column prop="userId" label="用户ID" width="100" />
        <el-table-column label="类型" width="110">
          <template #default="{ row }">
            <el-tag :type="row.type === 'refund' ? 'danger' : 'success'">
              {{ typeText(row.type) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="金额" min-width="130">
          <template #default="{ row }">{{ formatCents(row.amountCents) }}</template>
        </el-table-column>
        <el-table-column label="方式" width="110">
          <template #default="{ row }">{{ methodText(row.method) }}</template>
        </el-table-column>
        <el-table-column prop="voucherNo" label="凭证号" min-width="140" />
        <el-table-column prop="remark" label="备注" min-width="160" />
        <el-table-column label="时间" min-width="170">
          <template #default="{ row }">{{ formatDate(row.createdAt) }}</template>
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
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { ElMessage } from 'element-plus'
import { ledgersApi } from '@/api/wallets'
import type { Ledger } from '@/api/types'
import { formatCents, formatDate } from '@/utils/format'
import { useAuthStore } from '@/stores/auth'

const auth = useAuthStore()
const loading = ref(false)
const list = ref<Ledger[]>([])
const userIdFilter = ref(0)
const page = ref(1)
const pageSize = 10
const pagedList = computed(() =>
  list.value.slice((page.value - 1) * pageSize, page.value * pageSize),
)

function errorText(error: unknown): string {
  return error instanceof Error ? error.message : '加载失败'
}

async function load() {
  loading.value = true
  try {
    list.value = auth.isEndUser
      ? await ledgersApi.my()
      : await ledgersApi.list(userIdFilter.value || undefined)
    page.value = 1
  } catch (error) {
    ElMessage.error(errorText(error))
  } finally {
    loading.value = false
  }
}

onMounted(load)

function typeText(type: string): string {
  const map: Record<string, string> = {
    recharge: '充值',
    refund: '退款',
    charge: '充电扣费',
    settle: '结算',
  }
  return map[type] || type
}

function methodText(method: string): string {
  const map: Record<string, string> = {
    cash: '现金',
    bank: '银行转账',
    wechat: '微信',
    alipay: '支付宝',
    other: '其他',
  }
  return map[method] || method
}
</script>
