<script>
	export let data;

	const {
		classes: initialClasses = [],
		classMembers: initialClassMembers = [],
		eventSports: initialEventSports = [],
		allSports: initialAllSports = [],
		selectedClassId: initialSelectedClassId = null,
		error: initialError = null
	} = data ?? {};

	let classes = [...initialClasses];
	let selectedClassId = initialSelectedClassId;
	let classMembers = [...initialClassMembers];
	let eventSports = [...initialEventSports];
	let allSports = [...initialAllSports];
	let selectedSportId = null;
	let selectedMembers = [];
	let assignedMembers = [];

	let membersLoading = false;
	let teamMembersLoading = false;
	let sportsLoading = false;
	let assignLoading = false;

	let error = initialError;
	let success = null;
	const isAdmin = data.isAdmin || false;

if (selectedClassId !== null && typeof selectedClassId !== 'number') {
	const parsedInitialClassId = Number(selectedClassId);
	selectedClassId = Number.isNaN(parsedInitialClassId) ? null : parsedInitialClassId;
}

	$: selectedClass =
		selectedClassId !== null ? classes.find((c) => c.id === selectedClassId) : null;
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

	async function loadClassMembers(classId) {
		if (!classId) return;

		membersLoading = true;
		error = null;

		try {
			const response = await authorizedFetch(`/api/admin/class-team/classes/${classId}/members`);
			if (!response.ok) {
				const errorData = await response.json();
				throw new Error(errorData.error || 'クラスメンバーの取得に失敗しました');
			}
			classMembers = await response.json();
			selectedMembers = [];
		} catch (err) {
			console.error('Error loading class members:', err);
			error = err.message || 'クラスメンバーの取得に失敗しました';
		} finally {
			membersLoading = false;
		}
	}

	async function loadTeamMembers(sportId) {
		if (!selectedClass || !sportId) {
			assignedMembers = [];
			return;
		}

		teamMembersLoading = true;
		error = null;

		try {
			let url = `/api/admin/class-team/sports/${sportId}/members`;
			if (isAdmin) {
				url += `?class_id=${selectedClass.id}`;
			}

			const response = await authorizedFetch(url);
			if (!response.ok) {
				const errorData = await response.json();
				throw new Error(errorData.error || 'チームメンバーの取得に失敗しました');
			}
			assignedMembers = await response.json();
		} catch (err) {
			console.error('Error loading team members:', err);
			error = err.message || 'チームメンバーの取得に失敗しました';
		} finally {
			teamMembersLoading = false;
		}
	}

	async function handleClassChange(event) {
		const value = toNumber(event?.target?.value ?? selectedClassId);
		selectedClassId = value;
		if (selectedClassId !== null) {
			await loadClassMembers(selectedClassId);
			selectedSportId = null;
			assignedMembers = [];
			selectedMembers = [];
		}
	}

	async function handleSportChange(event) {
		selectedSportId = toNumber(event?.target?.value ?? selectedSportId);
		if (selectedSportId !== null) {
			await loadTeamMembers(selectedSportId);
		} else {
			assignedMembers = [];
		}
	}

	function toggleMemberSelection(user) {
		if (selectedMembers.find((m) => m.id === user.id)) {
			selectedMembers = selectedMembers.filter((m) => m.id !== user.id);
		} else {
			selectedMembers = [...selectedMembers, user];
		}
	}

	async function assignMembers() {
		if (!selectedClass || selectedSportId === null || selectedMembers.length === 0) {
			error = 'クラス、競技、およびメンバーを選択してください';
			return;
		}

		assignLoading = true;
		error = null;
		success = null;

		try {
			const requestBody = {
				sport_id: selectedSportId,
				user_ids: selectedMembers.map((m) => m.id)
			};

			if (isAdmin && selectedClass) {
				requestBody.class_id = selectedClass.id;
			}

			const response = await authorizedFetch('/api/admin/class-team/assign-members', {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify(requestBody)
			});

			if (!response.ok) {
				const errorData = await response.json();
				throw new Error(errorData.error || 'メンバーの割り当てに失敗しました');
			}

			const result = await response.json();
			success = result.message || 'メンバーの割り当てが完了しました';

			if (selectedSportId !== null) {
				await loadTeamMembers(selectedSportId);
			}
			selectedMembers = [];
		} catch (err) {
			console.error('Error assigning members:', err);
			error = err.message || 'メンバーの割り当てに失敗しました';
		} finally {
			assignLoading = false;
		}
	}

	async function removeMember(user) {
		if (!selectedClass || selectedSportId === null) return;
		if (!confirm(`${user.display_name || user.email} をチームから削除しますか？`)) return;

		assignLoading = true; // Use same loading state
		error = null;
		success = null;

		try {
			const requestBody = {
				sport_id: selectedSportId,
				user_id: user.id
			};

			if (isAdmin && selectedClass) {
				requestBody.class_id = selectedClass.id;
			}

			const response = await authorizedFetch('/api/admin/class-team/remove-member', {
				method: 'DELETE',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify(requestBody)
			});

			if (!response.ok) {
				const errorData = await response.json();
				throw new Error(errorData.error || 'メンバーの削除に失敗しました');
			}

			const result = await response.json();
			success = result.message || 'メンバーを削除しました';

			await loadTeamMembers(selectedSportId);
		} catch (err) {
			console.error('Error removing member:', err);
			error = err.message || 'メンバーの削除に失敗しました';
		} finally {
			assignLoading = false;
		}
	}

