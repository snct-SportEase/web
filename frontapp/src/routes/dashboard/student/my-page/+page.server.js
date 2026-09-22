import { env } from '$env/dynamic/private';
const BACKEND_URL = env.BACKEND_URL;

const toNumber = (value) => (typeof value === 'number' && !Number.isNaN(value) ? value : 0);

const canViewHiddenScores = (user) =>
	user?.roles?.some((role) => role?.name === 'admin' || role?.name === 'root') ?? false;

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

const getRoundLabel = (tournamentData, match) => {
	if (match?.isBronzeMatch) {
		return '3位決定戦';
	}
	return tournamentData?.rounds?.[match?.roundIndex]?.name || `Round ${toNumber(match?.roundIndex) + 1}`;
};

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

const buildTournamentMatchResults = (tournaments, teams) => {
	if (!Array.isArray(tournaments) || !Array.isArray(teams) || teams.length === 0) return [];

	const teamByID = new Map(teams.map((team) => [Number(team.id), team]));
	const results = [];

	for (const tournament of tournaments) {
		let tournamentData = tournament?.data;
		if (typeof tournamentData === 'string') {
			try {
				tournamentData = JSON.parse(tournamentData);
			} catch {
				continue;
			}
		}

		for (const match of tournamentData?.matches || []) {
			const sides = Array.isArray(match?.sides) ? match.sides : [];
			const mySide = sides.find((side) => teamByID.has(Number(side?.teamId)));
			if (!mySide) continue;

			const myScore = mySide?.scores?.[0]?.mainScore;
			const opponentSide = sides.find((side) => Number(side?.teamId) !== Number(mySide.teamId));
			const opponentScore = opponentSide?.scores?.[0]?.mainScore;
			const status = normalizeMatchStatus(match?.matchStatus);
			const isFinished = status === 'completed' || status === 'finished' || myScore !== undefined;
			if (!isFinished) continue;

			const result = mySide.isWinner === true
				? '勝利'
				: mySide.isWinner === false || (myScore !== undefined && opponentScore !== undefined && Number(myScore) < Number(opponentScore))
					? '敗戦'
					: Number(myScore) === Number(opponentScore)
						? '引き分け'
						: '終了';
			const playedAt = match?.startTime || match?.rainyModeStartTime || null;

			results.push({
				id: `tournament-result-${tournament.id}-${match.id}`,
				sport_name: teamByID.get(Number(mySide.teamId))?.sport_name || tournament?.name || '競技',
				opponent_name: getContestantName(tournamentData, opponentSide),
				round_label: getRoundLabel(tournamentData, match),
				result,
				score: myScore !== undefined && opponentScore !== undefined ? `${myScore} - ${opponentScore}` : '',
				played_at: playedAt,
				sort_value: playedAt ? toDateValue(playedAt) : Number.NEGATIVE_INFINITY
			});
		}
	}

	return results.sort((left, right) => right.sort_value - left.sort_value);
};

const buildNoonMatchResults = (sessionPayload, classId) => {
	if (!classId || !Array.isArray(sessionPayload?.matches)) return [];

	return sessionPayload.matches
		.filter((match) => getParticipantEntries(match, classId).length > 0)
		.filter((match) => match?.result || ['completed', 'finished'].includes(normalizeMatchStatus(match?.status)))
		.map((match) => {
			const participantEntries = getParticipantEntries(match, classId);
			const participantIDs = new Set(participantEntries.map((entry) => String(entry.id)));
			const detail = (match?.result?.details || []).find((item) => participantIDs.has(String(item.entry_id)));
			const opponent = (match?.entries || []).find((entry) => !participantIDs.has(String(entry.id)));
			const result = detail?.rank ? `${detail.rank}位` : match?.result?.winner_display || '終了';

			return {
				id: `noon-result-${match.id}`,
				sport_name: match?.title || sessionPayload?.session?.name || '昼競技',
				opponent_name: opponent?.resolved_name || opponent?.display_name || '',
				round_label: '昼競技',
				result,
				score: detail?.competition_score !== undefined ? `${detail.competition_score}点` : '',
				played_at: match?.scheduled_at || null,
				sort_value: toDateValue(match?.scheduled_at)
			};
		})
		.sort((left, right) => right.sort_value - left.sort_value);
};

