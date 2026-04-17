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
      if (request.url().includes('/api/root/notifications') && request.method() === 'POST') {
        requestSent = true;
      }
    });

    await expect(page.getByRole('button', { name: '通知を送信' })).toBeEnabled();
    await page.getByRole('button', { name: '通知を送信' }).click();

    await expect.poll(() => requestSent).toBe(false);
    await expect(page.getByRole('textbox', { name: 'タイトル' })).toHaveValue('');
  });

  test('新しい通知を送信できる', async ({ page }) => {
    await page.getByLabel('タイトル').fill('競技開始時間変更');
    await page.getByLabel('本文').fill('バスケットボールの開始時刻が変更になりました。');
    // チェックボックスをevaluateで操作
    await page.getByRole('checkbox', { name: '管理者' }).evaluate((el) => {
      el.checked = true;
      el.dispatchEvent(new Event('change', { bubbles: true }));
    });
    await page.waitForTimeout(500); // 状態更新待ち

    const createRequest = page.waitForRequest((request) => {
      if (request.url().includes('/api/root/notifications') && request.method() === 'POST') {
        const body = JSON.parse(request.postData() ?? '{}');
        return body.title === '競技開始時間変更';
      }

      return false;
    });

    await expect(page.getByRole('button', { name: '通知を送信' })).toBeEnabled();
    await page.getByRole('button', { name: '通知を送信' }).click();

    const request = await createRequest;
    const body = JSON.parse(request.postData() ?? '{}');
    expect(body.title).toBe('競技開始時間変更');
    expect(body.body).toBe('バスケットボールの開始時刻が変更になりました。');
    expect(body.target_roles).toContain('student');
    expect(body.target_roles).toContain('admin');
    expect(body.type).toBe('general');

    await expect(page.getByText('通知を送信しました。')).toBeVisible();
    await expect(page.getByLabel('タイトル')).toHaveValue('');
    await expect(page.getByLabel('本文')).toHaveValue('');
  });
});
