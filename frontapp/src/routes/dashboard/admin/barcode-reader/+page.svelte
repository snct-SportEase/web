<script>
	import { onMount } from 'svelte';
	import { SvelteMap, SvelteSet } from 'svelte/reactivity';

	const supportedBarcodeFormats = ['code_39', 'code_128', 'ean_13', 'ean_8', 'itf', 'upc_a', 'upc_e'];

	let barcodeVideo = $state();
	let barcodeStream = null;
	let barcodeScanFrame = null;
	let errorMessage = $state('');
	let verificationResult = $state(null);
	let activeEventId = $state(null);
	let activeEventName = $state('');
	let sports = $state([]);
	let tournaments = $state([]);
	let selectedSportId = $state('');
	let selectedMatchId = $state('');
	let manualBarcode = $state('');
	let isScanning = $state(false);
	let isVerifying = $state(false);
	let loading = $state(true);
	let matchCheckIns = $state([]);
	let matchCheckInCount = $state(0);
	let checkInsLoading = $state(false);
	let checkInsError = $state('');

	let selectedSport = $derived(
		selectedSportId ? sports.find((sport) => `${sport.id}` === `${selectedSportId}`) : null
	);
	let selectedSportTournaments = $derived(
		selectedSportId
			? tournaments.filter((tournament) => `${tournament.sport_id}` === `${selectedSportId}`)
			: []
	);
	let selectableMatches = $derived(buildMatchSelections(selectedSportTournaments));
	let selectedMatch = $derived(
		selectedMatchId
			? selectableMatches.find((match) => match.value === selectedMatchId) || null
			: null
	);

	onMount(() => {
		loadInitialData();
		window.addEventListener('keydown', handleKeydown);

		return () => {
			window.removeEventListener('keydown', handleKeydown);
			stopScan();
		};
	});

	async function loadInitialData() {
		loading = true;
		errorMessage = '';

		try {
			const eventResponse = await fetch('/api/events/active', { credentials: 'include' });

			if (!eventResponse.ok) {
				throw new Error('開催中イベントの取得に失敗しました');
			}

			const eventData = await eventResponse.json();
			activeEventId = eventData?.event_id ?? eventData?.id ?? null;
			activeEventName = eventData?.event_name ?? eventData?.name ?? '';

			if (!activeEventId) {
				sports = [];
				tournaments = [];
				return;
			}

			const [sportsResponse, tournamentsResponse] = await Promise.all([
				fetch(`/api/events/${activeEventId}/sports`, { credentials: 'include' }),
				fetch(`/api/admin/events/${activeEventId}/tournaments`, { credentials: 'include' })
			]);
			if (!sportsResponse.ok) {
				throw new Error('競技一覧の取得に失敗しました');
			}
			if (!tournamentsResponse.ok) {
				throw new Error('試合一覧の取得に失敗しました');
			}
			sports = dedupeSportsByID(await sportsResponse.json());
			tournaments = await tournamentsResponse.json();
		} catch (err) {
			errorMessage = err.message || '初期データの取得に失敗しました';
		} finally {
			loading = false;
		}
	}

	function canVerify() {
		return activeEventId !== null && selectedSportId !== '' && selectedMatch !== null;
	}

	function handleKeydown(event) {
		if (event.key === 'Escape' && verificationResult && !verificationResult.success) {
			closeFailureModal();
		}
	}

	function dedupeSportsByID(items) {
		const seen = new SvelteSet();
		return (items ?? [])
			.map(normalizeSport)
			.filter(Boolean)
			.filter((sport) => {
				const key = `${sport?.id}`;
				if (!sport?.id || seen.has(key)) {
					return false;
				}
				seen.add(key);
				return true;
			});
	}

	function normalizeSport(sport) {
		const id = sport?.id ?? sport?.sport_id;
		const name = sport?.name ?? sport?.sport_name;
		if (id === undefined || id === null || !name) {
			return null;
		}

		return {
			...sport,
			id,
			name
		};
	}

	function getTournamentData(tournament) {
		if (!tournament?.data) {
			return { rounds: [], matches: [] };
		}
		if (typeof tournament.data === 'string') {
			try {
				return JSON.parse(tournament.data);
			} catch {
				return { rounds: [], matches: [] };
			}
		}
		return tournament.data;
	}

	function getMatchRound(match) {
		return Number(match?.roundIndex ?? 0) + 1;
	}

	function getRoundName(tournament, match) {
		const data = getTournamentData(tournament);
		return data.rounds?.[match.roundIndex]?.name || `第${getMatchRound(match)}ラウンド`;
	}

	function buildMatchSelections(tournamentsForSport) {
		const timedGroups = new SvelteMap();
		const selections = [];

		tournamentsForSport.forEach((tournament, tournamentIndex) => {
			const data = getTournamentData(tournament);
			(data.matches ?? [])
				.filter((match) => match.id)
				.forEach((match, matchIndex) => {
					const startKey = getMatchStartTimeKey(match);
					const startLabel = getMatchStartTimeLabel(match);
					const entry = {
						tournament,
						data,
						match,
						matchupLabel: getMatchupLabel(data, match),
						sortIndex:
							tournamentIndex * 100000 +
							Number(match.roundIndex ?? 0) * 1000 +
							Number(match.order ?? matchIndex)
					};

					if (!startKey) {
						selections.push(createMatchSelection([entry]));
						return;
					}

					const roundName = getRoundName(tournament, match);
					const groupKey = `${Number(match.roundIndex ?? 0)}:${startKey}`;
					if (!timedGroups.has(groupKey)) {
						timedGroups.set(groupKey, {
							round: getMatchRound(match),
							roundName,
							startLabel,
							entries: []
						});
					}
					timedGroups.get(groupKey).entries.push(entry);
				});
		});

		for (const group of timedGroups.values()) {
			selections.push(createMatchSelection(group.entries, group));
		}

		return selections.sort((a, b) => a.sortIndex - b.sortIndex);
	}

	function createMatchSelection(entries, timedGroup = null) {
		const firstEntry = entries[0];
		const matchIds = entries.map((entry) => entry.match.id);
		const matchupLabels = entries.map((entry) => entry.matchupLabel).filter(Boolean);
		const matchupSummary = matchupLabels.join(', ');
		const sortIndex = Math.min(...entries.map((entry) => entry.sortIndex));

		if (timedGroup) {
			const summaryLabel = `${timedGroup.roundName} ${timedGroup.startLabel}開始試合`;
			return {
				id: matchIds[0],
				matchIds,
				value: `time:${matchIds.join('-')}`,
				label: matchupSummary ? `${summaryLabel}（${matchupSummary}）` : summaryLabel,
				summaryLabel,
				matchupLabel: matchupSummary,
				matchupLabels,
				round: timedGroup.round,
				sortIndex
			};
		}

		return {
			id: firstEntry.match.id,
			matchIds,
			value: `${firstEntry.tournament.id ?? 0}:${firstEntry.match.id}:${sortIndex}`,
			label: getMatchLabel(firstEntry.tournament, firstEntry.match),
			summaryLabel: getMatchLabel(firstEntry.tournament, firstEntry.match),
			matchupLabel: matchupSummary,
			matchupLabels,
			round: getMatchRound(firstEntry.match),
			sortIndex
		};
	}

	function getMatchStartTime(match) {
		return match?.startTime || match?.rainyModeStartTime || '';
	}

	function getMatchStartTimeKey(match) {
		const value = String(getMatchStartTime(match) || '').trim();
		if (!value) {
			return '';
		}

		const date = new Date(value.replace(' ', 'T'));
		if (Number.isNaN(date.getTime())) {
			return value;
		}

		const year = date.getFullYear();
		const month = String(date.getMonth() + 1).padStart(2, '0');
		const day = String(date.getDate()).padStart(2, '0');
		const hour = String(date.getHours()).padStart(2, '0');
		const minute = String(date.getMinutes()).padStart(2, '0');
		return `${year}-${month}-${day}T${hour}:${minute}`;
	}

	function getMatchStartTimeLabel(match) {
		const value = String(getMatchStartTime(match) || '').trim();
		if (!value) {
			return '';
		}

		const timeOnlyMatch = value.match(/^(\d{1,2}:\d{2})/);
		if (timeOnlyMatch) {
			return timeOnlyMatch[1];
		}

		const date = new Date(value.replace(' ', 'T'));
		if (Number.isNaN(date.getTime())) {
			return value;
		}

		return date.toLocaleTimeString('ja-JP', {
			hour: '2-digit',
			minute: '2-digit'
		});
	}

	function getMatchLabel(tournament, match) {
		const matchNumber = Number(match.order ?? 0) + 1;
		const matchupLabel = getMatchupLabel(getTournamentData(tournament), match);
		const baseLabel = `${tournament.name} / ${getRoundName(tournament, match)} 第${matchNumber}試合`;
		return matchupLabel ? `${baseLabel}（${matchupLabel}）` : baseLabel;
	}

	function getMatchupLabel(tournamentData, match) {
		const sides = Array.isArray(match?.sides) ? match.sides : [];
		if (sides.length === 0) {
			return '';
		}

		const teamNames = sides.slice(0, 2).map((side) => getSideName(tournamentData, side));
		if (teamNames.every((name) => name === '未定')) {
			return '';
		}

		return `${teamNames[0] ?? '未定'} vs ${teamNames[1] ?? '未定'}`;
	}

	function getSideName(tournamentData, side) {
		if (!side) {
			return '未定';
		}
		if (side.title) {
			return side.title;
		}

		const contestant = tournamentData?.contestants?.[side.contestantId];
		return contestant?.players?.[0]?.title || '未定';
	}

	function handleSportChange(event) {
		selectedSportId = event.currentTarget.value;
		selectedMatchId = '';
		clearMatchCheckIns();
	}

	async function handleMatchChange(event) {
		selectedMatchId = event.currentTarget.value;
		await loadMatchCheckIns();
	}

	function clearMatchCheckIns() {
		matchCheckIns = [];
		matchCheckInCount = 0;
		checkInsError = '';
	}

	async function loadMatchCheckIns() {
		if (!activeEventId || !selectedSportId || !selectedMatch) {
			clearMatchCheckIns();
			return;
		}

		checkInsLoading = true;
		checkInsError = '';

		try {
			// eslint-disable-next-line svelte/prefer-svelte-reactivity
			const params = new URLSearchParams({
				event_id: String(activeEventId),
				sport_id: String(selectedSportId)
			});
			params.set('match_ids', selectedMatch.matchIds.join(','));
			const response = await fetch(`/api/barcode/matches/${selectedMatch.id}/check-ins?${params}`, {
				credentials: 'include'
			});
			const data = await response.json();
			if (!response.ok) {
				throw new Error(data.error || 'チェックイン済み一覧の取得に失敗しました');
			}
			matchCheckIns = data.members || [];
			matchCheckInCount = data.count || matchCheckIns.length;
		} catch (err) {
			checkInsError = err.message || 'チェックイン済み一覧の取得に失敗しました';
			matchCheckIns = [];
			matchCheckInCount = 0;
		} finally {
			checkInsLoading = false;
		}
	}

	function formatCheckedInAt(value) {
		if (!value) {
			return '';
		}
		const date = new Date(value);
		if (Number.isNaN(date.getTime())) {
			return '';
		}
		return date.toLocaleString('ja-JP', {
			month: '2-digit',
			day: '2-digit',
			hour: '2-digit',
			minute: '2-digit'
		});
	}

	async function verifyBarcode(barcodeData) {
		const trimmedBarcode = barcodeData.trim();
		if (!trimmedBarcode) {
			errorMessage = 'バーコードを入力してください';
			return;
		}
		if (!canVerify()) {
			errorMessage = '競技と試合を選択してください';
			return;
		}

		isVerifying = true;
		errorMessage = '';

		try {
			const payload = {
				barcode_data: trimmedBarcode,
				event_id: Number(activeEventId),
				sport_id: Number(selectedSportId),
				match_id: Number(selectedMatch.id),
				match_ids: selectedMatch.matchIds.map(Number)
			};

			const response = await fetch('/api/barcode/check-in', {
				method: 'POST',
				credentials: 'include',
				headers: {
					'Content-Type': 'application/json'
				},
				body: JSON.stringify(payload)
			});

			const result = await response.json();

			if (response.ok) {
				verificationResult = { success: true, data: result };
				manualBarcode = '';
				await loadMatchCheckIns();
			} else {
				console.error('Barcode check-in failed:', result);
				verificationResult = { success: false, message: result.error };
			}
		} catch (err) {
			console.error('Barcode check-in request failed:', err);
			verificationResult = { success: false, message: 'バーコードの確認中にエラーが発生しました' };
		} finally {
			isVerifying = false;
		}
	}

	async function startScan() {
		if (!canVerify()) {
			errorMessage = '競技と試合を選択してください';
			return;
		}
		if (isScanning) {
			return;
		}

		errorMessage = '';
		verificationResult = null;

		try {
			const barcodeDetector = await createBarcodeDetector();
			if (!navigator.mediaDevices?.getUserMedia) {
				throw new Error('このブラウザではカメラを利用できません。手入力を使用してください');
			}

			barcodeStream = await navigator.mediaDevices.getUserMedia({
				video: {
					facingMode: { ideal: 'environment' }
				}
			});
			if (!barcodeVideo) {
				throw new Error('バーコード読み取り領域を初期化できませんでした');
			}
			barcodeVideo.srcObject = barcodeStream;
			await barcodeVideo.play();
			isScanning = true;
			scanBarcodeFrame(barcodeDetector);
		} catch (err) {
			await stopScan();
			errorMessage = err.message || `スキャナーの開始に失敗しました: ${err}`;
		}
	}

	async function createBarcodeDetector() {
		if (typeof window === 'undefined' || !('BarcodeDetector' in window)) {
			throw new Error('このブラウザはバーコード読み取りに対応していません。手入力を使用してください');
		}

		const detectorClass = window.BarcodeDetector;
		let formats = supportedBarcodeFormats;
		if (typeof detectorClass.getSupportedFormats === 'function') {
			const browserFormats = await detectorClass.getSupportedFormats();
			formats = supportedBarcodeFormats.filter((format) => browserFormats.includes(format));
		}
		if (formats.length === 0) {
			throw new Error('このブラウザは対応するバーコード形式を読み取れません。手入力を使用してください');
		}

		return new detectorClass({ formats });
	}

	async function scanBarcodeFrame(barcodeDetector) {
		if (!isScanning || !barcodeVideo) {
			return;
		}

		try {
			if (barcodeVideo.readyState >= HTMLMediaElement.HAVE_CURRENT_DATA) {
				const detectedBarcodes = await barcodeDetector.detect(barcodeVideo);
				const detectedValue = detectedBarcodes.find((barcode) => barcode.rawValue)?.rawValue;
				if (detectedValue) {
					await stopScan();
					await verifyBarcode(detectedValue);
					return;
				}
			}
		} catch {
			await stopScan();
			errorMessage = 'バーコード読み取りに失敗しました。手入力を使用してください';
			return;
		}

		barcodeScanFrame = window.requestAnimationFrame(() => scanBarcodeFrame(barcodeDetector));
	}

	async function stopScan() {
		if (barcodeScanFrame !== null) {
			window.cancelAnimationFrame(barcodeScanFrame);
			barcodeScanFrame = null;
		}
		if (barcodeStream) {
			barcodeStream.getTracks().forEach((track) => track.stop());
			barcodeStream = null;
		}
		if (barcodeVideo) {
			barcodeVideo.pause();
			barcodeVideo.srcObject = null;
		}
		isScanning = false;
	}

	async function submitManualBarcode(event) {
		event.preventDefault();
		await verifyBarcode(manualBarcode);
	}

	function resetScanner() {
		verificationResult = null;
		errorMessage = '';
		manualBarcode = '';
	}

	function closeFailureModal() {
		if (verificationResult && !verificationResult.success) {
			verificationResult = null;
		}
	}
