import { describe, expect, it } from 'vitest';
import { sportScoreGroups, sportScoreItems, scoreItemValue } from '$lib/utils/sportScores.js';

describe('sport score breakdown', () => {
  const score = { sport_scores: [
    { sport_id: 10, sport_name: 'サッカー', win1_points: 10, champion_points: 80 },
    { sport_id: 11, sport_name: 'キックベース', win1_points: 20 },
    { sport_id: 12, sport_name: '将棋', tournament_id: 101, slot_key: 'A', is_board_game: true, win_points: 7, rank_points: 60 },
    { sport_id: 12, sport_name: '将棋', tournament_id: 102, slot_key: 'B', is_board_game: true, win_points: 14, rank_points: 40 }
  ] };

  it('同じ場所の競技を競技名とIDで分け、各クラス自身の得点を表示する', () => {
    const items = sportScoreItems(score);
    expect(items.find((item) => item.label === 'サッカー1勝点').key).not.toBe(
      items.find((item) => item.label === 'キックベース1勝点').key
    );
    expect(scoreItemValue(score, items.find((item) => item.label === 'キックベース1勝点'))).toBe(20);
    expect(scoreItemValue({ sport_scores: [{ sport_id: 11, win1_points: 0 }] }, items.find((item) => item.label === 'キックベース1勝点'))).toBe(0);
    expect(sportScoreGroups(score).map((group) => group.label)).toEqual(['サッカー', 'キックベース', '将棋 Aブロック', '将棋 Bブロック']);
  });

  it('将棋はA・Bブロック別に勝利点・順位点を表示する', () => {
    const items = sportScoreItems(score).filter((item) => item.sportId === 12);
    expect(items.map((item) => item.label)).toEqual([
      '将棋 Aブロック勝利点',
      '将棋 Aブロック順位点',
      '将棋 Bブロック勝利点',
      '将棋 Bブロック順位点'
    ]);
    expect(new Set(items.map((item) => item.key)).size).toBe(4);
    expect(items.map((item) => scoreItemValue(score, item))).toEqual([7, 60, 14, 40]);
  });

  it('MAIN枠の盤上競技にはブロック名を追加しない', () => {
    const othello = { sport_scores: [
      { sport_id: 13, sport_name: 'オセロ', tournament_id: 103, slot_key: 'MAIN', is_board_game: true }
    ] };
    expect(sportScoreGroups(othello)[0].label).toBe('オセロ');
  });

  it('旧APIでは固定項目にフォールバックし、新APIの空配列は空のまま扱う', () => {
    expect(sportScoreItems({ sport_names: { ground: 'サッカー' } })).toBeNull();
    expect(scoreItemValue({ ground_win1_points: 10 }, { key: 'ground_win1_points' })).toBe(10);
    expect(sportScoreItems({ sport_scores: [] })).toEqual([]);
  });
});
