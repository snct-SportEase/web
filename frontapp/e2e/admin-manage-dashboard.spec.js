import { test, expect } from '@playwright/test';

const mockBackendUrl = process.env.MOCK_BACKEND_URL ?? 'http://127.0.0.1:8081';

test.describe('管理者ダッシュボード閲覧 (admin/root)', () => {
	test.beforeEach(async ({ page, context, request }) => {
		await request.post(`${mockBackendUrl}/__reset`);
		await context.addCookies([{ name: 'session_token', value: 'test-session-token', domain: 'localhost', path: '/' }]);
		await page.addInitScript(() => {
			class MockWebSocket {
				constructor(url) {
					this.url = url;
					this.readyState = 1;
					this.onopen = null;
					this.onmessage = null;
					this.onclose = null;
				}

				close() {
					this.readyState = 3;
					if (this.onclose) {
						this.onclose();
					}
				}
			}

			window.WebSocket = MockWebSocket;
		});
		await page.goto('/dashboard/admin/manage-dashboard');
	});

	test('統計情報と進行状況を表示できる', async ({ page }) => {
		await expect(page.getByRole('heading', { name: '管理者ダッシュボード' })).toBeVisible();
		await expect(page.getByText('84.20%')).toBeVisible();
		await expect(page.getByRole('heading', { name: '全体出席率' })).toBeVisible();
		await expect(page.getByRole('heading', { name: '競技ごとの参加率' })).toBeVisible();
		await expect(page.getByRole('heading', { name: 'クラス別スコア推移' })).toBeVisible();
		await expect(page.getByRole('heading', { name: 'リアルタイムイベント進行状況' })).toBeVisible();
		await expect(page.getByText('バスケットボール')).toBeVisible();
		await expect(page.getByText('進行中')).toBeVisible();
		await expect(page.locator('canvas')).toHaveCount(2);
	});
});
