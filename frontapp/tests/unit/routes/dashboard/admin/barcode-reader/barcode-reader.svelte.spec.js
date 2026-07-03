import { page } from '@vitest/browser/context';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { render } from 'vitest-browser-svelte';
import Page from '$src/routes/dashboard/admin/barcode-reader/+page.svelte';

function minutesFromNow(minutes) {
	// Match start times are registered as JST wall-clock values.
	const jstOffsetMs = 9 * 60 * 60 * 1000;
	const date = new Date(Date.now() + minutes * 60 * 1000 + jstOffsetMs);
	return `${date.getUTCFullYear()}-${String(date.getUTCMonth() + 1).padStart(2, '0')}-${String(
		date.getUTCDate()
	).padStart(2, '0')}T${String(date.getUTCHours()).padStart(2, '0')}:${String(
		date.getUTCMinutes()
	).padStart(2, '0')}:00`;
}

function jsonResponse(body, ok = true) {
	return Promise.resolve({
		ok,
		json: () => Promise.resolve(body)
	});
}

describe('Barcode Reader Page', () => {
	let fetchMock;
	let checkInResponse;
	let matchCheckInsResponse;
	let matchStartTime;

	beforeEach(() => {
		matchStartTime = minutesFromNow(9);
		checkInResponse = {
			ok: false,
			body: { error: 'まだあなたのクラスはこの試合にチェックインできません' }
		};
		matchCheckInsResponse = {
			members: [
				{
					user_id: 'user-1',
					email: 's2301059@sendai-nct.jp',
					display_name: '山田 太郎',
					class_name: '1-1',
					team_name: '1-1',
					team_id: 11,
					checked_in_at: '2026-07-03T09:55:00Z'
				}
			],
			count: 1,
			checked_in_members: [
				{
					user_id: 'user-1',
					email: 's2301059@sendai-nct.jp',
					display_name: '山田 太郎',
					class_name: '1-1',
					team_name: '1-1',
					team_id: 11,
					checked_in_at: '2026-07-03T09:55:00Z'
				}
			],
			checked_in_count: 1,
			unchecked_members: [
				{
					user_id: 'user-2',
					email: 's2301060@sendai-nct.jp',
					display_name: '佐藤 花子',
					class_name: '1-2',
					team_name: '1-2',
					team_id: 12
				}
			],
			unchecked_count: 1
		};
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
									startTime: matchStartTime,
									sides: [
										{ contestantId: 'c0', teamId: 11 },
										{ contestantId: 'c1', teamId: 12 }
									]
								},
								{
									id: 32,
									roundIndex: 0,
									order: 1,
									startTime: matchStartTime,
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
				return jsonResponse(checkInResponse.body, checkInResponse.ok);
			}

			if (String(url).startsWith('/api/barcode/matches/31/check-ins?')) {
				return jsonResponse(matchCheckInsResponse);
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
		const sportSelect = document.querySelector('#sport-select');
		expect(sportSelect?.value).toBe('7');
		expect(sportSelect?.selectedOptions?.[0]?.textContent?.trim()).toBe('バスケットボール');
		const matchOption = document.querySelector('#match-select option[value="time:31-32"]');
		expect(matchOption?.textContent).toContain('決勝');
		expect(matchOption?.textContent).toContain('開始試合（1-1 vs 1-2, 1-3 vs 1-4）');
		expect(matchOption?.disabled).toBe(false);

		await page.getByLabelText('試合').selectOptions('time:31-32');

		expect(document.body.textContent).toContain('試合: 決勝');
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

	it('開始10分より前の試合は選択できない', async () => {
		matchStartTime = minutesFromNow(12);

		render(Page);

		await page.getByLabelText('競技').selectOptions('7');

		const matchOption = document.querySelector('#match-select option[value="time:31-32"]');
		expect(matchOption?.disabled).toBe(true);
		expect(matchOption?.textContent).toContain('から選択可');
		await expect.element(page.getByRole('button', { name: 'チェックインする' })).toBeDisabled();
	});

	it('タイムゾーン付きの開始時刻も登録済みJST時刻のまま表示する', async () => {
		matchStartTime = '2026-07-03T10:00:00Z';

		render(Page);

		await page.getByLabelText('競技').selectOptions('7');

		const matchOption = document.querySelector('#match-select option[value="time:31-32"]');
		expect(matchOption?.textContent).toContain('決勝 10:00開始試合');
		expect(matchOption?.textContent).not.toContain('19:00開始試合');
	});

	it('チェックイン済みと未チェックインの学生をモーダルで確認できる', async () => {
		render(Page);

		await page.getByLabelText('競技').selectOptions('7');
		await page.getByLabelText('試合').selectOptions('time:31-32');

		await page.getByRole('button', { name: 'チェックイン済み（1人）' }).click();
		const checkedDialog = page.getByRole('dialog', { name: 'チェックイン済みの学生' });
		await expect.element(checkedDialog).toBeInTheDocument();
		await expect.element(checkedDialog.getByText('1-1')).toBeInTheDocument();
		await expect.element(checkedDialog.getByText('チェックイン済み: 1 / 1人')).toBeInTheDocument();
		await expect.element(checkedDialog.getByText('チェックイン済み: 0 / 1人')).toBeInTheDocument();
		await expect.element(checkedDialog.getByText('山田 太郎')).toBeInTheDocument();
		await page.getByRole('button', { name: '閉じる' }).click();

		await page.getByRole('button', { name: '未チェックイン（1人）' }).click();
		const uncheckedDialog = page.getByRole('dialog', { name: '未チェックインの学生' });
		await expect.element(uncheckedDialog).toBeInTheDocument();
		await expect.element(uncheckedDialog.getByText('1-2')).toBeInTheDocument();
		await expect.element(uncheckedDialog.getByText('未チェックイン: 0 / 1人')).toBeInTheDocument();
		await expect.element(uncheckedDialog.getByText('未チェックイン: 1 / 1人')).toBeInTheDocument();
		await expect.element(uncheckedDialog.getByText('佐藤 花子')).toBeInTheDocument();
		await page.getByRole('button', { name: '閉じる' }).click();

		await expect
			.element(page.getByRole('dialog', { name: '未チェックインの学生' }))
			.not.toBeInTheDocument();
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

	it('チェックイン済みをモーダルで表示できる', async () => {
		checkInResponse = {
			ok: false,
			body: {
				error: 'チェックイン済みです',
				already_checked_in: true
			}
		};

		render(Page);

		await page.getByLabelText('競技').selectOptions('7');
		await page.getByLabelText('試合').selectOptions('time:31-32');
		await page.getByLabelText('バーコード値').fill('H1023010590');
		await page.getByRole('button', { name: 'チェックインする' }).click();

		await expect
			.element(page.getByRole('dialog', { name: 'チェックイン済みです' }))
			.toBeInTheDocument();

		await page.getByRole('button', { name: '閉じる' }).click();

		await expect
			.element(page.getByRole('dialog', { name: 'チェックイン済みです' }))
			.not.toBeInTheDocument();
	});

	it('チェックイン成功をモーダルで表示できる', async () => {
		checkInResponse = {
			ok: true,
			body: {
				display_name: '山田 太郎',
				student_number: '2301059',
				sport_name: 'バスケットボール',
				round: 1
			}
		};

		render(Page);

		await page.getByLabelText('競技').selectOptions('7');
		await page.getByLabelText('試合').selectOptions('time:31-32');
		await page.getByLabelText('バーコード値').fill('H1023010590');
		await page.getByRole('button', { name: 'チェックインする' }).click();

		await expect
			.element(page.getByRole('dialog', { name: 'ラウンドチェックインを完了しました' }))
			.toBeInTheDocument();
		await expect.element(page.getByText('氏名: 山田 太郎')).toBeInTheDocument();
		await expect.element(page.getByText('学籍番号: 2301059')).toBeInTheDocument();

		await page.getByRole('button', { name: '閉じる' }).click();

		await expect
			.element(page.getByRole('dialog', { name: 'ラウンドチェックインを完了しました' }))
			.not.toBeInTheDocument();
	});
});
