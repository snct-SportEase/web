import { page } from '@vitest/browser/context';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { render } from 'vitest-browser-svelte';
import Page from './+page.svelte';

const mocks = vi.hoisted(() => {
	function ChartMock() {}

	return {
		chartMock: vi.fn(ChartMock),
		webSockets: []
	};
});

vi.mock('chart.js/auto', () => ({
	default: mocks.chartMock
}));

function jsonResponse(body) {
	return Promise.resolve({
		ok: true,
		json: () => Promise.resolve(body)
	});
}

class MockWebSocket {
	constructor(url) {
		this.url = url;
		this.readyState = 1;
		this.onopen = null;
		this.onmessage = null;
		this.onclose = null;
		mocks.webSockets.push(this);
	}

	close() {
		this.readyState = 3;
		if (this.onclose) {
			this.onclose();
		}
	}
}

describe('Manage Dashboard Page', () => {
	let fetchMock;
	let localStorageMock;

	beforeEach(() => {
		mocks.chartMock.mockClear();
		mocks.webSockets.length = 0;

		localStorageMock = {
			getItem: vi.fn(() => 'test-token')
		};

		fetchMock = vi.fn((url) => {
			if (url === '/api/admin/statistics/attendance') {
				return jsonResponse({ attendance_rate: 87.65 });
			}

			if (url === '/api/admin/statistics/participation') {
				return jsonResponse({ バスケットボール: 91, バレーボール: 84 });
			}

			if (url === '/api/admin/statistics/scores') {
				return jsonResponse({
					'2025春': [
						{ class_name: '1A', total_points_current_event: 12 },
						{ class_name: '1B', total_points_current_event: 9 }
					],
					'2025秋': [
						{ class_name: '1A', total_points_current_event: 15 },
						{ class_name: '1B', total_points_current_event: 10 }
					]
				});
			}

			if (url === '/api/admin/statistics/progress') {
				return jsonResponse({
					バスケットボール: '進行中',
					バレーボール: '準備中'
				});
			}

			if (url === '/api/admin/events/active') {
				return jsonResponse({ id: 1, name: '2025春季スポーツ大会', hide_scores: false });
			}

			return jsonResponse({});
		});

		vi.stubGlobal('fetch', fetchMock);
		vi.stubGlobal('localStorage', localStorageMock);
		vi.stubGlobal('WebSocket', MockWebSocket);
	});

	afterEach(() => {
		vi.unstubAllGlobals();
	});

	it('統計情報を表示し、スコア可視時はチャートを描画できる', async () => {
		render(Page);

		await expect.element(page.getByRole('heading', { name: '管理者ダッシュボード' })).toBeInTheDocument();
		await expect.element(page.getByText('87.65%')).toBeInTheDocument();
		await expect.element(page.getByText('バスケットボール')).toBeInTheDocument();
		await expect.element(page.getByText('進行中')).toBeInTheDocument();

		await vi.waitFor(() => {
			expect(mocks.chartMock).toHaveBeenCalledTimes(2);
		});

		expect(localStorageMock.getItem).toHaveBeenCalledWith('token');
		expect(mocks.webSockets).toHaveLength(1);
		expect(mocks.webSockets[0].url).toContain('/api/ws/progress');
	});

	it('スコア非表示時は案内文を表示し、WebSocket 更新を反映できる', async () => {
		fetchMock = vi.fn((url) => {
			if (url === '/api/admin/statistics/attendance') {
				return jsonResponse({ attendance_rate: 91.2 });
			}

			if (url === '/api/admin/statistics/participation') {
				return jsonResponse({ バスケットボール: 93 });
			}

			if (url === '/api/admin/statistics/scores') {
				return jsonResponse({
					'2025春': [{ class_name: '1A', total_points_current_event: 12 }]
				});
			}

			if (url === '/api/admin/statistics/progress') {
				return jsonResponse({
					サッカー: '試合前'
				});
			}

			if (url === '/api/admin/events/active') {
				return jsonResponse({ id: 1, name: '2025春季スポーツ大会', hide_scores: true });
			}

			return jsonResponse({});
		});

		vi.stubGlobal('fetch', fetchMock);

		render(Page);

		await expect.element(page.getByText('スコアは現在非表示に設定されています。')).toBeInTheDocument();
		expect(mocks.chartMock).not.toHaveBeenCalled();

		await vi.waitFor(() => {
			expect(mocks.webSockets).toHaveLength(1);
		});

		mocks.webSockets[0].onmessage({
			data: JSON.stringify({ サッカー: '試合中', バスケットボール: '終了' })
		});

		await expect.element(page.getByText('試合中')).toBeInTheDocument();
		await expect.element(page.getByText('終了')).toBeInTheDocument();
	});
});
