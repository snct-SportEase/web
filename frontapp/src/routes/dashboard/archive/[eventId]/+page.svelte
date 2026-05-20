<script>
    import { onMount } from 'svelte';
    import { page } from '$app/stores';
    
    let eventId = $page.params.eventId;
    let activeTab = $state('scores');
    
    let eventData = $state(null);
    let scores = $state([]);
    let tournaments = $state([]);
    let relayMatches = $state([]);
    let relayError = $state('');
    let loading = $state(true);
    let error = $state(null);

    function normalizeTournament(tournament) {
        const normalized = { ...tournament };

        if (typeof normalized.data === 'string') {
            try {
                normalized.data = JSON.parse(normalized.data);
            } catch (parseError) {
                console.error('Failed to parse tournament data:', parseError);
                normalized.data = null;
            }
        }

        return normalized;
    }

    function getSideName(tournament, side) {
        if (!side) return '未定';
        if (side.title) return side.title;

        const contestant = tournament?.data?.contestants?.[side.contestantId];
        const playerTitle = contestant?.players?.[0]?.title;
        return playerTitle || '未定';
    }

    function getSideScore(side) {
        return side?.scores?.[0]?.mainScore;
    }

    function isMatchCompleted(match) {
        return match?.sides?.some((side) => getSideScore(side) !== undefined);
    }

    function getWinnerName(tournament, match) {
        const winnerSide = match?.sides?.find((side) => side?.isWinner);
        return winnerSide ? getSideName(tournament, winnerSide) : '引き分け';
    }

    function isRelayMatch(match) {
        const title = match?.title || '';
        return title.includes('リレー');
    }

    function formatDateTime(value) {
        if (!value) return '日時未定';
        const date = new Date(value);
        if (Number.isNaN(date.getTime())) return '日時未定';
        return new Intl.DateTimeFormat('ja-JP', {
            month: 'numeric',
            day: 'numeric',
            weekday: 'short',
            hour: '2-digit',
            minute: '2-digit'
        }).format(date);
    }

    function formatStatus(status) {
        const labels = {
            scheduled: '予定',
            in_progress: '進行中',
            finished: '終了',
            cancelled: '中止'
        };
        return labels[status] || status || '未定';
    }

    function entryName(entry) {
        return entry?.resolved_name || entry?.display_name || '参加者未定';
    }

    function resultEntries(match) {
        return [...(match?.result?.details || [])].sort(
            (a, b) => (a.rank || 999) - (b.rank || 999)
        );
    }

    function resultEntryName(detail, match) {
        if (detail?.entry_resolved_name) return detail.entry_resolved_name;
        const entry = match?.entries?.find((item) => String(item.id) === String(detail?.entry_id));
        return entryName(entry);
    }

    onMount(async () => {
        try {
            // Fetch event info
            const eventsRes = await fetch('/api/events');
            if (eventsRes.ok) {
                const allEvents = await eventsRes.json();
                eventData = allEvents.find(e => e.id === parseInt(eventId));
            }

            // Fetch scores
            const scoreRes = await fetch(`/api/scores/class?event_id=${eventId}`);
            if (scoreRes.ok) {
                scores = await scoreRes.json();
            }

            // Fetch tournaments
            const tournRes = await fetch(`/api/student/events/${eventId}/tournaments`);
            if (tournRes.ok) {
                const fetchedTournaments = await tournRes.json();
                tournaments = Array.isArray(fetchedTournaments)
                    ? fetchedTournaments.map(normalizeTournament)
                    : [];
            }

            const relayRes = await fetch(`/api/student/events/${eventId}/noon-game/session`);
            if (relayRes.ok) {
                const relayPayload = await relayRes.json();
                relayMatches = (relayPayload.matches || []).filter(isRelayMatch);
            } else {
                const relayDetail = await relayRes.json().catch(() => null);
                relayError = relayDetail?.error || 'リレー結果を取得できませんでした。';
            }
        } catch (err) {
            error = err.message;
        } finally {
            loading = false;
        }
    });

    let season = $derived(eventData ? eventData.season : (scores.length > 0 ? scores[0].season : ''));

    // Simplistic rank sorting
    let sortedScores = $derived([...scores].sort((a, b) => {
        const rankA = season === 'spring' ? a.rank_current_event : a.rank_overall;
        const rankB = season === 'spring' ? b.rank_current_event : b.rank_overall;
        const normalizedRankA = (rankA === 0 || rankA === null || rankA === undefined) ? Infinity : rankA;
        const normalizedRankB = (rankB === 0 || rankB === null || rankB === undefined) ? Infinity : rankB;
        return normalizedRankA - normalizedRankB;
    }));

    function getRankBadge(rank) {
        if (rank === 0 || rank === null || rank === undefined) return '未開始';
        if (rank === 1) return '🥇 1位';
        if (rank === 2) return '🥈 2位';
        if (rank === 3) return '🥉 3位';
        return `${rank}位`;
    }
</script>

<svelte:head>
    <title>{eventData ? eventData.name + ' - ' : ''}アーカイブ - SportEase</title>
</svelte:head>

