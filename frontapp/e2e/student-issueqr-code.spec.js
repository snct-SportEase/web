import { expect, test } from '@playwright/test';

const mockBackendUrl = process.env.MOCK_BACKEND_URL ?? 'http://127.0.0.1:8081';

test.describe('参加競技確認 (student)', () => {
  test.beforeEach(async ({ page, context, request }) => {
    await request.post(`${mockBackendUrl}/__reset`);
    await context.addCookies([{ name: 'session_token', value: 'test-session-token', domain: 'localhost', path: '/' }]);
    await page.goto('/dashboard/student/issueqr-code');
    await expect(page.getByRole('heading', { name: '参加競技確認' })).toBeVisible();
  });

  test('開催中イベントの参加競技を確認できる', async ({ page }) => {
    await expect(page.getByText('開催中イベント')).toBeVisible();
    await expect(page.getByText('バスケットボール')).toBeVisible();
    await expect(page.getByText('Team A')).toBeVisible();
  });
});
