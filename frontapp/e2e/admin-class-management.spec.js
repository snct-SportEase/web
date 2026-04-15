import { expect, test } from '@playwright/test';

const mockBackendUrl = process.env.MOCK_BACKEND_URL ?? 'http://127.0.0.1:8081';

test.describe.configure({ mode: 'serial' });

test.describe('チームメンバー割り当て (admin/root)', () => {
  test.beforeEach(async ({ page, context, request }) => {
    await request.post(`${mockBackendUrl}/__reset`);
    await request.post(`${mockBackendUrl}/api/admin/events/1/sports`, {
      data: {
        sport_id: 1,
        location: 'gym1',
        description: '屋内メイン競技',
        rules: '# バスケットボール',
        rules_type: 'markdown'
      }
    });
    await context.addCookies([{ name: 'session_token', value: 'test-session-token', domain: 'localhost', path: '/' }]);
    await page.goto('/dashboard/admin/class-management');
    await page.waitForLoadState('networkidle');
    await expect(page.getByRole('heading', { name: 'クラス・チーム管理' })).toBeVisible();
  });

  test('競技参加メンバーを登録できる', async ({ page }) => {
    await expect(page.getByText('山田太郎')).toBeVisible();
    const memberRow = page.locator('tr', { has: page.getByText('山田太郎') });
    const memberCheckbox = memberRow.locator('input[type="checkbox"]');

    const assignRequest = page.waitForRequest((request) => {
      return request.url().endsWith('/api/admin/class-team/assign-members') && request.method() === 'POST';
    });

    await page.getByRole('combobox').nth(1).selectOption('1');
    await memberCheckbox.evaluate((element) => element.click());
    await expect(memberCheckbox).toBeChecked();
    await expect(page.getByRole('button', { name: '選択した1名を割り当てる' })).toBeVisible();
    await page.getByRole('button', { name: '選択した1名を割り当てる' }).click();

    const request = await assignRequest;
    expect(JSON.parse(request.postData() ?? '{}')).toEqual({
      sport_id: 1,
      class_id: 1,
      user_ids: ['class-user-1']
    });

    await expect(page.getByText('メンバーの割り当てが完了しました')).toBeVisible();
    await expect(page.getByRole('heading', { name: '割り当て済みメンバー (バスケットボール)' })).toBeVisible();
    await expect(page.getByRole('cell', { name: '山田太郎' }).nth(1)).toBeVisible();
  });

  test('割り当て済みメンバーを削除できる', async ({ page }) => {
    const memberRow = page.locator('tr', { has: page.getByText('山田太郎') });
    const memberCheckbox = memberRow.locator('input[type="checkbox"]');
    const assignRequest = page.waitForRequest((request) => {
      return request.url().endsWith('/api/admin/class-team/assign-members') && request.method() === 'POST';
    });

    await page.getByRole('combobox').nth(1).selectOption('1');
    await memberCheckbox.evaluate((element) => element.click());
    await expect(memberCheckbox).toBeChecked();
    await expect(page.getByRole('button', { name: '選択した1名を割り当てる' })).toBeVisible();
    await page.getByRole('button', { name: '選択した1名を割り当てる' }).click();
    await assignRequest;

    const removeRequest = page.waitForRequest((request) => {
      return request.url().endsWith('/api/admin/class-team/remove-member') && request.method() === 'DELETE';
    });

    page.once('dialog', async (dialog) => {
      await dialog.accept();
    });

    await page.getByRole('button', { name: '削除' }).click();
    const request = await removeRequest;
    expect(JSON.parse(request.postData() ?? '{}')).toEqual({
      sport_id: 1,
      class_id: 1,
      user_id: 'class-user-1'
    });

    await expect(page.getByText('メンバーを削除しました')).toBeVisible();
    await expect(page.getByText('メンバーが割り当てられていません')).toBeVisible();
  });
});
