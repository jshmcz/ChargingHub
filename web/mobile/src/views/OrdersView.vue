<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { Cell, CellGroup, Empty, Loading, NavBar, PullRefresh, Tag } from 'vant'
import { orderApi } from '@/api/orders'
import type { Order } from '@/api/types'
import { formatCents, formatTime, orderStatusText, orderStatusType } from '@/utils/format'

const orders = ref<Order[]>([])
const loading = ref(true)
const refreshing = ref(false)

async function load(): Promise<void> {
  loading.value = true
  try {
    orders.value = await orderApi.mine()
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
  <div class="page">
    <NavBar title="充电订单" />

    <PullRefresh v-model="refreshing" @refresh="onRefresh">
      <div v-if="loading" class="page-loading">
        <Loading type="spinner" />
      </div>
      <Empty v-else-if="orders.length === 0" image="search" description="暂无充电订单" />
      <CellGroup v-else inset>
        <Cell
          v-for="item in orders"
          :key="item.id"
          :title="`订单 ${String(item.orderNo ?? item.id).slice(0, 16)}`"
        >
          <template #value>
            <div class="col order-value">
              <Tag :type="orderStatusType(item.status)" size="medium">
                {{ orderStatusText(item.status) }}
              </Tag>
              <span :class="item.status === 'settled' ? 'muted' : 'money'">
                ¥{{ formatCents(item.amountCents) }}
              </span>
            </div>
          </template>
          <template #label>
            <div class="col order-label">
              <span>开始：{{ formatTime(item.startTime || item.createdAt) }}</span>
              <span v-if="item.endTime">结束：{{ formatTime(item.endTime) }}</span>
            </div>
          </template>
        </Cell>
      </CellGroup>
      <div v-if="!loading && orders.length > 0" class="muted list-foot">
        共 {{ orders.length }} 条订单
      </div>
    </PullRefresh>
  </div>
</template>

<style scoped>
.order-value {
  align-items: flex-end;
  gap: 6px;
}

.order-label {
  align-items: flex-start;
}

.list-foot {
  text-align: center;
  padding: 16px 0;
}
</style>
