import { describe, expect, it } from 'vitest';
import { applyCSRFProxyProtection } from '$lib/server/csrfProxy.js';
import { createBackendSessionHeaders } from '$lib/server/backendSessionHeaders.js';

describe('applyCSRFProxyProtection', () => {
	it('replaces an incoming token for a same-origin state-changing request', () => {
		const headers = new Headers({ 'x-csrf-token': 'attacker-token' });

		applyCSRFProxyProtection(headers, {
			method: 'POST',
			requestOrigin: 'https://sports.example',
			expectedOrigin: 'https://sports.example',
			csrfToken: 'trusted-token'
		});

		expect(headers.get('x-csrf-token')).toBe('trusted-token');
	});

	it('rejects a same-site but cross-origin request', () => {
		expect(() =>
			applyCSRFProxyProtection(new Headers(), {
				method: 'PUT',
				requestOrigin: 'https://evil.sports.example',
				expectedOrigin: 'https://sports.example',
				csrfToken: 'trusted-token'
			})
		).toThrow('invalid Origin');
	});

	it('rejects a state-changing request without a bound token', () => {
		expect(() =>
			applyCSRFProxyProtection(new Headers(), {
				method: 'DELETE',
				requestOrigin: 'https://sports.example',
				expectedOrigin: 'https://sports.example',
				csrfToken: undefined
			})
		).toThrow('no CSRF token');
	});

	it('removes an unnecessary client-provided token from safe requests', () => {
		const headers = new Headers({ 'x-csrf-token': 'attacker-token' });
		applyCSRFProxyProtection(headers, {
			method: 'GET',
			requestOrigin: null,
			expectedOrigin: 'https://sports.example',
			csrfToken: undefined
		});
		expect(headers.has('x-csrf-token')).toBe(false);
	});
});

describe('createBackendSessionHeaders', () => {
	it('forwards both credentials and copies the CSRF token to the header', () => {
		const cookies = {
			get(name) {
				return { session_token: 'session-value', csrf_token: 'csrf-value' }[name];
			}
		};

		const headers = createBackendSessionHeaders(cookies, { 'content-type': 'application/json' });

		expect(headers.get('cookie')).toBe('session_token=session-value; csrf_token=csrf-value');
		expect(headers.get('x-csrf-token')).toBe('csrf-value');
		expect(headers.get('content-type')).toBe('application/json');
	});
});
