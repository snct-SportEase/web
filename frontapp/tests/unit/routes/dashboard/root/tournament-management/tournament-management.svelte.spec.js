import { page } from '@vitest/browser/context';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { render } from 'vitest-browser-svelte';
import Page from '$src/routes/dashboard/root/tournament-management/+page.svelte';

const mocks = vi.hoisted(() => ({
  active: {
    id: 1,
    name: '2025春季スポーツ大会'
  }
}));

vi.mock('$lib/stores/eventStore.js', () => ({
  activeEvent: {
    subscribe(fn) {
      fn(mocks.active);
      return () => {};
    },
    init: vi.fn(async () => mocks.active)
  }
}));

vi.mock('bracketry', () => ({
  createBracket: vi.fn()
}));

vi.mock('svelte-dnd-action', () => ({
  dndzone: () => ({})
}));

describe('Tournament Management Page', () => {
  let fetchMock;
  let createObjectURLMock;
  let revokeObjectURLMock;
  let anchorClickMock;
  const originalCreateElement = document.createElement.bind(document);

  afterEach(() => {
    vi.restoreAllMocks();
  });

  beforeEach(() => {
    fetchMock = vi.fn((url, options = {}) => {
      if (url === '/api/root/events/1/tournaments' && !options.method) {
        return Promise.resolve({ ok: true, json: () => Promise.resolve([]) });
      }

      if (url === '/api/root/events/1/tournaments/generate-preview' && options.method === 'POST') {
        return Promise.resolve({ ok: true, json: () => Promise.resolve([]) });
      }

      if (url === '/api/root/events/1/tournaments/bulk-create' && options.method === 'POST') {
        return Promise.resolve({ ok: true, json: () => Promise.resolve({ message: 'saved' }) });
      }

      if (url === '/api/root/events/1/tournaments/export/excel' && !options.method) {
        return Promise.resolve({
          ok: true,
          blob: () => Promise.resolve(new Blob(['mock-excel'])),
          headers: {
            get: () => 'attachment; filename="event_1_tournaments.xlsx"'
          }
        });
      }

      return Promise.resolve({ ok: true, json: () => Promise.resolve({}) });
    });

    createObjectURLMock = vi.fn(() => 'blob:mock-excel');
    revokeObjectURLMock = vi.fn();
    anchorClickMock = vi.fn();

    vi.stubGlobal('fetch', fetchMock);
    vi.stubGlobal('confirm', vi.fn(() => true));
    vi.stubGlobal('alert', vi.fn());
    vi.stubGlobal('URL', {
      createObjectURL: createObjectURLMock,
      revokeObjectURL: revokeObjectURLMock
    });

    vi.spyOn(document, 'createElement').mockImplementation((tagName) => {
      const element = originalCreateElement(tagName);
      if (String(tagName).toLowerCase() === 'a') {
        element.click = anchorClickMock;
      }
      return element;
    });
  });

  it('初期表示で見出しを表示できる', async () => {
    render(Page);

    await expect.element(page.getByRole('heading', { name: 'トーナメント生成・管理' })).toBeInTheDocument();
    await expect.element(page.getByRole('button', { name: 'トーナメントプレビューを生成' })).toBeInTheDocument();
  });

  it('DBに保存済みのトーナメントを表示できる', async () => {
    const savedTournaments = [
      {
        id: 1,
        name: 'バスケットボール',
        data: JSON.stringify({
          rounds: [{ name: '決勝' }],
          matches: [
            {
              roundIndex: 0,
              order: 0,
              sides: [{ contestantId: 'c0' }, { contestantId: 'c1' }]
            }
          ],
          contestants: {
            c0: { players: [{ title: '1年A組' }] },
            c1: { players: [{ title: '2年B組' }] }
          }
        })
      }
    ];

    fetchMock = vi.fn((url, options = {}) => {
      if (url === '/api/root/events/1/tournaments' && !options.method) {
        return Promise.resolve({ ok: true, json: () => Promise.resolve(savedTournaments) });
      }
      return Promise.resolve({ ok: true, json: () => Promise.resolve({}) });
    });
    vi.stubGlobal('fetch', fetchMock);

    render(Page);

    await expect.element(page.getByRole('heading', { name: '生成済みトーナメント一覧' })).toBeInTheDocument();
    await expect.element(page.getByText('バスケットボール')).toBeInTheDocument();
    await expect.element(page.getByRole('button', { name: '保存済みトーナメントをExcel出力' })).toBeInTheDocument();
  });

  it('保存済みトーナメントをExcel出力できる', async () => {
    const savedTournaments = [
      {
        id: 1,
        name: 'バスケットボール',
        data: JSON.stringify({
          rounds: [{ name: '決勝' }],
          matches: [
            {
              roundIndex: 0,
              order: 0,
              sides: [{ contestantId: 'c0' }, { contestantId: 'c1' }]
            }
          ],
          contestants: {
            c0: { players: [{ title: '1年A組' }] },
            c1: { players: [{ title: '2年B組' }] }
          }
        })
      }
    ];

    fetchMock = vi.fn((url, options = {}) => {
      if (url === '/api/root/events/1/tournaments' && !options.method) {
        return Promise.resolve({ ok: true, json: () => Promise.resolve(savedTournaments) });
      }
      if (url === '/api/root/events/1/tournaments/export/excel' && !options.method) {
        return Promise.resolve({
          ok: true,
          blob: () => Promise.resolve(new Blob(['mock-excel'])),
          headers: {
            get: () => 'attachment; filename="event_1_tournaments.xlsx"'
          }
        });
      }
      return Promise.resolve({ ok: true, json: () => Promise.resolve({}) });
    });
    vi.stubGlobal('fetch', fetchMock);

    render(Page);

    await page.getByRole('button', { name: '保存済みトーナメントをExcel出力' }).click();

    const exportCall = fetchMock.mock.calls.find(([url]) => url === '/api/root/events/1/tournaments/export/excel');
    expect(exportCall).toBeTruthy();
    expect(createObjectURLMock).toHaveBeenCalled();
    expect(anchorClickMock).toHaveBeenCalled();
    expect(revokeObjectURLMock).toHaveBeenCalledWith('blob:mock-excel');
  });

  it('トーナメントプレビューを生成できる', async () => {
    render(Page);

    await page.getByRole('button', { name: 'トーナメントプレビューを生成' }).click();

    const previewCall = fetchMock.mock.calls.find(([url, options]) => url === '/api/root/events/1/tournaments/generate-preview' && options?.method === 'POST');
    expect(previewCall).toBeTruthy();
    await expect.element(page.getByRole('button', { name: 'プレビューをDBに保存' })).toBeInTheDocument();
  });

  it('生成したプレビューを保存できる', async () => {
    render(Page);

    await page.getByRole('button', { name: 'トーナメントプレビューを生成' }).click();
    await page.getByRole('button', { name: 'プレビューをDBに保存' }).click();

    const saveCall = fetchMock.mock.calls.find(([url, options]) => url === '/api/root/events/1/tournaments/bulk-create' && options?.method === 'POST');
    expect(saveCall).toBeTruthy();
  });

  it('盤上競技は全16クラス固定で、選手をrootから送信しない', async () => {
    const classes = Array.from({ length: 16 }, (_, index) => ({ id: index + 1, name: `C${index + 1}` }));
    fetchMock = vi.fn((url, options = {}) => {
      if (url === '/api/root/events/1/tournament-templates/board-game/classes') {
        return Promise.resolve({ ok: true, json: () => Promise.resolve(classes) });
      }
      if (url === '/api/admin/events/1/board-game-runs') {
        return Promise.resolve({ ok: true, json: () => Promise.resolve([]) });
      }
      if (url === '/api/root/events/1/tournaments') {
        return Promise.resolve({ ok: true, json: () => Promise.resolve([]) });
      }
      if (url === '/api/root/events/1/tournament-templates/board-game/run' && options.method === 'POST') {
        return Promise.resolve({ ok: true, json: () => Promise.resolve({ name: '将棋' }) });
      }
      return Promise.resolve({ ok: true, json: () => Promise.resolve([]) });
    });
    vi.stubGlobal('fetch', fetchMock);

    render(Page);

    await expect.element(page.getByText('参加クラス（全16クラス固定）')).not.toBeInTheDocument();
    await expect.element(page.getByText('代表選手・補欠', { exact: true })).not.toBeInTheDocument();
    await page.getByRole('button', { name: '盤上競技トーナメントを作成' }).click();

    const createCall = fetchMock.mock.calls.find(([url, options]) =>
      url === '/api/root/events/1/tournament-templates/board-game/run' && options?.method === 'POST'
    );
    expect(createCall).toBeTruthy();
    const body = JSON.parse(createCall[1].body);
    expect(body).not.toHaveProperty('participants');
    expect(body.seed_orders.A).toHaveLength(16);
    expect(body.seed_orders.B).toHaveLength(16);
  });

  it('ドラッグ＆ドロップで変更したシード順を送信する', async () => {
    const classes = Array.from({ length: 16 }, (_, index) => ({ id: index + 1, name: `C${index + 1}` }));
    fetchMock = vi.fn((url, options = {}) => {
      if (url === '/api/root/events/1/tournament-templates/board-game/classes') {
        return Promise.resolve({ ok: true, json: () => Promise.resolve(classes) });
      }
      if (url === '/api/admin/events/1/board-game-runs') {
        return Promise.resolve({ ok: true, json: () => Promise.resolve([]) });
      }
      if (url === '/api/root/events/1/tournaments') {
        return Promise.resolve({ ok: true, json: () => Promise.resolve([]) });
      }
      if (url === '/api/root/events/1/tournament-templates/board-game/run' && options.method === 'POST') {
        return Promise.resolve({ ok: true, json: () => Promise.resolve({ name: '将棋' }) });
      }
      return Promise.resolve({ ok: true, json: () => Promise.resolve([]) });
    });
    vi.stubGlobal('fetch', fetchMock);

    render(Page);

    await expect.element(page.getByText('クラスをドラッグ＆ドロップしてシード順を変更できます。')).toBeInTheDocument();
    const seedList = document.querySelector('ol[aria-label="Aブロックのシード順"]');
    const reorderedItems = [2, 1, ...Array.from({ length: 14 }, (_, index) => index + 3)]
      .map((classID) => ({ id: `A-${classID}`, classID }));
    seedList.dispatchEvent(new CustomEvent('finalize', { detail: { items: reorderedItems } }));

    await page.getByRole('button', { name: '盤上競技トーナメントを作成' }).click();

    const createCall = fetchMock.mock.calls.find(([url, options]) =>
      url === '/api/root/events/1/tournament-templates/board-game/run' && options?.method === 'POST'
    );
    const body = JSON.parse(createCall[1].body);
    expect(body.seed_orders.A).toEqual([2, 1, ...Array.from({ length: 14 }, (_, index) => index + 3)]);
    expect(body.seed_orders.B).toEqual(Array.from({ length: 16 }, (_, index) => index + 1));
  });
});
