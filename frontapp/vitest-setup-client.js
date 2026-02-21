/// <reference types="@vitest/browser/matchers" />
/// <reference types="@vitest/browser/providers/playwright" />
// Suppress transport disconnection errors that occur during test cleanup
import { afterAll } from 'vitest';

afterAll(() => {
	// Allow time for cleanup operations before process exit
	return new Promise(resolve => setTimeout(resolve, 100));
});