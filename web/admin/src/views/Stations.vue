<template>
  <div class="page-container">
    <div class="page-header">
      <h2 class="page-title">场站管理</h2>
      <el-button v-if="auth.isStaff" type="primary" @click="openCreate">新建场站</el-button>
    </div>

    <div class="table-wrap">
      <el-table :data="pagedList" border v-loading="loading" min-width="1100">
        <el-table-column prop="id" label="ID" width="80" />
        <el-table-column v-if="auth.isSuper" prop="tenantName" label="所属租户" min-width="140" />
        <el-table-column prop="name" label="名称" min-width="140" />
        <el-table-column label="所在地区" min-width="140">
          <template #default="{ row }">{{ row.areaName || '-' }}</template>
        </el-table-column>
        <el-table-column prop="address" label="地址" min-width="180" />
        <el-table-column prop="defaultStartTime" label="默认启动" width="110" />
        <el-table-column label="启动门槛" width="120">
          <template #default="{ row }">{{ formatCents(row.startThresholdCents) }}</template>
        </el-table-column>
        <el-table-column label="断电门槛" width="120">
          <template #default="{ row }">{{ formatCents(row.stopThresholdCents) }}</template>
        </el-table-column>
        <el-table-column label="预警门槛" width="120">
          <template #default="{ row }">{{ formatCents(row.warnThresholdCents) }}</template>
        </el-table-column>
        <el-table-column label="结算超时(h)" width="120">
          <template #default="{ row }">{{ row.settlementTimeoutHours }}</template>
        </el-table-column>
        <el-table-column v-if="auth.isStaff" label="操作" width="120" fixed="right">
          <template #default="{ row }">
            <el-button link type="primary" @click="openEdit(row)">编辑</el-button>
          </template>
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

    <el-dialog v-model="dialogVisible" :title="editing ? '编辑场站' : '新建场站'" width="620px">
      <el-form ref="formRef" :model="form" :rules="rules" label-width="120px">
        <el-row :gutter="16">
          <el-col v-if="auth.isSuper && !editing" :xs="24" :sm="12">
            <el-form-item label="所属租户" prop="tenantId">
              <el-select v-model="form.tenantId" style="width: 100%" filterable placeholder="选择租户">
                <el-option v-for="t in tenants" :key="t.id" :label="t.name" :value="t.id" />
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :xs="24" :sm="12">
            <el-form-item label="场站名称" prop="name">
              <el-input v-model="form.name" />
            </el-form-item>
          </el-col>
          <el-col :xs="24" :sm="12">
            <el-form-item label="所在地区" prop="areaCodes">
              <el-cascader
                v-model="form.areaCodes"
                :options="areaOptions"
                :props="{ expandTrigger: 'hover' }"
                placeholder="选择省 / 市 / 区县"
                style="width: 100%"
                clearable
                filterable
              />
            </el-form-item>
          </el-col>
          <el-col :xs="24" :sm="12">
            <el-form-item label="详细地址" prop="address">
              <el-input v-model="form.address" placeholder="街道、门牌号等" />
            </el-form-item>
          </el-col>
          <el-col :xs="24" :sm="12">
            <el-form-item label="默认启动时刻" prop="defaultStartTime">
              <el-input v-model="form.defaultStartTime" placeholder="如 22:00" />
            </el-form-item>
          </el-col>
          <el-col :xs="24" :sm="12">
            <el-form-item label="结算超时(小时)" prop="settlementTimeoutHours">
              <el-input-number v-model="form.settlementTimeoutHours" :min="1" :max="720" />
            </el-form-item>
          </el-col>
          <el-col :xs="24" :sm="12">
            <el-form-item label="启动门槛(元)">
              <el-input-number v-model="form.startYuan" :min="0" :precision="2" />
            </el-form-item>
          </el-col>
          <el-col :xs="24" :sm="12">
            <el-form-item label="断电门槛(元)">
              <el-input-number v-model="form.stopYuan" :min="0" :precision="2" />
            </el-form-item>
          </el-col>
          <el-col :xs="24" :sm="12">
            <el-form-item label="预警门槛(元)">
              <el-input-number v-model="form.warnYuan" :min="0" :precision="2" />
            </el-form-item>
          </el-col>
        </el-row>
      </el-form>
      <template #footer>
        <div class="dialog-footer">
          <el-button @click="dialogVisible = false">取消</el-button>
          <el-button type="primary" :loading="saving" @click="save">保存</el-button>
        </div>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { ElMessage, type FormInstance, type FormRules } from 'element-plus'
import { stationsApi } from '@/api/stations'
import { tenantsApi } from '@/api/tenants'
import type { StationInput } from '@/api/stations'
import type { Station } from '@/api/types'
import { centsToYuan, formatCents, yuanToCents } from '@/utils/format'
import { useAuthStore } from '@/stores/auth'

