import { page } from '@vitest/browser/context';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { render } from 'vitest-browser-svelte';
import NotificationSettings from './NotificationSettings.svelte';

class NotificationMock {}

vi.mock('$env/dynamic/public', () => ({
	env: {
		PUBLIC_WEBPUSH_PUBLIC_KEY: 'mock-public-key'
	}
}));

vi.mock('$lib/utils/push.js', () => ({
	userHasPushEligibleRole: (user) =>
		Boolean(user?.roles?.some((role) => ['student', 'admin', 'root'].includes(role.name))),
	ensurePushSubscription: vi.fn(async () => ({ status: 'subscribed' }))
}));

describe('NotificationSettings', () => {
	let fetchMock;
	let originalNotification;
	let originalServiceWorkerDescriptor;
	let originalPushManager;
	let serverSubscriptionPayload;
	let currentDeviceSubscription;

	beforeEach(() => {
		serverSubscriptionPayload = { subscribed: false, count: 0, endpoints: [] };
		currentDeviceSubscription = null;

		fetchMock = vi.fn((url) => {
			if (url === '/api/notifications/subscription') {
				return Promise.resolve({
					ok: true,
					json: () => Promise.resolve(serverSubscriptionPayload)
				});
			}

			return Promise.resolve({
				ok: true,
				json: () => Promise.resolve({})
			});
		});

		vi.stubGlobal('fetch', fetchMock);

		originalNotification = window.Notification;
		originalPushManager = window.PushManager;
		originalServiceWorkerDescriptor = Object.getOwnPropertyDescriptor(navigator, 'serviceWorker');

		NotificationMock.permission = 'default';
		NotificationMock.requestPermission = vi.fn(async () => 'granted');
		vi.stubGlobal('Notification', NotificationMock);
		vi.stubGlobal('PushManager', class PushManagerMock {});

		Object.defineProperty(navigator, 'serviceWorker', {
			configurable: true,
			value: {
				ready: Promise.resolve({
					pushManager: {
						getSubscription: vi.fn(async () => currentDeviceSubscription),
						subscribe: vi.fn(async () => ({
							toJSON: () => ({
								endpoint: 'https://example.com/push',
								keys: {
									auth: 'auth-key',
									p256dh: 'p256dh-key'
								}
							})
						}))
					}
				})
			}
		});
	});

	afterEach(() => {
		vi.restoreAllMocks();

		if (originalNotification) {
			window.Notification = originalNotification;
		} else {
			delete window.Notification;
		}

		if (originalPushManager) {
			window.PushManager = originalPushManager;
		} else {
			delete window.PushManager;
		}

		if (originalServiceWorkerDescriptor) {
			Object.defineProperty(navigator, 'serviceWorker', originalServiceWorkerDescriptor);
		} else {
			delete navigator.serviceWorker;
		}
	});

	it('対応ブラウザでは未対応メッセージを表示しない', async () => {
		render(NotificationSettings, {
			user: {
				roles: [{ name: 'root' }]
			}
		});

		await expect.element(page.getByText('未設定')).toBeInTheDocument();
		await expect.element(page.getByRole('button', { name: '通知を有効にする' })).toBeInTheDocument();
		await expect
			.element(
				page.getByText(
					'このブラウザはプッシュ通知をサポートしていません。モバイルブラウザまたは最新のデスクトップブラウザをご利用ください。'
				)
			)
			.not.toBeInTheDocument();
		expect(fetchMock).not.toHaveBeenCalledWith('/api/notifications/debug', expect.anything());
	});

	it('別デバイスだけ購読済みの場合はこのデバイスの有効化ボタンを表示する', async () => {
		NotificationMock.permission = 'granted';
		serverSubscriptionPayload = {
			subscribed: true,
			count: 1,
			endpoints: ['https://example.com/other-device']
		};

		render(NotificationSettings, {
			user: {
				roles: [{ name: 'student' }]
			}
		});

		await expect
			.element(page.getByText('ほかのデバイスでは通知が有効です。このデバイスでも受け取るには通知を有効にしてください。'))
			.toBeInTheDocument();
		await expect.element(page.getByRole('button', { name: '通知を有効にする' })).toBeInTheDocument();
		await expect.element(page.getByText('無効')).toBeInTheDocument();
	});
});
