import { expect, test } from '@playwright/test';

const mockBackendUrl =
  process.env.MOCK_BACKEND_URL ??
  `http://127.0.0.1:${process.env.MOCK_BACKEND_PORT ?? 8081}`;

function tournamentData(prefix: string, firstRoundMatchCount: number) {
  const roundCount = Math.log2(firstRoundMatchCount) + 1;
  const contestants: Record<string, { players: { title: string }[] }> = {};
  const matches = [];

  for (let index = 0; index < firstRoundMatchCount * 2; index += 1) {
    contestants[`c${index}`] = { players: [{ title: `${prefix} ${index + 1}` }] };
  }
  for (let index = 0; index < firstRoundMatchCount; index += 1) {
    matches.push({
      roundIndex: 0,
      order: index,
      sides: [{ contestantId: `c${index * 2}` }, { contestantId: `c${index * 2 + 1}` }]
    });
  }
  for (let roundIndex = 1; roundIndex < roundCount; roundIndex += 1) {
    const matchCount = firstRoundMatchCount / (2 ** roundIndex);
    for (let order = 0; order < matchCount; order += 1) {
      matches.push({ roundIndex, order });
    }
  }

  return {
    rounds: Array.from({ length: roundCount }, (_, index) => ({ name: `Round ${index + 1}` })),
    matches,
    contestants
  };
}

test.describe('トーナメント閲覧 (student)', () => {
  test.beforeEach(async ({ page, context, request }) => {
    await request.post(`${mockBackendUrl}/__reset`);
    await request.post(`${mockBackendUrl}/__set-user`, { data: { user: 'student' } });
    await request.post(`${mockBackendUrl}/__set-active-event`, { data: { event_id: 1 } });
    await context.addCookies([{
      name: 'session_token',
      value: 'test-session-token',
      domain: 'localhost',
      path: '/'
    }]);

    await page.goto('/dashboard/student/tournament');
  });

  test('開催中競技のトーナメントを一覧確認する', async ({ page }) => {
    await expect(page.getByRole('heading', { name: 'トーナメント一覧' })).toBeVisible();
    await expect(page.getByRole('heading', { name: 'バスケットボール', exact: true })).toBeVisible();
    await expect(page.getByRole('heading', { name: 'バスケットボール 敗者復活' })).toBeVisible();
    await expect(page.locator('#bracket-1 .ml-root')).toBeVisible();
  });

  test('参加者にホバーすると勝ち上がり経路を強調する', async ({ page }) => {
    const participant = page.locator('#bracket-1 .ml-team[data-contestant-key]').first();
    await expect(participant).toBeVisible();
    await participant.hover();
    await expect(participant).toHaveClass(/ml-team--highlighted/);
  });

  test.describe('タッチ端末', () => {
    test.use({ hasTouch: true, isMobile: true, viewport: { width: 390, height: 844 } });

    test('参加者をタップすると勝ち上がり経路を強調する', async ({ page }) => {
      const participant = page.locator('#bracket-1 .ml-team[data-contestant-key]').first();
      await expect(participant).toBeVisible();
      await participant.tap();
      await expect(participant).toHaveClass(/ml-team--highlighted/);
    });
  });

  test('通常競技と盤上競技のトーナメント表をどちらも表示する', async ({ page }) => {
    await page.route('**/api/student/events/1/tournaments', async (route) => {
      await route.fulfill({
        contentType: 'application/json',
        body: JSON.stringify([
          { id: 36, name: 'デモバスケットボール Tournament', sport_id: 1, data: tournamentData('通常', 4) },
          { id: 27, name: '将棋 Aブロック', sport_id: 10, data: tournamentData('将棋', 8) }
        ])
      });
    });
    await page.reload();

    await expect(page.locator('#bracket-36 .ml-root')).toBeVisible();
    await expect(page.locator('#bracket-36')).toContainText('通常 1');
    await expect(page.getByText('会場:')).toHaveCount(0);
    await expect(page.getByText('出場者')).toHaveCount(0);
    await expect(page.locator('#bracket-27 .ml-root')).toBeVisible();
    await expect(page.locator('#bracket-27')).toContainText('将棋 1');
  });
});
