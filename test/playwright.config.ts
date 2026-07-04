import { defineConfig, devices } from '@playwright/test';

export default defineConfig({
  testDir: './specs',
  timeout: 30_000,
  expect: { timeout: 10_000 },
  fullyParallel: false,
  retries: 0,
  workers: 1,
  reporter: [['list'], ['html', { open: 'never' }]],

  use: {
    baseURL: 'http://localhost:5173',
    trace: 'on-first-retry',
    screenshot: 'only-on-failure',
  },

  projects: [
    {
      name: 'chromium',
      use: {
        ...devices['Desktop Chrome'],
        // Use system Chromium (no network to download Playwright's bundled browser)
        channel: 'chromium',
        launchOptions: {
          executablePath: '/usr/bin/chromium',
        },
      },
    },
  ],

  webServer: [
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
