import { afterEach, describe, expect, it, vi } from 'vitest';

describe('activeEvent', () => {
	afterEach(() => {
		vi.unstubAllGlobals();
		vi.resetModules();
	});

	it('一般ユーザー向けAPIだけで開催中大会を初期化する', async () => {
		const fetchMock = vi.fn(async (url) => {
			if (url === '/api/events/active') {
				return response({ event_id: 7 });
			}
			if (url === '/api/events') {
				return response([{ id: 7, name: '秋季大会', is_rainy_mode: true }]);
			}
			throw new Error(`unexpected URL: ${url}`);
		});
		vi.stubGlobal('fetch', fetchMock);
		const { activeEvent } = await import('$lib/stores/eventStore.js');

		const event = await activeEvent.init();

		expect(event).toMatchObject({ id: 7, name: '秋季大会', is_rainy_mode: true });
		expect(fetchMock).not.toHaveBeenCalledWith(expect.stringContaining('/api/root/'), expect.anything());
	});
});

function response(body) {
	return { ok: true, json: async () => body };
}
