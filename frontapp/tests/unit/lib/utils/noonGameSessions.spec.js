import { describe, expect, it, vi } from 'vitest';
import { fetchPublishedNoonGameSessions, flattenNoonGameMatches } from '$lib/utils/noonGameSessions.js';

describe('noonGameSessions', () => {
	it('公開済みの全セッションの詳細を取得する', async () => {
		const fetcher = vi.fn(async (url) => {
			if (url.endsWith('/sessions')) {
				return response({ sessions: [{ id: 2 }, { id: 5 }] });
			}
			if (url.endsWith('/sessions/2')) return response({ id: 2, name: 'リレー', matches: [{ id: 20 }] });
			if (url.endsWith('/sessions/5')) return response({ id: 5, name: '綱引き', matches: [{ id: 50 }] });
			throw new Error(`unexpected URL: ${url}`);
		});

		const sessions = await fetchPublishedNoonGameSessions(7, fetcher);

		expect(sessions.map((session) => session.id)).toEqual([2, 5]);
		expect(fetcher).toHaveBeenCalledTimes(3);
		expect(flattenNoonGameMatches(sessions)).toEqual([
			{ id: 20, session_name: 'リレー' },
			{ id: 50, session_name: '綱引き' }
		]);
	});
});

function response(body, ok = true) {
	return { ok, json: async () => body };
}
