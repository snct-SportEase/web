import { browser } from '$app/env';
import { writable } from 'svelte/store';

const MAX_TRACKED_NOTIFICATIONS = 200;

type BadgeUser = { id?: string | number; email?: string; roles?: { name: string }[] } | null | undefined;
type BadgeNotification = { id?: string | number | null; title?: string; created_at?: string };
type BadgeSnapshot = { userKey: string; notifications: BadgeNotification[]; loadedAt: number };
type ForcedRefresh = { started: boolean; repeat: boolean; promise: Promise<BadgeNotification[]> };

export const notificationBadgeCount = writable(0);

let latestNotificationIds: string[] = [];
let latestUserKey: string | null = null;
let snapshot: BadgeSnapshot | null = null;
const pending = new Map<string, Promise<BadgeNotification[]>>();
const forcedPending = new Map<string, ForcedRefresh>();
let snapshotRevision = 0;
const SNAPSHOT_TTL = 15_000;

// Called with SSR data only in the browser. Snapshots are scoped to the viewer.
export function seedNotificationSnapshot(user: BadgeUser, notifications: BadgeNotification[] | null | undefined): void {
  if (!browser || !canUseNotificationBadge(user) || !Array.isArray(notifications)) return;
  snapshotRevision += 1;
  snapshot = { userKey: getUserKey(user), notifications, loadedAt: Date.now() };
  latestUserKey = snapshot.userKey;
  latestNotificationIds = notifications.map(getNotificationId);
}

async function getNotifications(user: BadgeUser, fetcher: typeof fetch, force: boolean): Promise<BadgeNotification[]> {
  const userKey = getUserKey(user);
  if (!force && snapshot?.userKey === userKey && Date.now() - snapshot.loadedAt < SNAPSHOT_TTL) {
    return snapshot.notifications;
  }
  if (pending.has(userKey)) {
    if (!force) return pending.get(userKey)!;
    const queued = forcedPending.get(userKey);
    if (queued) {
      if (queued.started) queued.repeat = true;
      return queued.promise;
    }
    // A Push can arrive after an existing read started. Read again once it ends.
    const previous = pending.get(userKey);
    const refresh: ForcedRefresh = { started: false, repeat: false, promise: Promise.resolve([]) };
    refresh.promise = (async () => {
      try { await previous; } catch { /* Retry even if the earlier read failed. */ }
      let notifications: BadgeNotification[] = [];
      do {
        refresh.started = true;
        refresh.repeat = false;
        try {
          notifications = await getNotifications(user, fetcher, true);
        } catch (error) {
          if (!refresh.repeat) throw error;
        }
      } while (refresh.repeat);
      return notifications;
    })();
    forcedPending.set(userKey, refresh);
    try { return await refresh.promise; }
    finally { forcedPending.delete(userKey); }
  }
  const revision = snapshotRevision;
  const request = (async () => {
    const response = await fetcher('/api/notifications?limit=50');
    if (!response.ok) throw new Error(`Failed to fetch notifications: ${response.status}`);
    const result = await response.json();
    const notifications = Array.isArray(result.notifications) ? result.notifications : [];
    if (snapshotRevision === revision && latestUserKey === userKey) seedNotificationSnapshot(user, notifications);
    return snapshot?.userKey === userKey ? snapshot.notifications : notifications;
  })();
  pending.set(userKey, request);
  try { return await request; }
  finally { pending.delete(userKey); }
}

function canUseNotificationBadge(user: BadgeUser): boolean {
  return Boolean(user?.roles?.some((role) => ['student', 'admin', 'root'].includes(role.name)));
}

function getUserKey(user: BadgeUser): string {
  return `sportease-notification-seen:${user?.id ?? user?.email ?? 'anonymous'}`;
}

function getNotificationId(notification: BadgeNotification): string {
  if (notification?.id !== undefined && notification?.id !== null) {
    return String(notification.id);
  }

  return `${notification?.title ?? ''}:${notification?.created_at ?? ''}`;
}

function getSeenIds(user: BadgeUser): string[] {
  if (!browser) return [];

  try {
    const raw = getStoredSeenIds(user);
    const parsed = raw ? JSON.parse(raw) : [];
    return Array.isArray(parsed) ? parsed.map(String) : [];
  } catch {
    return [];
  }
}

function getStoredSeenIds(user: BadgeUser): string | null {
  if (!browser) return null;
  return localStorage.getItem(getUserKey(user));
}

function saveSeenIds(user: BadgeUser, ids: string[]): void {
  if (!browser) return;
  const uniqueIds = [...new Set(ids.map(String))].slice(0, MAX_TRACKED_NOTIFICATIONS);
  localStorage.setItem(getUserKey(user), JSON.stringify(uniqueIds));
}

export async function refreshNotificationBadge(user: BadgeUser, { initializeSeen = false, fetcher = fetch, force = false }: { initializeSeen?: boolean; fetcher?: typeof fetch; force?: boolean } = {}): Promise<number> {
  if (!browser || !canUseNotificationBadge(user)) {
    latestUserKey = null;
    notificationBadgeCount.set(0);
    return 0;
  }

  const userKey = getUserKey(user);
  latestUserKey = userKey;
  const notifications = await getNotifications(user, fetcher, force);
  if (latestUserKey !== userKey) return 0;
  latestNotificationIds = notifications.map(getNotificationId);

  const hasStoredSeenIds = getStoredSeenIds(user) !== null;
  const seenIds = getSeenIds(user);
  if (initializeSeen && !hasStoredSeenIds) {
    saveSeenIds(user, latestNotificationIds);
    notificationBadgeCount.set(0);
    return 0;
  }

  const seenSet = new Set(seenIds);
  const unreadCount = latestNotificationIds.filter((id) => !seenSet.has(id)).length;
  notificationBadgeCount.set(unreadCount);
  return unreadCount;
}

export function markNotificationsSeen(user: BadgeUser, notifications: BadgeNotification[] | null = null): void {
  if (!browser || !canUseNotificationBadge(user)) return;

  const idsToMark = Array.isArray(notifications)
    ? notifications.map(getNotificationId)
    : latestUserKey === getUserKey(user) ? latestNotificationIds : [];

  saveSeenIds(user, [...idsToMark, ...getSeenIds(user)]);
  if (latestUserKey === getUserKey(user)) notificationBadgeCount.set(0);
}
