import { expect, test } from '@playwright/test';

const mockBackendUrl =
  process.env.MOCK_BACKEND_URL ??
  `http://127.0.0.1:${process.env.MOCK_BACKEND_PORT ?? 8081}`;

test.describe.configure({ mode: 'serial' });

test.describe('マスタロール管理 (root)', () => {
  test.beforeEach(async ({ page, context, request }) => {
    await request.post(`${mockBackendUrl}/__reset`);
    await context.addCookies([{
      name: 'session_token',
      value: 'test-session-token',
      domain: 'localhost',
      path: '/'
    }]);
    await page.goto('/dashboard/root/user-promotion');
    await expect(page.getByRole('heading', { name: '権限管理' })).toBeVisible();
    await expect(page.getByText('student1@sendai-nct.jp')).toBeVisible();
  });

  test('ユーザーを検索してstudentからadminへ切り替える', async ({ page }) => {
    await page.getByPlaceholder('検索キーワードを入力...').fill('student1@sendai-nct.jp');
    await page.getByRole('button', { name: '検索' }).click();

    const row = page.getByRole('row').filter({ hasText: 'student1@sendai-nct.jp' });
    await expect(row).toContainText('student');
    await expect(page.getByText('admin1@sendai-nct.jp')).not.toBeVisible();

    page.once('dialog', async (dialog) => dialog.accept());
    const promoteRequest = page.waitForRequest((request) => {
      if (!request.url().endsWith('/api/root/users/promote') || request.method() !== 'PUT') {
        return false;
      }
      const body = JSON.parse(request.postData() ?? '{}');
      return body.user_id === 'user-1' && body.role === 'admin';
    });

    await row.getByRole('button', { name: '切替' }).first().click();
    await promoteRequest;

    await expect(row).toContainText('admin');
    await expect(row.getByRole('button', { name: '保有中' })).toBeDisabled();
  });

  test('表示名検索と全件表示を切り替えられる', async ({ page }) => {
    await page.getByLabel('表示名').check();
    await page.getByPlaceholder('検索キーワードを入力...').fill('運営花子');
    await page.getByRole('button', { name: '検索' }).click();

    await expect(page.getByText('admin1@sendai-nct.jp')).toBeVisible();
    await expect(page.getByText('student1@sendai-nct.jp')).not.toBeVisible();

    await page.getByRole('button', { name: 'すべて表示' }).click();
    await expect(page.getByText('student1@sendai-nct.jp')).toBeVisible();
  });
});
