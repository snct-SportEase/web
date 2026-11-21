<script lang="ts">
  import { onMount } from 'svelte';
  import { get } from 'svelte/store';
  import { activeEvent } from '$lib/stores/eventStore.js';
  import { fetchPublishedNoonGameSessions, flattenNoonGameMatches } from '$lib/utils/noonGameSessions.js';
  import type { NoonGameMatch } from '$lib/utils/noonGameSessions.js';
  import type { BracketData, Match, Side } from '@snct-sportease/matchline';

  type TournamentMatch = Omit<Match, 'id'> & { id?: number | string };
  type TournamentData = Omit<BracketData, 'matches'> & { matches?: TournamentMatch[] };
  type Tournament = { id: number; name: string; data: TournamentData | null };
  type TournamentResponse = Omit<Tournament, 'data'> & { data: TournamentData | string | null };
  type NoonEntry = NonNullable<NoonGameMatch['entries']>[number];
  type NoonDetail = NonNullable<NonNullable<NoonGameMatch['result']>['details']>[number];

  let loading = $state(true);
  let error = $state('');
  let tournamentResults = $state<{ tournament: Tournament; match: TournamentMatch }[]>([]);
  let noonGameMatches = $state<(NoonGameMatch & { session_name: string })[]>([]);

  function normalizeTournament(tournament: TournamentResponse): Tournament {
    if (typeof tournament.data !== 'string') return { ...tournament, data: tournament.data };
    try {
      return { ...tournament, data: JSON.parse(tournament.data) };
    } catch {
      return { ...tournament, data: null };
    }
  }

  function sideName(tournament: Tournament, side: Side | undefined): string {
    if (side?.title) return side.title;
    return (side?.contestantId && tournament.data?.contestants?.[side.contestantId]?.players?.[0]?.title) || '未定';
  }

  function score(side: Side | undefined): number | string | undefined {
    return side?.scores?.[0]?.mainScore;
  }

  function completedMatch(match: TournamentMatch): boolean {
    return match?.sides?.some((side) => score(side) !== undefined) ?? false;
  }

  function tournamentMatchResults(tournaments: Tournament[]) {
    return tournaments.flatMap((tournament) =>
      (tournament.data?.matches || [])
        .filter(completedMatch)
        .map((match) => ({ tournament, match }))
    );
  }

  function formatStatus(status: string | undefined): string {
    const labels: Record<string, string> = { scheduled: '予定', in_progress: '進行中', completed: '終了', finished: '終了', cancelled: '中止' };
    return (status && labels[status]) || status || '未定';
  }

  function entryName(entry: NoonEntry | undefined): string {
    return entry?.resolved_name || entry?.display_name || '参加者未定';
  }

  function resultEntries(match: NoonGameMatch): NoonDetail[] {
    return [...(match?.result?.details || [])].sort((a, b) => (a.rank || 999) - (b.rank || 999));
  }

  function resultEntryName(detail: NoonDetail, match: NoonGameMatch): string {
    return detail?.entry_resolved_name || entryName(match?.entries?.find((entry) => String(entry.id) === String(detail?.entry_id)));
  }

  onMount(async () => {
    try {
      await activeEvent.init();
      const event = get(activeEvent);
      if (!event?.id) return;

      const [tournamentResponse, sessions] = await Promise.all([
        fetch(`/api/student/events/${event.id}/tournaments`),
        fetchPublishedNoonGameSessions(event.id)
      ]);
      if (!tournamentResponse.ok) throw new Error('通常競技の結果を取得できませんでした。');

      const payload: TournamentResponse[] = await tournamentResponse.json();
      const tournaments = payload.map(normalizeTournament);
      tournamentResults = tournamentMatchResults(tournaments);
      noonGameMatches = flattenNoonGameMatches(sessions);
    } catch (cause) {
      error = cause instanceof Error ? cause.message : '結果一覧を取得できませんでした。';
    } finally {
      loading = false;
    }
  });
</script>

<div class="mx-auto max-w-6xl space-y-8 p-4 md:p-8">
  <div>
    <h1 class="text-2xl font-bold text-gray-900">結果一覧</h1>
    <p class="mt-1 text-sm text-gray-600">通常競技と昼競技の結果をまとめて確認できます。</p>
  </div>

  {#if loading}
    <p class="rounded-md border border-gray-200 bg-gray-50 px-4 py-3 text-sm text-gray-600">読み込み中です。</p>
  {:else if error}
    <p class="rounded-md border border-red-200 bg-red-50 px-4 py-3 text-sm text-red-700">{error}</p>
  {:else}
    <section class="space-y-4">
      <h2 class="text-xl font-semibold text-gray-900">通常競技</h2>
      {#if tournamentResults.length === 0}
        <p class="rounded-md border border-gray-200 bg-gray-50 px-4 py-3 text-sm text-gray-600">結果が登録された通常競技はありません。</p>
      {:else}
        <div class="space-y-3">
          {#each tournamentResults as { tournament, match } (match.id)}
            {@const home = match.sides?.[0]}
            {@const away = match.sides?.[1]}
            <article class="rounded-lg border border-gray-200 bg-white p-4 shadow-sm">
              <h3 class="font-semibold text-gray-900">{tournament.name}</h3>
              <p class="mt-1 text-sm text-gray-500">{match.isBronzeMatch ? '3位決定戦' : `${(match.roundIndex ?? 0) + 1}回戦`}</p>
              <div class="mt-3 grid grid-cols-[1fr_auto_1fr] items-center gap-3 text-sm">
                <span class:font-bold={home?.isWinner}>{sideName(tournament, home)}</span>
                <span class="font-semibold text-indigo-700">{score(home) ?? '-'} - {score(away) ?? '-'}</span>
                <span class:font-bold={away?.isWinner} class="text-right">{sideName(tournament, away)}</span>
              </div>
            </article>
          {/each}
        </div>
      {/if}
    </section>

    <section class="space-y-4">
      <h2 class="text-xl font-semibold text-gray-900">昼競技</h2>
      {#if noonGameMatches.length === 0}
        <p class="rounded-md border border-gray-200 bg-gray-50 px-4 py-3 text-sm text-gray-600">表示できる昼競技はありません。</p>
      {:else}
        <div class="grid gap-4 lg:grid-cols-2">
          {#each noonGameMatches as match (match.id)}
            <article class="rounded-lg border border-indigo-100 bg-white p-4 shadow-sm">
              <h3 class="font-semibold text-gray-900">{match.title || match.session_name || '昼競技'}</h3>
              <p class="mt-1 text-sm text-gray-600">ステータス: {formatStatus(match.status)}</p>
              {#if resultEntries(match).length > 0}
                <ol class="mt-3 space-y-2">
                  {#each resultEntries(match) as detail (detail.id || detail.entry_id)}
                    <li class="flex justify-between rounded bg-indigo-50 px-3 py-2 text-sm">
                      <span class="font-semibold">{detail.rank ? `${detail.rank}位` : '-'}</span>
                      <span>{resultEntryName(detail, match)}</span>
                      <span class="font-semibold">{detail.points}点</span>
                    </li>
                  {/each}
                </ol>
              {:else if match.winner_display}
                <p class="mt-3 text-sm text-gray-700">勝者: {match.winner_display}</p>
              {:else}
                <p class="mt-3 text-sm text-gray-500">結果はまだ登録されていません。</p>
              {/if}
            </article>
          {/each}
        </div>
      {/if}
    </section>
  {/if}
</div>
