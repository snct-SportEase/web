<script>
    import { onMount } from 'svelte';
    import { createBracket } from 'bracketry'

    let sports = [];
    let teams = [];
    let selectedSportId = null;
    let matchDuration = 30; // Default match duration in minutes

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
                console.log('teams', teams);
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
                console.log('tournamentData', tournamentData);
                const wrapperElement = document.getElementById('viewer');
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

    $: {
        if (selectedSportId) {
            fetchTeams(selectedSportId);
        }
    }
</script>

<h1 class="text-2xl font-bold mb-4">トーナメント生成・管理</h1>

<div class="grid grid-cols-1 md:grid-cols-3 gap-4">
    <div class="col-span-1">
        <div class="mb-4">
            <label for="sport-select" class="block text-sm font-medium text-gray-700">競技選択</label>
            <select id="sport-select" bind:value={selectedSportId} class="mt-1 block w-full pl-3 pr-10 py-2 text-base border-gray-300 focus:outline-none focus:ring-indigo-500 focus:border-indigo-500 sm:text-sm rounded-md">
                <option value={null}>-- 競技を選択 --</option>
                {#each sports as sport}
                    <option value={sport.id}>{sport.name}</option>
                {/each}
            </select>
        </div>

        <div class="mb-4">
            <label for="match-duration" class="block text-sm font-medium text-gray-700">一試合の時間（分）</label>
            <input type="number" id="match-duration" bind:value={matchDuration} class="mt-1 block w-full pl-3 pr-10 py-2 text-base border-gray-300 focus:outline-none focus:ring-indigo-500 focus:border-indigo-500 sm:text-sm rounded-md" />
        </div>

        <button on:click={generateTournament} class="w-full inline-flex justify-center py-2 px-4 border border-transparent shadow-sm text-sm font-medium rounded-md text-white bg-indigo-600 hover:bg-indigo-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500" disabled={!selectedSportId || teams.length < 2}>
            トーナメント生成
        </button>
    </div>

    <div class="col-span-2">
        <div id="viewer"></div>
    </div>
</div>