<div class="px-4 py-6 sm:px-0 max-w-7xl mx-auto">
    <div class="mb-6 flex items-center justify-between">
        <div>
            <a href="/dashboard/archive" class="text-sm text-indigo-600 hover:text-indigo-800 flex items-center mb-2">
                <svg class="w-4 h-4 mr-1" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 19l-7-7 7-7"></path>
                </svg>
                アーカイブ一覧に戻る
            </a>
            <h1 class="text-2xl font-bold text-gray-900">
                {eventData ? eventData.name : '読み込み中...'}
            </h1>
        </div>
    </div>

    {#if loading}
        <div class="flex justify-center p-12">
            <div class="animate-spin rounded-full h-10 w-10 border-b-2 border-indigo-600"></div>
        </div>
    {:else if error}
        <div class="bg-red-50 border border-red-200 text-red-700 px-4 py-3 rounded relative">
            <span class="block sm:inline">{error}</span>
        </div>
    {:else}
        <!-- Tabs -->
        <div class="border-b border-gray-200 mb-6">
            <nav class="-mb-px flex space-x-8" aria-label="Tabs">
                <button
                    onclick={() => activeTab = 'scores'}
                    class="{activeTab === 'scores' ? 'border-indigo-500 text-indigo-600' : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300'} whitespace-nowrap py-4 px-1 border-b-2 font-medium text-sm transition-colors duration-200"
                >
                    スコア・順位
                </button>
                <button
                    onclick={() => activeTab = 'tournaments'}
                    class="{activeTab === 'tournaments' ? 'border-indigo-500 text-indigo-600' : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300'} whitespace-nowrap py-4 px-1 border-b-2 font-medium text-sm transition-colors duration-200"
                >
                    試合結果
                </button>
                <button
                    onclick={() => activeTab = 'relays'}
                    class="{activeTab === 'relays' ? 'border-indigo-500 text-indigo-600' : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300'} whitespace-nowrap py-4 px-1 border-b-2 font-medium text-sm transition-colors duration-200"
                >
                    リレー結果
                </button>
            </nav>
        </div>

        <!-- Content -->
        {#if activeTab === 'scores'}
            {#if scores.length === 0}
                <div class="bg-gray-50 rounded-lg p-8 text-center text-gray-500 border border-gray-200">
                    スコアデータが見つかりませんでした。
                </div>
            {:else}
                <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-6">
                    {#each sortedScores as score (score.id || score.class_id)}
                        {@const rank = season === 'spring' ? score.rank_current_event : score.rank_overall}
                        {@const totalPoints = season === 'spring' ? score.total_points_current_event : score.total_points_overall}
                        <div class="bg-white rounded-lg shadow border border-gray-200 p-6 flex flex-col hover:shadow-md transition-shadow">
                            <div class="text-2xl font-bold text-center mb-2">{getRankBadge(rank)}</div>
                            <div class="text-xl font-semibold text-center text-gray-800 mb-4">{score.class_name}</div>
                            
                            <div class="mt-auto pt-4 border-t border-gray-100 flex justify-between items-center">
                                <span class="text-gray-500 text-sm">合計スコア</span>
                                <span class="text-2xl font-bold text-indigo-600">{totalPoints}</span>
                            </div>
                        </div>
                    {/each}
                </div>
            {/if}
        {:else if activeTab === 'tournaments'}
            {#if tournaments.length === 0}
                <div class="bg-gray-50 rounded-lg p-8 text-center text-gray-500 border border-gray-200">
                    試合結果データが見つかりませんでした。
                </div>
            {:else}
                <div class="space-y-6">
                    {#each tournaments as tournament (tournament.id)}
                        <div class="bg-white rounded-lg shadow-sm border border-gray-200 overflow-hidden">
                            <div class="px-6 py-4 bg-gray-50 border-b border-gray-200">
                                <h3 class="text-lg font-bold text-gray-900">{tournament.name}</h3>
                            </div>
                            <div class="p-6">
                                <p class="text-sm text-gray-600 mb-4">
                                    アーカイブでの簡易トーナメント表示<br>
                                    ※詳細な対戦ツリーはダッシュボードの本機能から閲覧してください（本アーカイブは過去の履歴です）。
                                </p>
                                <div class="overflow-x-auto">
                                    <table class="min-w-full divide-y divide-gray-200">
                                        <thead class="bg-gray-50">
                                            <tr>
                                                <th class="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">試合ID</th>
                                                <th class="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">ラウンド</th>
                                                <th class="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">チーム1</th>
                                                <th class="px-4 py-3 text-center text-xs font-medium text-gray-500 uppercase tracking-wider">スコア</th>
                                                <th class="px-4 py-3 text-right text-xs font-medium text-gray-500 uppercase tracking-wider">チーム2</th>
                                                <th class="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">勝者</th>
                                            </tr>
                                        </thead>
                                        <tbody class="bg-white divide-y divide-gray-200">
                                            {#each tournament.data?.matches || [] as match (match.id)}
                                                {@const team1 = match.sides?.[0]}
                                                {@const team2 = match.sides?.[1]}
                                                <tr class="hover:bg-gray-50">
                                                    <td class="px-4 py-3 whitespace-nowrap text-sm text-gray-500">#{match.id}</td>
                                                    <td class="px-4 py-3 whitespace-nowrap text-sm font-medium text-gray-900">
                                                        {match.isBronzeMatch ? '3位決定戦' : (match.roundIndex + 1) + '回戦'}
                                                    </td>
                                                    <td class="px-4 py-3 whitespace-nowrap text-sm text-gray-900 {team1?.isWinner ? 'font-bold text-indigo-600' : ''}">
                                                        {getSideName(tournament, team1)}
                                                    </td>
                                                    <td class="px-4 py-3 whitespace-nowrap text-sm text-center font-medium">
                                                        {#if isMatchCompleted(match)}
                                                            {getSideScore(team1) ?? '-'} - {getSideScore(team2) ?? '-'}
                                                        {:else}
                                                            -
                                                        {/if}
                                                    </td>
                                                    <td class="px-4 py-3 whitespace-nowrap text-sm text-right text-gray-900 {team2?.isWinner ? 'font-bold text-indigo-600' : ''}">
                                                        {getSideName(tournament, team2)}
                                                    </td>
                                                    <td class="px-4 py-3 whitespace-nowrap text-sm text-gray-500">
                                                        {#if isMatchCompleted(match)}
                                                            {getWinnerName(tournament, match)}
                                                        {:else}
                                                            未完了
                                                        {/if}
                                                    </td>
                                                </tr>
                                            {/each}
                                        </tbody>
                                    </table>
                                </div>
                            </div>
                        </div>
                    {/each}
                </div>
            {/if}
        {:else if activeTab === 'relays'}
            {#if relayError}
                <p class="rounded-md border border-red-200 bg-red-50 px-4 py-3 text-sm text-red-700">
                    {relayError}
                </p>
            {:else if relayMatches.length === 0}
                <p class="rounded-md border border-gray-200 bg-gray-50 px-4 py-3 text-sm text-gray-600">
                    リレー結果データが見つかりませんでした。
                </p>
            {:else}
                <div class="grid gap-4 lg:grid-cols-2">
                    {#each relayMatches as match (match.id)}
                        <article class="rounded-lg border border-indigo-100 bg-white p-5 shadow-sm">
                            <div class="flex flex-wrap items-start justify-between gap-3">
                                <div>
                                    <h3 class="text-lg font-semibold text-gray-900">{match.title || 'リレー'}</h3>
                                    <p class="mt-1 text-sm text-gray-600">ステータス: {formatStatus(match.status)}</p>
                                </div>
                                <span class="rounded-full bg-indigo-50 px-3 py-1 text-sm font-semibold text-indigo-700">
                                    {formatDateTime(match.scheduled_at)}
                                </span>
                            </div>

                            <dl class="mt-4 grid gap-3 text-sm sm:grid-cols-2">
                                <div>
                                    <dt class="font-semibold text-gray-700">場所</dt>
                                    <dd class="mt-1 text-gray-600">{match.location || '未定'}</dd>
                                </div>
                                <div>
                                    <dt class="font-semibold text-gray-700">形式</dt>
                                    <dd class="mt-1 text-gray-600">{match.format || '未定'}</dd>
                                </div>
                            </dl>

                            {#if match.memo}
                                <div class="mt-4 rounded-md bg-gray-50 px-3 py-2 text-sm text-gray-700">
                                    {match.memo}
                                </div>
                            {/if}

                            <div class="mt-4">
                                <h4 class="text-sm font-semibold text-gray-700">参加予定</h4>
                                {#if match.entries && match.entries.length > 0}
                                    <div class="mt-2 flex flex-wrap gap-2">
                                        {#each match.entries as entry (entry.id)}
                                            <span class="rounded-full border border-gray-200 bg-gray-50 px-3 py-1 text-sm text-gray-700">
                                                {entryName(entry)}
                                            </span>
                                        {/each}
                                    </div>
                                {:else}
                                    <p class="mt-2 text-sm text-gray-500">参加予定はまだ登録されていません。</p>
                                {/if}
                            </div>

                            <div class="mt-4">
                                <h4 class="text-sm font-semibold text-gray-700">結果</h4>
                                {#if match.result?.details?.length}
                                    <ol class="mt-2 space-y-2">
                                        {#each resultEntries(match) as detail (detail.id)}
                                            <li class="flex items-center justify-between rounded-md border border-gray-100 bg-gray-50 px-3 py-2 text-sm">
                                                <span class="font-medium text-gray-900">
                                                    {detail.rank ? `${detail.rank}位` : '順位未設定'} {resultEntryName(detail, match)}
                                                </span>
                                                <span class="text-indigo-700 font-semibold">{detail.points} 点</span>
                                            </li>
                                        {/each}
                                    </ol>
                                {:else if match.winner_display}
                                    <p class="mt-2 text-sm text-gray-600">勝者: {match.winner_display}</p>
                                {:else}
                                    <p class="mt-2 text-sm text-gray-500">結果はまだ登録されていません。</p>
                                {/if}
                            </div>
                        </article>
                    {/each}
                </div>
            {/if}
        {/if}
    {/if}
</div>