</script>

<div class="container mx-auto p-6">
	<h1 class="text-2xl font-bold mb-6">MyIDバーコード読み取り</h1>

	{#if loading}
		<div class="rounded bg-white p-6 shadow-sm">
			<p class="text-gray-500">読み込み中...</p>
		</div>
	{:else}
		<div class="grid gap-6 lg:grid-cols-[minmax(0,1fr)_minmax(320px,420px)]">
			<section class="rounded bg-white p-6 shadow-sm">
				<div class="mb-4">
					<p class="text-sm text-gray-500">開催中イベント</p>
					<p class="text-lg font-semibold text-gray-900">{activeEventName || '未設定'}</p>
				</div>

				<label for="sport-select" class="mb-2 block text-sm font-medium text-gray-700">競技</label>
				<select
					id="sport-select"
					value={selectedSportId}
					onchange={handleSportChange}
					disabled={isScanning || isVerifying}
					class="w-full rounded border border-gray-300 px-4 py-2 focus:outline-none focus:ring-2 focus:ring-blue-500"
				>
					<option value="">競技を選択してください</option>
					{#each sports as sport, index (`${sport.id}:${index}`)}
						<option value={sport.id}>{sport.name}</option>
					{/each}
				</select>

				<label for="match-select" class="mb-2 mt-4 block text-sm font-medium text-gray-700">試合</label>
				<select
					id="match-select"
					bind:value={selectedMatchId}
					onchange={handleMatchChange}
					disabled={isScanning || isVerifying || !selectedSportId}
					class="w-full rounded border border-gray-300 px-4 py-2 focus:outline-none focus:ring-2 focus:ring-blue-500"
				>
					<option value="">
						{selectedSportId ? '試合を選択してください' : '先に競技を選択してください'}
					</option>
					{#each selectableMatches as match (match.value)}
						<option value={match.value}>{match.label}</option>
					{/each}
				</select>

				<div
					id="barcode-reader-region"
					class="mt-6 min-h-48 w-full overflow-hidden rounded border-2 border-gray-300 bg-gray-50"
				>
					<video
						bind:this={barcodeVideo}
						class="h-full min-h-48 w-full object-cover"
						playsinline
						muted
					></video>
				</div>

				<div class="mt-4 flex flex-wrap gap-3">
					<button
						type="button"
						onclick={startScan}
						disabled={!canVerify() || isScanning || isVerifying}
						class="rounded bg-blue-600 px-4 py-2 font-semibold text-white hover:bg-blue-700 disabled:cursor-not-allowed disabled:bg-gray-400"
					>
						読み取り開始
					</button>
					<button
						type="button"
						onclick={stopScan}
						disabled={!isScanning}
						class="rounded bg-red-600 px-4 py-2 font-semibold text-white hover:bg-red-700 disabled:cursor-not-allowed disabled:bg-gray-400"
					>
						停止
					</button>
					<button
						type="button"
						onclick={resetScanner}
						disabled={isVerifying}
						class="rounded bg-gray-600 px-4 py-2 font-semibold text-white hover:bg-gray-700 disabled:cursor-not-allowed disabled:bg-gray-400"
					>
						リセット
					</button>
				</div>
			</section>

			<aside class="rounded bg-white p-6 shadow-sm">
				<h2 class="mb-4 text-lg font-semibold text-gray-900">手入力</h2>
				<form class="space-y-4" onsubmit={submitManualBarcode}>
					<div>
						<label for="manual-barcode" class="mb-2 block text-sm font-medium text-gray-700">バーコード値</label>
						<input
							id="manual-barcode"
							type="text"
							inputmode="text"
							bind:value={manualBarcode}
							placeholder="H102301059"
							disabled={isVerifying}
							class="w-full rounded border border-gray-300 px-4 py-2 font-mono focus:outline-none focus:ring-2 focus:ring-blue-500"
						/>
					</div>
					<button
						type="submit"
						disabled={!canVerify() || isVerifying || manualBarcode.trim() === ''}
						class="w-full rounded bg-blue-600 px-4 py-2 font-semibold text-white hover:bg-blue-700 disabled:cursor-not-allowed disabled:bg-gray-400"
					>
						{isVerifying ? 'チェックイン中...' : 'チェックインする'}
					</button>
				</form>

				{#if selectedSport}
					<p class="mt-4 text-sm text-gray-600">選択中: {selectedSport.name}</p>
				{/if}
				{#if selectedMatch}
					<p class="mt-1 text-sm text-gray-600">試合: {selectedMatch.summaryLabel}</p>
					{#if selectedMatch.matchupLabels.length > 0}
						<div class="mt-1 text-sm text-gray-600">
							<p>対戦:</p>
							<ul class="mt-1 space-y-1">
								{#each selectedMatch.matchupLabels as matchupLabel (matchupLabel)}
									<li>{matchupLabel}</li>
								{/each}
							</ul>
						</div>
					{/if}
					<p class="mt-1 text-sm text-gray-600">ラウンド: {selectedMatch.round}</p>
				{/if}

				<div class="mt-6 border-t border-gray-200 pt-4">
					<div class="mb-3 flex items-center justify-between">
						<h3 class="text-base font-semibold text-gray-900">
							{selectedMatch?.matchIds?.length > 1 ? 'この時間帯のチェックイン済み' : 'この試合のチェックイン済み'}
						</h3>
						<span class="text-sm text-gray-600">{matchCheckInCount} 人</span>
					</div>

					{#if checkInsError}
						<p class="rounded border border-red-300 bg-red-50 px-3 py-2 text-sm text-red-700">
							{checkInsError}
						</p>
					{:else if !selectedMatch}
						<p class="text-sm text-gray-500">試合を選択してください</p>
					{:else if checkInsLoading}
						<p class="text-sm text-gray-500">読み込み中...</p>
					{:else if matchCheckIns.length === 0}
						<p class="text-sm text-gray-500">まだチェックインしていません</p>
					{:else}
						<ul class="max-h-72 divide-y divide-gray-200 overflow-y-auto rounded border border-gray-200">
							{#each matchCheckIns as member (member.user_id)}
								<li class="px-3 py-2">
									<p class="text-sm font-medium text-gray-900">
										{member.display_name || '未設定'}
									</p>
									<p class="text-xs text-gray-600">
										{member.class_name} / {member.team_name}
									</p>
									<p class="text-xs text-gray-500">
										{member.email}
										{#if formatCheckedInAt(member.checked_in_at)}
											・{formatCheckedInAt(member.checked_in_at)}
										{/if}
									</p>
								</li>
							{/each}
						</ul>
					{/if}
				</div>
			</aside>
		</div>
	{/if}

	{#if verificationResult?.success}
		<div class="mt-6 rounded border border-green-400 bg-green-100 p-4 text-green-800">
			<p class="font-bold">参加本登録とラウンドチェックインを完了しました</p>
			<p>氏名: {verificationResult.data.display_name || '未設定'}</p>
			<p>学籍番号: {verificationResult.data.student_number}</p>
			<p>競技: {verificationResult.data.sport_name}</p>
			<p>ラウンド: {verificationResult.data.round}</p>
			{#if verificationResult.data.capacity_warning}
				<p class="mt-2 font-semibold">{verificationResult.data.capacity_warning}</p>
			{/if}
		</div>
	{/if}

	{#if verificationResult && !verificationResult.success}
		<div
			class="fixed inset-0 z-50 flex items-center justify-center bg-black/50 px-4"
			role="presentation"
			onclick={closeFailureModal}
		>
			<div
				class="w-full max-w-md rounded bg-white p-6 shadow-xl"
				role="dialog"
				aria-modal="true"
				aria-labelledby="check-in-failure-title"
				tabindex="-1"
				onclick={(event) => event.stopPropagation()}
				onkeydown={(event) => event.stopPropagation()}
			>
				<h2 id="check-in-failure-title" class="text-lg font-semibold text-gray-900">
					チェックインできませんでした
				</h2>
				<p class="mt-3 text-sm leading-6 text-gray-700">{verificationResult.message}</p>
				<div class="mt-6 flex justify-end">
					<button
						type="button"
						onclick={closeFailureModal}
						class="rounded bg-blue-600 px-4 py-2 font-semibold text-white hover:bg-blue-700"
					>
						閉じる
					</button>
				</div>
			</div>
		</div>
	{/if}

	{#if errorMessage}
		<div class="mt-6 rounded border border-red-400 bg-red-100 p-4 text-red-700">
			<p class="font-bold">エラー:</p>
			<p>{errorMessage}</p>
		</div>
	{/if}
</div>
