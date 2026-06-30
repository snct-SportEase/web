<script>
	import { onMount } from 'svelte';
	import { Html5Qrcode, Html5QrcodeSupportedFormats } from 'html5-qrcode';

	let html5QrCode = $state();
	let errorMessage = $state('');
	let verificationResult = $state(null);
	let activeEventId = $state(null);
	let activeEventName = $state('');
	let sports = $state([]);
	let selectedSportId = $state('');
	let manualBarcode = $state('');
	let isScanning = $state(false);
	let isVerifying = $state(false);
	let loading = $state(true);

	let selectedSport = $derived(
		selectedSportId ? sports.find((sport) => `${sport.id}` === `${selectedSportId}`) : null
	);

	onMount(() => {
		loadInitialData();
		Html5Qrcode.getCameras().catch((err) => {
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
				return;
			}

			const sportsResponse = await fetch(`/api/events/${activeEventId}/sports`, { credentials: 'include' });
			if (!sportsResponse.ok) {
				throw new Error('競技一覧の取得に失敗しました');
			}
			sports = await sportsResponse.json();
		} catch (err) {
			errorMessage = err.message || '初期データの取得に失敗しました';
		} finally {
			loading = false;
		}
	}

	function canVerify() {
		return activeEventId !== null && selectedSportId !== '';
	}

	async function verifyBarcode(barcodeData) {
		const trimmedBarcode = barcodeData.trim();
		if (!trimmedBarcode) {
			errorMessage = 'バーコードを入力してください';
			return;
		}
		if (!canVerify()) {
			errorMessage = '競技を選択してください';
			return;
		}

		isVerifying = true;
		errorMessage = '';

		try {
			const response = await fetch('/api/qrcode/verify', {
				method: 'POST',
				credentials: 'include',
				headers: {
					'Content-Type': 'application/json'
				},
				body: JSON.stringify({
					barcode_data: trimmedBarcode,
					event_id: Number(activeEventId),
					sport_id: Number(selectedSportId)
				})
			});

			const result = await response.json();

			if (response.ok) {
				verificationResult = { success: true, data: result };
				manualBarcode = '';
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
			errorMessage = '競技を選択してください';
			return;
		}

		errorMessage = '';
		verificationResult = null;

		html5QrCode = new Html5Qrcode('barcode-reader-region', {
			formatsToSupport: [
				Html5QrcodeSupportedFormats.CODE_39,
				Html5QrcodeSupportedFormats.CODE_128,
				Html5QrcodeSupportedFormats.EAN_13,
				Html5QrcodeSupportedFormats.EAN_8,
				Html5QrcodeSupportedFormats.ITF,
				Html5QrcodeSupportedFormats.UPC_A,
				Html5QrcodeSupportedFormats.UPC_E
			]
		});

		html5QrCode
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
		if (html5QrCode?.isScanning) {
			try {
				await html5QrCode.stop();
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
					bind:value={selectedSportId}
					disabled={isScanning || isVerifying}
					class="w-full rounded border border-gray-300 px-4 py-2 focus:outline-none focus:ring-2 focus:ring-blue-500"
				>
					<option value="">競技を選択してください</option>
					{#each sports as sport (sport.id)}
						<option value={sport.id}>{sport.name}</option>
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
						{isVerifying ? '確認中...' : '本登録する'}
					</button>
				</form>

				{#if selectedSport}
					<p class="mt-4 text-sm text-gray-600">選択中: {selectedSport.name}</p>
				{/if}
			</aside>
		</div>
	{/if}

	{#if verificationResult}
		{#if verificationResult.success}
			<div class="mt-6 rounded border border-green-400 bg-green-100 p-4 text-green-800">
				<p class="font-bold">本登録しました</p>
				<p>氏名: {verificationResult.data.display_name || '未設定'}</p>
				<p>学籍番号: {verificationResult.data.student_number}</p>
				<p>競技: {verificationResult.data.sport_name}</p>
				{#if verificationResult.data.capacity_warning}
					<p class="mt-2 font-semibold">{verificationResult.data.capacity_warning}</p>
				{/if}
			</div>
		{:else}
			<div class="mt-6 rounded border border-red-400 bg-red-100 p-4 text-red-700">
				<p class="font-bold">本登録できませんでした</p>
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
