import { describe, expect, it, vi } from 'vitest';
import { openNotificationTarget } from '$lib/utils/notificationNavigation.js';

describe('openNotificationTarget', () => {
	it('既存画面を通知一覧へ移動してからフォーカスする', async () => {
		const client = {
			navigate: vi.fn(async () => client),
			focus: vi.fn(async () => client)
		};
		const clientsApi = {
			matchAll: vi.fn(async () => [client]),
			openWindow: vi.fn()
		};

		await openNotificationTarget(clientsApi, 'http://localhost:3300/dashboard/student/notification');

		expect(client.navigate).toHaveBeenCalledWith('http://localhost:3300/dashboard/student/notification');
		expect(client.focus).toHaveBeenCalledOnce();
		expect(clientsApi.openWindow).not.toHaveBeenCalled();
	});

	it('既存画面がなければ通知一覧を新しく開く', async () => {
		const clientsApi = {
			matchAll: vi.fn(async () => []),
			openWindow: vi.fn(async () => ({}))
		};

		await openNotificationTarget(clientsApi, 'https://example.test/dashboard/student/notification');

		expect(clientsApi.openWindow).toHaveBeenCalledWith('https://example.test/dashboard/student/notification');
	});
});
