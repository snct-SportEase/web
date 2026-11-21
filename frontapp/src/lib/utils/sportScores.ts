type ScoreField = 'win1_points' | 'win2_points' | 'win3_points' | 'champion_points' | 'loser_bracket_champion_points' | 'win_points' | 'rank_points';
export type SportScore = Partial<Record<ScoreField, number>> & {
  sport_id: number;
  sport_name: string;
  tournament_id?: number;
  slot_key?: string;
  is_board_game?: boolean;
};
export type ClassScore = { sport_scores?: SportScore[]; [key: string]: number | string | SportScore[] | Record<string, string> | null | undefined };
type ScoreItem = { key: string; label: string; field: ScoreField; sportId: number; tournamentId: number; value: number };
type LegacyScoreItem = { key: string; sportId?: undefined };
type ScoreGroup = { location: string; label: string; items: ScoreItem[] };

const tournamentFields: [ScoreField, string][] = [
  ['win1_points', '1勝点'],
  ['win2_points', '2勝点'],
  ['win3_points', '3勝点'],
  ['champion_points', '優勝点'],
  ['loser_bracket_champion_points', '敗者戦ブロック優勝']
];
const boardGameFields: [ScoreField, string][] = [
  ['win_points', '勝利点'],
  ['rank_points', '順位点']
];

export function sportScoreGroups(score: ClassScore | null | undefined): ScoreGroup[] | null {
  if (!Array.isArray(score?.sport_scores)) return null;
  return score.sport_scores.map((sport) => {
    const tournamentId = Number(sport.tournament_id) || 0;
    const blockLabel =
      sport.is_board_game && sport.slot_key && sport.slot_key !== 'MAIN'
        ? ` ${sport.slot_key}ブロック`
        : '';
    return {
      location: `sport_${sport.sport_id}${tournamentId ? `_tournament_${tournamentId}` : ''}`,
      label: `${sport.sport_name}${blockLabel}`,
      items: (sport.is_board_game ? boardGameFields : tournamentFields).map(([field, label]) => ({
        key: `sport_${sport.sport_id}${tournamentId ? `_tournament_${tournamentId}` : ''}_${field}`,
        label,
        field,
        sportId: sport.sport_id,
        tournamentId,
        value: Number(sport[field]) || 0
      }))
    };
  });
}

export function sportScoreItems(score: ClassScore | null | undefined): ScoreItem[] | null {
  return sportScoreGroups(score)?.flatMap((group) =>
    group.items.map((item) => ({ ...item, label: `${group.label}${item.label}` }))
  ) ?? null;
}

export function scoreItemValue(score: ClassScore, item: ScoreItem | LegacyScoreItem): number {
  if (item.sportId !== undefined) {
    return (
      score.sport_scores?.find(
        (sport) =>
          sport.sport_id === item.sportId &&
          (Number(sport.tournament_id) || 0) === (item.tournamentId || 0)
      )?.[item.field] || 0
    );
  }
  const value = score[item.key];
  return typeof value === 'number' ? value : 0;
}
