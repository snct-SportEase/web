<script>
	import { afterUpdate, onDestroy, onMount } from 'svelte';
	import Chart from 'chart.js/auto';

	export let data;

	let scoreBreakdownChart;

	$: user = data.user;
	$: myClassScore = data.myClassScore;
	$: errorMessage = data.error;
	$: upcomingMatches = data.upcomingMatches || [];
	$: scoreItems = data.scoreItems || [];
	$: categoryBreakdown = data.categoryBreakdown || [];
	$: pointHighlights = data.pointHighlights || [];
	$: sportSections = data.sportSections || [];
	$: hasScore = Boolean(myClassScore);
	$: primaryRankLabel = myClassScore
		? myClassScore.season === 'spring'
			? '現在の順位'
			: '総合順位'
		: '';
	$: secondaryRankLabel = myClassScore
		? myClassScore.season === 'spring'
			? '総合順位'
			: '現在の順位'
		: '';

	function createChart(ctx, labels, values) {
		scoreBreakdownChart = new Chart(ctx, {
			type: 'bar',
			data: {
				labels,
				datasets: [
					{
						label: '獲得ポイント',
						data: values,
						backgroundColor: [
							'rgba(79, 70, 229, 0.85)',
							'rgba(37, 99, 235, 0.85)',
							'rgba(16, 185, 129, 0.85)',
							'rgba(251, 191, 36, 0.85)',
							'rgba(245, 158, 11, 0.85)',
							'rgba(244, 63, 94, 0.85)'
						],
						borderRadius: 12,
						barThickness: 32
					}
				]
			},
			options: {
				responsive: true,
				maintainAspectRatio: false,
				scales: {
					x: {
						grid: {
							display: false
						}
					},
					y: {
						beginAtZero: true,
						grid: {
							color: 'rgba(148, 163, 184, 0.2)'
						},
						ticks: {
							stepSize: 5
						}
					}
				},
				plugins: {
					legend: {
						display: false
					},
					tooltip: {
						callbacks: {
							label: (ctx) => `${ctx.parsed.y} 点`
						}
					}
				}
			}
		});
	}

	function syncChart() {
		if (typeof document === 'undefined') {
			return;
		}

		const labels = categoryBreakdown.map((item) => item.label);
		const values = categoryBreakdown.map((item) => item.value);

		const canvas = document.getElementById('scoreBreakdownChart');
		if (!canvas) {
			return;
		}

		if (!categoryBreakdown.length) {
			if (scoreBreakdownChart) {
				scoreBreakdownChart.destroy();
				scoreBreakdownChart = null;
			}
			return;
		}

		if (!scoreBreakdownChart) {
			createChart(canvas, labels, values);
			return;
		}

		scoreBreakdownChart.data.labels = labels;
		scoreBreakdownChart.data.datasets[0].data = values;
		scoreBreakdownChart.update();
	}

	onMount(() => {
		syncChart();
	});

	afterUpdate(() => {
		syncChart();
	});

	onDestroy(() => {
		scoreBreakdownChart?.destroy();
	});
</script>

