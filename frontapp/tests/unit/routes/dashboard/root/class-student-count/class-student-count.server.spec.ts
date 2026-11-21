import { describe, expect, it, vi } from 'vitest';
import { load } from '$src/routes/dashboard/root/class-student-count/+page.server.js';

describe('class student count server load', () => {
 it('クラス取得時に認証情報をバックエンドへ転送する', async () => {
  const fetchMock = vi.fn(async (input: RequestInfo | URL, options?: RequestInit) => { void input; void options; return Response.json([]); });
  const request = new Request('http://localhost/dashboard/root/class-student-count', { headers: { cookie: 'session_token=root-session', authorization: 'Bearer root-token' } });
  await load({ fetch: fetchMock, request });
  expect(fetchMock.mock.calls[0][1]?.headers).toEqual({ cookie: 'session_token=root-session', authorization: 'Bearer root-token' });
 });
 it('専教を人数設定対象から除外する', async () => {
  const fetchMock = vi.fn(async () => Response.json([{ id: 1, name: '1-1', student_count: 40 }, { id: 16, name: '専教', student_count: 0 }]));
  const result = await load({ fetch: fetchMock, request: new Request('http://localhost/dashboard/root/class-student-count') });
  expect(result.classes).toEqual([{ id: 1, name: '1-1', student_count: 40 }]);
 });
});
