import { defineConfig, devices } from '@playwright/test';

// Dedicated config for tutorial screenshot generation.
// Not used in CI — run manually with: npm run screenshots
export default defineConfig({
  testDir: './screenshots',
  timeout: 60_000,
  expect: { timeout: 15_000 },
  fullyParallel: false,
  retries: 0,
  workers: 1,
  reporter: [['list']],

  use: {
    baseURL: 'http://localhost:5173',
    trace: 'off',
    screenshot: 'off',
  },

  projects: [
    {
      name: 'chromium',
      use: {
        ...devices['Desktop Chrome'],
        channel: 'chromium',
        launchOptions: { executablePath: '/usr/bin/chromium' },
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
