import { page } from '@vitest/browser/context';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { render } from 'vitest-browser-svelte';
import Page from './+page.svelte';

vi.mock('$lib/stores/eventStore.js', () => ({
  activeEvent: {
    init: vi.fn(async () => ({ id: 1, name: '2025春季スポーツ大会' })),
    subscribe: vi.fn((callback) => {
      callback({
        id: 1,
        name: '2025春季スポーツ大会',
        start_date: '2025-04-01T00:00:00Z'
      });

      return () => {};
    })
  }
}));

describe('Sport Management Page', () => {
  let fetchMock;
  let alertMock;
  let sports;

  beforeEach(() => {
    sports = [
      { id: 1, name: 'バスケットボール' },
      { id: 2, name: 'バレーボール' }
    ];

    fetchMock = vi.fn((url, options = {}) => {
      if (url === '/api/root/sports') {
        if (options.method === 'POST') {
          const body = JSON.parse(options.body);
          const nextSport = { id: sports.length + 1, name: body.name };
          sports = [...sports, nextSport];

          return Promise.resolve({
            ok: true,
            json: () => Promise.resolve(nextSport)
          });
        }

        return Promise.resolve({
          ok: true,
          json: () => Promise.resolve(sports)
        });
      }

      if (url === '/api/events/1/sports') {
        return Promise.resolve({
          ok: true,
          json: () => Promise.resolve([])
        });
      }

      return Promise.resolve({
        ok: true,
        json: () => Promise.resolve({})
      });
    });

    alertMock = vi.fn();

    vi.stubGlobal('fetch', fetchMock);
    vi.stubGlobal('alert', alertMock);
  });

  it('初期表示で競技マスタ一覧とアクティブ大会が表示されること', async () => {
    render(Page);

    await expect.element(page.getByRole('heading', { name: '大会競技管理ダッシュボード' })).toBeInTheDocument();
    await expect.element(page.getByText('2025春季スポーツ大会')).toBeInTheDocument();
    await expect.element(page.getByRole('list').getByText('バスケットボール')).toBeInTheDocument();
    await expect.element(page.getByRole('list').getByText('バレーボール')).toBeInTheDocument();
  });

  it('競技名が空のときは登録せずにアラートを表示すること', async () => {
    render(Page);

    await page.getByRole('button', { name: '競技をマスタに登録' }).click();

    expect(alertMock).toHaveBeenCalledWith('競技名を入力してください。');

    const createCall = fetchMock.mock.calls.find(([url, options]) => {
      return url === '/api/root/sports' && options?.method === 'POST';
    });
    expect(createCall).toBeFalsy();
  });

  it('新しい競技をマスタ登録するとPOST送信して一覧を更新すること', async () => {
    render(Page);

    await page.getByPlaceholder('新しい競技名を入力').fill('綱引き');
    await page.getByRole('button', { name: '競技をマスタに登録' }).click();

    const createCall = fetchMock.mock.calls.find(([url, options]) => {
      return url === '/api/root/sports' && options?.method === 'POST';
    });

    expect(createCall).toBeTruthy();
    expect(JSON.parse(createCall[1].body)).toEqual({ name: '綱引き' });
    expect(alertMock).toHaveBeenCalledWith('新しい競技を登録しました。');
    await expect.element(page.getByRole('list').getByText('綱引き')).toBeInTheDocument();
  });
});
