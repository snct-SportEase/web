import { expect, test } from '@playwright/test';

const mockBackendUrl = process.env.MOCK_BACKEND_URL ?? 'http://127.0.0.1:8081';
const pages = [
  '/dashboard', '/dashboard/guide', '/dashboard/maintenance', '/dashboard/privacy-policy',
  '/dashboard/archive', '/dashboard/archive/1',
  '/dashboard/admin/attendance-management', '/dashboard/admin/barcode-reader',
  '/dashboard/admin/class-management', '/dashboard/admin/insert-matche-result',
  '/dashboard/admin/manage-dashboard', '/dashboard/admin/noon-game-results',
  '/dashboard/admin/role-management', '/dashboard/admin/sport-details-registration',
  '/dashboard/admin/vorting-mic', '/dashboard/root/change-username',
  '/dashboard/root/class-student-count', '/dashboard/root/competition-guidelines-upload',
  '/dashboard/root/event-management', '/dashboard/root/identify-mic', '/dashboard/root/noon-game',
  '/dashboard/root/notification', '/dashboard/root/notification-requests', '/dashboard/root/rainy-mode',
  '/dashboard/root/sport-management', '/dashboard/root/tournament-management', '/dashboard/root/user-promotion',
  '/dashboard/student/class-info', '/dashboard/student/my-page', '/dashboard/student/noon-game',
  '/dashboard/student/notification', '/dashboard/student/notification-request', '/dashboard/student/results',
  '/dashboard/student/score-list', '/dashboard/student/sport-info', '/dashboard/student/timetable',
  '/dashboard/student/tournament'
];

test.use({ viewport: { width: 390, height: 844 } });

test('各画面がスマートフォン幅で表示され、ページ全体が横にはみ出さない', async ({ page, context, request }) => {
  test.setTimeout(120_000);
  await request.post(`${mockBackendUrl}/__reset`);
  await context.addCookies([{ name: 'session_token', value: 'test-session-token', domain: 'localhost', path: '/' }]);
  await context.addInitScript(() => localStorage.setItem('pwa-notification-seen', 'true'));
  const problems: string[] = [];
  for (const path of pages) {
    const response = await page.goto(path, { waitUntil: 'domcontentloaded' });
    if (!response || !response.ok() || new URL(page.url()).pathname !== path) {
      problems.push(`${path}: 表示できませんでした (${response?.status() ?? '応答なし'})`);
      continue;
    }
    const overflow = await page.locator('.main-content').evaluate((element) => element.scrollWidth - element.clientWidth);
    if (overflow > 1) problems.push(`${path}: ${overflow}px 横にはみ出しています`);
  }
  expect(problems).toEqual([]);
});

test('スマートフォンで表とナビゲーションを操作できる', async ({ page, context, request }) => {
  await request.post(`${mockBackendUrl}/__reset`);
  await context.addCookies([{ name: 'session_token', value: 'test-session-token', domain: 'localhost', path: '/' }]);
  await context.addInitScript(() => localStorage.setItem('pwa-notification-seen', 'true'));
  await page.goto('/dashboard/root/event-management', { waitUntil: 'networkidle' });

  const table = page.getByRole('region', { name: '大会一覧' });
  await table.focus();
  await expect(table).toBeFocused();
  expect(await table.evaluate((element) => element.scrollWidth > element.clientWidth)).toBe(true);

  await page.getByRole('button', { name: 'サイドバーを開く' }).click();
  const navigation = page.getByRole('navigation', { name: 'メインメニュー' });
  await expect(navigation.getByRole('link', { name: '大会情報登録・管理' })).toHaveAttribute('aria-current', 'page');
  await page.getByRole('button', { name: 'サイドバーを閉じる' }).click();
  await expect(page.locator('aside')).toHaveAttribute('inert', '');
});
