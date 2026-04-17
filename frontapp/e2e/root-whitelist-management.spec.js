import { expect, test } from '@playwright/test';

const mockBackendUrl = process.env.MOCK_BACKEND_URL ?? 'http://127.0.0.1:8081';

test.describe.configure({ mode: 'serial' });

test.describe('ホワイトリスト管理 (root)', () => {
  test.beforeEach(async ({ page, context, request }) => {
    await request.post(`${mockBackendUrl}/__reset`);
    await context.addCookies([{ name: 'session_token', value: 'test-session-token', domain: 'localhost', path: '/' }]);
    await page.goto('/dashboard/root/whitelist-management');
    await expect(page.getByRole('heading', { name: 'ホワイトリスト管理' })).toBeVisible();
    await expect(page.getByRole('button', { name: 'Add' })).toBeEnabled();
  });

  test('ホワイトリスト一覧を表示できる', async ({ page }) => {
    await expect(page.getByText('student1@sendai-nct.jp')).toBeVisible();
    await expect(page.getByText('admin1@sendai-nct.jp')).toBeVisible();
  });

  test('メールアドレスを追加できる', async ({ page }) => {
    const requestPromise = page.waitForRequest((request) => request.url().includes('/api/root/whitelist') && request.method() === 'POST');
    await page.getByLabel('メールアドレス').evaluate((el, val) => {
      el.value = val;
      el.dispatchEvent(new Event('input', { bubbles: true }));
    }, 'new.user');
    await page.locator('#role').evaluate((el, val) => {
      el.value = val;
      el.dispatchEvent(new Event('change', { bubbles: true }));
    }, 'admin');
    await page.waitForTimeout(500); // 同期待ち
    const addButton = page.locator('button', { hasText: 'Add' });
    await expect(addButton).toBeEnabled();
    await addButton.click();
    const req = await requestPromise;
    expect(JSON.parse(req.postData() ?? '{}')).toEqual({
      email: 'new.user@sendai-nct.jp',
      role: 'admin'
    });
  });

  test('選択した項目を削除できる', async ({ page }) => {
    await page.evaluate(() => {
      window.confirm = () => true;
    });

    const requestPromise = page.waitForRequest((request) => request.url().includes('/api/root/whitelist/bulk') && request.method() === 'DELETE');
    const targetRow = page.locator('tbody tr').filter({ hasText: 'student1@sendai-nct.jp' });
    await expect(targetRow).toHaveCount(1);
    await targetRow.locator('input[type="checkbox"]').check();
    const deleteButton = page.getByRole('button', { name: '選択した項目を削除' });
    await expect(deleteButton).toBeEnabled();
    await deleteButton.click();
    const req = await requestPromise;
    expect(JSON.parse(req.postData() ?? '{}')).toEqual({
      emails: ['student1@sendai-nct.jp']
    });
  });
});
