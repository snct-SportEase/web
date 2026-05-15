import { env } from '$env/dynamic/private';
const BACKEND_URL = env.BACKEND_URL;

const toNumber = (value) => (typeof value === 'number' && !Number.isNaN(value) ? value : 0);

const toDateValue = (value) => {
	if (!value) return Number.POSITIVE_INFINITY;
	const date = new Date(String(value).replace(' ', 'T'));
	return Number.isNaN(date.getTime()) ? Number.POSITIVE_INFINITY : date.getTime();
};

const normalizeMatchStatus = (status) => String(status || '').toLowerCase();

const getContestantName = (tournamentData, side) => {
	if (!side) return '未定';
	if (side.title) return side.title;
	const contestant = tournamentData?.contestants?.[side.contestantId];
	return contestant?.players?.[0]?.title || '未定';
};

const getRoundLabel = (tournamentData, match) =>
	tournamentData?.rounds?.[match?.roundIndex]?.name || `Round ${toNumber(match?.roundIndex) + 1}`;

const buildTournamentUpcomingMatches = (tournaments, teams) => {
	if (!Array.isArray(tournaments) || !Array.isArray(teams) || teams.length === 0) {
		return [];
	}

	const teamByID = new Map(teams.map((team) => [Number(team.id), team]));
	const upcoming = [];

	for (const tournament of tournaments) {
		let tournamentData = tournament?.data;
		if (typeof tournamentData === 'string') {
			try {
				tournamentData = JSON.parse(tournamentData);
			} catch {
				continue;
			}
		}

		if (!tournamentData?.matches) continue;

		for (const match of tournamentData.matches) {
			const status = normalizeMatchStatus(match?.matchStatus);
			if (status === 'completed' || status === 'finished') continue;

			const sides = Array.isArray(match?.sides) ? match.sides : [];
			const mySide = sides.find((side) => teamByID.has(Number(side?.teamId)));
			if (!mySide) continue;

			const team = teamByID.get(Number(mySide.teamId));
			const opponentSide = sides.find((side) => Number(side?.teamId) !== Number(mySide.teamId));
			const startTime = match?.startTime || match?.rainyModeStartTime;

			upcoming.push({
				id: `tournament-${tournament.id}-${match.id}`,
				start_time: startTime || null,
				sport_name: team?.sport_name || tournament?.name || '競技',
				opponent_name: getContestantName(tournamentData, opponentSide),
				location: getRoundLabel(tournamentData, match),
				sort_value: toDateValue(startTime),
				status
			});
		}
	}

	return upcoming;
};

const getParticipantEntries = (match, classId) =>
	(match?.entries || []).filter((entry) =>
		Array.isArray(entry?.class_ids) && entry.class_ids.some((id) => Number(id) === Number(classId))
	);

const getOpponentName = (match, classId) => {
	const participantEntries = getParticipantEntries(match, classId);
	if (participantEntries.length === 0) return '';

	const participantIds = new Set(participantEntries.map((entry) => String(entry.id)));
	const opponent = (match?.entries || []).find((entry) => !participantIds.has(String(entry.id)));
	return opponent?.resolved_name || opponent?.display_name || '未定';
};

const buildNoonUpcomingMatches = (sessionPayload, classId) => {
	if (!classId || !sessionPayload?.matches) return [];

	return sessionPayload.matches
		.filter((match) => getParticipantEntries(match, classId).length > 0)
		.filter((match) => {
			const status = normalizeMatchStatus(match?.status);
			return status !== 'finished' && status !== 'completed' && !match?.result;
		})
		.map((match) => ({
			id: `noon-${match.id}`,
			start_time: match?.scheduled_at || null,
			sport_name: '昼競技',
			opponent_name: getOpponentName(match, classId),
			location: match?.location || sessionPayload?.session?.name || '昼競技',
			sort_value: toDateValue(match?.scheduled_at),
			status: normalizeMatchStatus(match?.status)
		}));
};

