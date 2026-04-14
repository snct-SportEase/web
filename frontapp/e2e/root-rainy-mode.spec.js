import { test, expect } from '@playwright/test';

const mockBackendUrl = process.env.MOCK_BACKEND_URL ?? 'http://127.0.0.1:8081';

test.describe('雨天時モード管理 (root)', () => {
  test.beforeEach(async ({ page, context, request }) => {
    await request.post(`${mockBackendUrl}/__reset`);
    await context.addCookies([{ name: 'session_token', value: 'test-session-token', domain: 'localhost', path: '/' }]);
    await page.goto('/dashboard/root/rainy-mode');
  });

  test('アクティブ大会を表示できる', async ({ page }) => {
    await expect(page.getByRole('heading', { name: '雨天時モード管理' })).toBeVisible();
    await expect(page.getByText('2025春季スポーツ大会')).toBeVisible();
  });

  test('雨天時モードを有効化できる', async ({ page }) => {
    page.once('dialog', async (dialog) => { await dialog.accept(); });

    const req = page.waitForRequest((request) => request.url().endsWith('/api/root/events/1/rainy-mode') && request.method() === 'PUT');
    await page.getByRole('button', { name: '有効にする' }).click();
    const request = await req;
    expect(JSON.parse(request.postData() ?? '{}')).toEqual({ is_rainy_mode: true });
  });
});
