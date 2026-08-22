import { expect, test } from '@playwright/test';

const mockBackendUrl = process.env.MOCK_BACKEND_URL ?? 'http://127.0.0.1:8081';

const sampleTypingResults = {
  schema_version: 'typing-results-v1',
  export_id: '4ef87d60-2e74-477a-9c16-a93423d04c20',
  teams: [
    { team_name: '1年生', match_1_score: 120, match_2_score: 130, match_3_score: 140, total_score: 390, rank: 1 },
    { team_name: '2年生', match_1_score: 110, match_2_score: 120, match_3_score: 130, total_score: 360, rank: 2 },
    { team_name: '3年生', match_1_score: 100, match_2_score: 110, match_3_score: 120, total_score: 330, rank: 3 },
    { team_name: '4年生', match_1_score: 90, match_2_score: 100, match_3_score: 110, total_score: 300, rank: 4 },
    { team_name: '5年生', match_1_score: 80, match_2_score: 90, match_3_score: 100, total_score: 270, rank: 5 },
    { team_name: '専攻科・教員', match_1_score: 70, match_2_score: 80, match_3_score: 90, total_score: 240, rank: 6 }
  ]
};

test.describe('競技タイピング結果インポート (root)', () => {
  test('サンプルJSONを提出すると順位点が学生画面に反映される', async ({ page, context, request }) => {
    await request.post(`${mockBackendUrl}/__reset`);
    await context.addCookies([{ name: 'session_token', value: 'test-session-token', domain: 'localhost', path: '/' }]);
    page.on('dialog', (dialog) => void dialog.accept());

    await page.goto('/dashboard/root/noon-game');
    const typingCard = page.locator('div.border', { has: page.getByRole('heading', { name: '競技タイピング' }) });
    await typingCard.getByRole('button', { name: '昼競技を作成' }).click();
    await expect(page.getByRole('textbox', { name: 'グループ名' }).first()).toBeVisible();
    await page.getByLabel('公開状態').selectOption('published');
    await expect(page.getByLabel('公開状態')).toHaveValue('published');
    await page.getByRole('button', { name: '昼競技を作成' }).last().click();

    await expect(page.getByRole('heading', { name: '競技タイピング結果のインポート' })).toBeVisible();
    await page.locator('input[type="file"]').setInputFiles({
      name: 'typing-results-v1.json',
      mimeType: 'application/json',
      buffer: Buffer.from(JSON.stringify(sampleTypingResults))
    });

    const importResponse = page.waitForResponse((response) => response.url().includes('/typing-system/import') && response.request().method() === 'POST');
    await page.getByRole('button', { name: '結果をインポート' }).click();
    const submitted = await importResponse;
    expect(submitted.status()).toBe(200);
    await expect(submitted.json()).resolves.toMatchObject({ export_id: sampleTypingResults.export_id });

    await request.post(`${mockBackendUrl}/__set-user`, { data: { user: 'student' } });
    await page.goto('/dashboard/student/noon-game');
    await expect(page.getByRole('heading', { name: '競技タイピング結果' })).toBeVisible();
    const firstYearRow = page.locator('tr', { hasText: '1年生' });
    await expect(firstYearRow).toContainText('120');
    await expect(firstYearRow).toContainText('390');
    await expect(firstYearRow).toContainText('40点');
  });
});
