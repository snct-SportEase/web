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

describe('Noon Game Page', () => {
  let fetchMock;
  let session;
  const classes = [
    { id: 1, name: '1A' },
    { id: 2, name: '1B' }
  ];

  beforeEach(() => {
    session = null;

    fetchMock = vi.fn((url, options = {}) => {
      if (url === '/api/root/events/1/noon-game/session' && !options.method) {
        return Promise.resolve({
          ok: true,
          json: () => Promise.resolve({
            session,
            classes,
            groups: [],
            matches: [],
            points_summary: [],
            template_runs: []
          })
        });
      }

      if (url === '/api/root/events/1/noon-game/session' && options.method === 'POST') {
        const body = JSON.parse(options.body);
        session = { id: 1, ...body };
        return Promise.resolve({
          ok: true,
          json: () => Promise.resolve({
            session,
            classes,
            groups: [],
            matches: [],
            points_summary: [],
            template_runs: []
          })
        });
      }

      if (url === '/api/root/noon-game/templates/course_relay/default-groups' && !options.method) {
        return Promise.resolve({
          ok: true,
          json: () => Promise.resolve({
            groups: [{ group_name: '機械系', class_names: ['1A'] }]
          })
        });
      }

      if (url === '/api/root/events/1/noon-game/templates/course-relay/run' && options.method === 'POST') {
        return Promise.resolve({ ok: true, json: () => Promise.resolve({ ok: true }) });
      }

      return Promise.resolve({ ok: true, json: () => Promise.resolve({}) });
    });

    vi.stubGlobal('fetch', fetchMock);
    vi.stubGlobal('alert', vi.fn());
  });

  it('初期表示で昼競技管理画面を表示できる', async () => {
    render(Page);

    await expect.element(page.getByRole('heading', { name: '昼競技管理' })).toBeInTheDocument();
    await expect.element(page.getByRole('button', { name: 'セッションを作成' })).toBeInTheDocument();
  });

  it('昼競技セッションを作成できる', async () => {
    render(Page);

    await page.getByLabelText('セッション名').fill('昼休み競技 2025');
    await page.getByRole('button', { name: 'セッションを作成' }).click();

    const saveCall = fetchMock.mock.calls.find(([url, options]) => url === '/api/root/events/1/noon-game/session' && options?.method === 'POST');
    expect(saveCall).toBeTruthy();
    expect(JSON.parse(saveCall[1].body)).toEqual(expect.objectContaining({
      name: '昼休み競技 2025',
      mode: 'mixed'
    }));
  });

  it('テンプレートを実行できる', async () => {
    render(Page);

    await page.getByRole('button', { name: 'テンプレートを設定' }).nth(1).click();
    await expect.element(page.getByRole('heading', { name: /コース対抗リレー.*テンプレート設定/ })).toBeInTheDocument();
    await page.getByRole('button', { name: 'テンプレートを作成' }).click();

    const createCall = fetchMock.mock.calls.find(([url, options]) => url === '/api/root/events/1/noon-game/templates/course-relay/run' && options?.method === 'POST');
    expect(createCall).toBeTruthy();
  });
});
