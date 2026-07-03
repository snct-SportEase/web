import { describe, expect, it } from 'vitest';
import config from '../../svelte.config.js';

describe('SvelteKit CSRF configuration', () => {
	it('keeps cross-origin form submissions blocked by default', () => {
		const csrf = config.kit?.csrf ?? {};
		const trustedOrigins = csrf.trustedOrigins ?? [];

		expect(csrf.checkOrigin).not.toBe(false);
		expect(trustedOrigins).not.toContain('*');
	});
});
