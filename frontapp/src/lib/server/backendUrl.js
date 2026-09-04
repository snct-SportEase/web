const ALLOWED_BACKEND_HOSTNAMES = new Set([
	'sportease-backapp',
	'localhost',
	'127.0.0.1',
	'[::1]'
]);

export function resolveBackendOrigin(value) {
	if (!value) return null;

	const url = new URL(value);
	if (url.protocol !== 'http:' && url.protocol !== 'https:') {
		throw new Error('BACKEND_URL must use HTTP or HTTPS');
	}
	if (url.username || url.password) {
		throw new Error('BACKEND_URL must not contain credentials');
	}
	if (!ALLOWED_BACKEND_HOSTNAMES.has(url.hostname)) {
		throw new Error(`BACKEND_URL host is not allowed: ${url.hostname}`);
	}

	return url.origin;
}