const readJson = async (response, fallback) => {
	if (!response?.ok) return fallback;
	try {
		return await response.json();
	} catch {
		return fallback;
	}
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
	const isAutumn = classScore.season === 'autumn';

	const baseItems = [
		...(isAutumn ? [{ key: 'initial_points', label: '初期点', value: toNumber(classScore.initial_points) }] : []),
		...(isAutumn ? [{ key: 'survey_points', label: 'アンケート', value: toNumber(classScore.survey_points) }] : []),
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
		...(isAutumn ? [{ label: 'アンケート', value: toNumber(classScore.survey_points) }] : []),
		{ label: '出席', value: toNumber(classScore.attendance_points) },
		...sportSections.map((section) => ({
			label: section.label,
			value: section.total
		})),
		{ label: '昼競技', value: toNumber(classScore.noon_game_points) },
		...(isAutumn ? [{ label: '初期点', value: toNumber(classScore.initial_points) }] : [])
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
	const emptyData = {
		myClassScore: null,
		scoreItems: [],
		categoryBreakdown: [],
		pointHighlights: [],
		sportSections: [],
		assignedSports: [],
		upcomingMatches: [],
		matchResults: [],
		classInfo: null,
		classProgress: [],
		notifications: [],
		sportGuidelines: [],
		competitionGuidelinesUrl: null,
		surveyUrl: null,
		scoreHistory: []
	};

	if (!user) {
		return {
			...emptyData,
			user: null,
			error: 'ユーザー情報が見つかりません。'
		};
	}
	if (!user.class_id) {
		return {
			...emptyData,
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

		let activeEvent = null;
		const activeEventResponse = await fetch(`${BACKEND_URL}/api/events/active`, { headers });
		if (activeEventResponse.ok) {
			activeEvent = await activeEventResponse.json();
		}
		const activeEventId = activeEvent?.event_id ?? activeEvent?.id ?? null;
		let scoresHidden = Boolean(activeEvent?.hide_scores && !canViewHiddenScores(user));
		let scoreResponse = null;
		let classProgressResponse = null;
		let teamsResponse = null;
		let tournamentsResponse = null;
		let noonResponse = null;
		let notificationsResponse = null;
		let sportsResponse = null;

		if (activeEventId) {
			[
				scoreResponse,
				classProgressResponse,
				teamsResponse,
				tournamentsResponse,
				noonResponse,
				notificationsResponse,
				sportsResponse
			] = await Promise.all([
				scoresHidden ? Promise.resolve(null) : fetch(`${BACKEND_URL}/api/scores/class`, { headers }),
				fetch(`${BACKEND_URL}/api/student/class-progress`, { headers }),
				fetch(`${BACKEND_URL}/api/barcode/teams`, { headers }),
				fetch(`${BACKEND_URL}/api/student/events/${activeEventId}/tournaments`, { headers }),
				fetch(`${BACKEND_URL}/api/student/events/${activeEventId}/noon-game/session`, { headers }),
				fetch(`${BACKEND_URL}/api/notifications?limit=3`, { headers }),
				fetch(`${BACKEND_URL}/api/events/${activeEventId}/sports`, { headers })
			]);
		}

		if (scoreResponse?.status === 403) scoresHidden = true;
		const classScores = scoresHidden ? [] : await readJson(scoreResponse, []);
		const rawClassScore = Array.isArray(classScores)
			? classScores.find((score) => Number(score.class_id) === Number(user.class_id))
			: null;
		let myClassScore = null;
		let breakdown = buildScoreBreakdown(null);
		if (rawClassScore) {
			const season = rawClassScore.season;
			const primaryRankRaw = season === 'spring' ? rawClassScore.rank_current_event : rawClassScore.rank_overall;
			const secondaryRankRaw = season === 'spring' ? rawClassScore.rank_overall : rawClassScore.rank_current_event;
			myClassScore = {
				...rawClassScore,
				primaryRank: [0, null, undefined].includes(primaryRankRaw) ? null : primaryRankRaw,
				primaryPoints: season === 'spring' ? rawClassScore.total_points_current_event : rawClassScore.total_points_overall,
				secondaryRank: [0, null, undefined].includes(secondaryRankRaw) ? null : secondaryRankRaw,
				secondaryPoints: season === 'spring' ? rawClassScore.total_points_overall : rawClassScore.total_points_current_event
			};
			breakdown = buildScoreBreakdown(rawClassScore);
		}

		const classPayload = await readJson(classProgressResponse, {});
		const teamsPayload = await readJson(teamsResponse, []);
		const tournamentsPayload = await readJson(tournamentsResponse, []);
		const noonPayload = await readJson(noonResponse, {});
		const notificationPayload = await readJson(notificationsResponse, {});
		const sportsPayload = await readJson(sportsResponse, []);
		const currentEventTeams = Array.isArray(teamsPayload)
			? teamsPayload.filter((team) => Number(team?.event_id) === Number(activeEventId))
			: [];
		let upcomingMatches = [];
		if (activeEventId) {
			const tournamentMatches = buildTournamentUpcomingMatches(tournamentsPayload, currentEventTeams);
			const noonMatches = buildNoonUpcomingMatches(noonPayload, user.class_id);
			upcomingMatches = [...tournamentMatches, ...noonMatches]
				.filter((match) => match.start_time || match.opponent_name || match.location)
				.sort((left, right) => left.sort_value - right.sort_value)
				.map((match) => ({
					id: match.id,
					start_time: match.start_time,
					sport_name: match.sport_name,
					opponent_name: match.opponent_name,
					location: match.location
				}));
		}
		const matchResults = [
			...buildTournamentMatchResults(tournamentsPayload, currentEventTeams),
			...buildNoonMatchResults(noonPayload, user.class_id)
		].sort((left, right) => right.sort_value - left.sort_value).slice(0, 5);
		const sportGuidelines = (Array.isArray(sportsPayload) ? sportsPayload : [])
			.filter((sport) => sport?.rules_pdf_url)
			.map((sport) => ({
				id: sport.sport_id ?? sport.id,
				name: sport.sport_name || sport.name || '競技要項',
				url: sport.rules_pdf_url
			}));

		return {
			user,
			myClassScore,
			scoresHidden: scoresHidden || undefined,
			error: !scoresHidden && !myClassScore ? 'あなたのクラスの得点情報が見つかりませんでした。' : undefined,
			scoreItems: breakdown.scoreItems,
			categoryBreakdown: breakdown.categoryBreakdown,
			pointHighlights: breakdown.pointHighlights,
			sportSections: breakdown.sportSections,
			assignedSports: buildAssignedSports(currentEventTeams),
			upcomingMatches,
			matchResults,
			classInfo: classPayload?.class_info ?? null,
			classProgress: Array.isArray(classPayload?.progress) ? classPayload.progress : [],
			notifications: Array.isArray(notificationPayload?.notifications) ? notificationPayload.notifications.slice(0, 3) : [],
			sportGuidelines,
			competitionGuidelinesUrl: activeEvent?.competition_guidelines_pdf_url ?? null,
			surveyUrl: activeEvent?.is_survey_published && activeEvent?.survey_url ? activeEvent.survey_url : null,
			scoreHistory: []
		};
	} catch (error) {
		console.error(error);
		return {
			...emptyData,
			user,
			error: 'マイページ情報の読み込みに失敗しました。'
		};
	}
};
