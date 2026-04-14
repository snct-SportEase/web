import { defineConfig, devices } from '@playwright/test';

export default defineConfig({
  testDir: './e2e',
  fullyParallel: true,
  forbidOnly: !!process.env.CI,
  retries: process.env.CI ? 2 : 0,
  workers: process.env.CI ? 1 : undefined,
  reporter: 'html',
  use: {
    baseURL: 'http://localhost:5000',
    trace: 'on-first-retry',
  },
  projects: [
    {
      name: 'chromium',
      use: { ...devices['Desktop Chrome'] },
    },
  ],
  webServer: [
    {
      command: 'node scripts/mock-backend.js',
      url: 'http://127.0.0.1:8081/health',
      reuseExistingServer: !process.env.CI,
    },
    {
      command: 'npm run dev',
      url: 'http://localhost:5000',
      reuseExistingServer: !process.env.CI,
      env: {
        ...process.env,
        BACKEND_URL: 'http://127.0.0.1:8081',
        PUBLIC_BACKEND_URL: 'http://127.0.0.1:8081',
      },
    },
  ],
});
