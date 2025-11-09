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

<div class="container mx-auto space-y-8 p-6">
	<div class="space-y-2 text-center">
		<h1 class="text-3xl font-bold tracking-tight">マイページ</h1>
		<p class="text-base-content/70">クラスの現状とポイント内訳をまとめて確認できます</p>
	</div>

	{#if errorMessage && !hasScore}
		<div class="alert alert-error">
			<svg
				xmlns="http://www.w3.org/2000/svg"
				class="stroke-current h-6 w-6 flex-shrink-0"
				fill="none"
				viewBox="0 0 24 24"
				><path
					stroke-linecap="round"
					stroke-linejoin="round"
					stroke-width="2"
					d="M10 14l2-2m0 0l2-2m-2 2l-2-2m2 2l2 2m7-2a9 9 0 11-18 0 9 9 0 0118 0z"
				/></svg
			>
			<span>{errorMessage}</span>
		</div>
	{:else if user && hasScore}
		<div class="space-y-8">
			<section class="grid grid-cols-1 gap-6 lg:grid-cols-3">
				<div class="card border border-base-200 bg-base-100 shadow-xl">
					<div class="card-body space-y-4">
						<h2 class="card-title flex items-center gap-2 text-lg font-semibold">
							<svg xmlns="http://www.w3.org/2000/svg" class="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z" />
							</svg>
							プロフィール
						</h2>
						<div class="space-y-3">
							<div class="flex items-center justify-between">
								<span class="text-sm text-base-content/70">表示名</span>
								<span class="badge badge-lg badge-neutral">{user.display_name || '未設定'}</span>
							</div>
							<div class="flex items-center justify-between">
								<span class="text-sm text-base-content/70">クラス</span>
								<span class="badge badge-lg badge-primary">{myClassScore.class_name}</span>
							</div>
							<div class="flex items-center justify-between">
								<span class="text-sm text-base-content/70">メールアドレス</span>
								<span class="text-sm font-medium">{user.email}</span>
							</div>
						</div>
					</div>
				</div>

				<div class="card border border-primary/30 bg-gradient-to-br from-primary to-primary/90 text-primary-content shadow-xl">
					<div class="card-body space-y-4">
						<h2 class="card-title flex items-center gap-2 text-lg">
							<svg xmlns="http://www.w3.org/2000/svg" class="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 13l4 4L19 7" />
							</svg>
							クラス成績
						</h2>
						<div class="grid grid-cols-1 gap-4">
							<div class="rounded-2xl bg-primary/30 p-4 backdrop-blur">
								<p class="text-sm opacity-70">{primaryRankLabel}</p>
								<p class="text-4xl font-bold tracking-tight">{myClassScore.primaryRank}位</p>
								<p class="mt-1 text-sm opacity-80">獲得 {myClassScore.primaryPoints} 点</p>
							</div>
							{#if myClassScore.secondaryRank}
								<div class="rounded-2xl bg-primary/20 p-4 backdrop-blur">
									<p class="text-sm opacity-70">{secondaryRankLabel}</p>
									<p class="text-2xl font-semibold">{myClassScore.secondaryRank}位</p>
									<p class="mt-1 text-sm opacity-80">累計 {myClassScore.secondaryPoints} 点</p>
								</div>
							{/if}
						</div>
					</div>
				</div>

				<div class="card border border-base-200 bg-base-100 shadow-xl">
					<div class="card-body space-y-4">
						<h2 class="card-title flex items-center gap-2 text-lg font-semibold">
							<svg xmlns="http://www.w3.org/2000/svg" class="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 16h-1v-4h-1m1-4h.01M12 2a10 10 0 100 20 10 10 0 000-20z" />
							</svg>
							得点ハイライト
						</h2>
						<div class="space-y-3">
							{#if pointHighlights.length === 0}
								<p class="text-sm text-base-content/70">表示できる得点データがありません。</p>
							{:else}
								{#each pointHighlights as highlight, index}
									<div class="flex items-center justify-between rounded-xl border border-base-200 bg-base-100 p-3 shadow-sm">
										<div class="flex items-center gap-3">
											<div class="flex h-8 w-8 items-center justify-center rounded-full bg-primary/10 text-sm font-semibold text-primary">
												#{index + 1}
											</div>
											<span class="font-medium">{highlight.label}</span>
										</div>
										<span class="text-base font-semibold">{highlight.value} 点</span>
									</div>
								{/each}
							{/if}
						</div>
					</div>
				</div>
			</section>

			<section class="grid grid-cols-1 gap-6 xl:grid-cols-3">
				<div class="card border border-base-200 bg-base-100 shadow-xl xl:col-span-2">
					<div class="card-body space-y-4">
						<h2 class="card-title flex items-center gap-2 text-lg font-semibold">
							<svg xmlns="http://www.w3.org/2000/svg" class="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 17v-6h13M5 17l-3-3m0 0l3-3m-3 3h16" />
							</svg>
							獲得ポイント内訳
						</h2>
						<div class="overflow-hidden rounded-xl border border-base-200">
							<table class="table table-zebra">
								<thead>
									<tr class="bg-base-200 text-base-content/80">
										<th>項目</th>
										<th class="text-right">獲得点</th>
									</tr>
								</thead>
								<tbody>
									{#if scoreItems.length === 0}
										<tr>
											<td colspan="2" class="py-6 text-center text-sm text-base-content/70">
												得点データが登録されていません。
											</td>
										</tr>
									{:else}
										{#each scoreItems as item}
											<tr>
												<td class="font-medium">{item.label}</td>
												<td class="text-right font-semibold">{item.value} 点</td>
											</tr>
										{/each}
									{/if}
								</tbody>
							</table>
						</div>
					</div>
				</div>

				{#if categoryBreakdown.length > 0}
					<div class="card border border-base-200 bg-base-100 shadow-xl">
						<div class="card-body space-y-4">
							<h2 class="card-title flex items-center gap-2 text-lg font-semibold">
								<svg xmlns="http://www.w3.org/2000/svg" class="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
									<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M11 3.055A9.001 9.001 0 1020.945 13H11V3.055z" />
								</svg>
								カテゴリ別の貢献度
							</h2>
							<div class="h-72">
								<canvas id="scoreBreakdownChart" aria-label="カテゴリ別得点チャート"></canvas>
							</div>
						</div>
					</div>
				{/if}
			</section>

			{#if sportSections.length > 0}
				<section class="card border border-base-200 bg-base-100 shadow-xl">
					<div class="card-body space-y-4">
						<h2 class="card-title flex items-center gap-2 text-lg font-semibold">
							<svg xmlns="http://www.w3.org/2000/svg" class="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M18 9l-6 6-6-6" />
							</svg>
							種目別の詳細
						</h2>
						<div class="grid grid-cols-1 gap-4 md:grid-cols-2 lg:grid-cols-3">
							{#each sportSections as section}
								<div class="rounded-2xl border border-base-200 bg-base-100 p-5 shadow-sm">
									<div class="flex items-center justify-between">
										<h3 class="text-lg font-semibold">{section.label}</h3>
										<span class="badge badge-outline badge-lg">{section.total} 点</span>
									</div>
									<div class="divider my-3"></div>
									<ul class="space-y-2 text-sm">
										{#each section.entries as entry}
											<li class="flex items-center justify-between">
												<span class="text-base-content/70">{entry.label}</span>
												<span class="font-semibold">{entry.value} 点</span>
											</li>
										{/each}
									</ul>
								</div>
							{/each}
						</div>
					</div>
				</section>
			{/if}

			<section class="card border border-base-200 bg-base-100 shadow-xl">
				<div class="card-body space-y-4">
					<h2 class="card-title flex items-center gap-2 text-lg font-semibold">
						<svg xmlns="http://www.w3.org/2000/svg" class="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 7V3m8 4V3m-9 8h10M5 21h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z" />
						</svg>
						今後の試合予定
					</h2>
					{#if upcomingMatches.length > 0}
						<div class="overflow-x-auto rounded-xl border border-base-200">
							<table class="table table-zebra">
								<thead>
									<tr class="bg-base-200 text-base-content/80">
										<th>日時</th>
										<th>種目</th>
										<th>対戦相手</th>
										<th>場所</th>
									</tr>
								</thead>
								<tbody>
									{#each upcomingMatches as match}
										<tr>
											<td>{new Date(match.start_time).toLocaleString('ja-JP')}</td>
											<td>{match.sport_name}</td>
											<td>{match.opponent_name}</td>
											<td>{match.location}</td>
										</tr>
									{/each}
								</tbody>
							</table>
						</div>
					{:else}
						<div class="alert alert-info">
							<svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" class="stroke-current h-6 w-6 shrink-0">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"></path>
							</svg>
							<span>現在、予定されている試合はありません。</span>
						</div>
					{/if}
				</div>
			</section>
		</div>
	{:else}
		<div class="flex flex-col items-center justify-center gap-4 py-20">
			<span class="loading loading-lg loading-spinner text-primary"></span>
			<p class="text-base-content/70">マイページ情報を読み込んでいます...</p>
		</div>
	{/if}
</div>

