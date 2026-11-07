import { createWebKit } from '@novelia/web-kit';
import { PiniaColada } from '@pinia/colada';
import { createPinia } from 'pinia';
import { createApp } from 'vue';

import App from './App.vue';
import router from './router';
import './styles.css';

const webKit = createWebKit({
  auth: {
    app: 'c',
    url: __AUTH_URL__,
  },
  brand: '漫画',
  strikes: { enabled: false },
  repository: {
    url: 'https://github.com/auto-novel/auto-comic',
    buildTime: __BUILD_TIME__,
    commitSha: __COMMIT_SHA__,
  },
});

router.beforeEach(async () => {
  await webKit.checkSignedIn();
});

createApp(App)
  .use(webKit)
  .use(createPinia())
  .use(PiniaColada, {
    queryOptions: {
      staleTime: 60_000,
      gcTime: 5 * 60_000,
    },
  })
  .use(router)
  .mount('#app');
