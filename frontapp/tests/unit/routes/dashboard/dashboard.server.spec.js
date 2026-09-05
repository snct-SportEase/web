import { describe, expect, it, vi } from 'vitest';
import { load } from '$src/routes/dashboard/+page.server.js';

describe('dashboard server load', () => {
	it('クラス取得時に認証情報をバックエンドへ転送する', async () => {
		const fetchMock = vi.fn(async () => ({ ok: true, json: async () => [] }));
		const request = new Request('http://localhost/dashboard', {
			headers: { cookie: 'session_token=test-session', authorization: 'Bearer test-token' }
		});

		await load({ locals: { user: { roles: [] } }, fetch: fetchMock, request });

		const [, options] = fetchMock.mock.calls.find(([url]) => String(url).endsWith('/api/classes'));
		expect(options.headers).toEqual({
			cookie: 'session_token=test-session',
			authorization: 'Bearer test-token'
		});
	});
});
