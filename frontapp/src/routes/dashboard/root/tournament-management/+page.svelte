<script>
    import { onMount } from 'svelte';
    import { browser } from '$app/environment';
    import { activeEvent } from '$lib/stores/eventStore.js';
    import { get } from 'svelte/store';
    import { dndzone } from 'svelte-dnd-action';
    
    let createBracket = null;

    let isGenerating = false;
    let isSaving = false;
    let allTournaments = [];
    let generatedTournamentsPreview = null;

    let editingTournamentId = null;
    let teamsForEditing = [];
    const flipDurationMs = 300;

    onMount(async () => {
        if (browser) {
            const bracketryModule = await import('bracketry');
            createBracket = bracketryModule.createBracket;
        }
        await activeEvent.init();
        const currentEvent = get(activeEvent);
        if (currentEvent) {
            fetchTournamentsForActiveEvent();
        }
    });

    async function previewAllTournaments() {
        if (!browser) return;
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
                allTournaments = previewData.map((t, index) => ({
                    id: `preview-${index}`,
                    name: t.sport_name,
                    data: t.tournament_data,
                }));
                alert('トーナメントのプレビューが生成されました。内容を確認して保存してください。');
                renderAllBrackets();
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
                generatedTournamentsPreview = null;
                await fetchTournamentsForActiveEvent();
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
        if (!browser) return;
        const currentEvent = get(activeEvent);
        if (!currentEvent) return;

        try {
            const response = await fetch(`/api/root/events/${currentEvent.id}/tournaments`);
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

                generatedTournamentsPreview = null; // Ensure preview is cleared
                renderAllBrackets();
            } else {
                console.error('Failed to fetch tournaments');
                allTournaments = [];
            }
        } catch (error) {
            console.error('Error fetching tournaments:', error);
            allTournaments = [];
        }
    }

    function renderAllBrackets() {
        if (!browser) return;
        setTimeout(() => {
            allTournaments.forEach(tournament => {
                renderBracket(tournament);
            });
        }, 0);
    }

    function renderBracket(tournament) {
        if (!browser || !createBracket) return;
        const wrapper = document.getElementById(`bracket-${tournament.id}`);
        if (wrapper && tournament.data) {
            wrapper.innerHTML = '';
            createBracket(tournament.data, wrapper);
        }
    }

    function getOrderedContestantIds(bracketData) {
        if (!bracketData || !bracketData.matches) {
            return [];
        }
        const contestantIds = [];
        const firstRoundMatches = bracketData.matches.filter(m => m.roundIndex === 0).sort((a, b) => a.order - b.order);

        for (const match of firstRoundMatches) {
            if (match.sides) {
                for (const side of match.sides) {
                    if (side.contestantId) {
                        contestantIds.push(side.contestantId);
                    } else {
                        contestantIds.push(null); 
                    }
                }
            }
        }
        return contestantIds;
    }

    function getTeams(tournament) {
        const bracketData = tournament.data;
        if (!bracketData || !bracketData.contestants) {
            return [];
        }
        const orderedContestantIds = getOrderedContestantIds(bracketData);
        const teams = orderedContestantIds.map(contestantId => {
            if (!contestantId) return { name: 'BYE' };
            const contestant = bracketData.contestants[contestantId];
            if (contestant && contestant.players && contestant.players.length > 0 && contestant.players[0].title) {
                return { name: contestant.players[0].title };
            }
            return { name: 'Unknown Team' };
        });

        return teams.map((team, index) => ({ ...team, id: `${tournament.id}-team-${index}` }));
    }

    function updateBracketDataWithNewTeams(bracketData, newTeamOrder) {
        const newBracketData = JSON.parse(JSON.stringify(bracketData));
        const orderedContestantIds = getOrderedContestantIds(newBracketData);

        const teamDataMap = {};
        Object.values(newBracketData.contestants).forEach(c => {
            if (c.players && c.players.length > 0 && c.players[0].title) {
                teamDataMap[c.players[0].title] = c.players;
            }
        });

        orderedContestantIds.forEach((contestantId, index) => {
            if (!contestantId) return;

            const contestant = newBracketData.contestants[contestantId];
            if (contestant) {
                const newTeamName = newTeamOrder[index];
                if (newTeamName && newTeamName !== 'BYE') {
                    const newPlayers = teamDataMap[newTeamName];
                    if (newPlayers) {
                        contestant.players = newPlayers;
                    }
                }
            }
        });

        return newBracketData;
    }

    function toggleEdit(tournament) {
        if (editingTournamentId === tournament.id) {
            editingTournamentId = null;
            teamsForEditing = [];
        } else {
            editingTournamentId = tournament.id;
            teamsForEditing = getTeams(tournament);
        }
    }

    function handleDnd(e) {
        teamsForEditing = e.detail.items;
    }

    function saveTeamOrder(tournament) {
        if (!browser) return;
        const newTeamNames = teamsForEditing.map(t => t.name);
        const newBracketData = updateBracketDataWithNewTeams(tournament.data, newTeamNames);

        const tournamentIndex = allTournaments.findIndex(t => t.id === tournament.id);
        if (tournamentIndex !== -1) {
            allTournaments[tournamentIndex].data = newBracketData;
            allTournaments = [...allTournaments];

            if (generatedTournamentsPreview) {
                const previewIndex = parseInt(tournament.id.split('-')[1]);
                const currentPreview = generatedTournamentsPreview[previewIndex];

                if (currentPreview) {
                    const originalShuffledTeams = currentPreview.shuffled_teams || [];
                    const teamObjectMap = new Map(originalShuffledTeams.map(team => [team.name, team]));
                    const newShuffledTeams = newTeamNames.map(name => teamObjectMap.get(name)).filter(Boolean);

                    const newPreviewArray = [...generatedTournamentsPreview];
                    newPreviewArray[previewIndex] = {
                        ...currentPreview,
                        tournament_data: newBracketData,
                        shuffled_teams: newShuffledTeams
                    };
                    generatedTournamentsPreview = newPreviewArray;
                }
            }
            
            renderBracket(allTournaments[tournamentIndex]);
        }

        editingTournamentId = null;
        teamsForEditing = [];
    }

