import { page } from '@vitest/browser/context';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { render } from 'vitest-browser-svelte';
import Page from '$src/routes/dashboard/student/sport-info/+page.svelte';

function jsonResponse(body) {
  return Promise.resolve({
    ok: true,
    json: () => Promise.resolve(body)
  });
}

describe('Student Sport Info Page', () => {
  let openMock;

  beforeEach(() => {
    vi.stubGlobal('fetch', vi.fn((url) => {
      if (url === '/api/events/active') {
        return jsonResponse({ event_id: 1 });
      }

      if (url === '/api/events/1/sports') {
        return jsonResponse([
          {
            sport_id: 1,
            sport_name: 'バスケットボール',
            description: '屋内競技',
            location: 'gym1',
            rules: '# 旧ルール',
            rules_type: 'markdown',
            rules_pdf_url: '/uploads/rules/basketball.pdf'
          },
          {
            sport_id: 2,
            sport_name: 'バレーボール',
            location: 'gym2',
            rules: '# 旧ルール',
            rules_type: 'markdown',
            rules_pdf_url: null
          }
        ]);
      }

      return jsonResponse({});
    }));

    openMock = vi.fn();
    vi.stubGlobal('open', openMock);
  });

  afterEach(() => {
    vi.unstubAllGlobals();
  });

  it('PDFが登録された競技だけルールPDFを別タブで開ける', async () => {
    render(Page);

    const pdfButton = page.getByRole('button', { name: 'ルールPDFを見る' });
    await expect.element(pdfButton).toBeInTheDocument();
    await expect.element(page.getByText('旧ルール')).not.toBeInTheDocument();

    await pdfButton.click();

    expect(openMock).toHaveBeenCalledWith(
      '/uploads/rules/basketball.pdf',
      '_blank',
      'noopener,noreferrer'
    );
    await expect.element(page.getByRole('button', { name: 'ルールPDFを見る' })).toHaveLength(1);
  });
});
