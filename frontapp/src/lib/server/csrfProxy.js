const SAFE_METHODS = new Set(['GET', 'HEAD', 'OPTIONS', 'TRACE']);

function normalizedOrigin(value) {
	try {
		return value ? new URL(value).origin : null;
	} catch {
		return null;
	}
}

/**
 * Inject the session-bound CSRF token only for browser requests that prove
 * they originated from this exact origin. Incoming CSRF headers are discarded.
 *
 * @param {Headers} headers
 * @param {{ method: string, requestOrigin: string | null, expectedOrigin: string, csrfToken: string | undefined }} options
 */
export function applyCSRFProxyProtection(
	headers,
	{ method, requestOrigin, expectedOrigin, csrfToken }
) {
	headers.delete('x-csrf-token');
	if (SAFE_METHODS.has(method.toUpperCase())) {
		return headers;
	}

	if (normalizedOrigin(requestOrigin) !== normalizedOrigin(expectedOrigin)) {
		throw new Error('State-changing API request has an invalid Origin');
	}
	if (!csrfToken) {
		throw new Error('State-changing API request has no CSRF token');
	}

	headers.set('x-csrf-token', csrfToken);
	return headers;
}
