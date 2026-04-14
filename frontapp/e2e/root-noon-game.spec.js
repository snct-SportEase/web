import { expect, test } from '@playwright/test';

test.describe.configure({ mode: 'serial' });

test.describe('昼競技管理 (root)', () => {
  test.beforeEach(async ({ page, context }) => {
    await context.addCookies([{ name: 'session_token', value: 'test-session-token', domain: 'localhost', path: '/' }]);
    await page.goto('/dashboard/root/noon-game');
    await expect(page.getByRole('heading', { name: '昼競技管理' })).toBeVisible();
  });

  test('昼競技セッションを作成できる', async ({ page }) => {
    const requestPromise = page.waitForRequest((request) => request.url().endsWith('/api/root/events/1/noon-game/session') && request.method() === 'POST');
    await page.locator('input[placeholder="例: 昼休み競技 2025"]').fill('昼休み競技 2025');
    await page.getByRole('button', { name: 'セッションを作成' }).click();
    const req = await requestPromise;
    expect(JSON.parse(req.postData() ?? '{}')).toEqual(expect.objectContaining({
      name: '昼休み競技 2025',
      mode: 'mixed'
    }));
  });

  test('テンプレートを実行できる', async ({ page }) => {
    const openButton = page.getByRole('button', { name: 'テンプレートを設定' }).nth(1);
    await openButton.click();
    await expect(page.getByRole('heading', { name: /コース対抗リレー.*テンプレート設定/ })).toBeVisible();

    const requestPromise = page.waitForRequest((request) => request.url().endsWith('/api/root/events/1/noon-game/templates/course-relay/run') && request.method() === 'POST');
    await page.getByRole('button', { name: 'テンプレートを作成' }).click();
    await requestPromise;
  });
});
