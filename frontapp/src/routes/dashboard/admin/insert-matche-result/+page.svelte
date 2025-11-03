<script>
	import { onMount } from 'svelte';
	import InsertMatchResultModal from '$lib/components/InsertMatchResultModal.svelte';
	import ConfirmMatchResultModal from '$lib/components/ConfirmMatchResultModal.svelte';
	import { createBracket } from 'bracketry';

	let tournaments = [];
	let selectedTournamentId = '';
	let selectedMatch = null;
	let activeEventId = null;
	let showModal = false;
	let showConfirmModal = false;
	let scoresToSubmit = null;

	onMount(async () => {
		try {
			const eventResponse = await fetch('/api/events/active');
			if (!eventResponse.ok) throw new Error('Failed to get active event');
			const eventData = await eventResponse.json();
			activeEventId = eventData.event_id;

			if (activeEventId) {
				const tournamentsResponse = await fetch(`/api/admin/events/${activeEventId}/tournaments`);
				if (!tournamentsResponse.ok) throw new Error('Failed to fetch tournaments');
				tournaments = await tournamentsResponse.json();
			}
		} catch (error) {
			console.error(error);
		}
	});

	function openModal(match) {
		selectedMatch = match;
		showModal = true;
	}

	function closeModal() {
		showModal = false;
		selectedMatch = null;
	}

	function handleConfirm(event) {
		scoresToSubmit = event.detail;
		showConfirmModal = true;
	}

	function closeConfirmModal() {
		showConfirmModal = false;
		scoresToSubmit = null;
	}

	async function handleSubmit(event) {
		if (!selectedMatch || !scoresToSubmit) return;

		const { team1_score, team2_score } = scoresToSubmit;
		const winnerId = event?.detail?.winnerId;

		const body = {
			team1_score: team1_score,
			team2_score: team2_score
		};

		if (winnerId) {
			body.winner_id = winnerId;
		}

		const response = await fetch(`/api/admin/matches/${selectedMatch.id}/result`, {
			method: 'PUT',
			headers: {
				'Content-Type': 'application/json'
			},
			body: JSON.stringify(body)
		});

		if (response.ok) {
			alert('試合結果を更新しました');
			closeModal();
			closeConfirmModal();
			// Refresh tournaments data
			const tournamentsResponse = await fetch(`/api/admin/events/${activeEventId}/tournaments`);
			tournaments = await tournamentsResponse.json();
		} else {
			alert('試合結果の更新に失敗しました');
		}
	}

	$: selectedTournament = tournaments.find((t) => t.id === selectedTournamentId);

	function renderBracket() {
		setTimeout(() => {
			const wrapper = document.getElementById('bracket-container');
			if (wrapper) {
				wrapper.innerHTML = '';
				if (selectedTournament && selectedTournament.data) {
					console.log(selectedTournament.data);
					createBracket(selectedTournament.data, wrapper);
				} else {
					wrapper.innerHTML = '<p>このトーナメント情報はありません。</p>';
				}
			}
		}, 0);
	}

	$: if (selectedTournament) {
		renderBracket();
	}
</script>

<h1 class="text-2xl font-bold mb-4">試合結果入力</h1>

<div class="mb-4">
	<label for="tournament-select" class="block text-sm font-medium text-gray-700"
		>トーナメント選択</label
	>
	<select
		id="tournament-select"
		bind:value={selectedTournamentId}
		class="mt-1 block w-full pl-3 pr-10 py-2 text-base border-gray-300 focus:outline-none focus:ring-indigo-500 focus:border-indigo-500 sm:text-sm rounded-md"
	>
		<option value="">トーナメントを選択してください</option>
		{#each tournaments as tournament}
			<option value={tournament.id}>{tournament.name}</option>
		{/each}
	</select>
</div>

{#if selectedTournament}
	<h2 class="text-xl font-bold mt-6 mb-2">{selectedTournament.name}</h2>
	<div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
		{#each selectedTournament.data.matches as match}
			<div class="border rounded-lg p-4">
				<p class="font-semibold">Round {match.roundIndex + 1} - Match {match.order + 1}</p>
				<p>
					{selectedTournament.data.contestants[match.sides?.[0]?.contestantId]?.players?.[0]
						?.title ?? 'TBD'}
					vs
					{selectedTournament.data.contestants[match.sides?.[1]?.contestantId]?.players?.[0]
						?.title ?? 'TBD'}
				</p>
				{#if match.sides?.some((side) => side.isWinner)}
					{@const score1 = match.sides?.[0]?.scores?.[0]?.mainScore}
					{@const score2 = match.sides?.[1]?.scores?.[0]?.mainScore}
					<p>Score: {score1 ?? 'N/A'} - {score2 ?? 'N/A'}</p>
					{@const winnerSide = match.sides?.find((side) => side.isWinner)}
					{#if winnerSide}
						{@const winnerName =
							selectedTournament.data.contestants[winnerSide.contestantId]?.players?.[0]?.title ??
							'TBD'}
						<p class="font-bold text-green-600">Winner: {winnerName}</p>
					{:else if score1 !== undefined && score1 === score2}
						<p class="font-bold text-yellow-600">Draw</p>
					{/if}
				{:else}
					<button on:click={() => openModal(match)} class="text-blue-500 hover:underline"
						>結果を入力</button
					>
				{/if}
			</div>
		{/each}
	</div>

	<div id="bracket-container" class="mt-8"></div>
{/if}

<InsertMatchResultModal
	bind:showModal
	{selectedMatch}
	{selectedTournament}
	on:close={closeModal}
	on:confirm={handleConfirm}
/>

<ConfirmMatchResultModal
    bind:showModal={showConfirmModal}
    team1Score={scoresToSubmit?.team1_score}
    team2Score={scoresToSubmit?.team2_score}
    team1Name={selectedTournament?.data.contestants[selectedMatch?.sides?.[0]?.contestantId]?.players?.[0]?.title}
    team2Name={selectedTournament?.data.contestants[selectedMatch?.sides?.[1]?.contestantId]?.players?.[0]?.title}
    team1Id={selectedMatch?.sides?.[0]?.teamId}
    team2Id={selectedMatch?.sides?.[1]?.teamId}
    on:confirm={handleSubmit}
    on:cancel={closeConfirmModal}
/>
