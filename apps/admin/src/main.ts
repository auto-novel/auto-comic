import { createWebKit } from '@novelia/web-kit';
import { createApp } from 'vue';

import App from './App.vue';
import router from './router';
import './style.css';

const webKit = createWebKit({
  auth: {
    app: 'c',
    url: __AUTH_URL__,
  },
  brand: 'Auto Comic',
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

createApp(App).use(webKit).use(router).mount('#app');
