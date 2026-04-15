import { test, expect } from '@playwright/test';

const mockBackendUrl = process.env.MOCK_BACKEND_URL ?? 'http://127.0.0.1:8081';

test.describe('競技詳細登録 (admin/root)', () => {
  test.beforeEach(async ({ page, context, request }) => {
    await request.post(`${mockBackendUrl}/__reset`);
    // 競技を1つ大会に割り当てておく
    await request.post(`${mockBackendUrl}/api/admin/events/1/sports`, {
      data: {
        sport_id: 1,
        location: 'gym1',
        description: '初期概要',
        rules: '初期ルール'
      }
    });
    await context.addCookies([{ name: 'session_token', value: 'test-session-token', domain: 'localhost', path: '/' }]);
    await page.goto('/dashboard/admin/sport-details-registration');
    await expect(page.getByRole('heading', { name: '競技詳細情報登録' })).toBeVisible();
  });

  test('競技の概要とルールを更新できる', async ({ page }) => {
    await page.getByLabel('競技選択').selectOption('1');
    
    // 概要の入力
    // getByRole('textbox') を使うが、複数あるため見出しで特定
    const descriptionHeading = page.getByRole('heading', { name: '競技概要' });
    const descriptionTextarea = page.locator('h2:text("競技概要") + textarea');
    await descriptionTextarea.fill('新しいバスケットボールの概要');

    // ルールの入力
    await page.getByLabel('Markdown', { exact: true }).check();
    const rulesTextarea = page.locator('div.grid textarea');
    await rulesTextarea.fill('# 新ルール\n- 3ポイントシュートあり');

    // 保存
    const updateRequest = page.waitForRequest((request) => 
      request.url().endsWith('/api/admin/events/1/sports/1/details') && request.method() === 'PUT'
    );

    // ダイアログハンドリング (alert)
    page.on('dialog', dialog => dialog.accept());

    await page.getByRole('button', { name: /^保存$/ }).click();

    const request = await updateRequest;
    expect(JSON.parse(request.postData() ?? '{}')).toEqual({
      description: '新しいバスケットボールの概要',
      rules_type: 'markdown',
      rules: '# 新ルール\n- 3ポイントシュートあり',
      rules_pdf_url: null
    });

    // 完了メッセージの確認（ダイアログが閉じられた後）
    // handleSave 内で alert が呼ばれる
  });

  test('PDFによるルール登録ができる', async ({ page }) => {
    await page.getByLabel('競技選択').selectOption('1');

    await page.getByLabel('PDF', { exact: true }).check();

    // ファイルアップロードのモック（Playwrightの機能）
    const fileChooserPromise = page.waitForEvent('filechooser');
    await page.locator('input[type="file"]').click();
    const fileChooser = await fileChooserPromise;
    await fileChooser.setFiles({
      name: 'rules.pdf',
      mimeType: 'application/pdf',
      buffer: Buffer.from('%PDF-1.4 test')
    });

    const pdfUploadRequest = page.waitForRequest(req => 
      req.url().endsWith('/api/admin/pdfs') && req.method() === 'POST'
    );
    const updateRequest = page.waitForRequest(req => 
      req.url().endsWith('/api/admin/events/1/sports/1/details') && req.method() === 'PUT'
    );

    page.on('dialog', dialog => dialog.accept());
    await page.getByRole('button', { name: /^保存$/ }).click();

    await pdfUploadRequest;
    const req = await updateRequest;
    const body = JSON.parse(req.postData() ?? '{}');
    expect(body.rules_type).toBe('pdf');
    expect(body.rules_pdf_url).toBe('https://example.com/guidelines.pdf');
  });
});
