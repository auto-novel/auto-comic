<script setup lang="ts">
import {
  WebKitApp,
  WebKitLayout,
  type WebKitMenuOption,
  useWebKit,
} from '@novelia/web-kit';
import { DashboardOutlined } from '@vicons/material';
import { RouterView, useRoute } from 'vue-router';

const route = useRoute();
const { whoami } = useWebKit();
const navigationOptions: WebKitMenuOption[] = [
  {
    key: 'home',
    label: '概览',
    icon: DashboardOutlined,
    to: { name: 'home' },
  },
];
</script>

<template>
  <WebKitApp>
    <WebKitLayout
      :navigation-options="navigationOptions"
      :account-options="[]"
      :selected-navigation-key="String(route.name ?? '')"
    >
      <div class="page-container py-4 md:py-6">
        <RouterView v-if="whoami.isAdmin" />
        <section v-else-if="whoami.isSignedIn" class="py-12 text-center">
          <p class="text-4xl font-semibold text-muted">403</p>
          <h1 class="mt-4 text-xl font-medium text-ink">需要管理员权限</h1>
          <p class="mt-2 text-muted">当前账号没有访问管理后台的权限。</p>
        </section>
        <p v-else class="py-12 text-center text-muted">
          请先通过右上角登录账号，再访问管理后台。
        </p>
      </div>
    </WebKitLayout>
  </WebKitApp>
</template>
