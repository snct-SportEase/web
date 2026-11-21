import * as env from '$app/env/private';
import type { PageServerLoad } from './$types';

type NoonEntry = { id: number; class_ids?: number[]; resolved_name?: string; display_name?: string | null; entry_index: number };
type NoonResult = { winner: string; details?: { entry_id: number; rank?: number | null }[] };
type NoonMatch = { id: number; title?: string | null; status?: string; scheduled_at?: string | null; entries?: NoonEntry[]; result?: NoonResult | null };
type MatchSummary = { match_id: number; round: number; round_label: string; opponent_name: string; match_status: string; start_time?: string; result: string; score?: string | null };
type Progress = { sport_name: string; team_name: string | null; tournament_name: string; status: string; current_round: string; next_match?: MatchSummary; last_match?: MatchSummary };
type ClassInfo = { id: number; name: string; student_count: number; attend_count: number };

const BACKEND_URL = env.BACKEND_URL;

function createHeaders(request: Request): Record<string, string> {
	const headers: Record<string, string> = {
		cookie: request.headers.get('cookie') ?? ''
	};
	const authHeader = request.headers.get('Authorization');
	if (authHeader) {
		headers.Authorization = authHeader;
	}
	return headers;
}

function formatNoonStatus(status: string | undefined): string {
	const labels: Record<string, string> = {
		scheduled: '予定',
		in_progress: '進行中',
		finished: '終了',
		cancelled: '中止'
	};
	return (status && labels[status]) || status || '未定';
}

function toDateValue(value: string | null | undefined): number {
	if (!value) return Number.POSITIVE_INFINITY;
	const date = new Date(value);
	return Number.isNaN(date.getTime()) ? Number.POSITIVE_INFINITY : date.getTime();
}

function getParticipantEntries(match: NoonMatch, classId: number): NoonEntry[] {
	return (match?.entries || []).filter((entry) =>
		Array.isArray(entry?.class_ids) && entry.class_ids.some((id) => Number(id) === Number(classId))
	);
}

function getParticipantName(match: NoonMatch, classId: number, fallbackClassName: string | null): string | null {
	const entry = getParticipantEntries(match, classId)[0];
	return entry?.resolved_name || entry?.display_name || fallbackClassName;
}

function getOpponentName(match: NoonMatch, classId: number): string {
	const participantEntries = getParticipantEntries(match, classId);
	if (participantEntries.length === 0) return '';

	const participantIds = new Set(participantEntries.map((entry) => String(entry.id)));
	const opponent = (match?.entries || []).find((entry) => !participantIds.has(String(entry.id)));
	return opponent?.resolved_name || opponent?.display_name || '';
}

function buildNoonResult(match: NoonMatch, classId: number): string | null {
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

function buildNoonProgressEntries(session: { name?: string } | null | undefined, matches: NoonMatch[] | undefined, classId: number, className: string | null): Progress[] {
	const relevantMatches = (matches || [])
		.filter((match) => getParticipantEntries(match, classId).length > 0)
		.sort((left, right) => toDateValue(left?.scheduled_at) - toDateValue(right?.scheduled_at));

	return relevantMatches.map((match) => {
		const isFinished = match?.status === 'finished' || Boolean(match?.result);
		const resultLabel = buildNoonResult(match, classId);
		const participantName = getParticipantName(match, classId, className);
		const opponentName = getOpponentName(match, classId);
		const eventLabel = session?.name || match?.title || '昼競技';

		const entry: Progress = {
			sport_name: eventLabel,
			team_name: participantName,
			tournament_name: match?.title || eventLabel,
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

export const load = (async ({ fetch, request, locals }: Pick<Parameters<PageServerLoad>[0], 'fetch' | 'request' | 'locals'>) => {
	const user = locals.user;
	const configuredClassId = user?.class_id;

	if (!configuredClassId) {
		return {
			isClassMember: false,
			className: null,
			classInfo: null
		};
	}

	const className = null;
	const headers = createHeaders(request);

	try {
		const response = await fetch(`${BACKEND_URL}/api/student/class-progress?view=summary`, {
			headers
		});

		if (!response.ok) {
			if (response.status === 403) {
				return {
					isClassMember: false,
					className,
					classInfo: null,
					progress: []
				};
			}
			const errorText = await response.text();
			throw new Error(`Failed to fetch class progress: ${response.status} ${errorText}`);
		}

		const payload: { class_id?: number | null; event_id?: number | null; class_name?: string | null; class_info?: ClassInfo | null; progress?: Progress[] } = await response.json();
		const classId = payload.class_id ?? null;
		const progress = Array.isArray(payload.progress) ? [...payload.progress] : [];

        const eventId = payload.event_id;
        if (classId !== null && eventId) {
            const noonResponse = await fetch(
                `${BACKEND_URL}/api/student/events/${eventId}/noon-game/session`, { headers }
            );
            if (noonResponse.ok) {
                const noonPayload: { session?: { name?: string }; matches?: NoonMatch[] } = await noonResponse.json();
                progress.push(...buildNoonProgressEntries(
                    noonPayload?.session, noonPayload?.matches,
                    classId, payload.class_name ?? className
                ));
            }
        }

		return {
			isClassMember: true,
			classId,
			className: payload.class_name ?? className,
			classInfo: payload.class_info ?? null,
			progress
		};
	} catch (error) {
		console.error('Error loading class progress:', error);
		return {
			isClassMember: true,
			className,
			classInfo: null,
			progress: [],
			error: error instanceof Error ? error.message : String(error)
		};
	}
}) satisfies PageServerLoad;
