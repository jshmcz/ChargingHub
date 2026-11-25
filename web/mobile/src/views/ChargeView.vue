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
import { chargeApi } from '@/api/charge'
import { orderApi } from '@/api/orders'
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
  try {
    const [stationData, orderData] = await Promise.all([stationApi.list(), orderApi.mine()])
    stations.value = stationData || []
    orders.value = orderData || []
  } finally {
    loading.value = false
  }
}

async function startCharge(): Promise<void> {
  if (!selectedGunId.value) {
    showToast('请先选择充电枪')
    return
  }
  try {
    await showConfirmDialog({
      title: '启动充电',
      message: `确认启动 ${selectedGun.value?.label || '所选充电枪'}？余额需达到场站启动门槛。`,
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
      <CellGroup inset>
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
        <Button type="primary" block round :loading="acting" @click="startCharge">立即启动</Button>
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
