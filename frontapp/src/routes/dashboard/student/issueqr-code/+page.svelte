<script>
	import { onMount, onDestroy } from 'svelte';
	import { beforeNavigate, afterNavigate } from '$app/navigation';
	import QRCode from 'qrcode';
	import { browser } from '$app/environment';

	let teams = [];
	let selectedTeamId = '';
	let qrCodeData = null;

	$: selectedTeam = selectedTeamId
		? teams.find((t) => `${t.event_id}-${t.sport_id}` === selectedTeamId)
		: null;
	let qrCodeImage = null;
	let expiresAt = null;
	let remainingTime = null;
	let countdownInterval = null;
	let loading = false;
	let error = null;
	let activeEventId = null;

	// クリーンアップ関数
	function cleanup() {
		if (countdownInterval) {
			clearInterval(countdownInterval);
			countdownInterval = null;
		}
	}

	// データを初期化・再取得する関数
	async function initializeData() {
		try {
			// Get active event
			const eventResponse = await fetch('/api/events/active');
			if (!eventResponse.ok) throw new Error('Failed to get active event');
			const eventData = await eventResponse.json();
			activeEventId = eventData.event_id;

			// Get user teams
			const sessionToken = document.cookie
				.split('; ')
				.find((row) => row.startsWith('session_token='))
				?.split('=')[1];
			const headers = {
				'Content-Type': 'application/json',
				Cookie: `session_token=${sessionToken}`
			};

			const teamsResponse = await fetch('/api/qrcode/teams', { headers });
			if (!teamsResponse.ok) {
				const errorText = await teamsResponse.text();
				throw new Error(`Failed to load teams: ${errorText}`);
			}
			teams = await teamsResponse.json();

			// Filter teams by active event
			if (activeEventId) {
				teams = teams.filter((team) => team.event_id === activeEventId);
			}
		} catch (err) {
			console.error('Error loading teams:', err);
			error = err.message || 'チーム情報の取得に失敗しました';
		}
	}

	onMount(async () => {
		await initializeData();
	});

	// ページ遷移前にクリーンアップ（他のページに遷移する時）
	beforeNavigate(({ to, cancel }) => {
		// このページから他のページに遷移する時はクリーンアップ
		if (to && to.url.pathname !== '/dashboard/student/issueqr-code') {
			cleanup();
		}
	});

	// ページ遷移時にデータを再取得（このページに遷移した時のみ）
	afterNavigate(async ({ to, from }) => {
		// このページに遷移した時だけデータを再取得（他のページから遷移してきた場合のみ）
		if (to && to.url.pathname === '/dashboard/student/issueqr-code' && 
		    from && from.url.pathname !== '/dashboard/student/issueqr-code') {
			await initializeData();
		}
	});

	onDestroy(() => {
		// クリーンアップ: インターバルを確実にクリア
		cleanup();
	});

	function startCountdown(expiresAtTimestamp) {
		// 既存のインターバルをクリア
		cleanup();

		function updateCountdown() {
			const now = Math.floor(Date.now() / 1000);
			const remaining = expiresAtTimestamp - now;

			if (remaining <= 0) {
				remainingTime = '00:00';
				cleanup();
				qrCodeImage = null;
				return;
			}

			const minutes = Math.floor(remaining / 60);
			const seconds = remaining % 60;
			remainingTime = `${minutes.toString().padStart(2, '0')}:${seconds.toString().padStart(2, '0')}`;
		}

		updateCountdown();
		countdownInterval = setInterval(updateCountdown, 1000);
	}

	async function generateQRCode() {
		if (!selectedTeam) {
			error = '競技を選択してください';
			return;
		}

		loading = true;
		error = null;
		qrCodeImage = null;
		remainingTime = null;

		try {
			const sessionToken = document.cookie
				.split('; ')
				.find((row) => row.startsWith('session_token='))
				?.split('=')[1];
			const headers = {
				'Content-Type': 'application/json',
				Cookie: `session_token=${sessionToken}`
			};

			const response = await fetch('/api/qrcode/generate', {
				method: 'POST',
				headers,
				body: JSON.stringify({
					event_id: selectedTeam.event_id,
					sport_id: selectedTeam.sport_id
				})
			});

			if (!response.ok) {
				const errorData = await response.json();
				throw new Error(errorData.error || 'QRコードの生成に失敗しました');
			}

			const data = await response.json();
			qrCodeData = data.qr_code_data;
			expiresAt = data.expires_at;

			// Generate QR code image
			if (browser) {
				try {
					qrCodeImage = await QRCode.toDataURL(qrCodeData, {
						width: 300,
						margin: 2,
						color: {
							dark: '#000000',
							light: '#FFFFFF'
						}
					});
				} catch (qrError) {
					console.error('QR code generation error:', qrError);
					throw new Error('QRコード画像の生成に失敗しました');
				}
			}

			// Start countdown
			startCountdown(expiresAt);
		} catch (err) {
			console.error('Error generating QR code:', err);
			error = err.message || 'QRコードの生成に失敗しました';
		} finally {
			loading = false;
		}
	}

	function resetQRCode() {
		cleanup();
		qrCodeData = null;
		qrCodeImage = null;
		expiresAt = null;
		remainingTime = null;
		selectedTeamId = '';
		error = null;
	}
