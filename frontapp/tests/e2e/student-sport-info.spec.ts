import { expect, test } from '@playwright/test';

const mockBackendUrl =
  process.env.MOCK_BACKEND_URL ??
  `http://127.0.0.1:${process.env.MOCK_BACKEND_PORT ?? 8081}`;

test.describe('競技情報閲覧 (student)', () => {
  test.beforeEach(async ({ page, context, request }) => {
    await request.post(`${mockBackendUrl}/__reset`);
    await request.post(`${mockBackendUrl}/__set-user`, { data: { user: 'student' } });
    await request.post(`${mockBackendUrl}/__set-active-event`, { data: { event_id: 1 } });
    await request.post(`${mockBackendUrl}/api/admin/events/1/sports`, {
      data: {
        sport_id: 1,
        description: '5人制で実施します。',
        location: 'gym1',
        min_capacity: 5,
        max_capacity: 10
      }
    });
    await context.addCookies([{
      name: 'session_token',
      value: 'test-session-token',
      domain: 'localhost',
      path: '/'
    }]);

    await page.goto('/dashboard/student/sport-info');
  });

  test('開催競技の説明と会場を確認する', async ({ page }) => {
    await expect(page.getByRole('heading', { name: '競技一覧・詳細閲覧' })).toBeVisible();
    await expect(page.getByRole('heading', { name: 'バスケットボール' })).toBeVisible();
    await expect(page.getByText('5人制で実施します。')).toBeVisible();
    await expect(page.getByText('第一体育館')).toBeVisible();
  });
});
