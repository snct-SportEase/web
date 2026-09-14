import { expect, test } from '@playwright/test';

const mockBackendUrl =
  process.env.MOCK_BACKEND_URL ??
  `http://127.0.0.1:${process.env.MOCK_BACKEND_PORT ?? 8081}`;

test.describe('参加本登録確認 (admin)', () => {
  test.beforeEach(async ({ page, context, request }) => {
    await request.post(`${mockBackendUrl}/__reset`);
    await request.post(`${mockBackendUrl}/__set-user`, { data: { user: 'admin' } });
    await request.post(`${mockBackendUrl}/__set-active-event`, { data: { event_id: 1 } });
    await request.post(`${mockBackendUrl}/api/admin/events/1/sports`, {
      data: { sport_id: 1, min_capacity: 2, max_capacity: 5 }
    });
    await request.post(`${mockBackendUrl}/api/admin/class-team/assign-members`, {
      data: { class_id: 1, sport_id: 1, user_ids: ['user-1'] }
    });
    await context.addCookies([{
      name: 'session_token',
      value: 'test-session-token',
      domain: 'localhost',
      path: '/'
    }]);

    const sportsResponse = page.waitForResponse((response) =>
      response.url().endsWith('/api/events/1/sports') && response.request().method() === 'GET'
    );
    await page.goto('/dashboard/admin/confirmed-participants');
    await expect(page.getByRole('heading', { name: '参加本登録済みメンバー確認' })).toBeVisible();
    await sportsResponse;
  });

  test('本登録済みメンバーと最低人数不足を確認できる', async ({ page }) => {
    const selects = page.locator('select');
    await selects.first().selectOption({ label: '1A' });
    await expect(selects).toHaveCount(2);
    await selects.nth(1).selectOption({ label: 'バスケットボール' });

    await expect(page.getByText('student1@sendai-nct.jp')).toBeVisible();
    await expect(page.getByText('1 / 2 人')).toBeVisible();
    await expect(page.getByText('（最低人数未満）')).toBeVisible();
    await expect(page.getByText(/1Aクラスのバスケットボールの参加本登録済みメンバー数/)).toBeVisible();
  });
});
