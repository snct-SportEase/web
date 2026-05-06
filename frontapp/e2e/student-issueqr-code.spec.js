import { expect, test } from '@playwright/test';

const mockBackendUrl = process.env.MOCK_BACKEND_URL ?? 'http://127.0.0.1:8081';

test.describe('QRコード発行 (student)', () => {
  test.beforeEach(async ({ page, context, request }) => {
    await request.post(`${mockBackendUrl}/__reset`);
    await context.addCookies([{ name: 'session_token', value: 'test-session-token', domain: 'localhost', path: '/' }]);
    await page.goto('/dashboard/student/issueqr-code');
    await expect(page.getByRole('heading', { name: 'QRコード発行' })).toBeVisible();
  });

  test('競技を選択してQRコードを生成できる', async ({ page }) => {
    await expect(page.getByText('QRコードは発行から10秒間のみ有効です')).toBeVisible();

    await page.locator('select').selectOption('1-1');

    const requestPromise = page.waitForRequest((request) => request.url().endsWith('/api/qrcode/generate') && request.method() === 'POST');
    await page.getByRole('button', { name: 'QRコードを生成' }).click();
    const request = await requestPromise;

    expect(JSON.parse(request.postData() ?? '{}')).toEqual({
      event_id: 1,
      sport_id: 1
    });

    await expect(page.getByRole('heading', { name: 'QRコード' })).toBeVisible();
    await expect(page.getByAltText('QR Code')).toBeVisible();
    await expect(page.getByText('競技: バスケットボール')).toBeVisible();
    await expect(page.getByText('有効期限まで残り時間:')).toBeVisible();
  });

  test('有効期限が切れると新しいQRコードを再発行できる', async ({ page }) => {
    await page.locator('select').selectOption('1-1');
    await page.getByRole('button', { name: 'QRコードを生成' }).click();

    await expect(page.getByAltText('QR Code')).toBeVisible();
    await page.waitForFunction(() => !document.querySelector('img[alt="QR Code"]'), null, { timeout: 15000 });

    await expect(page.getByRole('button', { name: 'QRコードを生成' })).toBeVisible();
    await expect(page.getByAltText('QR Code')).toHaveCount(0);
  });
});
