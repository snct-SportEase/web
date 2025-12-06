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
		{ key: 'gym2_loser_bracket_champion_points', label: '敗者戦ブロック優勝' },
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

	// Sort scores by rank (1st, 2nd, 3rd, etc.)
	// Rank 0 (未開始) should be sorted last
	$: sortedScores = [...scores].sort((a, b) => {
		const rankA = season === 'spring' ? a.rank_current_event : a.rank_overall;
		const rankB = season === 'spring' ? b.rank_current_event : b.rank_overall;
		// If rank is 0, null, or undefined, treat it as Infinity for sorting (put it last)
		const normalizedRankA = (rankA === 0 || rankA === null || rankA === undefined) ? Infinity : rankA;
		const normalizedRankB = (rankB === 0 || rankB === null || rankB === undefined) ? Infinity : rankB;
		return normalizedRankA - normalizedRankB;
	});

	// Helper function to get rank style classes
	function getRankStyle(rank) {
		if (rank === 0 || rank === null || rank === undefined) {
			return 'bg-gray-100 border-2 border-gray-300 shadow';
		}
		if (rank === 1) {
			return 'rank-first relative overflow-hidden scale-105';
		} else if (rank === 2) {
			return 'bg-gradient-to-br from-gray-300 via-gray-200 to-gray-300 border-[3px] border-gray-400 shadow-lg scale-105';
		} else if (rank === 3) {
			return 'bg-gradient-to-br from-amber-700 via-amber-500 to-amber-700 border-[3px] border-amber-800 shadow-lg scale-105';
		}
		return 'bg-white border-2 border-gray-200 shadow';
	}

	// Helper function to get rank badge text
	function getRankBadge(rank) {
		if (rank === 0 || rank === null || rank === undefined) return '未開始';
		if (rank === 1) return '🥇';
		if (rank === 2) return '🥈';
		if (rank === 3) return '🥉';
		return `${rank}位`;
	}
</script>

<style>
	.rank-first {
		background: linear-gradient(135deg, #ffd700 0%, #ffed4e 30%, #ffd700 60%, #ffed4e 100%);
		border: 4px solid #ffb300;
		box-shadow: 
			0 10px 40px rgba(255, 215, 0, 0.5),
			0 0 30px rgba(255, 215, 0, 0.4),
			inset 0 0 20px rgba(255, 255, 255, 0.3);
		transform: scale(1.08);
		position: relative;
		overflow: hidden;
		animation: pulse-gold 2s ease-in-out infinite;
	}

	.rank-first::before {
		content: '';
		position: absolute;
		top: -50%;
		left: -50%;
		width: 200%;
		height: 200%;
		background: linear-gradient(45deg, transparent, rgba(255, 255, 255, 0.4), transparent);
		animation: shine 3s infinite;
		pointer-events: none;
	}

	.rank-first::after {
		content: '✨';
		position: absolute;
		top: 10px;
		right: 10px;
		font-size: 1.5rem;
		animation: twinkle 1.5s ease-in-out infinite;
		pointer-events: none;
	}

	@keyframes shine {
		0% {
			transform: translateX(-100%) translateY(-100%) rotate(45deg);
		}
		100% {
			transform: translateX(100%) translateY(100%) rotate(45deg);
		}
	}

	@keyframes pulse-gold {
		0%, 100% {
			box-shadow: 
				0 10px 40px rgba(255, 215, 0, 0.5),
				0 0 30px rgba(255, 215, 0, 0.4),
				inset 0 0 20px rgba(255, 255, 255, 0.3);
		}
		50% {
			box-shadow: 
				0 12px 50px rgba(255, 215, 0, 0.7),
				0 0 40px rgba(255, 215, 0, 0.6),
				inset 0 0 25px rgba(255, 255, 255, 0.4);
		}
	}

	@keyframes twinkle {
		0%, 100% {
			opacity: 0.5;
			transform: scale(1);
		}
		50% {
			opacity: 1;
			transform: scale(1.2);
		}
	}

	/* Custom styles for rank-first that require complex animations and gradients */
	.rank-first {
		background: linear-gradient(135deg, #ffd700 0%, #ffed4e 30%, #ffd700 60%, #ffed4e 100%);
		border: 4px solid #ffb300;
		box-shadow: 
			0 10px 40px rgba(255, 215, 0, 0.5),
			0 0 30px rgba(255, 215, 0, 0.4),
			inset 0 0 20px rgba(255, 255, 255, 0.3);
		animation: pulse-gold 2s ease-in-out infinite;
	}

	.rank-first:hover {
		transform: translateY(-6px) scale(1.1) !important;
		box-shadow: 
			0 15px 60px rgba(255, 215, 0, 0.6),
			0 0 50px rgba(255, 215, 0, 0.5),
			inset 0 0 30px rgba(255, 255, 255, 0.4);
	}
</style>

<h1 class="text-2xl font-bold mb-6">点数一覧</h1>

{#if scores.length > 0}
	<div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-6">
		{#each sortedScores as score}
			{@const rank = season === 'spring' ? score.rank_current_event : score.rank_overall}
			{@const totalPoints = season === 'spring' ? score.total_points_current_event : score.total_points_overall}
			{@const isNotStarted = rank === 0 || rank === null || rank === undefined}
			<div class="transition-all duration-300 rounded-xl p-6 mb-6 hover:-translate-y-1 hover:shadow-xl {getRankStyle(rank)}">
				{#if rank === 1}
					<span class="absolute top-2.5 right-2.5 text-2xl pointer-events-none animate-pulse">✨</span>
				{/if}
				<div class="text-3xl font-bold text-center mb-4 drop-shadow-md">{getRankBadge(rank)}</div>
				<div class="text-2xl font-bold text-center mb-4 {rank === 1 ? 'text-amber-900 text-[1.75rem] drop-shadow-[2px_2px_4px_rgba(0,0,0,0.3),0_0_10px_rgba(255,255,255,0.5)]' : rank === 2 ? 'text-gray-700 drop-shadow-sm' : rank === 3 ? 'text-amber-900 drop-shadow-sm' : isNotStarted ? 'text-gray-600 drop-shadow-sm' : 'text-gray-800 drop-shadow-sm'}">
					{score.class_name}
				</div>
				
				<div class="space-y-1">
					{#each filteredScoreItems as item}
						{#if item.key !== 'rank_current_event' && item.key !== 'rank_overall' && item.key !== 'total_points_current_event' && item.key !== 'total_points_overall'}
							<div class="flex justify-between py-2 border-b border-black/10">
								<span class="text-gray-500">{item.label}:</span>
								<span class="font-semibold text-gray-800">{score[item.key] || 0}</span>
							</div>
						{/if}
					{/each}
					
					<div class="flex justify-between py-3 mt-2 border-t-2 border-black/20 font-bold {rank === 1 ? 'text-[1.75rem]' : 'text-xl'}">
						<span class="text-gray-500">合計点:</span>
						<span class="font-bold {rank === 1 ? 'text-amber-900 drop-shadow-[1px_1px_2px_rgba(0,0,0,0.2)]' : 'text-gray-800'}">
							{totalPoints}
						</span>
					</div>
				</div>
			</div>
		{/each}
	</div>
{:else}
	<p class="text-gray-500">点数情報がまだありません。</p>
{/if}