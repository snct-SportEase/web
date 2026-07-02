import { page } from '@vitest/browser/context';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { render } from 'vitest-browser-svelte';
import Page from './+page.svelte';

const barcodeMocks = vi.hoisted(() => ({
	getCameras: vi.fn()
}));

vi.mock('html5-qrcode', () => ({
	Html5Qrcode: Object.assign(
		vi.fn(() => ({
			isScanning: false,
			start: vi.fn(() => Promise.resolve()),
			stop: vi.fn(() => Promise.resolve())
		})),
		{
			getCameras: barcodeMocks.getCameras
		}
	),
	Html5QrcodeSupportedFormats: {
		CODE_39: 'CODE_39',
		CODE_128: 'CODE_128',
		EAN_13: 'EAN_13',
		EAN_8: 'EAN_8',
		ITF: 'ITF',
		UPC_A: 'UPC_A',
		UPC_E: 'UPC_E'
	}
}));

function jsonResponse(body) {
	return Promise.resolve({
		ok: true,
		json: () => Promise.resolve(body)
	});
}

describe('Barcode Reader Page', () => {
	let fetchMock;

	beforeEach(() => {
		barcodeMocks.getCameras.mockResolvedValue([]);

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
							matches: [{ id: 31, roundIndex: 0, order: 0 }]
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
		await expect.element(page.getByRole('option', { name: 'バスケットボール / 決勝 第1試合' })).toBeInTheDocument();
	});
});
