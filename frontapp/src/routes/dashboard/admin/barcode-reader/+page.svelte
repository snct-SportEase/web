<script>
	import { onMount } from 'svelte';
	import { SvelteSet } from 'svelte/reactivity';
	import {
		Html5Qrcode as Html5BarcodeScanner,
		Html5QrcodeSupportedFormats as Html5BarcodeSupportedFormats
	} from 'html5-qrcode';

	let html5BarcodeScanner = $state();
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
	let selectableMatches = $derived(
		selectedSportTournaments.flatMap((tournament, tournamentIndex) => {
			const data = getTournamentData(tournament);
			return (data.matches ?? [])
				.filter((match) => match.id)
				.map((match, matchIndex) => ({
					id: match.id,
					value: `${tournament.id ?? tournamentIndex}:${match.id}:${matchIndex}`,
					label: getMatchLabel(tournament, match),
					round: getMatchRound(match)
				}));
		})
	);
	let selectedMatch = $derived(
		selectedMatchId
			? selectableMatches.find((match) => match.value === selectedMatchId) || null
			: null
	);

	onMount(() => {
		loadInitialData();
		Html5BarcodeScanner.getCameras().catch((err) => {
			errorMessage = `カメラの取得に失敗しました: ${err}`;
		});

		return () => {
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

	function getMatchLabel(tournament, match) {
		const matchNumber = Number(match.order ?? 0) + 1;
		return `${tournament.name} / ${getRoundName(tournament, match)} 第${matchNumber}試合`;
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
			const params = new URLSearchParams({
				event_id: String(activeEventId),
				sport_id: String(selectedSportId)
			});
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
			const response = await fetch('/api/barcode/check-in', {
				method: 'POST',
				credentials: 'include',
				headers: {
					'Content-Type': 'application/json'
				},
				body: JSON.stringify({
					barcode_data: trimmedBarcode,
					event_id: Number(activeEventId),
					sport_id: Number(selectedSportId),
					match_id: Number(selectedMatch.id)
				})
			});

			const result = await response.json();

			if (response.ok) {
				verificationResult = { success: true, data: result };
				manualBarcode = '';
				await loadMatchCheckIns();
			} else {
				verificationResult = { success: false, message: result.error };
			}
		} catch {
			verificationResult = { success: false, message: 'バーコードの確認中にエラーが発生しました' };
		} finally {
			isVerifying = false;
		}
	}

	function startScan() {
		if (!canVerify()) {
			errorMessage = '競技と試合を選択してください';
			return;
		}

		errorMessage = '';
		verificationResult = null;

		html5BarcodeScanner = new Html5BarcodeScanner('barcode-reader-region', {
			formatsToSupport: [
				Html5BarcodeSupportedFormats.CODE_39,
				Html5BarcodeSupportedFormats.CODE_128,
				Html5BarcodeSupportedFormats.EAN_13,
				Html5BarcodeSupportedFormats.EAN_8,
				Html5BarcodeSupportedFormats.ITF,
				Html5BarcodeSupportedFormats.UPC_A,
				Html5BarcodeSupportedFormats.UPC_E
			]
		});

		html5BarcodeScanner
			.start(
				{ facingMode: 'environment' },
				{
					fps: 10,
					qrbox: { width: 320, height: 140 }
				},
				async (decodedText) => {
					await stopScan();
					await verifyBarcode(decodedText);
				},
				() => {}
			)
			.then(() => {
				isScanning = true;
			})
			.catch((err) => {
				errorMessage = `スキャナーの開始に失敗しました: ${err}`;
				isScanning = false;
			});
	}

	async function stopScan() {
		if (html5BarcodeScanner?.isScanning) {
			try {
				await html5BarcodeScanner.stop();
			} catch (err) {
				console.error(`Error stopping scanner: ${err}`);
			}
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
				></div>

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
					<p class="mt-1 text-sm text-gray-600">試合: {selectedMatch.label}</p>
					<p class="mt-1 text-sm text-gray-600">ラウンド: {selectedMatch.round}</p>
				{/if}

				<div class="mt-6 border-t border-gray-200 pt-4">
					<div class="mb-3 flex items-center justify-between">
						<h3 class="text-base font-semibold text-gray-900">この試合のチェックイン済み</h3>
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

	{#if verificationResult}
		{#if verificationResult.success}
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
		{:else}
			<div class="mt-6 rounded border border-red-400 bg-red-100 p-4 text-red-700">
				<p class="font-bold">チェックインできませんでした</p>
				<p>{verificationResult.message}</p>
			</div>
		{/if}
	{/if}

	{#if errorMessage}
		<div class="mt-6 rounded border border-red-400 bg-red-100 p-4 text-red-700">
			<p class="font-bold">エラー:</p>
			<p>{errorMessage}</p>
		</div>
	{/if}
</div>
