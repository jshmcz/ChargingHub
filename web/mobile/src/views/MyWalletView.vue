<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { Cell, CellGroup, Empty, Loading, NavBar, PullRefresh, Tag } from 'vant'
import { walletApi } from '@/api/wallets'
import type { Ledger, Wallet } from '@/api/types'
import { formatCents, formatTime, ledgerText, signedYuan } from '@/utils/format'

const wallets = ref<Wallet[]>([])
const ledgers = ref<Ledger[]>([])
const loading = ref(true)
const refreshing = ref(false)

async function load(): Promise<void> {
  loading.value = true
  try {
    const [walletData, ledgerData] = await Promise.all([
      walletApi.myWallets(),
      walletApi.myLedgers(),
    ])
    wallets.value = walletData || []
    ledgers.value = ledgerData || []
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
    <NavBar title="我的钱包" />

    <Empty
      v-if="!loading && wallets.length === 0"
      image="search"
      description="暂无钱包，请联系租户管理员充值"
    />
    <div v-for="item in wallets" :key="item.id" class="balance-card">
      <div class="muted">{{ item.tenantName || '默认租户' }} · 钱包余额（元）</div>
      <div class="balance-amount">¥{{ formatCents(item.balanceCents) }}</div>
      <div class="balance-foot">
        <span v-if="item.status === 'frozen'" class="frozen">账户已冻结，请联系管理员</span>
        <span v-else>充电启动与预约将校验该余额</span>
      </div>
    </div>

    <div class="section-title">资金流水</div>

    <PullRefresh v-model="refreshing" @refresh="onRefresh">
      <div v-if="loading" class="page-loading">
        <Loading type="spinner" />
      </div>
      <Empty v-else-if="ledgers.length === 0" image="search" description="暂无流水记录" />
      <CellGroup v-else inset>
        <Cell
          v-for="item in ledgers"
          :key="item.id"
          :title="ledgerText(item.type)"
        >
          <template #value>
            <div class="col ledger-value">
              <span :class="Number(item.amountCents) >= 0 ? 'income' : 'outcome'">
                {{ signedYuan(item.amountCents) }}
              </span>
              <Tag v-if="item.method" size="medium" type="default">{{ item.method }}</Tag>
            </div>
          </template>
          <template #label>
            <div class="col">
              <span>{{ formatTime(item.createdAt || item.createTime) }}</span>
              <span v-if="item.remark" class="ellipsis remark">{{ item.remark }}</span>
            </div>
          </template>
        </Cell>
      </CellGroup>
    </PullRefresh>
  </div>
</template>

<style scoped>
.balance-card {
  margin: 12px;
  padding: 24px 20px;
  border-radius: var(--ch-radius);
  color: #fff;
  background: linear-gradient(135deg, #1989fa, #4fc0ff);
}

.balance-card :deep(.muted) {
  color: rgba(255, 255, 255, 0.8);
}

.balance-amount {
  margin: 8px 0 12px;
  font-size: 34px;
  font-weight: 600;
  line-height: 1.2;
}

.balance-foot {
  font-size: 12px;
  opacity: 0.9;
}

.frozen {
  color: #fff0f0;
}

.ledger-value {
  align-items: flex-end;
}

.income {
  color: var(--ch-success);
  font-weight: 600;
}

.outcome {
  color: var(--ch-danger);
  font-weight: 600;
}

.remark {
  max-width: 180px;
}
</style>
