<script>
    import { onMount } from 'svelte';
    import { browser } from '$app/environment';
    import { activeEvent } from '$lib/stores/eventStore.js';
    import { get } from 'svelte/store';

    let allTournaments = $state([]);
    let isLoading = $state(false);
    let isRainyMode = $state(false);
    let boardGameRuns = $state([]);

    onMount(async () => {
        await activeEvent.init();
        const currentEvent = get(activeEvent);
        if (currentEvent) {
            isRainyMode = currentEvent.is_rainy_mode || false;
            await fetchTournamentsForActiveEvent();
        }
        
        // ページがフォーカスされた時にトーナメントデータを再取得
        const handleFocus = async () => {
            const currentEvent = get(activeEvent);
            if (currentEvent) {
                // イベント情報を再取得して雨天時モードの状態を確認
                try {
                    const res = await fetch('/api/root/events');
                    if (res.ok) {
                        const events = await res.json();
                        const active = events.find(e => e.id === currentEvent.id);
                        if (active) {
                            const newIsRainyMode = active.is_rainy_mode || false;
                            if (newIsRainyMode !== isRainyMode || newIsRainyMode) {
                                // 雨天時モードが変更された場合、または雨天時モードが有効な場合は再取得
                                isRainyMode = newIsRainyMode;
                                activeEvent._set(active);
                                await fetchTournamentsForActiveEvent();
                            }
                        }
                    }
                } catch (error) {
                    console.error('Error checking rainy mode:', error);
                }
            }
        };
        
        if (browser) {
            window.addEventListener('focus', handleFocus);
        }
        
        return () => {
            if (browser) {
                window.removeEventListener('focus', handleFocus);
            }
        };
    });

    async function fetchTournamentsForActiveEvent() {
        const currentEvent = get(activeEvent);
        if (!currentEvent) return;

        isLoading = true;
        try {
            const [response, boardGameResponse] = await Promise.all([
                fetch(`/api/student/events/${currentEvent.id}/tournaments`),
                fetch(`/api/student/events/${currentEvent.id}/board-game-runs`),
            ]);
            if (response.ok) {
                const fetchedTournaments = await response.json();

                allTournaments = fetchedTournaments.map(t => {
                    if (typeof t.data === 'string') {
                        try {
                            t.data = JSON.parse(t.data);
                        } catch (e) {
                            console.error('Failed to parse tournament data:', e);
                            t.data = null;
                        }
                    }
                    return t;
                });

                renderAllBrackets();
            } else {
                console.error('Failed to fetch tournaments');
                allTournaments = [];
            }
            boardGameRuns = boardGameResponse.ok ? await boardGameResponse.json() : [];
        } catch (error) {
            console.error('Error fetching tournaments:', error);
            allTournaments = [];
        } finally {
            isLoading = false;
        }
    }

    async function renderAllBrackets() {
        if (!browser) return;
        setTimeout(async () => {
            for (const tournament of allTournaments) {
                await renderBracket(tournament);
            }
        }, 0);
    }

    async function renderBracket(tournament) {
        if (!browser) return;
        const wrapper = document.getElementById(`bracket-${tournament.id}`);
        if (wrapper && tournament.data) {
            wrapper.innerHTML = '';
            try {
                const { createBracket } = await import('bracketry');
                createBracket(tournament.data, wrapper);
            } catch (error) {
                console.error('Failed to load createBracket:', error);
                wrapper.innerHTML = '<p>ブラケットの読み込みに失敗しました。</p>';
            }
        }
    }

    function boardGameForTournament(tournamentID) {
        for (const run of boardGameRuns) {
            const tournament = run.tournaments?.find((item) => item.id === tournamentID);
            if (tournament) return { run, tournament };
        }
        return null;
    }

    function aggregateBoardGameScores(run) {
		const byClass = {};
        for (const tournament of run.tournaments || []) {
            for (const ranking of tournament.rankings || []) {
				const current = byClass[ranking.class_id] || { classID: ranking.class_id, className: ranking.class_name, points: 0 };
                current.points += ranking.total_points;
				byClass[ranking.class_id] = current;
            }
        }
		return Object.values(byClass).sort((a, b) => b.points - a.points || a.className.localeCompare(b.className, 'ja'));
    }

    function boardGameMemberName(member) {
        return member.display_name || member.email;
    }
</script>

