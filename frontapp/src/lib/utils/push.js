import { browser } from '$app/environment';
import { env as publicEnv } from '$env/dynamic/public';

const SUPPORTED_ROLES = new Set(['student', 'admin', 'root']);

let pushSubscriptionPromise = null;

export function userHasPushEligibleRole(user) {
	if (!user?.roles) {
		return false;
	}
	return user.roles.some((role) => SUPPORTED_ROLES.has(role.name));
}

export async function ensurePushSubscription() {
	if (!browser) {
		return { status: 'skipped', reason: 'not-browser' };
	}

	// PUBLIC_WEBPUSH_PUBLIC_KEY または PUBLIC_WEBPUSH_KEY のどちらかをサポート
	const vapidKey = publicEnv.PUBLIC_WEBPUSH_PUBLIC_KEY ?? publicEnv.PUBLIC_WEBPUSH_KEY ?? '';

	if (!vapidKey) {
		return { status: 'skipped', reason: 'missing-vapid-key' };
	}

	if (!('Notification' in window) || !('serviceWorker' in navigator) || !('PushManager' in window)) {
		return { status: 'skipped', reason: 'unsupported' };
	}

	if (Notification.permission === 'denied') {
		return { status: 'skipped', reason: 'permission-denied' };
	}

	if (Notification.permission === 'default') {
		const permission = await Notification.requestPermission();
		if (permission !== 'granted') {
			return { status: 'skipped', reason: 'permission-denied' };
		}
	}

	if (!pushSubscriptionPromise) {
		pushSubscriptionPromise = setupSubscription(vapidKey).finally(() => {
			pushSubscriptionPromise = null;
		});
	}

	return pushSubscriptionPromise;
}

async function setupSubscription(vapidKey) {
	try {
		const registration = await navigator.serviceWorker.ready;

		const applicationServerKey = urlBase64ToUint8Array(vapidKey);
		let subscription = await registration.pushManager.getSubscription();

		if (!subscription) {
			subscription = await registration.pushManager.subscribe({
				userVisibleOnly: true,
				applicationServerKey
			});
		}

		const body = JSON.stringify(subscription.toJSON());
		console.log('[push] 購読情報をサーバーに送信します:', subscription.endpoint);
		
		const response = await fetch('/api/notifications/subscription', {
			method: 'POST',
			headers: {
				'Content-Type': 'application/json'
			},
			credentials: 'include',
			body
		});

		if (!response.ok) {
			const text = await response.text();
			console.error('[push] 購読情報の保存に失敗しました:', response.status, text);
			throw new Error(`Failed to register push subscription: ${response.status} ${text}`);
		}

		const result = await response.json();
		console.log('[push] 購読情報の保存に成功しました:', result);
		return { status: 'subscribed' };
	} catch (error) {
		console.error('[push] push subscription failed', error);
		return { status: 'failed', reason: error.message };
	}
}

function urlBase64ToUint8Array(base64String) {
	const padding = '='.repeat((4 - (base64String.length % 4)) % 4);
	const base64 = (base64String + padding).replace(/-/g, '+').replace(/_/g, '/');

	const rawData = window.atob(base64);
	const outputArray = new Uint8Array(rawData.length);

	for (let i = 0; i < rawData.length; ++i) {
		outputArray[i] = rawData.charCodeAt(i);
	}
	return outputArray;
}