</script>

<div class="max-w-6xl mx-auto p-6">
	<h1 class="text-2xl font-bold mb-6">クラス・チーム管理</h1>

	{#if error}
		<div class="mb-4 p-4 bg-red-100 border border-red-400 text-red-700 rounded">
			{error}
		</div>
	{/if}

	{#if success}
		<div class="mb-4 p-4 bg-green-100 border border-green-400 text-green-700 rounded">
			{success}
		</div>
	{/if}

	<div class="space-y-6">
		<!-- Class Selection -->
		<div class="bg-white p-6 rounded-lg shadow">
			<h2 class="text-xl font-semibold mb-4">クラス選択</h2>
			{#if isAdmin}
				<select
					value={selectedClassId ?? ''}
					on:change={handleClassChange}
					class="w-full p-3 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
					disabled={membersLoading || assignLoading}
				>
					<option value=''>クラスを選択してください</option>
					{#each classes as cls}
						<option value={cls.id} selected={selectedClassId === cls.id}>{cls.name}</option>
					{/each}
				</select>
			{:else if classes.length > 0}
				<p class="text-gray-700 font-semibold">{classes[0].name}</p>
				<p class="text-sm text-gray-500">あなたの担当クラス</p>
			{/if}
		</div>

		{#if selectedClass}
			<!-- Class Members -->
			<div class="bg-white p-6 rounded-lg shadow">
				<h2 class="text-xl font-semibold mb-4">クラスメンバー</h2>
				{#if membersLoading && classMembers.length === 0}
					<p class="text-gray-500">読み込み中...</p>
				{:else if classMembers.length === 0}
					<p class="text-gray-500">メンバーが登録されていません</p>
				{:else}
					<div class="overflow-x-auto">
						<table class="min-w-full divide-y divide-gray-200">
							<thead class="bg-gray-50">
								<tr>
									<th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">選択</th>
									<th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">表示名</th>
									<th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">メールアドレス</th>
								</tr>
							</thead>
							<tbody class="bg-white divide-y divide-gray-200">
								{#each classMembers as member}
									<tr class="hover:bg-gray-50">
										<td class="px-6 py-4 whitespace-nowrap">
											<input
												type="checkbox"
												checked={selectedMembers.some((m) => m.id === member.id)}
												on:change={() => toggleMemberSelection(member)}
												class="rounded border-gray-300 text-blue-600 focus:ring-blue-500"
											/>
										</td>
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

			<!-- Sport Selection -->
			<div class="bg-white p-6 rounded-lg shadow">
				<h2 class="text-xl font-semibold mb-4">競技選択</h2>
				<select
					value={selectedSportId ?? ''}
					on:change={handleSportChange}
					class="w-full p-3 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
					disabled={sportsLoading || membersLoading || assignLoading}
				>
					<option value=''>競技を選択してください</option>
					{#each eventSports as eventSport}
						<option value={eventSport.sport_id} selected={selectedSportId === eventSport.sport_id}>
							{getSportName(eventSport.sport_id)}
						</option>
					{/each}
				</select>
			</div>

			{#if selectedSportId !== null && selectedMembers.length > 0}
				<!-- Assign Button -->
				<div class="bg-white p-6 rounded-lg shadow">
					<button
						on:click={assignMembers}
						disabled={assignLoading}
						class="w-full bg-blue-600 text-white py-2 px-4 rounded-md hover:bg-blue-700 disabled:bg-gray-400 disabled:cursor-not-allowed"
					>
						{assignLoading ? '割り当て中...' : `選択した${selectedMembers.length}名を割り当てる`}
					</button>
				</div>
			{/if}

			{#if selectedSportId !== null}
				<!-- Assigned Team Members -->
				<div class="bg-white p-6 rounded-lg shadow">
					<h2 class="text-xl font-semibold mb-4">
						割り当て済みメンバー ({getSportName(selectedSportId)})
					</h2>
					{#if teamMembersLoading && assignedMembers.length === 0}
						<p class="text-gray-500">読み込み中...</p>
					{:else if assignedMembers.length === 0}
						<p class="text-gray-500">メンバーが割り当てられていません</p>
					{:else}
						<div class="overflow-x-auto">
							<table class="min-w-full divide-y divide-gray-200">
								<thead class="bg-gray-50">
									<tr>
										<th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">表示名</th>
										<th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">メールアドレス</th>
										<th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">操作</th>
									</tr>
								</thead>
								<tbody class="bg-white divide-y divide-gray-200">
									{#each assignedMembers as member}
										<tr class="hover:bg-gray-50">
											<td class="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
												{member.display_name || '未設定'}
											</td>
											<td class="px-6 py-4 whitespace-nowrap text-sm text-gray-500">{member.email}</td>
											<td class="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
												<button
													on:click={() => removeMember(member)}
													disabled={assignLoading}
													class="text-red-600 hover:text-red-900 disabled:text-gray-400"
												>
													削除
												</button>
											</td>
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
