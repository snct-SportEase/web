import { expect, test } from '@playwright/test';

const mockBackendUrl = process.env.MOCK_BACKEND_URL ?? 'http://127.0.0.1:8081';

test.describe('トーナメント生成・管理 (root)', () => {
  test.beforeEach(async ({ page, context, request }) => {
    await request.post(`${mockBackendUrl}/__reset`);
    await context.addCookies([{ name: 'session_token', value: 'test-session-token', domain: 'localhost', path: '/' }]);
    await page.goto('/dashboard/root/tournament-management');
    await expect(page.getByRole('heading', { name: 'トーナメント生成・管理' })).toBeVisible();
  });

  test('ページを表示できる', async ({ page }) => {
    await expect(page.getByRole('button', { name: 'トーナメントプレビューを生成' })).toBeVisible();
  });

  test('トーナメントプレビューを生成して保存できる', async ({ page }) => {
    page.on('dialog', (dialog) => {
      void dialog.accept().catch(() => {});
    });

    const previewRequest = page.waitForRequest((request) => request.url().endsWith('/api/root/events/1/tournaments/generate-preview') && request.method() === 'POST');
    await page.getByRole('button', { name: 'トーナメントプレビューを生成' }).click();
    await previewRequest;
    await expect(page.getByRole('button', { name: 'プレビューをDBに保存' })).toBeVisible();

    const saveRequest = page.waitForRequest((request) => request.url().endsWith('/api/root/events/1/tournaments/bulk-create') && request.method() === 'POST');
    await page.getByRole('button', { name: 'プレビューをDBに保存' }).click();
    await saveRequest;
  });
});
