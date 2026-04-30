import { expect, test } from '@playwright/test';

test.describe.configure({ mode: 'serial' });

const mockBackendUrl = process.env.MOCK_BACKEND_URL ?? 'http://127.0.0.1:8081';

test.describe('ユーザー管理 (root)', () => {
  test.beforeEach(async ({ page, context, request }) => {
    await request.post(`${mockBackendUrl}/__reset`);
    await context.addCookies([{ name: 'session_token', value: 'test-session-token', domain: 'localhost', path: '/' }]);
    await page.goto('/dashboard/root/change-username');
    await expect(page.getByRole('heading', { name: 'ユーザー管理' })).toBeVisible();
    await page.evaluate(() => {
      window.alert = () => {};
      window.confirm = () => true;
    });
  });

  test('ユーザー一覧を表示できる', async ({ page }) => {
    await expect(page.getByText('student1@sendai-nct.jp')).toBeVisible();
  });

  test('表示名を更新できる', async ({ page }) => {
    await page.locator('tbody button').first().click({ force: true });
    await expect(page.locator('#displayNameInput')).toBeVisible();
    const requestPromise = page.waitForRequest((request) => request.url().endsWith('/api/root/users/display-name') && request.method() === 'PUT');
    await page.locator('#displayNameInput').fill('新しい表示名');
    await page.getByRole('button', { name: '更新' }).click();
    const req = await requestPromise;
    expect(JSON.parse(req.postData() ?? '{}')).toEqual({
      user_id: 'user-1',
      display_name: '新しい表示名'
    });
  });

  test('ロールを追加できる', async ({ page }) => {
    await page.locator('tbody button').first().click({ force: true });
    await expect(page.locator('#newRoleInput')).toBeVisible();
    const requestPromise = page.waitForRequest((request) => request.url().endsWith('/api/admin/users/role') && request.method() === 'PUT');
    await page.locator('#newRoleInput').fill('score_keeper');
    await page.getByRole('button', { name: '追加' }).click();
    const req = await requestPromise;
    expect(JSON.parse(req.postData() ?? '{}')).toEqual({
      user_id: 'user-1',
      role: 'score_keeper'
    });
  });

  test('rootロールに切り替えできる', async ({ page }) => {
    const targetRow = page.locator('tbody tr', { has: page.getByText('student1@sendai-nct.jp') });
    await targetRow.getByRole('button', { name: '管理' }).click({ force: true });
    await expect(page.getByRole('button', { name: 'root に切り替え' })).toBeVisible();
    const requestPromise = page.waitForRequest((request) => request.url().endsWith('/api/root/users/promote') && request.method() === 'PUT');
    await page.getByRole('button', { name: 'root に切り替え' }).click();
    const req = await requestPromise;
    expect(JSON.parse(req.postData() ?? '{}')).toEqual({
      user_id: 'user-1',
      role: 'root'
    });
    await expect(page.locator('[aria-labelledby="modal-title"]').getByText('root', { exact: true })).toBeVisible();
  });

  test('ロールを削除できる', async ({ page }) => {
    await page.locator('tbody button').first().click({ force: true });
    await expect(page.locator('button[title="ロールを削除"]').first()).toBeVisible();
    const requestPromise = page.waitForRequest((request) => request.url().endsWith('/api/admin/users/role') && request.method() === 'DELETE');
    await page.locator('button[title="ロールを削除"]').first().click();
    const req = await requestPromise;
    expect(JSON.parse(req.postData() ?? '{}')).toEqual({
      user_id: 'user-1',
      role: 'judge'
    });
  });

  test('masterロールはその他のロール削除対象に出ない', async ({ page }) => {
    const targetRow = page.locator('tbody tr', { has: page.getByText('admin1@sendai-nct.jp') });
    await targetRow.getByRole('button', { name: '管理' }).click({ force: true });
    await expect(page.getByText('admin を保有中')).toBeVisible();
    await expect(page.locator('button[title="ロールを削除"]')).toHaveCount(0);
  });

  test('クラス所属ロールを付け替えできる', async ({ page }) => {
    await page.locator('tbody button').first().click({ force: true });
    await expect(page.locator('#classRepSelect')).toBeVisible();
    await page.locator('#classRepSelect').selectOption('2');
    const addRequestPromise = page.waitForRequest((request) => request.url().endsWith('/api/admin/users/role') && request.method() === 'PUT');
    await page.getByRole('button', { name: '変更・保存' }).click();
    const req = await addRequestPromise;
    expect(JSON.parse(req.postData() ?? '{}')).toEqual({
      user_id: 'user-1',
      role: '1B_rep'
    });
  });
});
