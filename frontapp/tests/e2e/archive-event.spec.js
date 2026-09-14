import { expect, test } from '@playwright/test';

const mockBackendUrl =
  process.env.MOCK_BACKEND_URL ??
  `http://127.0.0.1:${process.env.MOCK_BACKEND_PORT ?? 8081}`;

test.describe('大会アーカイブ', () => {
  test.beforeEach(async ({ page, context, request }) => {
    await request.post(`${mockBackendUrl}/__reset`);
    await request.put(`${mockBackendUrl}/api/root/events/1`, {
      data: {
        name: '2025春季スポーツ大会',
        year: 2025,
        season: 'spring',
        start_date: '2025-04-01',
        end_date: '2025-04-02',
        status: 'archived',
        survey_url: 'https://example.com/survey'
      }
    });
    await request.post(`${mockBackendUrl}/__set-user`, { data: { user: 'student' } });
    await context.addCookies([{
      name: 'session_token',
      value: 'test-session-token',
      domain: 'localhost',
      path: '/'
    }]);

    await page.goto('/dashboard/archive');
  });

  test('終了した大会の順位と試合結果を振り返る', async ({ page }) => {
    await expect(page.getByRole('heading', { name: '過去の大会アーカイブ' })).toBeVisible();
    await page.getByRole('link', { name: /2025春季スポーツ大会/ }).click();

    await expect(page.getByRole('heading', { name: '2025春季スポーツ大会' })).toBeVisible();
    await expect(page.getByText('🥇 1位')).toBeVisible();
    await expect(page.getByText('1A', { exact: true }).first()).toBeVisible();
    await expect(page.getByText('60', { exact: true })).toBeVisible();

    await page.getByRole('button', { name: '試合結果' }).click();
    await expect(page.getByRole('heading', { name: 'バスケットボール', exact: true })).toBeVisible();
    await expect(page.getByText('3 - 1').first()).toBeVisible();
    await expect(page.getByText('1A', { exact: true }).last()).toBeVisible();
  });
});