</script>

<style>
    .draggable-list li {
        padding: 8px;
        margin-bottom: 4px;
        border: 1px solid #ccc;
        border-radius: 4px;
        cursor: grab;
        background-color: #f9f9f9;
    }
    .draggable-list li:active {
        cursor: grabbing;
        background-color: #e9e9e9;
    }
</style>

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
                    <div class="flex justify-between items-center mb-2">
                        <h3 class="text-lg font-bold">{tournament.name}</h3>
                        {#if generatedTournamentsPreview}
                        <button on:click={() => toggleEdit(tournament)} class="py-1 px-3 border border-transparent shadow-sm text-sm font-medium rounded-md text-white {editingTournamentId === tournament.id ? 'bg-red-600 hover:bg-red-700' : 'bg-indigo-600 hover:bg-indigo-700'}">
                            {editingTournamentId === tournament.id ? 'キャンセル' : 'シード順を編集'}
                        </button>
                        {/if}
                    </div>

                    {#if editingTournamentId === tournament.id}
                        <div class="my-4 p-4 border rounded-lg bg-gray-50">
                            <h4 class="font-semibold mb-2">チームの並び替え (ドラッグ＆ドロップで編集)</h4>
                            <ul class="draggable-list" use:dndzone={{ items: teamsForEditing, flipDurationMs }} on:consider={handleDnd} on:finalize={handleDnd}>
                                {#each teamsForEditing as team (team.id)}
                                    <li class="bg-white">{team.name}</li>
                                {/each}
                            </ul>
                            <button on:click={() => saveTeamOrder(tournament)} class="mt-4 w-full inline-flex justify-center py-2 px-4 border border-transparent shadow-sm text-sm font-medium rounded-md text-white bg-blue-600 hover:bg-blue-700">
                                この順序でブラケットを更新
                            </button>
                        </div>
                    {/if}

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