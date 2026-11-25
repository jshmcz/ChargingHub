<template>
  <div class="page-container">
    <div class="page-header">
      <h2 class="page-title">概览</h2>
      <el-button :loading="loading" @click="load">刷新</el-button>
    </div>
    <div class="stats-grid" v-loading="loading">
      <el-card v-for="stat in stats" :key="stat.label" class="stat-card" shadow="never">
        <div class="text-muted">{{ stat.label }}</div>
        <div class="stat-value">{{ stat.value }}</div>
      </el-card>
    </div>
  </div>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { ElMessage } from 'element-plus'
import { useAuthStore } from '@/stores/auth'
import { tenantsApi } from '@/api/tenants'
import { usersApi } from '@/api/users'
import { stationsApi } from '@/api/stations'
import { pilesApi } from '@/api/piles'
import { ledgersApi, walletsApi } from '@/api/wallets'
import { ordersApi } from '@/api/orders'
import { reservationsApi } from '@/api/reservations'
import { formatCents } from '@/utils/format'

interface Stat {
  label: string
  value: string
}

const auth = useAuthStore()
const loading = ref(false)
const stats = ref<Stat[]>([])

function errorText(error: unknown): string {
  return error instanceof Error ? error.message : '加载失败'
}

async function load() {
  loading.value = true
  try {
    if (auth.isSuper) {
      const [tenants, users] = await Promise.all([tenantsApi.list(), usersApi.list()])
      stats.value = [
        { label: '租户总数', value: String(tenants.length) },
        { label: '用户总数', value: String(users.length) },
      ]
    } else if (auth.isEndUser) {
      const [wallet, ledgers, orders, reservations] = await Promise.all([
        walletsApi.my(),
        ledgersApi.my(),
        ordersApi.my(),
        reservationsApi.my(),
      ])
      stats.value = [
        { label: '我的余额', value: formatCents(wallet.balanceCents) },
        { label: '我的流水', value: String(ledgers.length) },
        { label: '我的订单', value: String(orders.length) },
        { label: '我的预约', value: String(reservations.length) },
      ]
    } else {
      const [stations, piles, users, wallets, orders, reservations] = await Promise.all([
        stationsApi.list(),
        pilesApi.list(),
        usersApi.list(),
        walletsApi.list(),
        ordersApi.list(),
        reservationsApi.list(),
      ])
      const totalBalance = wallets.reduce((sum, wallet) => sum + (wallet.balanceCents || 0), 0)
      stats.value = [
        { label: '场站总数', value: String(stations.length) },
        { label: '充电桩总数', value: String(piles.length) },
        { label: '用户总数', value: String(users.length) },
        { label: '钱包总余额', value: formatCents(totalBalance) },
        { label: '订单总数', value: String(orders.length) },
        { label: '预约总数', value: String(reservations.length) },
      ]
    }
  } catch (error) {
    ElMessage.error(errorText(error))
  } finally {
    loading.value = false
  }
}

onMounted(load)
</script>
