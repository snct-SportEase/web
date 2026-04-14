import { page } from '@vitest/browser/context';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { render } from 'vitest-browser-svelte';
import Page from './+page.svelte';

vi.mock('$lib/stores/eventStore.js', () => ({
  activeEvent: {
    init: vi.fn(async () => ({
      id: 1,
      name: '2025春季スポーツ大会',
      year: 2025,
      season: 'spring',
      is_rainy_mode: false
    }))
  }
}));

describe('Rainy Mode Page', () => {
  let fetchMock;
  let alertMock;

  beforeEach(() => {
    fetchMock = vi.fn((url, options = {}) => {
      if (url === '/api/root/events/1/rainy-mode' && options.method === 'PUT') {
        const body = JSON.parse(options.body);
        return Promise.resolve({
          ok: true,
          json: () => Promise.resolve({ is_rainy_mode: body.is_rainy_mode })
        });
      }
      return Promise.resolve({ ok: true, json: () => Promise.resolve({}) });
    });

    alertMock = vi.fn();
    vi.stubGlobal('fetch', fetchMock);
    vi.stubGlobal('alert', alertMock);
  });

  it('アクティブ大会と現在の状態を表示できる', async () => {
    render(Page);

    await expect.element(page.getByRole('heading', { name: '雨天時モード管理' })).toBeInTheDocument();
    await expect.element(page.getByText('2025春季スポーツ大会')).toBeInTheDocument();
    await expect.element(page.getByText('現在の状態:')).toBeInTheDocument();
    await expect.element(page.getByText('無効')).toBeInTheDocument();
  });

  it('雨天時モードを有効化できる', async () => {
    render(Page);

    await page.getByRole('button', { name: '有効にする' }).click();

    const saveCall = fetchMock.mock.calls.find(([url, options]) => url === '/api/root/events/1/rainy-mode' && options?.method === 'PUT');
    expect(saveCall).toBeTruthy();
    expect(JSON.parse(saveCall[1].body)).toEqual({ is_rainy_mode: true });
    expect(alertMock).toHaveBeenCalled();
  });
});
