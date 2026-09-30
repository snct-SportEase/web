import { page } from '@vitest/browser/context';
import { describe, expect, it } from 'vitest';
import { render } from 'vitest-browser-svelte';
import Page from '$src/routes/dashboard/student/score-list/+page.svelte';

describe('得点一覧の競技別内訳', () => {
  it('その他の複数競技と将棋のA・Bブロック得点をクラスごとに表示する', async () => {
    render(Page, { data: { scores: [{
      class_id: 1, class_name: '1A', season: 'spring', rank_current_event: 1, total_points_current_event: 120,
      mic_points: 3,
      sport_scores: [
        { sport_id: 10, sport_name: 'サッカー', win1_points: 10, champion_points: 80 },
        { sport_id: 11, sport_name: 'キックベース', win1_points: 20 },
        { sport_id: 12, sport_name: '将棋', tournament_id: 101, slot_key: 'A', is_board_game: true, win_points: 7 },
        { sport_id: 12, sport_name: '将棋', tournament_id: 102, slot_key: 'B', is_board_game: true, win_points: 14 }
      ]
    }] } });
    await page.getByText('点数項目を表示').click();
    await expect.element(page.getByText('MIC点:')).toBeVisible();
    await expect.element(page.getByText('3', { exact: true })).toBeVisible();
    await expect.element(page.getByText('サッカー1勝点:')).toBeVisible();
    await expect.element(page.getByText('キックベース1勝点:')).toBeVisible();
    await expect.element(page.getByText('将棋 Aブロック勝利点:')).toBeVisible();
    await expect.element(page.getByText('将棋 Aブロック順位点:')).toBeVisible();
    await expect.element(page.getByText('将棋 Bブロック勝利点:')).toBeVisible();
    await expect.element(page.getByText('将棋 Bブロック順位点:')).toBeVisible();
    await expect.element(page.getByText('gym1')).not.toBeInTheDocument();
  });
});
