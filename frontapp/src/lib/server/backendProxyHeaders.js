const HOP_BY_HOP_HEADERS = [
	'connection',
	'keep-alive',
	'proxy-authenticate',
	'proxy-authorization',
	'te',
	'trailer',
	'transfer-encoding',
	'upgrade',
	'host',
	'content-length'
];

/**
 * Build headers for requests sent to the backend.
 *
 * The client address must come from SvelteKit's trusted adapter configuration.
 * Never forward client-supplied forwarding headers across this trust boundary.
 *
 * @param {Headers} requestHeaders
 * @param {{ clientAddress: string, host: string, protocol: string }} forwarding
 */
export function createBackendProxyHeaders(requestHeaders, { clientAddress, host, protocol }) {
	const normalizedClientAddress = clientAddress?.trim();
	if (!normalizedClientAddress) {
		throw new Error('Trusted client address is unavailable');
	}

	const headers = new Headers(requestHeaders);
	const connectionHeaders = (headers.get('connection') ?? '')
		.split(',')
		.map((name) => name.trim())
		.filter(Boolean);

	for (const header of HOP_BY_HOP_HEADERS) {
		headers.delete(header);
	}
	for (const header of connectionHeaders) {
		headers.delete(header);
	}
	for (const header of [...headers.keys()]) {
		if (header === 'forwarded' || header === 'x-real-ip' || header.startsWith('x-forwarded-')) {
			headers.delete(header);
		}
	}

	headers.set('x-forwarded-for', normalizedClientAddress);
	headers.set('x-real-ip', normalizedClientAddress);
	headers.set('x-forwarded-host', host);
	headers.set('x-forwarded-proto', protocol);

	return headers;
}
