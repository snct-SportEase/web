import { expect, test } from '@playwright/test';

test.describe.configure({ mode: 'serial' });

test.describe('通知申請管理 (root)', () => {
  test.beforeEach(async ({ page, context }) => {
    await context.addCookies([{ name: 'session_token', value: 'test-session-token', domain: 'localhost', path: '/' }]);
    await page.goto('/dashboard/root/notification-requests');
    await expect(page.getByRole('heading', { name: '通知申請管理' })).toBeVisible();
  });

  test('申請一覧と詳細を表示できる', async ({ page }) => {
    await expect(page.getByText('お知らせ配信依頼').nth(0)).toBeVisible();
    await expect(page.getByText('内容を確認お願いします。')).toBeVisible();
  });

  test('メッセージを送信できる', async ({ page }) => {
    const requestPromise = page.waitForRequest((request) => request.url().endsWith('/api/root/notification-requests/1/messages') && request.method() === 'POST');
    await page.getByRole('textbox', { name: 'メッセージを送信' }).pressSequentially('了解しました。');
    await page.getByRole('button', { name: 'メッセージを送信' }).click();
    const req = await requestPromise;
    expect(JSON.parse(req.postData() ?? '{}')).toEqual({ message: '了解しました。' });
    await expect(page.getByText('了解しました。')).toBeVisible();
  });

  test('通知申請を承認できる', async ({ page }) => {
    const requestPromise = page.waitForRequest((request) => request.url().endsWith('/api/root/notification-requests/1/decision') && request.method() === 'POST');
    await page.getByRole('button', { name: '承認する' }).click();
    const req = await requestPromise;
    expect(JSON.parse(req.postData() ?? '{}')).toEqual({ status: 'approved' });
    await expect(page.getByText('承認済み')).toBeVisible();
  });
});
