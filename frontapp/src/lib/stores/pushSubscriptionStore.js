import { writable } from 'svelte/store';

export const pushSubscriptionStatus = writable({
	loaded: false,
	isSubscribed: false,
	isSupported: false,
	canEnable: false,
	permission: 'default',
	vapidKeySet: false
});

export function updatePushSubscriptionStatus(nextStatus) {
	pushSubscriptionStatus.update((current) => ({
		...current,
		...nextStatus,
		loaded: true
	}));
}
