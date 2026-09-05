export async function fetchPublishedNoonGameSessions(eventId, fetcher = fetch) {
	const listResponse = await fetcher(`/api/student/events/${eventId}/noon-game/sessions`);
	if (!listResponse.ok) {
		const detail = await safeJson(listResponse);
		throw new Error(detail?.error || '昼競技情報を取得できませんでした。');
	}

	const listPayload = await listResponse.json();
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

async function safeJson(response) {
	try {
		return await response.json();
	} catch {
		return null;
	}
}

export function flattenNoonGameMatches(sessions) {
	return (sessions || []).flatMap((session) =>
		(session?.matches || []).map((match) => ({ ...match, session_name: session.name }))
	);
}
