<script>
	import { onMount } from 'svelte';

	let teams = $state([]);
	let loading = $state(true);
	let error = $state(null);
	let activeEventId = $state(null);
	let activeEventName = $state('');

	let activeTeams = $derived(
		activeEventId && Array.isArray(teams)
			? teams.filter((team) => team.event_id === activeEventId)
			: teams
	);

	onMount(async () => {
		await initializeData();
	});

	async function initializeData() {
		loading = true;
		error = null;

		try {
			const [eventResponse, teamsResponse] = await Promise.all([
				fetch('/api/events/active', { credentials: 'include' }),
				fetch('/api/qrcode/teams', { credentials: 'include' })
			]);

			if (!eventResponse.ok) {
				throw new Error('開催中イベントの取得に失敗しました');
			}
			if (!teamsResponse.ok) {
				throw new Error('参加競技の取得に失敗しました');
			}

			const eventData = await eventResponse.json();
			activeEventId = eventData?.event_id ?? eventData?.id ?? null;
			activeEventName = eventData?.event_name ?? eventData?.name ?? '';

			const teamsData = await teamsResponse.json();
			teams = Array.isArray(teamsData) ? teamsData : [];
		} catch (err) {
			error = err.message || '参加競技の取得に失敗しました';
			teams = [];
		} finally {
			loading = false;
		}
	}
</script>

<div class="mx-auto max-w-4xl p-6 page-content">
	<h1 class="mb-6 text-2xl font-bold">参加競技確認</h1>

	{#if error}
		<div class="mb-4 rounded border border-red-400 bg-red-100 p-4 text-red-700">
			{error}
		</div>
	{/if}

	<section class="rounded bg-white p-6 shadow-sm">
		<div class="mb-6">
			<p class="text-sm text-gray-500">開催中イベント</p>
			<p class="text-lg font-semibold text-gray-900">{activeEventName || '未設定'}</p>
		</div>

		{#if loading}
			<p class="text-gray-500">読み込み中...</p>
		{:else if !activeTeams || activeTeams.length === 0}
			<p class="text-gray-600">参加登録されている競技がありません。</p>
		{:else}
			<div class="overflow-x-auto">
				<table class="min-w-full divide-y divide-gray-200">
					<thead class="bg-gray-50">
						<tr>
							<th class="px-4 py-3 text-left text-xs font-medium uppercase text-gray-500">競技</th>
							<th class="px-4 py-3 text-left text-xs font-medium uppercase text-gray-500">チーム</th>
						</tr>
					</thead>
					<tbody class="divide-y divide-gray-200 bg-white">
						{#each activeTeams as team (team.id)}
							<tr class="hover:bg-gray-50">
								<td class="px-4 py-3 text-sm font-medium text-gray-900">{team.sport_name}</td>
								<td class="px-4 py-3 text-sm text-gray-600">{team.name}</td>
							</tr>
						{/each}
					</tbody>
				</table>
			</div>
		{/if}
	</section>
</div>
