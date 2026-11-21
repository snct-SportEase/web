import { writable } from 'svelte/store';

interface PushSubscriptionStatus {
	loaded: boolean;
	isSubscribed: boolean;
	isSupported: boolean;
	canEnable: boolean;
	permission: NotificationPermission | 'unsupported';
	vapidKeySet: boolean;
}

export const pushSubscriptionStatus = writable<PushSubscriptionStatus>({
	loaded: false,
	isSubscribed: false,
	isSupported: false,
	canEnable: false,
	permission: 'default',
	vapidKeySet: false
});

export function updatePushSubscriptionStatus(nextStatus: Partial<PushSubscriptionStatus>): void {
	pushSubscriptionStatus.update((current) => ({
		...current,
		...nextStatus,
		loaded: true
	}));
}