const buildAssignedSports = (teams) => {
	if (!Array.isArray(teams)) return [];

	const seen = new Set();
	return teams.filter((team) => {
		const key = `${team?.sport_id}-${team?.sport_name}-${team?.name}`;
		if (seen.has(key)) return false;
		seen.add(key);
		return true;
	});
};

const buildScoreBreakdown = (classScore) => {
	if (!classScore) {
		return {
			scoreItems: [],
			categoryBreakdown: [],
			pointHighlights: [],
			sportSections: []
		};
	}

	const sportNames = classScore.sport_names || {};

	const baseItems = [
		{ key: 'initial_points', label: '初期点', value: toNumber(classScore.initial_points) },
		{ key: 'survey_points', label: 'アンケート', value: toNumber(classScore.survey_points) },
		{ key: 'attendance_points', label: '出席', value: toNumber(classScore.attendance_points) },
		{ key: 'noon_game_points', label: '昼競技', value: toNumber(classScore.noon_game_points) }
	];

	const sportGroups = [
		{
			location: 'gym1',
			label: sportNames.gym1 || '体育館１',
			items: [
				{ key: 'gym1_win1_points', label: '1勝点' },
				{ key: 'gym1_win2_points', label: '2勝点' },
				{ key: 'gym1_win3_points', label: '3勝点' },
				{ key: 'gym1_champion_points', label: '優勝点' }
			]
		},
		{
			location: 'gym2',
			label: sportNames.gym2 || '体育館２',
			items: [
				{ key: 'gym2_win1_points', label: '1勝点' },
				{ key: 'gym2_win2_points', label: '2勝点' },
				{ key: 'gym2_win3_points', label: '3勝点' },
				{ key: 'gym2_champion_points', label: '優勝点' },
				{ key: 'gym2_loser_bracket_champion_points', label: '敗者戦ブロック優勝' }
			]
		},
		{
			location: 'ground',
			label: sportNames.ground || 'グラウンド',
			items: [
				{ key: 'ground_win1_points', label: '1勝点' },
				{ key: 'ground_win2_points', label: '2勝点' },
				{ key: 'ground_win3_points', label: '3勝点' },
				{ key: 'ground_champion_points', label: '優勝点' }
			]
		}
	];

	const sportSections = sportGroups
		.map((group) => {
			const entries = group.items
				.map((item) => ({
					key: item.key,
					label: item.label,
					value: toNumber(classScore[item.key])
				}))
				.filter((entry) => entry.value > 0);

			const total = entries.reduce((acc, item) => acc + item.value, 0);

			return {
				location: group.location,
				label: group.label,
				total,
				entries
			};
		})
		.filter((section) => section.total > 0);

	const scoreItems = [
		...baseItems,
		...sportSections.flatMap((section) =>
			section.entries.map((entry) => ({
				key: entry.key,
				label: `${section.label} ${entry.label}`,
				value: entry.value
			}))
		)
	].filter((item) => item.value > 0);

	const categoryBreakdown = [
		{ label: 'アンケート', value: toNumber(classScore.survey_points) },
		{ label: '出席', value: toNumber(classScore.attendance_points) },
		...sportSections.map((section) => ({
			label: section.label,
			value: section.total
		})),
		{ label: '昼競技', value: toNumber(classScore.noon_game_points) },
		{ label: '初期点', value: toNumber(classScore.initial_points) }
	].filter((item) => item.value > 0);

	const pointHighlights = [...scoreItems]
		.sort((a, b) => b.value - a.value)
		.slice(0, 3);

	return {
		scoreItems,
		categoryBreakdown,
		pointHighlights,
		sportSections
	};
};

