import { page } from '@vitest/browser/context';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { render } from 'vitest-browser-svelte';
import Page from './+page.svelte';

describe('Competition Guidelines Upload Page', () => {
  let fetchMock;

  beforeEach(() => {
    const events = [
      { id: 1, name: '2025春季スポーツ大会', competition_guidelines_pdf_url: null }
    ];

    fetchMock = vi.fn((url, options = {}) => {
      if (url === '/api/events/active') {
        return Promise.resolve({ ok: true, json: () => Promise.resolve({ event_id: 1 }) });
      }
      if (url === '/api/root/events') {
        return Promise.resolve({ ok: true, json: () => Promise.resolve(events) });
      }
      if (url === '/api/admin/pdfs' && options.method === 'POST') {
        return Promise.resolve({ ok: true, json: () => Promise.resolve({ url: 'https://example.com/guidelines.pdf' }) });
      }
      if (url === '/api/root/events/1/competition-guidelines' && options.method === 'PUT') {
        return Promise.resolve({ ok: true, json: () => Promise.resolve({ message: 'updated' }) });
      }
      return Promise.resolve({ ok: true, json: () => Promise.resolve({}) });
    });

    vi.stubGlobal('fetch', fetchMock);
    vi.stubGlobal('URL', {
      createObjectURL: vi.fn(() => 'blob:preview'),
      revokeObjectURL: vi.fn()
    });
  });

  it('初期表示で大会選択を表示できる', async () => {
    render(Page);

    await expect.element(page.getByRole('heading', { name: '大会要項アップロード' })).toBeInTheDocument();
    await expect.element(page.getByLabelText('大会選択')).toBeInTheDocument();
    await expect.element(page.getByText('2025春季スポーツ大会')).toBeInTheDocument();
  });

  it('PDF未選択ではアップロードしない', async () => {
    render(Page);

    await expect.element(page.getByRole('button', { name: 'アップロード' })).toBeDisabled();
  });
});
