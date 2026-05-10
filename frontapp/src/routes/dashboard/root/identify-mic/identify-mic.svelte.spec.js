import { page } from '@vitest/browser/context';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { render } from 'vitest-browser-svelte';
import Page from './+page.svelte';

describe('Identify Mic Page', () => {
  beforeEach(() => {
    vi.stubGlobal('fetch', vi.fn((url) => {
      if (url === '/api/events/active') {
        return Promise.resolve({
          ok: true,
          json: () => Promise.resolve({ event_id: 1 })
        });
      }
      if (url === '/api/root/mic/class?event_id=1') {
        return Promise.resolve({
          ok: true,
          json: () => Promise.resolve({ class_name: '1A', vote_count: 7, total_points: 120, season: 'spring' })
        });
      }
      return Promise.resolve({ ok: true, json: () => Promise.resolve({}) });
    }));
  });

  it('行事委員会賞結果を表示できる', async () => {
    render(Page);

    await expect.element(page.getByRole('heading', { name: '行事委員会賞確認' })).toBeInTheDocument();
    await expect.element(page.getByText('行事委員会賞クラス')).toBeInTheDocument();
    await expect.element(page.getByText('1A')).toBeInTheDocument();
    await expect.element(page.getByText('7')).toBeInTheDocument();
    await expect.element(page.getByText('120')).toBeInTheDocument();
  });
});
