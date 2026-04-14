import { page } from '@vitest/browser/context';
import { beforeEach, describe, expect, it, vi } from 'vitest';
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

      return Promise.resolve({ ok: true, json: () => Promise.resolve({}) });
    });

    vi.stubGlobal('fetch', fetchMock);
    vi.stubGlobal('confirm', vi.fn(() => true));
    vi.stubGlobal('alert', vi.fn());
  });

  it('初期表示で見出しを表示できる', async () => {
    render(Page);

    await expect.element(page.getByRole('heading', { name: 'トーナメント生成・管理' })).toBeInTheDocument();
    await expect.element(page.getByRole('button', { name: 'トーナメントプレビューを生成' })).toBeInTheDocument();
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
