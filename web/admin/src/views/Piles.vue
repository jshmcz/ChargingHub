<template>
  <div class="page-container">
    <div class="page-header">
      <h2 class="page-title">桩与枪管理</h2>
      <el-button type="primary" @click="openCreate">新建桩</el-button>
    </div>

    <div class="table-toolbar">
      <span class="text-muted">按场站筛选</span>
      <el-select v-model="filterStationId" style="width: 220px">
        <el-option label="全部场站" :value="0" />
        <el-option v-for="station in stations" :key="station.id" :label="station.name" :value="station.id" />
      </el-select>
    </div>

    <div class="table-wrap">
      <el-table :data="pagedList" border v-loading="loading" min-width="1000">
        <el-table-column prop="id" label="ID" width="80" />
        <el-table-column prop="deviceNo" label="设备编号" min-width="160" />
        <el-table-column label="所属场站" min-width="140">
          <template #default="{ row }">{{ stationName(row.stationId) }}</template>
        </el-table-column>
        <el-table-column label="枪数" width="90">
          <template #default="{ row }">{{ row.guns?.length ?? '-' }}</template>
        </el-table-column>
        <el-table-column label="设备密钥" min-width="200">
          <template #default="{ row }">
            <span class="secret-text">{{ secretVisible[row.id] ? row.secret : '••••••••' }}</span>
            <el-button link type="primary" size="small" @click="toggleSecret(row)">
              {{ secretVisible[row.id] ? '隐藏' : '查看' }}
            </el-button>
            <el-button link type="primary" size="small" @click="copySecret(row.secret)">复制</el-button>
          </template>
        </el-table-column>
        <el-table-column label="状态" width="110">
          <template #default="{ row }">
            <el-tag :type="row.status === 'online' ? 'success' : 'info'">
              {{ statusText(row.status) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="300" fixed="right">
          <template #default="{ row }">
            <el-button link type="primary" @click="openGuns(row)">查看枪</el-button>
            <el-button link type="primary" @click="toggleStatus(row)">
              {{ row.status === 'online' ? '停用' : '启用' }}
            </el-button>
            <el-button link type="warning" @click="regenerate(row)">重置密钥</el-button>
          </template>
        </el-table-column>
      </el-table>
    </div>

    <div class="pagination-wrapper">
      <el-pagination
        background
        layout="total, prev, pager, next"
        :total="filteredList.length"
        v-model:current-page="page"
        :page-size="pageSize"
      />
    </div>

    <el-dialog v-model="createVisible" title="新建桩" width="460px">
      <el-form ref="formRef" :model="form" :rules="rules" label-width="100px">
        <el-form-item label="所属场站" prop="stationId">
          <el-select v-model="form.stationId" style="width: 100%">
            <el-option v-for="station in stations" :key="station.id" :label="station.name" :value="station.id" />
          </el-select>
        </el-form-item>
        <el-form-item label="设备编号" prop="deviceNo">
          <el-input v-model="form.deviceNo" placeholder="设备唯一编号" />
        </el-form-item>
        <el-form-item label="枪数量" prop="gunCount">
          <el-input-number v-model="form.gunCount" :min="1" :max="32" />
        </el-form-item>
      </el-form>
      <template #footer>
        <div class="dialog-footer">
          <el-button @click="createVisible = false">取消</el-button>
          <el-button type="primary" :loading="saving" @click="create">确定</el-button>
        </div>
      </template>
    </el-dialog>

    <el-dialog v-model="gunVisible" title="设备枪列表" width="640px">
      <div class="text-muted" style="margin-bottom: 12px">
        场站：{{ gunStationName }} · 设备：{{ gunDeviceNo }}
      </div>
      <div class="table-wrap">
        <el-table :data="guns" border size="small" min-width="560">
          <el-table-column prop="id" label="ID" width="80" />
          <el-table-column prop="gunNo" label="枪编号" min-width="140" />
          <el-table-column label="状态" width="120">
            <template #default="{ row }">
              <el-tag>{{ gunStatusText(row.status) }}</el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="pileId" label="桩ID" width="100" />
        </el-table>
      </div>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { ElMessage, ElMessageBox, type FormInstance, type FormRules } from 'element-plus'
import { gunsApi, pilesApi } from '@/api/piles'
import type { PileInput } from '@/api/piles'
import type { Gun, Pile, Station } from '@/api/types'
import { stationsApi } from '@/api/stations'

const loading = ref(false)
const stations = ref<Station[]>([])
const piles = ref<Pile[]>([])
const filterStationId = ref(0)
const page = ref(1)
const pageSize = 10
const secretVisible = ref<Record<number, boolean>>({})

function errorText(error: unknown): string {
  return error instanceof Error ? error.message : '操作失败'
}

async function load() {
  loading.value = true
  try {
    const [stationList, pileList] = await Promise.all([stationsApi.list(), pilesApi.list()])
    stations.value = stationList
    piles.value = pileList
  } catch (error) {
    ElMessage.error(errorText(error))
  } finally {
    loading.value = false
  }
}

onMounted(load)

const filteredList = computed(() =>
  filterStationId.value
    ? piles.value.filter((pile) => pile.stationId === filterStationId.value)
    : piles.value,
)
const pagedList = computed(() =>
  filteredList.value.slice((page.value - 1) * pageSize, page.value * pageSize),
)

function stationName(stationId: number): string {
  return stations.value.find((item) => item.id === stationId)?.name ?? String(stationId)
}

function statusText(status: string): string {
  if (status === 'online') return '在线'
  if (status === 'offline') return '离线'
  return status
}

function gunStatusText(status: string): string {
  const map: Record<string, string> = {
    idle: '空闲',
    reserved: '已预约',
    charging: '充电中',
    offline: '离线',
  }
  return map[status] || status
}

function toggleSecret(row: Pile) {
  secretVisible.value[row.id] = !secretVisible.value[row.id]
}

async function copySecret(secret: string) {
  try {
    await navigator.clipboard.writeText(secret)
    ElMessage.success('已复制')
  } catch {
    ElMessage.warning('复制失败，请手动选择密钥')
  }
}

async function toggleStatus(row: Pile) {
  const next = row.status === 'online' ? 'offline' : 'online'
  try {
    await pilesApi.updateStatus(row.id, next)
    row.status = next
    ElMessage.success('状态已更新')
  } catch (error) {
    ElMessage.error(errorText(error))
  }
}

async function regenerate(row: Pile) {
  try {
    await ElMessageBox.confirm('重置后旧密钥立即失效，确认继续？', '提示', { type: 'warning' })
  } catch {
    return
  }
  try {
    const result = await pilesApi.regenerateSecret(row.id)
    row.secret = result.secret
    ElMessage.success('密钥已重置')
  } catch (error) {
    ElMessage.error(errorText(error))
  }
}

const gunVisible = ref(false)
const guns = ref<Gun[]>([])
const gunStationName = ref('')
const gunDeviceNo = ref('')

async function openGuns(pile: Pile) {
  try {
    const all = await gunsApi.list(pile.stationId)
    const matched = all.filter((gun) => gun.pileId === pile.id)
    guns.value = matched.length > 0 ? matched : all
    gunStationName.value = stationName(pile.stationId)
    gunDeviceNo.value = pile.deviceNo
    gunVisible.value = true
  } catch (error) {
    ElMessage.error(errorText(error))
  }
}

const createVisible = ref(false)
const saving = ref(false)
const formRef = ref<FormInstance>()
const form = reactive({ stationId: 0, deviceNo: '', gunCount: 1 })

const rules: FormRules = {
  stationId: [{ required: true, message: '请选择场站', trigger: 'change' }],
  deviceNo: [{ required: true, message: '请输入设备编号', trigger: 'blur' }],
}

function openCreate() {
  Object.assign(form, { stationId: filterStationId.value || 0, deviceNo: '', gunCount: 1 })
  createVisible.value = true
}

async function create() {
  if (!formRef.value) return
  try {
    await formRef.value.validate()
  } catch {
    return
  }
  const payload: PileInput = {
    stationId: form.stationId,
    deviceNo: form.deviceNo,
    gunCount: form.gunCount,
  }
  saving.value = true
  try {
    await pilesApi.create(payload)
    ElMessage.success('新建成功，枪已自动生成')
    createVisible.value = false
    await load()
  } catch (error) {
    ElMessage.error(errorText(error))
  } finally {
    saving.value = false
  }
}
</script>
