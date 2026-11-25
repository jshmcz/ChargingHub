<script setup lang="ts">
import { computed } from 'vue'
import { useRoute } from 'vue-router'
import { Tabbar, TabbarItem } from 'vant'

const route = useRoute()
const showTabbar = computed(() => Boolean(route.meta.showTabbar))
</script>

<template>
  <div class="app-root">
    <router-view v-slot="{ Component }">
      <component :is="Component" />
    </router-view>

    <Tabbar v-show="showTabbar" route :placeholder="true" :safe-area-inset-bottom="true">
      <TabbarItem to="/home" icon="home-o">首页</TabbarItem>
      <TabbarItem to="/wallet" icon="wallet-o">钱包</TabbarItem>
      <TabbarItem to="/orders" icon="orders-o">订单</TabbarItem>
      <TabbarItem to="/profile" icon="user-o">我的</TabbarItem>
    </Tabbar>
  </div>
</template>