<div class="container mx-auto space-y-8 p-4 md:p-6 lg:p-8 max-w-7xl">
	<div class="space-y-2 text-center mb-8">
		<h1 class="text-3xl md:text-4xl font-bold tracking-tight text-gray-900">マイページ</h1>
		<p class="text-base text-gray-600">クラスの現状とポイント内訳をまとめて確認できます</p>
	</div>

	{#if errorMessage && !hasScore}
		<div class="bg-red-50 border-l-4 border-red-500 rounded-lg p-4 flex items-start gap-3 shadow-sm">
			<svg
				xmlns="http://www.w3.org/2000/svg"
				class="h-6 w-6 text-red-500 flex-shrink-0 mt-0.5"
				fill="none"
				viewBox="0 0 24 24"
				stroke="currentColor"
			>
				<path
					stroke-linecap="round"
					stroke-linejoin="round"
					stroke-width="2"
					d="M10 14l2-2m0 0l2-2m-2 2l-2-2m2 2l2 2m7-2a9 9 0 11-18 0 9 9 0 0118 0z"
				/>
			</svg>
			<span class="text-red-800 font-medium">{errorMessage}</span>
		</div>
	{:else if user && hasScore}
		<div class="space-y-8">
			<section class="grid grid-cols-1 gap-6 lg:grid-cols-3">
				<!-- プロフィールカード -->
				<div class="bg-white rounded-xl border border-gray-200 shadow-lg p-6">
					<h2 class="flex items-center gap-2 text-lg font-semibold text-gray-900 mb-4">
						<svg xmlns="http://www.w3.org/2000/svg" class="h-6 w-6 text-indigo-600" fill="none" viewBox="0 0 24 24" stroke="currentColor">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z" />
						</svg>
						プロフィール
					</h2>
					<div class="space-y-4">
						<div class="flex items-center justify-between py-2 border-b border-gray-100">
							<span class="text-sm text-gray-600">表示名</span>
							<span class="px-3 py-1 rounded-full text-sm font-medium bg-gray-100 text-gray-800">{user.display_name || '未設定'}</span>
						</div>
						<div class="flex items-center justify-between py-2 border-b border-gray-100">
							<span class="text-sm text-gray-600">クラス</span>
							<span class="px-3 py-1 rounded-full text-sm font-medium bg-indigo-100 text-indigo-800">{myClassScore.class_name}</span>
						</div>
						<div class="flex items-center justify-between py-2">
							<span class="text-sm text-gray-600">メールアドレス</span>
							<span class="text-sm font-medium text-gray-900">{user.email}</span>
						</div>
					</div>
				</div>

				<!-- クラス成績カード -->
				<div class="bg-gradient-to-br from-indigo-600 to-indigo-700 rounded-xl border border-indigo-500 shadow-xl p-6 text-white">
					<h2 class="flex items-center gap-2 text-lg font-semibold mb-4">
						<svg xmlns="http://www.w3.org/2000/svg" class="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 13l4 4L19 7" />
						</svg>
						クラス成績
					</h2>
					<div class="grid grid-cols-1 gap-4">
						<div class="rounded-xl bg-white/20 backdrop-blur-sm p-4 border border-white/30">
							<p class="text-sm text-white/80 mb-1">{primaryRankLabel}</p>
							<p class="text-4xl font-bold tracking-tight mb-1">{myClassScore.primaryRank}位</p>
							<p class="text-sm text-white/90">獲得 {myClassScore.primaryPoints} 点</p>
						</div>
						{#if myClassScore.secondaryRank}
							<div class="rounded-xl bg-white/15 backdrop-blur-sm p-4 border border-white/20">
								<p class="text-sm text-white/80 mb-1">{secondaryRankLabel}</p>
								<p class="text-2xl font-semibold mb-1">{myClassScore.secondaryRank}位</p>
								<p class="text-sm text-white/90">累計 {myClassScore.secondaryPoints} 点</p>
							</div>
						{/if}
					</div>
				</div>

				<!-- 得点ハイライトカード -->
				<div class="bg-white rounded-xl border border-gray-200 shadow-lg p-6">
					<h2 class="flex items-center gap-2 text-lg font-semibold text-gray-900 mb-4">
						<svg xmlns="http://www.w3.org/2000/svg" class="h-6 w-6 text-indigo-600" fill="none" viewBox="0 0 24 24" stroke="currentColor">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 16h-1v-4h-1m1-4h.01M12 2a10 10 0 100 20 10 10 0 000-20z" />
						</svg>
						得点ハイライト
					</h2>
					<div class="space-y-3">
						{#if pointHighlights.length === 0}
							<p class="text-sm text-gray-500 text-center py-4">表示できる得点データがありません。</p>
						{:else}
							{#each pointHighlights as highlight, index}
								<div class="flex items-center justify-between rounded-lg border border-gray-200 bg-gray-50 p-3 hover:bg-gray-100 transition-colors">
									<div class="flex items-center gap-3">
										<div class="flex h-8 w-8 items-center justify-center rounded-full bg-indigo-100 text-sm font-semibold text-indigo-700">
											#{index + 1}
										</div>
										<span class="font-medium text-gray-900">{highlight.label}</span>
									</div>
									<span class="text-base font-semibold text-indigo-600">{highlight.value} 点</span>
								</div>
							{/each}
						{/if}
					</div>
				</div>
			</section>

			<section class="grid grid-cols-1 gap-6 xl:grid-cols-3">
				<!-- 獲得ポイント内訳 -->
				<div class="bg-white rounded-xl border border-gray-200 shadow-lg p-6 xl:col-span-2">
					<h2 class="flex items-center gap-2 text-lg font-semibold text-gray-900 mb-4">
						<svg xmlns="http://www.w3.org/2000/svg" class="h-6 w-6 text-indigo-600" fill="none" viewBox="0 0 24 24" stroke="currentColor">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 17v-6h13M5 17l-3-3m0 0l3-3m-3 3h16" />
						</svg>
						獲得ポイント内訳
					</h2>
					<div class="overflow-hidden rounded-lg border border-gray-200">
						<table class="w-full">
							<thead>
								<tr class="bg-gray-50 border-b border-gray-200">
									<th class="px-4 py-3 text-left text-sm font-semibold text-gray-700">項目</th>
									<th class="px-4 py-3 text-right text-sm font-semibold text-gray-700">獲得点</th>
								</tr>
							</thead>
							<tbody>
								{#if scoreItems.length === 0}
									<tr>
										<td colspan="2" class="py-8 text-center text-sm text-gray-500">
											得点データが登録されていません。
										</td>
									</tr>
								{:else}
									{#each scoreItems as item, index}
										<tr class="border-b border-gray-100 hover:bg-gray-50 transition-colors {index % 2 === 0 ? 'bg-white' : 'bg-gray-50/50'}">
											<td class="px-4 py-3 font-medium text-gray-900">{item.label}</td>
											<td class="px-4 py-3 text-right font-semibold text-indigo-600">{item.value} 点</td>
										</tr>
									{/each}
								{/if}
							</tbody>
						</table>
					</div>
				</div>

				<!-- カテゴリ別の貢献度 -->
				{#if categoryBreakdown.length > 0}
					<div class="bg-white rounded-xl border border-gray-200 shadow-lg p-6">
						<h2 class="flex items-center gap-2 text-lg font-semibold text-gray-900 mb-4">
							<svg xmlns="http://www.w3.org/2000/svg" class="h-6 w-6 text-indigo-600" fill="none" viewBox="0 0 24 24" stroke="currentColor">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M11 3.055A9.001 9.001 0 1020.945 13H11V3.055z" />
							</svg>
							カテゴリ別の貢献度
						</h2>
						<div class="h-72">
							<canvas id="scoreBreakdownChart" aria-label="カテゴリ別得点チャート"></canvas>
						</div>
					</div>
				{/if}
			</section>

			{#if sportSections.length > 0}
				<section class="bg-white rounded-xl border border-gray-200 shadow-lg p-6">
					<h2 class="flex items-center gap-2 text-lg font-semibold text-gray-900 mb-4">
						<svg xmlns="http://www.w3.org/2000/svg" class="h-6 w-6 text-indigo-600" fill="none" viewBox="0 0 24 24" stroke="currentColor">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M18 9l-6 6-6-6" />
						</svg>
						種目別の詳細
					</h2>
					<div class="grid grid-cols-1 gap-4 md:grid-cols-2 lg:grid-cols-3">
						{#each sportSections as section}
							<div class="rounded-xl border border-gray-200 bg-gray-50 p-5 shadow-sm hover:shadow-md transition-shadow">
								<div class="flex items-center justify-between mb-4">
									<h3 class="text-lg font-semibold text-gray-900">{section.label}</h3>
									<span class="px-3 py-1 rounded-full text-sm font-semibold bg-indigo-100 text-indigo-800 border border-indigo-200">{section.total} 点</span>
								</div>
								<div class="border-t border-gray-200 my-3"></div>
								<ul class="space-y-2 text-sm">
									{#each section.entries as entry}
										<li class="flex items-center justify-between py-1">
											<span class="text-gray-600">{entry.label}</span>
											<span class="font-semibold text-gray-900">{entry.value} 点</span>
										</li>
									{/each}
								</ul>
							</div>
						{/each}
					</div>
				</section>
			{/if}

			<!-- 今後の試合予定 -->
			<section class="bg-white rounded-xl border border-gray-200 shadow-lg p-6">
				<h2 class="flex items-center gap-2 text-lg font-semibold text-gray-900 mb-4">
					<svg xmlns="http://www.w3.org/2000/svg" class="h-6 w-6 text-indigo-600" fill="none" viewBox="0 0 24 24" stroke="currentColor">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 7V3m8 4V3m-9 8h10M5 21h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z" />
					</svg>
					今後の試合予定
				</h2>
				{#if upcomingMatches.length > 0}
					<div class="overflow-x-auto rounded-lg border border-gray-200">
						<table class="w-full">
							<thead>
								<tr class="bg-gray-50 border-b border-gray-200">
									<th class="px-4 py-3 text-left text-sm font-semibold text-gray-700">日時</th>
									<th class="px-4 py-3 text-left text-sm font-semibold text-gray-700">種目</th>
									<th class="px-4 py-3 text-left text-sm font-semibold text-gray-700">対戦相手</th>
									<th class="px-4 py-3 text-left text-sm font-semibold text-gray-700">場所</th>
								</tr>
							</thead>
							<tbody>
								{#each upcomingMatches as match, index}
									<tr class="border-b border-gray-100 hover:bg-gray-50 transition-colors {index % 2 === 0 ? 'bg-white' : 'bg-gray-50/50'}">
										<td class="px-4 py-3 text-sm text-gray-900">{new Date(match.start_time).toLocaleString('ja-JP')}</td>
										<td class="px-4 py-3 text-sm text-gray-900">{match.sport_name}</td>
										<td class="px-4 py-3 text-sm text-gray-900">{match.opponent_name}</td>
										<td class="px-4 py-3 text-sm text-gray-900">{match.location}</td>
									</tr>
								{/each}
							</tbody>
						</table>
					</div>
				{:else}
					<div class="bg-blue-50 border-l-4 border-blue-500 rounded-lg p-4 flex items-start gap-3">
						<svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" class="h-6 w-6 text-blue-500 flex-shrink-0 mt-0.5" stroke="currentColor">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"></path>
						</svg>
						<span class="text-blue-800 font-medium">現在、予定されている試合はありません。</span>
					</div>
				{/if}
			</section>
		</div>
	{:else}
		<div class="flex flex-col items-center justify-center gap-4 py-20">
			<div class="animate-spin rounded-full h-12 w-12 border-4 border-gray-200 border-t-indigo-600"></div>
			<p class="text-gray-600 font-medium">マイページ情報を読み込んでいます...</p>
		</div>
	{/if}
</div>