</script>

<div class="max-w-4xl mx-auto p-6">
	<h1 class="text-2xl font-bold mb-6">QRコード発行</h1>

	{#if error}
		<div class="mb-4 p-4 bg-red-100 border border-red-400 text-red-700 rounded">
			{error}
		</div>
	{/if}

	{#if teams.length === 0}
		<div class="p-6 bg-gray-100 rounded-lg text-center">
			<p class="text-gray-600">
				参加登録されている競技がありません。競技に参加登録してからQRコードを発行してください。
			</p>
		</div>
	{:else}
		<div class="space-y-6">
			<!-- Team Selection -->
			<div class="bg-white p-6 rounded-lg shadow">
				<h2 class="text-xl font-semibold mb-4">競技選択</h2>
				<select
					bind:value={selectedTeamId}
					class="w-full p-3 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
					disabled={loading || qrCodeImage !== null}
				>
					<option value="">競技を選択してください</option>
					{#each teams as team}
						<option value={`${team.event_id}-${team.sport_id}`}>{team.sport_name}</option>
					{/each}
				</select>

				{#if selectedTeam && !qrCodeImage}
					<button
						on:click={generateQRCode}
						disabled={loading}
						class="mt-4 w-full bg-blue-600 text-white py-2 px-4 rounded-md hover:bg-blue-700 disabled:bg-gray-400 disabled:cursor-not-allowed"
					>
						{loading ? '生成中...' : 'QRコードを生成'}
					</button>
				{/if}
			</div>

			<!-- QR Code Display -->
			{#if qrCodeImage}
				<div class="bg-white p-6 rounded-lg shadow text-center">
					<h2 class="text-xl font-semibold mb-4">QRコード</h2>
					<div class="mb-4">
						<img src={qrCodeImage} alt="QR Code" class="mx-auto border-4 border-gray-200 rounded-lg" />
					</div>

					{#if remainingTime}
						<div class="mb-4">
							<p class="text-sm text-gray-600">有効期限まで残り時間:</p>
							<p class="text-2xl font-bold {remainingTime === '00:00' ? 'text-red-600' : 'text-blue-600'}">
								{remainingTime}
							</p>
						</div>
					{/if}

					{#if selectedTeam}
						<div class="mb-4 text-sm text-gray-600">
							<p>競技: {selectedTeam.sport_name}</p>
							<p>イベントID: {selectedTeam.event_id}</p>
						</div>
					{/if}

					<button
						on:click={resetQRCode}
						class="mt-4 bg-gray-600 text-white py-2 px-4 rounded-md hover:bg-gray-700"
					>
						新しいQRコードを発行
					</button>
				</div>
			{/if}

			<!-- Info Box -->
			<div class="bg-blue-50 border border-blue-200 rounded-lg p-4">
				<h3 class="font-semibold text-blue-900 mb-2">ご注意</h3>
				<ul class="text-sm text-blue-800 space-y-1 list-disc list-inside">
					<li>QRコードは発行から5分間のみ有効です</li>
					<li>有効期限が切れたQRコードは使用できません</li>
					<li>事前に競技割り当てされた方のみQRコードを発行できます</li>
					<li>QRコードにはイベントID、競技名、表示名などの情報が含まれます</li>
				</ul>
			</div>
		</div>
	{/if}
</div>

<style>
	:global(body) {
		background-color: #f3f4f6;
	}
</style>
