import { page } from '@vitest/browser/context';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { render } from 'vitest-browser-svelte';
import Page from './+page.svelte';

function jsonResponse(body, ok = true) {
	return Promise.resolve({
		ok,
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
									startTime: '2026-07-03T10:00:00',
									sides: [
										{ contestantId: 'c0', teamId: 11 },
										{ contestantId: 'c1', teamId: 12 }
									]
								},
								{
									id: 32,
									roundIndex: 0,
									order: 1,
									startTime: '2026-07-03T10:00:00',
									sides: [
										{ contestantId: 'c2', teamId: 13 },
										{ contestantId: 'c3', teamId: 14 }
									]
								}
							],
							contestants: {
								c0: { players: [{ title: '1-1' }] },
								c1: { players: [{ title: '1-2' }] },
								c2: { players: [{ title: '1-3' }] },
								c3: { players: [{ title: '1-4' }] }
							}
						})
					}
				]);
			}

			if (url === '/api/barcode/check-in') {
				return jsonResponse({ error: 'まだあなたのクラスはこの試合にチェックインできません' }, false);
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
		await expect.element(page.getByRole('option', { name: '決勝 10:00開始試合（1-1 vs 1-2, 1-3 vs 1-4）' })).toBeInTheDocument();

		await page.getByLabelText('試合').selectOptions('time:31-32');

		await expect.element(page.getByText('試合: 決勝 10:00開始試合')).toBeInTheDocument();
		const matchupItems = Array.from(document.querySelectorAll('li')).map((item) =>
			item.textContent?.trim()
		);
		expect(matchupItems).toContain('1-1 vs 1-2');
		expect(matchupItems).toContain('1-3 vs 1-4');
		expect(
			fetchMock.mock.calls.some(([url]) =>
				String(url).includes('/api/barcode/matches/31/check-ins?event_id=1&sport_id=7&match_ids=31%2C32')
			)
		).toBe(true);
	});

	it('チェックイン失敗をモーダルで表示できる', async () => {
		render(Page);

		await page.getByLabelText('競技').selectOptions('7');
		await page.getByLabelText('試合').selectOptions('time:31-32');
		await page.getByLabelText('バーコード値').fill('H1023010590');
		await page.getByRole('button', { name: 'チェックインする' }).click();

		await expect
			.element(page.getByRole('dialog', { name: 'チェックインできませんでした' }))
			.toBeInTheDocument();
		await expect
			.element(page.getByText('まだあなたのクラスはこの試合にチェックインできません'))
			.toBeInTheDocument();

		await page.getByRole('button', { name: '閉じる' }).click();

		await expect
			.element(page.getByRole('dialog', { name: 'チェックインできませんでした' }))
			.not.toBeInTheDocument();
	});
});
