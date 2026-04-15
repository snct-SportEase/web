import { page, userEvent } from '@vitest/browser/context';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { render } from 'vitest-browser-svelte';
import Page from './+page.svelte';

const classes = [
  { id: 1, name: '1A' },
  { id: 2, name: '1B' }
];

const sports = [
  { id: 1, name: 'バスケットボール' },
  { id: 2, name: 'バレーボール' }
];

function jsonResponse(body) {
  return Promise.resolve({
    ok: true,
    json: () => Promise.resolve(body)
  });
}

describe('Sport Details Registration Page', () => {
  let fetchMock;
  let alertMock;
  let rainyModeSettings;

  beforeEach(() => {
    rainyModeSettings = [];

    fetchMock = vi.fn((url, options = {}) => {
      if (url === '/api/events/active') {
        return jsonResponse({
          event_id: 1,
          event_name: '2025春季スポーツ大会'
        });
      }

      if (url === '/api/root/events') {
        return jsonResponse([
          {
            id: 1,
            name: '2025春季スポーツ大会',
            is_rainy_mode: false
          }
        ]);
      }

      if (url === '/api/admin/allsports') {
        return jsonResponse(sports);
      }

      if (url === '/api/admin/class-team/managed-class') {
        return jsonResponse(classes);
      }

      if (url === '/api/admin/events/1/tournaments') {
        return jsonResponse([]);
      }

      if (url === '/api/admin/events/1/sports/1/details') {
        return jsonResponse({
          description: '',
          rules: '',
          rules_type: 'markdown',
          rules_pdf_url: null,
          min_capacity: null,
          max_capacity: null
        });
      }

      if (url === '/api/root/sports/1/teams') {
        return jsonResponse([]);
      }

      if (url === '/api/root/events/1/rainy-mode/settings' && !options.method) {
        return jsonResponse(rainyModeSettings);
      }

      if (url === '/api/root/events/1/rainy-mode/settings' && options.method === 'POST') {
        const body = JSON.parse(options.body);
        const nextSetting = {
          event_id: 1,
          sport_id: Number(body.sport_id),
          class_id: Number(body.class_id),
          min_capacity: body.min_capacity,
          max_capacity: body.max_capacity
        };

        rainyModeSettings = [
          ...rainyModeSettings.filter(
            (item) => !(item.sport_id === nextSetting.sport_id && item.class_id === nextSetting.class_id)
          ),
          nextSetting
        ];

        return jsonResponse(nextSetting);
      }

      return jsonResponse({});
    });

    alertMock = vi.fn();
    vi.stubGlobal('fetch', fetchMock);
    vi.stubGlobal('alert', alertMock);
  });

  afterEach(() => {
    vi.unstubAllGlobals();
  });

  it('一括設定の雨天時定員を保存できる', async () => {
    rainyModeSettings = [
      { event_id: 1, sport_id: 1, class_id: 1, min_capacity: 3, max_capacity: 5 },
      { event_id: 1, sport_id: 1, class_id: 2, min_capacity: 3, max_capacity: 5 }
    ];

    render(Page);

    await expect.element(page.getByRole('heading', { name: '競技詳細情報登録' })).toBeInTheDocument();
    await expect.element(page.getByRole('option', { name: 'バスケットボール' })).toBeInTheDocument();

    await page.getByLabelText('競技選択').selectOptions('1');

    await expect.element(page.getByText('現在の設定: 定員 3 〜 5')).toBeInTheDocument();

    await page.getByRole('button', { name: '雨天時定員設定を保存' }).click();

    const saveCalls = fetchMock.mock.calls.filter(([url, options]) => {
      return url === '/api/root/events/1/rainy-mode/settings' && options?.method === 'POST';
    });

    const requestBodies = saveCalls
      .map(([, options]) => JSON.parse(options.body))
      .sort((left, right) => Number(left.class_id) - Number(right.class_id));

    expect(requestBodies).toEqual([
      { sport_id: '1', class_id: 1, min_capacity: 3, max_capacity: 5 },
      { sport_id: '1', class_id: 2, min_capacity: 3, max_capacity: 5 }
    ]);
    expect(alertMock).toHaveBeenCalledWith('雨天時定員設定を更新しました。');
  });

  it('クラスごとの雨天時定員を保存できる', async () => {
    rainyModeSettings = [
      { event_id: 1, sport_id: 1, class_id: 1, min_capacity: 2, max_capacity: 4 },
      { event_id: 1, sport_id: 1, class_id: 2, min_capacity: 5, max_capacity: 7 }
    ];

    render(Page);

    await expect.element(page.getByRole('option', { name: 'バスケットボール' })).toBeInTheDocument();
    await page.getByLabelText('競技選択').selectOptions('1');

    const saveButton = page.getByRole('button', { name: 'すべてのクラスの雨天時定員設定を保存' });
    await expect.element(saveButton).toBeInTheDocument();

    await saveButton.click();

    const saveCalls = fetchMock.mock.calls.filter(([url, options]) => {
      return url === '/api/root/events/1/rainy-mode/settings' && options?.method === 'POST';
    });

    const requestBodies = saveCalls
      .map(([, options]) => JSON.parse(options.body))
      .sort((left, right) => Number(left.class_id) - Number(right.class_id));

    expect(requestBodies).toEqual([
      { sport_id: '1', class_id: '1', min_capacity: 2, max_capacity: 4 },
      { sport_id: '1', class_id: '2', min_capacity: 5, max_capacity: 7 }
    ]);
    expect(alertMock).toHaveBeenCalledWith('すべてのクラスの雨天時定員設定を更新しました。');
  });

  it('競技の概要とルール(Markdown)を保存できる', async () => {
    render(Page);

    await expect.element(page.getByRole('option', { name: 'バスケットボール' })).toBeInTheDocument();
    await page.getByLabelText('競技選択').selectOptions('1');

    // 競技概要 (1番目のtextarea)
    const textareas = page.getByRole('textbox');
    await textareas.nth(0).fill('バスケットボールの概要です。');

    // ルール形式をMarkdownに
    await userEvent.click(page.getByLabelText('Markdown', { exact: true }));

    // ルール詳細 (最後のtextarea)
    await textareas.last().fill('# ルール\n- 5人制\n- ドリブル必須');

    const saveButton = page.getByRole('button', { name: /^保存$/ });
    await saveButton.click();

    const saveCall = fetchMock.mock.calls.find(([url, options]) => {
      return url === '/api/admin/events/1/sports/1/details' && options?.method === 'PUT';
    });

    expect(saveCall).toBeTruthy();
    expect(JSON.parse(saveCall[1].body)).toEqual({
      description: 'バスケットボールの概要です。',
      rules_type: 'markdown',
      rules: '# ルール\n- 5人制\n- ドリブル必須',
      rules_pdf_url: null
    });
    expect(alertMock).toHaveBeenCalledWith('Sport details saved successfully');
  });
});
