import { page } from '@vitest/browser/context';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { render } from 'vitest-browser-svelte';
import Page from '$src/routes/dashboard/+page.svelte';

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
	email: 'student@example.com',
	roles: [{ name: 'student' }]
};

function renderDashboard(user = rootUser) {
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
				progress: []
			}
		}
	});
}

describe('Dashboard shortcuts', () => {
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
						matches: [{ id: 101, title: '借り物競走', status: 'scheduled', entries: [] }]
					})
				});
			}
			return Promise.resolve({ ok: true, json: () => Promise.resolve({ count: 0, endpoints: [] }) });
		});

		renderDashboard(studentUser);

		await expect.element(page.getByRole('heading', { name: '昼競技情報' })).toBeInTheDocument();
		await expect.element(page.getByText('借り物競走')).toBeInTheDocument();
	});
});
