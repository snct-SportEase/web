<script lang="ts">
    import Modal from '$lib/components/Modal.svelte';
    import ModalFooter from '$lib/components/ModalFooter.svelte';
    import type { BracketData, Match } from '@snct-sportease/matchline';
    let { selectedMatch, selectedTournament, showModal = $bindable(false), onclose, onconfirm }: { selectedMatch: Pick<Match, 'sides'> | null; selectedTournament: { data: Pick<BracketData, 'contestants'> } | null | undefined; showModal?: boolean; onclose?: () => void; onconfirm?: (scores: { team1_score: number; team2_score: number }) => void } = $props();

    function getTeamName(index: number, fallback: string): string {
        const id = selectedMatch?.sides?.[index]?.contestantId;
        return (id && selectedTournament?.data.contestants?.[id]?.players?.[0]?.title) || fallback;
    }

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
<Modal bind:open={showModal} title={`結果入力: ${getTeamName(0, 'TBD')} vs ${getTeamName(1, 'TBD')}`} onclose={closeModal}>
    <div class="mt-4">
                        <label for="team1-score" class="block text-sm font-medium text-gray-700">{getTeamName(0, 'Team 1')} Score</label>
                        <input type="number" id="team1-score" bind:value={team1Score} min="0" class="mt-1 block w-full pl-3 pr-10 py-2 text-base border-gray-300 focus:outline-none focus:ring-indigo-500 focus:border-indigo-500 sm:text-sm rounded-md">
    </div>
    <div class="mt-4">
                        <label for="team2-score" class="block text-sm font-medium text-gray-700">{getTeamName(1, 'Team 2')} Score</label>
                        <input type="number" id="team2-score" bind:value={team2Score} min="0" class="mt-1 block w-full pl-3 pr-10 py-2 text-base border-gray-300 focus:outline-none focus:ring-indigo-500 focus:border-indigo-500 sm:text-sm rounded-md">
    </div>
    {#snippet footer()}
        <ModalFooter confirmLabel="確認" onconfirm={handleConfirm} oncancel={closeModal} />
    {/snippet}
</Modal>
{/if}
