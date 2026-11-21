import { expect, test } from '@playwright/test';

const mockBackendUrl =
  process.env.MOCK_BACKEND_URL ??
  `http://127.0.0.1:${process.env.MOCK_BACKEND_PORT ?? 8081}`;

test.describe('通知作成申請 (student)', () => {
  test.beforeEach(async ({ page, context, request }) => {
    await request.post(`${mockBackendUrl}/__reset`);
    await request.post(`${mockBackendUrl}/__set-user`, { data: { user: 'student' } });
    await context.addCookies([{
      name: 'session_token',
      value: 'test-session-token',
      domain: 'localhost',
      path: '/'
    }]);

    await page.goto('/dashboard/student/notification-request');
    await page.waitForLoadState('networkidle');
    await expect(page.getByRole('heading', { name: '通知申請' })).toBeVisible();
    await expect(page.getByText('お知らせ配信依頼', { exact: true }).first()).toBeVisible();
  });

  test('通知申請を作成しrootへ補足メッセージを送る', async ({ page }) => {
    await page.getByRole('button', { name: '申請フォームを開く' }).click();
    await page.getByLabel('タイトル').fill('決勝戦の招集通知');
    await page.getByLabel('対象ロール / 申請先').fill('決勝進出クラス');
    await page.getByLabel('内容').fill('決勝開始10分前の集合を通知してください。');

    const createRequest = page.waitForRequest((request) =>
      request.url().endsWith('/api/student/notification-requests') && request.method() === 'POST'
    );
    await page.getByRole('button', { name: '申請を送信' }).click();
    await createRequest;

    await expect(page).toHaveURL(/request_id=2/);
    await page.reload();
    await page.waitForLoadState('networkidle');
    await expect(page.getByText('決勝戦の招集通知', { exact: true }).first()).toBeVisible();
    await expect(page.getByText('決勝開始10分前の集合を通知してください。', { exact: true }).first()).toBeVisible();
    await expect(page.getByText('審査中').first()).toBeVisible();

    await page.getByLabel('メッセージを送信').fill('集合場所は第1体育館です。');
    const messageResponse = page.waitForResponse((response) =>
      response.url().endsWith('/api/student/notification-requests/2/messages')
        && response.request().method() === 'POST'
    );
    await page.getByRole('button', { name: 'メッセージを送信' }).click();
    await messageResponse;

    await expect(page.getByText('集合場所は第1体育館です。')).toBeVisible();
    await expect(page.getByText('山田太郎', { exact: true }).last()).toBeVisible();
  });
});
