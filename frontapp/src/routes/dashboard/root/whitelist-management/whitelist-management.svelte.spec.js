import { page } from '@vitest/browser/context';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { render } from 'vitest-browser-svelte';
import Page from './+page.svelte';

const mocks = vi.hoisted(() => ({
  pageData: {}
}));

vi.mock('$app/stores', () => ({
  page: {
    subscribe(fn) {
      fn({ data: mocks.pageData });
      return () => {};
    }
  }
}));

describe('Whitelist Management Page', () => {
  let fetchMock;
  let whitelist;

  beforeEach(() => {
    whitelist = [
      { id: 1, email: 'student1@sendai-nct.jp', role: 'student' },
      { id: 2, email: 'admin1@sendai-nct.jp', role: 'admin' }
    ];

    mocks.pageData = {
      whitelist,
      error: null
    };

    fetchMock = vi.fn((url, options = {}) => {
      if (url === '/api/root/whitelist' && !options.method) {
        return Promise.resolve({ ok: true, json: () => Promise.resolve(whitelist) });
      }

      if (url === '/api/root/whitelist' && options.method === 'POST') {
        const body = JSON.parse(options.body);
        whitelist = [...whitelist, { id: whitelist.length + 1, ...body }];
        return Promise.resolve({ ok: true, json: () => Promise.resolve({ ok: true }) });
      }

      if (url === '/api/root/whitelist/bulk' && options.method === 'DELETE') {
        const body = JSON.parse(options.body);
        whitelist = whitelist.filter((entry) => !body.emails.includes(entry.email));
        return Promise.resolve({ ok: true, json: () => Promise.resolve({ ok: true }) });
      }

      return Promise.resolve({ ok: true, json: () => Promise.resolve({}) });
    });

    vi.stubGlobal('fetch', fetchMock);
    vi.stubGlobal('confirm', vi.fn(() => true));
  });

  it('初期表示でホワイトリスト一覧を表示できる', async () => {
    render(Page);

    await expect.element(page.getByRole('heading', { name: 'ホワイトリスト管理' })).toBeInTheDocument();
    await expect.element(page.getByText('student1@sendai-nct.jp')).toBeInTheDocument();
    await expect.element(page.getByText('admin1@sendai-nct.jp')).toBeInTheDocument();
  });

  it('メールアドレスを追加できる', async () => {
    render(Page);

    await page.getByLabelText('メールアドレス').fill('new.user');
    await page.getByLabelText('Role', { exact: true }).selectOptions('admin');
    await page.getByRole('button', { name: 'Add' }).click();

    const postCall = fetchMock.mock.calls.find(([url, options]) => url === '/api/root/whitelist' && options?.method === 'POST');
    expect(postCall).toBeTruthy();
    expect(JSON.parse(postCall[1].body)).toEqual({
      email: 'new.user@sendai-nct.jp',
      role: 'admin'
    });
  });

  it('選択した項目を一括削除できる', async () => {
    render(Page);

    await page.getByRole('checkbox').nth(1).click();
    await page.getByRole('button', { name: '選択した項目を削除' }).click();

    const deleteCall = fetchMock.mock.calls.find(([url, options]) => url === '/api/root/whitelist/bulk' && options?.method === 'DELETE');
    expect(deleteCall).toBeTruthy();
    expect(JSON.parse(deleteCall[1].body)).toEqual({
      emails: ['admin1@sendai-nct.jp']
    });
  });
});
