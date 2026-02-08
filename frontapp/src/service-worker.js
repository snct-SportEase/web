/// <reference types="@sveltejs/kit" />
import { build, files, version } from '$service-worker';

const CACHE = `cache-${version}`;

const ASSETS = [
	...build, // represents all code and data needed to render the routes of your app
	...files  // represents all static assets in your static directory
];

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
	// ignore POST requests etc
	if (event.request.method !== 'GET') return;

	async function respond() {
		const url = new URL(event.request.url);
		
		// Skip caching for unsupported schemes (chrome-extension, etc.)
		if (url.protocol !== 'http:' && url.protocol !== 'https:') {
			return fetch(event.request);
		}

		const cache = await caches.open(CACHE);

		// `build`/`files` can always be served from the cache
		if (ASSETS.includes(url.pathname)) {
			const response = await cache.match(url.pathname);

			if (response) {
				return response;
			}
		}

		// Identify critical API endpoints for Stale-While-Revalidate strategy
		// These are data that users might want to see even if offline (schedule, tournaments, etc.)
		const isCriticalApi = 
			url.pathname.match(/^\/api\/student\/events\/[^/]+\/tournaments/) ||
			url.pathname.match(/^\/api\/student\/events\/[^/]+\/noon-game\/session/) ||
			url.pathname.match(/^\/api\/root\/events/) ||
			url.pathname.match(/^\/api\/student\/class-info/);

		if (isCriticalApi) {
			// Stale-While-Revalidate strategy
			// 1. Return from cache immediately if available
			// 2. Fetch from network and update cache in background
			const cachedResponse = await cache.match(event.request);
			
			const networkFetch = fetch(event.request).then(response => {
				// Update cache if response is valid
				if (response.ok && (url.protocol === 'http:' || url.protocol === 'https:')) {
					cache.put(event.request, response.clone());
				}
				return response;
			}).catch(err => {
				console.log('Network fetch failed for critical API, using cache if available');
				throw err;
			});

			// If we have a cached response, return it immediately, but still trigger the network fetch to update cache
			if (cachedResponse) {
				// We don't await the network fetch here, we just start it
				// theoretically to update the cache for *next* time.
				// However, standard SW Stale-While-Revalidate usually returns cache and updates in background.
				// To make the UI update eventually, the app would need to handle updates, 
				// but for offline availability, returning stale cache is the priority.
				return cachedResponse;
			}
			
			// If no cache, wait for network
			try {
				return await networkFetch;
			} catch (err) {
				// If both cache and network fail, we can't do much for API calls
				// maybe return a fallback JSON if needed, but for now just throw
				throw err;
			}
		}

		// for everything else, try the network first, but
		// fall back to the cache if we're offline
		try {
			const response = await fetch(event.request);

			// if we're offline, fetch can return a value that is not a Response
			// instead of throwing - and we can't pass this value to cache.put()
			if (!(response instanceof Response)) {
				throw new Error('invalid response from fetch');
			}

			// Only cache successful responses with http/https protocol
			if (response.status === 200 && (url.protocol === 'http:' || url.protocol === 'https:')) {
				cache.put(event.request, response.clone());
			}

			return response;
		} catch (err) {
			const response = await cache.match(event.request);

			if (response) {
				return response;
			}

			// if there's no cache, throw an error
			// so that the browser shows its offline page
			throw err;
		}
	}

	event.respondWith(respond());
});

self.addEventListener('push', (event) => {
	if (!event.data) {
		return;
	}

	let payload;
	try {
		payload = event.data.json();
	} catch (error) {
		console.error('[service-worker] push payload JSON parse error', error);
		payload = {
			title: '新しい通知',
			body: event.data.text()
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
