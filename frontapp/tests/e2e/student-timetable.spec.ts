import { expect, test } from '@playwright/test';

const mockBackendUrl =
  process.env.MOCK_BACKEND_URL ??
  `http://127.0.0.1:${process.env.MOCK_BACKEND_PORT ?? 8081}`;

test.describe('大会タイムテーブル (student)', () => {
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

    await page.goto('/dashboard/student/timetable');
  });

  test('試合開始時刻・対戦カード・進行状態を確認する', async ({ page }) => {
    await expect(page.getByRole('heading', { name: 'タイムテーブル' })).toBeVisible();
    await expect(page.getByText('09:30')).toBeVisible();
    await expect(page.getByText('10:30')).toBeVisible();
    await expect(page.getByRole('heading', { name: 'バスケットボール', exact: true })).toBeVisible();
    await expect(page.getByText('1A vs 1B')).toHaveCount(2);
    await expect(page.getByText('進行中', { exact: true })).toHaveCount(2);
  });
});
