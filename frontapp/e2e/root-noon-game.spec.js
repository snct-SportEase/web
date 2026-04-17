import { expect, test } from '@playwright/test';

test.describe.configure({ mode: 'serial' });

test.describe('昼競技管理 (root)', () => {
  test.beforeEach(async ({ page, context }) => {
    page.on('console', (msg) => {
      console.log(`[Browser Console] ${msg.type()}: ${msg.text()}`);
    });

    await context.addCookies([{ name: 'session_token', value: 'test-session-token', domain: 'localhost', path: '/' }]);
    await page.goto('/dashboard/root/noon-game');
    await expect(page.getByRole('heading', { name: '昼競技管理' })).toBeVisible();
    // activeEvent.init() とクライアント側の初期化完了を待つ
    await expect(page.getByRole('button', { name: /セッションを(作成|更新)/ })).toBeEnabled({ timeout: 15000 });
  });

  test('昼競技セッションを作成できる', async ({ page }) => {
    page.once('dialog', (dialog) => {
      void dialog.accept().catch(() => {});
    });

    const requestPromise = page.waitForRequest((request) => request.url().includes('/noon-game/session') && request.method() === 'POST');
    const inputSelector = 'input[placeholder="例: 昼休み競技 2025"]';
    await page.locator(inputSelector).evaluate((el, val) => {
      el.value = val;
      el.dispatchEvent(new Event('input', { bubbles: true }));
      el.dispatchEvent(new Event('change', { bubbles: true }));
    }, '昼休み競技 2025');
    await page.locator(inputSelector).blur();
    await expect(page.getByRole('button', { name: 'セッションを作成' })).toBeEnabled();
    // 少し待機してイベントハンドラーのアタッチを確実にする
    await page.waitForTimeout(500);
    console.log('[E2E] Clicking button...');
    await page.locator('section', { hasText: 'テンプレートを使用しない場合の設定' }).getByRole('button', { name: 'セッションを作成' }).click();
    console.log('[E2E] Button clicked.');
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
