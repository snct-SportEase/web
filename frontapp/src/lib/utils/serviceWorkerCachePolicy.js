/**
 * Returns true only for a same-origin, predeclared public static asset.
 * API paths are denied explicitly even if a conflicting static file is added.
 */
export function isCacheableStaticAsset(requestUrl, serviceWorkerOrigin, assetPaths) {
	const url = new URL(requestUrl);
	if (url.origin !== serviceWorkerOrigin) return false;
	if (isApiPath(url.pathname)) return false;
	return assetPaths.has(url.pathname);
}

export function isApiPath(pathname) {
	return pathname === '/api' || pathname.startsWith('/api/');
}
