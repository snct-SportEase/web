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
});
