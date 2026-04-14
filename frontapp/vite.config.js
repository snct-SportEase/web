import tailwindcss from '@tailwindcss/vite';
import { defineConfig } from 'vitest/config';
import { sveltekit } from '@sveltejs/kit/vite';

const backendUrl = process.env.PUBLIC_BACKEND_URL || process.env.BACKEND_URL || 'http://localhost:8080';

export default defineConfig({
	plugins: [tailwindcss(), sveltekit()],
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
	optimizeDeps: {
		// Disable pre-bundling during tests to avoid dependency scanning issues
		disabled: process.env.VITEST === 'true'
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
