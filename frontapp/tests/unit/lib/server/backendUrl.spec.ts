import { describe, expect, it } from 'vitest';
import { resolveBackendOrigin } from '$lib/server/backendUrl.js';

describe('resolveBackendOrigin', () => {
	it.each([
		['http://sportease-backapp:8080', 'http://sportease-backapp:8080'],
		['http://localhost:8080/', 'http://localhost:8080'],
		['https://127.0.0.1:8443/api', 'https://127.0.0.1:8443'],
		['http://[::1]:8080', 'http://[::1]:8080']
	])('allows a trusted backend origin: %s', (value, expected) => {
		expect(resolveBackendOrigin(value)).toBe(expected);
	});

	it('returns null when the backend is not configured', () => {
		expect(resolveBackendOrigin(undefined)).toBeNull();
	});

	it.each([
		'https://example.com',
		'http://169.254.169.254/latest/meta-data',
		'file:///etc/passwd',
		'http://user:password@localhost:8080'
	])('rejects an untrusted backend URL: %s', (value) => {
		expect(() => resolveBackendOrigin(value)).toThrow();
	});
});
