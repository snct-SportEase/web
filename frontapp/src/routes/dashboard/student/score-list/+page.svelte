<script>
	/** @type {import('./$types').PageData} */
	export let data;

	$: scores = data.scores || [];
	$: season = scores.length > 0 ? scores[0].season : '';
	$: sportNames = scores.length > 0 ? scores[0].sport_names : {};

	// Helper function to get sport name for a given location
	function getSportName(location) {
		return sportNames[location] || location; // Fallback to location if sport name not found
	}

	// Define all possible score items and their labels
	$: scoreItemDefinitions = [
		{ key: 'initial_points', label: '初期点' },
		{ key: 'survey_points', label: 'アンケート' },
		{ key: 'attendance_points', label: '出席点' },
		{ key: 'gym1_win1_points', label: `${getSportName('gym1')}1勝点` },
		{ key: 'gym1_win2_points', label: `${getSportName('gym1')}2勝点` },
		{ key: 'gym1_win3_points', label: `${getSportName('gym1')}3勝点` },
		{ key: 'gym1_champion_points', label: `${getSportName('gym1')}優勝点` },
		{ key: 'gym2_win1_points', label: `${getSportName('gym2')}1勝点` },
		{ key: 'gym2_win2_points', label: `${getSportName('gym2')}2勝点` },
		{ key: 'gym2_win3_points', label: `${getSportName('gym2')}3勝点` },
		{ key: 'gym2_champion_points', label: `${getSportName('gym2')}優勝点` },
		{ key: 'ground_win1_points', label: `${getSportName('ground')}1勝点` },
		{ key: 'ground_win2_points', label: `${getSportName('ground')}2勝点` },
		{ key: 'ground_win3_points', label: `${getSportName('ground')}3勝点` },
		{ key: 'ground_champion_points', label: `${getSportName('ground')}優勝点` },
		{ key: 'noon_game_points', label: '昼競技' },
		{ key: 'total_points_current_event', label: '合計点' },
		{ key: 'rank_current_event', label: '順位' },
		{ key: 'total_points_overall', label: '総合点' },
		{ key: 'rank_overall', label: '総合順位' }
	];

	// Filter score items based on season
	$: filteredScoreItems = scoreItemDefinitions.filter(item => {
		if (season === 'spring') {
			return item.key !== 'initial_points' && item.key !== 'total_points_overall' && item.key !== 'rank_overall';
		}
		return true; // For autumn, include all
	});

	// Extract unique class names for column headers
	$: classNames = [...new Set(scores.map(s => s.class_name))].sort();

	// Pivot the data
	$: pivotedScores = filteredScoreItems.map(itemDef => {
		const row = { label: itemDef.label };
		classNames.forEach(className => {
			const scoreForClass = scores.find(s => s.class_name === className);
			row[className] = scoreForClass ? scoreForClass[itemDef.key] : '-'; // Use '-' for missing scores
		});
		return row;
	});
</script>

<h1 class="text-2xl font-bold mb-4">点数一覧</h1>

{#if scores.length > 0}
	<div class="overflow-x-auto relative shadow-md rounded-lg">
		<table class="table">
			<thead>
				<tr>
					<th class="sticky left-0 z-20 bg-black text-white whitespace-nowrap">得点項目</th>
					{#each classNames as className}
						<th class="bg-black text-white min-w-[120px] text-center text-lg">{className}</th>
					{/each}
				</tr>
			</thead>
			<tbody>
				{#each pivotedScores as row, i}
					<tr class="hover">
						<td class="sticky left-0 z-10 bg-black text-white whitespace-nowrap">{row.label}</td>
						{#each classNames as className}
							<td class="text-center">{row[className]}</td>
						{/each}
					</tr>
				{/each}
			</tbody>
		</table>
	</div>
{:else}
	<p>点数情報がまだありません。</p>
{/if}