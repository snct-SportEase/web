import { test, expect } from '@playwright/test';

const mockBackendUrl = process.env.MOCK_BACKEND_URL ?? 'http://127.0.0.1:8081';

test.describe('大会要項アップロード (root)', () => {
  test.beforeEach(async ({ page, context, request }) => {
    await request.post(`${mockBackendUrl}/__reset`);
    await context.addCookies([{ name: 'session_token', value: 'test-session-token', domain: 'localhost', path: '/' }]);
    await page.goto('/dashboard/root/competition-guidelines-upload');
    await expect(page.getByRole('heading', { name: '大会要項アップロード' })).toBeVisible();
  });

  test('大会選択を表示できる', async ({ page }) => {
    await expect(page.getByLabel('大会選択')).toHaveValue('1');
  });

  test('PDF未選択ではアップロードしない', async ({ page }) => {
    await expect(page.getByRole('button', { name: 'アップロード' })).toBeDisabled();
  });
});
