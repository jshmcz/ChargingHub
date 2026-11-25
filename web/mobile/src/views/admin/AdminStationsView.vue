<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import {
  Button,
  Cell,
  CellGroup,
  Empty,
  Field,
  Icon,
  Loading,
  NavBar,
  Popup,
  PullRefresh,
  Tag,
  showConfirmDialog,
  showDialog,
  showToast,
} from 'vant'
import { stationApi } from '@/api/stations'
import type { Pile, Station, StationPayload } from '@/api/types'
import { formatCents } from '@/utils/format'

const stations = ref<Station[]>([])
const piles = ref<Pile[]>([])
const loading = ref(true)
const refreshing = ref(false)

const stationOpen = ref(false)
const savingStation = ref(false)
const stationForm = reactive({
  id: '',
  name: '',
  address: '',
  areaCode: '',
  areaName: '',
  defaultStartTime: '22:00',
  startThresholdCents: '5000',
  stopThresholdCents: '2000',
  warnThresholdCents: '3000',
  settlementTimeoutHours: '24',
})

const pileOpen = ref(false)
const pileStationId = ref('')
const savingPile = ref(false)
const pileForm = reactive({ gunCount: '1' })
const pileStatusMap = reactive<Record<string, string>>({})

const currentStation = computed(
  () => stations.value.find((item) => item.id === pileStationId.value) || null,
)
const stationPiles = computed(() =>
  piles.value.filter((pile) => pile.stationId === pileStationId.value),
)

function resetStationForm(): void {
  stationForm.id = ''
  stationForm.name = ''
  stationForm.address = ''
  stationForm.areaCode = ''
  stationForm.areaName = ''
  stationForm.defaultStartTime = '22:00'
  stationForm.startThresholdCents = '5000'
  stationForm.stopThresholdCents = '2000'
  stationForm.warnThresholdCents = '3000'
  stationForm.settlementTimeoutHours = '24'
}

function openCreate(): void {
  resetStationForm()
  stationOpen.value = true
}

function openEdit(station: Station): void {
  stationForm.id = station.id
  stationForm.name = station.name || ''
  stationForm.address = station.address || ''
  stationForm.areaCode = station.areaCode || ''
  stationForm.areaName = station.areaName || ''
  stationForm.defaultStartTime = station.defaultStartTime || '22:00'
  stationForm.startThresholdCents = String(station.startThresholdCents ?? 5000)
  stationForm.stopThresholdCents = String(station.stopThresholdCents ?? 2000)
  stationForm.warnThresholdCents = String(station.warnThresholdCents ?? 3000)
  stationForm.settlementTimeoutHours = String(station.settlementTimeoutHours ?? 24)
  stationOpen.value = true
}

function buildStationPayload(): StationPayload | null {
  const numbers = [
    Number(stationForm.startThresholdCents),
    Number(stationForm.stopThresholdCents),
    Number(stationForm.warnThresholdCents),
    Number(stationForm.settlementTimeoutHours),
  ]
  if (numbers.some((value) => !Number.isFinite(value) || value < 0)) {
    showToast('金额与超时时长需为有效数字')
    return null
  }
  if (!stationForm.name.trim()) {
    showToast('请填写场站名称')
    return null
  }
  const areaCode = stationForm.areaCode.trim()
  if (!/^\d{6}$/.test(areaCode)) {
    showToast('请填写 6 位行政区划码（如 320106，用于生成桩编号）')
    return null
  }
  return {
    name: stationForm.name.trim(),
    address: stationForm.address.trim(),
    areaCode,
    areaName: stationForm.areaName.trim() || areaCode,
    defaultStartTime: stationForm.defaultStartTime.trim() || '22:00',
    startThresholdCents: numbers[0],
    stopThresholdCents: numbers[1],
    warnThresholdCents: numbers[2],
    settlementTimeoutHours: numbers[3],
  }
}

