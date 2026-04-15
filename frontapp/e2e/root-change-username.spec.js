import { expect, test } from '@playwright/test';

const mockBackendUrl = process.env.MOCK_BACKEND_URL ?? 'http://127.0.0.1:8081';

test.describe.configure({ mode: 'serial' });

test.describe('ユーザー管理 (root)', () => {
  test.beforeEach(async ({ page, context, request }) => {
    await request.post(`${mockBackendUrl}/__reset`);
    await context.addCookies([{ name: 'session_token', value: 'test-session-token', domain: 'localhost', path: '/' }]);
    await page.goto('/dashboard/root/change-username');
    await expect(page.getByRole('heading', { name: 'ユーザー管理' })).toBeVisible();
  });

  test('ユーザー一覧を表示できる', async ({ page }) => {
    await expect(page.getByText('student1@sendai-nct.jp')).toBeVisible();
  });

  test('表示名を更新できる', async ({ page }) => {
    await page.locator('tbody button').first().click({ force: true });
    await expect(page.locator('#displayNameInput')).toBeVisible();
    const requestPromise = page.waitForRequest((request) => request.url().endsWith('/api/root/users/display-name') && request.method() === 'PUT');
    await page.locator('#displayNameInput').fill('新しい表示名');
    
    await Promise.all([
      page.waitForEvent('dialog').then(dialog => dialog.accept()),
      page.getByRole('button', { name: '更新' }).click()
    ]);

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
    await page.locator('#newRoleInput').fill('admin');
    
    await Promise.all([
      page.waitForEvent('dialog').then(dialog => dialog.accept()),
      page.getByRole('button', { name: '追加' }).click()
    ]);

    const req = await requestPromise;
    expect(JSON.parse(req.postData() ?? '{}')).toEqual({
      user_id: 'user-1',
      role: 'admin'
    });
  });

  test('rootロールを追加できる', async ({ page }) => {
    await page.locator('tbody tr', { has: page.getByText('student1@sendai-nct.jp') }).getByRole('button', { name: '管理' }).click({ force: true });
    await expect(page.locator('#newRoleInput')).toBeVisible();
    const requestPromise = page.waitForRequest((request) => request.url().endsWith('/api/admin/users/role') && request.method() === 'PUT');
    await page.locator('#newRoleInput').fill('root');
    
    await Promise.all([
      page.waitForEvent('dialog').then(dialog => dialog.accept()),
      page.getByRole('button', { name: '追加' }).click()
    ]);

    const req = await requestPromise;
    expect(JSON.parse(req.postData() ?? '{}')).toEqual({
      user_id: 'user-1',
      role: 'root'
    });
    await expect(page.getByText('root')).toBeVisible();
  });

  test('ロールを削除できる', async ({ page }) => {
    await page.locator('tbody button').first().click({ force: true });
    await expect(page.locator('button[title="ロールを削除"]').first()).toBeVisible();
    const requestPromise = page.waitForRequest((request) => request.url().endsWith('/api/admin/users/role') && request.method() === 'DELETE');
    
    // Deletion has two dialogs: 1. confirm, 2. alert
    const dialogs = [];
    page.on('dialog', async (dialog) => {
      dialogs.push(dialog.message());
      await dialog.accept();
    });

    await page.locator('button[title="ロールを削除"]').first().click();

    const req = await requestPromise;
    expect(JSON.parse(req.postData() ?? '{}')).toEqual({
      user_id: 'user-1',
      role: 'student'
    });
    
    // Wait for the final alert to be processed (implicitly handled by the listener)
    await expect.poll(() => dialogs.length).toBe(2);
  });

  test('adminロールを削除できる', async ({ page }) => {
    await page.locator('tbody tr', { has: page.getByText('admin1@sendai-nct.jp') }).getByRole('button', { name: '管理' }).click({ force: true });
    await expect(page.locator('button[title="ロールを削除"]').first()).toBeVisible();
    const requestPromise = page.waitForRequest((request) => request.url().endsWith('/api/admin/users/role') && request.method() === 'DELETE');

    const dialogs = [];
    page.on('dialog', async (dialog) => {
      dialogs.push(dialog.message());
      await dialog.accept();
    });

    await page.locator('button[title="ロールを削除"]').first().click();

    const req = await requestPromise;
    expect(JSON.parse(req.postData() ?? '{}')).toEqual({
      user_id: 'user-2',
      role: 'admin'
    });
    await expect(page.getByText('ロールなし')).toBeVisible();
    await expect.poll(() => dialogs.length).toBe(2);
  });

  test('クラス所属ロールを付け替えできる', async ({ page }) => {
    await page.locator('tbody button').first().click({ force: true });
    await expect(page.locator('#classRepSelect')).toBeVisible();
    await page.locator('#classRepSelect').selectOption('2');
    const addRequestPromise = page.waitForRequest((request) => request.url().endsWith('/api/admin/users/role') && request.method() === 'PUT');
    
    await Promise.all([
      page.waitForEvent('dialog').then(dialog => dialog.accept()),
      page.getByRole('button', { name: '変更・保存' }).click()
    ]);

    const req = await addRequestPromise;
    expect(JSON.parse(req.postData() ?? '{}')).toEqual({
      user_id: 'user-1',
      role: '1B_rep'
    });
  });
});
