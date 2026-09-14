import { expect, test } from '@playwright/test';

const mockBackendUrl =
  process.env.MOCK_BACKEND_URL ??
  `http://127.0.0.1:${process.env.MOCK_BACKEND_PORT ?? 8081}`;

test.describe('クラス進捗確認 (student)', () => {
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

    await page.goto('/dashboard/student/class-info');
  });

  test('出席率と次の対戦予定を確認する', async ({ page }) => {
    await expect(page.getByRole('heading', { name: 'クラス情報' })).toBeVisible();
    await expect(page.getByText('1A', { exact: true }).first()).toBeVisible();
    await expect(page.getByText('40 名')).toBeVisible();
    await expect(page.getByText('36 名').first()).toBeVisible();
    await expect(page.getByText('90.0%')).toBeVisible();

    await expect(page.getByRole('heading', { name: '勝ち進み状況' })).toBeVisible();
    await expect(page.getByText('バスケットボール', { exact: true }).first()).toBeVisible();
    await expect(page.getByText(/決勝\s*・対 1B/)).toBeVisible();
    await expect(page.getByText('開始予定: 2025/04/01 09:30')).toBeVisible();
  });
});
