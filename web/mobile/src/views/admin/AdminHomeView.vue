<script setup lang="ts">
import { Cell, CellGroup, NavBar, Tag } from 'vant'
import { useAuthStore } from '@/stores/auth'
import { roleText } from '@/utils/format'

const auth = useAuthStore()
</script>

<template>
  <div class="page--plain">
    <NavBar title="管理后台" />

    <div class="card card--pad admin-head">
      <div class="row-between">
        <span>当前角色</span>
        <Tag type="primary" size="medium">{{ roleText(auth.role) }}</Tag>
      </div>
      <div class="row-between admin-line">
        <span>租户 ID</span>
        <span class="muted ellipsis tenant-id">{{ auth.tenantId || '-' }}</span>
      </div>
    </div>

    <CellGroup inset>
      <Cell title="用户管理" label="查询、创建与编辑用户" icon="friends-o" is-link to="/admin/users" />
      <Cell title="场站管理" label="场站、充电桩与充电枪" icon="shop-o" is-link to="/admin/stations" />
      <Cell title="充值 / 退款" label="钱包余额与现金流水" icon="balance-list-o" is-link to="/admin/wallet" />
    </CellGroup>

    <div class="card card--pad admin-note">
      <p class="muted">操作员账号可管理用户与场站，但无退款权限。</p>
      <p class="muted">所有金额以分为单位提交，页面展示时换算为元。</p>
    </div>
  </div>
</template>

<style scoped>
.admin-head {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.admin-line {
  font-size: 13px;
}

.tenant-id {
  max-width: 60%;
}

.admin-note p {
  margin: 4px 0;
  line-height: 1.5;
}
</style>