<div class="space-y-8 p-4 md:p-8">
    <h1 class="text-2xl md:text-3xl font-bold text-gray-800 border-b pb-2">トーナメント一覧</h1>

    {#if isLoading}
        <div class="flex justify-center items-center py-8">
            <p class="text-gray-600">読み込み中...</p>
        </div>
    {:else if allTournaments && allTournaments.length > 0}
        <div class="space-y-8">
            {#each allTournaments as tournament (tournament.id)}
                {@const isLoserBracket = tournament.name.includes('敗者戦')}
                {@const boardGame = boardGameForTournament(tournament.id)}
                {#if !isLoserBracket || isRainyMode}
                    <div class="p-4 border rounded-lg bg-white shadow-sm">
                        <div class="mb-4">
                            <h3 class="text-lg font-bold text-gray-800">{tournament.name}</h3>
                        </div>
                        {#if boardGame}
                            <section class="mb-5 rounded-lg border border-amber-200 bg-amber-50 p-4 text-sm">
                                <div class="flex flex-wrap justify-between gap-3">
                                    <div>
                                        <p><span class="font-semibold">会場:</span> {boardGame.run.location}</p>
                                        {#if boardGame.run.scheduled_date}<p><span class="font-semibold">実施日:</span> {boardGame.run.scheduled_date}</p>{/if}
                                        <p><span class="font-semibold">試合時間:</span> 通常 {boardGame.run.regular_minutes}分／決勝 {boardGame.run.final_minutes}分</p>
                                        <p><span class="font-semibold">得点:</span> 1勝 {boardGame.run.win_points}点</p>
                                    </div>
                                    {#if boardGame.run.rules_pdf_url}
                                        <a class="h-fit font-medium text-blue-700 underline" href={boardGame.run.rules_pdf_url} target="_blank" rel="noreferrer">ルールPDFを開く</a>
                                    {/if}
                                </div>
                                {#if boardGame.run.description}<p class="mt-2 whitespace-pre-wrap text-gray-700">{boardGame.run.description}</p>{/if}

                                <div class="mt-3">
                                    <h4 class="font-semibold">出場者</h4>
                                    <ul class="mt-1 grid gap-1 sm:grid-cols-2 lg:grid-cols-3">
                                        {#each boardGame.tournament.entries as entry (entry.id)}
                                            <li class="rounded bg-white px-2 py-1">
                                                {entry.team_name}: {entry.members.filter((member) => !member.is_substitute).map(boardGameMemberName).join('、')}
                                                {#if entry.members.some((member) => member.is_substitute)}<span class="text-gray-500">（補欠: {entry.members.filter((member) => member.is_substitute).map(boardGameMemberName).join('、')}）</span>{/if}
                                            </li>
                                        {/each}
                                    </ul>
                                </div>

                                {#if boardGame.tournament.rankings?.length}
                                    <div class="mt-4 overflow-x-auto">
                                        <h4 class="mb-1 font-semibold">確定順位・獲得点</h4>
                                        <table class="min-w-full divide-y divide-amber-200 bg-white text-left">
                                            <thead><tr><th class="px-2 py-1">順位</th><th class="px-2 py-1">出場枠</th><th class="px-2 py-1">勝数</th><th class="px-2 py-1">勝利点</th><th class="px-2 py-1">順位点</th><th class="px-2 py-1">合計</th></tr></thead>
                                            <tbody class="divide-y divide-gray-100">
                                                {#each boardGame.tournament.rankings as ranking (ranking.id)}
                                                    <tr><td class="px-2 py-1">{ranking.rank}位</td><td class="px-2 py-1">{ranking.team_name}</td><td class="px-2 py-1">{ranking.win_count}</td><td class="px-2 py-1">{ranking.win_points}</td><td class="px-2 py-1">{ranking.rank_points}</td><td class="px-2 py-1 font-semibold">{ranking.total_points}</td></tr>
                                                {/each}
                                            </tbody>
                                        </table>
                                    </div>
                                {/if}

                                {#if boardGame.run.tournaments?.[0]?.id === tournament.id && aggregateBoardGameScores(boardGame.run).length}
                                    <div class="mt-4">
                                        <h4 class="font-semibold">クラス別合計{boardGame.run.game_type === 'shogi' ? '（A・B合算）' : ''}</h4>
                                        <ol class="mt-1 flex flex-wrap gap-2">
                                            {#each aggregateBoardGameScores(boardGame.run) as score, index (score.classID)}
                                                <li class="rounded bg-white px-3 py-1">{index + 1}位 {score.className}: <strong>{score.points}点</strong></li>
                                            {/each}
                                        </ol>
                                    </div>
                                {/if}
                            </section>
                        {/if}
                        <div id="bracket-{tournament.id}"></div>
                    </div>
                {/if}
            {/each}
        </div>
    {:else}
        <div class="bg-blue-100 border-l-4 border-blue-500 text-blue-700 p-4" role="alert">
            <p class="font-bold">情報</p>
            <p>表示するトーナメントがありません。</p>
        </div>
    {/if}
</div>
