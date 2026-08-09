import { defineConfig, loadEnv } from 'vite'
import react from '@vitejs/plugin-react'
import tailwindcss from '@tailwindcss/vite'

export default defineConfig(({ command, mode }) => {
  const environment = loadEnv(mode, process.cwd(), '')

  if (command === 'build' && !environment.VITE_API_BASE_URL?.trim()) {
    throw new Error('VITE_API_BASE_URL must be set for production builds.')
  }

  const configuredBase =
    command === 'serve'
      ? environment.VITE_DEV_PROXY_BASE?.trim() || ''
      : ''

  const base = configuredBase
    ? `/${configuredBase.replace(/^\/+|\/+$/g, '')}/`
    : '/'

  const proxyPrefix = base === '/' ? '' : base.slice(0, -1)
  const allowedHost = environment.VITE_DEV_ALLOWED_HOST?.trim()
  const apiTarget =
    environment.VITE_DEV_API_TARGET?.trim() ||
    'http://127.0.0.1:8080'

  return {
    base,
    plugins: [react(), tailwindcss()],
    server: {
      allowedHosts: allowedHost ? [allowedHost] : [],
      proxy: {
        [`${proxyPrefix}/api`]: {
          target: apiTarget,
          changeOrigin: true,
          rewrite: proxyPrefix
            ? (path) => path.replace(proxyPrefix, '')
            : undefined,
        },
      },
    },
  }
})
