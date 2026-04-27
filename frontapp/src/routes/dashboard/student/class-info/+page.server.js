import { env } from '$env/dynamic/private';

const BACKEND_URL = env.BACKEND_URL;

function createHeaders(request) {
	const headers = {
		cookie: request.headers.get('cookie')
	};
	const authHeader = request.headers.get('Authorization');
	if (authHeader) {
		headers.Authorization = authHeader;
	}
	return headers;
}

function formatNoonStatus(status) {
	const labels = {
		scheduled: '予定',
		in_progress: '進行中',
		finished: '終了',
		cancelled: '中止'
	};
	return labels[status] || status || '未定';
}

function toDateValue(value) {
	if (!value) return Number.POSITIVE_INFINITY;
	const date = new Date(value);
	return Number.isNaN(date.getTime()) ? Number.POSITIVE_INFINITY : date.getTime();
}

function getParticipantEntries(match, classId) {
	return (match?.entries || []).filter((entry) =>
		Array.isArray(entry?.class_ids) && entry.class_ids.some((id) => Number(id) === Number(classId))
	);
}

function getParticipantName(match, classId, fallbackClassName) {
	const entry = getParticipantEntries(match, classId)[0];
	return entry?.resolved_name || entry?.display_name || fallbackClassName;
}

function getOpponentName(match, classId) {
	const participantEntries = getParticipantEntries(match, classId);
	if (participantEntries.length === 0) return '';

	const participantIds = new Set(participantEntries.map((entry) => String(entry.id)));
	const opponent = (match?.entries || []).find((entry) => !participantIds.has(String(entry.id)));
	return opponent?.resolved_name || opponent?.display_name || '';
}

function buildNoonResult(match, classId) {
	if (!match?.result) {
		return null;
	}

	const detail = (match.result.details || []).find((item) => {
		const entry = (match.entries || []).find((candidate) => String(candidate.id) === String(item.entry_id));
		return entry?.class_ids?.some((id) => Number(id) === Number(classId));
	});

	if (detail?.rank) {
		return `${detail.rank}位`;
	}

	const participantEntries = getParticipantEntries(match, classId);
	const homeIncludesClass = participantEntries.some((entry) => entry.entry_index === 0);
	const awayIncludesClass = participantEntries.some((entry) => entry.entry_index !== 0);

	if (match.result.winner === 'draw') {
		return '引き分け';
	}
	if (match.result.winner === 'home') {
		return homeIncludesClass ? '勝利' : '敗退';
	}
	if (match.result.winner === 'away') {
		return awayIncludesClass ? '勝利' : '敗退';
	}

	return '終了';
}

function buildNoonProgressEntries(session, matches, classId, className) {
	const relevantMatches = (matches || [])
		.filter((match) => getParticipantEntries(match, classId).length > 0)
		.sort((left, right) => toDateValue(left?.scheduled_at) - toDateValue(right?.scheduled_at));

	return relevantMatches.map((match) => {
		const isFinished = match?.status === 'finished' || Boolean(match?.result);
		const resultLabel = buildNoonResult(match, classId);
		const participantName = getParticipantName(match, classId, className);
		const opponentName = getOpponentName(match, classId);

		const entry = {
			sport_name: '昼競技',
			team_name: participantName,
			tournament_name: match?.title || session?.name || '昼競技',
			status: isFinished ? resultLabel || '終了' : formatNoonStatus(match?.status),
			current_round: match?.title || formatNoonStatus(match?.status),
			next_match: undefined,
			last_match: undefined
		};

		const matchSummary = {
			match_id: match?.id,
			round: 0,
			round_label: match?.title || '昼競技',
			opponent_name: opponentName,
			match_status: formatNoonStatus(match?.status),
			start_time: match?.scheduled_at || undefined,
			result: resultLabel || ''
		};

		if (isFinished) {
			entry.last_match = matchSummary;
		} else {
			entry.next_match = matchSummary;
		}

		return entry;
	});
}

/** @type {import('./$types').PageServerLoad} */
export async function load({ fetch, request, locals }) {
	const user = locals.user;
	const classRole = user?.roles?.find(
		(role) => typeof role?.name === 'string' && role.name.endsWith('_rep')
	);

	if (!classRole) {
		return {
			isClassRep: false,
			className: null,
			classInfo: null
		};
	}

	const className = classRole.name.slice(0, -4);
	const headers = createHeaders(request);

	try {
		const response = await fetch(`${BACKEND_URL}/api/student/class-progress`, {
			headers
		});

		if (!response.ok) {
			if (response.status === 403) {
				return {
					isClassRep: false,
					className,
					classInfo: null,
					progress: []
				};
			}
			const errorText = await response.text();
			throw new Error(`Failed to fetch class progress: ${response.status} ${errorText}`);
		}

		const payload = await response.json();
		const classId = payload.class_id ?? null;
		const progress = Array.isArray(payload.progress) ? [...payload.progress] : [];

		if (classId !== null) {
			const activeEventResponse = await fetch(`${BACKEND_URL}/api/events/active`, { headers });
			if (activeEventResponse.ok) {
				const activeEventPayload = await activeEventResponse.json();
				const activeEventId = activeEventPayload?.event_id ?? null;

				if (activeEventId) {
					const noonResponse = await fetch(
						`${BACKEND_URL}/api/student/events/${activeEventId}/noon-game/session`,
						{ headers }
					);

					if (noonResponse.ok) {
						const noonPayload = await noonResponse.json();
						const noonProgress = buildNoonProgressEntries(
							noonPayload?.session,
							noonPayload?.matches,
							classId,
							payload.class_name ?? className
						);
						progress.push(...noonProgress);
					}
				}
			}
		}

		return {
			isClassRep: true,
			classId,
			className: payload.class_name ?? className,
			classInfo: payload.class_info ?? null,
			progress
		};
	} catch (error) {
		console.error('Error loading class progress:', error);
		return {
			isClassRep: true,
			className,
			classInfo: null,
			progress: [],
			error: error.message
		};
	}
}
