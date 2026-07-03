import { page } from '@vitest/browser/context';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { render } from 'vitest-browser-svelte';
import Page from './+page.svelte';

function jsonResponse(body) {
	return Promise.resolve({
		ok: true,
		json: () => Promise.resolve(body)
	});
}

describe('Barcode Reader Page', () => {
	let fetchMock;

	beforeEach(() => {
		fetchMock = vi.fn((url) => {
			if (url === '/api/events/active') {
				return jsonResponse({
					event_id: 1,
					event_name: '2025春季スポーツ大会'
				});
			}

			if (url === '/api/events/1/sports') {
				return jsonResponse([
					{
						event_id: 1,
						sport_id: 7,
						sport_name: 'バスケットボール',
						location: 'gym1'
					}
				]);
			}

			if (url === '/api/admin/events/1/tournaments') {
				return jsonResponse([
					{
						id: 3,
						sport_id: 7,
						name: 'バスケットボール',
						data: JSON.stringify({
							rounds: [{ name: '決勝' }],
							matches: [
								{
									id: 31,
									roundIndex: 0,
									order: 0,
									sides: [
										{ contestantId: 'c0', teamId: 11 },
										{ contestantId: 'c1', teamId: 12 }
									]
								}
							],
							contestants: {
								c0: { players: [{ title: '1-1' }] },
								c1: { players: [{ title: '1-2' }] }
							}
						})
					}
				]);
			}

			return jsonResponse({});
		});

		vi.stubGlobal('fetch', fetchMock);
	});

	afterEach(() => {
		vi.unstubAllGlobals();
		vi.clearAllMocks();
	});

	it('event_sports形式の競技一覧をバーコード読み取り画面の選択肢に表示できる', async () => {
		render(Page);

		await expect.element(page.getByText('2025春季スポーツ大会')).toBeInTheDocument();
		await expect.element(page.getByRole('option', { name: 'バスケットボール' })).toBeInTheDocument();

		const optionTexts = Array.from(document.querySelectorAll('#sport-select option')).map((option) =>
			option.textContent?.trim()
		);
		expect(optionTexts).toEqual(['競技を選択してください', 'バスケットボール']);

		await page.getByLabelText('競技').selectOptions('7');

		await expect.element(page.getByText('選択中: バスケットボール')).toBeInTheDocument();
		await expect.element(page.getByRole('option', { name: 'バスケットボール / 決勝 第1試合（1-1 vs 1-2）' })).toBeInTheDocument();

		await page.getByLabelText('試合').selectOptions('3:31:0');

		await expect.element(page.getByText('対戦: 1-1 vs 1-2')).toBeInTheDocument();
	});
});
