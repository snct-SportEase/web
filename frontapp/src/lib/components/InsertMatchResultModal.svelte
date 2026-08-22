<script>
    import Modal from '$lib/components/Modal.svelte';
    let { selectedMatch, selectedTournament, showModal = $bindable(), onclose, onconfirm } = $props();

    let team1Score = $state(0);
    let team2Score = $state(0);

    // モーダルが開かれたときにスコアをリセット
    $effect(() => {
        if (showModal && selectedMatch) {
            team1Score = 0;
            team2Score = 0;
        }
    });

    function closeModal() {
        onclose?.();
    }

    function handleConfirm() {
        if (team1Score < 0 || team2Score < 0) {
            alert("スコアは0以上で入力してください。");
            return;
        }
        onconfirm?.({
            team1_score: team1Score,
            team2_score: team2Score
        });
    }
</script>

{#if selectedMatch}
<Modal bind:open={showModal} title={`結果入力: ${selectedTournament.data.contestants[selectedMatch.sides?.[0]?.contestantId]?.players?.[0]?.title ?? 'TBD'} vs ${selectedTournament.data.contestants[selectedMatch.sides?.[1]?.contestantId]?.players?.[0]?.title ?? 'TBD'}`} onclose={closeModal}>
    <div class="mt-4">
                        <label for="team1-score" class="block text-sm font-medium text-gray-700">{selectedTournament.data.contestants[selectedMatch.sides?.[0]?.contestantId]?.players?.[0]?.title ?? 'Team 1'} Score</label>
                        <input type="number" id="team1-score" bind:value={team1Score} min="0" class="mt-1 block w-full pl-3 pr-10 py-2 text-base border-gray-300 focus:outline-none focus:ring-indigo-500 focus:border-indigo-500 sm:text-sm rounded-md">
    </div>
    <div class="mt-4">
                        <label for="team2-score" class="block text-sm font-medium text-gray-700">{selectedTournament.data.contestants[selectedMatch.sides?.[1]?.contestantId]?.players?.[0]?.title ?? 'Team 2'} Score</label>
                        <input type="number" id="team2-score" bind:value={team2Score} min="0" class="mt-1 block w-full pl-3 pr-10 py-2 text-base border-gray-300 focus:outline-none focus:ring-indigo-500 focus:border-indigo-500 sm:text-sm rounded-md">
    </div>
    {#snippet footer()}
        <button type="button" onclick={handleConfirm} class="w-full inline-flex justify-center rounded-md border border-transparent shadow-sm px-4 py-2 bg-indigo-600 text-base font-medium text-white hover:bg-indigo-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500 sm:ml-3 sm:w-auto sm:text-sm">確認</button>
        <button type="button" onclick={closeModal} class="mt-3 w-full inline-flex justify-center rounded-md border border-gray-300 shadow-sm px-4 py-2 bg-white text-base font-medium text-gray-700 hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500 sm:mt-0 sm:ml-3 sm:w-auto sm:text-sm">キャンセル</button>
    {/snippet}
</Modal>
{/if}
