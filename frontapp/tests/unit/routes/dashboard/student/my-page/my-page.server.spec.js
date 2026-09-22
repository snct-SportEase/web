import { beforeEach, describe, expect, it, vi } from 'vitest';
import { load } from '$src/routes/dashboard/student/my-page/+page.server.js';

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

	it('参加試合・結果・クラス状況・通知・資料リンクを返す', async () => {
		const fetchMock = vi.fn((url) => {
			const value = String(url);
			if (value.endsWith('/api/events/active')) return Promise.resolve({ ok: true, json: () => Promise.resolve({ event_id: 1, competition_guidelines_pdf_url: 'https://example.com/event.pdf', survey_url: 'https://example.com/survey', is_survey_published: true }) });
			if (value.endsWith('/api/scores/class')) return Promise.resolve({ ok: true, json: () => Promise.resolve([{ class_id: 10, class_name: '1A', season: 'spring', rank_current_event: 1, total_points_current_event: 50 }]) });
			if (value.endsWith('/api/student/class-progress')) return Promise.resolve({ ok: true, json: () => Promise.resolve({ class_info: { name: '1A', student_count: 40, attend_count: 38 }, progress: [{ sport_name: 'バスケットボール', team_name: '1A', status: '決勝進出', current_round: '決勝' }] }) });
			if (value.endsWith('/api/barcode/teams')) return Promise.resolve({ ok: true, json: () => Promise.resolve([{ id: 101, event_id: 1, sport_id: 1, sport_name: 'バスケットボール', name: '1A' }]) });
			if (value.endsWith('/api/student/events/1/tournaments')) return Promise.resolve({ ok: true, json: () => Promise.resolve([{ id: 1, name: 'バスケットボール', data: { rounds: [{ name: '決勝' }], contestants: { c0: { players: [{ title: '1A' }] }, c1: { players: [{ title: '1B' }] } }, matches: [{ id: 1, roundIndex: 0, order: 0, matchStatus: 'completed', sides: [{ contestantId: 'c0', teamId: 101, isWinner: true, scores: [{ mainScore: 3 }] }, { contestantId: 'c1', teamId: 102, scores: [{ mainScore: 1 }] }] }] } }]) });
			if (value.endsWith('/api/student/events/1/noon-game/session')) return Promise.resolve({ ok: true, json: () => Promise.resolve({}) });
			if (value.includes('/api/notifications?limit=3')) return Promise.resolve({ ok: true, json: () => Promise.resolve({ notifications: [{ id: 1, title: '集合時刻変更', body: '9時集合です。' }] }) });
			if (value.endsWith('/api/events/1/sports')) return Promise.resolve({ ok: true, json: () => Promise.resolve([{ sport_id: 1, sport_name: 'バスケットボール', rules_pdf_url: 'https://example.com/basketball.pdf' }]) });
			return Promise.resolve({ ok: false, status: 404, json: () => Promise.resolve({}) });
		});

		const result = await load({ fetch: fetchMock, locals: { user: student }, request: makeRequest() });

		expect(result.classInfo).toEqual({ name: '1A', student_count: 40, attend_count: 38 });
		expect(result.classProgress).toHaveLength(1);
		expect(result.matchResults[0]).toMatchObject({ result: '勝利', score: '3 - 1', opponent_name: '1B' });
		expect(result.notifications[0].title).toBe('集合時刻変更');
		expect(result.sportGuidelines[0]).toEqual({ id: 1, name: 'バスケットボール', url: 'https://example.com/basketball.pdf' });
		expect(result.competitionGuidelinesUrl).toBe('https://example.com/event.pdf');
		expect(result.surveyUrl).toBe('https://example.com/survey');
	});
});
