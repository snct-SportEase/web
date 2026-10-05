import { afterEach, describe, expect, it, vi } from 'vitest';

afterEach(() => { vi.unstubAllGlobals(); vi.unstubAllEnvs(); vi.resetModules(); });

describe('dashboard active event sharing', () => {
 it('リクエスト内で一度取得した大会情報をSSRへ共有する', async () => {
  vi.stubEnv('BACKEND_URL', 'http://127.0.0.1:8080');
  const user = { id: 'student', roles: [{ name: 'student' }] };
  const activeEvent = { event_id: 7, event_name: '大会', test_run_state: '' };
  const fetch = vi.fn(async (url) => new Response(JSON.stringify(
   String(url).endsWith('/api/auth/user') ? user : activeEvent
  )));
  vi.stubGlobal('fetch', fetch);
  const { handle } = await import('$src/hooks.server.js');
  const event = {
   url: new URL('http://localhost/dashboard/student/my-page'),
   cookies: { get: () => 'session' }, locals: {}
  };
  const resolve = vi.fn(async ({ locals }) => new Response(JSON.stringify(locals.activeEvent)));
  const response = await handle({ event, resolve });
  expect(await response.json()).toEqual(activeEvent);
  expect(event.locals.user).toEqual(user);
  expect(fetch).toHaveBeenCalledTimes(2);
  expect(response.headers.get('Server-Timing')).toMatch(/^ssr;dur=/);
 });
});
