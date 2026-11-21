export interface NoonGameMatch {
	id: number;
	title?: string | null;
	scheduled_at?: string | null;
	location?: string | null;
	format?: string | null;
	memo?: string | null;
	status?: string;
	home_display_name?: string;
	away_display_name?: string;
	winner_display?: string | null;
	home_class_ids?: number[];
	away_class_ids?: number[];
	entries?: { id: number; resolved_name?: string; display_name?: string | null; class_ids?: number[] }[];
	result?: { winner: string; details?: { id: number; entry_id: number; rank?: number | null; points: number; competition_score?: number | null; entry_resolved_name?: string }[] } | null;
}

export interface NoonGameSession {
	id: number;
	name: string;
	matches?: NoonGameMatch[];
	session?: { id: number; name: string; template_key?: string };
	typing_results?: { team_name: string; match_1_score: number; match_2_score: number; match_3_score: number; total_score: number; rank: number; points: number }[];
}

export async function fetchPublishedNoonGameSessions(eventId: number | string, fetcher: typeof fetch = fetch): Promise<NoonGameSession[]> {
	const listResponse = await fetcher(`/api/student/events/${eventId}/noon-game/sessions`);
	if (!listResponse.ok) {
		const detail = await safeJson(listResponse);
		throw new Error(detail?.error || '昼競技情報を取得できませんでした。');
	}

	const listPayload: { sessions?: { id: number }[] } = await listResponse.json();
	const sessions = Array.isArray(listPayload?.sessions) ? listPayload.sessions : [];
	return Promise.all(
		sessions.map(async (session) => {
			const detailResponse = await fetcher(
				`/api/student/events/${eventId}/noon-game/sessions/${session.id}`
			);
			if (!detailResponse.ok) {
				const detail = await safeJson(detailResponse);
				throw new Error(detail?.error || '昼競技の詳細を取得できませんでした。');
			}
			return detailResponse.json();
		})
	);
}

async function safeJson(response: Response): Promise<{ error?: string } | null> {
	try {
		return await response.json();
	} catch {
		return null;
	}
}

export function flattenNoonGameMatches<T extends { name: string; matches?: M[] }, M extends NoonGameMatch>(sessions: T[] | null | undefined): (M & { session_name: string })[] {
	return (sessions || []).flatMap((session) =>
		(session?.matches || []).map((match) => ({ ...match, session_name: session.name }))
	);
}
