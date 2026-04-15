import { test, expect } from '@playwright/test';

const mockBackendUrl = process.env.MOCK_BACKEND_URL ?? 'http://127.0.0.1:8081';

test.describe('試合結果入力 (admin/root)', () => {
  test.beforeEach(async ({ page, context, request }) => {
    await request.post(`${mockBackendUrl}/__reset`);
    await context.addCookies([{ name: 'session_token', value: 'test-session-token', domain: 'localhost', path: '/' }]);
    await page.goto('/dashboard/admin/insert-matche-result');
    await expect(page.getByRole('heading', { name: '試合結果入力' })).toBeVisible();
  });

  test('スコアを入力して登録できる', async ({ page }) => {
    // トーナメント選択
    await page.getByLabel('トーナメント選択').selectOption('t1');
    
    // 試合カードが表示されるのを待機
    await expect(page.getByText('1A vs 1B')).toBeVisible();

    // 結果を入力ボタン
    await page.getByRole('button', { name: '結果を入力' }).click();

    // モーダル
    await expect(page.getByRole('heading', { name: /結果入力: 1A vs 1B/ })).toBeVisible();

    // スコア入力
    await page.getByLabel('1A Score').fill('15');
    await page.getByLabel('1B Score').fill('12');

    // 確認ボタン
    await page.getByRole('button', { name: '確認' }).click();

    // 確認モーダル
    // デバッグ：スコアが表示されているか
    await expect(page.getByTestId('confirm-team1-score')).toHaveText('1A: 15');
    await expect(page.getByTestId('confirm-team2-score')).toHaveText('1B: 12');
    await expect(page.getByTestId('winner-name')).toHaveText('勝者: 1A');

    // 登録ボタンと通信待機
    const updateRequest = page.waitForRequest(req => 
      req.url().endsWith('/api/admin/matches/m1/result') && req.method() === 'PUT'
    );

    await Promise.all([
      page.waitForEvent('dialog').then((dialog) => dialog.accept()),
      page.getByRole('button', { name: '登録' }).click()
    ]);

    const request = await updateRequest;
    expect(JSON.parse(request.postData() ?? '{}')).toEqual({
      team1_score: 15,
      team2_score: 12
    });
  });

  test('同点の場合に勝者を選択して登録できる', async ({ page }) => {
    await page.getByLabel('トーナメント選択').selectOption('t1');
    await expect(page.getByText('1A vs 1B')).toBeVisible();
    await page.getByRole('button', { name: '結果を入力' }).click();

    await page.getByLabel('1A Score').fill('10');
    await page.getByLabel('1B Score').fill('10');
    await page.getByRole('button', { name: '確認' }).click();

    // 確認モーダル
    await expect(page.getByTestId('confirm-team1-score')).toHaveText('1A: 10');
    await expect(page.getByTestId('confirm-team2-score')).toHaveText('1B: 10');

    // 勝者選択肢
    await expect(page.getByTestId('winner-selection')).toBeVisible();
    await page.getByRole('radio', { name: '1B' }).check();

    const updateRequest = page.waitForRequest(req => 
      req.url().endsWith('/api/admin/matches/m1/result') && req.method() === 'PUT'
    );

    await Promise.all([
      page.waitForEvent('dialog').then((dialog) => dialog.accept()),
      page.getByRole('button', { name: '登録' }).click()
    ]);

    const request = await updateRequest;
    expect(JSON.parse(request.postData() ?? '{}')).toEqual({
      team1_score: 10,
      team2_score: 10,
      winner_id: 2
    });
  });
});
