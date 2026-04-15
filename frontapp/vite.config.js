import tailwindcss from '@tailwindcss/vite';
import { defineConfig } from 'vitest/config';
import { sveltekit } from '@sveltejs/kit/vite';
import { createLogger } from 'vite';

const backendUrl = process.env.PUBLIC_BACKEND_URL || process.env.BACKEND_URL || 'http://localhost:8080';

const logger = createLogger();
const originalError = logger.error;
logger.error = (msg, options) => {
	// Suppress harmless ECONNRESET and EPIPE errors from Vite's WebSocket proxy during tests
	if (msg.includes('ws proxy') || msg.includes('EPIPE') || msg.includes('ECONNRESET')) {
		return;
	}

	originalError(msg, options);
};

export default defineConfig({
	plugins: [tailwindcss(), sveltekit()],
	customLogger: logger,
	server: {
		proxy: {
			'/api': {
				target: backendUrl,
				changeOrigin: true,
				ws: true
			}
		}
	},
	ssr: {
		noExternal: ['bracketry', 'marked', 'svelte-dnd-action', 'html2pdf.js']
	},
	test: {
		expect: { requireAssertions: true },
		projects: [
			{
				extends: './vite.config.js',
				test: {
					name: 'client',
					environment: 'browser',
					browser: {
						enabled: true,
						provider: 'playwright',
						instances: [{ browser: 'chromium' }],
						headless: true
					},
					include: ['src/**/*.svelte.{test,spec}.{js,ts}', 'tests/pwa.spec.js'],
					exclude: ['src/lib/server/**'],
					setupFiles: ['./vitest-setup-client.js'],
					// Timeout for browser operations
					testTimeout: 30000
				}
			},
			{
				extends: './vite.config.js',
				test: {
					name: 'server',
					environment: 'node',
					include: ['src/**/*.{test,spec}.{js,ts}'],
					exclude: ['src/**/*.svelte.{test,spec}.{js,ts}']
				}
			}
		]
	}
});
