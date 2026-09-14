import { expect, test } from '@playwright/test';

const mockBackendUrl =
  process.env.MOCK_BACKEND_URL ??
  `http://127.0.0.1:${process.env.MOCK_BACKEND_PORT ?? 8081}`;

test.describe('運営ロール管理 (admin)', () => {
  test.beforeEach(async ({ page, context, request }) => {
    await request.post(`${mockBackendUrl}/__reset`);
    await request.post(`${mockBackendUrl}/__set-user`, { data: { user: 'admin' } });
    await context.addCookies([{
      name: 'session_token',
      value: 'test-session-token',
      domain: 'localhost',
      path: '/'
    }]);

    page.on('dialog', (dialog) => void dialog.accept());
    await page.goto('/dashboard/admin/role-management');
    await page.waitForLoadState('networkidle');
    await expect(page.getByRole('heading', { name: 'ロール管理' })).toBeVisible();
  });

  test('ユーザーを検索して大会運営ロールを割り当てる', async ({ page }) => {
    const search = page.getByLabel('ユーザー検索');
    await search.focus();
    await search.fill('admin1');
    await page.getByRole('button', { name: 'ユーザー admin1@sendai-nct.jp を選択' }).click();
    await page.getByLabel('ロール名').fill('scorekeeper');

    const updateResponse = page.waitForResponse((response) =>
      response.url().endsWith('/api/admin/users/role')
        && response.request().method() === 'PUT'
    );
    await page.getByRole('button', { name: '割り当て' }).click();
    await updateResponse;

    await expect(page.getByRole('heading', { name: 'ロール管理' })).toBeVisible();
    const assignedRow = page.locator('tr', { hasText: 'admin1@sendai-nct.jp' });
    await expect(assignedRow).toContainText('scorekeeper');
  });
});
