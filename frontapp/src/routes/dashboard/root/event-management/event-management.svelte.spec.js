import { page } from '@vitest/browser/context';
import { describe, expect, it, vi, beforeEach } from 'vitest';
import { render } from 'vitest-browser-svelte';
import Page from './+page.svelte';

const originalCreateElement = document.createElement.bind(document);

// Mock activeEvent store
vi.mock('$lib/stores/eventStore.js', () => ({
  activeEvent: {
    init: vi.fn(),
    subscribe: vi.fn(() => () => {})
  }
}));

describe('Event Management Page', () => {
  const mockEvents = [
    {
      id: 1,
      name: '2025春季スポーツ大会',
      year: 2025,
      season: 'spring',
      start_date: '2025-04-01T00:00:00Z',
      end_date: '2025-04-02T00:00:00Z',
      status: 'upcoming',
      survey_url: 'https://example.com/survey',
      hide_scores: false
    }
  ];

  let fetchMock;
  let createObjectURLMock;
  let revokeObjectURLMock;
  let anchorClickMock;

  beforeEach(() => {
    vi.restoreAllMocks();

    fetchMock = vi.fn((url, options = {}) => {
      if (url === '/api/root/events') {
        if (options.method === 'POST') {
          return Promise.resolve({
            ok: true,
            json: () => Promise.resolve({ message: 'Event created' })
          });
        }

        return Promise.resolve({
          ok: true,
          json: () => Promise.resolve(mockEvents)
        });
      }

      if (url === '/api/root/events/1') {
        return Promise.resolve({
          ok: true,
          json: () => Promise.resolve({ message: 'Event updated' })
        });
      }

      if (url === '/api/root/events/1/notify-survey' && options.method === 'POST') {
        return Promise.resolve({
          ok: true,
          json: () => Promise.resolve({ message: 'Notification sent' })
        });
      }

      if (url === '/api/root/events/1/import-survey-scores' && options.method === 'POST') {
        return Promise.resolve({
          ok: true,
          json: () => Promise.resolve({ imported_classes_count: 2 })
        });
      }

      if (url === '/api/root/events/1/export/csv') {
        return Promise.resolve({
          ok: true,
          blob: () => Promise.resolve(new Blob(['class,score\n1A,100\n1B,90\n'], { type: 'text/csv' }))
        });
      }

      if (url === '/api/root/db/export') {
        return Promise.resolve({
          ok: true,
          blob: () => Promise.resolve(new Blob(['-- mock dump'], { type: 'application/sql' })),
          headers: {
            get: (name) => name === 'Content-Disposition' ? 'attachment; filename="mock_dump.sql"' : null
          }
        });
      }

      if (typeof url === 'string' && url.startsWith('/api/scores/class?')) {
        return Promise.resolve({
          ok: true,
          json: () => Promise.resolve([
            {
              class_name: '1A',
              rank_overall: 1,
              total_points_overall: 120,
              total_points_current_event: 60
            },
            {
              class_name: '1B',
              rank_overall: 2,
              total_points_overall: 100,
              total_points_current_event: 50
            }
          ])
        });
      }

      return Promise.resolve({ ok: true, json: () => Promise.resolve({}) });
    });

    createObjectURLMock = vi.fn(() => 'blob:mock-export');
    revokeObjectURLMock = vi.fn();
    anchorClickMock = vi.fn();

    vi.stubGlobal('fetch', fetchMock);
    vi.stubGlobal('alert', vi.fn());
    vi.stubGlobal('confirm', vi.fn(() => true));
    vi.stubGlobal('URL', {
      createObjectURL: createObjectURLMock,
      revokeObjectURL: revokeObjectURLMock
    });

    vi.spyOn(document, 'createElement').mockImplementation((tagName) => {
      const element = originalCreateElement(tagName);
      if (String(tagName).toLowerCase() === 'a') {
        element.click = anchorClickMock;
      }
      return element;
    });
  });

  it('初期表示で大会一覧が表示されること', async () => {
    render(Page);
    
    const title = page.getByRole('heading', { name: '大会情報登録・管理' });
    await expect.element(title).toBeInTheDocument();

    // モックデータが表示されているか確認
    const eventName = page.getByText('2025春季スポーツ大会');
    await expect.element(eventName).toBeInTheDocument();
  });

  it('「新規作成」ボタンをクリックするとモーダルが開くこと', async () => {
    render(Page);
    
    const createButton = page.getByRole('button', { name: '新規作成' });
    await createButton.click();

    const modalTitle = page.getByText('大会作成');
    await expect.element(modalTitle).toBeInTheDocument();
  });

  it('年度やシーズンを変更すると大会名が自動生成されること', async () => {
    render(Page);
    
    const createButton = page.getByRole('button', { name: '新規作成' });
    await createButton.click();

    const yearInput = page.getByRole('spinbutton', { name: '年度' });
    await yearInput.fill('2026');
    
    const seasonSelect = page.getByRole('combobox', { name: 'シーズン' });
    await seasonSelect.selectOptions('autumn');

    const nameInput = page.getByRole('textbox', { name: '大会名' });
    await expect.element(nameInput).toHaveValue('2026秋季スポーツ大会');
  });

  it('大会名を手動で変更した後は自動生成が停止すること', async () => {
    render(Page);
    
    const createButton = page.getByRole('button', { name: '新規作成' });
    await createButton.click();

    const nameInput = page.getByRole('textbox', { name: '大会名' });
    await nameInput.fill('カスタム大会名');
    
    const yearInput = page.getByRole('spinbutton', { name: '年度' });
    await yearInput.fill('2027');

    await expect.element(nameInput).toHaveValue('カスタム大会名');
  });

  it('イベント行をクリックすると編集モーダルが開くこと', async () => {
    render(Page);
    
    // 一覧の行をクリック
    const eventRow = page.getByText('2025春季スポーツ大会');
    await eventRow.click();

    const modalTitle = page.getByText('大会編集');
    await expect.element(modalTitle).toBeInTheDocument();

    const nameInput = page.getByRole('textbox', { name: '大会名' });
    await expect.element(nameInput).toHaveValue('2025春季スポーツ大会');
  });

  it('新規作成ではスコア非表示設定が初期値falseであること', async () => {
    render(Page);

    await page.getByRole('button', { name: '新規作成' }).click();

    const hideScoresCheckbox = page.getByLabelText('スコアを非表示にする');
    await expect.element(hideScoresCheckbox).not.toBeChecked();
  });

  it('新規作成を保存するとPOSTで大会情報を送信すること', async () => {
    render(Page);

    await page.getByRole('button', { name: '新規作成' }).click();
    await page.getByRole('spinbutton', { name: '年度' }).fill('2026');
    await page.getByRole('combobox', { name: 'シーズン' }).selectOptions('autumn');
    await page.getByLabelText('開始日').fill('2026-10-01');
    await page.getByLabelText('終了日').fill('2026-10-02');
    await page.getByRole('button', { name: '保存' }).click();

    const saveCall = fetchMock.mock.calls.find(([url, options]) => {
      return url === '/api/root/events' && options?.method === 'POST';
    });

    expect(saveCall).toBeTruthy();
    expect(saveCall[1]).toEqual(expect.objectContaining({
      method: 'POST',
      headers: { 'Content-Type': 'application/json' }
    }));
    expect(JSON.parse(saveCall[1].body)).toEqual(expect.objectContaining({
      name: '2026秋季スポーツ大会',
      year: 2026,
      season: 'autumn',
      start_date: '2026-10-01',
      end_date: '2026-10-02'
    }));
  });

  it('スコア非表示を有効にして保存するとhide_scores=trueで送信すること', async () => {
    render(Page);

    await page.getByRole('button', { name: '新規作成' }).click();
    await page.getByRole('spinbutton', { name: '年度' }).fill('2026');
    await page.getByLabelText('開始日').fill('2026-10-01');
    await page.getByLabelText('終了日').fill('2026-10-02');
    await page.getByLabelText('スコアを非表示にする').click();
    await page.getByRole('button', { name: '保存' }).click();

    const saveCall = fetchMock.mock.calls.find(([url, options]) => {
      return url === '/api/root/events' && options?.method === 'POST';
    });

    expect(saveCall).toBeTruthy();
    expect(JSON.parse(saveCall[1].body)).toEqual(expect.objectContaining({
      hide_scores: true
    }));
  });

  it('編集保存するとPUTで大会情報を送信すること', async () => {
    render(Page);

    await page.getByText('2025春季スポーツ大会').click();
    await page.getByRole('combobox', { name: 'ステータス' }).selectOptions('active');
    await page.getByRole('button', { name: '保存' }).click();

    const saveCall = fetchMock.mock.calls.find(([url, options]) => {
      return url === '/api/root/events/1' && options?.method === 'PUT';
    });

    expect(saveCall).toBeTruthy();
    expect(saveCall[1]).toEqual(expect.objectContaining({
      method: 'PUT',
      headers: { 'Content-Type': 'application/json' }
    }));
    expect(JSON.parse(saveCall[1].body)).toEqual(expect.objectContaining({
      id: 1,
      name: '2025春季スポーツ大会',
      status: 'active'
    }));
  });

  it('編集時にスコア非表示を有効にして保存するとPUTで反映されること', async () => {
    render(Page);

    await page.getByText('2025春季スポーツ大会').click();
    await page.getByLabelText('スコアを非表示にする').click();
    await page.getByRole('button', { name: '保存' }).click();

    const saveCall = fetchMock.mock.calls.find(([url, options]) => {
      return url === '/api/root/events/1' && options?.method === 'PUT';
    });

    expect(saveCall).toBeTruthy();
    expect(JSON.parse(saveCall[1].body)).toEqual(expect.objectContaining({
      id: 1,
      hide_scores: true
    }));
  });

  it('アンケート通知を送信すると全体通知APIを呼び出すこと', async () => {
    render(Page);

    await page.getByText('2025春季スポーツ大会').click();

    const notifyButton = page.getByRole('button', { name: '通知を送信' });
    await expect.element(notifyButton).toBeInTheDocument();
    await notifyButton.click();

    const notifyCall = fetchMock.mock.calls.find(([url, options]) => {
      return url === '/api/root/events/1/notify-survey' && options?.method === 'POST';
    });

    expect(notifyCall).toBeTruthy();
    expect(confirm).toHaveBeenCalledWith('アンケート通知を全ユーザーに送信します。よろしいですか？');
    expect(alert).toHaveBeenCalledWith('アンケート通知を送信しました。');
  });

  it('春季大会の得点CSVをインポートできること', async () => {
    const view = render(Page);

    await expect.element(page.getByText('点数インポート(CSV)')).toBeInTheDocument();

    const fileInput = view.container.querySelector('input[type="file"]');
    expect(fileInput).toBeTruthy();

    const file = new File(['class,score\n1A,100\n1B,90\n'], 'scores.csv', { type: 'text/csv' });
    const dataTransfer = new DataTransfer();
    dataTransfer.items.add(file);
    Object.defineProperty(fileInput, 'files', {
      value: dataTransfer.files,
      configurable: true
    });

    fileInput.dispatchEvent(new Event('change', { bubbles: true }));

    await vi.waitFor(() => {
      const uploadCall = fetchMock.mock.calls.find(([url, options]) => {
        return url === '/api/root/events/1/import-survey-scores' && options?.method === 'POST';
      });

      expect(uploadCall).toBeTruthy();
      expect(uploadCall[1].body).toBeInstanceOf(FormData);
      expect(alert).toHaveBeenCalledWith('インポート成功: 2 クラス分の点数が反映されました。');
      expect(fileInput.value).toBe('');
    });
  });

  it('クラス別スコア集計をCSV出力できること', async () => {
    render(Page);

    await page.getByRole('button', { name: 'CSV出力' }).click();

    await vi.waitFor(() => {
      const exportCall = fetchMock.mock.calls.find(([url]) => url === '/api/root/events/1/export/csv');
      expect(exportCall).toBeTruthy();
      expect(createObjectURLMock).toHaveBeenCalled();
      expect(anchorClickMock).toHaveBeenCalled();
      expect(revokeObjectURLMock).toHaveBeenCalledWith('blob:mock-export');
    });
  });

  it('クラス別スコア集計をPDF出力できること', async () => {
    render(Page);

    await page.getByRole('button', { name: 'PDF出力' }).click();

    await vi.waitFor(() => {
      const scoreCall = fetchMock.mock.calls.find(([url]) => url === '/api/scores/class?event_id=1');
      expect(scoreCall).toBeTruthy();
      expect(alert).not.toHaveBeenCalled();
    });
  });

  it('DBダンプを出力できること', async () => {
    render(Page);

    await page.getByRole('button', { name: 'DBダンプ出力' }).click();

    await vi.waitFor(() => {
      const dumpCall = fetchMock.mock.calls.find(([url]) => url === '/api/root/db/export');
      expect(dumpCall).toBeTruthy();
      expect(createObjectURLMock).toHaveBeenCalled();
      expect(anchorClickMock).toHaveBeenCalled();
      const anchor = document.createElement.mock.results.find((result) => result.value?.download === 'mock_dump.sql')?.value;
      expect(anchor?.download).toBe('mock_dump.sql');
      expect(revokeObjectURLMock).toHaveBeenCalledWith('blob:mock-export');
    });
  });
});
