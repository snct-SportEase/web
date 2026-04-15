import { defineConfig, devices } from '@playwright/test';

const appPort = Number(process.env.PLAYWRIGHT_APP_PORT ?? 5000);
const backendPort = Number(process.env.MOCK_BACKEND_PORT ?? 8081);

export default defineConfig({
  testDir: './e2e',
  fullyParallel: false,
  forbidOnly: !!process.env.CI,
  retries: process.env.CI ? 2 : 0,
  workers: 1,
  reporter: 'html',
  use: {
    baseURL: `http://localhost:${appPort}`,
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
      url: `http://127.0.0.1:${backendPort}/health`,
      reuseExistingServer: !process.env.CI,
      env: {
        ...process.env,
        MOCK_BACKEND_PORT: String(backendPort),
        MOCK_BACKEND_URL: `http://127.0.0.1:${backendPort}`
      },
    },
    {
      command: `npm run dev -- --port ${appPort}`,
      url: `http://localhost:${appPort}`,
      reuseExistingServer: !process.env.CI,
      env: {
        ...process.env,
        BACKEND_URL: `http://127.0.0.1:${backendPort}`,
        PUBLIC_BACKEND_URL: `http://127.0.0.1:${backendPort}`,
        MOCK_BACKEND_URL: `http://127.0.0.1:${backendPort}`
      },
    },
  ],
});
