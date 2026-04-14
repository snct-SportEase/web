import { page } from '@vitest/browser/context';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { render } from 'vitest-browser-svelte';
import Page from './+page.svelte';

const mocks = vi.hoisted(() => ({
  pageData: {},
  goto: vi.fn(),
  invalidateAll: vi.fn(async () => {})
}));

vi.mock('$app/stores', () => ({
  page: {
    subscribe(fn) {
      fn({ data: mocks.pageData });
      return () => {};
    }
  }
}));

vi.mock('$app/navigation', () => ({
  goto: mocks.goto,
  invalidateAll: mocks.invalidateAll
}));

describe('Notification Requests Page', () => {
  let fetchMock;
  let activeRequest;

  beforeEach(() => {
    activeRequest = {
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

    mocks.pageData = {
      requests: [
        {
          id: 1,
          title: activeRequest.title,
          body: activeRequest.body,
          status: activeRequest.status,
          target_text: activeRequest.target_text,
          requester: activeRequest.requester
        }
      ],
      activeRequest,
      error: null
    };

    fetchMock = vi.fn((url, options = {}) => {
      if (url === '/api/root/notification-requests/1' && !options.method) {
        return Promise.resolve({ ok: true, json: () => Promise.resolve({ request: activeRequest }) });
      }

      if (url === '/api/root/notification-requests/1/messages' && options.method === 'POST') {
        const body = JSON.parse(options.body);
        activeRequest = {
          ...activeRequest,
          messages: [
            ...activeRequest.messages,
            {
              id: 2,
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
        return Promise.resolve({ ok: true, json: () => Promise.resolve({ ok: true }) });
      }

      if (url === '/api/root/notification-requests/1/decision' && options.method === 'POST') {
        const body = JSON.parse(options.body);
        activeRequest = { ...activeRequest, status: body.status };
        return Promise.resolve({ ok: true, json: () => Promise.resolve({ ok: true }) });
      }

      return Promise.resolve({ ok: true, json: () => Promise.resolve({}) });
    });

    vi.stubGlobal('fetch', fetchMock);
  });

  it('初期表示で申請一覧と詳細を表示できる', async () => {
    render(Page);

    await expect.element(page.getByRole('heading', { name: '通知申請管理' })).toBeInTheDocument();
    await expect.element(page.getByText('お知らせ配信依頼').nth(0)).toBeInTheDocument();
    await expect.element(page.getByText('内容を確認お願いします。')).toBeInTheDocument();
  });

  it('メッセージ送信時にPOSTして詳細を更新する', async () => {
    render(Page);

    await page.getByLabelText('メッセージを送信').fill('了解しました。');
    await page.getByRole('button', { name: 'メッセージを送信' }).click();

    const postCall = fetchMock.mock.calls.find(([url, options]) => url === '/api/root/notification-requests/1/messages' && options?.method === 'POST');
    expect(postCall).toBeTruthy();
    expect(JSON.parse(postCall[1].body)).toEqual({ message: '了解しました。' });
    await expect.element(page.getByText('了解しました。')).toBeInTheDocument();
  });

  it('承認操作でdecision APIを呼ぶ', async () => {
    render(Page);

    await page.getByRole('button', { name: '承認する' }).click();

    const postCall = fetchMock.mock.calls.find(([url, options]) => url === '/api/root/notification-requests/1/decision' && options?.method === 'POST');
    expect(postCall).toBeTruthy();
    expect(JSON.parse(postCall[1].body)).toEqual({ status: 'approved' });
  });
});