async function saveStation(): Promise<void> {
  const payload = buildStationPayload()
  if (!payload) return
  savingStation.value = true
  try {
    if (stationForm.id) {
      await stationApi.update(stationForm.id, payload)
      showToast({ type: 'success', message: '场站已更新' })
    } else {
      await stationApi.create(payload)
      showToast({ type: 'success', message: '场站已创建' })
    }
    stationOpen.value = false
    await load()
  } catch {
    // 统一由 request.ts 提示
  } finally {
    savingStation.value = false
  }
}

async function load(): Promise<void> {
  loading.value = true
  try {
    const [stationData, pileData] = await Promise.all([stationApi.list(), stationApi.piles()])
    stations.value = stationData || []
    piles.value = pileData || []
    piles.value.forEach((pile) => {
      if (!(pile.id in pileStatusMap)) pileStatusMap[pile.id] = pile.status || ''
    })
  } finally {
    loading.value = false
  }
}

async function onRefresh(): Promise<void> {
  await load()
  refreshing.value = false
}

function openPiles(station: Station): void {
  pileStationId.value = station.id
  pileForm.gunCount = '1'
  pileOpen.value = true
}

async function createPile(): Promise<void> {
  const gunCount = Number(pileForm.gunCount)
  if (!Number.isInteger(gunCount) || gunCount <= 0 || gunCount > 32) {
    showToast('请填写有效枪数（1-32）')
    return
  }
  savingPile.value = true
  try {
    await stationApi.createPile({
      stationId: pileStationId.value,
      gunCount,
    })
    showToast({ type: 'success', message: '充电桩已创建，编号自动生成' })
    await load()
  } catch {
    // 统一由 request.ts 提示
  } finally {
    savingPile.value = false
  }
}

async function savePileStatus(pile: Pile): Promise<void> {
  const status = (pileStatusMap[pile.id] || '').trim()
  if (!status) {
    showToast('请填写桩状态')
    return
  }
  try {
    await stationApi.updatePile(pile.id, status)
    showToast({ type: 'success', message: '状态已更新' })
    await load()
  } catch {
    // 统一由 request.ts 提示
  }
}

async function regenerateSecret(pile: Pile): Promise<void> {
  try {
    await showConfirmDialog({
      title: '重置设备密钥',
      message: '重置后旧密钥将立即失效，确认继续？',
      confirmButtonText: '重置',
    })
  } catch {
    return
  }
  try {
    const result = await stationApi.regenerateSecret(pile.id)
    await showDialog({
      title: '新的设备密钥',
      message: result.secret || '重置成功',
    })
  } catch {
    // 统一由 request.ts 提示
  }
}

onMounted(load)
</script>

