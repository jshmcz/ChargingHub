<template>
  <div class="page-container">
    <div class="page-header">
      <h2 class="page-title">{{ auth.isEndUser ? '我的订单' : '订单管理' }}</h2>
    </div>

    <div v-if="!auth.isEndUser" class="table-toolbar">
      <span class="text-muted">按用户 ID 筛选</span>
      <el-input-number v-model="userIdFilter" :min="0" :controls="false" placeholder="用户ID" />
      <el-button @click="load">查询</el-button>
      <el-button @click="userIdFilter = 0; load()">全部</el-button>
    </div>

    <div class="table-wrap">
      <el-table :data="pagedList" border v-loading="loading" min-width="940">
        <el-table-column prop="id" label="ID" width="90" />
        <el-table-column prop="userId" label="用户ID" width="100" />
        <el-table-column prop="stationId" label="场站ID" width="100" />
        <el-table-column prop="gunId" label="枪ID" width="90" />
        <el-table-column label="金额" min-width="130">
          <template #default="{ row }">{{ formatCents(row.amountCents) }}</template>
        </el-table-column>
        <el-table-column label="状态" width="130">
          <template #default="{ row }">
            <el-tag :type="statusTagType(row.status)">{{ statusText(row.status) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="开始时间" min-width="170">
          <template #default="{ row }">{{ formatDate(row.startTime) }}</template>
        </el-table-column>
        <el-table-column label="结束时间" min-width="170">
          <template #default="{ row }">{{ formatDate(row.endTime) }}</template>
        </el-table-column>
        <el-table-column v-if="!auth.isEndUser" label="操作" width="130" fixed="right">
          <template #default="{ row }">
            <el-button
              v-if="row.status === 'pending_settle'"
              link
              type="primary"
              @click="settle(row)"
            >
              手动结算
            </el-button>
            <span v-else class="text-muted">-</span>
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
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { ordersApi } from '@/api/orders'
import type { Order } from '@/api/types'
import { formatCents, formatDate } from '@/utils/format'
import { useAuthStore } from '@/stores/auth'

const auth = useAuthStore()
const loading = ref(false)
const list = ref<Order[]>([])
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
    list.value = auth.isEndUser
      ? await ordersApi.my()
      : await ordersApi.list(userIdFilter.value || undefined)
    page.value = 1
  } catch (error) {
    ElMessage.error(errorText(error))
  } finally {
    loading.value = false
  }
}

onMounted(load)

function statusText(status: string): string {
  const map: Record<string, string> = {
    charging: '充电中',
    pending_settle: '待结算',
    settled: '已结算',
  }
  return map[status] || status
}

function statusTagType(status: string): string {
  if (status === 'settled') return 'success'
  if (status === 'pending_settle') return 'warning'
  return 'primary'
}

async function settle(row: Order) {
  try {
    await ElMessageBox.confirm('确认按最后上报数据手动结算该订单？', '提示', { type: 'warning' })
  } catch {
    return
  }
  try {
    await ordersApi.settle(row.id)
    row.status = 'settled'
    ElMessage.success('结算成功')
  } catch (error) {
    ElMessage.error(errorText(error))
  }
}
</script>
