import { page } from '@vitest/browser/context';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { render } from 'vitest-browser-svelte';
import Page from './+page.svelte';

describe('Class Management Page', () => {
  let fetchMock;
  let confirmMock;
  let teamMembers;

  beforeEach(() => {
    teamMembers = [];

    fetchMock = vi.fn((url, options = {}) => {
      if (url === '/api/admin/class-team/sports/1/members') {
        return Promise.resolve({
          ok: true,
          json: () => Promise.resolve(teamMembers)
        });
      }

      if (url === '/api/admin/class-team/assign-members' && options.method === 'POST') {
        const body = JSON.parse(options.body);
        teamMembers = [
          ...teamMembers,
          { id: body.user_ids[0], display_name: '山田太郎', email: 'student1@sendai-nct.jp' }
        ];

        return Promise.resolve({
          ok: true,
          json: () => Promise.resolve({ message: 'メンバーの割り当てが完了しました' })
        });
      }

      if (url === '/api/admin/class-team/remove-member' && options.method === 'DELETE') {
        const body = JSON.parse(options.body);
        teamMembers = teamMembers.filter((member) => member.id !== body.user_id);

        return Promise.resolve({
          ok: true,
          json: () => Promise.resolve({ message: 'メンバーを削除しました' })
        });
      }

      if (url === '/api/admin/class-team/classes/1/members') {
        return Promise.resolve({
          ok: true,
          json: () =>
            Promise.resolve([
              { id: 'user-1', display_name: '山田太郎', email: 'student1@sendai-nct.jp' },
              { id: 'user-3', display_name: '佐藤花子', email: 'student3@sendai-nct.jp' }
            ])
        });
      }

      return Promise.resolve({
        ok: true,
        json: () => Promise.resolve({})
      });
    });

    confirmMock = vi.fn(() => true);

    vi.stubGlobal('fetch', fetchMock);
    vi.stubGlobal('confirm', confirmMock);
  });

  afterEach(() => {
    vi.unstubAllGlobals();
  });

  it('選択したメンバーを競技チームへ割り当てできる', async () => {
    render(Page, {
      props: {
        data: {
          classes: [{ id: 1, name: '1A' }],
          classMembers: [
            { id: 'user-1', display_name: '山田太郎', email: 'student1@sendai-nct.jp' },
            { id: 'user-3', display_name: '佐藤花子', email: 'student3@sendai-nct.jp' }
          ],
          eventSports: [{ id: 11, sport_id: 1 }],
          allSports: [{ id: 1, name: 'バスケットボール' }],
          selectedClassId: 1,
          isAdmin: false,
          error: null
        }
      }
    });

    await expect.element(page.getByRole('heading', { name: 'クラス・チーム管理' })).toBeInTheDocument();
    await page.getByRole('combobox').selectOptions('1');
    await page.getByRole('checkbox').first().click();
    await page.getByRole('button', { name: '選択した1名を割り当てる' }).click();

    const assignCall = fetchMock.mock.calls.find(([url, options]) => {
      return url === '/api/admin/class-team/assign-members' && options?.method === 'POST';
    });

    expect(assignCall).toBeTruthy();
    expect(JSON.parse(assignCall[1].body)).toEqual({
      sport_id: 1,
      user_ids: ['user-1']
    });

    await expect.element(page.getByText('メンバーの割り当てが完了しました')).toBeInTheDocument();
    await expect.element(page.getByText('割り当て済みメンバー (バスケットボール)')).toBeInTheDocument();
    await expect.element(page.getByRole('cell', { name: '山田太郎' }).nth(1)).toBeInTheDocument();
  });

  it('割り当て済みメンバーを削除できる', async () => {
    teamMembers = [{ id: 'user-1', display_name: '山田太郎', email: 'student1@sendai-nct.jp' }];

    render(Page, {
      props: {
        data: {
          classes: [{ id: 1, name: '1A' }],
          classMembers: [{ id: 'user-1', display_name: '山田太郎', email: 'student1@sendai-nct.jp' }],
          eventSports: [{ id: 11, sport_id: 1 }],
          allSports: [{ id: 1, name: 'バスケットボール' }],
          selectedClassId: 1,
          isAdmin: false,
          error: null
        }
      }
    });

    await page.getByRole('combobox').selectOptions('1');
    await expect.element(page.getByRole('cell', { name: '山田太郎' }).first()).toBeInTheDocument();
    await page.getByRole('button', { name: '削除' }).click();

    const removeCall = fetchMock.mock.calls.find(([url, options]) => {
      return url === '/api/admin/class-team/remove-member' && options?.method === 'DELETE';
    });

    expect(confirmMock).toHaveBeenCalled();
    expect(removeCall).toBeTruthy();
    expect(JSON.parse(removeCall[1].body)).toEqual({
      sport_id: 1,
      user_id: 'user-1'
    });

    await expect.element(page.getByText('メンバーを削除しました')).toBeInTheDocument();
    await expect.element(page.getByText('メンバーが割り当てられていません')).toBeInTheDocument();
  });
});
