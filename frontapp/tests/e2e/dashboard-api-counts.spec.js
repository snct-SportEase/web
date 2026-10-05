import { expect, test } from '@playwright/test';

const mockBackendUrl = process.env.MOCK_BACKEND_URL ?? 'http://127.0.0.1:8081';

for (const path of ['my-page', 'score-list', 'class-info']) {
 test(`${path}で大会APIを一度だけ取得する`, async ({ page, context, request }, testInfo) => {
  await request.post(`${mockBackendUrl}/__reset`);
  await request.post(`${mockBackendUrl}/__set-user`, { data: { user: 'student' } });
  await context.addCookies([{ name: 'session_token', value: 'test-session-token', domain: 'localhost', path: '/' }]);
  const started = performance.now();
  const response = await page.goto(`/dashboard/student/${path}`);
  await page.waitForLoadState('networkidle');
  const counts = await (await request.get(`${mockBackendUrl}/__request-counts`)).json();
  expect(response.status()).toBe(200);
  expect(counts['GET /api/events/active']).toBe(1);
  expect(counts['GET /api/events']).toBeUndefined();
  await testInfo.attach('request-counts', {
   body: JSON.stringify({ path, counts, serverTiming: response.headers()['server-timing'], navigationMs: performance.now() - started }, null, 2),
   contentType: 'application/json'
  });
 });
}
