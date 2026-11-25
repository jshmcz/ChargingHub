<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { Cell, CellGroup, Empty, Icon, Loading, NavBar, PullRefresh, Tag } from 'vant'
import { stationApi } from '@/api/stations'
import type { Gun, Station } from '@/api/types'
import { useAuthStore } from '@/stores/auth'
import { formatCents } from '@/utils/format'

const router = useRouter()
const auth = useAuthStore()

const stations = ref<Station[]>([])
const loading = ref(true)
const refreshing = ref(false)

function stationGuns(station: Station): Gun[] {
  return station.guns || station.piles?.flatMap((pile) => pile.guns || []) || []
}

function availableCount(station: Station): number {
  if (typeof station.availableGuns === 'number') return station.availableGuns
  return stationGuns(station).filter((gun) =>
    ['available', 'idle', 'free'].includes(gun.status),
  ).length
}

function totalCount(station: Station): number {
  if (typeof station.totalGuns === 'number') return station.totalGuns
  return stationGuns(station).length
}

async function load(): Promise<void> {
  loading.value = true
  try {
    stations.value = await stationApi.list()
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
    <NavBar :title="`你好，${auth.user?.name || '用户'}`">
      <template #right>
        <Icon name="bell" size="20" @click="router.push('/notifications')" />
      </template>
    </NavBar>

    <CellGroup inset class="quick-group">
      <Cell title="立即充电" label="选择空闲枪启动" is-link to="/charge" />
      <Cell title="我的预约" label="查看预约与启动时间" is-link to="/reservations" />
      <Cell title="通知记录" label="查看发送记录" is-link to="/notifications" />
    </CellGroup>

    <div class="section-title">附近场站</div>

    <PullRefresh v-model="refreshing" @refresh="onRefresh">
      <div v-if="loading" class="page-loading">
        <Loading type="spinner" />
      </div>
      <Empty v-else-if="stations.length === 0" image="search" description="暂无可用场站" />
      <CellGroup v-else inset>
        <Cell
          v-for="station in stations"
          :key="station.id"
          :title="station.name"
          :label="station.address || '暂无地址信息'"
          is-link
          @click="router.push(`/station/${station.id}`)"
        >
          <template #value>
            <div class="station-meta">
              <Tag :type="availableCount(station) > 0 ? 'success' : 'danger'" size="medium">
                空闲 {{ availableCount(station) }}/{{ totalCount(station) }}
              </Tag>
              <span class="muted">启动门槛 ¥{{ formatCents(station.startThresholdCents) }}</span>
            </div>
          </template>
        </Cell>
      </CellGroup>
      <div v-if="!loading && stations.length > 0" class="muted list-foot">
        共 {{ stations.length }} 个场站 · 下拉可刷新枪状态
      </div>
    </PullRefresh>
  </div>
</template>

<style scoped>
.quick-group {
  margin-top: 12px;
}

.station-meta {
  display: flex;
  flex-direction: column;
  align-items: flex-end;
  gap: 4px;
}

.list-foot {
  text-align: center;
  padding: 16px 0;
}
</style>
