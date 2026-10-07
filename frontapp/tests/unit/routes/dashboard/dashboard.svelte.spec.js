import { page } from '@vitest/browser/context';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { render } from 'vitest-browser-svelte';
import Page from '$src/routes/dashboard/+page.svelte';
import { activeEvent } from '$lib/stores/eventStore.js';

vi.mock('$env/dynamic/public', () => ({
	env: {
		PUBLIC_WEBPUSH_PUBLIC_KEY: ''
	}
}));

const rootUser = {
	id: 'root-user-1',
	email: 'root@example.com',
	display_name: 'Root User',
	is_profile_complete: true,
	is_init_root_first_login: false,
	roles: [{ name: 'root' }]
};

const studentUser = {
	...rootUser,
	id: 'student-user-1',
	class_id: 7,
	email: 'student@example.com',
	roles: [{ name: 'student' }]
};

function renderDashboard(user = rootUser, eventSnapshot = null, dataOverrides = {}) {
	// Rendering Page alone omits Layout's SSR snapshot initialization.
	// Supply it for every case, including an explicit absence of an event.
	activeEvent.seed(eventSnapshot);
	return render(Page, {
		props: {
			data: {
				user,
				classes: [],
				events: [{ id: 1, name: '2026春季スポーツ大会' }],
				form: {},
				isClassMember: false,
				className: null,
				classInfo: null,
				members: [],
				progress: [],
				...dataOverrides
			}
		}
	});
}

