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
  let groups;
  const classes = [
    { id: 1, name: '1A' },
    { id: 2, name: '1B' },
    { id: 3, name: '1-1' },
    { id: 4, name: '1-2' },
    { id: 5, name: 'IE2' }
  ];

  beforeEach(() => {
    session = null;
    groups = [];

    fetchMock = vi.fn((url, options = {}) => {
      if (url === '/api/root/events/1/noon-game/session' && !options.method) {
        return Promise.resolve({
          ok: true,
          json: () => Promise.resolve({
            session,
            classes,
            groups,
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
            groups,
            matches: [],
            points_summary: [],
            template_runs: []
          })
        });
      }

      if (url === '/api/root/noon-game/sessions/1/groups/10' && options.method === 'PUT') {
        const body = JSON.parse(options.body);
        const updatedGroup = {
          id: 10,
          name: body.name,
          description: body.description,
          members: body.class_ids.map((classId, index) => ({
            id: index + 1,
            group_id: 10,
            class_id: classId,
            weight: 1,
            class: classes.find((cls) => cls.id === classId)
          }))
        };
        groups = groups.map((group) => (group.id === 10 ? updatedGroup : group));
        return Promise.resolve({
          ok: true,
          json: () => Promise.resolve({ group: updatedGroup })
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
    expect(JSON.parse(createCall[1].body)).toEqual(expect.objectContaining({
      session: expect.objectContaining({
        name: 'コース対抗リレー'
      })
    }));
  });

  it('グループ編集でクラスをクリックで複数選択と解除できる', async () => {
    session = {
      id: 1,
      name: '昼休み競技 2025',
      mode: 'mixed'
    };
    groups = [
      {
        id: 10,
        name: '機械系',
        description: '既存グループ',
        members: [
          { id: 1, group_id: 10, class_id: 1, weight: 1, class: classes[0] }
        ]
      }
    ];

    render(Page);

    await expect.element(page.getByRole('heading', { name: '現在の昼競技設定' })).toBeInTheDocument();
    await expect.element(page.getByText('機械系')).toBeInTheDocument();
    await page.getByRole('button', { name: '編集' }).click();

    const class1Button = page.getByRole('button', { name: '1A' });
    const class2Button = page.getByRole('button', { name: '1B' });

    await class2Button.click();
    await class1Button.click();

    await page.getByRole('button', { name: 'グループを更新' }).click();

    const updateCall = fetchMock.mock.calls.find(([url, options]) => url === '/api/root/noon-game/sessions/1/groups/10' && options?.method === 'PUT');
    expect(updateCall).toBeTruthy();
    expect(JSON.parse(updateCall[1].body)).toEqual(expect.objectContaining({
      class_ids: [2]
    }));
  });

  it('自動命名グループは所属クラスの変更に合わせて名前も更新される', async () => {
    session = {
      id: 1,
      name: '昼休み競技 2025',
      mode: 'mixed'
    };
    groups = [
      {
        id: 10,
        name: '1-1 & IEコース',
        description: '既存グループ',
        members: [
          { id: 1, group_id: 10, class_id: 3, weight: 1, class: classes[2] },
          { id: 2, group_id: 10, class_id: 5, weight: 1, class: classes[4] }
        ]
      }
    ];

    render(Page);

    await expect.element(page.getByText('1-1 & IEコース')).toBeInTheDocument();
    await page.getByRole('button', { name: '編集' }).click();
    await expect.element(page.getByLabelText('グループ名')).toHaveValue('1-1 & IEコース');

    await page.getByRole('button', { name: '1-2' }).click();
    await page.getByRole('button', { name: '1-1' }).click();

    await expect.element(page.getByLabelText('グループ名')).toHaveValue('1-2 & IEコース');
    await page.getByRole('button', { name: 'グループを更新' }).click();

    const updateCall = fetchMock.mock.calls.find(([url, options]) => url === '/api/root/noon-game/sessions/1/groups/10' && options?.method === 'PUT');
    expect(updateCall).toBeTruthy();
    expect(JSON.parse(updateCall[1].body)).toEqual(expect.objectContaining({
      name: '1-2 & IEコース',
      class_ids: [5, 4]
    }));
  });
});
