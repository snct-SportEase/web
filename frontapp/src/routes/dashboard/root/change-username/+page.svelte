<script>
  import { onMount } from 'svelte';
  import { browser } from '$app/environment';

  let users = [];
  let searchQuery = '';
  let searchType = 'email'; // 'email' or 'display_name'
  let selectedUser = null;
  let newDisplayName = '';
  let showModal = false;

  async function fetchUsers(query = '', type = '') {
    if (!browser) return;
    try {
      const url = `/api/root/users?query=${encodeURIComponent(query)}&searchType=${encodeURIComponent(type)}`;
      const response = await fetch(url);
      if (response.ok) {
        users = await response.json();
      } else {
        console.error('Failed to fetch users:', response.statusText);
      }
    } catch (error) {
      console.error('Error fetching users:', error);
    }
  }

  async function searchUsers() {
    await fetchUsers(searchQuery, searchType);
  }

  function openEditModal(user) {
    selectedUser = user;
    newDisplayName = user.display_name || '';
    showModal = true;
  }

  function closeEditModal() {
    showModal = false;
    selectedUser = null;
    newDisplayName = '';
  }

  async function handleDisplayNameUpdate() {
    if (!selectedUser || !browser) return;
    try {
      const response = await fetch('/api/root/users/display-name', {
        method: 'PUT',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          user_id: selectedUser.id,
          display_name: newDisplayName,
        }),
      });
      if (response.ok) {
      // Update the user in the local list
      const updatedUser = { ...selectedUser, display_name: newDisplayName };
      const index = users.findIndex(u => u.id === selectedUser.id);
      if (index !== -1) {
        users[index] = updatedUser;
        users = [...users]; // Trigger reactivity
      }
        closeEditModal();
        location.reload();
      } else {
        console.error('Failed to update display name:', response.statusText);
      }
    } catch (error) {
      console.error('Error updating display name:', error);
    }
  }

  onMount(async () => {
    await fetchUsers();
  });
</script>

<div class="container mx-auto p-4 lg:p-8">
  <h1 class="text-3xl font-bold mb-6 text-center">ユーザー名変更</h1>

  <div class="card bg-base-100 shadow-xl mb-8">
    <div class="card-body">
      <h2 class="card-title">ユーザー検索</h2>
      <div class="flex flex-col md:flex-row items-center space-y-4 md:space-y-0 md:space-x-4">
        <div class="join">
          <input class="join-item btn" type="radio" name="searchType" value="email" bind:group={searchType} aria-label="Email" checked />
          <input class="join-item btn" type="radio" name="searchType" value="display_name" bind:group={searchType} aria-label="表示名" />
        </div>
        <div class="form-control w-full md:max-w-xs">
          <label class="input input-bordered flex items-center gap-2">
            <input type="text" class="grow" placeholder="検索..." bind:value={searchQuery} on:keydown.enter={searchUsers} />
            <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 16 16" fill="currentColor" class="w-4 h-4 opacity-70"><path fill-rule="evenodd" d="M9.965 11.026a5 5 0 1 1 1.06-1.06l2.755 2.754a.75.75 0 1 1-1.06 1.06l-2.755-2.754ZM10.5 7a3.5 3.5 0 1 1-7 0 3.5 3.5 0 0 1 7 0Z" clip-rule="evenodd" /></svg>
          </label>
        </div>
        <button class="btn btn-primary" on:click={searchUsers}>
          検索
        </button>
      </div>
    </div>
  </div>

  <div class="card bg-base-100 shadow-xl">
    <div class="card-body">
      <div class="overflow-x-auto">
        <table class="table table-zebra w-full">
          <thead>
            <tr class="text-base">
              <th class="uppercase">Email</th>
              <th class="uppercase">表示名</th>
              <th class="uppercase">クラスID</th>
              <th class="uppercase">アクション</th>
            </tr>
          </thead>
          <tbody>
            {#each users as user (user.id)}
              <tr>
                <td>{user.email}</td>
                <td>{user.display_name || '-'}</td>
                <td>{user.class_id || '-'}</td>
                <td>
                  <button class="btn btn-sm btn-outline btn-primary" on:click={() => openEditModal(user)}>
                    <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" viewBox="0 0 20 20" fill="currentColor">
                      <path d="M17.414 2.586a2 2 0 00-2.828 0L7 10.172V13h2.828l7.586-7.586a2 2 0 000-2.828z" />
                      <path fill-rule="evenodd" d="M2 6a2 2 0 012-2h4a1 1 0 010 2H4v10h10v-4a1 1 0 112 0v4a2 2 0 01-2 2H4a2 2 0 01-2-2V6z" clip-rule="evenodd" />
                    </svg>
                    編集
                  </button>
                </td>
              </tr>
            {/each}
          </tbody>
        </table>
      </div>
    </div>
  </div>
</div>

{#if showModal}
  <div class="modal modal-open">
    <div class="modal-box">
      <h3 class="font-bold text-lg mb-4">表示名を編集</h3>
      <p class="py-2">ユーザー: <span class="font-mono">{selectedUser.email}</span></p>
      <div class="form-control w-full">
        <label class="label" for="displayNameInput">
          <span class="label-text">新しい表示名</span>
        </label>
        <input id="displayNameInput" type="text" placeholder="新しい表示名を入力" class="input input-bordered w-full" bind:value={newDisplayName} />
      </div>
      <div class="modal-action mt-6">
        <button class="btn btn-primary" on:click={handleDisplayNameUpdate}>
          <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" viewBox="0 0 20 20" fill="currentColor">
            <path fill-rule="evenodd" d="M16.707 5.293a1 1 0 010 1.414l-8 8a1 1 0 01-1.414 0l-4-4a1 1 0 011.414-1.414L8 12.586l7.293-7.293a1 1 0 011.414 0z" clip-rule="evenodd" />
          </svg>
          保存
        </button>
        <button class="btn" on:click={closeEditModal}>
          <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" viewBox="0 0 20 20" fill="currentColor">
            <path fill-rule="evenodd" d="M4.293 4.293a1 1 0 011.414 0L10 8.586l4.293-4.293a1 1 0 111.414 1.414L11.414 10l4.293 4.293a1 1 0 01-1.414 1.414L10 11.414l-4.293 4.293a1 1 0 01-1.414-1.414L8.586 10 4.293 5.707a1 1 0 010-1.414z" clip-rule="evenodd" />
          </svg>
          キャンセル
        </button>
      </div>
    </div>
  </div>
{/if}
