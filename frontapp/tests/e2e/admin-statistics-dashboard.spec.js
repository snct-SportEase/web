import { expect, test } from '@playwright/test';

const mockBackendUrl =
  process.env.MOCK_BACKEND_URL ??
  `http://127.0.0.1:${process.env.MOCK_BACKEND_PORT ?? 8081}`;

async function openDashboard({ page, context, request }, hideScores) {
  await request.post(`${mockBackendUrl}/__reset`);
  await request.post(`${mockBackendUrl}/__set-user`, { data: { user: 'admin' } });
  await request.post(`${mockBackendUrl}/__set-active-event`, {
    data: { event_id: 1, hide_scores: hideScores }
  });
  await context.addCookies([{
    name: 'session_token',
    value: 'test-session-token',
    domain: 'localhost',
    path: '/'
  }]);
  await page.goto('/dashboard/admin/manage-dashboard');
  await expect(page.getByRole('heading', { name: '管理者ダッシュボード' })).toBeVisible();
}

test.describe('大会統計ダッシュボード (admin)', () => {
  test('出席率と競技進行状況を一覧できる', async ({ page, context, request }) => {
    await openDashboard({ page, context, request }, false);

    await expect(page.getByText('92.50%')).toBeVisible();
    await expect(page.getByText('決勝戦を実施中')).toBeVisible();
    await expect(page.getByText('準決勝まで完了')).toBeVisible();
    await expect(page.locator('#participationChart')).toBeVisible();
    await expect(page.locator('#scoreChart')).toBeVisible();
  });

  test('得点非表示中はスコア推移を描画しない', async ({ page, context, request }) => {
    await openDashboard({ page, context, request }, true);

    await expect(page.getByText('スコアは現在非表示に設定されています。')).toBeVisible();
    await expect(page.locator('#scoreChart')).toHaveCount(0);
    await expect(page.getByText('決勝戦を実施中')).toBeVisible();
  });
});
