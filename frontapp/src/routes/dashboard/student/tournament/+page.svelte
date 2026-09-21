<script>
    import { onMount } from 'svelte';
    import { browser } from '$app/environment';
    import { activeEvent } from '$lib/stores/eventStore.js';
    import { get } from 'svelte/store';

    let allTournaments = $state([]);
    let isLoading = $state(false);
    let isRainyMode = $state(false);

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
                    const active = await activeEvent.init();
                    if (active) {
                        const newIsRainyMode = active.is_rainy_mode || false;
                        if (newIsRainyMode !== isRainyMode || newIsRainyMode) {
                            // 雨天時モードが変更された場合、または雨天時モードが有効な場合は再取得
                            isRainyMode = newIsRainyMode;
                            await fetchTournamentsForActiveEvent();
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

    function bracket(node, initialData) {
        let bracketInstance = null;
        let renderVersion = 0;

        async function render(data) {
            const currentVersion = ++renderVersion;
            bracketInstance?.uninstall?.();
            bracketInstance = null;
            node.replaceChildren();

            if (!data) {
                node.textContent = 'このトーナメント情報はありません。';
                return;
            }

            try {
                const { createBracket } = await import('bracketry');
                if (currentVersion !== renderVersion) return;
                bracketInstance = createBracket(data, node);
            } catch (error) {
                console.error('Failed to load createBracket:', error);
                if (currentVersion === renderVersion) {
                    node.textContent = 'ブラケットの読み込みに失敗しました。';
                }
            }
        }

        void render(initialData);

        return {
            update(data) {
                void render(data);
            },
            destroy() {
                renderVersion += 1;
                bracketInstance?.uninstall?.();
            },
        };
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
                {#if !isLoserBracket || isRainyMode}
                    <div class="p-4 border rounded-lg bg-white shadow-sm">
                        <div class="mb-4">
                            <h3 class="text-lg font-bold text-gray-800">{tournament.name}</h3>
                        </div>
                        <div id="bracket-{tournament.id}" use:bracket={tournament.data}></div>
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
