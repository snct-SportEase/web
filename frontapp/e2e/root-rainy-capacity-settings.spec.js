import { test, expect } from '@playwright/test';

const mockBackendUrl = process.env.MOCK_BACKEND_URL ?? 'http://127.0.0.1:8081';

test.describe('雨天時定員設定 (root)', () => {
  test.beforeEach(async ({ page, context, request }) => {
    await request.post(`${mockBackendUrl}/__reset`);
    await request.post(`${mockBackendUrl}/api/admin/events/1/sports`, {
      data: {
        sport_id: 1,
        location: 'gym1',
        description: 'バスケットボール',
        rules: 'ルール'
      }
    });
    await context.addCookies([{ name: 'session_token', value: 'test-session-token', domain: 'localhost', path: '/' }]);
    await page.goto('/dashboard/admin/sport-details-registration');
  });

  test('競技詳細登録画面から雨天時定員を一括保存できる', async ({ page }) => {
    const saveRequests = [];

    page.on('request', (request) => {
      if (
        request.url().endsWith('/api/root/events/1/rainy-mode/settings') &&
        request.method() === 'POST'
      ) {
        saveRequests.push(JSON.parse(request.postData() ?? '{}'));
      }
    });

    await expect(page.getByRole('heading', { name: '競技詳細情報登録' })).toBeVisible();
    await expect(page.getByText('2025春季スポーツ大会')).toBeVisible();

    await Promise.all([
      page.waitForResponse(res => res.url().includes('/api/admin/events/1/sports/1/details')),
      page.waitForResponse(res => res.url().includes('/api/root/events/1/rainy-mode/settings')),
      page.getByLabel('競技選択').selectOption('1')
    ]);

    await expect(page.getByText('現在の設定: 定員 未設定 〜 未設定')).toBeVisible();
    await expect(page.locator('#rainy-min-capacity')).toBeVisible();

    await page.locator('#rainy-min-capacity').fill('6');
    await page.locator('#rainy-max-capacity').fill('8');
    await expect(page.getByText('現在の設定: 定員 6 〜 8')).toBeVisible();

    const dialogPromise = page.waitForEvent('dialog');

    await page.getByRole('button', { name: '雨天時定員設定を保存' }).click();
    const dialog = await dialogPromise;
    expect(dialog.message()).toBe('雨天時定員設定を更新しました。');
    await dialog.accept();

    await expect.poll(() => saveRequests.length).toBe(2);

    const requestBodies = [...saveRequests].sort((left, right) => Number(left.class_id) - Number(right.class_id));
    expect(requestBodies).toEqual([
      { sport_id: '1', class_id: 1, min_capacity: 6, max_capacity: 8 },
      { sport_id: '1', class_id: 2, min_capacity: 6, max_capacity: 8 }
    ]);
  });
});
