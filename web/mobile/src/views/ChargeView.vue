<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { useRoute } from 'vue-router'
import {
  Button,
  Cell,
  CellGroup,
  Loading,
  NavBar,
  Popup,
  Radio,
  RadioGroup,
  Tag,
  showConfirmDialog,
  showToast,
} from 'vant'
import { chargeApi, type GunByToken } from '@/api/charge'
import { orderApi } from '@/api/orders'
import { reservationApi } from '@/api/reservations'
import { stationApi } from '@/api/stations'
import type { Order, Station } from '@/api/types'
import {
  formatCents,
  formatTime,
  gunStatusText,
  gunStatusType,
  orderStatusText,
  orderStatusType,
} from '@/utils/format'

interface GunOption {
  id: string
  label: string
  stationName: string
  status: string
}

const route = useRoute()
const stations = ref<Station[]>([])
const orders = ref<Order[]>([])
const selectedGunId = ref(String(route.query.gunId ?? ''))
const gunPopup = ref(false)
const loading = ref(false)
const acting = ref(false)

// 扫码模式：URL 带 token 时按令牌定位枪并锁定展示
const scanToken = computed(() => (typeof route.query.token === 'string' ? route.query.token : ''))
const scannedGun = ref<GunByToken | null>(null)
const scanError = ref('')
const isScanMode = computed(() => scanToken.value.length > 0)

const gunOptions = computed<GunOption[]>(() =>
  stations.value.flatMap((station) => {
    const guns = station.guns || station.piles?.flatMap((pile) => pile.guns || []) || []
    return guns.map((gun) => ({
      id: gun.id,
      label: `枪 ${String(gun.gunNo ?? gun.no ?? gun.name ?? gun.id).slice(0, 12)}`,
      stationName: station.name,
      status: gun.status,
    }))
  }),
)

const selectedGun = computed(
  () => gunOptions.value.find((item) => item.id === selectedGunId.value) || null,
)

const chargingOrder = computed<Order | null>(() => {
  const charging = orders.value.filter((item) => item.status === 'charging')
  return selectedGunId.value
    ? charging.find((item) => item.gunId === selectedGunId.value) || null
    : charging[0] || null
})

watch(
  () => route.query.gunId,
  (value) => {
    if (typeof value === 'string') selectedGunId.value = value
  },
)

function onGunChange(value: unknown): void {
  selectedGunId.value = String(value ?? '')
  gunPopup.value = false
}

async function load(): Promise<void> {
  loading.value = true
  scanError.value = ''
  try {
    const tasks: Promise<unknown>[] = [orderApi.mine()]
    if (isScanMode.value) {
      tasks.push(
        chargeApi.gunByToken(scanToken.value)
          .then((gun) => {
            scannedGun.value = gun
            selectedGunId.value = String(gun.gunId)
          })
          .catch((error: unknown) => {
            scannedGun.value = null
            scanError.value = error instanceof Error ? error.message : '二维码无效'
          }),
      )
    } else {
      tasks.push(stationApi.list().then((data) => { stations.value = data || [] }))
    }
    const [orderData] = await Promise.all(tasks)
    orders.value = (orderData as Order[]) || []
  } finally {
    loading.value = false
  }
}

async function reserveCharge(): Promise<void> {
  if (!scannedGun.value) return
  try {
    await showConfirmDialog({
      title: '预约充电',
      message: `确认预约这把枪？启动时间为场站设定的固定时刻（当天过点自动顺延次日），用户不可修改。`,
      confirmButtonText: '确认预约',
    })
  } catch {
    return
  }
  acting.value = true
  try {
    await reservationApi.create(String(scannedGun.value.stationId), String(scannedGun.value.gunId))
    showToast({ type: 'success', message: '预约成功' })
    await load()
  } catch {
    // 统一由 request.ts 提示
  } finally {
    acting.value = false
  }
}

async function startCharge(): Promise<void> {
  if (!selectedGunId.value) {
    showToast('请先选择充电枪')
    return
  }
  const gunLabel = isScanMode.value && scannedGun.value
    ? `枪 ${scannedGun.value.gunCode}`
    : (selectedGun.value?.label || '所选充电枪')
  try {
    await showConfirmDialog({
      title: '启动充电',
      message: `确认启动 ${gunLabel}？余额需达到场站启动门槛。`,
      confirmButtonText: '启动',
    })
  } catch {
    return
  }

  acting.value = true
  try {
    await chargeApi.start(selectedGunId.value)
    showToast({ type: 'success', message: '启动成功' })
    await load()
  } catch {
    // 统一由 request.ts 提示
  } finally {
    acting.value = false
  }
}

