import { expect, test } from '@playwright/test';

const mockBackendUrl =
  process.env.MOCK_BACKEND_URL ??
  `http://127.0.0.1:${process.env.MOCK_BACKEND_PORT ?? 8081}`;

function tournamentData(prefix, firstRoundMatchCount) {
  const roundCount = Math.log2(firstRoundMatchCount) + 1;
  const contestants = {};
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
    await expect(page.locator('#bracket-1 .bracket-root')).toBeVisible();
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
    await page.route('**/api/student/events/1/board-game-runs', async (route) => {
      await route.fulfill({
        contentType: 'application/json',
        body: JSON.stringify([{
          game_type: 'shogi',
          location: 'ICTメディア室',
          regular_minutes: 15,
          final_minutes: 30,
          win_points: 5,
          tournaments: [{
            id: 27,
            entries: [{ id: 1, team_name: '1-1 A', members: [{ display_name: '参加者A', is_substitute: false }] }],
            rankings: []
          }]
        }])
      });
    });

    await page.reload();

    await expect(page.locator('#bracket-36 .bracket-root')).toBeVisible();
    await expect(page.locator('#bracket-36')).toContainText('通常 1');
    await expect(page.getByText('参加者A')).toBeVisible();
    await expect(page.locator('#bracket-27 .bracket-root')).toBeVisible();
    await expect(page.locator('#bracket-27')).toContainText('将棋 1');
  });
});
