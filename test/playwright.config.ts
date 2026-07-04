import { defineConfig, devices } from '@playwright/test';

const IS_CI = !!process.env.CI;

export default defineConfig({
  testDir: './specs',
  timeout: 30_000,
  expect: { timeout: 10_000 },
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
        ? { ...devices['Desktop Chrome'] }
        : {
            ...devices['Desktop Chrome'],
            channel: 'chromium',
            launchOptions: { executablePath: '/usr/bin/chromium' },
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
