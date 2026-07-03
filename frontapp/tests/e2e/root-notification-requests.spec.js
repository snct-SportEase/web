import { expect, test } from '@playwright/test';

test.describe.configure({ mode: 'serial' });

test.describe('通知申請管理 (root)', () => {
  test.beforeEach(async ({ page, context }) => {
    await context.addCookies([{ name: 'session_token', value: 'test-session-token', domain: 'localhost', path: '/' }]);
    await page.goto('/dashboard/root/notification-requests');
    await expect(page.getByRole('heading', { name: '通知申請管理' })).toBeVisible();
    await expect(page.getByRole('button', { name: 'メッセージを送信' })).toBeEnabled();
  });

  test('申請一覧と詳細を表示できる', async ({ page }) => {
    await expect(page.getByText('お知らせ配信依頼').nth(0)).toBeVisible();
    await expect(page.getByText('内容を確認お願いします。')).toBeVisible();
  });

  test('メッセージを送信できる', async ({ page }) => {
    const requestPromise = page.waitForRequest((request) => request.url().includes('/messages') && request.method() === 'POST');
    await page.getByRole('textbox', { name: 'メッセージを送信' }).fill('了解しました。');
    await page.getByRole('textbox', { name: 'メッセージを送信' }).blur();
    await page.waitForTimeout(500); // 状態更新待ち
    // メッセージ送信ボタンを特定（textareaのすぐ下にあるボタン）
    const sendButton = page.locator('form:has(textarea#rootMessageInput)').getByRole('button', { name: 'メッセージを送信' });
    await expect(sendButton).toBeEnabled();
    await sendButton.click();
    const req = await requestPromise;
    expect(JSON.parse(req.postData() ?? '{}')).toEqual({ message: '了解しました。' });
    await expect(page.getByText('了解しました。')).toBeVisible({ timeout: 15000 });
  });

  test('通知申請を承認できる', async ({ page }) => {
    const requestPromise = page.waitForRequest((request) => request.url().includes('/decision') && request.method() === 'POST');
    // 承認ボタンを特定
    const approveButton = page.locator('div.flex:has(button:text("承認する"))').getByRole('button', { name: '承認する' });
    await expect(approveButton).toBeEnabled();
    await approveButton.click();
    const req = await requestPromise;
    expect(JSON.parse(req.postData() ?? '{}')).toEqual({ status: 'approved' });
    await expect(page.getByText('承認済み')).toBeVisible();
  });
});
