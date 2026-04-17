import { test, expect } from '@playwright/test';

const mockBackendUrl = process.env.MOCK_BACKEND_URL ?? 'http://127.0.0.1:8081';

test.describe.configure({ mode: 'serial' });

test.describe('競技情報登録・管理 (root)', () => {
  test.beforeEach(async ({ page, context, request }) => {
    await request.post(`${mockBackendUrl}/__reset`);

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
    await page.evaluate(() => {
      window.alert = () => {};
    });

    const createRequest = page.waitForRequest((request) => {
      if (request.url().endsWith('/api/root/sports') && request.method() === 'POST') {
        const body = JSON.parse(request.postData() ?? '{}');
        return body.name === '綱引き';
      }

      return false;
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

  test('競技未選択では大会へ割り当てボタンが無効', async ({ page }) => {
    await expect(page.getByRole('button', { name: '大会に競技を割り当てる' })).toBeDisabled();
    await expect(page.getByText('割り当て済み競技一覧 (0件)')).toBeVisible();
  });

  test('競技を大会へ割り当てできる', async ({ page }) => {
    await page.getByLabel('割り当てる競技').selectOption('1');
    await page.getByLabel('場所').selectOption('gym1');
    await page.getByLabel('概要 (任意)').fill('屋内メイン競技');
    await page.getByLabel('ルール詳細 (任意)').fill('# バスケットボール');
    await page.evaluate(() => {
      window.alert = () => {};
    });

    const assignRequest = page.waitForRequest((request) => {
      if (request.url().endsWith('/api/admin/events/1/sports') && request.method() === 'POST') {
        const body = JSON.parse(request.postData() ?? '{}');
        return body.sport_id === 1 && body.location === 'gym1';
      }

      return false;
    });

    await page.getByRole('button', { name: '大会に競技を割り当てる' }).click();

    const request = await assignRequest;
    const body = JSON.parse(request.postData() ?? '{}');
    expect(body).toEqual(expect.objectContaining({
      sport_id: 1,
      location: 'gym1',
      description: '屋内メイン競技',
      rules: '# バスケットボール',
      rules_type: 'markdown'
    }));

    await expect(page.getByLabel('割り当てる競技')).toHaveValue('');
    await expect(page.getByLabel('場所')).toHaveValue('other');
    await expect(page.getByLabel('概要 (任意)')).toHaveValue('');
    await expect(page.getByLabel('ルール詳細 (任意)')).toHaveValue('');
  });
});
