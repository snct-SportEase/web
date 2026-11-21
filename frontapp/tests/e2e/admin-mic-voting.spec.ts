import { expect, test } from '@playwright/test';

const mockBackendUrl =
  process.env.MOCK_BACKEND_URL ??
  `http://127.0.0.1:${process.env.MOCK_BACKEND_PORT ?? 8081}`;

test.describe('行事委員会賞投票 (admin)', () => {
  test.beforeEach(async ({ page, context, request }) => {
    await request.post(`${mockBackendUrl}/__reset`);
    await request.post(`${mockBackendUrl}/__set-user`, { data: { user: 'admin' } });
    await request.post(`${mockBackendUrl}/__set-active-event`, { data: { event_id: 1 } });
    await context.addCookies([{
      name: 'session_token',
      value: 'test-session-token',
      domain: 'localhost',
      path: '/'
    }]);

    await page.goto('/dashboard/admin/vorting-mic');
    await expect(page.getByRole('heading', { name: '行事委員会賞投票' })).toBeVisible();
    await expect(page.getByText('投票状態: 有効')).toBeVisible();
  });

  test('投票理由と対象クラスを送信し再投票を防止する', async ({ page }) => {
    await page.getByLabel('投票対象クラス').selectOption('1');
    await page.getByLabel('理由').fill('大会を通して応援と運営協力が素晴らしかったため');

    page.once('dialog', async (dialog) => dialog.accept());
    const voteRequest = page.waitForRequest((request) => {
      if (!request.url().endsWith('/api/admin/mic/vote') || request.method() !== 'POST') {
        return false;
      }
      const body = JSON.parse(request.postData() ?? '{}');
      return body.event_id === 1
        && body.voted_for_class_id === 1
        && body.reason.includes('応援と運営協力');
    });
    await page.getByRole('button', { name: '投票する' }).click();
    await voteRequest;

    await expect(page.getByRole('heading', { name: '投票済みです' })).toBeVisible();
    await expect(page.getByText('あなたは 1A に投票しました。')).toBeVisible();
    await expect(page.getByRole('button', { name: '投票する' })).not.toBeVisible();
  });
});
