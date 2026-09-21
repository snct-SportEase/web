import { page } from '@vitest/browser/context';
import { afterEach, describe, expect, it, vi } from 'vitest';
import { render } from 'vitest-browser-svelte';
import Page from '$src/routes/dashboard/admin/class-management/+page.svelte';

describe('Class Management Page', () => {
  afterEach(() => vi.unstubAllGlobals());

  it.each([null, []])('空のAPI応答 %j でもクラス・競技の切り替えを続けられる', async (emptyMembers) => {
    const fetchMock = vi.fn(async () => ({ ok: true, json: async () => emptyMembers }));
    vi.stubGlobal('fetch', fetchMock);
    render(Page, {
      props: {
        data: {
          classes: [{ id: 1, name: 'IS4' }, { id: 2, name: 'IT4' }],
          classMembers: [{ id: 'member', email: 'member@example.com', display_name: '登録済みメンバー' }],
          allSports: [{ id: 1, name: '綱引き' }],
          availableSports: [{ id: 1, name: '綱引き' }],
          selectedClassId: 1,
          isAdmin: true
        }
      }
    });

    await expect.element(page.getByText('登録済みメンバー', { exact: true })).toBeInTheDocument();
    await page.getByRole('combobox').nth(0).selectOptions('2');
    await expect.element(page.getByText('メンバーが登録されていません')).toBeInTheDocument();
    await expect.element(page.getByText('登録済みメンバー', { exact: true })).not.toBeInTheDocument();
    await page.getByRole('combobox').nth(1).selectOptions('1');
    await expect.element(page.getByText('メンバーが割り当てられていません')).toBeInTheDocument();
    expect(fetchMock.mock.calls.map(([url]) => url)).toEqual([
      '/api/admin/class-team/classes/2/members',
      '/api/admin/class-team/sports/1/members?class_id=2'
    ]);
  });

  it('作成済みの昼競技に対応する競技マスタを選択肢に表示する', async () => {
    render(Page, {
      props: {
        data: {
          classes: [{ id: 1, name: '1A' }],
          classMembers: [],
          allSports: [
            { id: 1, name: 'バスケットボール' },
            { id: 2, name: '綱引き' }
          ],
          availableSports: [
            { id: 1, name: 'バスケットボール' },
            { id: 2, name: '綱引き' }
          ],
          selectedClassId: 1,
          noonSessionName: '綱引き',
          noonSessionSportMatched: true,
          isAdmin: true
        }
      }
    });

    await expect.element(page.getByRole('heading', { name: 'クラス競技割り当て・管理' })).toBeInTheDocument();
    await expect.element(page.getByRole('option', { name: '綱引き' })).toBeInTheDocument();
  });

  it('昼競技セッション名に対応する競技マスタがない場合は案内を表示する', async () => {
    render(Page, {
      props: {
        data: {
          classes: [{ id: 1, name: '1A' }],
          classMembers: [],
          allSports: [{ id: 1, name: 'バスケットボール' }],
          availableSports: [{ id: 1, name: 'バスケットボール' }],
          selectedClassId: 1,
          noonSessionName: '綱引き',
          noonSessionSportMatched: false,
          isAdmin: true
        }
      }
    });

    await expect
      .element(page.getByText('昼競技セッション「綱引き」は競技マスタに同名の競技がないため、割り当て候補に表示できません。'))
      .toBeInTheDocument();
  });

  it('将棋の代表A・代表B・補欠を選択順で割り当てられる', async () => {
    const fetchMock = vi.fn(async (url, options = {}) => {
      if (url.includes('/members') && !options.method) {
        return { ok: true, json: async () => [] };
      }
      return { ok: true, json: async () => ({ message: '登録しました' }) };
    });
    vi.stubGlobal('fetch', fetchMock);
    render(Page, {
      props: {
        data: {
          classes: [{ id: 1, name: '1A' }],
          classMembers: [
            { id: 'a', email: 'a@example.com', display_name: '選手A' },
            { id: 'b', email: 'b@example.com', display_name: '選手B' },
            { id: 'sub', email: 'sub@example.com', display_name: '選手C' }
          ],
          allSports: [{ id: 9, name: '将棋', templateKey: 'board_game_tournament', gameType: 'shogi' }],
          availableSports: [{ id: 9, name: '将棋', templateKey: 'board_game_tournament', gameType: 'shogi' }],
          selectedClassId: 1,
          isAdmin: true
        }
      }
    });

    await page.getByRole('combobox').nth(1).selectOptions('9');
    await expect.element(page.getByText('選択順に「代表A・代表B・補欠」として登録します（3名まで）。')).toBeInTheDocument();
    for (const index of [0, 1, 2]) {
      await page.getByRole('checkbox').nth(index).click();
    }
    await expect.element(page.getByText('代表A', { exact: true })).toBeInTheDocument();
    await expect.element(page.getByText('代表B', { exact: true })).toBeInTheDocument();
    await expect.element(page.getByText('補欠', { exact: true })).toBeInTheDocument();
    await page.getByRole('button', { name: '選択した3名を選択順で割り当てる' }).click();

    const assignCall = fetchMock.mock.calls.find(([url, options]) =>
      url === '/api/admin/class-team/assign-members' && options?.method === 'POST'
    );
    expect(assignCall).toBeTruthy();
    expect(JSON.parse(assignCall[1].body).user_ids).toEqual(['a', 'b', 'sub']);
  });
});
