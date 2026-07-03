import { expect, test } from '@playwright/test';

test.describe.configure({ mode: 'serial' });

const mockBackendUrl = process.env.MOCK_BACKEND_URL ?? 'http://127.0.0.1:8081';

test.describe('昼競技管理 (root)', () => {
  test.beforeEach(async ({ page, context, request }) => {
    await request.post(`${mockBackendUrl}/__reset`);
    page.on('console', (msg) => {
      console.log(`[Browser Console] ${msg.type()}: ${msg.text()}`);
    });
    page.on('dialog', (dialog) => {
      void dialog.accept().catch(() => {});
    });

    await context.addCookies([{ name: 'session_token', value: 'test-session-token', domain: 'localhost', path: '/' }]);
    await page.goto('/dashboard/root/noon-game');
    await expect(page.getByRole('heading', { name: '昼競技管理' })).toBeVisible();
    // activeEvent.init() とクライアント側の初期化完了を待つ
    await expect(page.getByRole('button', { name: /セッションを(作成|更新)/ })).toBeEnabled({ timeout: 15000 });
  });

  test('昼競技セッションを作成できる', async ({ page }) => {
    const requestPromise = page.waitForRequest((request) => request.url().includes('/noon-game/session') && request.method() === 'POST');
    const inputSelector = 'input[placeholder="例: 昼休み競技 2025"]';
    await page.locator(inputSelector).evaluate((el, val) => {
      el.value = val;
      el.dispatchEvent(new Event('input', { bubbles: true }));
      el.dispatchEvent(new Event('change', { bubbles: true }));
    }, '昼休み競技 2025');
    await page.locator(inputSelector).blur();
    await expect(page.getByRole('button', { name: 'セッションを作成' })).toBeEnabled();
    // 少し待機してイベントハンドラーのアタッチを確実にする
    await page.waitForTimeout(500);
    console.log('[E2E] Clicking button...');
    await page.locator('section', { hasText: 'テンプレートを使用しない場合の設定' }).getByRole('button', { name: 'セッションを作成' }).click();
    console.log('[E2E] Button clicked.');
    const req = await requestPromise;
    expect(JSON.parse(req.postData() ?? '{}')).toEqual(expect.objectContaining({
      name: '昼休み競技 2025',
      mode: 'mixed'
    }));
  });

  test('テンプレートを実行できる', async ({ page }) => {
    const openButton = page.getByRole('button', { name: 'テンプレートを設定' }).nth(1);
    await openButton.click();
    await expect(page.getByRole('heading', { name: /コース対抗リレー.*テンプレート設定/ })).toBeVisible();

    const requestPromise = page.waitForRequest((request) => request.url().endsWith('/api/root/events/1/noon-game/templates/course-relay/run') && request.method() === 'POST');
    await page.getByRole('button', { name: 'テンプレートを作成' }).click();
    await requestPromise;
  });

  test('グループ編集でクリックで複数選択と解除ができる', async ({ page }) => {
    await page.route('**/api/root/events/1/noon-game/session', async (route) => {
      if (route.request().method() !== 'POST') {
        await route.continue();
        return;
      }

      const body = JSON.parse(route.request().postData() ?? '{}');
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          session: {
            id: 1,
            ...body
          },
          classes: [
            { id: 1, name: '1A' },
            { id: 2, name: '1B' }
          ],
          groups: [],
          matches: [],
          points_summary: [],
          template_runs: []
        })
      });
    });

    await page.route('**/api/root/noon-game/sessions/1/groups', async (route) => {
      const body = JSON.parse(route.request().postData() ?? '{}');
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          group: {
            id: 1,
            name: body.name,
            description: body.description ?? null,
            members: (body.class_ids ?? []).map((classId, index) => ({
              id: index + 1,
              group_id: 1,
              class_id: classId,
              weight: 1,
              class: {
                id: classId,
                name: classId === 1 ? '1A' : '1B'
              }
            }))
          }
        })
      });
    });

    await page.route('**/api/root/noon-game/sessions/1/groups/1', async (route) => {
      const body = JSON.parse(route.request().postData() ?? '{}');
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          group: {
            id: 1,
            name: body.name,
            description: body.description ?? null,
            members: (body.class_ids ?? []).map((classId, index) => ({
              id: index + 1,
              group_id: 1,
              class_id: classId,
              weight: 1,
              class: {
                id: classId,
                name: classId === 1 ? '1A' : '1B'
              }
            }))
          }
        })
      });
    });

    const inputSelector = 'input[placeholder="例: 昼休み競技 2025"]';
    await page.locator(inputSelector).evaluate((el, val) => {
      el.value = val;
      el.dispatchEvent(new Event('input', { bubbles: true }));
      el.dispatchEvent(new Event('change', { bubbles: true }));
    }, '昼休み競技 2025');
    await page.locator(inputSelector).blur();
    await page.getByRole('button', { name: 'セッションを作成' }).click();

    await expect(page.getByRole('heading', { name: '現在の昼競技設定' })).toBeVisible();
    await page.getByLabel('グループ名').fill('機械系');
    await page.getByRole('button', { name: '1A' }).click();
    await page.getByRole('button', { name: 'グループを登録' }).click();

    const machineGroup = page.locator('li', { hasText: '機械系' });
    await expect(machineGroup).toBeVisible();

    await machineGroup.getByRole('button', { name: '編集' }).click();
    await expect(page.getByRole('button', { name: '1A' })).toHaveAttribute('aria-pressed', 'true');
    await expect(page.getByRole('button', { name: '1B' })).toHaveAttribute('aria-pressed', 'false');

    const updateGroupRequest = page.waitForRequest((request) =>
      request.url().endsWith('/api/root/noon-game/sessions/1/groups/1') && request.method() === 'PUT'
    );
    await page.getByRole('button', { name: '1B' }).click();
    await page.getByRole('button', { name: '1A' }).click();
    await page.getByRole('button', { name: 'グループを更新' }).click();

    const req = await updateGroupRequest;
    expect(JSON.parse(req.postData() ?? '{}')).toEqual(expect.objectContaining({
      class_ids: [2]
    }));
    await expect(machineGroup.getByText('メンバー: 1B')).toBeVisible();
  });
});
