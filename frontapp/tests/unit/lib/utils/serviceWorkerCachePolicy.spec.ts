import { describe, expect, it } from 'vitest';
import { isApiPath, isCacheableStaticAsset } from '$lib/utils/serviceWorkerCachePolicy.js';

const origin = 'https://sports.example';
const assets = new Set(['/_app/immutable/app.js', '/icon-192.png', '/api/accidental.json']);

describe('isCacheableStaticAsset', () => {
	it('allows an exact same-origin build asset', () => {
		expect(isCacheableStaticAsset(`${origin}/_app/immutable/app.js`, origin, assets)).toBe(true);
	});

	it('allows a declared asset when the request has a query string', () => {
		expect(isCacheableStaticAsset(`${origin}/icon-192.png?v=2`, origin, assets)).toBe(true);
	});

	it('rejects API responses even if a conflicting static path is declared', () => {
		expect(isCacheableStaticAsset(`${origin}/api/accidental.json`, origin, assets)).toBe(false);
	});

	it('rejects authenticated application routes', () => {
		expect(isCacheableStaticAsset(`${origin}/dashboard/root`, origin, assets)).toBe(false);
	});

	it('rejects cross-origin assets', () => {
		expect(
			isCacheableStaticAsset('https://cdn.example/_app/immutable/app.js', origin, assets)
		).toBe(false);
	});
});

describe('isApiPath', () => {
	it('identifies the API root and descendants', () => {
		expect(isApiPath('/api')).toBe(true);
		expect(isApiPath('/api/root/db/export')).toBe(true);
		expect(isApiPath('/apiary')).toBe(false);
	});
});
