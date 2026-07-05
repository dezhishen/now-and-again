import { defineConfig, type Plugin } from 'vite'
import vue from '@vitejs/plugin-vue'
import { resolve } from 'path'

function stripTestAttrsPlugin(): Plugin {
  return {
    name: 'strip-test-attrs',
    enforce: 'pre',
    transform(code, id) {
      if (!id.endsWith('.vue')) return null

      const stripped = code
        .replace(/\s(?:data-testid|data-test|data-cy)=(['"]).*?\1/g, '')
        .replace(/\s(?::|v-bind:)(?:data-testid|data-test|data-cy)=(['"]).*?\1/g, '')

      return stripped === code ? null : stripped
    },
  }
}

export default defineConfig(({ command }) => {
  // Keep test attributes when explicitly requested (for CI E2E against built assets).
  // Default behavior remains stripping in production build.
  const stripTestAttrs = command === 'build' && process.env.NA_STRIP_TEST_ATTRS !== '0'

  return {
    plugins: [
      ...(stripTestAttrs ? [stripTestAttrsPlugin()] : []),
      vue(),
    ],
    resolve: {
      alias: {
        '@': resolve(__dirname, 'src'),
      },
    },
    server: {
      host: '0.0.0.0',
      port: 5173,
      proxy: {
        '/api': {
          target: 'http://localhost:8080',
          changeOrigin: true,
        },
        '/uploads': {
          target: 'http://localhost:8080',
          changeOrigin: true,
        },
      },
    },
  }
})
