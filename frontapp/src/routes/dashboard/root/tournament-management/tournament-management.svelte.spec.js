import { page } from '@vitest/browser/context';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { render } from 'vitest-browser-svelte';
import Page from './+page.svelte';

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
});
