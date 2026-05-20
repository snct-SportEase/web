import { page } from '@vitest/browser/context';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { render } from 'vitest-browser-svelte';
import Page from './+page.svelte';

vi.mock('$app/stores', async () => {
  const { readable } = await import('svelte/store');

  return {
    page: readable({
      params: {
        eventId: '1'
      }
    })
  };
});

describe('Archive Event Detail Page', () => {
  let mockTournaments;
  let mockRelayMatches;
  let fetchMock;

  beforeEach(() => {
    vi.restoreAllMocks();

    mockTournaments = [
      {
        id: 10,
        name: 'バスケットボール',
        sport_id: 1,
        data: JSON.stringify({
          rounds: [{ name: 'Round 1' }],
          matches: [
            {
              id: 101,
              roundIndex: 0,
              order: 0,
              sides: [
                {
                  contestantId: 'c1',
                  scores: [{ mainScore: 3 }],
                  isWinner: true
                },
                {
                  contestantId: 'c2',
                  scores: [{ mainScore: 1 }]
                }
              ]
            }
          ],
          contestants: {
            c1: { players: [{ title: '1A' }] },
            c2: { players: [{ title: '1B' }] }
          }
        })
      }
    ];

    mockRelayMatches = [
      {
        id: 201,
        title: '学年対抗リレー Aブロック',
        status: 'finished',
        scheduled_at: '2025-04-01T09:00:00Z',
        location: 'グラウンド',
        format: '順位決定',
        entries: [
          { id: 1, resolved_name: '1A' },
          { id: 2, resolved_name: '1B' }
        ],
        result: {
          details: [
            { id: 11, entry_id: 1, entry_resolved_name: '1A', rank: 1, points: 10 },
            { id: 12, entry_id: 2, entry_resolved_name: '1B', rank: 2, points: 8 }
          ]
        }
      }
    ];

    fetchMock = vi.fn((url) => {
      if (url === '/api/events') {
        return Promise.resolve({
          ok: true,
          json: () =>
            Promise.resolve([
              {
                id: 1,
                name: '2025春季スポーツ大会',
                season: 'spring',
                status: 'archived'
              }
            ])
        });
      }

      if (url === '/api/scores/class?event_id=1') {
        return Promise.resolve({
          ok: true,
          json: () => Promise.resolve([])
        });
      }

      if (url === '/api/student/events/1/tournaments') {
        return Promise.resolve({
          ok: true,
          json: () => Promise.resolve(mockTournaments)
        });
      }

      if (url === '/api/student/events/1/noon-game/session') {
        return Promise.resolve({
          ok: true,
          json: () => Promise.resolve({
            matches: mockRelayMatches
          })
        });
      }

      return Promise.resolve({
        ok: false,
        json: () => Promise.resolve({})
      });
    });

    vi.stubGlobal('fetch', fetchMock);
  });

  it('トーナメントAPIが配列を返したときアーカイブで試合一覧を表示できること', async () => {
    render(Page);

    const tournamentTab = page.getByRole('button', { name: '試合結果' });
    await tournamentTab.click();

    await expect.element(page.getByText('バスケットボール')).toBeInTheDocument();
    await expect.element(page.getByText('#101')).toBeInTheDocument();
    await expect.element(page.getByRole('cell', { name: '1A' }).first()).toBeInTheDocument();
    await expect.element(page.getByRole('cell', { name: '1B' })).toBeInTheDocument();
    await expect.element(page.getByText('3 - 1')).toBeInTheDocument();
    await expect.element(page.getByText('試合結果データが見つかりませんでした。')).not.toBeInTheDocument();
  });

  it('トーナメントが空配列なら未検出メッセージを表示すること', async () => {
    mockTournaments = [];

    render(Page);

    const tournamentTab = page.getByRole('button', { name: '試合結果' });
    await tournamentTab.click();

    await expect.element(page.getByText('試合結果データが見つかりませんでした。')).toBeInTheDocument();
  });

  it('リレー結果タブでリレーの順位結果を表示できること', async () => {
    render(Page);

    const relayTab = page.getByRole('button', { name: 'リレー結果' });
    await relayTab.click();

    await expect.element(page.getByText('学年対抗リレー Aブロック')).toBeInTheDocument();
    await expect.element(page.getByText('ステータス: 終了')).toBeInTheDocument();
    await expect.element(page.getByText('1位 1A')).toBeInTheDocument();
    await expect.element(page.getByText('10 点')).toBeInTheDocument();
    await expect.element(page.getByText('2位 1B')).toBeInTheDocument();
  });
});
