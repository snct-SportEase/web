<script>
    import { onMount } from 'svelte';
    import { browser } from '$app/environment';
    import { activeEvent } from '$lib/stores/eventStore.js';
    import { get } from 'svelte/store';

    let allTournaments = [];
    let isLoading = false;

    onMount(async () => {
        await activeEvent.init();
        const currentEvent = get(activeEvent);
        if (currentEvent) {
            fetchTournamentsForActiveEvent();
        }
    });

    async function fetchTournamentsForActiveEvent() {
        const currentEvent = get(activeEvent);
        if (!currentEvent) return;

        isLoading = true;
        try {
            const response = await fetch(`/api/student/events/${currentEvent.id}/tournaments`);
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
                <div class="p-4 border rounded-lg bg-white shadow-sm">
                    <div class="mb-4">
                        <h3 class="text-lg font-bold text-gray-800">{tournament.name}</h3>
                    </div>
                    <div id="bracket-{tournament.id}"></div>
                </div>
            {/each}
        </div>
    {:else}
        <div class="bg-blue-100 border-l-4 border-blue-500 text-blue-700 p-4" role="alert">
            <p class="font-bold">情報</p>
            <p>表示するトーナメントがありません。</p>
        </div>
    {/if}
</div>

