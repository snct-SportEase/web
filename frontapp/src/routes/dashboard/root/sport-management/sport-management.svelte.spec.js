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
  let assignedSports;

  beforeEach(() => {
    sports = [
      { id: 1, name: 'バスケットボール' },
      { id: 2, name: 'バレーボール' }
    ];
    assignedSports = [];

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
          json: () => Promise.resolve(assignedSports)
        });
      }

      if (url === '/api/admin/events/1/sports' && options.method === 'POST') {
        const body = JSON.parse(options.body);
        const nextAssignedSport = {
          event_id: 1,
          sport_id: body.sport_id,
          description: body.description ?? '',
          rules: body.rules ?? '',
          location: body.location ?? 'other',
          rules_type: body.rules_type ?? 'markdown'
        };
        assignedSports = [...assignedSports, nextAssignedSport];

        return Promise.resolve({
          ok: true,
          json: () => Promise.resolve(nextAssignedSport)
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

  it('競技未選択では大会へ割り当てボタンが無効であること', async () => {
    render(Page);

    await expect.element(page.getByRole('button', { name: '大会に競技を割り当てる' })).toBeDisabled();

    const assignCall = fetchMock.mock.calls.find(([url, options]) => {
      return url === '/api/admin/events/1/sports' && options?.method === 'POST';
    });
    expect(assignCall).toBeFalsy();
  });

  it('競技を大会に割り当てるとPOST送信して割り当て済み一覧に表示されること', async () => {
    render(Page);

    await page.getByLabelText('割り当てる競技').selectOptions('1');
    await page.getByLabelText('場所').selectOptions('gym1');
    await page.getByLabelText('概要 (任意)').fill('屋内メイン競技');
    await page.getByLabelText('ルール詳細 (任意)').fill('# バスケットボール');
    await page.getByRole('button', { name: '大会に競技を割り当てる' }).click();

    const assignCall = fetchMock.mock.calls.find(([url, options]) => {
      return url === '/api/admin/events/1/sports' && options?.method === 'POST';
    });

    expect(assignCall).toBeTruthy();
    expect(JSON.parse(assignCall[1].body)).toEqual(expect.objectContaining({
      sport_id: 1,
      location: 'gym1',
      description: '屋内メイン競技',
      rules: '# バスケットボール',
      rules_type: 'markdown'
    }));
    expect(alertMock).toHaveBeenCalledWith('競技を大会に割り当てました。');
    await expect.element(page.getByText('割り当て済み競技一覧 (1件)')).toBeInTheDocument();
    await expect.element(page.getByRole('cell', { name: 'バスケットボール' })).toBeInTheDocument();
    await expect.element(page.getByRole('cell', { name: 'gym1' })).toBeInTheDocument();
  });
});
