import { htmlBootstrapPlugin } from '@novelia/web-kit/vite';
import tailwindcss from '@tailwindcss/vite';
import vue from '@vitejs/plugin-vue';
import { fileURLToPath, URL } from 'node:url';
import { defineConfig, loadEnv } from 'vite';

export default defineConfig(({ command, mode }) => {
  const env = loadEnv(
    mode,
    fileURLToPath(new URL('.', import.meta.url)),
    'VITE_',
  );
  const isLocalApi = env.VITE_LOCAL_API === 'true';

  return {
    define: {
      __AUTH_URL__: JSON.stringify(
        command === 'serve' ? '/auth-proxy/' : 'https://auth.novelia.cc',
      ),
      __BUILD_TIME__: JSON.stringify(
        env.VITE_BUILD_TIME || new Date().toISOString(),
      ),
      __COMMIT_SHA__: JSON.stringify(env.VITE_COMMIT_SHA || 'unknown'),
    },
    plugins: [htmlBootstrapPlugin(), vue(), tailwindcss()],
    build: {
      cssCodeSplit: false,
      rolldownOptions: {
        output: {
          // 按更新频率分离应用代码、web-kit 和第三方依赖，减少更新时的缓存失效范围；
          // 第三方依赖集中为 vendor，兼顾高带宽、高 RTT 环境下的请求数量与缓存复用。
          manualChunks: (id) => {
            const normalizedId = id.replaceAll('\\', '/');
            if (normalizedId.includes('/node_modules/@novelia/web-kit/')) {
              return 'web-kit';
            }
            if (normalizedId.includes('/node_modules/')) {
              return 'vendor';
            }
            return undefined;
          },
        },
      },
    },
    optimizeDeps: {
      exclude: ['@novelia/web-kit'],
    },
    server: {
      port: 5173,
      proxy: {
        '/auth-proxy/': {
          target: 'https://auth.novelia.cc',
          changeOrigin: true,
          rewrite: (path) => path.replace(/^\/auth-proxy/, ''),
        },
        '/api': {
          target: isLocalApi ? 'http://localhost:7001' : 'https://c.novelia.cc',
          changeOrigin: true,
          rewrite: isLocalApi
            ? (path) => path.replace(/^\/api/, '')
            : undefined,
        },
      },
    },
    resolve: {
      alias: {
        '@': fileURLToPath(new URL('./src', import.meta.url)),
      },
    },
  };
});
