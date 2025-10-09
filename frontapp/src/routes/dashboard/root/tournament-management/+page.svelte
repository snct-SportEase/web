<script>
    import { onMount } from 'svelte';

    let sports = [];
    let teams = [];
    let selectedSportId = null;
    let participants = [];
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
            } else {
                console.error('Failed to fetch teams');
                teams = [];
            }
        } catch (error) {
            console.error('Error fetching teams:', error);
            teams = [];
        }
    }

    function addParticipant(team) {
        if (!participants.find(p => p.id === team.id)) {
            participants = [...participants, team];
        }
    }

    function removeParticipant(team) {
        participants = participants.filter(p => p.id !== team.id);
    }

    async function generateTournament() {
        if (participants.length < 2) {
            alert('トーナメントを生成するには、少なくとも2チームの参加者が必要です。');
            return;
        }

        const viewer = new window.BracketsViewer();
        const manager = new window.BracketsManager(viewer);

        const participantNames = participants.map(p => p.name);

        await manager.create({
            name: 'My awesome tournament',
            tournamentId: 0,
            type: 'single_elimination',
            seeding: participantNames,
            settings: { seedOrdering: ['natural'] },
        });

        const viewerElement = document.querySelector('#viewer');
        viewer.render(viewerElement, manager.get.storage());
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

        {#if selectedSportId}
            <div class="mb-4">
                <h3 class="text-lg font-medium text-gray-900">チーム一覧</h3>
                <ul class="mt-2 border border-gray-200 rounded-md divide-y divide-gray-200">
                    {#each teams as team}
                        <li class="pl-3 pr-4 py-3 flex items-center justify-between text-sm">
                            <div class="w-0 flex-1 flex items-center">
                                <span>{team.name}</span>
                            </div>
                            <div class="ml-4 flex-shrink-0">
                                <button on:click={() => addParticipant(team)} class="font-medium text-indigo-600 hover:text-indigo-500">追加</button>
                            </div>
                        </li>
                    {/each}
                </ul>
            </div>
        {/if}

        <div class="mb-4">
            <label for="match-duration" class="block text-sm font-medium text-gray-700">一試合の時間（分）</label>
            <input type="number" id="match-duration" bind:value={matchDuration} class="mt-1 block w-full pl-3 pr-10 py-2 text-base border-gray-300 focus:outline-none focus:ring-indigo-500 focus:border-indigo-500 sm:text-sm rounded-md" />
        </div>

        <div class="mb-4">
            <h3 class="text-lg font-medium text-gray-900">参加チーム</h3>
            <ul class="mt-2 border border-gray-200 rounded-md divide-y divide-gray-200">
                {#each participants as participant}
                    <li class="pl-3 pr-4 py-3 flex items-center justify-between text-sm">
                        <div class="w-0 flex-1 flex items-center">
                            <span>{participant.name}</span>
                        </div>
                        <div class="ml-4 flex-shrink-0">
                            <button on:click={() => removeParticipant(participant)} class="font-medium text-red-600 hover:text-red-500">削除</button>
                        </div>
                    </li>
                {/each}
            </ul>
        </div>

        <button on:click={generateTournament} class="w-full inline-flex justify-center py-2 px-4 border border-transparent shadow-sm text-sm font-medium rounded-md text-white bg-indigo-600 hover:bg-indigo-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500">
            トーナメント生成
        </button>
    </div>

    <div class="col-span-2">
        <div id="viewer" class="brackets-viewer"></div>
    </div>
</div>
