import { describe, expect, it, vi } from 'vitest';
import { load } from '$src/routes/dashboard/student/score-list/+page.server.js';

const request = new Request('http://localhost/dashboard/student/score-list', { headers: { cookie: 'session_token=test', Authorization: 'Bearer token' } });
const user = { id: 'student', email: 'student@example.com', display_name: null, class_id: 1, is_profile_complete: true, created_at: '', updated_at: '' };

describe('score-list load', () => {
 it('得点APIを直接取得し、403を非表示として扱う', async () => {
  const fetch = vi.fn(async (input: RequestInfo | URL, options?: RequestInit) => { void input; void options; return new Response(null, { status: 403 }); });
  const result = await load({ fetch, request, locals: { user, activeEvent: null } });
  expect(result).toEqual({ scores: [], error: '得点一覧は現在非表示です。' });
  expect(fetch).toHaveBeenCalledTimes(1);
  expect(String(fetch.mock.calls[0][0])).toMatch(/\/api\/scores\/class$/);
  expect(fetch.mock.calls[0][1]?.headers).toEqual({ cookie: 'session_token=test', Authorization: 'Bearer token' });
 });
 it('バックエンドが返した得点を表示する', async () => {
  const scores = [{ class_id: 1, points: 10 }];
  const fetch = vi.fn(async () => Response.json(scores));
  expect(await load({ fetch, request, locals: { user: { ...user, id: 'admin' }, activeEvent: null } })).toEqual({ scores });
 });
});
