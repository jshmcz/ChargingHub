<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import {
  Button,
  Cell,
  CellGroup,
  Empty,
  Loading,
  NavBar,
  Popup,
  PullRefresh,
  Radio,
  RadioGroup,
  Tag,
  showConfirmDialog,
  showToast,
} from 'vant'
import { reservationApi } from '@/api/reservations'
import { stationApi } from '@/api/stations'
import type { Gun, Reservation, Station } from '@/api/types'
import {
  formatTime,
  gunStatusText,
  gunStatusType,
  reservationStatusText,
  reservationStatusType,
} from '@/utils/format'

const stations = ref<Station[]>([])
const guns = ref<Gun[]>([])
const reservations = ref<Reservation[]>([])
const stationPick = ref('')
const gunPick = ref('')
const stationPopup = ref(false)
const gunPopup = ref(false)
const loading = ref(true)
const gunLoading = ref(false)
const submitting = ref(false)
const refreshing = ref(false)

const selectedStation = computed(
  () => stations.value.find((item) => item.id === stationPick.value) || null,
)
const selectedGun = computed(() => guns.value.find((item) => item.id === gunPick.value) || null)

function gunLabel(gun: Gun): string {
  return String(gun.gunNo ?? gun.no ?? gun.name ?? gun.id).slice(0, 12)
}

function canCancel(item: Reservation): boolean {
  return ['pending', 'active'].includes(item.status)
}

async function loadGuns(stationId: string): Promise<void> {
  gunLoading.value = true
  try {
    const remote = await stationApi.guns(stationId)
    if (remote.length > 0) guns.value = remote
  } finally {
    gunLoading.value = false
  }
}

async function onStationChange(value: unknown): Promise<void> {
  const id = String(value ?? '')
  const found = stations.value.find((item) => item.id === id) || null
  stationPick.value = id
  gunPick.value = ''
  guns.value = found?.guns || found?.piles?.flatMap((pile) => pile.guns || []) || []
  stationPopup.value = false
  if (found) await loadGuns(id)
}

function onGunChange(value: unknown): void {
  gunPick.value = String(value ?? '')
  gunPopup.value = false
}

async function loadAll(): Promise<void> {
  loading.value = true
  try {
    const [stationData, reservationData] = await Promise.all([
      stationApi.list(),
      reservationApi.mine(),
    ])
    stations.value = stationData || []
    reservations.value = reservationData || []
  } finally {
    loading.value = false
  }
}

async function onRefresh(): Promise<void> {
  await loadAll()
  refreshing.value = false
}

async function submit(): Promise<void> {
  if (!stationPick.value || !gunPick.value) {
    showToast('请先选择场站和充电枪')
    return
  }
  submitting.value = true
  try {
    const result = await reservationApi.create(stationPick.value, gunPick.value)
    showToast({
      type: 'success',
      message: `预约成功，预计启动：${formatTime(result.startAt)}`,
      duration: 2500,
    })
    gunPick.value = ''
    await loadAll()
  } catch {
    // 统一由 request.ts 提示
  } finally {
    submitting.value = false
  }
}

async function cancel(item: Reservation): Promise<void> {
  try {
    await showConfirmDialog({ title: '取消预约', message: '确认取消这条预约？' })
  } catch {
    return
  }
  try {
    await reservationApi.cancel(item.id)
    showToast({ type: 'success', message: '预约已取消' })
    await loadAll()
  } catch {
    // 统一由 request.ts 提示
  }
}

onMounted(loadAll)
</script>

