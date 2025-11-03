<script>
    import { createEventDispatcher } from 'svelte';

    export let selectedMatch;
    export let selectedTournament;
    export let showModal;

    let team1Score = 0;
    let team2Score = 0;

    const dispatch = createEventDispatcher();

    function closeModal() {
        dispatch('close');
    }

    function handleConfirm() {
        if (team1Score < 0 || team2Score < 0) {
            alert("スコアは0以上で入力してください。");
            return;
        }
        dispatch('confirm', {
            team1_score: team1Score,
            team2_score: team2Score
        });
    }
</script>

{#if showModal && selectedMatch}
<div class="fixed z-50 inset-0 overflow-y-auto">
    <div class="flex items-end justify-center min-h-screen pt-4 px-4 pb-20 text-center sm:block sm:p-0">
        <div class="fixed inset-0 transition-opacity" aria-hidden="true">
            <div class="absolute inset-0 bg-gray-500 opacity-75"></div>
        </div>

        <span class="hidden sm:inline-block sm:align-middle sm:h-screen" aria-hidden="true">&#8203;</span>

        <div class="relative z-50 inline-block align-bottom bg-white rounded-lg text-left overflow-hidden shadow-xl transform transition-all sm:my-8 sm:align-middle sm:max-w-lg sm:w-full">
            <form on:submit|preventDefault={handleConfirm}>
                <div class="bg-white px-4 pt-5 pb-4 sm:p-6 sm:pb-4">
                    <h3 class="text-lg leading-6 font-medium text-gray-900">結果入力: {selectedTournament.data.contestants[selectedMatch.sides?.[0]?.contestantId]?.players?.[0]?.title ?? 'TBD'} vs {selectedTournament.data.contestants[selectedMatch.sides?.[1]?.contestantId]?.players?.[0]?.title ?? 'TBD'}</h3>
                    <div class="mt-4">
                        <label for="team1-score" class="block text-sm font-medium text-gray-700">{selectedTournament.data.contestants[selectedMatch.sides?.[0]?.contestantId]?.players?.[0]?.title ?? 'Team 1'} Score</label>
                        <input type="number" id="team1-score" bind:value={team1Score} min="0" class="mt-1 block w-full pl-3 pr-10 py-2 text-base border-gray-300 focus:outline-none focus:ring-indigo-500 focus:border-indigo-500 sm:text-sm rounded-md">
                    </div>
                    <div class="mt-4">
                        <label for="team2-score" class="block text-sm font-medium text-gray-700">{selectedTournament.data.contestants[selectedMatch.sides?.[1]?.contestantId]?.players?.[0]?.title ?? 'Team 2'} Score</label>
                        <input type="number" id="team2-score" bind:value={team2Score} min="0" class="mt-1 block w-full pl-3 pr-10 py-2 text-base border-gray-300 focus:outline-none focus:ring-indigo-500 focus:border-indigo-500 sm:text-sm rounded-md">
                    </div>
                </div>
                <div class="bg-gray-50 px-4 py-3 sm:px-6 sm:flex sm:flex-row-reverse">
                    <button type="submit" class="w-full inline-flex justify-center rounded-md border border-transparent shadow-sm px-4 py-2 bg-indigo-600 text-base font-medium text-white hover:bg-indigo-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500 sm:ml-3 sm:w-auto sm:text-sm">確認</button>
                    <button type="button" on:click={closeModal} class="mt-3 w-full inline-flex justify-center rounded-md border border-gray-300 shadow-sm px-4 py-2 bg-white text-base font-medium text-gray-700 hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500 sm:mt-0 sm:ml-3 sm:w-auto sm:text-sm">キャンセル</button>
                </div>
            </form>
        </div>
    </div>
</div>
{/if}
