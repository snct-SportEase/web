
<script>
    import { createEventDispatcher } from 'svelte';

    export let showModal = false;
    export let team1Score = 0;
    export let team2Score = 0;
    export let team1Name = 'Team 1';
    export let team2Name = 'Team 2';

    const dispatch = createEventDispatcher();

    function confirm() {
        dispatch('confirm');
    }

    function cancel() {
        dispatch('cancel');
    }

    $: winnerName = team1Score > team2Score ? team1Name : team2Name;
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
                    <p class="font-bold mt-4">勝者: {winnerName}</p>
                </div>
            </div>
            <div class="bg-gray-50 px-4 py-3 sm:px-6 sm:flex sm:flex-row-reverse">
                <button type="button" on:click={confirm} class="w-full inline-flex justify-center rounded-md border border-transparent shadow-sm px-4 py-2 bg-indigo-600 text-base font-medium text-white hover:bg-indigo-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500 sm:ml-3 sm:w-auto sm:text-sm">登録</button>
                <button type="button" on:click={cancel} class="mt-3 w-full inline-flex justify-center rounded-md border border-gray-300 shadow-sm px-4 py-2 bg-white text-base font-medium text-gray-700 hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500 sm:mt-0 sm:ml-3 sm:w-auto sm:text-sm">キャンセル</button>
            </div>
        </div>
    </div>
</div>
{/if}
