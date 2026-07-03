import { page } from '@vitest/browser/context';
import { afterEach, describe, expect, it, vi } from 'vitest';
import { render } from 'vitest-browser-svelte';
import Page from './+page.svelte';

describe('/+page.svelte', () => {
	afterEach(() => {
		vi.restoreAllMocks();
		window.history.replaceState({}, '', '/');
	});

	it('should render h1', async () => {
		render(Page);
		
		const heading = page.getByRole('heading', { level: 1 });
		await expect.element(heading).toBeInTheDocument();
	});

	it('shows a retry message for invalid OAuth state', async () => {
		window.history.replaceState({}, '', '/?error=invalid_state');
		const alertSpy = vi.spyOn(window, 'alert').mockImplementation(() => {});

		render(Page);

		await vi.waitFor(() => {
			expect(alertSpy).toHaveBeenCalledWith(
				expect.stringContaining('もう一度Googleでサインインしてください')
			);
		});
	});
});
