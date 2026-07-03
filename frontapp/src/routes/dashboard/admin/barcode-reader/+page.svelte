<script>
	import { onMount } from 'svelte';
	import { SvelteMap, SvelteSet } from 'svelte/reactivity';

	const supportedBarcodeFormats = ['code_39', 'code_128', 'ean_13', 'ean_8', 'itf', 'upc_a', 'upc_e'];
	const matchSelectionOpenOffsetMs = 10 * 60 * 1000;

	let barcodeVideo = $state();
	let barcodeStream = null;
	let barcodeScanFrame = null;
	let selectionClockInterval = null;
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
	let checkedInMembers = $state([]);
	let uncheckedMembers = $state([]);
	let matchCheckInCount = $state(0);
	let checkInsLoading = $state(false);
	let checkInsError = $state('');
	let checkInStatusModalMode = $state(null);
	let currentTime = $state(Date.now());

	let selectedSport = $derived(
		selectedSportId ? sports.find((sport) => `${sport.id}` === `${selectedSportId}`) : null
	);
	let selectedSportTournaments = $derived(
		selectedSportId
			? tournaments.filter((tournament) => `${tournament.sport_id}` === `${selectedSportId}`)
			: []
	);
	let selectableMatches = $derived(buildMatchSelections(selectedSportTournaments, currentTime));
	let selectedMatch = $derived(
		selectedMatchId
			? selectableMatches.find((match) => match.value === selectedMatchId) || null
			: null
	);

	onMount(() => {
		loadInitialData();
		window.addEventListener('keydown', handleKeydown);
		selectionClockInterval = window.setInterval(() => {
			currentTime = Date.now();
		}, 1000);

		return () => {
			window.removeEventListener('keydown', handleKeydown);
			if (selectionClockInterval !== null) {
				window.clearInterval(selectionClockInterval);
				selectionClockInterval = null;
			}
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
		return activeEventId !== null && selectedSportId !== '' && selectedMatch?.isSelectable === true;
	}

	function handleKeydown(event) {
		if (event.key === 'Escape' && verificationResult) {
			closeResultModal();
		}
		if (event.key === 'Escape' && checkInStatusModalMode) {
			closeCheckInStatusModal();
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

	function buildMatchSelections(tournamentsForSport, nowMs) {
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
						selections.push(createMatchSelection([entry], null, nowMs));
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
			selections.push(createMatchSelection(group.entries, group, nowMs));
		}

		return selections.sort((a, b) => a.sortIndex - b.sortIndex);
	}

	function createMatchSelection(entries, timedGroup = null, nowMs = Date.now()) {
		const firstEntry = entries[0];
		const matchIds = entries.map((entry) => entry.match.id);
		const matchupLabels = entries.map((entry) => entry.matchupLabel).filter(Boolean);
		const matchupSummary = matchupLabels.join(', ');
		const teamNamesById = getTeamNamesById(entries);
		const sortIndex = Math.min(...entries.map((entry) => entry.sortIndex));
		const startTimeMs = getSelectionStartTimeMs(entries);
		const opensAtMs = startTimeMs === null ? null : startTimeMs - matchSelectionOpenOffsetMs;
		const isSelectable = opensAtMs === null || nowMs >= opensAtMs;
		const lockedLabel = isSelectable ? '' : `（${formatMatchSelectionOpensAt(opensAtMs)}から選択可）`;

		if (timedGroup) {
			const summaryLabel = `${timedGroup.roundName} ${timedGroup.startLabel}開始試合`;
			return {
				id: matchIds[0],
				matchIds,
				value: `time:${matchIds.join('-')}`,
				label: `${matchupSummary ? `${summaryLabel}（${matchupSummary}）` : summaryLabel}${lockedLabel}`,
				summaryLabel,
				matchupLabel: matchupSummary,
				matchupLabels,
				teamNamesById,
				round: timedGroup.round,
				isSelectable,
				opensAtMs,
				sortIndex
			};
		}

		const baseLabel = getMatchLabel(firstEntry.tournament, firstEntry.match);
		return {
			id: firstEntry.match.id,
			matchIds,
			value: `${firstEntry.tournament.id ?? 0}:${firstEntry.match.id}:${sortIndex}`,
			label: `${baseLabel}${lockedLabel}`,
			summaryLabel: baseLabel,
			matchupLabel: matchupSummary,
			matchupLabels,
			teamNamesById,
			round: getMatchRound(firstEntry.match),
			isSelectable,
			opensAtMs,
			sortIndex
		};
	}

	function getMatchStartTime(match) {
		return match?.startTime || match?.rainyModeStartTime || '';
	}

	function getTeamNamesById(entries) {
		return entries.reduce((names, entry) => {
			const sides = Array.isArray(entry.match?.sides) ? entry.match.sides : [];
			sides.forEach((side) => {
				if (!side?.teamId) {
					return;
				}
				names[String(side.teamId)] = getSideName(entry.data, side);
			});
			return names;
		}, {});
	}

	function getSelectionStartTimeMs(entries) {
		const startTimes = entries
			.map((entry) => parseMatchStartTimeMs(getMatchStartTime(entry.match)))
			.filter((value) => value !== null);

		if (startTimes.length === 0) {
			return null;
		}
		return Math.min(...startTimes);
	}

	function parseMatchStartTimeMs(value) {
		const parts = parseRegisteredJstDateTime(value);
		if (!parts) {
			return null;
		}

		return Date.UTC(parts.year, parts.month - 1, parts.day, parts.hour - 9, parts.minute, parts.second);
	}

	function formatMatchSelectionOpensAt(value) {
		if (value === null) {
			return '';
		}

		return new Date(value).toLocaleTimeString('ja-JP', {
			timeZone: 'Asia/Tokyo',
			hour: '2-digit',
			minute: '2-digit'
		});
	}

	function getMatchStartTimeKey(match) {
		const rawValue = getMatchStartTime(match);
		const value = String(rawValue || '').trim();
		if (!value) {
			return '';
		}

		const parts = parseRegisteredJstDateTime(rawValue);
		if (!parts) {
			return value;
		}

		const year = parts.year;
		const month = String(parts.month).padStart(2, '0');
		const day = String(parts.day).padStart(2, '0');
		const hour = String(parts.hour).padStart(2, '0');
		const minute = String(parts.minute).padStart(2, '0');
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

		const parts = parseRegisteredJstDateTime(value);
		if (!parts) {
			return value;
		}

		return `${String(parts.hour).padStart(2, '0')}:${String(parts.minute).padStart(2, '0')}`;
	}

	function parseRegisteredJstDateTime(value) {
		const trimmedValue = String(value || '').trim();
		if (!trimmedValue) {
			return null;
		}

		const match = trimmedValue.match(
			/^(\d{4})-(\d{1,2})-(\d{1,2})(?:T|\s)(\d{1,2}):(\d{2})(?::(\d{2}))?/
		);
		if (!match) {
			return null;
		}

		const [, year, month, day, hour, minute, second = '0'] = match;
		return {
			year: Number(year),
			month: Number(month),
			day: Number(day),
			hour: Number(hour),
			minute: Number(minute),
			second: Number(second)
		};
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
		const nextMatchId = event.currentTarget.value;
		const nextMatch = selectableMatches.find((match) => match.value === nextMatchId);
		selectedMatchId = nextMatch?.isSelectable === false ? '' : nextMatchId;
		await loadMatchCheckIns();
	}

	function clearMatchCheckIns() {
		matchCheckIns = [];
		checkedInMembers = [];
		uncheckedMembers = [];
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
			checkedInMembers = data.checked_in_members || data.members || [];
			uncheckedMembers = data.unchecked_members || [];
			matchCheckIns = checkedInMembers;
			matchCheckInCount = data.checked_in_count ?? data.count ?? matchCheckIns.length;
		} catch (err) {
			checkInsError = err.message || 'チェックイン済み一覧の取得に失敗しました';
			matchCheckIns = [];
			checkedInMembers = [];
			uncheckedMembers = [];
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
				verificationResult = {
					success: false,
					alreadyCheckedIn: Boolean(result.already_checked_in),
					message: result.error
				};
				if (result.already_checked_in) {
					await loadMatchCheckIns();
				}
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
		checkInStatusModalMode = null;
		errorMessage = '';
		manualBarcode = '';
	}

	function closeResultModal() {
		verificationResult = null;
	}

	function openCheckInStatusModal(mode) {
		if (!selectedMatch) {
			return;
		}
		checkInStatusModalMode = mode;
	}

	function closeCheckInStatusModal() {
		checkInStatusModalMode = null;
	}

	function getCheckInStatusModalTitle() {
		return checkInStatusModalMode === 'unchecked'
			? '未チェックインの学生'
			: 'チェックイン済みの学生';
	}

	function getCheckInStatusModalMembers() {
		return checkInStatusModalMode === 'unchecked' ? uncheckedMembers : checkedInMembers;
	}

	function getCheckInStatusClassGroups() {
		const groups = new SvelteMap();

		for (const member of [...checkedInMembers, ...uncheckedMembers]) {
			const key = getMemberGroupKey(member);
			if (!groups.has(key)) {
				groups.set(key, {
					key,
					name: getMemberClassName(member),
					checked: [],
					unchecked: []
				});
			}
		}

		for (const member of checkedInMembers) {
			groups.get(getMemberGroupKey(member))?.checked.push(member);
		}
		for (const member of uncheckedMembers) {
			groups.get(getMemberGroupKey(member))?.unchecked.push(member);
		}

		return Array.from(groups.values()).sort((a, b) => a.name.localeCompare(b.name, 'ja-JP'));
	}

	function getCheckInStatusGroupMembers(group) {
		return checkInStatusModalMode === 'unchecked' ? group.unchecked : group.checked;
	}

	function getCheckInStatusGroupCountLabel(group) {
		const totalCount = group.checked.length + group.unchecked.length;
		if (checkInStatusModalMode === 'unchecked') {
			return `未チェックイン: ${group.unchecked.length} / ${totalCount}人`;
		}
		return `チェックイン済み: ${group.checked.length} / ${totalCount}人`;
	}

	function getMemberDisplayName(member) {
		return member?.display_name || member?.email || '未設定';
	}

	function getMemberGroupKey(member) {
		return `${member?.class_id ?? 'unknown'}:${member?.team_id ?? 'unknown'}:${getMemberClassName(member)}`;
	}

	function getMemberClassName(member) {
		return (
			member?.class_name ||
			selectedMatch?.teamNamesById?.[String(member?.team_id)] ||
			member?.team_name ||
			'クラス未設定'
		);
	}

	function getMemberMeta(member) {
		const values = [];
		if (member?.team_name && member.team_name !== member?.class_name) {
			values.push(member.team_name);
		}
		if (member?.email) {
			values.push(member.email);
		}
		return values.join(' / ');
	}
</script>

<div class="container mx-auto p-6">
	<h1 class="text-2xl font-bold mb-6">MyIDバーコード読み取り</h1>

	{#if loading}
		<div class="rounded bg-white p-6 shadow-sm">
			<p class="text-gray-500">読み込み中...</p>
		</div>
	{:else}
		{#if selectedMatch}
			<div class="mb-4 rounded bg-white p-4 shadow-sm">
				<div class="flex flex-col gap-3 md:flex-row md:items-center md:justify-between">
					<div>
						<p class="text-sm text-gray-500">選択中の試合</p>
						<p class="text-base font-semibold text-gray-900">{selectedMatch.summaryLabel}</p>
					</div>
					<div class="grid grid-cols-1 gap-2 sm:grid-cols-2 md:min-w-80">
						<button
							type="button"
							onclick={() => openCheckInStatusModal('checked')}
							disabled={checkInsLoading}
							class="rounded border border-blue-200 bg-blue-50 px-3 py-2 text-sm font-semibold text-blue-700 hover:bg-blue-100 disabled:cursor-not-allowed disabled:border-gray-200 disabled:bg-gray-100 disabled:text-gray-500"
						>
							チェックイン済み（{checkedInMembers.length}人）
						</button>
						<button
							type="button"
							onclick={() => openCheckInStatusModal('unchecked')}
							disabled={checkInsLoading}
							class="rounded border border-amber-200 bg-amber-50 px-3 py-2 text-sm font-semibold text-amber-800 hover:bg-amber-100 disabled:cursor-not-allowed disabled:border-gray-200 disabled:bg-gray-100 disabled:text-gray-500"
						>
							未チェックイン（{uncheckedMembers.length}人）
						</button>
					</div>
				</div>
			</div>
		{/if}

		<div class="grid gap-6 lg:grid-cols-[minmax(0,1fr)_minmax(320px,420px)]">
			<section class="rounded bg-white p-6 shadow-sm">
				<div class="mb-4">
					<p class="text-sm text-gray-500">開催中イベント</p>
					<p class="text-lg font-semibold text-gray-900">{activeEventName || '未設定'}</p>
				</div>

				<label for="sport-select" class="mb-2 block text-sm font-medium text-gray-700">競技</label>
				<select
					id="sport-select"
					bind:value={selectedSportId}
					onchange={handleSportChange}
					disabled={isScanning || isVerifying}
					class="w-full rounded border border-gray-300 px-4 py-2 focus:outline-none focus:ring-2 focus:ring-blue-500"
				>
					<option value="">競技を選択してください</option>
					{#each sports as sport, index (`${sport.id}:${index}`)}
						<option value={`${sport.id}`}>{sport.name}</option>
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
						<option value={match.value} disabled={!match.isSelectable}>{match.label}</option>
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

	{#if verificationResult}
		<div
			class="fixed inset-0 z-50 flex items-center justify-center bg-black/50 px-4"
			role="presentation"
			onclick={closeResultModal}
		>
			<div
				class="w-full max-w-md rounded bg-white p-6 shadow-xl"
				role="dialog"
				aria-modal="true"
				aria-labelledby="check-in-result-title"
				tabindex="-1"
				onclick={(event) => event.stopPropagation()}
				onkeydown={(event) => event.stopPropagation()}
			>
				{#if verificationResult.success}
					<h2 id="check-in-result-title" class="text-lg font-semibold text-gray-900">
						ラウンドチェックインを完了しました
					</h2>
					<div class="mt-3 space-y-1 text-sm leading-6 text-gray-700">
						<p>氏名: {verificationResult.data.display_name || '未設定'}</p>
						<p>学籍番号: {verificationResult.data.student_number}</p>
						<p>競技: {verificationResult.data.sport_name}</p>
						<p>ラウンド: {verificationResult.data.round}</p>
						{#if verificationResult.data.capacity_warning}
							<p class="mt-2 font-semibold text-amber-700">
								{verificationResult.data.capacity_warning}
							</p>
						{/if}
					</div>
				{:else}
					<h2 id="check-in-result-title" class="text-lg font-semibold text-gray-900">
						{verificationResult.alreadyCheckedIn ? 'チェックイン済みです' : 'チェックインできませんでした'}
					</h2>
					<p class="mt-3 text-sm leading-6 text-gray-700">{verificationResult.message}</p>
				{/if}
				<div class="mt-6 flex justify-end">
					<button
						type="button"
						onclick={closeResultModal}
						class="rounded bg-blue-600 px-4 py-2 font-semibold text-white hover:bg-blue-700"
					>
						閉じる
					</button>
				</div>
			</div>
		</div>
	{/if}

	{#if checkInStatusModalMode}
		<div
			class="fixed inset-0 z-50 flex items-center justify-center bg-black/50 px-4"
			role="presentation"
			onclick={closeCheckInStatusModal}
		>
			<div
				class="max-h-[85vh] w-full max-w-lg overflow-hidden rounded bg-white shadow-xl"
				role="dialog"
				aria-modal="true"
				aria-labelledby="check-in-status-title"
				tabindex="-1"
				onclick={(event) => event.stopPropagation()}
				onkeydown={(event) => event.stopPropagation()}
			>
				<div class="border-b border-gray-200 p-6">
					<div class="flex items-center justify-between gap-4">
						<h2 id="check-in-status-title" class="text-lg font-semibold text-gray-900">
							{getCheckInStatusModalTitle()}
						</h2>
						<span class="text-sm text-gray-600">{getCheckInStatusModalMembers().length} 人</span>
					</div>
					{#if selectedMatch}
						<p class="mt-1 text-sm text-gray-600">{selectedMatch.summaryLabel}</p>
					{/if}
				</div>

				<div class="max-h-[55vh] overflow-y-auto p-6">
					{#if getCheckInStatusClassGroups().length === 0}
						<p class="text-sm text-gray-500">
							{checkInStatusModalMode === 'unchecked'
								? '未チェックインの学生はいません'
								: 'チェックイン済みの学生はいません'}
						</p>
					{:else}
						<div class="space-y-3">
							{#each getCheckInStatusClassGroups() as group (group.key)}
								<section class="rounded border border-gray-200">
									<div class="border-b border-gray-200 bg-gray-50 px-3 py-2">
										<div class="flex items-center justify-between gap-3">
											<h3 class="text-sm font-semibold text-gray-900">{group.name}</h3>
											<span class="text-xs font-medium text-gray-600">
												{getCheckInStatusGroupCountLabel(group)}
											</span>
										</div>
									</div>
									{#if getCheckInStatusGroupMembers(group).length === 0}
										<p class="px-3 py-2 text-sm text-gray-500">
											{checkInStatusModalMode === 'unchecked'
												? '未チェックインの学生はいません'
												: 'チェックイン済みの学生はいません'}
										</p>
									{:else}
										<ul class="divide-y divide-gray-200">
											{#each getCheckInStatusGroupMembers(group) as member (`${member.user_id}:${member.team_id}`)}
												<li class="px-3 py-2">
													<p class="text-sm font-medium text-gray-900">
														{getMemberDisplayName(member)}
													</p>
													{#if getMemberMeta(member)}
														<p class="text-xs text-gray-600">{getMemberMeta(member)}</p>
													{/if}
													{#if checkInStatusModalMode === 'checked' && formatCheckedInAt(member.checked_in_at)}
														<p class="text-xs text-gray-500">
															{formatCheckedInAt(member.checked_in_at)}
														</p>
													{/if}
												</li>
											{/each}
										</ul>
									{/if}
								</section>
							{/each}
						</div>
					{/if}
				</div>

				<div class="flex justify-end border-t border-gray-200 p-4">
					<button
						type="button"
						onclick={closeCheckInStatusModal}
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
