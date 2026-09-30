const tournamentFields = [
  ['win1_points', '1勝点'],
  ['win2_points', '2勝点'],
  ['win3_points', '3勝点'],
  ['champion_points', '優勝点'],
  ['loser_bracket_champion_points', '敗者戦ブロック優勝']
];

export function sportScoreGroups(score) {
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
      items: (sport.is_board_game
        ? [
            ['win_points', '勝利点'],
            ['rank_points', '順位点']
          ]
        : tournamentFields
      ).map(([field, label]) => ({
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

export function sportScoreItems(score) {
  return sportScoreGroups(score)?.flatMap((group) =>
    group.items.map((item) => ({ ...item, label: `${group.label}${item.label}` }))
  ) ?? null;
}

export function scoreItemValue(score, item) {
  if (item.sportId !== undefined) {
    return (
      score.sport_scores?.find(
        (sport) =>
          sport.sport_id === item.sportId &&
          (Number(sport.tournament_id) || 0) === (item.tournamentId || 0)
      )?.[item.field] || 0
    );
  }
  return score[item.key] || 0;
}
