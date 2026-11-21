import { describe, expect, it, vi } from 'vitest';
import { load } from '$src/routes/dashboard/+page.server.js';

const user = { id: 'student', email: 'student@example.com', display_name: null, class_id: 1, is_profile_complete: true, created_at: '', updated_at: '', roles: [{ id: 1, name: 'student' }] };
const request = new Request('http://localhost/dashboard', { headers: { cookie: 'session_token=test-session', authorization: 'Bearer test-token' } });

describe('dashboard server load', () => {
 it('クラス取得時に認証情報をバックエンドへ転送する', async () => {
  const fetchMock = vi.fn(async (url: RequestInfo | URL, options?: RequestInit) => { void url; void options; return Response.json([]); });
  await load({ locals: { user: { ...user, is_profile_complete: false, roles: [] }, activeEvent: null }, fetch: fetchMock, request });
  const call = fetchMock.mock.calls.find(([url]) => String(url).endsWith('/api/classes'));
  expect(call?.[1]?.headers).toEqual({ cookie: 'session_token=test-session', authorization: 'Bearer test-token' });
 });
 it('プロフィール設定済みユーザーにはクラス一覧を取得しない', async () => {
  const fetch = vi.fn(async (url: RequestInfo | URL) => { void url; return Response.json({}); });
  await load({ locals: { user, activeEvent: null }, fetch, request });
  expect(fetch).toHaveBeenCalledTimes(1);
  expect(String(fetch.mock.calls[0][0])).toMatch(/class-progress$/);
 });
 it('大会一覧とクラス状況を並列取得し、一方の403でも他方の結果を返す', async () => {
  const releases: ((response: Response) => void)[] = [];
  const fetch = vi.fn(() => new Promise<Response>(resolve => releases.push(resolve)));
  const pending = load({ locals: { user: { ...user, roles: [{ id: 3, name: 'root' }] }, activeEvent: null }, fetch, request });
  expect(fetch).toHaveBeenCalledTimes(2);
  releases[0](Response.json([{ id: 7 }]));
  releases[1](new Response(null, { status: 403 }));
  const result = await pending;
  expect(result.events).toEqual([{ id: 7 }]);
  expect(result.isClassMember).toBe(false);
 });
});
