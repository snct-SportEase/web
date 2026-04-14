import { page } from '@vitest/browser/context';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { render } from 'vitest-browser-svelte';
import Page from './+page.svelte';

describe('Class Student Count Page', () => {
  let fetchMock;

  beforeEach(() => {
    const classes = [
      { id: 1, name: '1A', student_count: 40 },
      { id: 2, name: '1B', student_count: 38 }
    ];

    fetchMock = vi.fn((url, options = {}) => {
      if (url === '/api/root/classes/student-counts' && options.method === 'PUT') {
        return Promise.resolve({ ok: true, json: () => Promise.resolve({ message: 'updated' }) });
      }
      if (url === '/api/root/classes/student-counts/csv' && options.method === 'POST') {
        return Promise.resolve({ ok: true, json: () => Promise.resolve({ message: 'csv updated' }) });
      }
      if (url === '/api/classes') {
        return Promise.resolve({ ok: true, json: () => Promise.resolve(classes) });
      }
      return Promise.resolve({ ok: true, json: () => Promise.resolve({}) });
    });

    vi.stubGlobal('fetch', fetchMock);
  });

  it('初期表示でクラス一覧を表示できる', async () => {
    render(Page, {
      props: {
        data: {
          classes: [
            { id: 1, name: '1A', student_count: 40 },
            { id: 2, name: '1B', student_count: 38 }
          ]
        }
      }
    });

    await expect.element(page.getByRole('heading', { name: '各クラス人数設定' })).toBeInTheDocument();
    await expect.element(page.getByText('1A')).toBeInTheDocument();
    await expect.element(page.getByText('1B')).toBeInTheDocument();
  });

  it('手動更新で生徒数を保存できる', async () => {
    render(Page, {
      props: {
        data: {
          classes: [
            { id: 1, name: '1A', student_count: 40 }
          ]
        }
      }
    });

    await page.getByRole('spinbutton').fill('42');
    await page.getByRole('button', { name: '保存' }).click();

    const saveCall = fetchMock.mock.calls.find(([url, options]) => url === '/api/root/classes/student-counts' && options?.method === 'PUT');
    expect(saveCall).toBeTruthy();
    expect(JSON.parse(saveCall[1].body)).toEqual([{ class_id: 1, student_count: 42 }]);
    await expect.element(page.getByText('生徒数を更新しました。')).toBeInTheDocument();
  });
});
