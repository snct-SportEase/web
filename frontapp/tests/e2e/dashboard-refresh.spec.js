import { expect, test } from '@playwright/test';

const mockBackendUrl = process.env.MOCK_BACKEND_URL ?? 'http://127.0.0.1:8081';

for (const [name, width] of [['PC', 1280], ['スマホ', 375]]) {
  test.describe(`手動更新 (${name})`, () => {
    test.use({ viewport: { width, height: 812 } });

    test('現在のURLとログインを維持し、最新のユーザー情報を取得する', async ({ page, context, request }) => {
      await request.post(`${mockBackendUrl}/__reset`);
      await context.addCookies([
        { name: 'session_token', value: 'test-session-token', domain: 'localhost', path: '/' }
      ]);
      await context.addInitScript(() => {
        localStorage.setItem('pwa-notification-seen', 'true');
      });
      await page.goto('/dashboard/root/change-username?refresh_test=1#refresh-check');
      const originalUrl = page.url();
      const studentRow = page.getByRole('row').filter({ hasText: 'student1@sendai-nct.jp' });
      await expect(studentRow).toContainText('山田太郎');

      const update = await request.put(`${mockBackendUrl}/api/root/users/display-name`, {
        data: { user_id: 'user-1', display_name: '更新後の表示名' }
      });
      expect(update.ok()).toBe(true);
      await expect(studentRow).not.toContainText('更新後の表示名');

      const refreshButton = page.getByRole('button', { name: 'ページを更新', exact: true });
      if (width < 768) {
        await expect(refreshButton).toBeHidden();
        await page.getByLabel('ヘッダメニューを開く').click();
      }
      await Promise.all([page.waitForEvent('load'), refreshButton.click()]);

      await expect(page).toHaveURL(originalUrl);
      await expect(page.getByRole('button', { name: 'Logout', exact: true })).toBeVisible();
      await expect(studentRow).toContainText('更新後の表示名');
      if (width < 768) await expect(refreshButton).toBeHidden();
    });
  });
}
