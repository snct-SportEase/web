import { expect, test } from '@playwright/test';

const mockBackendUrl =
  process.env.MOCK_BACKEND_URL ??
  `http://127.0.0.1:${process.env.MOCK_BACKEND_PORT ?? 8081}`;

test.describe.configure({ mode: 'serial' });

test.describe('クラス競技割り当て (admin)', () => {
  test.beforeEach(async ({ page, context, request }) => {
    await request.post(`${mockBackendUrl}/__reset`);
    await request.post(`${mockBackendUrl}/__set-user`, { data: { user: 'admin' } });
    await request.post(`${mockBackendUrl}/__set-active-event`, { data: { event_id: 1 } });
    await request.post(`${mockBackendUrl}/api/admin/events/1/sports`, {
      data: { sport_id: 1, location: 'gym1', description: '屋内競技' }
    });
    await context.addCookies([{
      name: 'session_token',
      value: 'test-session-token',
      domain: 'localhost',
      path: '/'
    }]);

    await page.goto('/dashboard/admin/class-management');
    await expect(page.getByRole('heading', { name: 'クラス競技割り当て・管理' })).toBeVisible();
    await expect(page.getByText('student1@sendai-nct.jp')).toBeVisible();
  });

  test('クラスの学生を競技へ割り当て、その後削除できる', async ({ page }) => {
    await page.locator('select').nth(1).selectOption('1');
    await expect(page.getByText('メンバーが割り当てられていません')).toBeVisible();

    const memberRow = page.getByRole('row').filter({ hasText: 'student1@sendai-nct.jp' }).first();
    await memberRow.getByRole('checkbox').check();

    const assignRequest = page.waitForRequest((request) => {
      if (!request.url().endsWith('/api/admin/class-team/assign-members') || request.method() !== 'POST') {
        return false;
      }
      const body = JSON.parse(request.postData() ?? '{}');
      return body.class_id === 1 && body.sport_id === 1 && body.user_ids?.[0] === 'user-1';
    });
    await page.getByRole('button', { name: '選択した1名を割り当てる' }).click();
    await assignRequest;

    await expect(page.getByText('メンバーの割り当てが完了しました')).toBeVisible();
    const assignedSection = page.getByRole('heading', { name: '割り当て済みメンバー (バスケットボール)' }).locator('..');
    await expect(assignedSection.getByText('student1@sendai-nct.jp')).toBeVisible();

    page.once('dialog', async (dialog) => dialog.accept());
    const removeRequest = page.waitForRequest((request) => {
      if (!request.url().endsWith('/api/admin/class-team/remove-member') || request.method() !== 'DELETE') {
        return false;
      }
      const body = JSON.parse(request.postData() ?? '{}');
      return body.class_id === 1 && body.sport_id === 1 && body.user_id === 'user-1';
    });
    await assignedSection.getByRole('button', { name: '削除' }).click();
    await removeRequest;

    await expect(page.getByText('メンバーを削除しました')).toBeVisible();
    await expect(page.getByText('メンバーが割り当てられていません')).toBeVisible();
  });
});