<template>
  <div class="page--plain">
    <NavBar title="我的预约" />

    <div class="card card--pad">
      <p class="muted rule">预约启动时间取场站固定时刻，用户不可修改；当天过点自动顺延次日。</p>
      <p class="muted rule">创建时需保证余额 ≥ 场站启动门槛，且该枪未被预约或充电占用。</p>

      <CellGroup class="pick-group">
        <Cell
          title="选择场站"
          :value="selectedStation?.name || '请选择'"
          is-link
          @click="stationPopup = true"
        />
        <Cell
          title="选择充电枪"
          :value="selectedGun ? `枪 ${gunLabel(selectedGun)}` : '请选择'"
          is-link
          :disabled="!selectedStation"
          @click="gunPopup = true"
        />
        <Cell
          title="预计启动时间"
          :value="selectedStation ? `${selectedStation.defaultStartTime || '--'}（以服务端确认为准）` : '--'"
        />
      </CellGroup>

      <div class="submit-wrap">
        <Button
          type="primary"
          block
          round
          :loading="submitting"
          :disabled="!selectedStation || !selectedGun"
          @click="submit"
        >
          提交预约
        </Button>
      </div>
    </div>

    <div class="section-title">预约记录</div>

    <PullRefresh v-model="refreshing" @refresh="onRefresh">
      <div v-if="loading" class="page-loading">
        <Loading type="spinner" />
      </div>
      <Empty v-else-if="reservations.length === 0" image="search" description="暂无预约记录" />
      <CellGroup v-else inset>
        <Cell
          v-for="item in reservations"
          :key="item.id"
          :title="item.stationName || '充电场站'"
        >
          <template #value>
            <Tag :type="reservationStatusType(item.status)" size="medium">
              {{ reservationStatusText(item.status) }}
            </Tag>
          </template>
          <template #label>
            <div class="col reservation-label">
              <span>预计启动：{{ formatTime(item.startAt) }}</span>
              <span class="muted">下单时间：{{ formatTime(item.createdAt) }}</span>
            </div>
          </template>
          <template #right-icon>
            <Button
              v-if="canCancel(item)"
              size="mini"
              plain
              type="danger"
              @click="cancel(item)"
            >
              取消
            </Button>
          </template>
        </Cell>
      </CellGroup>
    </PullRefresh>

    <Popup v-model:show="stationPopup" position="bottom" round>
      <div class="popup-title">选择场站</div>
      <RadioGroup v-model="stationPick" class="picker-list" @change="onStationChange">
        <Cell v-for="item in stations" :key="item.id" clickable @click="onStationChange(item.id)">
          <template #title>
            <Radio :name="item.id">{{ item.name }}</Radio>
          </template>
          <template #label>{{ item.address }}</template>
        </Cell>
      </RadioGroup>
      <div class="popup-foot">
        <Button block type="primary" @click="stationPopup = false">关闭</Button>
      </div>
    </Popup>

    <Popup v-model:show="gunPopup" position="bottom" round>
      <div class="popup-title">选择充电枪</div>
      <div v-if="gunLoading" class="page-loading"><Loading type="spinner" /></div>
      <RadioGroup v-else v-model="gunPick" class="picker-list" @change="onGunChange">
        <Cell v-for="gun in guns" :key="gun.id" clickable @click="onGunChange(gun.id)">
          <template #title>
            <Radio :name="gun.id">枪 {{ gunLabel(gun) }}</Radio>
          </template>
          <template #value>
            <Tag :type="gunStatusType(gun.status)" size="medium">{{ gunStatusText(gun.status) }}</Tag>
          </template>
        </Cell>
      </RadioGroup>
      <div class="popup-foot">
        <Button block type="primary" @click="gunPopup = false">关闭</Button>
      </div>
    </Popup>
  </div>
</template>

<style scoped>
.rule {
  margin: 0 0 4px;
  line-height: 1.5;
}

.pick-group {
  margin-top: 12px;
}

.submit-wrap {
  margin-top: 16px;
}

.reservation-label {
  align-items: flex-start;
}

.popup-title {
  padding: 16px;
  font-size: 15px;
  font-weight: 600;
  text-align: center;
}

.popup-foot {
  padding: 12px 16px calc(12px + env(safe-area-inset-bottom));
}

.picker-list {
  max-height: 55vh;
  overflow-y: auto;
}
</style>
