/// <reference types="@sveltejs/kit" />
import { build, files, version } from '$service-worker';
import { isApiPath, isCacheableStaticAsset } from '$lib/utils/serviceWorkerCachePolicy.js';

// This cache must contain public static assets only. Never put application
// routes or API responses in Cache Storage because cache keys are not scoped
// to the authenticated user.
const CACHE = `static-assets-v2-${version}`;

const ASSETS = [
	...build, // represents all code and data needed to render the routes of your app
	...files  // represents all static assets in your static directory
].filter((pathname) => !isApiPath(pathname));
const ASSET_PATHS = new Set(ASSETS);

self.addEventListener('install', (event) => {
	// Create a new cache and add all files to it
	async function addFilesToCache() {
		const cache = await caches.open(CACHE);
		await cache.addAll(ASSETS);
	}

	// Skip waiting to activate the new service worker immediately
	event.waitUntil(
		addFilesToCache().then(() => {
			return self.skipWaiting();
		})
	);
});

self.addEventListener('activate', (event) => {
	// Remove previous cached data from disk
	async function deleteOldCaches() {
		for (const key of await caches.keys()) {
			if (key !== CACHE) await caches.delete(key);
		}
	}

	// Claim all clients to take control immediately
	event.waitUntil(
		deleteOldCaches().then(() => {
			return self.clients.claim();
		})
	);
});

self.addEventListener('fetch', (event) => {
	// Let the browser handle every request except exact, same-origin static
	// assets. In particular, never intercept or cache /api requests,
	// authenticated pages, navigations, exports, or downloads.
	if (event.request.method !== 'GET') return;
	if (event.request.mode === 'navigate') return;

	const url = new URL(event.request.url);
	if (!isCacheableStaticAsset(url, self.location.origin, ASSET_PATHS)) return;

	event.respondWith(
		caches.open(CACHE).then(async (cache) => {
			const cachedResponse = await cache.match(url.pathname);
			if (cachedResponse) return cachedResponse;

			const response = await fetch(event.request);
			if (response.ok) {
				await cache.put(url.pathname, response.clone());
			}
			return response;
		})
	);
});

self.addEventListener('push', (event) => {
	if (!event.data) {
		return;
	}

	let payload;
	const textData = event.data.text();
	try {
		payload = JSON.parse(textData);
	} catch (error) {
		console.error('[service-worker] push payload JSON parse error', error);
		payload = {
			title: '新しい通知',
			body: textData
		};
	}

	const title = payload.title || '新しい通知';
	const options = {
		body: payload.body || '',
		data: payload.data || {},
		tag: 'sportease-notification',
		badge: '/icon-96.png',
		icon: '/icon-192.png',
		requireInteraction: false,
		silent: false,
		renotify: true
	};

	event.waitUntil(self.registration.showNotification(title, options));
	event.waitUntil(
		self.clients.matchAll({ type: 'window', includeUncontrolled: true }).then((clientList) => {
			for (const client of clientList) {
				client.postMessage({ type: 'sportease:new-notification' });
			}
		})
	);
});

self.addEventListener('notificationclick', (event) => {
	event.notification.close();

	const targetUrl = '/dashboard/student/notification';
	event.waitUntil(
		self.clients.matchAll({ type: 'window', includeUncontrolled: true }).then((clientList) => {
			for (const client of clientList) {
				if ('focus' in client) {
					client.focus();
					return;
				}
			}
			if (self.clients.openWindow) {
				return self.clients.openWindow(targetUrl);
			}
		})
	);
});
