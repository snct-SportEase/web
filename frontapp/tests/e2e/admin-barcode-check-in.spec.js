import { expect, test } from '@playwright/test';

const mockBackendUrl =
  process.env.MOCK_BACKEND_URL ??
  `http://127.0.0.1:${process.env.MOCK_BACKEND_PORT ?? 8081}`;

test.describe.configure({ mode: 'serial' });

async function setupAdminBarcodePage({ page, context, request }) {
  await request.post(`${mockBackendUrl}/__reset`);
  await request.post(`${mockBackendUrl}/__set-user`, { data: { user: 'admin' } });
  await request.post(`${mockBackendUrl}/api/admin/events/1/sports`, {
    data: {
      sport_id: 1,
      location: 'gym1',
      description: '屋内競技',
    }
  });

  await context.addCookies([{
    name: 'session_token',
    value: 'test-session-token',
    domain: 'localhost',
    path: '/'
  }]);

  await page.goto('/dashboard/admin/barcode-reader');
  await expect(page.getByRole('heading', { name: 'MyIDバーコード読み取り' })).toBeVisible();
}

test.describe('MyIDバーコード読み取り (admin)', () => {
  test.beforeEach(setupAdminBarcodePage);

  test('手入力したMyIDバーコードでラウンドチェックインできる', async ({ page }) => {
    await page.getByLabel('競技').selectOption('1');
    await page.getByLabel('試合').selectOption({ index: 1 });
    await page.getByLabel('バーコード値').fill('H1023010590');

    const checkInRequest = page.waitForRequest((request) => {
      if (!request.url().endsWith('/api/barcode/check-in') || request.method() !== 'POST') {
        return false;
      }

      const body = JSON.parse(request.postData() ?? '{}');
      return body.barcode_data === 'H1023010590'
        && body.event_id === 1
        && body.sport_id === 1
        && body.match_id === 1;
    });

    await page.getByRole('button', { name: 'チェックインする' }).click();

    await checkInRequest;
    const successDialog = page.getByRole('dialog', { name: 'ラウンドチェックインを完了しました' });
    await expect(successDialog).toBeVisible();
    await expect(page.getByText('氏名: 山田太郎')).toBeVisible();
    await expect(page.getByText('学籍番号: 2301059')).toBeVisible();
    await expect(page.getByText('競技: バスケットボール')).toBeVisible();
    await expect(page.getByText('ラウンド: 1').nth(1)).toBeVisible();
    await expect(page.getByLabel('バーコード値')).toHaveValue('');
    await successDialog.getByRole('button', { name: '閉じる' }).click();
    await expect(page.getByText('この試合のチェックイン済み')).toBeVisible();
    await expect(page.getByText('1 人')).toBeVisible();
    await expect(page.getByText('s2301059@sendai-nct.jp')).toBeVisible();

    await page.getByRole('button', { name: 'チェックイン済み（1人）' }).click();
    const checkedDialog = page.getByRole('dialog', { name: 'チェックイン済みの学生' });
    await expect(checkedDialog).toBeVisible();
    await expect(checkedDialog.getByText('山田太郎')).toBeVisible();
    await checkedDialog.getByRole('button', { name: '閉じる' }).click();

    await page.getByRole('button', { name: '未チェックイン（1人）' }).click();
    const uncheckedDialog = page.getByRole('dialog', { name: '未チェックインの学生' });
    await expect(uncheckedDialog).toBeVisible();
    await expect(uncheckedDialog.getByText('佐藤花子')).toBeVisible();
    await uncheckedDialog.getByRole('button', { name: '閉じる' }).click();
  });

  test('MyID形式ではないバーコードはrejectされる', async ({ page }) => {
    await page.getByLabel('競技').selectOption('1');
    await page.getByLabel('試合').selectOption({ index: 1 });
    await page.getByLabel('バーコード値').fill('2301059');

    await page.getByRole('button', { name: 'チェックインする' }).click();

    await expect(page.getByText('チェックインできませんでした')).toBeVisible();
    await expect(page.getByText('バーコード形式が不正です')).toBeVisible();
  });
});
