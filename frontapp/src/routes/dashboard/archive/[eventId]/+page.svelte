<script>
    import { onMount } from 'svelte';
    import { page } from '$app/stores';
    
    let eventId = $page.params.eventId;
    let activeTab = 'scores';
    
    let eventData = null;
    let scores = [];
    let tournaments = [];
    let loading = true;
    let error = null;

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
                const resData = await tournRes.json();
                tournaments = resData.tournaments || [];
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
                    トーナメント結果
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
                    {#each sortedScores as score}
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
                    トーナメントデータが見つかりませんでした。
                </div>
            {:else}
                <div class="space-y-6">
                    {#each tournaments as tournament}
                        <div class="bg-white rounded-lg shadow-sm border border-gray-200 overflow-hidden">
                            <div class="px-6 py-4 bg-gray-50 border-b border-gray-200 flex justify-between items-center">
                                <h3 class="text-lg font-bold text-gray-900">{tournament.name}</h3>
                                <span class="px-3 py-1 rounded-full text-xs font-semibold bg-indigo-100 text-indigo-800">
                                    {tournament.sport_name}
                                </span>
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
                                            {#each tournament.matches as match}
                                                <tr class="hover:bg-gray-50">
                                                    <td class="px-4 py-3 whitespace-nowrap text-sm text-gray-500">#{match.match_id}</td>
                                                    <td class="px-4 py-3 whitespace-nowrap text-sm font-medium text-gray-900">
                                                        {match.is_bronze_match ? '3位決定戦' : (match.round + 1) + '回戦'}
                                                    </td>
                                                    <td class="px-4 py-3 whitespace-nowrap text-sm text-gray-900 {match.winner_team_id === match.team1_id ? 'font-bold text-indigo-600' : ''}">
                                                        {match.team1_name || '未定'}
                                                    </td>
                                                    <td class="px-4 py-3 whitespace-nowrap text-sm text-center font-medium">
                                                        {#if match.status === 'completed'}
                                                            {match.team1_score} - {match.team2_score}
                                                        {:else}
                                                            -
                                                        {/if}
                                                    </td>
                                                    <td class="px-4 py-3 whitespace-nowrap text-sm text-right text-gray-900 {match.winner_team_id === match.team2_id ? 'font-bold text-indigo-600' : ''}">
                                                        {match.team2_name || '未定'}
                                                    </td>
                                                    <td class="px-4 py-3 whitespace-nowrap text-sm text-gray-500">
                                                        {#if match.status === 'completed'}
                                                            {#if match.winner_team_id === match.team1_id}
                                                                {match.team1_name}
                                                            {:else if match.winner_team_id === match.team2_id}
                                                                {match.team2_name}
                                                            {:else}
                                                                引き分け
                                                            {/if}
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
        {/if}
    {/if}
</div>
