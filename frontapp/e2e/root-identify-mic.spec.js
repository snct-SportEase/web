import { test, expect } from '@playwright/test';

const mockBackendUrl = process.env.MOCK_BACKEND_URL ?? 'http://127.0.0.1:8081';

test.describe('MIC確認 (root)', () => {
  test.beforeEach(async ({ page, context, request }) => {
    await request.post(`${mockBackendUrl}/__reset`);
    await context.addCookies([{ name: 'session_token', value: 'test-session-token', domain: 'localhost', path: '/' }]);
    await page.goto('/dashboard/root/identify-mic');
  });

  test('MIC結果を表示できる', async ({ page }) => {
    await expect(page.getByRole('heading', { name: 'MIC確認' })).toBeVisible();
    await expect(page.getByText('MIC Class')).toBeVisible();
    await expect(page.getByText('1A')).toBeVisible();
    await expect(page.getByText('120')).toBeVisible();
  });
});
