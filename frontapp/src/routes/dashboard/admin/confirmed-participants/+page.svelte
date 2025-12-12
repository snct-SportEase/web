<script>
	import { onMount } from 'svelte';

	let classes = [];
	let allSports = [];
	let selectedClassId = null;
	let selectedSportId = null;
	let confirmedMembers = [];
	let confirmedCount = 0;
	let minCapacity = null;
	let capacityOK = true;
	let loading = false;
	let error = null;

	$: selectedClass = selectedClassId !== null ? classes.find((c) => c.id === selectedClassId) : null;

	function toNumber(value) {
		if (value === '' || value === null || value === undefined) {
			return null;
		}
		const parsed = Number(value);
		return Number.isNaN(parsed) ? null : parsed;
	}

	function authorizedFetch(url, options = {}) {
		const { headers, ...rest } = options;
		return fetch(url, {
			credentials: 'include',
			...rest,
			headers: {
				...(headers ?? {})
			}
		});
	}

	function getSportName(sportId) {
		const sport = allSports.find((s) => s.id === sportId);
		return sport ? sport.name : '不明な競技';
	}

	async function loadClasses() {
		try {
			const response = await authorizedFetch('/api/classes');
			if (!response.ok) {
				throw new Error('クラス一覧の取得に失敗しました');
			}
			classes = await response.json();
		} catch (err) {
			console.error('Error loading classes:', err);
			error = err.message || 'クラス一覧の取得に失敗しました';
		}
	}

	async function loadSports() {
		try {
			const response = await authorizedFetch('/api/admin/allsports');
			if (!response.ok) {
				throw new Error('競技一覧の取得に失敗しました');
			}
			allSports = await response.json();
		} catch (err) {
			console.error('Error loading sports:', err);
			error = err.message || '競技一覧の取得に失敗しました';
		}
	}

	async function loadConfirmedMembers() {
		if (!selectedClass || selectedSportId === null) {
			confirmedMembers = [];
			confirmedCount = 0;
			minCapacity = null;
			capacityOK = true;
			return;
		}

		loading = true;
		error = null;

		try {
			const url = `/api/admin/class-team/sports/${selectedSportId}/confirmed-members?class_id=${selectedClass.id}`;
			const response = await authorizedFetch(url);
			if (!response.ok) {
				const errorData = await response.json();
				throw new Error(errorData.error || '参加本登録済みメンバーの取得に失敗しました');
			}
			const data = await response.json();
			confirmedMembers = data.members || [];
			confirmedCount = data.confirmed_count || 0;
			minCapacity = data.min_capacity;
			capacityOK = data.capacity_ok !== false;
		} catch (err) {
			console.error('Error loading confirmed members:', err);
			error = err.message || '参加本登録済みメンバーの取得に失敗しました';
			confirmedMembers = [];
			confirmedCount = 0;
			minCapacity = null;
			capacityOK = true;
		} finally {
			loading = false;
		}
	}

	async function handleClassChange(event) {
		const value = toNumber(event?.target?.value ?? selectedClassId);
		selectedClassId = value;
		selectedSportId = null;
		confirmedMembers = [];
		confirmedCount = 0;
		minCapacity = null;
		capacityOK = true;
	}

	async function handleSportChange(event) {
		selectedSportId = toNumber(event?.target?.value ?? selectedSportId);
		await loadConfirmedMembers();
	}

	onMount(async () => {
		await Promise.all([loadClasses(), loadSports()]);
	});
</script>

<div class="container mx-auto p-6">
	<h1 class="text-3xl font-bold mb-6">参加本登録済みメンバー確認</h1>

	{#if error}
		<div class="mb-4 p-4 bg-red-100 border border-red-400 text-red-700 rounded">
			<p class="font-bold">エラー:</p>
			<p>{error}</p>
		</div>
	{/if}

	<div class="space-y-6">
		<!-- Class Selection -->
		<div class="bg-white p-6 rounded-lg shadow">
			<h2 class="text-xl font-semibold mb-4">クラス選択</h2>
			<select
				bind:value={selectedClassId}
				on:change={handleClassChange}
				class="w-full md:w-1/3 px-4 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
			>
				<option value={null}>クラスを選択してください</option>
				{#each classes as classItem}
					<option value={classItem.id}>{classItem.name}</option>
				{/each}
			</select>
		</div>

		{#if selectedClass}
			<!-- Sport Selection -->
			<div class="bg-white p-6 rounded-lg shadow">
				<h2 class="text-xl font-semibold mb-4">競技選択</h2>
				<select
					bind:value={selectedSportId}
					on:change={handleSportChange}
					class="w-full md:w-1/3 px-4 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
				>
					<option value={null}>競技を選択してください</option>
					{#each allSports as sport}
						<option value={sport.id}>{sport.name}</option>
					{/each}
				</select>
			</div>

			{#if selectedSportId !== null}
				<!-- Confirmed Members -->
				<div class="bg-white p-6 rounded-lg shadow">
					<div class="flex items-center justify-between mb-4">
						<h2 class="text-xl font-semibold">
							参加本登録済みメンバー ({getSportName(selectedSportId)})
						</h2>
						<div class="flex items-center space-x-4">
							{#if minCapacity !== null}
								<span class="text-sm {capacityOK ? 'text-green-600' : 'text-red-600'} font-semibold">
									{confirmedCount} / {minCapacity} 人
									{#if !capacityOK}
										（最低人数未満）
									{/if}
								</span>
							{:else}
								<span class="text-sm text-gray-600">
									{confirmedCount} 人
								</span>
							{/if}
						</div>
					</div>

					{#if !capacityOK}
						<div class="mb-4 p-4 bg-red-100 border border-red-400 text-red-700 rounded">
							<p class="font-bold">警告:</p>
							<p>
								{selectedClass.name}クラスの{getSportName(selectedSportId)}の参加本登録済みメンバー数（{confirmedCount}人）が最低人数（{minCapacity}人）に達していません。
							</p>
						</div>
					{/if}

					{#if loading}
						<p class="text-gray-500">読み込み中...</p>
					{:else if confirmedMembers.length === 0}
						<p class="text-gray-500">参加本登録済みメンバーがいません</p>
					{:else}
						<div class="overflow-x-auto">
							<table class="min-w-full divide-y divide-gray-200">
								<thead class="bg-gray-50">
									<tr>
										<th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">表示名</th>
										<th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">メールアドレス</th>
									</tr>
								</thead>
								<tbody class="bg-white divide-y divide-gray-200">
									{#each confirmedMembers as member}
										<tr class="hover:bg-gray-50">
											<td class="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
												{member.display_name || '未設定'}
											</td>
											<td class="px-6 py-4 whitespace-nowrap text-sm text-gray-500">{member.email}</td>
										</tr>
									{/each}
								</tbody>
							</table>
						</div>
					{/if}
				</div>
			{/if}
		{/if}
	</div>
</div>

<style>
	:global(body) {
		background-color: #f3f4f6;
	}
</style>

