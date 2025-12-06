import { BACKEND_URL } from '$env/static/private';

const toNumber = (value) => (typeof value === 'number' && !Number.isNaN(value) ? value : 0);

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
			upcomingMatches: [],
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