interface AreaNode { value: string; label: string; children?: AreaNode[] }
const auth = useAuthStore()
const loading = ref(false)
const list = ref<Station[]>([])
const tenants = ref<{ id: number; name: string }[]>([])
const areaOptions = ref<AreaNode[]>([])
const page = ref(1)
const pageSize = 10
const pagedList = computed(() =>
  list.value.slice((page.value - 1) * pageSize, page.value * pageSize),
)

function errorText(error: unknown): string {
  return error instanceof Error ? error.message : '操作失败'
}

async function load() {
  loading.value = true
  try {
    list.value = await stationsApi.list()
  } catch (error) {
    ElMessage.error(errorText(error))
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  load()
  fetch('/areas.json').then((r) => r.json()).then((d) => { areaOptions.value = d }).catch(() => {})
  if (auth.isSuper) {
    tenantsApi.list().then((rows) => { tenants.value = rows }).catch(() => {})
  }
})

function resolveArea(codes: string[]): { code: string; name: string } {
  if (!codes.length) return { code: '', name: '' }
  let nodes = areaOptions.value
  const labels: string[] = []
  for (const code of codes) {
    const hit = nodes.find((n) => n.value === code)
    if (!hit) break
    labels.push(hit.label)
    nodes = hit.children || []
  }
  return { code: codes[codes.length - 1], name: labels.join(' / ') }
}

// 根据末级区划码反查完整级联路径（省/市/区）
function areaCodesFromCode(code: string): string[] {
  if (!code) return []
  const path: string[] = []
  function walk(nodes: AreaNode[], trail: string[]): boolean {
    for (const n of nodes) {
      const next = [...trail, n.value]
      if (n.value === code) { path.push(...next); return true }
      if (n.children && walk(n.children, next)) return true
    }
    return false
  }
  walk(areaOptions.value, [])
  return path
}

const dialogVisible = ref(false)
const editing = ref<Station | null>(null)
const saving = ref(false)
const formRef = ref<FormInstance>()
const form = reactive({
  tenantId: 0,
  name: '',
  address: '',
  areaCodes: [] as string[],
  defaultStartTime: '22:00',
  settlementTimeoutHours: 24,
  startYuan: 50,
  stopYuan: 20,
  warnYuan: 30,
})

const rules: FormRules = {
  tenantId: [{ validator: (_r, v, cb) => (!auth.isSuper || editing.value || v > 0) ? cb() : cb(new Error('请选择所属租户')), trigger: 'change' }],
  name: [{ required: true, message: '请输入场站名称', trigger: 'blur' }],
  address: [{ required: true, message: '请输入详细地址', trigger: 'blur' }],
  areaCodes: [{ type: 'array', required: true, message: '请选择所在地区（用于生成桩编号）', trigger: 'change' }],
  defaultStartTime: [{ required: true, message: '请输入默认启动时刻', trigger: 'blur' }],
}

function openCreate() {
  editing.value = null
  Object.assign(form, {
    tenantId: tenants.value[0]?.id || 0,
    name: '',
    address: '',
    areaCodes: [],
    defaultStartTime: '22:00',
    settlementTimeoutHours: 24,
    startYuan: 50,
    stopYuan: 20,
    warnYuan: 30,
  })
  dialogVisible.value = true
}

function openEdit(row: Station) {
  editing.value = row
  Object.assign(form, {
    name: row.name,
    address: row.address,
    areaCodes: areaCodesFromCode(row.areaCode || ''),
    defaultStartTime: row.defaultStartTime,
    settlementTimeoutHours: row.settlementTimeoutHours,
    startYuan: centsToYuan(row.startThresholdCents),
    stopYuan: centsToYuan(row.stopThresholdCents),
    warnYuan: centsToYuan(row.warnThresholdCents),
  })
  dialogVisible.value = true
}

async function save() {
  if (!formRef.value) return
  try {
    await formRef.value.validate()
  } catch {
    return
  }
  const area = resolveArea(form.areaCodes)
  const payload: StationInput = {
    tenantId: form.tenantId,
    name: form.name,
    address: form.address,
    areaCode: area.code,
    areaName: area.name,
    defaultStartTime: form.defaultStartTime,
    settlementTimeoutHours: form.settlementTimeoutHours,
    startThresholdCents: yuanToCents(form.startYuan),
    stopThresholdCents: yuanToCents(form.stopYuan),
    warnThresholdCents: yuanToCents(form.warnYuan),
  }
  saving.value = true
  try {
    if (editing.value) {
      await stationsApi.update(editing.value.id, payload)
    } else {
      await stationsApi.create(payload)
    }
    ElMessage.success('保存成功')
    dialogVisible.value = false
    await load()
  } catch (error) {
    ElMessage.error(errorText(error))
  } finally {
    saving.value = false
  }
}
</script>
