import { page } from '@vitest/browser/context';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { render } from 'vitest-browser-svelte';
import Page from './+page.svelte';

const mocks = vi.hoisted(() => ({
  pageValue: {
    data: {
      user: {
        id: 'current-root',
        roles: [{ name: 'root' }]
      }
    }
  }
}));

vi.mock('$app/stores', () => ({
  page: {
    subscribe(fn) {
      fn(mocks.pageValue);
      return () => {};
    }
  }
}));

vi.mock('$app/navigation', () => ({
  invalidateAll: vi.fn(async () => {})
}));

describe('Change Username Page', () => {
  let fetchMock;
  let users;

  beforeEach(() => {
    users = [
      {
        id: 'user-1',
        email: 'student1@sendai-nct.jp',
        display_name: '山田太郎',
        class_id: 1,
        roles: [
          { id: 1, name: 'student' },
          { id: 2, name: '1A_rep' },
          { id: 3, name: 'judge' }
        ]
      },
      {
        id: 'user-2',
        email: 'admin1@sendai-nct.jp',
        display_name: '運営花子',
        class_id: 2,
        roles: [
          { id: 4, name: 'admin' }
        ]
      }
    ];

    fetchMock = vi.fn((url, options = {}) => {
      if (url.startsWith('/api/root/users?')) {
        return Promise.resolve({ ok: true, json: () => Promise.resolve(users) });
      }

      if (url === '/api/classes') {
        return Promise.resolve({
          ok: true,
          json: () => Promise.resolve([
            { id: 1, name: '1A' },
            { id: 2, name: '1B' }
          ])
        });
      }

      if (url === '/api/root/users/display-name' && options.method === 'PUT') {
        const body = JSON.parse(options.body);
        users = users.map((user) => user.id === body.user_id ? { ...user, display_name: body.display_name } : user);
        return Promise.resolve({ ok: true, json: () => Promise.resolve({ ok: true }) });
      }

      if (url === '/api/root/users/promote' && options.method === 'PUT') {
        const body = JSON.parse(options.body);
        const masterRoles = ['student', 'admin', 'root'];
        users = users.map((user) => {
          if (user.id !== body.user_id) return user;
          return {
            ...user,
            roles: [
              ...user.roles.filter((role) => !masterRoles.includes(role.name)),
              { id: `master-${body.role}`, name: body.role }
            ]
          };
        });
        return Promise.resolve({ ok: true, json: () => Promise.resolve({ ok: true }) });
      }

      if (url === '/api/admin/users/role' && options.method === 'PUT') {
        const body = JSON.parse(options.body);
        users = users.map((user) => {
          if (user.id !== body.user_id) return user;
          if (user.roles.some((role) => role.name === body.role)) return user;
          return {
            ...user,
            roles: [...user.roles, { id: Date.now(), name: body.role }]
          };
        });
        return Promise.resolve({ ok: true, json: () => Promise.resolve({ ok: true }) });
      }

      if (url === '/api/admin/users/role' && options.method === 'DELETE') {
        const body = JSON.parse(options.body);
        users = users.map((user) => user.id === body.user_id ? {
          ...user,
          roles: user.roles.filter((role) => role.name !== body.role)
        } : user);
        return Promise.resolve({ ok: true, json: () => Promise.resolve({ ok: true }) });
      }

      return Promise.resolve({ ok: true, json: () => Promise.resolve({}) });
    });

    vi.stubGlobal('fetch', fetchMock);
    vi.stubGlobal('alert', vi.fn());
    vi.stubGlobal('confirm', vi.fn(() => true));
  });

  it('初期表示でユーザー一覧を表示できる', async () => {
    render(Page);

    await expect.element(page.getByRole('heading', { name: 'ユーザー管理' })).toBeInTheDocument();
    await expect.element(page.getByText('student1@sendai-nct.jp')).toBeInTheDocument();
  });

  it('表示名を更新できる', async () => {
    render(Page);

    await page.getByRole('button', { name: '管理' }).first().click();
    await page.getByRole('textbox', { name: '表示名' }).fill('新しい表示名');
    await page.getByRole('button', { name: '更新' }).click();

    const updateCall = fetchMock.mock.calls.find(([url, options]) => url === '/api/root/users/display-name' && options?.method === 'PUT');
    expect(updateCall).toBeTruthy();
    expect(JSON.parse(updateCall[1].body)).toEqual({
      user_id: 'user-1',
      display_name: '新しい表示名'
    });
  });

  it('一般ロールを追加できる', async () => {
    render(Page);

    await page.getByRole('button', { name: '管理' }).first().click();
    await page.getByLabelText('新規ロール追加').fill('score_keeper');
    await page.getByRole('button', { name: '追加' }).click();

    const addRoleCall = fetchMock.mock.calls.find(([url, options]) => url === '/api/admin/users/role' && options?.method === 'PUT');
    expect(addRoleCall).toBeTruthy();
    expect(JSON.parse(addRoleCall[1].body)).toEqual({
      user_id: 'user-1',
      role: 'score_keeper'
    });
  });

  it('rootロールに切り替えると表示が更新される', async () => {
    render(Page);

    await page.getByRole('button', { name: '管理' }).first().click();
    await page.getByRole('button', { name: 'root に切り替え' }).click();

    const replaceCall = fetchMock.mock.calls.find(([url, options]) => url === '/api/root/users/promote' && options?.method === 'PUT');
    expect(replaceCall).toBeTruthy();
    expect(JSON.parse(replaceCall[1].body)).toEqual({
      user_id: 'user-1',
      role: 'root'
    });
    await expect.element(page.getByRole('row', { name: /student1@sendai-nct\.jp.*root/ })).toBeInTheDocument();
  });

  it('一般ロールを削除できる', async () => {
    render(Page);

    await page.getByRole('button', { name: '管理' }).first().click();
    await page.getByTitle('ロールを削除').first().click();

    const deleteRoleCall = fetchMock.mock.calls.find(([url, options]) => url === '/api/admin/users/role' && options?.method === 'DELETE');
    expect(deleteRoleCall).toBeTruthy();
    expect(confirm).toHaveBeenCalledWith('ロール "judge" を削除しますか？');
    expect(JSON.parse(deleteRoleCall[1].body)).toEqual({
      user_id: 'user-1',
      role: 'judge'
    });
  });

  it('masterロールはその他のロール削除対象に出ない', async () => {
    render(Page);

    await page.getByRole('button', { name: '管理' }).nth(1).click();
    await expect.element(page.getByText('admin を保有中')).toBeInTheDocument();
    await expect.element(page.getByTitle('ロールを削除')).not.toBeInTheDocument();
  });

  it('クラス所属ロールを付け替えできる', async () => {
    render(Page);

    await page.getByRole('button', { name: '管理' }).first().click();
    await page.getByLabelText('担当クラスを選択').selectOptions('2');
    await page.getByRole('button', { name: '変更・保存' }).click();

    const addRoleCall = fetchMock.mock.calls.find(([url, options]) => url === '/api/admin/users/role' && options?.method === 'PUT');
    const deleteRoleCall = fetchMock.mock.calls.find(([url, options]) => url === '/api/admin/users/role' && options?.method === 'DELETE');
    expect(addRoleCall).toBeTruthy();
    expect(deleteRoleCall).toBeTruthy();
    expect(JSON.parse(addRoleCall[1].body)).toEqual({
      user_id: 'user-1',
      role: '1B_rep'
    });
    expect(JSON.parse(deleteRoleCall[1].body)).toEqual({
      user_id: 'user-1',
      role: '1A_rep'
    });
  });
});
