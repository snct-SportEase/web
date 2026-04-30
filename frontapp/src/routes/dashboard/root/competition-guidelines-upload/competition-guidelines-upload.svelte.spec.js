import { page } from '@vitest/browser/context';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { render } from 'vitest-browser-svelte';
import Page from './+page.svelte';

describe('Guide Management Page', () => {
  let fetchMock;

  beforeEach(() => {
    const events = [
      { id: 1, name: '2025春季スポーツ大会', competition_guidelines_pdf_url: null }
    ];
    const documents = [
      { id: 10, title: '会場案内', description: '当日の導線資料', pdf_url: 'https://example.com/map.pdf' }
    ];

    fetchMock = vi.fn((url, options = {}) => {
      if (url === '/api/events/active') {
        return Promise.resolve({ ok: true, json: () => Promise.resolve({ event_id: 1 }) });
      }
      if (url === '/api/root/events') {
        return Promise.resolve({ ok: true, json: () => Promise.resolve(events) });
      }
      if (url === '/api/root/guide-documents') {
        if (options.method === 'POST') {
          return Promise.resolve({ ok: true, json: () => Promise.resolve({ document: { id: 11 } }) });
        }
        return Promise.resolve({ ok: true, json: () => Promise.resolve({ documents }) });
      }
      if (url === '/api/root/guide-documents/10' && options.method === 'DELETE') {
        return Promise.resolve({ ok: true, json: () => Promise.resolve({ message: 'deleted' }) });
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
    vi.stubGlobal('confirm', vi.fn(() => true));
    vi.stubGlobal('URL', {
      createObjectURL: vi.fn(() => 'blob:preview'),
      revokeObjectURL: vi.fn()
    });
  });

  it('初期表示で資料管理画面と登録済み資料を表示できる', async () => {
    render(Page);

    await expect.element(page.getByRole('heading', { name: '資料管理' })).toBeInTheDocument();
    await expect.element(page.getByText('2025春季スポーツ大会')).toBeInTheDocument();
    await expect.element(page.getByText('会場案内')).toBeInTheDocument();
  });

  it('大会要項PDF未選択ではアップロードしない', async () => {
    render(Page);

    await expect.element(page.getByRole('button', { name: 'アップロード' })).toBeDisabled();
  });

  it('任意資料PDF未選択では登録しない', async () => {
    render(Page);

    await expect.element(page.getByRole('button', { name: '資料を登録' })).toBeDisabled();
  });
});
