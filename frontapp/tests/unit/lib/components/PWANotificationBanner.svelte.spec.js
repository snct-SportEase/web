import { page } from '@vitest/browser/context';
import { describe, expect, it, vi } from 'vitest';
import { render } from 'vitest-browser-svelte';
import PWANotificationBanner from '$src/lib/components/PWANotificationBanner.svelte';

vi.mock('$lib/utils/pwa.js', () => ({
	isPWAInstalled: () => true,
	isPWAInstallable: () => false
}));

describe('PWANotificationBanner', () => {
	it('マウント後にshowがtrueになっても表示される', async () => {
		const view = render(PWANotificationBanner, { props: { show: false } });

		await view.rerender({ show: true });
		await new Promise((resolve) => setTimeout(resolve, 550));

		const banner = page.getByRole('alert');
		await expect.element(banner).toBeInTheDocument();
		await expect.element(banner).toHaveClass(/opacity-100/);
		await expect.element(banner).toHaveClass(/top-20/);
		await expect.element(banner).toHaveClass(/app-layer-notification/);
		await expect.element(banner).toHaveClass(/pointer-events-none/);
		await expect.element(page.getByRole('button', { name: '通知を閉じる' })).toHaveClass(/pointer-events-auto/);
	});
});
