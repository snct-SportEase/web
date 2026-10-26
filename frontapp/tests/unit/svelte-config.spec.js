import { describe, expect, it } from 'vitest';
import config from '../../sveltekit.config.js';

describe('SvelteKit CSRF configuration', () => {
	it('keeps cross-origin form submissions blocked by default', () => {
		const csrf = config.csrf ?? {};
		const trustedOrigins = csrf.trustedOrigins ?? [];

		expect(csrf.checkOrigin).not.toBe(false);
		expect(trustedOrigins).not.toContain('*');
	});
});
