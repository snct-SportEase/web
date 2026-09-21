import { page } from '@vitest/browser/context';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { render } from 'vitest-browser-svelte';
import Page from '$src/routes/dashboard/student/tournament/+page.svelte';

const mocks = vi.hoisted(() => ({
  active: {
    id: 1,
    name: '2025春季スポーツ大会',
    is_rainy_mode: false
  },
  createBracket: vi.fn((data, wrapper) => {
    wrapper.textContent = `表示済み: ${data.contestants.c0.players[0].title}`;
  })
}));

vi.mock('$lib/stores/eventStore.js', () => ({
  activeEvent: {
    subscribe(fn) {
      fn(mocks.active);
      return () => {};
    },
    init: vi.fn(async () => mocks.active)
  }
}));

vi.mock('bracketry', () => ({
  createBracket: mocks.createBracket
}));

describe('Student Tournament Page', () => {
  beforeEach(() => {
    mocks.createBracket.mockClear();

    vi.stubGlobal('fetch', vi.fn((url) => {
      if (url === '/api/student/events/1/tournaments') {
        return Promise.resolve({
          ok: true,
          json: () => Promise.resolve([
            {
              id: 1,
              name: 'バスケットボール Tournament',
              sport_id: 1,
              data: {
                rounds: [{ name: '決勝' }],
                matches: [{ roundIndex: 0, order: 0, sides: [{ contestantId: 'c0' }, { contestantId: 'c1' }] }],
                contestants: {
                  c0: { players: [{ title: '1A' }] },
                  c1: { players: [{ title: '1B' }] }
                }
              }
            },
            {
              id: 2,
              name: '将棋 Aブロック',
              sport_id: 10,
              data: {
                rounds: [{ name: '決勝' }],
                matches: [{ roundIndex: 0, order: 0, sides: [{ contestantId: 'c0' }, { contestantId: 'c1' }] }],
                contestants: {
                  c0: { players: [{ title: 'IE2 A' }] },
                  c1: { players: [{ title: 'IS2 A' }] }
                }
              }
            }
          ])
        });
      }

      if (url === '/api/student/events/1/board-game-runs') {
        return Promise.resolve({
          ok: true,
          json: () => new Promise((resolve) => setTimeout(() => resolve([
            {
              game_type: 'shogi',
              location: 'ICTメディア室',
              regular_minutes: 15,
              final_minutes: 30,
              win_points: 5,
              tournaments: [{ id: 2, entries: [], rankings: [] }]
            }
          ]), 20))
        });
      }

      return Promise.resolve({ ok: false, json: () => Promise.resolve({}) });
    }));
  });

  it('通常競技と盤上競技の表示領域がマウントされた後に各トーナメントを描画する', async () => {
    render(Page);

    await expect.element(page.getByRole('heading', { name: 'バスケットボール Tournament' })).toBeInTheDocument();
    await expect.element(page.getByText('表示済み: 1A')).toBeInTheDocument();
    await expect.element(page.getByRole('heading', { name: '将棋 Aブロック' })).toBeInTheDocument();
    await expect.element(page.getByText('会場: ICTメディア室')).toBeInTheDocument();
    await expect.element(page.getByText('表示済み: IE2 A')).toBeInTheDocument();
    expect(mocks.createBracket).toHaveBeenCalledTimes(2);
  });
});
