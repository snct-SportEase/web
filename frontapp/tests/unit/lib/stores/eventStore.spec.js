import { afterEach, describe, expect, it, vi } from 'vitest';

describe('activeEvent', () => {
	afterEach(() => {
		vi.unstubAllGlobals();
		vi.resetModules();
	});

	it('一般ユーザー向けAPIだけで開催中大会を初期化する', async () => {
		const fetchMock = vi.fn(async (url) => {
			if (url === '/api/events/active') {
				return response({ event_id: 7, event_name: '秋季大会', is_rainy_mode: true, test_run_state: 'testing' });
			}
			throw new Error(`unexpected URL: ${url}`);
		});
		vi.stubGlobal('fetch', fetchMock);
		const { activeEvent } = await import('$lib/stores/eventStore.js');

		const event = await activeEvent.init();

		expect(event).toMatchObject({ id: 7, name: '秋季大会', is_rainy_mode: true });
		expect(fetchMock).toHaveBeenCalledExactlyOnceWith('/api/events/active');
 expect(event.test_run_state).toBe('testing');
	});
 it('同時初期化と再初期化は取得を共有し、明示更新時だけ再取得する', async () => {
  const fetchMock = vi.fn(async () => response({ event_id: 7, event_name: '大会' }));
  vi.stubGlobal('fetch', fetchMock);
  const { activeEvent } = await import('$lib/stores/eventStore.js');
  const [first, second] = await Promise.all([activeEvent.init(), activeEvent.init()]);
  expect(first).toEqual(second);
  await activeEvent.init();
  expect(fetchMock).toHaveBeenCalledTimes(1);
  await activeEvent.init({ force: true });
  expect(fetchMock).toHaveBeenCalledTimes(2);
 });

 it('SSRから渡された大会なしの状態でも追加取得しない', async () => {
  const fetchMock = vi.fn();
  vi.stubGlobal('fetch', fetchMock);
  const { activeEvent } = await import('$lib/stores/eventStore.js');
  activeEvent.seed(null);
  expect(await activeEvent.init()).toBeNull();
  expect(fetchMock).not.toHaveBeenCalled();
 });

 it('初期化中の古いレスポンスは新しい大会を上書きしない', async () => {
  let resolveFetch;
  vi.stubGlobal('fetch', vi.fn(() => new Promise(resolve => { resolveFetch = resolve; })));
  const { activeEvent } = await import('$lib/stores/eventStore.js');
  const pending = activeEvent.init();
  activeEvent.seed({ id: 8, name: '新大会' });
  resolveFetch(response({ event_id: 7, event_name: '古い大会' }));
  expect(await pending).toEqual({ id: 8, name: '新大会' });
 });

});

function response(body) {
	return { ok: true, json: async () => body };
}
