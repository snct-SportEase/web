import { expect, test } from '@playwright/test';

const mockBackendUrl =
  process.env.MOCK_BACKEND_URL ??
  `http://127.0.0.1:${process.env.MOCK_BACKEND_PORT ?? 8081}`;

test.describe('トーナメント閲覧 (student)', () => {
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

    await page.goto('/dashboard/student/tournament');
  });

  test('開催中競技のトーナメントを一覧確認する', async ({ page }) => {
    await expect(page.getByRole('heading', { name: 'トーナメント一覧' })).toBeVisible();
    await expect(page.getByRole('heading', { name: 'バスケットボール', exact: true })).toBeVisible();
    await expect(page.getByRole('heading', { name: 'バスケットボール 敗者復活' })).toBeVisible();
  });
});
