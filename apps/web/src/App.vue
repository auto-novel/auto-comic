<script setup lang="ts">
import {
  WebKitApp,
  WebKitLayout,
  type WebKitMenuOption,
} from '@novelia/web-kit';
import { HomeOutlined } from '@vicons/material';
import { RouterView, useRoute } from 'vue-router';

const route = useRoute();
const navigationOptions: WebKitMenuOption[] = [
  {
    key: 'home',
    label: '首页',
    icon: HomeOutlined,
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
      <RouterView />
    </WebKitLayout>
  </WebKitApp>
</template>
