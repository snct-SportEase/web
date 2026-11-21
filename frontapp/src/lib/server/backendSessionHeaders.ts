/**
 * Build credentials for a direct server-to-backend request.
 *
 */
export function createBackendSessionHeaders(cookies: Pick<import('@sveltejs/kit').Cookies, 'get'>, initialHeaders?: HeadersInit): Headers {
	const headers = new Headers(initialHeaders);
	const sessionToken = cookies.get('session_token');
	const csrfToken = cookies.get('csrf_token');
	const cookieValues = [];

	if (sessionToken) {
		cookieValues.push(`session_token=${sessionToken}`);
	}
	if (csrfToken) {
		cookieValues.push(`csrf_token=${csrfToken}`);
		headers.set('x-csrf-token', csrfToken);
	}
	if (cookieValues.length > 0) {
		headers.set('cookie', cookieValues.join('; '));
	}

	return headers;
}
