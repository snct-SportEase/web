import { test, expect } from '@playwright/test';

test.describe.configure({ mode: 'serial' });

test.describe('競技情報登録・管理 (root)', () => {
  test.beforeEach(async ({ page, context, request }) => {
    await request.post('http://127.0.0.1:8081/__reset');

    await context.addCookies([{
      name: 'session_token',
      value: 'test-session-token',
      domain: 'localhost',
      path: '/'
    }]);

    await page.goto('/dashboard/root/sport-management');
    await expect(page.getByRole('list').getByText('バスケットボール')).toBeVisible();
  });

  test('競技マスタ一覧を表示できる', async ({ page }) => {
    await expect(page.getByRole('heading', { name: '大会競技管理ダッシュボード' })).toBeVisible();
    await expect(page.getByText('2025春季スポーツ大会')).toBeVisible();
    await expect(page.getByRole('list').getByText('バレーボール')).toBeVisible();
  });

  test('新しい競技をマスタ登録できる', async ({ page }) => {
    await page.getByPlaceholder('新しい競技名を入力').fill('綱引き');

    const createRequest = page.waitForRequest((request) => {
      if (request.url().endsWith('/api/root/sports') && request.method() === 'POST') {
        const body = JSON.parse(request.postData() ?? '{}');
        return body.name === '綱引き';
      }

      return false;
    });

    page.once('dialog', async (dialog) => {
      await dialog.accept();
    });

    await page.getByRole('button', { name: '競技をマスタに登録' }).click();

    const request = await createRequest;
    const body = JSON.parse(request.postData() ?? '{}');
    expect(body).toEqual({ name: '綱引き' });

    await expect(page.getByRole('list').getByText('綱引き')).toBeVisible();
  });

  test('競技名が空のときは登録しない', async ({ page }) => {
    let dialogMessage = '';

    page.once('dialog', async (dialog) => {
      dialogMessage = dialog.message();
      await dialog.accept();
    });

    await page.getByRole('button', { name: '競技をマスタに登録' }).click();

    await expect.poll(() => dialogMessage).toBe('競技名を入力してください。');
    await expect(page.getByRole('list').getByText('綱引き')).not.toBeVisible();
  });
});
