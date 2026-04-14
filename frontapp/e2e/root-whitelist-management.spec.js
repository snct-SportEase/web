import { expect, test } from '@playwright/test';

test.describe.configure({ mode: 'serial' });

test.describe('ホワイトリスト管理 (root)', () => {
  test.beforeEach(async ({ page, context }) => {
    await context.addCookies([{ name: 'session_token', value: 'test-session-token', domain: 'localhost', path: '/' }]);
    await page.goto('/dashboard/root/whitelist-management');
    await expect(page.getByRole('heading', { name: 'ホワイトリスト管理' })).toBeVisible();
  });

  test('ホワイトリスト一覧を表示できる', async ({ page }) => {
    await expect(page.getByText('student1@sendai-nct.jp')).toBeVisible();
    await expect(page.getByText('admin1@sendai-nct.jp')).toBeVisible();
  });

  test('メールアドレスを追加できる', async ({ page }) => {
    const requestPromise = page.waitForRequest((request) => request.url().endsWith('/api/root/whitelist') && request.method() === 'POST');
    await page.getByLabel('メールアドレス').pressSequentially('new.user');
    await page.locator('#role').selectOption('admin');
    await page.getByRole('button', { name: 'Add' }).click();
    const req = await requestPromise;
    expect(JSON.parse(req.postData() ?? '{}')).toEqual({
      email: 'new.user@sendai-nct.jp',
      role: 'admin'
    });
  });

  test('選択した項目を削除できる', async ({ page }) => {
    page.once('dialog', async (dialog) => {
      await dialog.accept();
    });

    const requestPromise = page.waitForRequest((request) => request.url().endsWith('/api/root/whitelist/bulk') && request.method() === 'DELETE');
    await page.getByRole('checkbox').nth(1).check();
    await page.getByRole('button', { name: '選択した項目を削除' }).click();
    const req = await requestPromise;
    expect(JSON.parse(req.postData() ?? '{}')).toEqual({
      emails: ['admin1@sendai-nct.jp']
    });
  });
});
