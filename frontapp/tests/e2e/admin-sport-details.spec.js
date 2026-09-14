import { expect, test } from '@playwright/test';

const mockBackendUrl =
  process.env.MOCK_BACKEND_URL ??
  `http://127.0.0.1:${process.env.MOCK_BACKEND_PORT ?? 8081}`;

test.describe('競技詳細設定 (admin)', () => {
  test.beforeEach(async ({ page, context, request }) => {
    await request.post(`${mockBackendUrl}/__reset`);
    await request.post(`${mockBackendUrl}/__set-user`, { data: { user: 'admin' } });
    await request.post(`${mockBackendUrl}/__set-active-event`, { data: { event_id: 1 } });
    await request.post(`${mockBackendUrl}/api/admin/events/1/sports`, {
      data: {
        sport_id: 1,
        description: '既存の競技説明',
        location: 'gym1',
        min_capacity: 5,
        max_capacity: 10
      }
    });
    await context.addCookies([{
      name: 'session_token',
      value: 'test-session-token',
      domain: 'localhost',
      path: '/'
    }]);
    page.on('dialog', (dialog) => void dialog.accept());

    await page.goto('/dashboard/admin/sport-details-registration');
    await expect(page.getByRole('heading', { name: '競技詳細情報登録' })).toBeVisible();
    await expect(page.getByLabel('競技選択')).toContainText('バスケットボール');
    await page.getByLabel('競技選択').selectOption('1');
    await expect(page.getByRole('heading', { name: '競技概要' })).toBeVisible();
  });

  test('競技説明と定員を更新して再取得する', async ({ page }) => {
    const description = page.getByRole('heading', { name: '競技概要' }).locator('..').locator('textarea');
    await expect(description).toHaveValue('既存の競技説明');
    await description.fill('当日は5人制、交代自由で実施します。');

    const detailsResponse = page.waitForResponse((response) =>
      response.url().endsWith('/api/admin/events/1/sports/1/details')
        && response.request().method() === 'PUT'
    );
    await page.getByRole('button', { name: '保存', exact: true }).click();
    await detailsResponse;
    await expect(description).toHaveValue('当日は5人制、交代自由で実施します。');

    await page.getByLabel('最低定員').first().fill('6');
    await page.getByLabel('最高定員').first().fill('12');
    const capacityResponse = page.waitForResponse((response) =>
      response.url().endsWith('/api/admin/events/1/sports/1/capacity')
        && response.request().method() === 'PUT'
    );
    await page.getByRole('button', { name: '定員設定を保存', exact: true }).click();
    await capacityResponse;

    await expect(page.getByText('現在の設定: 6 〜 12')).toBeVisible();
  });
});
