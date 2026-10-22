import { defineConfig, loadEnv } from 'vite';
import vue from '@vitejs/plugin-vue';
import { fileURLToPath, URL } from 'node:url';
import { readFileSync } from 'node:fs';

export default defineConfig(({ mode }) => {
  const env = loadEnv(mode, process.cwd(), '');
  const target = env.API_PROXY_TARGET || 'http://127.0.0.1:8080';
  const { version } = JSON.parse(readFileSync(new URL('./package.json', import.meta.url), 'utf8'));
  const buildInfo = { version, builtAt: new Date().toISOString() };
  return {
    define: { __BUILD_INFO__: JSON.stringify(buildInfo) },
    plugins: [vue(), {
      name: 'build-info',
      generateBundle() {
        this.emitFile({ type: 'asset', fileName: 'build-info.json', source: JSON.stringify(buildInfo, null, 2) + '\n' });
      },
    }],
    resolve: { alias: { '@': fileURLToPath(new URL('./src', import.meta.url)) } },
    server: { proxy: { '/api': { target, changeOrigin: true }, '/uploads': { target, changeOrigin: true } } },
    build: { outDir: '.verification-build', emptyOutDir: true },
  };
});
