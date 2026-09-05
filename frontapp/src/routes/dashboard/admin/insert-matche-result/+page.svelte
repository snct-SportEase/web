<script>
	import { onMount } from 'svelte';
	import { browser } from '$app/environment';
	import { page } from '$app/stores';
	import InsertMatchResultModal from '$lib/components/InsertMatchResultModal.svelte';
	import ConfirmMatchResultModal from '$lib/components/ConfirmMatchResultModal.svelte';

	let { data } = $page;
	let user = $derived(data.user);
	let isRoot = $derived(user?.roles?.some(role => role.name === 'root'));

	let tournaments = $state([]);
	let selectedTournamentId = $state('');
	let selectedMatch = $state(null);
	let activeEventId = $state(null);
	let activeEventStatus = $state('');
	let showModal = $state(false);
	let showConfirmModal = $state(false);
	let scoresToSubmit = $state(null);
	let isRainyMode = $state(false);
	let boardGameRuns = $state([]);
	let rankingSelections = $state([]);
	let rankingTournamentId = $state(null);
	let isSavingRankings = $state(false);

	let ws;

	$effect(() => {
		if (selectedTournamentId && typeof window !== 'undefined') {
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
					fetchBoardGameRuns();
				}
			};

			ws.onclose = () => {
				console.log('WebSocket connection closed');
			};

			ws.onerror = (error) => {
				console.error('WebSocket error:', error);
			};
		}
	});

	onMount(async () => {
		try {
			const eventResponse = await fetch('/api/events/active');
			if (!eventResponse.ok) throw new Error('Failed to get active event');
			const eventData = await eventResponse.json();
			activeEventId = eventData.event_id;
			activeEventStatus = eventData.status || '';
			isRainyMode = eventData.is_rainy_mode || false;

			if (activeEventId) {
				const tournamentsResponse = await fetch(`/api/admin/events/${activeEventId}/tournaments`);
				if (!tournamentsResponse.ok) throw new Error('Failed to fetch tournaments');
				tournaments = await tournamentsResponse.json();
				await fetchBoardGameRuns();
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
		if (activeEventStatus !== 'active') {
			alert('試合結果は開催中の大会でのみ入力できます。');
			return;
		}
		selectedMatch = match;
		showModal = true;
	}

	function closeModal() {
		showModal = false;
		selectedMatch = null;
	}

	function handleConfirm(scores) {
		scoresToSubmit = scores;
		showConfirmModal = true;
	}

	function closeConfirmModal() {
		showConfirmModal = false;
		scoresToSubmit = null;
	}

	async function handleSubmit(result = {}) {
		if (activeEventStatus !== 'active') {
			alert('試合結果は開催中の大会でのみ入力できます。');
			return;
		}
		if (!selectedMatch || !scoresToSubmit) return;

		const { team1_score, team2_score } = scoresToSubmit;
		const winnerId = result?.winnerId;

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
			await fetchBoardGameRuns();
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

	let selectedTournament = $derived(tournaments.find((t) => t.id === selectedTournamentId));
	let selectedBoardGame = $derived.by(() => {
		for (const run of boardGameRuns) {
			const tournament = run.tournaments?.find((item) => item.id === selectedTournamentId);
			if (tournament) return { run, tournament };
		}
		return null;
	});

	async function fetchBoardGameRuns() {
		if (!activeEventId) return;
		const response = await fetch(`/api/admin/events/${activeEventId}/board-game-runs`);
		if (response.ok) boardGameRuns = await response.json();
	}

	function updateRankingSelection(index, value) {
		const next = [...rankingSelections];
		next[index] = value ? Number(value) : '';
		rankingSelections = next;
	}

	async function saveBoardGameRankings() {
		if (!selectedBoardGame || activeEventStatus !== 'active') {
			alert('順位は開催中の大会でのみ登録できます。');
			return;
		}
		const required = Math.min(4, selectedBoardGame.tournament.entries.length);
		if (rankingSelections.slice(0, required).some((value) => !value) || new Set(rankingSelections.slice(0, required)).size !== required) {
			alert('1位から順に、重複しない出場枠を選択してください。');
			return;
		}
		isSavingRankings = true;
		try {
			const response = await fetch(`/api/admin/board-game-runs/${selectedBoardGame.run.id}/tournaments/${selectedBoardGame.tournament.id}/rankings`, {
				method: 'PUT',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({ rankings: rankingSelections.slice(0, required).map((entryID, index) => ({ entry_id: entryID, rank: index + 1 })) }),
			});
			const result = await response.json();
			if (!response.ok) throw new Error(result.error || '順位の登録に失敗しました');
			boardGameRuns = boardGameRuns.map((run) => run.id === result.id ? result : run);
			alert('順位点を登録しました。');
		} catch (error) {
			alert(error.message || '順位の登録に失敗しました');
		} finally {
			isSavingRankings = false;
		}
	}
	// 敗者戦トーナメントが選択されていて、雨天時モードが無効な場合は選択を解除
	$effect(() => {
		if (selectedTournament && selectedTournament.name.includes('敗者戦') && !isRainyMode) {
			selectedTournamentId = '';
		}
	});

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

	$effect(() => {
		if (selectedTournament) {
			renderBracket();
		}
	});

	$effect(() => {
		if (selectedBoardGame && rankingTournamentId !== selectedBoardGame.tournament.id) {
			rankingTournamentId = selectedBoardGame.tournament.id;
			const byRank = [...(selectedBoardGame.tournament.rankings || [])].sort((a, b) => a.rank - b.rank);
			rankingSelections = byRank.map((ranking) => ranking.entry_id);
		} else if (!selectedBoardGame) {
			rankingTournamentId = null;
			rankingSelections = [];
		}
	});
</script>

<h1 class="text-2xl font-bold mb-4">試合結果入力</h1>

{#if activeEventId && activeEventStatus !== 'active'}
	<p class="mb-4 rounded-md border border-amber-200 bg-amber-50 px-4 py-3 text-sm text-amber-800">
		試合結果は大会が「開催中」になってから入力できます。
	</p>
{/if}

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
		{#each tournaments as tournament (tournament.id)}
			{@const isLoserBracket = tournament.name.includes('敗者戦')}
			{#if !isLoserBracket || isRainyMode}
				<option value={tournament.id}>{tournament.name}</option>
			{/if}
		{/each}
	</select>
</div>

{#if selectedTournament}
	<h2 class="text-xl font-bold mt-6 mb-2">{selectedTournament.name}</h2>
	{#if selectedBoardGame}
		<section class="mb-6 rounded-lg border border-amber-200 bg-amber-50 p-4">
			<div class="flex flex-wrap items-start justify-between gap-3">
				<div>
					<h3 class="font-semibold">盤上競技の順位登録</h3>
					<p class="text-sm text-gray-600">会場: {selectedBoardGame.run.location}／1勝 {selectedBoardGame.run.win_points}点。全試合の結果入力後に確定してください。</p>
				</div>
				{#if selectedBoardGame.run.rules_pdf_url}
					<a class="text-sm text-blue-700 underline" href={selectedBoardGame.run.rules_pdf_url} target="_blank" rel="noreferrer">ルールPDF</a>
				{/if}
			</div>
			<div class="mt-3 grid gap-2 sm:grid-cols-2 lg:grid-cols-4">
				{#each Array.from({ length: Math.min(4, selectedBoardGame.tournament.entries.length) }, (_, index) => index) as rankIndex (rankIndex)}
					<label class="text-sm font-medium">{rankIndex + 1}位（{selectedBoardGame.run.rank_points?.[String(rankIndex + 1)] ?? 0}点）
						<select class="mt-1 w-full rounded-md border-gray-300" value={rankingSelections[rankIndex] || ''} onchange={(event) => updateRankingSelection(rankIndex, event.currentTarget.value)}>
							<option value="">選択してください</option>
							{#each selectedBoardGame.tournament.entries as entry (entry.id)}
								<option value={entry.id}>{entry.team_name}</option>
							{/each}
						</select>
					</label>
				{/each}
			</div>
			<button class="mt-3 rounded-md bg-amber-700 px-4 py-2 text-sm font-medium text-white hover:bg-amber-800 disabled:opacity-50" disabled={isSavingRankings || activeEventStatus !== 'active'} onclick={saveBoardGameRankings}>
				{isSavingRankings ? '登録中...' : selectedBoardGame.tournament.rankings?.length ? '順位を修正して再集計' : '順位を確定して得点へ反映'}
			</button>
		</section>
	{/if}
	<div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
		{#each selectedTournament.data.matches as match (match.id)}
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
							<button onclick={() => openModal(match)} class="mt-2 text-orange-600 hover:underline text-sm"
								>結果を修正（rootのみ）</button
							>
						{/if}
					{:else}
						<button onclick={() => openModal(match)} class="text-blue-500 hover:underline"
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
	onclose={closeModal}
	onconfirm={handleConfirm}
/>

<ConfirmMatchResultModal
    bind:showModal={showConfirmModal}
    team1Score={scoresToSubmit?.team1_score ?? 0}
    team2Score={scoresToSubmit?.team2_score ?? 0}
    team1Name={selectedTournament?.data.contestants[selectedMatch?.sides?.[0]?.contestantId]?.players?.[0]?.title}
    team2Name={selectedTournament?.data.contestants[selectedMatch?.sides?.[1]?.contestantId]?.players?.[0]?.title}
    team1Id={selectedMatch?.sides?.[0]?.teamId}
    team2Id={selectedMatch?.sides?.[1]?.teamId}
    onconfirm={handleSubmit}
    oncancel={closeConfirmModal}
/>
