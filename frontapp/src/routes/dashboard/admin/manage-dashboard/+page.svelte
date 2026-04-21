<script>
	import { onMount, onDestroy } from 'svelte';
	import Chart from 'chart.js/auto';

	let attendanceRate = $state(null);
	let attendanceError = $state('');
	let participationRates = $state({});
	let scoreTrends = $state({});
	let eventProgress = $state({});

	let hideScores = $state(false);

	let ws;

	onMount(async () => {
		const token = localStorage.getItem('token');
		const headers = token ? { 'Authorization': `Bearer ${token}` } : {};

		try {
			const [attendance, participation, scores, progress, activeEvent] = await Promise.allSettled([
				fetchJson('/api/admin/statistics/attendance', { headers }),
				fetchJson('/api/admin/statistics/participation', { headers }),
				fetchJson('/api/admin/statistics/scores', { headers }),
				fetchJson('/api/admin/statistics/progress', { headers }),
				fetchJson('/api/events/active', { headers })
			]);

			if (attendance.status === 'fulfilled') {
				const rate = Number(attendance.value.attendance_rate);
				if (Number.isFinite(rate)) {
					attendanceRate = rate;
					attendanceError = '';
				} else {
					attendanceRate = null;
					attendanceError = '出席率を取得できませんでした。';
				}
			} else {
				console.error(attendance.reason);
				attendanceRate = null;
				attendanceError = '出席率を取得できませんでした。';
			}

			if (participation.status === 'fulfilled') {
				participationRates = participation.value;
			} else {
				console.error(participation.reason);
			}

			if (scores.status === 'fulfilled' && !scores.value.message) {
				scoreTrends = scores.value;
			} else if (scores.status === 'rejected') {
				console.error(scores.reason);
			}

			if (progress.status === 'fulfilled') {
				eventProgress = progress.value;
			} else {
				console.error(progress.reason);
			}

			if (activeEvent.status === 'fulfilled') {
				hideScores = Boolean(activeEvent.value.hide_scores);
			} else {
				console.error(activeEvent.reason);
			}

			// グラフを描画
			setTimeout(() => {
				if (!hideScores) {
					drawCharts();
				}
			}, 100); // DOMがレンダリングされるのを待つ

			// WebSocket接続は現在のオリジンに追従させる
			const wsProtocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
			ws = new WebSocket(`${wsProtocol}//${window.location.host}/api/ws/progress`);
			ws.onopen = () => {
				console.log('WebSocket connected');
			};
			ws.onmessage = (event) => {
				eventProgress = JSON.parse(event.data);
			};
			ws.onclose = () => {
				console.log('WebSocket closed');
			};
		} catch (e) {
			console.error(e);
		}
	});

	onDestroy(() => {
		if (ws) ws.close();
	});

	function drawCharts() {
		// 参加率のグラフ
		const ctx = document.getElementById('participationChart');
		if (ctx) {
			new Chart(ctx, {
				type: 'bar',
				data: {
					labels: Object.keys(participationRates),
					datasets: [{
						label: '参加率 (%)',
						data: Object.values(participationRates),
						backgroundColor: 'rgba(54, 162, 235, 0.6)',
						borderColor: 'rgba(54, 162, 235, 1)',
						borderWidth: 1
					}]
				},
				options: {
					responsive: true,
					maintainAspectRatio: false,
					scales: {
						y: {
							beginAtZero: true,
							max: 100
						}
					},
					plugins: {
						legend: {
							display: false
						}
					}
				}
			});
		}

		// スコア推移のグラフ、クラスごとに
		const classScores = {};
		Object.keys(scoreTrends).forEach(event => {
			if (!Array.isArray(scoreTrends[event])) return;
			scoreTrends[event].forEach(score => {
				if (!classScores[score.class_name]) classScores[score.class_name] = [];
				classScores[score.class_name].push({ event, score: score.total_points_current_event });
			});
		});

		const ctx2 = document.getElementById('scoreChart');
		if (ctx2) {
			const datasets = Object.keys(classScores).map(className => ({
				label: className,
				data: classScores[className].map(d => d.score),
				borderColor: getRandomColor(),
				backgroundColor: getRandomColor(0.2),
				fill: false,
				tension: 0.1
			}));

			new Chart(ctx2, {
				type: 'line',
				data: {
					labels: Object.keys(scoreTrends),
					datasets
				},
				options: {
					responsive: true,
					maintainAspectRatio: false,
					scales: {
						y: {
							beginAtZero: true
						}
					},
					plugins: {
						legend: {
							position: 'bottom'
						}
					}
				}
			});
		}
	}

	function getRandomColor(alpha = 1) {
		const colors = [
			'rgba(255, 99, 132, ' + alpha + ')',
			'rgba(54, 162, 235, ' + alpha + ')',
			'rgba(255, 205, 86, ' + alpha + ')',
			'rgba(75, 192, 192, ' + alpha + ')',
			'rgba(153, 102, 255, ' + alpha + ')',
			'rgba(255, 159, 64, ' + alpha + ')'
		];
		return colors[Math.floor(Math.random() * colors.length)];
	}

	async function fetchJson(url, options) {
		const response = await fetch(url, options);
		if (!response.ok) {
			throw new Error(`${url} failed with status ${response.status}`);
		}
		return response.json();
	}