describe('Dashboard', () => {
	let fetchMock;

	beforeEach(() => {
		window.localStorage.clear();
		fetchMock = vi.fn((url) => {
			if (url === '/api/events/active') {
				return Promise.resolve({
					ok: true,
					json: () => Promise.resolve({})
				});
			}

			if (url === '/api/notifications/subscription') {
				return Promise.resolve({
					ok: true,
					json: () => Promise.resolve({ count: 0, endpoints: [] })
				});
			}

			return Promise.resolve({
				ok: true,
				json: () => Promise.resolve({})
			});
		});
		vi.stubGlobal('fetch', fetchMock);
	});

	afterEach(() => {
		activeEvent.seed(null);
		vi.unstubAllGlobals();
		window.localStorage.clear();
	});

	it('ショートカットを表示設定から非表示にできる', async () => {
		renderDashboard();

		await expect.element(page.getByRole('link', { name: /通知管理/ })).toBeInTheDocument();

		await page.getByRole('button', { name: '表示設定' }).click();
		await page.getByRole('checkbox', { name: /通知管理/ }).click();

		await expect.element(page.getByRole('link', { name: /通知管理/ })).not.toBeInTheDocument();
		expect(
			JSON.parse(window.localStorage.getItem('sportease.dashboard.hiddenShortcuts.root-user-1'))
		).toContain('/dashboard/root/notification');
	});

	it('非表示にしたショートカットをまとめて再表示できる', async () => {
		window.localStorage.setItem(
			'sportease.dashboard.hiddenShortcuts.root-user-1',
			JSON.stringify(['/dashboard/root/notification'])
		);

		renderDashboard();

		await expect.element(page.getByRole('link', { name: /通知管理/ })).not.toBeInTheDocument();

		await page.getByRole('button', { name: /表示設定/ }).click();
		await page.getByRole('button', { name: 'すべて表示' }).click();

		await expect.element(page.getByRole('link', { name: /通知管理/ })).toBeInTheDocument();
		expect(
			JSON.parse(window.localStorage.getItem('sportease.dashboard.hiddenShortcuts.root-user-1'))
		).toEqual([]);
	});

	it('昼競技情報にリレー以外の公開昼競技も表示する', async () => {
		fetchMock.mockImplementation((url) => {
			if (url === '/api/events/active') {
				return Promise.resolve({
					ok: true,
					json: () => Promise.resolve({ event_id: 1, event_name: '2026春季スポーツ大会' })
				});
			}
			if (url === '/api/student/events/1/noon-game/sessions') {
				return Promise.resolve({ ok: true, json: () => Promise.resolve({ sessions: [{ id: 10 }] }) });
			}
			if (url === '/api/student/events/1/noon-game/sessions/10') {
				return Promise.resolve({
					ok: true,
					json: () => Promise.resolve({
						name: '借り物競走',
						matches: [{ id: 101, title: '借り物競走', status: 'scheduled', entries: [{ id: 1, class_ids: [7] }] }]
					})
				});
			}
			if (url === '/api/barcode/teams') {
				return Promise.resolve({
					ok: true,
					json: () => Promise.resolve([{ event_id: 1, location: 'noon_game', sport_name: '借り物競走' }])
				});
			}
			return Promise.resolve({ ok: true, json: () => Promise.resolve({ count: 0, endpoints: [] }) });
		});

		renderDashboard(studentUser, { id: 1, name: '2026春季スポーツ大会' });

		await expect.element(page.getByRole('heading', { name: '昼競技情報' })).toBeInTheDocument();
		await expect.element(page.getByText('借り物競走')).toBeInTheDocument();
		expect(fetchMock).toHaveBeenCalledWith('/api/student/events/1/noon-game/sessions');
		expect(fetchMock).toHaveBeenCalledWith('/api/student/events/1/noon-game/sessions/10');
		expect(fetchMock).toHaveBeenCalledWith('/api/barcode/teams');
		expect(fetchMock).not.toHaveBeenCalledWith('/api/events/active');
	});
	it('割り当てがない昼競技は表示しない', async () => {
		fetchMock.mockImplementation((url) => {
			if (url === '/api/events/active') {
				return Promise.resolve({ ok: true, json: () => Promise.resolve({ event_id: 1 }) });
			}
			if (url === '/api/student/events/1/noon-game/sessions') {
				return Promise.resolve({ ok: true, json: () => Promise.resolve({ sessions: [{ id: 10 }] }) });
			}
			if (url === '/api/student/events/1/noon-game/sessions/10') {
				return Promise.resolve({
					ok: true,
					json: () => Promise.resolve({
						name: '借り物競走',
						matches: [{ id: 101, title: '借り物競走', entries: [{ id: 1, class_ids: [7] }] }]
					})
				});
			}
			if (url === '/api/barcode/teams') {
				return Promise.resolve({ ok: true, json: () => Promise.resolve([]) });
			}
			return Promise.resolve({ ok: true, json: () => Promise.resolve({}) });
		});

		renderDashboard(studentUser, { id: 1, name: '2026春季スポーツ大会' });

		await expect.poll(() => fetchMock.mock.calls.some(([url]) => url === '/api/barcode/teams')).toBe(true);
		await expect.element(page.getByText('現在表示できる昼競技情報はありません。')).toBeInTheDocument();
		await expect.element(page.getByText('借り物競走')).not.toBeInTheDocument();
		expect(fetchMock).toHaveBeenCalledWith('/api/student/events/1/noon-game/sessions/10');
		expect(fetchMock).not.toHaveBeenCalledWith('/api/events/active');
	});

	describe('競技別メンバー一覧', () => {
		const members = [
			{ id: 'soccer', display_name: '山田 太郎', email: 'yamada@example.test', assignments: [{ sport_name: 'サッカー', team_name: '3A_サッカー' }] },
			{ id: 'basketball', display_name: '佐藤 花子', email: 'sato@example.test', assignments: [{ sport_name: 'バスケットボール', team_name: '3A_バスケットボール' }] },
			{ id: 'both', display_name: '鈴木 一郎', email: 'suzuki@example.test', assignments: [{ sport_name: 'サッカー', team_name: '3A_サッカー' }, { sport_name: 'バスケットボール', team_name: '3A_バスケットボール' }] },
			{ id: 'unassigned', display_name: '田中 美咲', email: 'tanaka@example.test' }
		];

		function renderRoster(roster = members) {
			return renderDashboard(studentUser, null, { isClassMember: true, members: roster });
		}

		it('初期状態では未割り当てのメンバーも含めて全員と人数を表示する', async () => {
			renderRoster();

			for (const member of members) {
				await expect.element(page.getByRole('cell', { name: member.display_name, exact: true })).toBeInTheDocument();
			}
			await expect.element(page.getByText('4 名', { exact: true })).toBeInTheDocument();
			await expect.element(page.getByRole('cell', { name: '未割り当て', exact: true })).toBeInTheDocument();
			expect(page.getByRole('combobox', { name: '競技' }).element().value).toBe('');
			expect(page.getByRole('combobox', { name: '競技' }).element().options.length).toBe(3);
		});

		it('競技を切り替えると該当者だけを表示し、複数競技の登録者も含める', async () => {
			renderRoster();
			const select = page.getByRole('combobox', { name: '競技' });

			await select.selectOptions('サッカー');
			await expect.element(page.getByRole('cell', { name: '山田 太郎', exact: true })).toBeInTheDocument();
			await expect.element(page.getByRole('cell', { name: '鈴木 一郎', exact: true })).toBeInTheDocument();
			await expect.element(page.getByRole('cell', { name: '佐藤 花子', exact: true })).not.toBeInTheDocument();
			await expect.element(page.getByRole('cell', { name: '田中 美咲', exact: true })).not.toBeInTheDocument();
			await expect.element(page.getByText('2 名', { exact: true })).toBeInTheDocument();

			await select.selectOptions('バスケットボール');
			await expect.element(page.getByRole('cell', { name: '佐藤 花子', exact: true })).toBeInTheDocument();
			await expect.element(page.getByRole('cell', { name: '鈴木 一郎', exact: true })).toBeInTheDocument();
			await expect.element(page.getByRole('cell', { name: '山田 太郎', exact: true })).not.toBeInTheDocument();
			await expect.element(page.getByRole('cell', { name: '田中 美咲', exact: true })).not.toBeInTheDocument();
			await expect.element(page.getByText('2 名', { exact: true })).toBeInTheDocument();
		});

		it('すべての競技に戻すと全員と人数を再表示する', async () => {
			renderRoster();
			const select = page.getByRole('combobox', { name: '競技' });
			await select.selectOptions('サッカー');
			await expect.element(page.getByRole('cell', { name: '佐藤 花子', exact: true })).not.toBeInTheDocument();

			await select.selectOptions('');
			for (const member of members) {
				await expect.element(page.getByRole('cell', { name: member.display_name, exact: true })).toBeInTheDocument();
			}
			await expect.element(page.getByText('4 名', { exact: true })).toBeInTheDocument();
		});

		it('メンバーがいない場合は0人と空状態の案内を表示する', async () => {
			renderRoster([]);

			await expect.element(page.getByText('0 名', { exact: true })).toBeInTheDocument();
			await expect.element(page.getByText('クラスメンバーがまだ登録されていません。')).toBeInTheDocument();
			await expect.element(page.getByRole('table')).not.toBeInTheDocument();
			expect(page.getByRole('combobox', { name: '競技' }).element().options.length).toBe(1);
		});
	});

});
