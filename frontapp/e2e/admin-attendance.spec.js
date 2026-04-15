import { test, expect } from '@playwright/test';

const mockBackendUrl = process.env.MOCK_BACKEND_URL ?? 'http://127.0.0.1:8081';

test.describe('出席者登録 (admin/root)', () => {
  test.beforeEach(async ({ page, context, request }) => {
    await request.post(`${mockBackendUrl}/__reset`);
    await context.addCookies([{ name: 'session_token', value: 'test-session-token', domain: 'localhost', path: '/' }]);
    await page.goto('/dashboard/admin/attendance-management');
    await expect(page.getByRole('heading', { name: '出席点管理' })).toBeVisible();
  });

  test('クラスを選択して出席人数を登録できる', async ({ page }) => {
    // 1Aを選択 (1Aはstudent_count: 40)
    await page.getByLabel('対象クラスを選択').selectOption('1');
    
    // 見出しが表示されるのを待機
    await expect(page.getByRole('heading', { name: '1A', exact: true })).toBeVisible();
    await expect(page.getByText('クラスの総人数: 40人')).toBeVisible();

    await page.getByLabel('出席人数').fill('38');
    
    const registerRequest = page.waitForRequest((request) => 
      request.url().endsWith('/api/admin/attendance/register') && request.method() === 'POST'
    );

    await page.getByRole('button', { name: '出席を登録する' }).click();

    const request = await registerRequest;
    expect(JSON.parse(request.postData() ?? '{}')).toEqual({
      class_id: 1,
      attendance_count: 38
    });

    await expect(page.getByText(/出席を正常に登録しました/)).toBeVisible();
  });

  test('総人数を超える人数は登録できない', async ({ page }) => {
    await page.getByLabel('対象クラスを選択').selectOption('1');
    await expect(page.getByRole('heading', { name: '1A', exact: true })).toBeVisible();

    await page.getByLabel('出席人数').fill('41');
    await page.getByRole('button', { name: '出席を登録する' }).click();

    await expect(page.getByText(/出席人数がクラスの総人数.*超えています/)).toBeVisible();
  });
});
