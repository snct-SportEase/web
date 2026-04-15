import { page, userEvent } from '@vitest/browser/context';
import { describe, expect, it, vi } from 'vitest';
import { render } from 'vitest-browser-svelte';
import AttendanceManagement from './+page.svelte';

describe('AttendanceManagement', () => {
  const mockClasses = [
    { id: 1, name: '1A' },
    { id: 2, name: '1B' }
  ];

  const mockClassDetails = {
    id: 1,
    name: '1A',
    studentCount: 40,
    attendancePoints: 10
  };

  it('初期表示が正しいこと (管理者)', async () => {
    render(AttendanceManagement, {
      props: {
        data: {
          classes: mockClasses,
          managedClass: null
        }
      }
    });

    await expect.element(page.getByRole('heading', { name: '出席点管理' })).toBeInTheDocument();
    await expect.element(page.getByRole('combobox', { name: '対象クラスを選択' })).toBeInTheDocument();
  });

  it('クラスを選択すると詳細が表示されること', async () => {
    // APIモック
    const fetchSpy = vi.spyOn(window, 'fetch').mockImplementation((url) => {
      if (url.includes('/api/admin/attendance/class-details/1')) {
        return Promise.resolve({
          ok: true,
          json: () => Promise.resolve(mockClassDetails)
        });
      }
      return Promise.resolve({ ok: false });
    });

    render(AttendanceManagement, {
      props: {
        data: {
          classes: mockClasses,
          managedClass: null
        }
      }
    });

    const selector = page.getByRole('combobox', { name: '対象クラスを選択' });
    await userEvent.selectOptions(selector, '1');
    
    // クラス名が見出しとして表示されるのを待機
    await expect.element(page.getByRole('heading', { name: '1A', exact: true })).toBeInTheDocument();
    await expect.element(page.getByText('クラスの総人数: 40人')).toBeInTheDocument();
    await expect.element(page.getByText('現在の出席ポイント: 10ポイント')).toBeInTheDocument();

    fetchSpy.mockRestore();
  });

  it('出席人数を入力して登録できること', async () => {
    vi.spyOn(window, 'fetch').mockImplementation((url) => {
      if (url.includes('/api/admin/attendance/class-details/1')) {
        return Promise.resolve({
          ok: true,
          json: () => Promise.resolve(mockClassDetails)
        });
      }
      if (url === '/api/admin/attendance/register') {
        return Promise.resolve({
          ok: true,
          json: () => Promise.resolve({ message: '出席を正常に登録しました。' })
        });
      }
      return Promise.resolve({ ok: false });
    });

    render(AttendanceManagement, {
      props: {
        data: {
          classes: mockClasses,
          managedClass: null
        }
      }
    });

    // クラス選択
    const selector = page.getByRole('combobox', { name: '対象クラスを選択' });
    await userEvent.selectOptions(selector, '1');
    
    // 入力
    const input = page.getByRole('spinbutton', { name: '出席人数' });
    await expect.element(input).toBeInTheDocument();
    await userEvent.fill(input, '35');

    // 送信
    const submitBtn = page.getByRole('button', { name: '出席を登録する' });
    await userEvent.click(submitBtn);

    // 成功メッセージ (曖昧マッチ)
    await expect.element(page.getByText(/出席を正常に登録しました/)).toBeInTheDocument();
  });

  it('総人数を超える入力をするとエラーが表示されること', async () => {
    vi.spyOn(window, 'fetch').mockImplementation((url) => {
      if (url.includes('/api/admin/attendance/class-details/1')) {
        return Promise.resolve({
          ok: true,
          json: () => Promise.resolve(mockClassDetails)
        });
      }
      return Promise.resolve({ ok: false });
    });

    render(AttendanceManagement, {
      props: {
        data: {
          classes: mockClasses,
          managedClass: null
        }
      }
    });

    const selector = page.getByRole('combobox', { name: '対象クラスを選択' });
    await userEvent.selectOptions(selector, '1');
    
    // 入力フォームが出るまで待つ
    const input = page.getByRole('spinbutton', { name: '出席人数' });
    await expect.element(input).toBeInTheDocument();
    
    // 値を入力
    await userEvent.fill(input, '45');
    
    // ボタンをクリック
    const submitBtn = page.getByRole('button', { name: '出席を登録する' });
    await userEvent.click(submitBtn);

    // エラーメッセージが表示されるのを待つ (曖昧マッチ)
    await expect.element(page.getByText(/出席人数がクラスの総人数.*超えています/)).toBeInTheDocument();
  });

  it('担当クラスがある場合は自動的に選択されること', async () => {
    vi.spyOn(window, 'fetch').mockImplementation((url) => {
      if (url.includes('/api/admin/attendance/class-details/2')) {
        return Promise.resolve({
          ok: true,
          json: () => Promise.resolve({
            id: 2,
            name: '1B',
            studentCount: 38,
            attendancePoints: 5
          })
        });
      }
      return Promise.resolve({ ok: false });
    });

    render(AttendanceManagement, {
      props: {
        data: {
          classes: mockClasses,
          managedClass: { id: 2, name: '1B' }
        }
      }
    });

    await expect.element(page.getByText('対象クラス: 1B')).toBeInTheDocument();
    await expect.element(page.getByText('クラスの総人数: 38人')).toBeInTheDocument();
  });
});
