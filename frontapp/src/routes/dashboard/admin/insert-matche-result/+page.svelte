<script>
	import { onMount } from 'svelte';
	import { browser } from '$app/environment';
	import { page } from '$app/stores';
	import InsertMatchResultModal from '$lib/components/InsertMatchResultModal.svelte';
	import ConfirmMatchResultModal from '$lib/components/ConfirmMatchResultModal.svelte';

	let { data } = $page;
	$: user = data.user;
	$: isRoot = user?.roles?.some(role => role.name === 'root');

	let tournaments = [];
	let selectedTournamentId = '';
	let selectedMatch = null;
	let activeEventId = null;
	let showModal = false;
	let showConfirmModal = false;
	let scoresToSubmit = null;
	let isRainyMode = false;

	let ws;

	$: if (selectedTournamentId && typeof window !== 'undefined') {
		if (ws) {
			ws.close();
		}
		const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
		const host = window.location.host;
		ws = new WebSocket(`${protocol}//${host}/api/ws/tournaments/${selectedTournamentId}`);

		ws.onmessage = (event) => {
			const data = JSON.parse(event.data);
			if (data.type === 'update') {
				// refetch tournament data
				fetch(`/api/admin/events/${activeEventId}/tournaments`)
					.then((res) => res.json())
					.then((data) => {
						tournaments = data;
					});
			}
		};

		ws.onclose = () => {
			console.log('WebSocket connection closed');
		};

		ws.onerror = (error) => {
			console.error('WebSocket error:', error);
		};
	}

	onMount(async () => {
		try {
			const eventResponse = await fetch('/api/events/active');
			if (!eventResponse.ok) throw new Error('Failed to get active event');
			const eventData = await eventResponse.json();
			activeEventId = eventData.event_id;

			if (activeEventId) {
				// アクティブイベントの詳細を取得して雨天時モードの状態を確認
				const eventsResponse = await fetch('/api/root/events');
				if (eventsResponse.ok) {
					const events = await eventsResponse.json();
					const activeEvent = events.find(e => e.id === activeEventId);
					if (activeEvent) {
						isRainyMode = activeEvent.is_rainy_mode || false;
					}
				}

				const tournamentsResponse = await fetch(`/api/admin/events/${activeEventId}/tournaments`);
				if (!tournamentsResponse.ok) throw new Error('Failed to fetch tournaments');
				tournaments = await tournamentsResponse.json();
			}
		} catch (error) {
			console.error(error);
		}

		return () => {
			if (ws) {
				ws.close();
			}
		};
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
			let errorMessage = '試合結果の更新に失敗しました';
			if (response.status === 403) {
				const errorData = await response.json().catch(() => ({}));
				errorMessage = errorData.error || '既に入力済みの試合結果の修正はroot権限のみ可能です';
			} else if (response.status === 400 || response.status === 500) {
				const errorData = await response.json().catch(() => ({}));
				if (errorData.error) {
					errorMessage = errorData.error;
				}
			}
			alert(errorMessage);
		}
	}

	$: selectedTournament = tournaments.find((t) => t.id === selectedTournamentId);
	// 敗者戦トーナメントが選択されていて、雨天時モードが無効な場合は選択を解除
	$: if (selectedTournament && selectedTournament.name.includes('敗者戦') && !isRainyMode) {
		selectedTournamentId = '';
		selectedTournament = null;
	}

	async function renderBracket() {
		if (!browser) return;
		setTimeout(async () => {
			const wrapper = document.getElementById('bracket-container');
			if (wrapper) {
				wrapper.innerHTML = '';
				if (selectedTournament && selectedTournament.data) {
					try {
						const { createBracket } = await import('bracketry');
						createBracket(selectedTournament.data, wrapper);
					} catch (error) {
						console.error('Failed to load createBracket:', error);
						wrapper.innerHTML = '<p>ブラケットの読み込みに失敗しました。</p>';
					}
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
			{@const isLoserBracket = tournament.name.includes('敗者戦')}
			{#if !isLoserBracket || isRainyMode}
				<option value={tournament.id}>{tournament.name}</option>
			{/if}
		{/each}
	</select>
</div>

{#if selectedTournament}
	<h2 class="text-xl font-bold mt-6 mb-2">{selectedTournament.name}</h2>
	<div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
		{#each selectedTournament.data.matches as match}
			{@const isLoserBracketMatch = match.isLoserBracketMatch}
			{#if !isLoserBracketMatch || isRainyMode}
				<div class="border rounded-lg p-4 {isLoserBracketMatch ? 'bg-yellow-50' : ''}">
					<p class="font-semibold">
						{#if isLoserBracketMatch}
							敗者戦{match.loserBracketBlock ? match.loserBracketBlock + 'ブロック' : ''} Round {match.roundIndex + 1} - Match {match.order + 1}
						{:else}
							Round {match.roundIndex + 1} - Match {match.order + 1}
						{/if}
					</p>
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
						{#if isRoot}
							<button on:click={() => openModal(match)} class="mt-2 text-orange-600 hover:underline text-sm"
								>結果を修正（rootのみ）</button
							>
						{/if}
					{:else}
						<button on:click={() => openModal(match)} class="text-blue-500 hover:underline"
							>結果を入力</button
						>
					{/if}
				</div>
			{/if}
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
