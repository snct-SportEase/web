import tailwindcss from '@tailwindcss/vite';
import { defineConfig } from 'vitest/config';
import { playwright } from '@vitest/browser-playwright';
import { sveltekit } from '@sveltejs/kit/vite';
import { fileURLToPath } from 'node:url';

const backendUrl = process.env.PUBLIC_BACKEND_URL || process.env.BACKEND_URL || 'http://localhost:8080';
const chromiumExecutablePath = process.env.PLAYWRIGHT_CHROMIUM_EXECUTABLE_PATH;

export default defineConfig({
	plugins: [tailwindcss(), sveltekit()],
	resolve: {
		alias: {
			$src: fileURLToPath(new URL('./src', import.meta.url))
		}
	},
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
		include: ['html2pdf.js']
	},
	test: {
		expect: { requireAssertions: true },
		projects: [
			{
				extends: './vite.config.js',
				test: {
					name: 'client',
					browser: {
						enabled: true,
						provider: playwright(
							chromiumExecutablePath
								? { launchOptions: { executablePath: chromiumExecutablePath } }
								: undefined
						),
						instances: [{ browser: 'chromium' }],
						headless: true
					},
					include: ['tests/unit/**/*.svelte.{test,spec}.{js,ts}', 'tests/pwa.spec.js'],
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
					include: ['tests/unit/**/*.{test,spec}.{js,ts}'],
					exclude: ['tests/unit/**/*.svelte.{test,spec}.{js,ts}']
				}
			}
		]
	}
});