<template>
  <div class="page--plain">
    <NavBar title="场站管理">
      <template #right>
        <Icon name="plus" size="20" @click="openCreate" />
      </template>
    </NavBar>

    <PullRefresh v-model="refreshing" @refresh="onRefresh">
      <div v-if="loading" class="page-loading">
        <Loading type="spinner" />
      </div>
      <Empty v-else-if="stations.length === 0" image="search" description="暂无场站" />
      <CellGroup v-else inset>
        <Cell
          v-for="station in stations"
          :key="station.id"
          :title="station.name"
          :label="station.address || '暂无地址'"
        >
          <template #value>
            <div class="col station-value">
              <span class="muted">默认 {{ station.defaultStartTime || '--' }} 启动</span>
              <span class="muted">门槛 ¥{{ formatCents(station.startThresholdCents) }}</span>
            </div>
          </template>
          <template #right-icon>
            <div class="station-actions">
              <Button size="mini" plain type="primary" @click="openEdit(station)">编辑</Button>
              <Button size="mini" plain type="primary" @click="openPiles(station)">桩管理</Button>
            </div>
          </template>
        </Cell>
      </CellGroup>
    </PullRefresh>

    <Popup v-model:show="stationOpen" position="bottom" round>
      <div class="popup-title">{{ stationForm.id ? '编辑场站' : '新建场站' }}</div>
      <div class="popup-form">
        <Field v-model="stationForm.name" label="名称" placeholder="请输入场站名称" />
        <Field v-model="stationForm.areaCode" label="区划码" type="digit" maxlength="6" placeholder="6位，如 320106" />
        <Field v-model="stationForm.areaName" label="所在地区" placeholder="如 安徽省 / 合肥市 / 蜀山区" />
        <div class="muted" style="padding: 0 16px 8px; font-size: 12px">区划码是桩编号前缀，六位数字，可在 PC 管理端按地址自动选择</div>
        <Field v-model="stationForm.address" label="详细地址" placeholder="街道、门牌号等" />
        <Field v-model="stationForm.defaultStartTime" label="启动时刻" placeholder="22:00" />
        <Field v-model="stationForm.startThresholdCents" label="启动门槛(分)" type="digit" />
        <Field v-model="stationForm.stopThresholdCents" label="断电门槛(分)" type="digit" />
        <Field v-model="stationForm.warnThresholdCents" label="预警门槛(分)" type="digit" />
        <Field v-model="stationForm.settlementTimeoutHours" label="结算超时(小时)" type="digit" />
      </div>
      <div class="form-actions">
        <Button block @click="stationOpen = false">取消</Button>
        <Button block type="primary" :loading="savingStation" @click="saveStation">保存</Button>
      </div>
    </Popup>

    <Popup v-model:show="pileOpen" position="bottom" round>
      <div class="popup-title">桩管理 · {{ currentStation?.name || '' }}</div>
      <div class="pile-area">
        <div class="section-title">新增充电桩</div>
        <Field v-model="pileForm.gunCount" label="枪数量" type="digit" placeholder="1" />
        <div class="muted" style="padding: 0 16px 8px; font-size: 12px">桩编号按场站区划码自动生成，每把枪自动分配 18 位编号</div>
        <div class="pile-actions">
          <Button block type="primary" :loading="savingPile" @click="createPile">添加桩</Button>
        </div>

        <div class="section-title">已有充电桩</div>
        <Empty
          v-if="stationPiles.length === 0"
          image="search"
          description="暂无充电桩"
        />
        <CellGroup v-else inset>
          <Cell
            v-for="pile in stationPiles"
            :key="pile.id"
            :title="`设备 ${pile.deviceNo || pile.id}`"
            :label="`枪数 ${pile.gunCount ?? '-'}`"
          >
            <template #value>
              <Tag size="medium" type="default">{{ pile.status || '未知状态' }}</Tag>
            </template>
          </Cell>
          <Cell v-for="pile in stationPiles" :key="`${pile.id}-op`" title="状态与密钥">
            <template #label>
              <div class="pile-op">
                <Field v-model="pileStatusMap[pile.id]" placeholder="输入新状态" />
                <Button size="mini" type="primary" @click="savePileStatus(pile)">保存</Button>
                <Button size="mini" plain type="danger" @click="regenerateSecret(pile)">重置密钥</Button>
              </div>
            </template>
          </Cell>
        </CellGroup>
      </div>
      <div class="popup-foot">
        <Button block @click="pileOpen = false">关闭</Button>
      </div>
    </Popup>
  </div>
</template>

<style scoped>
.station-value {
  align-items: flex-end;
  gap: 4px;
}

.station-actions {
  display: flex;
  gap: 6px;
}

.popup-title {
  padding: 16px;
  font-size: 15px;
  font-weight: 600;
  text-align: center;
}

.popup-form,
.pile-area {
  max-height: 58vh;
  overflow-y: auto;
}

.pile-actions {
  padding: 12px 16px;
}

.pile-op {
  display: flex;
  align-items: center;
  gap: 8px;
}

.pile-op :deep(.van-cell) {
  padding: 6px 8px;
  border: 1px solid #ebedf0;
  border-radius: 6px;
}

.popup-foot,
.form-actions {
  padding-bottom: calc(16px + env(safe-area-inset-bottom));
}
</style>
