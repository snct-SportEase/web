import { beforeEach, describe, expect, it, vi } from 'vitest';
import { load } from './+page.server.js';

const makeRequest = () =>
	new Request('http://localhost/dashboard/student/my-page', {
		headers: {
			cookie: 'session=student-session',
			Authorization: 'Bearer test-token'
		}
	});

const student = {
	id: 'student-1',
	class_id: 10,
	roles: [{ name: 'student' }]
};

const admin = {
	id: 'admin-1',
	class_id: 10,
	roles: [{ name: 'admin' }]
};

const hasPath = (url, path) => String(url).endsWith(path);

describe('student my-page server load', () => {
	beforeEach(() => {
		vi.clearAllMocks();
	});

	it('得点非表示中の一般ユーザーには得点APIを呼ばず非表示状態を返す', async () => {
		expect.assertions(5);

		const fetchMock = vi.fn((url) => {
			if (hasPath(url, '/api/events/active')) {
				return Promise.resolve({
					ok: true,
					json: () => Promise.resolve({ event_id: 1, hide_scores: true })
				});
			}

			return Promise.resolve({ ok: true, json: () => Promise.resolve([]) });
		});

		const result = await load({
			fetch: fetchMock,
			locals: { user: student },
			request: makeRequest()
		});

		expect(result.scoresHidden).toBe(true);
		expect(result.myClassScore).toBeNull();
		expect(result.scoreItems).toEqual([]);
		expect(fetchMock).toHaveBeenCalledWith(expect.stringMatching(/\/api\/events\/active$/), expect.any(Object));
		expect(fetchMock).not.toHaveBeenCalledWith(expect.stringMatching(/\/api\/scores\/class$/), expect.any(Object));
	});

	it('得点APIが403を返した場合も非表示状態として扱う', async () => {
		expect.assertions(4);

		const fetchMock = vi.fn((url) => {
			if (hasPath(url, '/api/events/active')) {
				return Promise.resolve({
					ok: true,
					json: () => Promise.resolve({ event_id: 1, hide_scores: false })
				});
			}

			if (hasPath(url, '/api/scores/class')) {
				return Promise.resolve({
					ok: false,
					status: 403,
					json: () => Promise.resolve({ error: '得点一覧は現在非表示です。' })
				});
			}

			return Promise.resolve({ ok: true, json: () => Promise.resolve([]) });
		});

		const result = await load({
			fetch: fetchMock,
			locals: { user: student },
			request: makeRequest()
		});

		expect(result.scoresHidden).toBe(true);
		expect(result.myClassScore).toBeNull();
		expect(fetchMock).toHaveBeenCalledWith(expect.stringMatching(/\/api\/events\/active$/), expect.any(Object));
		expect(fetchMock).toHaveBeenCalledWith(expect.stringMatching(/\/api\/scores\/class$/), expect.any(Object));
	});

	it('得点非表示中でも管理者は得点情報を取得できる', async () => {
		expect.assertions(4);

		const fetchMock = vi.fn((url) => {
			if (hasPath(url, '/api/events/active')) {
				return Promise.resolve({
					ok: true,
					json: () => Promise.resolve({ event_id: 1, hide_scores: true })
				});
			}

			if (hasPath(url, '/api/scores/class')) {
				return Promise.resolve({
					ok: true,
					json: () =>
						Promise.resolve([
							{
								class_id: 10,
								class_name: '1A',
								season: 'spring',
								rank_current_event: 2,
								rank_overall: 3,
								total_points_current_event: 80,
								total_points_overall: 120
							}
						])
				});
			}

			return Promise.resolve({ ok: true, json: () => Promise.resolve([]) });
		});

		const result = await load({
			fetch: fetchMock,
			locals: { user: admin },
			request: makeRequest()
		});

		expect(result.scoresHidden).toBeUndefined();
		expect(result.myClassScore.primaryRank).toBe(2);
		expect(result.myClassScore.primaryPoints).toBe(80);
		expect(fetchMock).toHaveBeenCalledWith(expect.stringMatching(/\/api\/scores\/class$/), expect.any(Object));
	});
});
