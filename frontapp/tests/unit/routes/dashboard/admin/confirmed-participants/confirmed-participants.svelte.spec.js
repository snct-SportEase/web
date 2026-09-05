import { page } from '@vitest/browser/context';
import { afterEach, describe, expect, it, vi } from 'vitest';
import { render } from 'vitest-browser-svelte';
import Page from '$src/routes/dashboard/admin/confirmed-participants/+page.svelte';

describe('Confirmed participants page', () => {
	afterEach(() => vi.unstubAllGlobals());

	it('開催中大会に割り当てられた競技だけを表示する', async () => {
		const fetchMock = vi.fn(async (url) => {
			if (url === '/api/events/active') return response({ event_id: 7 });
			if (url === '/api/events/7/sports') {
				return response([{ sport_id: 2, sport_name: 'バレーボール' }]);
			}
			throw new Error(`unexpected URL: ${url}`);
		});
		vi.stubGlobal('fetch', fetchMock);

		render(Page, { props: { data: { classes: [{ id: 1, name: '1-A' }] } } });
		await page.getByRole('combobox').first().selectOptions('1');

		await expect.element(page.getByRole('option', { name: 'バレーボール' })).toBeInTheDocument();
		expect(fetchMock).not.toHaveBeenCalledWith('/api/admin/allsports', expect.anything());
	});
});

function response(body) {
	return { ok: true, json: async () => body };
}
