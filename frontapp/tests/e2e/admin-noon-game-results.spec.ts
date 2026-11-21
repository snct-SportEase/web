import { expect, test } from '@playwright/test';

const mockBackendUrl =
  process.env.MOCK_BACKEND_URL ??
  `http://127.0.0.1:${process.env.MOCK_BACKEND_PORT ?? 8081}`;

test.describe('昼競技結果入力 (admin)', () => {
  test.beforeEach(async ({ page, context, request }) => {
    await request.post(`${mockBackendUrl}/__reset`);
    await request.post(`${mockBackendUrl}/__set-active-event`, { data: { event_id: 1 } });
    await request.post(`${mockBackendUrl}/api/root/events/1/noon-game/session`, {
      data: { name: '昼休みドッジボール', mode: 'match', allow_manual_points: false }
    });
    await request.put(`${mockBackendUrl}/api/root/events/1/noon-game/sessions/1`, {
      data: { status: 'published' }
    });
    await request.post(`${mockBackendUrl}/api/root/noon-game/sessions/1/matches`, {
      data: {
        title: '決勝戦',
        home_display_name: '1A',
        away_display_name: '1B',
        scheduled_at: '2025-04-01T12:30:00Z',
        status: 'scheduled',
        allow_draw: false
      }
    });
    await request.post(`${mockBackendUrl}/__set-user`, { data: { user: 'admin' } });
    await context.addCookies([{
      name: 'session_token',
      value: 'test-session-token',
      domain: 'localhost',
      path: '/'
    }]);
    page.on('dialog', (dialog) => void dialog.accept());

    await page.goto('/dashboard/admin/noon-game-results');
    await expect(page.getByRole('heading', { name: '昼競技結果入力' })).toBeVisible();
    await expect(page.getByText('決勝戦', { exact: true })).toBeVisible();
  });

  test('昼競技の勝者を登録して確定結果を確認する', async ({ page }) => {
    await page.getByLabel('1A').check();
    const resultResponse = page.waitForResponse((response) =>
      response.url().endsWith('/api/admin/noon-game/matches/1/result')
        && response.request().method() === 'PUT'
    );
    await page.getByRole('button', { name: '結果を登録' }).click();
    await resultResponse;

    await expect(page.getByText('登録済み結果: 1A')).toBeVisible();
    await expect(page.getByText('ステータス: finished')).toBeVisible();
  });
});