export const load = async ({ fetch, locals, request }) => {
	const user = locals.user;
	if (!user) {
		return {
			myClassScore: null,
			user: null,
			error: 'ユーザー情報が見つかりません。'
		};
	}
	if (!user.class_id) {
		return {
			myClassScore: null,
			user,
			error: 'クラスに所属していません。'
		};
	}

	try {
		const headers = {
			cookie: request.headers.get('cookie')
		};
		const authHeader = request.headers.get('Authorization');
		if (authHeader) {
			headers.Authorization = authHeader;
		}

		const scoreResponse = await fetch(`${BACKEND_URL}/api/scores/class`, {
			headers
		});
		if (!scoreResponse.ok) {
			throw new Error('クラスの得点一覧の取得に失敗しました。');
		}
		const classScores = await scoreResponse.json();
		const myClassScore = classScores.find((score) => score.class_id === user.class_id);

		if (!myClassScore) {
			return {
				myClassScore: null,
				user,
				error: 'あなたのクラスの得点情報が見つかりませんでした。'
			};
		}

		const season = myClassScore.season;
		const primaryRankRaw = season === 'spring' ? myClassScore.rank_current_event : myClassScore.rank_overall;
		const primaryRank = (primaryRankRaw === 0 || primaryRankRaw === null || primaryRankRaw === undefined) ? null : primaryRankRaw;
		const primaryPoints =
			season === 'spring' ? myClassScore.total_points_current_event : myClassScore.total_points_overall;
		const secondaryRankRaw = season === 'spring' ? myClassScore.rank_overall : myClassScore.rank_current_event;
		const secondaryRank = (secondaryRankRaw === 0 || secondaryRankRaw === null || secondaryRankRaw === undefined) ? null : secondaryRankRaw;
		const secondaryPoints =
			season === 'spring' ? myClassScore.total_points_overall : myClassScore.total_points_current_event;

		const breakdown = buildScoreBreakdown(myClassScore);

		let upcomingMatches = [];
		let assignedSports = [];
		const activeEventResponse = await fetch(`${BACKEND_URL}/api/events/active`, { headers });
		if (activeEventResponse.ok) {
			const activeEventPayload = await activeEventResponse.json();
			const activeEventId = activeEventPayload?.event_id;

			if (activeEventId) {
				const [teamsResponse, tournamentsResponse, noonResponse] = await Promise.all([
					fetch(`${BACKEND_URL}/api/qrcode/teams`, { headers }),
					fetch(`${BACKEND_URL}/api/student/events/${activeEventId}/tournaments`, { headers }),
					fetch(`${BACKEND_URL}/api/student/events/${activeEventId}/noon-game/session`, { headers })
				]);

				const teams = teamsResponse.ok ? await teamsResponse.json() : [];
				const currentEventTeams = Array.isArray(teams)
					? teams.filter((team) => Number(team?.event_id) === Number(activeEventId))
					: [];
				assignedSports = buildAssignedSports(currentEventTeams);

				const tournamentMatches = tournamentsResponse.ok
					? buildTournamentUpcomingMatches(await tournamentsResponse.json(), currentEventTeams)
					: [];

				const noonMatches = noonResponse.ok
					? buildNoonUpcomingMatches(await noonResponse.json(), user.class_id)
					: [];

				upcomingMatches = [...tournamentMatches, ...noonMatches]
					.filter((match) => match.start_time || match.opponent_name || match.location)
					.sort((left, right) => left.sort_value - right.sort_value)
					.map(({ sort_value, status, ...match }) => match);
			}
		}

		return {
			user,
			myClassScore: {
				...myClassScore,
				primaryRank,
				primaryPoints,
				secondaryRank,
				secondaryPoints
			},
			scoreItems: breakdown.scoreItems,
			categoryBreakdown: breakdown.categoryBreakdown,
			pointHighlights: breakdown.pointHighlights,
			sportSections: breakdown.sportSections,
			assignedSports,
			upcomingMatches,
			scoreHistory: []
		};
	} catch (error) {
		console.error(error);
		return {
			myClassScore: null,
			user,
			error: 'クラスの得点の読み込みに失敗しました。'
		};
	}
};
