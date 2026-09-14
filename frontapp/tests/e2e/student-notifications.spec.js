import { expect, test } from '@playwright/test';

const mockBackendUrl =
  process.env.MOCK_BACKEND_URL ??
  `http://127.0.0.1:${process.env.MOCK_BACKEND_PORT ?? 8081}`;

test.describe('通知確認・フィルタ設定 (student)', () => {
  test.beforeEach(async ({ page, context, request }) => {
    await request.post(`${mockBackendUrl}/__reset`);
    await request.post(`${mockBackendUrl}/__set-user`, { data: { user: 'student' } });
    await context.addCookies([{
      name: 'session_token',
      value: 'test-session-token',
      domain: 'localhost',
      path: '/'
    }]);

    await page.goto('/dashboard/student/notification');
    await expect(page.getByRole('heading', { name: '通知一覧' })).toBeVisible();
  });

  test('大会通知を確認し受信対象を保存する', async ({ page }) => {
    await expect(page.getByRole('heading', { name: '大会開催のお知らせ' })).toBeVisible();
    await expect(page.getByText('春季スポーツ大会を開催します。')).toBeVisible();
    await expect(page.getByText('学生', { exact: true })).toBeVisible();

    await expect(page.getByLabel('一般通知')).toBeChecked();
    await expect(page.getByLabel('一般通知')).toBeDisabled();
    await page.getByLabel('自分のクラスの試合').check();
    await page.getByLabel('決勝戦').check();
    await page.getByRole('button', { name: '設定を保存' }).click();

    await expect(page.getByText('通知フィルタを更新しました')).toBeVisible();
    await page.reload();
    await expect(page.getByLabel('自分のクラスの試合')).toBeChecked();
    await expect(page.getByLabel('決勝戦')).toBeChecked();
    await expect(page.getByLabel('全ての試合')).not.toBeChecked();
  });
});
