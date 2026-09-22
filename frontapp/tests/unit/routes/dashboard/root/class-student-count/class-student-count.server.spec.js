import { describe, expect, it, vi } from 'vitest';
import { load } from '$src/routes/dashboard/root/class-student-count/+page.server.js';

describe('class student count server load', () => {
	it('クラス取得時に認証情報をバックエンドへ転送する', async () => {
		const fetchMock = vi.fn(async () => ({ ok: true, json: async () => [] }));
		const request = new Request('http://localhost/dashboard/root/class-student-count', {
			headers: { cookie: 'session_token=root-session', authorization: 'Bearer root-token' }
		});

		await load({ fetch: fetchMock, request });

		const [, options] = fetchMock.mock.calls[0];
		expect(options.headers).toEqual({
			cookie: 'session_token=root-session',
			authorization: 'Bearer root-token'
		});
	});

	it('専教を人数設定対象から除外する', async () => {
		const fetchMock = vi.fn(async () => ({
			ok: true,
			json: async () => [
				{ id: 1, name: '1-1', student_count: 40 },
				{ id: 16, name: '専教', student_count: 0 }
			]
		}));
		const request = new Request('http://localhost/dashboard/root/class-student-count');

		const result = await load({ fetch: fetchMock, request });

		expect(result.classes).toEqual([{ id: 1, name: '1-1', student_count: 40 }]);
	});
});
