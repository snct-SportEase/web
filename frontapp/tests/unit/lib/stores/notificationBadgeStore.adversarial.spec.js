import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { get } from 'svelte/store';

vi.mock('$app/environment', () => ({ browser: true }));

const alice = { id: 'alice', roles: [{ name: 'student' }] };
const bob = { id: 'bob', roles: [{ name: 'student' }] };
const response = notifications => ({ ok: true, json: async () => ({ notifications }) });
const deferred = () => {
 let resolve;
 const promise = new Promise(complete => { resolve = complete; });
 return { promise, resolve };
};

beforeEach(() => {
 const values = new Map();
 vi.stubGlobal('localStorage', {
  getItem: vi.fn(key => values.get(key) ?? null),
  setItem: vi.fn((key, value) => values.set(key, value))
 });
});
afterEach(() => { vi.unstubAllGlobals(); vi.resetModules(); });

describe('notification badge adversarial interleavings', () => {
 it.each(['success', '503'])('追加取得が%sでも開始後に届いたPushを反映する', async followupStatus => {
  const { refreshNotificationBadge, notificationBadgeCount } = await import('$lib/stores/notificationBadgeStore.js');
  const initialRead = deferred();
  const followupRead = deferred();
  const followupStarted = deferred();
  const fetcher = vi.fn()
   .mockImplementationOnce(() => initialRead.promise)
   .mockImplementationOnce(() => { followupStarted.resolve(); return followupRead.promise; })
   .mockResolvedValueOnce(response([{ id: 1 }, { id: 2 }, { id: 3 }]));
  const initial = refreshNotificationBadge(alice, { fetcher });
  const firstPush = refreshNotificationBadge(alice, { fetcher, force: true });
  initialRead.resolve(response([{ id: 1 }]));
  await followupStarted.promise;
  const latePush = refreshNotificationBadge(alice, { fetcher, force: true });
  const simultaneousPush = refreshNotificationBadge(alice, { fetcher, force: true });
  followupRead.resolve(followupStatus === '503' ? { ok: false, status: 503 } : response([{ id: 1 }, { id: 2 }]));
  await initial;
  await firstPush;
  expect(await latePush).toBe(3);
  expect(await simultaneousPush).toBe(3);
  expect(fetcher).toHaveBeenCalledTimes(3);
  expect(get(notificationBadgeCount)).toBe(3);
 });

 it('先行取得が503でも待機中のPush更新は成功できる', async () => {
  const { refreshNotificationBadge, notificationBadgeCount } = await import('$lib/stores/notificationBadgeStore.js');
  const firstRead = deferred();
  const fetcher = vi.fn()
   .mockImplementationOnce(() => firstRead.promise)
   .mockResolvedValueOnce(response([{ id: 2 }]));
  const initial = refreshNotificationBadge(alice, { fetcher }).catch(error => error);
  const push = refreshNotificationBadge(alice, { fetcher, force: true });
  firstRead.resolve({ ok: false, status: 503 });
  expect(await initial).toBeInstanceOf(Error);
  expect(await push).toBe(1);
  expect(fetcher).toHaveBeenCalledTimes(2);
  expect(get(notificationBadgeCount)).toBe(1);
 });

 it('ログアウト後に完了した通知取得でバッジを復活させない', async () => {
  const { refreshNotificationBadge, notificationBadgeCount } = await import('$lib/stores/notificationBadgeStore.js');
  const oldRead = deferred();
  const pending = refreshNotificationBadge(alice, { fetcher: () => oldRead.promise });
  expect(await refreshNotificationBadge(null)).toBe(0);
  oldRead.resolve(response([{ id: 1 }]));
  expect(await pending).toBe(0);
  expect(get(notificationBadgeCount)).toBe(0);
  expect(localStorage.setItem).not.toHaveBeenCalled();
 });

 it('ユーザー変更前に待機したPushが別ユーザーのsnapshotと既読を変更しない', async () => {
  const { refreshNotificationBadge, seedNotificationSnapshot, notificationBadgeCount } = await import('$lib/stores/notificationBadgeStore.js');
  const oldRead = deferred();
  const fetcher = vi.fn()
   .mockImplementationOnce(() => oldRead.promise)
   .mockResolvedValueOnce(response([{ id: 'alice-only' }]));
  const initial = refreshNotificationBadge(alice, { fetcher, initializeSeen: true });
  const queuedPush = refreshNotificationBadge(alice, { fetcher, force: true, initializeSeen: true });
  seedNotificationSnapshot(bob, [{ id: 'bob-1' }, { id: 'bob-2' }]);
  const bobFetch = vi.fn();
  expect(await refreshNotificationBadge(bob, { fetcher: bobFetch })).toBe(2);
  oldRead.resolve(response([{ id: 'old-alice' }]));
  expect(await initial).toBe(0);
  expect(await queuedPush).toBe(0);
  expect(get(notificationBadgeCount)).toBe(2);
  expect(await refreshNotificationBadge(bob, { fetcher: bobFetch })).toBe(2);
  expect(bobFetch).not.toHaveBeenCalled();
  expect(localStorage.getItem('sportease-notification-seen:alice')).toBeNull();
  expect(localStorage.getItem('sportease-notification-seen:bob')).toBeNull();
 });

 it('JSON解析失敗後に進行中の取得を残さず再試行する', async () => {
  const { refreshNotificationBadge } = await import('$lib/stores/notificationBadgeStore.js');
  const fetcher = vi.fn()
   .mockResolvedValueOnce({ ok: true, json: async () => { throw new SyntaxError('invalid JSON'); } })
   .mockResolvedValueOnce(response([{ id: 1 }]));
  await expect(refreshNotificationBadge(alice, { fetcher })).rejects.toThrow('invalid JSON');
  expect(await refreshNotificationBadge(alice, { fetcher })).toBe(1);
  expect(fetcher).toHaveBeenCalledTimes(2);
 });

 it('前ユーザーの遅い既読処理で現在のユーザーのバッジを消さない', async () => {
  const { refreshNotificationBadge, seedNotificationSnapshot, markNotificationsSeen, notificationBadgeCount } = await import('$lib/stores/notificationBadgeStore.js');
  seedNotificationSnapshot(bob, [{ id: 'bob-1' }, { id: 'bob-2' }]);
  expect(await refreshNotificationBadge(bob, { fetcher: vi.fn() })).toBe(2);
  markNotificationsSeen(alice, [{ id: 'alice-only' }]);
  expect(get(notificationBadgeCount)).toBe(2);
  expect(localStorage.getItem('sportease-notification-seen:bob')).toBeNull();
  expect(JSON.parse(localStorage.getItem('sportease-notification-seen:alice'))).toEqual(['alice-only']);
  markNotificationsSeen(alice);
  expect(get(notificationBadgeCount)).toBe(2);
  markNotificationsSeen(bob);
  expect(get(notificationBadgeCount)).toBe(0);
  expect(JSON.parse(localStorage.getItem('sportease-notification-seen:bob'))).toEqual(['bob-1', 'bob-2']);
 });

 it('破損した既読データでsnapshot全体を既読にしない', async () => {
  const { refreshNotificationBadge, seedNotificationSnapshot } = await import('$lib/stores/notificationBadgeStore.js');
  localStorage.setItem('sportease-notification-seen:alice', '{invalid');
  seedNotificationSnapshot(alice, [{ id: 1 }, { id: 2 }]);
  expect(await refreshNotificationBadge(alice, { fetcher: vi.fn(), initializeSeen: true })).toBe(2);
  expect(localStorage.getItem('sportease-notification-seen:alice')).toBe('{invalid');
 });
});
