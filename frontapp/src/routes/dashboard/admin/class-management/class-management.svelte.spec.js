import { page } from '@vitest/browser/context';
import { describe, expect, it } from 'vitest';
import { render } from 'vitest-browser-svelte';
import Page from './+page.svelte';

describe('Class Management Page', () => {
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
});
