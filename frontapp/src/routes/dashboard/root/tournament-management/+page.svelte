<script>
    import { onMount } from 'svelte';
    import { createBracket } from 'bracketry';
    import { activeEvent } from '$lib/stores/eventStore.js';
    import { get } from 'svelte/store';

    let sports = [];
    let teams = [];
    let selectedSportId = null;
    let isGeneratingAll = false;
    let allTournaments = [];

    onMount(async () => {
        try {
            const response = await fetch(`/api/root/sports`);
            if (response.ok) {
                sports = await response.json();
            } else {
                console.error('Failed to fetch sports');
            }
        } catch (error) {
            console.error('Error fetching sports:', error);
        }
        await activeEvent.init();
        const currentEvent = get(activeEvent);
        if (currentEvent) {
            fetchTournamentsForActiveEvent();
        }
    });

    async function fetchTeams(sportId) {
        if (!sportId) {
            teams = [];
            return;
        }
        try {
            const response = await fetch(`/api/root/sports/${sportId}/teams`);
            if (response.ok) {
                teams = await response.json();
            } else {
                console.error('Failed to fetch teams');
                teams = [];
            }
        } catch (error) {
            console.error('Error fetching teams:', error);
            teams = [];
        }
    }

    async function generateTournament() {
        try {
            const response = await fetch('/api/root/tournaments/generate', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify(teams),
            });

            if (response.ok) {
                const tournamentData = await response.json();
                const wrapperElement = document.getElementById('viewer-individual');
                wrapperElement.innerHTML = ''; // Clear previous bracket
                createBracket(tournamentData, wrapperElement);
            } else {
                const error = await response.json();
                alert(`トーナメントの生成に失敗しました: ${error.error}`);
            }
        } catch (error) {
            console.error('Error generating tournament:', error);
            alert('トーナメントの生成中にエラーが発生しました。');
        }
    }

    async function generateAllTournaments() {
        const currentEvent = get(activeEvent);
        if (!currentEvent) {
            alert('アクティブな大会が設定されていません。');
            return;
        }

        if (!confirm(`「${currentEvent.name}」の全競技のトーナメントを生成します。よろしいですか？既存のトーナメントは削除されます。`)) {
            return;
        }

        isGeneratingAll = true;
        allTournaments = [];
        try {
            const response = await fetch(`/api/root/events/${currentEvent.id}/tournaments/generate-all`, {
                method: 'POST',
            });

            const result = await response.json();
            if (response.ok) {
                alert(result.message || '全トーナメントの生成が正常に完了しました。');
                await fetchTournamentsForActiveEvent();
            } else {
                alert(`エラー: ${result.error || '不明なエラー'}`);
            }
        } catch (error) {
            console.error('Error generating all tournaments:', error);
            alert('全トーナメントの生成中にエラーが発生しました。');
        } finally {
            isGeneratingAll = false;
        }
    }

    async function fetchTournamentsForActiveEvent() {
        const currentEvent = get(activeEvent);
        if (!currentEvent) return;

        try {
            const response = await fetch(`/api/root/events/${currentEvent.id}/tournaments`);
            if (response.ok) {
                allTournaments = await response.json();
                // Use timeout to ensure DOM is updated before creating brackets
                setTimeout(() => {
                    allTournaments.forEach(tournament => {
                        const wrapper = document.getElementById(`bracket-${tournament.id}`);
                        if (wrapper && tournament.data) {
                            createBracket(tournament.data, wrapper);
                        }
                    });
                }, 0);
            } else {
                console.error('Failed to fetch tournaments');
                allTournaments = [];
            }
        } catch (error) {
            console.error('Error fetching tournaments:', error);
            allTournaments = [];
        }
    }

    $: {
        if (selectedSportId) {
            fetchTeams(selectedSportId);
        }
    }
</script>

<h1 class="text-2xl font-bold mb-4">トーナメント生成・管理</h1>

<div class="grid grid-cols-1 md:grid-cols-3 gap-8">
    <div class="md:col-span-1 space-y-6">
        <div>
            <h2 class="text-xl font-semibold mb-3">一括トーナメント生成</h2>
            <div class="space-y-4 p-4 border rounded-lg">
                <p class="text-sm text-gray-600">現在アクティブな大会に登録されている全ての競技のトーナメントを一括で生成します。</p>
                {#if $activeEvent}
                    <p class="text-sm">アクティブな大会: <span class="font-bold">{$activeEvent.name}</span></p>
                {/if}
                <button on:click={generateAllTournaments} class="w-full inline-flex justify-center py-2 px-4 border border-transparent shadow-sm text-sm font-medium rounded-md text-white bg-red-600 hover:bg-red-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-red-500" disabled={!$activeEvent || isGeneratingAll}>
                    {isGeneratingAll ? '生成中...' : '全トーナメントを一括生成'}
                </button>
            </div>
        </div>
        <div>
            <h2 class="text-xl font-semibold mb-3">個別トーナメント表示</h2>
            <div class="space-y-4 p-4 border rounded-lg">
                <div>
                    <label for="sport-select" class="block text-sm font-medium text-gray-700">競技選択</label>
                    <select id="sport-select" bind:value={selectedSportId} class="mt-1 block w-full pl-3 pr-10 py-2 text-base border-gray-300 focus:outline-none focus:ring-indigo-500 focus:border-indigo-500 sm:text-sm rounded-md">
                        <option value={null}>-- 競技を選択 --</option>
                        {#each sports as sport}
                            <option value={sport.id}>{sport.name}</option>
                        {/each}
                    </select>
                </div>
                <button on:click={generateTournament} class="w-full inline-flex justify-center py-2 px-4 border border-transparent shadow-sm text-sm font-medium rounded-md text-white bg-indigo-600 hover:bg-indigo-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500" disabled={!selectedSportId || teams.length < 2}>
                    個別トーナメント表示
                </button>
                 <div class="pt-4">
                    <div id="viewer-individual"></div>
                </div>
            </div>
        </div>
    </div>

    <div class="md:col-span-2 space-y-6">
        {#if allTournaments.length > 0}
            <h2 class="text-xl font-semibold">生成済みトーナメント一覧</h2>
            {#each allTournaments as tournament (tournament.id)}
                <div class="p-4 border rounded-lg">
                    <h3 class="text-lg font-bold mb-2">{tournament.name}</h3>
                    <div id="bracket-{tournament.id}"></div>
                </div>
            {/each}
        {:else}
            <div class="p-4 border rounded-lg text-center text-gray-500">
                <p>表示するトーナメントがありません。</p>
                <p>「全トーナメントを一括生成」ボタンを押して、トーナメントを作成してください。</p>
            </div>
        {/if}
    </div>
</div>
