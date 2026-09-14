import { expect, test } from '@playwright/test';

const mockBackendUrl =
  process.env.MOCK_BACKEND_URL ??
  `http://127.0.0.1:${process.env.MOCK_BACKEND_PORT ?? 8081}`;

test.describe('試合結果入力 (admin)', () => {
  test.beforeEach(async ({ page, context, request }) => {
    await request.post(`${mockBackendUrl}/__reset`);
    await request.post(`${mockBackendUrl}/__set-user`, { data: { user: 'admin' } });
    await request.post(`${mockBackendUrl}/__set-active-event`, { data: { event_id: 1 } });
    await request.post(`${mockBackendUrl}/__set-match-pending`);
    await context.addCookies([{
      name: 'session_token',
      value: 'test-session-token',
      domain: 'localhost',
      path: '/'
    }]);

    const tournamentsResponse = page.waitForResponse((response) =>
      response.url().endsWith('/api/admin/events/1/tournaments')
    );
    await page.goto('/dashboard/admin/insert-matche-result');
    await expect(page.getByRole('heading', { name: '試合結果入力' })).toBeVisible();
    await tournamentsResponse;
  });

  test('対戦スコアを確認して結果を登録する', async ({ page }) => {
    await page.getByLabel('トーナメント選択').selectOption('1');
    await expect(page.getByText('1A vs 1B')).toBeVisible();
    await page.getByRole('button', { name: '結果を入力' }).click();

    const inputDialog = page.getByRole('dialog', { name: '結果入力: 1A vs 1B' });
    await inputDialog.getByLabel('1A Score').fill('5');
    await inputDialog.getByLabel('1B Score').fill('3');
    await inputDialog.getByRole('button', { name: '確認' }).click();

    await expect(page.getByRole('heading', { name: '試合結果確認' })).toBeVisible();
    await expect(page.getByText('勝者: 1A')).toBeVisible();

    page.once('dialog', async (dialog) => dialog.accept());
    const resultRequest = page.waitForRequest((request) => {
      if (!request.url().endsWith('/api/admin/matches/1/result') || request.method() !== 'PUT') {
        return false;
      }
      const body = JSON.parse(request.postData() ?? '{}');
      return body.team1_score === 5 && body.team2_score === 3;
    });
    await page.getByRole('button', { name: '登録' }).click();
    await resultRequest;

    await expect(page.getByText('勝者: 1A')).toBeVisible();
  });
});
