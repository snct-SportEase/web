import { expect, test } from '@playwright/test';

const mockBackendUrl = process.env.MOCK_BACKEND_URL ?? 'http://127.0.0.1:8081';

test.describe.configure({ mode: 'serial' });

test.describe('通知申請管理 (root)', () => {
  test.beforeEach(async ({ page, context, request }) => {
    await request.post(`${mockBackendUrl}/__reset`);

    let activeRequest = {
      id: 1,
      title: 'お知らせ配信依頼',
      body: '明日の集合時刻変更を通知したいです。',
      status: 'pending',
      target_text: '全学生',
      requester: {
        id: 'student-user-1',
        email: 'student1@sendai-nct.jp',
        display_name: '1A 代表'
      },
      messages: [
        {
          id: 1,
          message: '内容を確認お願いします。',
          created_at: '2025-04-01T09:30:00Z',
          sender: {
            id: 'student-user-1',
            email: 'student1@sendai-nct.jp',
            display_name: '1A 代表'
          }
        }
      ]
    };

    await page.route('**/api/root/notification-requests/1', async (route) => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({ request: activeRequest })
      });
    });

    await page.route('**/api/root/notification-requests/1/messages', async (route, request) => {
      const body = JSON.parse(request.postData() ?? '{}');
      activeRequest = {
        ...activeRequest,
        messages: [
          ...activeRequest.messages,
          {
            id: activeRequest.messages.length + 1,
            message: body.message,
            created_at: '2025-04-02T10:00:00Z',
            sender: {
              id: 'test-root-id',
              email: 'root@example.com',
              display_name: '管理者ユーザー'
            }
          }
        ]
      };

      await route.fulfill({
        status: 201,
        contentType: 'application/json',
        body: JSON.stringify({ ok: true })
      });
    });

    await page.route('**/api/root/notification-requests/1/decision', async (route, request) => {
      const body = JSON.parse(request.postData() ?? '{}');
      activeRequest = { ...activeRequest, status: body.status };

      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({ ok: true })
      });
    });

    await context.addCookies([{ name: 'session_token', value: 'test-session-token', domain: 'localhost', path: '/' }]);
    await page.goto('/dashboard/root/notification-requests');
    await expect(page.getByRole('heading', { name: '通知申請管理' })).toBeVisible();
    await expect(page.getByRole('button', { name: 'メッセージを送信' })).toBeEnabled();
  });

  test('申請一覧と詳細を表示できる', async ({ page }) => {
    await expect(page.getByText('お知らせ配信依頼').nth(0)).toBeVisible();
    await expect(page.getByText('内容を確認お願いします。')).toBeVisible();
  });

  test('メッセージを送信できる', async ({ page }) => {
    const messageInput = page.getByRole('textbox', { name: 'メッセージを送信' });
    await messageInput.fill('了解しました。');
    await expect(messageInput).toHaveValue('了解しました。');

    const requestPromise = page.waitForRequest((request) => request.url().endsWith('/api/root/notification-requests/1/messages') && request.method() === 'POST');
    const refreshPromise = page.waitForRequest((request) => request.url().endsWith('/api/root/notification-requests/1') && request.method() === 'GET');
    await page.getByRole('button', { name: 'メッセージを送信' }).click();
    const req = await requestPromise;
    await refreshPromise;
    expect(JSON.parse(req.postData() ?? '{}')).toEqual({ message: '了解しました。' });
    await expect(messageInput).toHaveValue('');
  });

  test('通知申請を承認できる', async ({ page }) => {
    const requestPromise = page.waitForRequest((request) => request.url().endsWith('/api/root/notification-requests/1/decision') && request.method() === 'POST');
    const refreshPromise = page.waitForRequest((request) => request.url().endsWith('/api/root/notification-requests/1') && request.method() === 'GET');
    await page.getByRole('button', { name: '承認する' }).click();
    const req = await requestPromise;
    await refreshPromise;
    expect(JSON.parse(req.postData() ?? '{}')).toEqual({ status: 'approved' });
  });
});
