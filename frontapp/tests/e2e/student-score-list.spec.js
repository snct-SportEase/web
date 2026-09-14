import { expect, test } from '@playwright/test';

const mockBackendUrl =
  process.env.MOCK_BACKEND_URL ??
  `http://127.0.0.1:${process.env.MOCK_BACKEND_PORT ?? 8081}`;

test.describe('大会得点一覧 (student)', () => {
  test.beforeEach(async ({ page, context, request }) => {
    await request.post(`${mockBackendUrl}/__reset`);
    await request.post(`${mockBackendUrl}/__set-user`, { data: { user: 'student' } });
    await request.post(`${mockBackendUrl}/__set-active-event`, { data: { event_id: 1 } });
    await context.addCookies([{
      name: 'session_token',
      value: 'test-session-token',
      domain: 'localhost',
      path: '/'
    }]);

    await page.goto('/dashboard/student/score-list');
  });

  test('クラス順位と得点内訳を確認する', async ({ page }) => {
    await expect(page.getByRole('heading', { name: '点数一覧' })).toBeVisible();
    await expect(page.getByText('1A', { exact: true })).toBeVisible();
    await expect(page.getByText('1B', { exact: true })).toBeVisible();
    await expect(page.getByText('60', { exact: true })).toBeVisible();
    await expect(page.getByText('50', { exact: true })).toBeVisible();

    await page.getByText('点数項目を表示').first().click();
    const firstClass = page.locator('details').first();
    await expect(firstClass.getByText('出席点:')).toBeVisible();
    await expect(firstClass.getByText('昼競技:')).toBeVisible();
    await expect(firstClass.getByText('10', { exact: true })).toBeVisible();
    await expect(firstClass.getByText('20', { exact: true })).toBeVisible();
  });
});
