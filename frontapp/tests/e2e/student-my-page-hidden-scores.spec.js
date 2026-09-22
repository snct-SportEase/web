import { expect, test } from '@playwright/test';

const mockBackendUrl = process.env.MOCK_BACKEND_URL ?? 'http://127.0.0.1:8081';

test.describe.configure({ mode: 'serial' });

test.describe('マイページ 得点非表示 (student)', () => {
  test.beforeEach(async ({ context, request }) => {
    await request.post(`${mockBackendUrl}/__reset`);
    await context.addCookies([{ name: 'session_token', value: 'test-session-token', domain: 'localhost', path: '/' }]);
  });

  test('得点非表示中の一般ユーザーには順位と得点を表示しない', async ({ page, request }) => {
    await request.post(`${mockBackendUrl}/__set-user`, {
      data: { user: 'student' }
    });
    await request.post(`${mockBackendUrl}/__set-active-event`, {
      data: { event_id: 1, hide_scores: true }
    });

    await page.goto('/dashboard/student/my-page');

    await expect(page.getByRole('heading', { name: 'マイページ' })).toBeVisible();
    await expect(page.getByText('得点・順位は現在非表示です。')).toBeVisible();
    await expect(page.getByText('公開されるまでお待ちください。')).toBeVisible();

    await expect(page.getByText('クラス成績')).toHaveCount(0);
    await expect(page.getByText('獲得ポイント内訳')).toHaveCount(0);
    await expect(page.getByText('1位')).toHaveCount(0);
    await expect(page.getByText('獲得 60 点')).toHaveCount(0);
  });

  test('得点非表示中でも管理者はマイページの得点を確認できる', async ({ page, request }) => {
    await request.post(`${mockBackendUrl}/__set-user`, {
      data: { user: 'admin' }
    });
    await request.post(`${mockBackendUrl}/__set-active-event`, {
      data: { event_id: 1, hide_scores: true }
    });

    await page.goto('/dashboard/student/my-page');

    await expect(page.getByRole('heading', { name: 'マイページ' })).toBeVisible();
    await expect(page.getByText('得点・順位は現在非表示です。')).toHaveCount(0);
    await expect(page.getByText('クラス成績')).toBeVisible();
    await expect(page.getByText('2位')).toHaveCount(2);
    await expect(page.getByText('獲得 50 点')).toBeVisible();
  });

  test('参加試合・結果・クラス状況・通知・大会資料をまとめて確認できる', async ({ page, request }) => {
    await request.post(`${mockBackendUrl}/__set-user`, { data: { user: 'student' } });
    await request.post(`${mockBackendUrl}/__set-active-event`, {
      data: {
        event_id: 1,
        competition_guidelines_pdf_url: 'https://example.com/event-guidelines.pdf',
        survey_url: 'https://example.com/survey',
        is_survey_published: true
      }
    });
    await request.post(`${mockBackendUrl}/api/admin/events/1/sports`, {
      data: { sport_id: 1, description: '', location: 'gym1' }
    });
    await request.put(`${mockBackendUrl}/api/admin/events/1/sports/1/details`, {
      data: { description: '', rules_pdf_url: 'https://example.com/basketball-rules.pdf' }
    });

    await page.goto('/dashboard/student/my-page');

    await expect(page.getByRole('heading', { name: '参加試合' })).toBeVisible();
    await expect(page.getByRole('heading', { name: '試合結果' })).toBeVisible();
    await expect(page.getByRole('heading', { name: 'クラス状況' })).toBeVisible();
    await expect(page.getByRole('heading', { name: '最新のお知らせ' })).toBeVisible();
    await expect(page.getByText('大会開催のお知らせ')).toBeVisible();
    await expect(page.getByRole('link', { name: /大会要項PDF/ })).toHaveAttribute('href', 'https://example.com/event-guidelines.pdf');
    await expect(page.getByRole('link', { name: /バスケットボール 競技要項PDF/ })).toHaveAttribute('href', 'https://example.com/basketball-rules.pdf');
    await expect(page.getByRole('link', { name: /アンケートに回答する/ })).toHaveAttribute('href', 'https://example.com/survey');
  });
});
