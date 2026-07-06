import { defineConfig, devices } from '@playwright/test';

const IS_CI = !!process.env.CI;

const gpuArgs: string[] = [];

export default defineConfig({
  testDir: './specs',
  timeout: 60_000,
  expect: { timeout: 15_000 },
  fullyParallel: false,
  retries: 0,
  workers: 1,
  reporter: [['list'], ['html', { open: 'never' }]],

  use: {
    baseURL: IS_CI ? 'http://localhost:8080' : 'http://localhost:5173',
    trace: 'on-first-retry',
    screenshot: 'only-on-failure',
  },

  projects: [
    {
      name: 'chromium',
      use: IS_CI
        ? {
            ...devices['Desktop Chrome'],
            launchOptions: {
              args: gpuArgs.length > 0 ? gpuArgs : undefined,
            },
          }
        : {
            ...devices['Desktop Chrome'],
            channel: 'chromium',
            launchOptions: {
              executablePath: '/usr/bin/chromium',
              args: gpuArgs.length > 0 ? gpuArgs : undefined,
            },
          },
    },
  ],

  // In CI servers are started manually; locally use webServer for convenience.
  webServer: IS_CI
    ? undefined
    : [
        {
          command: 'cd .. && make dev-backend',
          port: 8080,
          timeout: 15_000,
          reuseExistingServer: true,
        },
        {
          command: 'cd ../frontend && pnpm run dev',
          port: 5173,
          timeout: 15_000,
          reuseExistingServer: true,
        },
      ],
});
