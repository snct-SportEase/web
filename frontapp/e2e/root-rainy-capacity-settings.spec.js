import { test, expect } from '@playwright/test';

const mockBackendUrl = process.env.MOCK_BACKEND_URL ?? 'http://127.0.0.1:8081';

test.describe('雨天時定員設定 (root)', () => {
  test.beforeEach(async ({ page, context, request }) => {
    await request.post(`${mockBackendUrl}/__reset`);
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

    await page.getByLabel('競技選択').selectOption('1');
    await expect(page.getByText('現在の設定: 定員 未設定 〜 未設定')).toBeVisible();
    await expect(page.locator('#rainy-min-capacity')).toBeVisible();

    await page.locator('#rainy-min-capacity').evaluate((input) => {
      input.value = '6';
      input.dispatchEvent(new Event('input', { bubbles: true }));
      input.dispatchEvent(new Event('change', { bubbles: true }));
    });
    await page.locator('#rainy-max-capacity').evaluate((input) => {
      input.value = '8';
      input.dispatchEvent(new Event('input', { bubbles: true }));
      input.dispatchEvent(new Event('change', { bubbles: true }));
    });

    const [dialog] = await Promise.all([
      page.waitForEvent('dialog'),
      page.getByRole('button', { name: '雨天時定員設定を保存' }).click()
    ]);
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
