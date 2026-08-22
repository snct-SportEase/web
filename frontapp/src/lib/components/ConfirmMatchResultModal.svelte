
<script>
    import Modal from '$lib/components/Modal.svelte';
    import ModalFooter from '$lib/components/ModalFooter.svelte';
    let { showModal = $bindable(false), team1Score = 0, team2Score = 0, team1Name = 'Team 1', team2Name = 'Team 2', team1Id = null, team2Id = null, onconfirm, oncancel } = $props();

    let selectedWinnerId = $state(null);

    function confirm() {
        if (isTie) {
            onconfirm?.({ winnerId: selectedWinnerId });
        } else {
            onconfirm?.();
        }
    }

    function cancel() {
        oncancel?.();
    }

    let isTie = $derived(team1Score === team2Score);
    let winnerName = $derived(team1Score > team2Score ? team1Name : team2Name);

    $effect(() => {
        if (showModal) {
            selectedWinnerId = null;
        }
    });
</script>

<Modal bind:open={showModal} title="試合結果確認" onclose={cancel}>
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
    {#snippet footer()}
        <ModalFooter confirmLabel="登録" confirmDisabled={isTie && !selectedWinnerId} onconfirm={confirm} oncancel={cancel} />
    {/snippet}
</Modal>
