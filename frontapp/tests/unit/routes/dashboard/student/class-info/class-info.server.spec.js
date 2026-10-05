import { describe, expect, it, vi } from 'vitest';
import { load } from '$src/routes/dashboard/student/class-info/+page.server.js';

describe('class-info load', () => {
 it('class-progressの大会IDで昼競技を取得し、大会APIを再取得しない', async () => {
  const fetch = vi.fn(async (url) => ({
   ok: true,
   json: async () => String(url).endsWith('/class-progress?view=summary')
    ? { event_id: 7, class_id: 1, class_name: '1A', progress: [] }
    : { session: { name: '昼競技' }, matches: [] }
  }));
  const result = await load({ fetch, locals: { user: { class_id: 1 } }, request: new Request('http://localhost') });
  expect(result.isClassMember).toBe(true);
  expect(fetch).toHaveBeenCalledTimes(2);
  expect(String(fetch.mock.calls[1][0])).toMatch(/\/events\/7\/noon-game\/session$/);
 });
});
