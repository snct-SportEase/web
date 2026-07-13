import { describe, expect, it } from 'vitest';
import { createBackendProxyHeaders } from '$lib/server/backendProxyHeaders.js';

describe('createBackendProxyHeaders', () => {
	it('replaces all client-supplied forwarding headers with trusted values', () => {
		const incoming = new Headers({
			cookie: 'session_token=example',
			forwarded: 'for=attacker.example;proto=http',
			'x-forwarded-for': '198.51.100.10',
			'x-forwarded-host': 'attacker.example',
			'x-forwarded-proto': 'http',
			'x-forwarded-port': '1234',
			'x-forwarded-custom': 'untrusted',
			'x-real-ip': '198.51.100.11'
		});

		const headers = createBackendProxyHeaders(incoming, {
			clientAddress: '203.0.113.7',
			host: 'sports.example',
			protocol: 'https'
		});

		expect(headers.get('x-forwarded-for')).toBe('203.0.113.7');
		expect(headers.get('x-real-ip')).toBe('203.0.113.7');
		expect(headers.get('x-forwarded-host')).toBe('sports.example');
		expect(headers.get('x-forwarded-proto')).toBe('https');
		expect(headers.has('forwarded')).toBe(false);
		expect(headers.has('x-forwarded-port')).toBe(false);
		expect(headers.has('x-forwarded-custom')).toBe(false);
		expect(headers.get('cookie')).toBe('session_token=example');
	});

	it('removes hop-by-hop headers named by Connection', () => {
		const incoming = new Headers({
			connection: 'keep-alive, x-remove-me',
			'x-remove-me': 'secret',
			upgrade: 'websocket'
		});

		const headers = createBackendProxyHeaders(incoming, {
			clientAddress: '2001:db8::1',
			host: 'sports.example',
			protocol: 'https'
		});

		expect(headers.has('connection')).toBe(false);
		expect(headers.has('x-remove-me')).toBe(false);
		expect(headers.has('upgrade')).toBe(false);
	});

	it('fails closed when the trusted client address is unavailable', () => {
		expect(() =>
			createBackendProxyHeaders(new Headers(), {
				clientAddress: '',
				host: 'sports.example',
				protocol: 'https'
			})
		).toThrow('Trusted client address is unavailable');
	});
});
