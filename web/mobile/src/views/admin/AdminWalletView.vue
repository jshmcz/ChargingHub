<script setup lang="ts">
import { computed, onMounted, reactive, ref, watch } from 'vue'
import {
  Button,
  Cell,
  CellGroup,
  Empty,
  Field,
  Loading,
  NavBar,
  Popup,
  Radio,
  RadioGroup,
  Tag,
  showToast,
} from 'vant'
import { adminApi } from '@/api/admin'
import type { Ledger, UserInfo, Wallet } from '@/api/types'
import { walletApi, type WalletMutationPayload } from '@/api/wallets'
import { useAuthStore } from '@/stores/auth'
import { formatCents, formatTime, ledgerText, signedYuan } from '@/utils/format'

const auth = useAuthStore()

const users = ref<UserInfo[]>([])
const wallets = ref<Wallet[]>([])
const ledgers = ref<Ledger[]>([])
const selectedUserId = ref('')
const userPopup = ref(false)
const loading = ref(true)
const submitting = ref(false)

const form = reactive({
  amountCents: '',
  method: 'cash',
  voucherNo: '',
  remark: '',
})

const selectedUser = computed(
  () => users.value.find((item) => item.id === selectedUserId.value) || null,
)
const selectedWallet = computed<Wallet | null>(() => {
  return wallets.value.find((item) => item.userId === selectedUserId.value) || null
})
const amountYuan = computed(() => formatCents(Number(form.amountCents || 0)))

async function loadLedgers(): Promise<void> {
  ledgers.value = selectedUserId.value
    ? await walletApi.ledgers(selectedUserId.value)
    : []
}

async function load(): Promise<void> {
  loading.value = true
  try {
    const [userData, walletData] = await Promise.all([adminApi.users(), walletApi.list()])
    users.value = userData || []
    wallets.value = walletData || []
    if (!selectedUserId.value && users.value.length > 0) {
      selectedUserId.value = users.value[0].id
    }
    await loadLedgers()
  } finally {
    loading.value = false
  }
}

watch(selectedUserId, () => {
  void loadLedgers()
})

function onUserChange(value: unknown): void {
  selectedUserId.value = String(value ?? '')
  userPopup.value = false
}

async function submit(kind: 'recharge' | 'refund'): Promise<void> {
  if (!selectedUserId.value) {
    showToast('请先选择用户')
    return
  }
  const amount = Number(form.amountCents)
  if (!Number.isInteger(amount) || amount <= 0) {
    showToast('金额需为正整数（单位：分）')
    return
  }

  submitting.value = true
  try {
    const payload: WalletMutationPayload = {
      userId: selectedUserId.value,
      amountCents: amount,
      method: form.method.trim() || 'cash',
      voucherNo: form.voucherNo.trim(),
      remark: form.remark.trim(),
    }
    if (auth.role === 'super' && selectedUser.value?.tenantId) {
      payload.tenantId = Number(selectedUser.value.tenantId)
    }
    if (kind === 'recharge') {
      await walletApi.recharge(payload)
      showToast({ type: 'success', message: '充值成功' })
    } else {
      await walletApi.refund(payload)
      showToast({ type: 'success', message: '退款成功' })
    }
    form.amountCents = ''
    form.voucherNo = ''
    form.remark = ''
    await load()
  } catch {
    // 统一由 request.ts 提示
  } finally {
    submitting.value = false
  }
}

onMounted(load)
</script>

<template>
  <div class="page--plain">
    <NavBar title="充值 / 退款" />

    <div v-if="loading" class="page-loading">
      <Loading type="spinner" />
    </div>

    <template v-else>
      <CellGroup inset>
        <Cell title="选择用户" is-link @click="userPopup = true">
          <template #value>
            <span>{{ selectedUser ? `${selectedUser.name} · ${selectedUser.phone}` : '请选择' }}</span>
          </template>
        </Cell>
      </CellGroup>

      <div class="card card--pad balance-card">
        <div class="muted">当前余额（元）</div>
        <div class="balance-amount">¥{{ formatCents(selectedWallet?.balanceCents) }}</div>
      </div>

      <div class="card card--pad">
        <div class="section-title form-title">现金流水操作</div>
        <Field v-model="form.amountCents" label="金额(分)" type="digit" placeholder="请输入整数分" />
        <div class="muted amount-hint">折合人民币 ¥{{ amountYuan }}</div>
        <Field v-model="form.method" label="方式" placeholder="cash / wechat / alipay / bank" />
        <Field v-model="form.voucherNo" label="凭证号" placeholder="选填" />
        <Field v-model="form.remark" label="备注" placeholder="选填" />
        <div class="form-actions">
          <Button
            v-if="auth.canRefund"
            block
            type="primary"
            :loading="submitting"
            @click="submit('recharge')"
          >
            充值
          </Button>
          <Button
            v-if="auth.canRefund"
            block
            type="danger"
            plain
            :loading="submitting"
            @click="submit('refund')"
          >
            退款
          </Button>
        </div>
        <p class="muted refund-note">存在充电中/待结算订单或未关闭预约时，退款会被服务端拒绝。</p>
      </div>

      <div class="section-title">流水记录</div>
      <Empty v-if="ledgers.length === 0" image="search" description="暂无流水" />
      <CellGroup v-else inset>
        <Cell
          v-for="item in ledgers"
          :key="item.id"
          :title="ledgerText(item.type)"
          :label="formatTime(item.createdAt || item.createTime)"
        >
          <template #value>
            <div class="col ledger-value">
              <span :class="Number(item.amountCents) >= 0 ? 'income' : 'outcome'">
                {{ signedYuan(item.amountCents) }}
              </span>
              <Tag v-if="item.method" size="medium" type="default">{{ item.method }}</Tag>
            </div>
          </template>
        </Cell>
      </CellGroup>
    </template>

    <Popup v-model:show="userPopup" position="bottom" round>
      <div class="popup-title">选择用户</div>
      <RadioGroup v-model="selectedUserId" class="picker-list" @change="onUserChange">
        <Cell v-for="user in users" :key="user.id" clickable @click="onUserChange(user.id)">
          <template #title>
            <Radio :name="user.id">{{ user.name }} · {{ user.phone }}</Radio>
          </template>
          <template #value>
            <span class="muted">¥{{ formatCents(user.balanceCents) }}</span>
          </template>
        </Cell>
      </RadioGroup>
      <div class="popup-foot">
        <Button block type="primary" @click="userPopup = false">关闭</Button>
      </div>
    </Popup>
  </div>
</template>

<style scoped>
.balance-card {
  margin-top: 12px;
}

.balance-amount {
  margin-top: 6px;
  font-size: 28px;
  font-weight: 600;
  color: var(--ch-primary);
}

.form-title {
  margin: 0 0 8px;
}

.amount-hint {
  padding: 8px 0 0;
}

.refund-note {
  margin: 4px 0 0;
  line-height: 1.5;
}

.ledger-value {
  align-items: flex-end;
  gap: 4px;
}

.income {
  color: var(--ch-success);
  font-weight: 600;
}

.outcome {
  color: var(--ch-danger);
  font-weight: 600;
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
