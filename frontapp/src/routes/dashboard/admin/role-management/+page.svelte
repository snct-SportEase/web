<script>
	import { onMount } from 'svelte';
	import EditRoleModal from '../../../../lib/components/EditRoleModal.svelte';

	let usersWithRoles = [];
	let allUsers = [];
	let filteredUsers = [];
	let selectedUsers = [];
	let emailSearch = '';
	let role = '';
	let showUserList = false;

	let showEditModal = false;
	let selectedUserForEdit = null;

	let sortColumn = '';
	let sortAsc = true;

	const defaultRoles = ['root', 'admin', 'student'];

	function sortUsers(column) {
		if (sortColumn === column) {
			sortAsc = !sortAsc;
		} else {
			sortColumn = column;
			sortAsc = true;
		}

		usersWithRoles.sort((a, b) => {
			let aValue, bValue;

			if (column === 'roles') {
				aValue = a.roles?.find(r => !defaultRoles.includes(r.name))?.name || '';
				bValue = b.roles?.find(r => !defaultRoles.includes(r.name))?.name || '';
			} else {
				aValue = a[column] || '';
				bValue = b[column] || '';
			}

			if (aValue < bValue) {
				return sortAsc ? -1 : 1;
			}
			if (aValue > bValue) {
				return sortAsc ? 1 : -1;
			}
			return 0;
		});
		usersWithRoles = [...usersWithRoles];
	}

	function openEditModal(user) {
		selectedUserForEdit = user;
		showEditModal = true;
	}

	async function fetchUsersWithRoles() {
		const res = await fetch('/api/admin/users');
		if (res.ok) {
			const data = await res.json();
			console.log("fetchUsersWithRoles:", data);
			if (Array.isArray(data)) {
				usersWithRoles = data.filter(u => u.roles && u.roles.some(r => !defaultRoles.includes(r.name)));
			}
		}
	}

	async function fetchAllUsers() {
		const res = await fetch('/api/admin/users');
		if (res.ok) {
			allUsers = await res.json();
		}
	}

	function searchUsers() {
		if (emailSearch.trim() === '') {
			filteredUsers = allUsers;
		} else {
			filteredUsers = allUsers.filter(user =>
				user.email.toLowerCase().includes(emailSearch.toLowerCase())
			);
		}
	}

	function selectUser(user) {
		if (!selectedUsers.find(u => u.id === user.id)) {
			selectedUsers = [...selectedUsers, user];
		}
		emailSearch = '';
		showUserList = false;
	}

	function removeUser(user) {
		selectedUsers = selectedUsers.filter(u => u.ID !== user.ID);
	}

	async function assignRole() {
		if (selectedUsers.length === 0) {
			alert('ユーザーを選択してください。');
			return;
		}
		if (role.trim() === '') {
			alert('ロール名を入力してください。');
			return;
		}

		if (defaultRoles.includes(role.trim().toLowerCase())) {
			alert('デフォルトのロール（root, admin, student）は割り当てられません。');
			return;
		}

		for (const user of selectedUsers) {
			const updateRes = await fetch('/api/admin/users/role', {
				method: 'PUT',
				headers: {
					'Content-Type': 'application/json'
				},
				body: JSON.stringify({ user_id: user.id, role: role })
			});

			if (!updateRes.ok) {
				const error = await updateRes.json();
				alert(`ユーザー ${user.email} へのロール割り当てに失敗しました: ${error.message}`);
				// Optionally stop or continue
			}
		}

		alert('選択したユーザーにロールが割り当てられました。');
		selectedUsers = [];
		role = '';
		fetchUsersWithRoles();
	}

	onMount(() => {
		fetchUsersWithRoles();
	});
</script>

<EditRoleModal bind:showModal={showEditModal} user={selectedUserForEdit} on:roleDeleted={fetchUsersWithRoles} />

<h1 class="text-2xl font-bold mb-4">ロール管理</h1>