</script>

<div class="min-h-screen bg-gray-50 p-6">
	<div class="max-w-7xl mx-auto">
		<h1 class="text-3xl font-bold text-gray-900 mb-8">管理者ダッシュボード</h1>

		<!-- 全体出席率 -->
		<div class="bg-white rounded-lg shadow-md p-6 mb-6">
			<h2 class="text-xl font-semibold text-gray-800 mb-4">全体出席率</h2>
			{#if attendanceRate !== null}
				<div class="text-4xl font-bold text-blue-600">{attendanceRate.toFixed(2)}%</div>
			{:else}
				<div class="text-base font-medium text-gray-600">{attendanceError || '出席率を読み込み中です。'}</div>
			{/if}
		</div>

		<!-- グリッドレイアウト -->
		<div class="grid grid-cols-1 lg:grid-cols-2 gap-6 mb-6">
			<!-- 競技ごとの参加率 -->
			<div class="bg-white rounded-lg shadow-md p-6">
				<h2 class="text-xl font-semibold text-gray-800 mb-4">競技ごとの参加率</h2>
				<div class="h-64">
					<canvas id="participationChart"></canvas>
				</div>
			</div>

			<!-- クラス別スコア推移 -->
			{#if !hideScores}
			<div class="bg-white rounded-lg shadow-md p-6">
				<h2 class="text-xl font-semibold text-gray-800 mb-4">クラス別スコア推移</h2>
				<div class="h-64">
					<canvas id="scoreChart"></canvas>
				</div>
			</div>
			{:else}
			<div class="bg-white rounded-lg shadow-md p-6">
				<h2 class="text-xl font-semibold text-gray-800 mb-4">クラス別スコア推移</h2>
				<p class="text-gray-600">スコアは現在非表示に設定されています。</p>
			</div>
			{/if}
		</div>

		<!-- リアルタイムイベント進行状況 -->
		<div class="bg-white rounded-lg shadow-md p-6">
			<h2 class="text-xl font-semibold text-gray-800 mb-4">リアルタイムイベント進行状況</h2>
			<div class="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3 lg:grid-cols-4 gap-4">
				{#each Object.entries(eventProgress) as [sport, status] (sport)}
					<div class="bg-gray-50 rounded-lg p-4 border border-gray-200">
						<h3 class="font-medium text-gray-900">{sport}</h3>
						<p class="text-sm text-gray-600 mt-1">{status}</p>
					</div>
				{/each}
			</div>
		</div>
	</div>
</div>
