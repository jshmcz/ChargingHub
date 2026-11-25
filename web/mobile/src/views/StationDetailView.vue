<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import {
  Button,
  Cell,
  CellGroup,
  Empty,
  Loading,
  NavBar,
  PullRefresh,
  Tag,
  showConfirmDialog,
  showToast,
} from 'vant'
import { chargeApi } from '@/api/charge'
import { reservationApi } from '@/api/reservations'
import { stationApi } from '@/api/stations'
import type { Gun, Station } from '@/api/types'
import {
  formatCents,
  formatTime,
  gunStatusText,
  gunStatusType,
} from '@/utils/format'

const route = useRoute()
const router = useRouter()
const stationId = computed(() => String(route.params.id ?? ''))

const station = ref<Station | null>(null)
const guns = ref<Gun[]>([])
const loading = ref(true)
const refreshing = ref(false)

function gunLabel(gun: Gun): string {
  return String(gun.gunNo ?? gun.no ?? gun.name ?? gun.id).slice(0, 12)
}

function isIdle(gun: Gun): boolean {
  return ['available', 'idle', 'free'].includes(gun.status)
}

async function load(): Promise<void> {
  loading.value = true
  try {
    const list = await stationApi.list()
    const found = list.find((item) => item.id === stationId.value) || null
    station.value = found
    const remoteGuns = await stationApi.guns(stationId.value)
    guns.value = remoteGuns.length
      ? remoteGuns
      : found?.guns || found?.piles?.flatMap((pile) => pile.guns || []) || []
  } finally {
    loading.value = false
  }
}

async function onRefresh(): Promise<void> {
  await load()
  refreshing.value = false
}

async function handleStart(gun: Gun): Promise<void> {
  try {
    await showConfirmDialog({
      title: '立即充电',
      message: `确认启动「${gunLabel(gun)}」？需余额不低于 ¥${formatCents(
        station.value?.startThresholdCents,
      )}。`,
      confirmButtonText: '启动',
    })
  } catch {
    return
  }

  try {
    const order = await chargeApi.start(gun.id)
    showToast({ type: 'success', message: '启动成功' })
    await router.push({ path: '/charge', query: { gunId: gun.id, orderId: order.id } })
  } catch {
    // 统一由 request.ts 提示
  }
}

async function handleReserve(gun: Gun): Promise<void> {
  try {
    await showConfirmDialog({
      title: '预约充电枪',
      message: `预约「${gunLabel(gun)}」？启动时间由场站规则自动计算，当天过点自动顺延次日。`,
      confirmButtonText: '预约',
    })
  } catch {
    return
  }

  try {
    const reservation = await reservationApi.create(stationId.value, gun.id)
    showToast({
      type: 'success',
      message: `预约成功，预计 ${formatTime(reservation.startAt)} 启动`,
      duration: 2500,
    })
    await router.push('/reservations')
  } catch {
    // 统一由 request.ts 提示
  }
}

onMounted(load)
</script>

<template>
  <div class="page--plain">
    <NavBar :title="station?.name || '场站详情'" />

    <div v-if="loading" class="page-loading">
      <Loading type="spinner" />
    </div>

    <template v-else>
      <CellGroup inset>
        <Cell title="场站地址" :value="station?.address || '-'" />
        <Cell title="默认启动时间" :value="station?.defaultStartTime || '-'" />
        <Cell title="启动门槛" :value="`¥${formatCents(station?.startThresholdCents)}`" />
        <Cell title="断电 / 预警门槛" :value="`¥${formatCents(station?.stopThresholdCents)} / ¥${formatCents(station?.warnThresholdCents)}`" />
      </CellGroup>

      <div class="section-title">充电枪状态</div>

      <PullRefresh v-model="refreshing" @refresh="onRefresh">
        <Empty v-if="guns.length === 0" image="search" description="该场站暂无充电枪" />
        <CellGroup v-else inset>
          <Cell v-for="gun in guns" :key="gun.id">
            <template #title>
              <div class="row-between gun-title">
                <span class="gun-name">枪 {{ gunLabel(gun) }}</span>
                <Tag :type="gunStatusType(gun.status)" size="medium">
                  {{ gunStatusText(gun.status) }}
                </Tag>
              </div>
            </template>
            <template #value>
              <div class="gun-actions">
                <Button
                  size="small"
                  type="primary"
                  :disabled="!isIdle(gun)"
                  @click="handleStart(gun)"
                >
                  立即充电
                </Button>
                <Button size="small" plain type="primary" @click="handleReserve(gun)">
                  预约
                </Button>
              </div>
            </template>
          </Cell>
        </CellGroup>
      </PullRefresh>

      <div class="card card--pad rule-tip">
        <div class="section-title rule-title">预约规则</div>
        <p class="muted">启动时间取场站固定时刻，用户不可修改；当天过点自动顺延次日。</p>
        <p class="muted">启动需余额 ≥ 启动门槛；充电中余额低于断电门槛将自动断电。</p>
      </div>
    </template>
  </div>
</template>

<style scoped>
.gun-title {
  width: 100%;
}

.gun-name {
  font-weight: 500;
}

.gun-actions {
  display: flex;
  gap: 8px;
}

.rule-tip {
  margin-top: 16px;
}

.rule-tip p {
  margin: 6px 0;
  line-height: 1.5;
}

.rule-title {
  margin: 0 0 8px;
}
</style>
