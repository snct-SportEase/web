import { afterEach, describe, expect, it, vi } from 'vitest';

afterEach(() => { vi.unstubAllGlobals(); vi.unstubAllEnvs(); vi.resetModules(); });

const viewer = id => ({ id, roles: [{ name: 'student' }] });
const event = token => ({
 url: new URL('http://localhost/dashboard/student/my-page'),
 cookies: { get: () => token }, locals: {}
});
const deferred = () => {
 let resolve;
 const promise = new Promise(complete => { resolve = complete; });
 return { promise, resolve };
};

describe('dashboard SSR adversarial responses', () => {
 it.each([
  ['401', () => new Response('unauthorized', { status: 401 })],
  ['破損した認証JSON', () => new Response('{invalid')],
  ['認証の通信エラー', () => Promise.reject(new Error('connection lost'))]
 ])('大会取得が成功しても%sならSSRを実行しない', async (_name, authResponse) => {
  vi.stubEnv('BACKEND_URL', 'http://127.0.0.1:8080');
  vi.stubGlobal('fetch', vi.fn(url => String(url).endsWith('/api/auth/user')
   ? authResponse() : new Response(JSON.stringify({ event_id: 7 }))));
  const { handle } = await import('$src/hooks.server.js');
  const current = event('expired');
  const resolve = vi.fn();
  const response = await handle({ event: current, resolve });
  expect(response.status).toBe(302);
  expect(response.headers.get('location')).toBe('http://localhost/');
  expect(current.locals.user).toBeNull();
  expect(current.locals.activeEvent).toBeUndefined();
  expect(resolve).not.toHaveBeenCalled();
 });

 it.each([
  ['破損した大会JSON', () => new Response('{invalid')],
  ['大会取得の通信エラー', () => Promise.reject(new Error('connection lost'))]
 ])('%sなら学生をメンテナンスへ送り、SSRに進まない', async (_name, activeResponse) => {
  vi.stubEnv('BACKEND_URL', 'http://127.0.0.1:8080');
  vi.stubGlobal('fetch', vi.fn(url => String(url).endsWith('/api/auth/user')
   ? new Response(JSON.stringify(viewer('alice'))) : activeResponse()));
  const { handle } = await import('$src/hooks.server.js');
  const current = event('alice');
  const resolve = vi.fn();
  const response = await handle({ event: current, resolve });
  expect(response.status).toBe(303);
  expect(response.headers.get('location')).toBe('http://localhost/dashboard/maintenance');
  expect(current.locals.activeEvent).toBeNull();
  expect(resolve).not.toHaveBeenCalled();
 });

 it('同時SSRの大会と認証が逆順に完了してもリクエスト間で混ざらない', async () => {
  vi.stubEnv('BACKEND_URL', 'http://127.0.0.1:8080');
  const reads = new Map();
  vi.stubGlobal('fetch', vi.fn((url, options) => {
   const read = deferred();
   reads.set(`${options.headers.cookie}:${new URL(url).pathname}`, read);
   return read.promise;
  }));
  const { handle } = await import('$src/hooks.server.js');
  const alice = event('alice');
  const bob = event('bob');
  const resolve = async current => new Response(JSON.stringify(current.locals));
  const aliceSSR = handle({ event: alice, resolve });
  const bobSSR = handle({ event: bob, resolve });
  expect(reads.size).toBe(4);
  reads.get('session_token=alice:/api/events/active').resolve(new Response(JSON.stringify({ event_id: 1, event_name: 'Alice event' })));
  reads.get('session_token=bob:/api/auth/user').resolve(new Response(JSON.stringify(viewer('bob'))));
  reads.get('session_token=bob:/api/events/active').resolve(new Response(JSON.stringify({ event_id: 2, event_name: 'Bob event' })));
  const bobResult = await (await bobSSR).json();
  expect(bobResult.user.id).toBe('bob');
  expect(bobResult.activeEvent).toEqual({ event_id: 2, event_name: 'Bob event' });
  reads.get('session_token=alice:/api/auth/user').resolve(new Response(JSON.stringify(viewer('alice'))));
  const aliceResult = await (await aliceSSR).json();
  expect(aliceResult.user.id).toBe('alice');
  expect(aliceResult.activeEvent).toEqual({ event_id: 1, event_name: 'Alice event' });
 });
});
