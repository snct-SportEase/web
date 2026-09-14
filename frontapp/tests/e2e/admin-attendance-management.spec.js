import { expect, test } from '@playwright/test';

const mockBackendUrl =
  process.env.MOCK_BACKEND_URL ??
  `http://127.0.0.1:${process.env.MOCK_BACKEND_PORT ?? 8081}`;

test.describe.configure({ mode: 'serial' });

test.describe('出席点管理 (admin)', () => {
  test.beforeEach(async ({ page, context, request }) => {
    await request.post(`${mockBackendUrl}/__reset`);
    await request.post(`${mockBackendUrl}/__set-user`, { data: { user: 'admin' } });
    await request.post(`${mockBackendUrl}/__set-active-event`, { data: { event_id: 1 } });
    await context.addCookies([{
      name: 'session_token',
      value: 'test-session-token',
      domain: 'localhost',
      path: '/'
    }]);

    await page.goto('/dashboard/admin/attendance-management');
    await expect(page.getByRole('heading', { name: '出席点管理' })).toBeVisible();
  });

  test('クラスの出席人数を登録して最新の出席点を表示する', async ({ page }) => {
    await expect(page.getByRole('heading', { name: '1A' })).toBeVisible();
    await expect(page.getByText('クラスの総人数: 40人')).toBeVisible();

    const registerRequest = page.waitForRequest((request) => {
      if (!request.url().endsWith('/api/admin/attendance/register') || request.method() !== 'POST') {
        return false;
      }
      const body = JSON.parse(request.postData() ?? '{}');
      return body.class_id === 1 && body.attendance_count === 36;
    });

    await page.getByLabel('出席人数').fill('36');
    await page.getByRole('button', { name: '出席を登録する' }).click();
    await registerRequest;

    await expect(page.getByText('Successfully registered attendance for class 1A. Points awarded: 36')).toBeVisible();
    await expect(page.getByText('現在の出席ポイント: 36ポイント')).toBeVisible();
  });
});