<div class="bg-white shadow-md rounded-lg p-6 mb-8">
    <h2 class="text-xl font-semibold mb-4">ロールの割り当て</h2>
    <div class="flex items-end space-x-4">
        <div class="w-full max-w-xs">
            <label for="email-search" class="block text-sm font-medium text-gray-700">ユーザー検索</label>
            <div class="mt-1 relative">
                <input
                    id="email-search"
                    type="email"
                    bind:value={emailSearch}
                    on:focus={async () => {
                        await fetchAllUsers();
                        searchUsers();
                        showUserList = true;
                    }}
                    on:input={searchUsers}
                    placeholder="メールアドレスで検索"
                    class="block w-full rounded-md border-gray-300 shadow-sm focus:border-indigo-500 focus:ring-indigo-500 sm:text-sm"
                />
                {#if showUserList && filteredUsers.length > 0}
                    <ul class="absolute z-10 mt-1 w-full bg-white shadow-lg max-h-60 rounded-md py-1 text-base ring-1 ring-black ring-opacity-5 overflow-auto focus:outline-none sm:text-sm">
                        {#each filteredUsers as user}
                            <li class="text-gray-900 cursor-default select-none relative py-2 pl-3 pr-9 hover:bg-gray-100" on:click={() => selectUser(user)}>
                                <span class="block truncate">{user.email}</span>
                            </li>
                        {/each}
                    </ul>
                {/if}
            </div>
            <div class="mt-2 space-x-1">
                {#each selectedUsers as user}
                    <span class="inline-flex items-center gap-x-1.5 rounded-full bg-indigo-100 px-2.5 py-1 text-sm font-semibold text-indigo-800">
                        {user.email}
                        <button on:click={() => removeUser(user)} class="-mr-0.5 h-5 w-5 p-0.5 rounded-full inline-flex items-center justify-center text-indigo-500 hover:bg-indigo-200 hover:text-indigo-600">
							<svg class="h-3 w-3" fill="none" viewBox="0 0 12 12" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 9l6-6M9 9L3 3"/></svg>
						</button>
                    </span>
                {/each}
            </div>
        </div>

        <div>
            <label for="role-name" class="block text-sm font-medium text-gray-700">ロール名</label>
            <input
                id="role-name"
                type="text"
                bind:value={role}
                placeholder="例: role"
                class="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-indigo-500 focus:ring-indigo-500 sm:text-sm"
            />
        </div>

        <button on:click={assignRole} class="inline-flex items-center justify-center rounded-md border border-transparent bg-indigo-600 px-4 py-2 text-sm font-medium text-white shadow-sm hover:bg-indigo-700 focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:ring-offset-2">
            <svg xmlns="http://www.w3.org/2000/svg" class="-ml-1 mr-2 h-5 w-5" viewBox="0 0 20 20" fill="currentColor">
                <path fill-rule="evenodd" d="M16.707 5.293a1 1 0 010 1.414l-8 8a1 1 0 01-1.414 0l-4-4a1 1 0 011.414-1.414L8 12.586l7.293-7.293a1 1 0 011.414 0z" clip-rule="evenodd" />
            </svg>
            割り当て
        </button>
    </div>
</div>

<div class="bg-white shadow-md rounded-lg p-6">
	<h2 class="text-xl font-semibold mb-4">現在のロールを持つユーザー</h2>
	<div class="overflow-x-auto">
		<table class="min-w-full divide-y divide-gray-200">
			<thead class="bg-gray-50">
				<tr>
					<th scope="col" class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider cursor-pointer" on:click={() => sortUsers('email')}>
						メールアドレス
						{#if sortColumn === 'email'}
							<span>{sortAsc ? '▲' : '▼'}</span>
						{/if}
					</th>
					<th scope="col" class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider cursor-pointer" on:click={() => sortUsers('display_name')}>
						表示名
						{#if sortColumn === 'display_name'}
							<span>{sortAsc ? '▲' : '▼'}</span>
						{/if}
					</th>
					<th scope="col" class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider cursor-pointer" on:click={() => sortUsers('roles')}>
						ロール
						{#if sortColumn === 'roles'}
							<span>{sortAsc ? '▲' : '▼'}</span>
						{/if}
					</th>
				</tr>
			</thead>
			<tbody class="bg-white divide-y divide-gray-200">
				{#each usersWithRoles as user}
					<tr class="hover:bg-gray-50 cursor-pointer" on:click={() => openEditModal(user)}>
						<td class="px-6 py-4 whitespace-nowrap text-sm font-medium text-gray-900">
							{user.email}
						</td>
						<td class="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
							{user.display_name || ''}
						</td>
						<td class="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
							{#if user.roles}
								<div class="flex space-x-2">
									{#each user.roles as role}
										{#if !defaultRoles.includes(role.name)}
											<span class="px-2 inline-flex text-xs leading-5 font-semibold rounded-full bg-indigo-100 text-indigo-800"										>
												{role.name}
											</span>
										{/if}
									{/each}
								</div>
							{/if}
						</td>
					</tr>
				{/each}
			</tbody>
		</table>
	</div>
</div>
