import { test, expect } from '@playwright/test';

const mockBackendUrl = process.env.MOCK_BACKEND_URL ?? 'http://127.0.0.1:8081';

test.describe('各クラス人数設定 (root)', () => {
  test.beforeEach(async ({ page, context, request }) => {
    await request.post(`${mockBackendUrl}/__reset`);
    await context.addCookies([{ name: 'session_token', value: 'test-session-token', domain: 'localhost', path: '/' }]);
    await page.goto('/dashboard/root/class-student-count');
    await expect(page.getByText('1A')).toBeVisible();
  });

  test('クラス一覧を表示できる', async ({ page }) => {
    await expect(page.getByRole('heading', { name: '各クラス人数設定' })).toBeVisible();
    await expect(page.getByText('1B')).toBeVisible();
  });

  test('手動で生徒数を更新できる', async ({ page }) => {
    const req = page.waitForRequest((request) => request.url().endsWith('/api/root/classes/student-counts') && request.method() === 'PUT');
    await page.getByRole('spinbutton').first().fill('42');
    await page.getByRole('button', { name: '保存' }).click();
    const request = await req;
    expect(JSON.parse(request.postData() ?? '{}')[0]).toEqual({ class_id: 1, student_count: 42 });
    await expect(page.getByText('生徒数を更新しました。')).toBeVisible();
  });
});
