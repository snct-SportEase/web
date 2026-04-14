import { page } from '@vitest/browser/context';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { render } from 'vitest-browser-svelte';
import Page from './+page.svelte';

vi.mock('$app/stores', () => ({
  page: {
    subscribe: (callback) => {
      callback({
        data: {
          notifications: [
            {
              id: 1,
              title: '大会開催のお知らせ',
              body: '春季スポーツ大会を開催します。',
              target_roles: ['student'],
              created_at: '2025-04-01T09:00:00Z'
            }
          ],
          roles: [
            { id: 1, name: 'student' },
            { id: 2, name: 'admin' },
            { id: 3, name: 'root' }
          ]
        }
      });

      return () => {};
    }
  }
}));

describe('Notification Management Page', () => {
  let fetchMock;

  beforeEach(() => {
    const notifications = [
      {
        id: 1,
        title: '大会開催のお知らせ',
        body: '春季スポーツ大会を開催します。',
        target_roles: ['student'],
        created_at: '2025-04-01T09:00:00Z'
      }
    ];

    fetchMock = vi.fn((url, options = {}) => {
      if (url === '/api/root/notifications' && options.method === 'POST') {
        const body = JSON.parse(options.body);
        notifications.unshift({
          id: 2,
          ...body,
          created_at: '2025-04-02T10:00:00Z'
        });

        return Promise.resolve({
          ok: true,
          json: () => Promise.resolve({ notification: notifications[0] })
        });
      }

      if (typeof url === 'string' && url.startsWith('/api/notifications?')) {
        return Promise.resolve({
          ok: true,
          json: () => Promise.resolve({ notifications })
        });
      }

      return Promise.resolve({
        ok: true,
        json: () => Promise.resolve({})
      });
    });

    vi.stubGlobal('fetch', fetchMock);
  });

  it('初期表示で通知一覧を表示できる', async () => {
    render(Page);

    await expect.element(page.getByRole('heading', { name: '通知管理' })).toBeInTheDocument();
    await expect.element(page.getByText('大会開催のお知らせ')).toBeInTheDocument();
    await expect.element(page.getByText('春季スポーツ大会を開催します。')).toBeInTheDocument();
  });

  it('タイトルと本文が空のときは送信せずエラーを表示する', async () => {
    render(Page);

    await page.getByRole('button', { name: '通知を送信' }).click();

    await expect.element(page.getByText('タイトルと本文を入力してください。')).toBeInTheDocument();

    const createCall = fetchMock.mock.calls.find(([url, options]) => {
      return url === '/api/root/notifications' && options?.method === 'POST';
    });
    expect(createCall).toBeFalsy();
  });

  it('通知を作成して送信済み一覧を更新できる', async () => {
    render(Page);

    await page.getByLabelText('タイトル').fill('競技開始時間変更');
    await page.getByLabelText('本文').fill('バスケットボールの開始時刻が変更になりました。');
    await page.getByLabelText('管理者').click();
    await page.getByRole('button', { name: '通知を送信' }).click();

    const createCall = fetchMock.mock.calls.find(([url, options]) => {
      return url === '/api/root/notifications' && options?.method === 'POST';
    });

    expect(createCall).toBeTruthy();
    expect(JSON.parse(createCall[1].body)).toEqual({
      title: '競技開始時間変更',
      body: 'バスケットボールの開始時刻が変更になりました。',
      type: 'general',
      target_roles: ['student', 'admin']
    });

    await expect.element(page.getByText('通知を送信しました。')).toBeInTheDocument();
    await expect.element(page.getByText('競技開始時間変更')).toBeInTheDocument();
  });
});
