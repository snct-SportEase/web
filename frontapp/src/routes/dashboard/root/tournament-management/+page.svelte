<script>
    import { onMount } from 'svelte';
    import { createBracket } from 'bracketry';
    import { activeEvent } from '$lib/stores/eventStore.js';
    import { get } from 'svelte/store';

    let sports = [];
    let teams = [];
    let selectedSportId = null;
    let isGenerating = false;
    let isSaving = false;
    let allTournaments = [];
    let generatedTournamentsPreview = null;

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

    async function previewAllTournaments() {
        const currentEvent = get(activeEvent);
        if (!currentEvent) {
            alert('アクティブな大会が設定されていません。');
            return;
        }

        if (!confirm(`「${currentEvent.name}」の全競技のトーナメントプレビューを生成します。よろしいですか？`)) {
            return;
        }

        isGenerating = true;
        generatedTournamentsPreview = null;
        try {
            const response = await fetch(`/api/root/events/${currentEvent.id}/tournaments/generate-preview`, {
                method: 'POST',
            });

            if (response.ok) {
                const previewData = await response.json();
                generatedTournamentsPreview = previewData;
                // Adapt preview data to the format expected by the UI
                allTournaments = previewData.map((t, index) => ({
                    id: `preview-${index}`,
                    name: t.sport_name,
                    data: t.tournament_data,
                }));
                alert('トーナメントのプレビューが生成されました。内容を確認して保存してください。');
                // Render brackets for preview
                setTimeout(() => {
                    allTournaments.forEach(tournament => {
                        const wrapper = document.getElementById(`bracket-${tournament.id}`);
                        if (wrapper && tournament.data) {
                            wrapper.innerHTML = ''; // Clear previous bracket
                            createBracket(tournament.data, wrapper);
                        }
                    });
                }, 0);
            } else {
                const result = await response.json();
                alert(`エラー: ${result.error || '不明なエラー'}`);
            }
        } catch (error) {
            console.error('Error generating tournament preview:', error);
            alert('トーナメントプレビューの生成中にエラーが発生しました。');
        } finally {
            isGenerating = false;
        }
    }

    async function saveAllTournaments() {
        const currentEvent = get(activeEvent);
        if (!currentEvent || !generatedTournamentsPreview) {
            alert('保存するトーナメントデータがありません。');
            return;
        }

        if (!confirm('現在のプレビューをデータベースに保存します。よろしいですか？既存のトーナメントは上書きされます。')) {
            return;
        }

        isSaving = true;
        try {
            const response = await fetch(`/api/root/events/${currentEvent.id}/tournaments/bulk-create`, {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify(generatedTournamentsPreview),
            });

            const result = await response.json();
            if (response.ok) {
                alert(result.message || 'トーナメントが正常に保存されました。');
                generatedTournamentsPreview = null; // Clear preview data
                await fetchTournamentsForActiveEvent(); // Refresh with data from DB
            } else {
                alert(`エラー: ${result.error || '不明なエラー'}`);
            }
        } catch (error) {
            console.error('Error saving all tournaments:', error);
            alert('トーナメントの保存中にエラーが発生しました。');
        } finally {
            isSaving = false;
        }
    }

    async function fetchTournamentsForActiveEvent() {
        const currentEvent = get(activeEvent);
        if (!currentEvent) return;

        try {
            const response = await fetch(`/api/root/events/${currentEvent.id}/tournaments`);
            if (response.ok) {
                allTournaments = await response.json();
                generatedTournamentsPreview = null; // Ensure preview is cleared
                // Use timeout to ensure DOM is updated before creating brackets
                setTimeout(() => {
                    allTournaments.forEach(tournament => {
                        const wrapper = document.getElementById(`bracket-${tournament.id}`);
                        if (wrapper && tournament.data) {
                            wrapper.innerHTML = ''; // Clear previous bracket
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

<div>
    <div>
        <div>
            <h2 class="text-xl font-semibold mb-3">一括トーナメント生成</h2>
            <div class="space-y-4 p-4 border rounded-lg">
                <p class="text-sm text-gray-600">現在アクティブな大会に登録されている全ての競技のトーナメントをプレビューし、保存します。</p>
                {#if $activeEvent}
                    <p class="text-sm">アクティブな大会: <span class="font-bold">{$activeEvent.name}</span></p>
                {/if}
                <button on:click={previewAllTournaments} class="w-full inline-flex justify-center py-2 px-4 border border-transparent shadow-sm text-sm font-medium rounded-md text-white bg-blue-600 hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500" disabled={!$activeEvent || isGenerating}>
                    {isGenerating ? 'プレビュー生成中...' : 'トーナメントプレビューを生成'}
                </button>
                {#if generatedTournamentsPreview}
                    <button on:click={saveAllTournaments} class="w-full inline-flex justify-center py-2 px-4 border border-transparent shadow-sm text-sm font-medium rounded-md text-white bg-green-600 hover:bg-green-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-green-500" disabled={isSaving}>
                        {isSaving ? '保存中...' : 'プレビューをDBに保存'}
                    </button>
                {/if}
            </div>
        </div>
    </div>

    <div class="mt-16">
        {#if allTournaments && allTournaments.length > 0}
            <h2 class="text-xl font-semibold">{generatedTournamentsPreview ? 'プレビュー中のトーナメント' : '生成済みトーナメント一覧'}</h2>
            {#each allTournaments as tournament (tournament.id)}
                <div class="p-4 border rounded-lg mb-8">
                    <h3 class="text-lg font-bold mb-2">{tournament.name}</h3>
                    <div id="bracket-{tournament.id}"></div>
                </div>
            {/each}
        {:else}
            <div class="p-4 border rounded-lg text-center text-gray-500">
                <p>表示するトーナメントがありません。</p>
                <p>「トーナメントプレビューを生成」ボタンを押して、トーナメントを作成してください。</p>
            </div>
        {/if}
    </div>
</div>