async function stopCharge(): Promise<void> {
  if (!selectedGunId.value) {
    showToast('请先选择充电枪')
    return
  }
  try {
    await showConfirmDialog({
      title: '停止充电',
      message: '确认停止当前充电中的订单？',
      confirmButtonText: '停止',
    })
  } catch {
    return
  }

  acting.value = true
  try {
    await chargeApi.stop(selectedGunId.value)
    showToast({ type: 'success', message: '已停止充电' })
    await load()
  } catch {
    // 统一由 request.ts 提示
  } finally {
    acting.value = false
  }
}

onMounted(load)
</script>

<template>
  <div class="page--plain">
    <NavBar title="充电控制" />

    <div v-if="loading" class="page-loading">
      <Loading type="spinner" />
    </div>

    <template v-else>
        <div v-if="isScanMode" class="scan-card card card--pad">
          <template v-if="scanError">
            <div class="scan-error">{{ scanError }}</div>
            <div class="muted">请确认二维码属于当前账号所在租户，或联系场站管理员。</div>
          </template>
          <template v-else-if="scannedGun">
            <div class="scan-station">{{ scannedGun.stationName }}</div>
            <div class="scan-code">{{ scannedGun.gunCode }}</div>
            <div class="scan-meta">
              <Tag :type="gunStatusType(scannedGun.status)" size="medium">{{ gunStatusText(scannedGun.status) }}</Tag>
              <span class="muted">余额 ¥{{ formatCents(scannedGun.balanceCents) }}</span>
            </div>
          </template>
        </div>
        <CellGroup v-if="!isScanMode" inset>
          <Cell title="当前充电枪" is-link @click="gunPopup = true">
          <template #label>
            <span class="muted">{{ selectedGun ? selectedGun.stationName : '请选择场站下的空闲枪' }}</span>
          </template>
          <template #value>
            <div class="row-start">
              <span>{{ selectedGun ? selectedGun.label : '未选择' }}</span>
              <Tag v-if="selectedGun" :type="gunStatusType(selectedGun.status)" size="medium">
                {{ gunStatusText(selectedGun.status) }}
              </Tag>
            </div>
          </template>
        </Cell>
      </CellGroup>

      <div class="charge-actions">
        <Button type="primary" block round :loading="acting" :disabled="!selectedGunId || !!scanError" @click="startCharge">立即启动</Button>
        <Button v-if="isScanMode" plain type="primary" block round :loading="acting" :disabled="!scannedGun || !!scanError" @click="reserveCharge">
          预约充电
        </Button>
        <Button plain type="danger" block round :disabled="!chargingOrder" @click="stopCharge">
          停止充电
        </Button>
      </div>

      <div class="section-title">充电中订单</div>
      <CellGroup v-if="chargingOrder" inset>
        <Cell title="订单状态">
          <template #value>
            <Tag :type="orderStatusType(chargingOrder.status)" size="medium">
              {{ orderStatusText(chargingOrder.status) }}
            </Tag>
          </template>
        </Cell>
        <Cell title="开始时间" :value="formatTime(chargingOrder.startTime || chargingOrder.createdAt)" />
        <Cell title="当前金额" :value="`¥${formatCents(chargingOrder.amountCents)}`" />
      </CellGroup>
      <div v-else class="empty-tip">当前没有充电中的订单</div>

      <div class="card card--pad charge-tip">
        <p class="muted">启动需余额 ≥ 场站启动门槛；余额低于断电门槛时自动断电并生成待结算订单。</p>
        <p class="muted">低于预警门槛会发送提醒，可在「通知记录」查看。</p>
      </div>
    </template>

    <Popup v-model:show="gunPopup" position="bottom" round>
      <div class="popup-title">选择充电枪</div>
      <RadioGroup v-model="selectedGunId" class="picker-list" @change="onGunChange">
        <Cell v-for="gun in gunOptions" :key="gun.id" clickable @click="onGunChange(gun.id)">
          <template #title>
            <Radio :name="gun.id">{{ gun.label }}</Radio>
          </template>
          <template #label>{{ gun.stationName }}</template>
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
.scan-card {
  margin: 12px;
  text-align: center;
}
.scan-station {
  font-size: 15px;
  font-weight: 500;
  margin-bottom: 6px;
}
.scan-code {
  font-size: 20px;
  font-weight: 600;
  letter-spacing: 1px;
  margin-bottom: 12px;
  word-break: break-all;
}
.scan-meta {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 12px;
}
.scan-error {
  color: var(--van-danger-color, #ee0a24);
  font-weight: 500;
  margin-bottom: 6px;
}
.charge-actions {
  display: flex;
  flex-direction: column;
  gap: 12px;
  padding: 16px 12px;
}

.charge-tip p {
  margin: 4px 0;
  line-height: 1.5;
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
