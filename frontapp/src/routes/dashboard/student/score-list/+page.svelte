<script>
	/** @type {import('./$types').PageData} */
	export let data;

	$: scores = data.scores || [];
	$: season = scores.length > 0 ? scores[0].season : '';

	const springHeaders = [
		{ key: 'class_name', label: 'クラス' },
		{ key: 'initial_points', label: '初期点' },
		{ key: 'survey_points', label: 'アンケート' },
		{ key: 'attendance_points', label: '出席点' },
		{ key: 'gym1_win1_points', label: '体育館1 1勝' },
		{ key: 'gym1_win2_points', label: '体育館1 2勝' },
		{ key: 'gym1_win3_points', label: '体育館1 3勝' },
		{ key: 'gym1_champion_points', label: '体育館1 優勝' },
		{ key: 'gym2_win1_points', label: '体育館2 1勝' },
		{ key: 'gym2_win2_points', label: '体育館2 2勝' },
		{ key: 'gym2_win3_points', label: '体育館2 3勝' },
		{ key: 'gym2_champion_points', label: '体育館2 優勝' },
		{ key: 'ground_win1_points', label: 'グラウンド1 1勝' },
		{ key: 'ground_win2_points', label: 'グラウンド1 2勝' },
		{ key: 'ground_win3_points', label: 'グラウンド1 3勝' },
		{ key: 'ground_champion_points', label: 'グラウンド1 優勝' },
		{ key: 'noon_game_points', label: '昼企画' },
		{ key: 'total_points_current_event', label: '合計' },
		{ key: 'rank_current_event', label: '順位' }
	];

	const autumnHeaders = [
		...springHeaders,
		{ key: 'total_points_overall', label: '総合計' },
		{ key: 'rank_overall', label: '総合順位' }
	];

	$: headers = season === 'autumn' ? autumnHeaders : springHeaders;
</script>

<h1 class="text-2xl font-bold mb-4">点数一覧</h1>

{#if scores.length > 0}
	<div class="overflow-x-auto">
		<table class="table table-zebra w-full">
			<thead>
				<tr>
					{#each headers as header}
						<th>{header.label}</th>
					{/each}
				</tr>
			</thead>
			<tbody>
				{#each scores as score}
					<tr>
						{#each headers as header}
							<td>{score[header.key]}</td>
						{/each}
					</tr>
				{/each}
			</tbody>
		</table>
	</div>
{:else}
	<p>点数情報がまだありません。</p>
{/if}