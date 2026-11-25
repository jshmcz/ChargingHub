<template>
  <div class="page-container">
    <div class="page-header">
      <h2 class="page-title">{{ auth.isEndUser ? '我的通知记录' : '通知记录' }}</h2>
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
        <el-table-column prop="type" label="类型" width="140" />
        <el-table-column prop="title" label="标题" min-width="160" />
        <el-table-column prop="content" label="内容" min-width="240" show-overflow-tooltip />
        <el-table-column prop="status" label="状态" width="110" />
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
import { notificationsApi } from '@/api/notifications'
import type { Notification } from '@/api/types'
import { formatDate } from '@/utils/format'
import { useAuthStore } from '@/stores/auth'

const auth = useAuthStore()
const loading = ref(false)
const list = ref<Notification[]>([])
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
      ? await notificationsApi.list()
      : await notificationsApi.list(userIdFilter.value || undefined)
    page.value = 1
  } catch (error) {
    ElMessage.error(errorText(error))
  } finally {
    loading.value = false
  }
}

onMounted(load)
</script>
