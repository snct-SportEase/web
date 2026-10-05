import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { get } from 'svelte/store';

vi.mock('$app/environment', () => ({ browser: true }));
const student = { id: 'alice', roles: [{ name: 'student' }] };
const otherStudent = { id: 'bob', roles: [{ name: 'student' }] };

beforeEach(() => {
 const values = new Map();
 vi.stubGlobal('localStorage', {
  getItem: vi.fn(key => values.get(key) ?? null),
  setItem: vi.fn((key, value) => values.set(key, value))
 });
});
afterEach(() => { vi.unstubAllGlobals(); vi.resetModules(); vi.useRealTimers(); });

const response = notifications => ({ ok: true, json: async () => ({ notifications }) });

describe('notification snapshots', () => {
 it('SSRの50件分を利用し、初期表示で通知APIを再取得しない', async () => {
  const { seedNotificationSnapshot, refreshNotificationBadge, notificationBadgeCount } = await import('$lib/stores/notificationBadgeStore.js');
  const notifications = Array.from({ length: 50 }, (_, index) => ({ id: index + 1 }));
  localStorage.setItem('sportease-notification-seen:alice', JSON.stringify(['1']));
  const fetcher = vi.fn();
  seedNotificationSnapshot(student, notifications);
  expect(await refreshNotificationBadge(student, { fetcher })).toBe(49);
  expect(get(notificationBadgeCount)).toBe(49);
  expect(fetcher).not.toHaveBeenCalled();
 });
 it('ユーザー変更時に前のユーザーのsnapshotを使用しない', async () => {
  const { seedNotificationSnapshot, refreshNotificationBadge } = await import('$lib/stores/notificationBadgeStore.js');
  seedNotificationSnapshot(student, [{ id: 1 }]);
  const fetcher = vi.fn(async () => response([{ id: 2 }, { id: 3 }]));
  expect(await refreshNotificationBadge(otherStudent, { fetcher })).toBe(2);
  expect(fetcher).toHaveBeenCalledTimes(1);
 });
 it('同時更新を共有し、Push受信時の明示更新はsnapshotを再取得する', async () => {
  const { refreshNotificationBadge } = await import('$lib/stores/notificationBadgeStore.js');
  const fetcher = vi.fn(async () => response([{ id: 1 }]));
  await Promise.all([refreshNotificationBadge(student, { fetcher }), refreshNotificationBadge(student, { fetcher })]);
  expect(fetcher).toHaveBeenCalledTimes(1);
  await refreshNotificationBadge(student, { fetcher, force: true });
  expect(fetcher).toHaveBeenCalledTimes(2);
 });
 it('古いリクエストは新しいSSR snapshotを上書きしない', async () => {
  const { seedNotificationSnapshot, refreshNotificationBadge } = await import('$lib/stores/notificationBadgeStore.js');
  let resolveFetch;
  const fetcher = vi.fn(() => new Promise(resolve => { resolveFetch = resolve; }));
  const pending = refreshNotificationBadge(student, { fetcher });
  seedNotificationSnapshot(student, [{ id: 2 }, { id: 3 }]);
  resolveFetch(response([{ id: 1 }]));
  expect(await pending).toBe(2);
 });
 it('TTL経過後は通知APIを再取得する', async () => {
  vi.useFakeTimers();
  const { seedNotificationSnapshot, refreshNotificationBadge } = await import('$lib/stores/notificationBadgeStore.js');
  seedNotificationSnapshot(student, [{ id: 1 }]);
  vi.advanceTimersByTime(15_001);
  const fetcher = vi.fn(async () => response([]));
  expect(await refreshNotificationBadge(student, { fetcher })).toBe(0);
  expect(fetcher).toHaveBeenCalledTimes(1);
 });
 it('ユーザー変更前の遅いレスポンスで現在の通知バッジを上書きしない', async () => {
  const { refreshNotificationBadge, notificationBadgeCount } = await import('$lib/stores/notificationBadgeStore.js');
  let resolveAlice;
  const oldRequest = refreshNotificationBadge(student, { fetcher: () => new Promise(resolve => { resolveAlice = resolve; }) });
  expect(await refreshNotificationBadge(otherStudent, { fetcher: async () => response([{ id: 2 }, { id: 3 }]) })).toBe(2);
  resolveAlice(response([{ id: 1 }]));
  await oldRequest;
  expect(get(notificationBadgeCount)).toBe(2);
 });
 it('APIエラーをキャッシュせず、次回更新で再取得する', async () => {
  const { refreshNotificationBadge } = await import('$lib/stores/notificationBadgeStore.js');
  const fetcher = vi.fn().mockResolvedValueOnce({ ok: false, status: 503 }).mockResolvedValueOnce(response([{ id: 1 }]));
  await expect(refreshNotificationBadge(student, { fetcher })).rejects.toThrow('503');
  expect(await refreshNotificationBadge(student, { fetcher })).toBe(1);
  expect(fetcher).toHaveBeenCalledTimes(2);
 });

});
