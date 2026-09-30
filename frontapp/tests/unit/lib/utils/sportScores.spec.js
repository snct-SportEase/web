import { describe, expect, it } from 'vitest';
import { sportScoreGroups, sportScoreItems, scoreItemValue } from '$lib/utils/sportScores.js';

describe('sport score breakdown', () => {
  const score = { sport_scores: [
    { sport_id: 10, sport_name: 'サッカー', win1_points: 10, champion_points: 80 },
    { sport_id: 11, sport_name: 'キックベース', win1_points: 20 },
    { sport_id: 12, sport_name: 'オセロ・将棋', is_board_game: true, win_points: 7, rank_points: 60 }
  ] };

  it('同じ場所の競技を競技名とIDで分け、各クラス自身の得点を表示する', () => {
    const items = sportScoreItems(score);
    expect(items.find((item) => item.label === 'サッカー1勝点').key).not.toBe(
      items.find((item) => item.label === 'キックベース1勝点').key
    );
    expect(scoreItemValue(score, items.find((item) => item.label === 'キックベース1勝点'))).toBe(20);
    expect(scoreItemValue({ sport_scores: [{ sport_id: 11, win1_points: 0 }] }, items.find((item) => item.label === 'キックベース1勝点'))).toBe(0);
    expect(sportScoreGroups(score).map((group) => group.label)).toEqual(['サッカー', 'キックベース', 'オセロ・将棋']);
  });

  it('盤上競技は設定された勝利点・順位点を表示する', () => {
    const items = sportScoreItems(score).filter((item) => item.sportId === 12);
    expect(items.map((item) => item.label)).toEqual(['オセロ・将棋勝利点', 'オセロ・将棋順位点']);
    expect(items.map((item) => scoreItemValue(score, item))).toEqual([7, 60]);
  });

  it('旧APIでは固定項目にフォールバックし、新APIの空配列は空のまま扱う', () => {
    expect(sportScoreItems({ sport_names: { ground: 'サッカー' } })).toBeNull();
    expect(scoreItemValue({ ground_win1_points: 10 }, { key: 'ground_win1_points' })).toBe(10);
    expect(sportScoreItems({ sport_scores: [] })).toEqual([]);
  });
});
