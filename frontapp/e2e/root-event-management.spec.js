import { test, expect } from '@playwright/test';

const mockBackendUrl = process.env.MOCK_BACKEND_URL ?? 'http://127.0.0.1:8081';

test.describe.configure({ mode: 'serial' });

test.describe('大会情報登録・管理 (root)', () => {
  test.beforeEach(async ({ page, context, request }) => {
    await request.post(`${mockBackendUrl}/__reset`);

    await context.addCookies([{
      name: 'session_token',
      value: 'test-session-token',
      domain: 'localhost',
      path: '/'
    }]);

    await page.goto('/dashboard/root/event-management');
    await expect(page.getByText('2025春季スポーツ大会')).toBeVisible();
  });

  test('大会一覧を表示できる', async ({ page }) => {
    await expect(page.getByRole('heading', { name: '大会情報登録・管理' })).toBeVisible();
    await expect(page.getByText('予定')).toBeVisible();
  });

  test('新しい大会を作成できる', async ({ page }) => {
    await page.getByRole('button', { name: '新規作成' }).click();

    await expect(page.getByText('大会作成')).toBeVisible();

    await page.getByRole('spinbutton', { name: '年度' }).fill('2026');
    await page.getByRole('combobox', { name: 'シーズン' }).selectOption('autumn');
    await expect(page.getByRole('textbox', { name: '大会名' })).toHaveValue('2026秋季スポーツ大会');

    await page.getByLabel('開始日').fill('2026-10-01');
    await page.getByLabel('終了日').fill('2026-10-02');

    const saveRequest = page.waitForRequest((request) => {
      if (request.url().endsWith('/api/root/events') && request.method() === 'POST') {
        const body = JSON.parse(request.postData() ?? '{}');
        return body.name === '2026秋季スポーツ大会' && body.year === 2026;
      }

      return false;
    });

    await page.getByRole('button', { name: '保存' }).click();

    const request = await saveRequest;
    const body = JSON.parse(request.postData() ?? '{}');
    expect(body).toMatchObject({
      name: '2026秋季スポーツ大会',
      year: 2026,
      season: 'autumn',
      start_date: '2026-10-01',
      end_date: '2026-10-02',
      status: 'upcoming',
      hide_scores: false
    });

    await expect(page.getByText('大会作成')).not.toBeVisible();
  });

  test('新規作成でスコア非表示設定を保存できる', async ({ page }) => {
    await page.getByRole('button', { name: '新規作成' }).click();

    const hideScoresCheckbox = page.getByLabel('スコアを非表示にする');
    await expect(hideScoresCheckbox).not.toBeChecked();

    await page.getByRole('spinbutton', { name: '年度' }).fill('2026');
    await page.getByRole('combobox', { name: 'シーズン' }).selectOption('autumn');
    await page.getByLabel('開始日').fill('2026-10-01');
    await page.getByLabel('終了日').fill('2026-10-02');
    await hideScoresCheckbox.check({ force: true });

    const saveRequest = page.waitForRequest((request) => {
      if (request.url().endsWith('/api/root/events') && request.method() === 'POST') {
        const body = JSON.parse(request.postData() ?? '{}');
        return body.hide_scores === true;
      }

      return false;
    });

    await page.getByRole('button', { name: '保存' }).click();

    const request = await saveRequest;
    const body = JSON.parse(request.postData() ?? '{}');
    expect(body).toMatchObject({
      hide_scores: true
    });
  });

  test('大会名を手動で変更した後は自動生成が停止する', async ({ page }) => {
    await page.getByRole('button', { name: '新規作成' }).click();

    const nameInput = page.getByRole('textbox', { name: '大会名' });
    await nameInput.fill('カスタム大会名');
    await page.getByRole('spinbutton', { name: '年度' }).fill('2027');
    await page.getByRole('combobox', { name: 'シーズン' }).selectOption('autumn');

    await expect(nameInput).toHaveValue('カスタム大会名');
  });

  test('既存の大会を編集できる', async ({ page }) => {
    await page.getByText('2025春季スポーツ大会').click();

    await expect(page.getByText('大会編集')).toBeVisible();
    await expect(page.getByRole('textbox', { name: '大会名' })).toHaveValue('2025春季スポーツ大会');
    await expect(page.getByLabel('開始日')).toHaveValue('2025-04-01');
    await expect(page.getByLabel('終了日')).toHaveValue('2025-04-02');

    await page.getByRole('combobox', { name: 'ステータス' }).selectOption('active');
    await page.getByLabel('アンケートURL').fill('https://example.com/updated-survey');

    const saveRequest = page.waitForRequest((request) => {
      if (request.url().endsWith('/api/root/events/1') && request.method() === 'PUT') {
        const body = JSON.parse(request.postData() ?? '{}');
        return body.status === 'active' && body.survey_url === 'https://example.com/updated-survey';
      }

      return false;
    });

    await page.getByRole('button', { name: '保存' }).click();

    const request = await saveRequest;
    const body = JSON.parse(request.postData() ?? '{}');
    expect(body).toMatchObject({
      id: 1,
      name: '2025春季スポーツ大会',
      year: 2025,
      season: 'spring',
      status: 'active',
      survey_url: 'https://example.com/updated-survey'
    });

    await expect(page.getByText('大会編集')).not.toBeVisible();
  });

  test('既存大会のスコア非表示設定を更新できる', async ({ page }) => {
    await page.getByText('2025春季スポーツ大会').click();

    const hideScoresCheckbox = page.getByLabel('スコアを非表示にする');
    await expect(hideScoresCheckbox).not.toBeChecked();
    await hideScoresCheckbox.check({ force: true });

    const saveRequest = page.waitForRequest((request) => {
      if (request.url().endsWith('/api/root/events/1') && request.method() === 'PUT') {
        const body = JSON.parse(request.postData() ?? '{}');
        return body.hide_scores === true;
      }

      return false;
    });

    await page.getByRole('button', { name: '保存' }).click();

    const request = await saveRequest;
    const body = JSON.parse(request.postData() ?? '{}');
    expect(body).toMatchObject({
      id: 1,
      hide_scores: true
    });
  });

  test('既存大会のアンケート通知を送信できる', async ({ page }) => {
    page.on('dialog', (dialog) => {
      if (dialog.type() === 'confirm') {
        void dialog.accept().catch(() => {});
        return;
      }

      void dialog.dismiss().catch(() => {});
    });

    await page.getByText('2025春季スポーツ大会').click();
    await expect(page.getByRole('button', { name: '通知を送信' })).toBeVisible();

    const notifyRequest = page.waitForRequest((request) => {
      return request.url().endsWith('/api/root/events/1/notify-survey') && request.method() === 'POST';
    });

    await page.getByRole('button', { name: '通知を送信' }).click();

    await notifyRequest;
  });

  test('春季大会の得点CSVをインポートできる', async ({ page }) => {
    page.once('dialog', (dialog) => {
      void dialog.accept().catch(() => {});
    });

    const uploadRequest = page.waitForRequest((request) => {
      return request.url().endsWith('/api/root/events/1/import-survey-scores') && request.method() === 'POST';
    });

    await page.locator('input[type="file"]').setInputFiles({
      name: 'scores.csv',
      mimeType: 'text/csv',
      buffer: Buffer.from('class,score\n1A,100\n1B,90\n')
    });

    await uploadRequest;
    await expect(page.locator('input[type="file"]')).toHaveValue('');
  });

  test('クラス別スコア集計をCSV出力できる', async ({ page }) => {
    const exportRequest = page.waitForRequest((request) => {
      return request.url().endsWith('/api/root/events/1/export/csv') && request.method() === 'GET';
    });

    await page.getByRole('button', { name: 'CSV出力' }).click();

    await exportRequest;
  });

  test('クラス別スコア集計をPDF出力できる', async ({ page }) => {
    const scoreRequest = page.waitForRequest((request) => {
      return request.url().includes('/api/scores/class?event_id=1') && request.method() === 'GET';
    });

    await page.getByRole('button', { name: 'PDF出力' }).click();

    await scoreRequest;
  });

  test('DBダンプを出力できる', async ({ page }) => {
    const dumpRequest = page.waitForRequest((request) => {
      return request.url().endsWith('/api/root/db/export') && request.method() === 'GET';
    });

    await page.getByRole('button', { name: 'DBダンプ出力' }).click();

    await dumpRequest;
  });
});
