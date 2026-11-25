<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { Cell, CellGroup, Empty, Loading, NavBar, PullRefresh, Tag } from 'vant'
import { notificationApi } from '@/api/notifications'
import type { Notification } from '@/api/types'
import { useAuthStore } from '@/stores/auth'
import { formatTime } from '@/utils/format'

const auth = useAuthStore()
const notifications = ref<Notification[]>([])
const loading = ref(true)
const refreshing = ref(false)

async function load(): Promise<void> {
  loading.value = true
  try {
    notifications.value = await notificationApi.list(auth.user?.id)
  } finally {
    loading.value = false
  }
}

async function onRefresh(): Promise<void> {
  await load()
  refreshing.value = false
}

onMounted(load)
</script>

<template>
  <div class="page--plain">
    <NavBar title="通知记录" />

    <PullRefresh v-model="refreshing" @refresh="onRefresh">
      <div v-if="loading" class="page-loading">
        <Loading type="spinner" />
      </div>
      <Empty v-else-if="notifications.length === 0" image="search" description="暂无通知记录" />
      <CellGroup v-else inset>
        <Cell
          v-for="item in notifications"
          :key="item.id"
          :title="item.title || item.type || '系统通知'"
        >
          <template #value>
            <Tag v-if="item.status" size="medium" type="default">{{ item.status }}</Tag>
          </template>
          <template #label>
            <div class="col notification-label">
              <span v-if="item.content" class="content">{{ item.content }}</span>
              <span class="muted">{{ formatTime(item.sentAt || item.createdAt) }}</span>
            </div>
          </template>
        </Cell>
      </CellGroup>
    </PullRefresh>
  </div>
</template>

<style scoped>
.notification-label {
  align-items: flex-start;
  gap: 4px;
}

.content {
  color: var(--ch-text);
  font-size: 13px;
  white-space: normal;
  line-height: 1.5;
}
</style>
