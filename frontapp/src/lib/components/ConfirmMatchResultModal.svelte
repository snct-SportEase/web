
<script>
    import { createEventDispatcher } from 'svelte';

    export let showModal = false;
    export let team1Score = 0;
    export let team2Score = 0;
    export let team1Name = 'Team 1';
    export let team2Name = 'Team 2';
    export let team1Id = null;
    export let team2Id = null;

    let selectedWinnerId = null;

    const dispatch = createEventDispatcher();

    function confirm() {
        if (isTie) {
            dispatch('confirm', { winnerId: selectedWinnerId });
        } else {
            dispatch('confirm');
        }
    }

    function cancel() {
        dispatch('cancel');
    }

    $: isTie = team1Score === team2Score;
    $: winnerName = team1Score > team2Score ? team1Name : team2Name;

    $: if (showModal) {
        selectedWinnerId = null;
    }
</script>

{#if showModal}
<div class="fixed z-50 inset-0 overflow-y-auto">
    <div class="flex items-end justify-center min-h-screen pt-4 px-4 pb-20 text-center sm:block sm:p-0">
        <div class="fixed inset-0 transition-opacity" aria-hidden="true">
            <div class="absolute inset-0 bg-gray-500 opacity-75"></div>
        </div>

        <span class="hidden sm:inline-block sm:align-middle sm:h-screen" aria-hidden="true">&#8203;</span>

        <div class="relative z-50 inline-block align-bottom bg-white rounded-lg text-left overflow-hidden shadow-xl transform transition-all sm:my-8 sm:align-middle sm:max-w-lg sm:w-full">
            <div class="bg-white px-4 pt-5 pb-4 sm:p-6 sm:pb-4">
                <h3 class="text-lg leading-6 font-medium text-gray-900">試合結果確認</h3>
                <div class="mt-4">
                    <p>{team1Name}: {team1Score}</p>
                    <p>{team2Name}: {team2Score}</p>
                    {#if isTie}
                        <div class="mt-4">
                            <p class="font-bold">勝者を選択してください:</p>
                            <div class="flex items-center mt-2">
                                <input type="radio" id="team1" name="winner" value={team1Id} bind:group={selectedWinnerId} class="focus:ring-indigo-500 h-4 w-4 text-indigo-600 border-gray-300">
                                <label for="team1" class="ml-3 block text-sm font-medium text-gray-700">{team1Name}</label>
                            </div>
                            <div class="flex items-center mt-2">
                                <input type="radio" id="team2" name="winner" value={team2Id} bind:group={selectedWinnerId} class="focus:ring-indigo-500 h-4 w-4 text-indigo-600 border-gray-300">
                                <label for="team2" class="ml-3 block text-sm font-medium text-gray-700">{team2Name}</label>
                            </div>
                        </div>
                    {:else}
                        <p class="font-bold mt-4">勝者: {winnerName}</p>
                    {/if}
                </div>
            </div>
            <div class="bg-gray-50 px-4 py-3 sm:px-6 sm:flex sm:flex-row-reverse">
                <button type="button" on:click={confirm} disabled={isTie && !selectedWinnerId} class="w-full inline-flex justify-center rounded-md border border-transparent shadow-sm px-4 py-2 bg-indigo-600 text-base font-medium text-white hover:bg-indigo-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500 sm:ml-3 sm:w-auto sm:text-sm disabled:opacity-50">登録</button>
                <button type="button" on:click={cancel} class="mt-3 w-full inline-flex justify-center rounded-md border border-gray-300 shadow-sm px-4 py-2 bg-white text-base font-medium text-gray-700 hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500 sm:mt-0 sm:ml-3 sm:w-auto sm:text-sm">キャンセル</button>
            </div>
        </div>
    </div>
</div>
{/if}
