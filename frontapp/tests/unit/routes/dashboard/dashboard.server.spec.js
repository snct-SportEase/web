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
 it('プロフィール設定済みユーザーにはクラス一覧を取得しない', async () => {
  const fetch = vi.fn(async () => ({ ok: true, json: async () => ({}) }));
  await load({ locals: { user: { is_profile_complete: true, class_id: 1, roles: [{ name: 'student' }] } }, fetch, request: new Request('http://localhost/dashboard') });
  expect(fetch).toHaveBeenCalledTimes(1);
  expect(String(fetch.mock.calls[0][0])).toMatch(/class-progress$/);
 });
 it('大会一覧とクラス状況を並列取得し、一方の403でも他方の結果を返す', async () => {
  const releases = [];
  const fetch = vi.fn(() => new Promise(resolve => releases.push(resolve)));
  const pending = load({ locals: { user: { is_profile_complete: true, class_id: 1, roles: [{ name: 'root' }] } }, fetch, request: new Request('http://localhost/dashboard') });
  expect(fetch).toHaveBeenCalledTimes(2);
  releases[0]({ ok: true, json: async () => [{ id: 7 }] });
  releases[1]({ ok: false, status: 403 });
  const result = await pending;
  expect(result.events).toEqual([{ id: 7 }]);
  expect(result.isClassMember).toBe(false);
 });

});
