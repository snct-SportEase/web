import { browser } from '$app/environment';
import { writable } from 'svelte/store';

const MAX_TRACKED_NOTIFICATIONS = 200;

export const notificationBadgeCount = writable(0);

let latestNotificationIds = [];
let latestUserKey = null;
let snapshot = null;
const pending = new Map();
let snapshotRevision = 0;
const SNAPSHOT_TTL = 15_000;

// Called with SSR data only in the browser. Snapshots are scoped to the viewer.
export function seedNotificationSnapshot(user, notifications) {
  if (!browser || !canUseNotificationBadge(user) || !Array.isArray(notifications)) return;
  snapshotRevision += 1;
  snapshot = { userKey: getUserKey(user), notifications, loadedAt: Date.now() };
  latestUserKey = snapshot.userKey;
  latestNotificationIds = notifications.map(getNotificationId);
}

async function getNotifications(user, fetcher, force) {
  const userKey = getUserKey(user);
  if (!force && snapshot?.userKey === userKey && Date.now() - snapshot.loadedAt < SNAPSHOT_TTL) {
    return snapshot.notifications;
  }
  if (pending.has(userKey)) return pending.get(userKey);
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

function canUseNotificationBadge(user) {
  return Boolean(user?.roles?.some((role) => ['student', 'admin', 'root'].includes(role.name)));
}

function getUserKey(user) {
  return `sportease-notification-seen:${user?.id ?? user?.email ?? 'anonymous'}`;
}

function getNotificationId(notification) {
  if (notification?.id !== undefined && notification?.id !== null) {
    return String(notification.id);
  }

  return `${notification?.title ?? ''}:${notification?.created_at ?? ''}`;
}

function getSeenIds(user) {
  if (!browser) return [];

  try {
    const raw = getStoredSeenIds(user);
    const parsed = raw ? JSON.parse(raw) : [];
    return Array.isArray(parsed) ? parsed.map(String) : [];
  } catch {
    return [];
  }
}

function getStoredSeenIds(user) {
  if (!browser) return null;
  return localStorage.getItem(getUserKey(user));
}

function saveSeenIds(user, ids) {
  if (!browser) return;
  const uniqueIds = [...new Set(ids.map(String))].slice(0, MAX_TRACKED_NOTIFICATIONS);
  localStorage.setItem(getUserKey(user), JSON.stringify(uniqueIds));
}

export async function refreshNotificationBadge(user, { initializeSeen = false, fetcher = fetch, force = false } = {}) {
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

export function markNotificationsSeen(user, notifications = null) {
  if (!browser || !canUseNotificationBadge(user)) return;

  const idsToMark = Array.isArray(notifications)
    ? notifications.map(getNotificationId)
    : latestUserKey === getUserKey(user) ? latestNotificationIds : [];

  saveSeenIds(user, [...idsToMark, ...getSeenIds(user)]);
  notificationBadgeCount.set(0);
}
