<script>
	import { onMount } from 'svelte';
	import { Html5Qrcode } from 'html5-qrcode';

	let html5QrCode;
	let cameraId;
	let scannedData = '';
	let errorMessage = '';

	let verificationResult = null;
	let remainingTime = null;
	let countdownInterval = null;

	onMount(() => {
		Html5Qrcode.getCameras()
			.then((devices) => {
				if (devices && devices.length) {
					cameraId = devices[0].id;
				}
			})
			.catch((err) => {
				errorMessage = `Error getting cameras: ${err}`;
			});
		
		return () => {
			if (countdownInterval) {
				clearInterval(countdownInterval);
			}
		};
	});

	function startCountdown(expiresAtTimestamp) {
		if (countdownInterval) {
			clearInterval(countdownInterval);
		}

		function updateCountdown() {
			const now = Math.floor(Date.now() / 1000);
			const remaining = expiresAtTimestamp - now;

			if (remaining <= 0) {
				remainingTime = 'Expired';
				clearInterval(countdownInterval);
				return;
			}

			const minutes = Math.floor(remaining / 60);
			const seconds = remaining % 60;
			remainingTime = `${minutes.toString().padStart(2, '0')}:${seconds.toString().padStart(2, '0')}`;
		}

		updateCountdown();
		countdownInterval = setInterval(updateCountdown, 1000);
	}

	async function verifyQRCode(qrCodeData) {
		try {
			const response = await fetch('/api/qrcode/verify', {
				method: 'POST',
				headers: {
					'Content-Type': 'application/json'
				},
				body: JSON.stringify({ qr_code_data: qrCodeData })
			});

			const result = await response.json();

			if (response.ok) {
				verificationResult = { success: true, data: result };
				startCountdown(result.expires_at);
			} else {
				verificationResult = { success: false, message: result.error };
			}
		} catch {
			verificationResult = { success: false, message: 'An error occurred while verifying the QR code.' };
		}
	}

	function startScan() {
		if (!cameraId) {
			errorMessage = 'No camera found.';
			return;
		}

		html5QrCode = new Html5Qrcode('qr-code-full-region');
		html5QrCode
			.start(
				cameraId,
				{
					fps: 10,
					qrbox: { width: 250, height: 250 }
				},
				(decodedText) => {
					scannedData = decodedText;
					verifyQRCode(decodedText);
					stopScan();
				},
				() => {
					// console.warn(`QR code scanning error: ${error}`);
				}
			)
			.catch((err) => {
				errorMessage = `Error starting scanner: ${err}`;
			});
	}

	function stopScan() {
		if (html5QrCode && html5QrCode.isScanning) {
			html5QrCode
				.stop()
				.then(() => {
					console.log('QR Code scanning stopped.');
				})
				.catch((err) => {
					console.error(`Error stopping scanner: ${err}`);
				});
		}
	}

	function resetScanner() {
		scannedData = '';
		verificationResult = null;
		errorMessage = '';
		if (countdownInterval) {
			clearInterval(countdownInterval);
		}
		remainingTime = null;
	}
</script>

<div class="container mx-auto p-4">
	<h1 class="text-2xl font-bold mb-4">QR Code Scanner</h1>

	<div id="qr-code-full-region" class="w-full md:w-1/2 lg:w-1/3 mx-auto border-2 border-gray-300 rounded-lg overflow-hidden"></div>

	<div class="mt-4 flex justify-center space-x-4">
		<button on:click={startScan} class="bg-blue-500 hover:bg-blue-700 text-white font-bold py-2 px-4 rounded">
			Start Scan
		</button>
		<button on:click={stopScan} class="bg-red-500 hover:bg-red-700 text-white font-bold py-2 px-4 rounded">
			Stop Scan
		</button>
		<button on:click={resetScanner} class="bg-gray-500 hover:bg-gray-700 text-white font-bold py-2 px-4 rounded">
			Reset
		</button>
	</div>

	{#if scannedData}
		<div class="mt-4 p-4 bg-gray-100 border border-gray-400 text-gray-700 rounded">
			<p class="font-bold">Scanned QR Code:</p>
			<p class="break-all">{scannedData}</p>
		</div>
	{/if}

	{#if verificationResult}
		{#if verificationResult.success}
			<div class="mt-4 p-4 bg-green-100 border border-green-400 text-green-700 rounded">
				<p class="font-bold">Verification Successful!</p>
				<p>User: {verificationResult.data.display_name}</p>
				<p>Sport: {verificationResult.data.sport_name}</p>
				{#if remainingTime}
					<p>Remaining Time: {remainingTime}</p>
				{/if}
			</div>
		{:else}
			<div class="mt-4 p-4 bg-red-100 border border-red-400 text-red-700 rounded">
				<p class="font-bold">Verification Failed!</p>
				<p>{verificationResult.message}</p>
			</div>
		{/if}
	{/if}

	{#if errorMessage}
		<div class="mt-4 p-4 bg-red-100 border border-red-400 text-red-700 rounded">
			<p class="font-bold">Error:</p>
			<p>{errorMessage}</p>
		</div>
	{/if}
</div>
