import { page } from '@vitest/browser/context';
import { describe, expect, it } from 'vitest';
import { render } from 'vitest-browser-svelte';
import Page from '$src/routes/dashboard/student/score-list/+page.svelte';

describe('得点一覧の競技別内訳', () => {
  it('その他の複数競技と盤上競技の得点をクラスごとに表示する', async () => {
    render(Page, { data: { scores: [{
      class_id: 1, class_name: '1A', season: 'spring', rank_current_event: 1, total_points_current_event: 117,
      sport_scores: [
        { sport_id: 10, sport_name: 'サッカー', win1_points: 10, champion_points: 80 },
        { sport_id: 11, sport_name: 'キックベース', win1_points: 20 },
        { sport_id: 12, sport_name: 'オセロ・将棋', is_board_game: true, win_points: 7 }
      ]
    }] } });
    await page.getByText('点数項目を表示').click();
    await expect.element(page.getByText('サッカー1勝点:')).toBeVisible();
    await expect.element(page.getByText('キックベース1勝点:')).toBeVisible();
    await expect.element(page.getByText('オセロ・将棋勝利点:')).toBeVisible();
    await expect.element(page.getByText('オセロ・将棋順位点:')).toBeVisible();
    await expect.element(page.getByText('gym1')).not.toBeInTheDocument();
  });
});
