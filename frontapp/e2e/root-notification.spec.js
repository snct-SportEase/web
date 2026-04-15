import { test, expect } from '@playwright/test';

const mockBackendUrl = process.env.MOCK_BACKEND_URL ?? 'http://127.0.0.1:8081';

test.describe.configure({ mode: 'serial' });

test.describe('通知管理 (root)', () => {
  test.beforeEach(async ({ page, context, request }) => {
    await request.post(`${mockBackendUrl}/__reset`);

    await context.addCookies([{
      name: 'session_token',
      value: 'test-session-token',
      domain: 'localhost',
      path: '/'
    }]);

    await page.goto('/dashboard/root/notification');
    await expect(page.getByText('大会開催のお知らせ')).toBeVisible();
    await expect(page.getByRole('button', { name: '通知を送信' })).toBeEnabled();
  });

  test('送信済み通知一覧を表示できる', async ({ page }) => {
    await expect(page.getByRole('heading', { name: '通知管理' })).toBeVisible();
    await expect(page.getByText('春季スポーツ大会を開催します。')).toBeVisible();
  });

  test('タイトルと本文が空のときは送信しない', async ({ page }) => {
    let requestSent = false;
    page.on('request', (request) => {
      if (request.url().endsWith('/api/root/notifications') && request.method() === 'POST') {
        requestSent = true;
      }
    });

    await page.getByRole('button', { name: '通知を送信' }).click();

    await expect.poll(() => requestSent).toBe(false);
    await expect(page.getByRole('textbox', { name: 'タイトル' })).toHaveValue('');
  });

  test('新しい通知を送信できる', async ({ page }) => {
    const titleInput = page.getByLabel('タイトル');
    const bodyInput = page.getByLabel('本文');
    const adminCheckbox = page.getByRole('checkbox', { name: '管理者' });

    await titleInput.fill('競技開始時間変更');
    await bodyInput.fill('バスケットボールの開始時刻が変更になりました。');
    await adminCheckbox.click();
    await expect(titleInput).toHaveValue('競技開始時間変更');
    await expect(bodyInput).toHaveValue('バスケットボールの開始時刻が変更になりました。');
    await expect(adminCheckbox).toBeChecked();

    const createRequest = page.waitForRequest((request) => {
      if (request.url().endsWith('/api/root/notifications') && request.method() === 'POST') {
        const body = JSON.parse(request.postData() ?? '{}');
        return body.title === '競技開始時間変更';
      }

      return false;
    });

    await page.getByRole('button', { name: '通知を送信' }).click();

    const request = await createRequest;
    const body = JSON.parse(request.postData() ?? '{}');
    expect(body).toEqual({
      title: '競技開始時間変更',
      body: 'バスケットボールの開始時刻が変更になりました。',
      type: 'general',
      target_roles: ['student', 'admin']
    });

    await expect(page.getByText('通知を送信しました。')).toBeVisible();
    await expect(page.getByText('競技開始時間変更')).toBeVisible({ timeout: 15000 });
  });
});
