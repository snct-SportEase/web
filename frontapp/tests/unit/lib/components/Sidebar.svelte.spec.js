import { page } from 'vitest/browser';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { render } from 'vitest-browser-svelte';
import Sidebar from '$src/lib/components/Sidebar.svelte';
import { isPWAInstalled } from '$lib/utils/pwa.js';
import { closePWAInstallDialog, pwaInstallDialogOpen } from '$lib/stores/pwaInstallStore.js';
import { get } from 'svelte/store';

vi.mock('$env/dynamic/public', () => ({
	env: {
		PUBLIC_WEBPUSH_PUBLIC_KEY: ''
	}
}));

vi.mock('$lib/utils/pwa.js', () => ({
	isPWAInstalled: vi.fn()
}));

const studentUser = {
	id: 'student-1',
	email: 'student@example.com',
	roles: [{ name: 'student' }]
};

describe('Sidebar PWA status', () => {
	beforeEach(() => {
		closePWAInstallDialog();
		vi.mocked(isPWAInstalled).mockReturnValue(false);
		vi.stubGlobal(
			'fetch',
			vi.fn(async () => ({
				ok: true,
				json: async () => ({ notifications: [] })
			}))
		);
	});

	it('ブラウザ表示ではPWA未設定ラベルを表示する', async () => {
		render(Sidebar, { user: studentUser });

		const pwaSetupButton = page.getByRole('button', { name: 'PWA未設定', exact: true });
		await expect.element(pwaSetupButton).toBeInTheDocument();
		await pwaSetupButton.click();
		expect(get(pwaInstallDialogOpen)).toBe(true);
	});

	it('PWAとして起動している場合はPWA未設定ラベルを表示しない', async () => {
		vi.mocked(isPWAInstalled).mockReturnValue(true);

		render(Sidebar, { user: studentUser });

		await expect
			.element(page.getByRole('button', { name: 'PWA未設定', exact: true }))
			.not.toBeInTheDocument();
	});
});
