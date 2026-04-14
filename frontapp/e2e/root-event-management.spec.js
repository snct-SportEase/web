import { test, expect } from '@playwright/test';

test.describe.configure({ mode: 'serial' });

test.describe('大会情報登録・管理 (root)', () => {
  test.beforeEach(async ({ page, context, request }) => {
    await request.post('http://127.0.0.1:8081/__reset');

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

  test('既存大会のアンケート通知を送信できる', async ({ page }) => {
    page.on('dialog', async (dialog) => {
      if (dialog.type() === 'confirm') {
        await dialog.accept();
        return;
      }

      await dialog.dismiss();
    });

    await page.getByText('2025春季スポーツ大会').click();
    await expect(page.getByRole('button', { name: '通知を送信' })).toBeVisible();

    const notifyRequest = page.waitForRequest((request) => {
      return request.url().endsWith('/api/root/events/1/notify-survey') && request.method() === 'POST';
    });

    await page.getByRole('button', { name: '通知を送信' }).click();

    await notifyRequest;
  });
});
