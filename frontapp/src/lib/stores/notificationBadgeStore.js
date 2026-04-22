import { browser } from '$app/environment';
import { writable } from 'svelte/store';

const MAX_TRACKED_NOTIFICATIONS = 200;

export const notificationBadgeCount = writable(0);

let latestNotificationIds = [];

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

export async function refreshNotificationBadge(user, { initializeSeen = false, fetcher = fetch } = {}) {
  if (!browser || !canUseNotificationBadge(user)) {
    notificationBadgeCount.set(0);
    return 0;
  }

  const response = await fetcher('/api/notifications?limit=50');
  if (!response.ok) {
    throw new Error(`Failed to fetch notifications: ${response.status}`);
  }

  const result = await response.json();
  const notifications = Array.isArray(result.notifications) ? result.notifications : [];
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
    : latestNotificationIds;

  saveSeenIds(user, [...idsToMark, ...getSeenIds(user)]);
  notificationBadgeCount.set(0);
}
