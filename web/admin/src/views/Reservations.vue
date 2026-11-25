<template>
  <div class="page-container">
    <div class="page-header">
      <h2 class="page-title">{{ auth.isEndUser ? '我的预约' : '预约管理' }}</h2>
    </div>

    <div v-if="!auth.isEndUser" class="table-toolbar">
      <span class="text-muted">按用户 ID 筛选</span>
      <el-input-number v-model="userIdFilter" :min="0" :controls="false" placeholder="用户ID" />
      <el-button @click="load">查询</el-button>
      <el-button @click="userIdFilter = 0; load()">全部</el-button>
    </div>

    <div class="table-wrap">
      <el-table :data="pagedList" border v-loading="loading" min-width="900">
        <el-table-column prop="id" label="ID" width="90" />
        <el-table-column prop="userId" label="用户ID" width="100" />
        <el-table-column label="场站" min-width="140">
          <template #default="{ row }">{{ row.stationName || row.stationId }}</template>
        </el-table-column>
        <el-table-column prop="gunId" label="枪ID" width="100" />
        <el-table-column label="预约启动时间" min-width="180">
          <template #default="{ row }">{{ formatDate(row.startTime) }}</template>
        </el-table-column>
        <el-table-column label="状态" width="110">
          <template #default="{ row }">
            <el-tag :type="statusTagType(row.status)">{{ statusText(row.status) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="120" fixed="right">
          <template #default="{ row }">
            <el-button v-if="canCancel(row.status)" link type="danger" @click="cancel(row)">
              取消
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
import { reservationsApi } from '@/api/reservations'
import type { Reservation } from '@/api/types'
import { formatDate } from '@/utils/format'
import { useAuthStore } from '@/stores/auth'

const auth = useAuthStore()
const loading = ref(false)
const list = ref<Reservation[]>([])
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
      ? await reservationsApi.my()
      : await reservationsApi.list(userIdFilter.value || undefined)
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
    pending: '待生效',
    active: '生效中',
    cancelled: '已取消',
    completed: '已完成',
  }
  return map[status] || status
}

function statusTagType(status: string): string {
  if (status === 'active') return 'primary'
  if (status === 'cancelled') return 'info'
  if (status === 'completed') return 'success'
  return 'warning'
}

function canCancel(status: string): boolean {
  return status === 'pending' || status === 'active'
}

async function cancel(row: Reservation) {
  try {
    await ElMessageBox.confirm('确认取消该预约？', '提示', { type: 'warning' })
  } catch {
    return
  }
  try {
    await reservationsApi.cancel(row.id)
    row.status = 'cancelled'
    ElMessage.success('已取消')
  } catch (error) {
    ElMessage.error(errorText(error))
  }
}
</script>
